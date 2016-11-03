package main

import (
	"archive/zip"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/docker/docker/pkg/jsonmessage"
	"github.com/urfave/cli"

	aws_credentials "github.com/aws/aws-sdk-go/aws/credentials"
	aws_session "github.com/aws/aws-sdk-go/aws/session"
	aws_lambda "github.com/aws/aws-sdk-go/service/lambda"
	lambdaImpl "github.com/iron-io/lambda/lambda"
)

var availableRuntimes = []string{"nodejs", "python2.7", "java8"}

type dockerJSONWriter struct {
	under io.Writer
	w     io.Writer
}

func newdockerJSONWriter(under io.Writer) *dockerJSONWriter {
	r, w := io.Pipe()
	go func() {
		err := jsonmessage.DisplayJSONMessagesStream(r, under, 1, true, nil)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}()
	return &dockerJSONWriter{under, w}
}

func (djw *dockerJSONWriter) Write(p []byte) (int, error) {
	return djw.w.Write(p)
}

func getFlags() []cli.Flag {
	return []cli.Flag{
		cli.StringFlag{
			Name:  "payload",
			Usage: "Payload to pass to the Lambda function. This is usually a JSON object.",
			Value: "{}",
		},
		cli.StringFlag{
			Name:  "version",
			Usage: "Version of the function to import.",
			Value: "$LATEST",
		},
		cli.BoolFlag{
			Name:  "download-only",
			Usage: "Only download the function into a directory. Will not create a Docker image.",
		},
	}
}

func downloadToFile(url string) (string, error) {
	downloadResp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer downloadResp.Body.Close()

	// zip reader needs ReaderAt, hence the indirection.
	tmpFile, err := ioutil.TempFile("", "lambda-function-")
	if err != nil {
		return "", err
	}

	if _, err := io.Copy(tmpFile, downloadResp.Body); err != nil {
		return "", err
	}
	if err := tmpFile.Close(); err != nil {
		return "", err
	}
	return tmpFile.Name(), nil
}

func unzipAndGetTopLevelFiles(dst, src string) (files []lambdaImpl.FileLike, topErr error) {
	files = make([]lambdaImpl.FileLike, 0)

	zipReader, err := zip.OpenReader(src)
	if err != nil {
		return files, err
	}
	defer zipReader.Close()

	var fd *os.File
	for _, f := range zipReader.File {
		path := filepath.Join(dst, f.Name)
		fmt.Printf("Extracting '%s' to '%s'\n", f.Name, path)
		if f.FileInfo().IsDir() {
			if err := os.Mkdir(path, 0644); err != nil {
				return nil, err
			}
			// Only top-level dirs go into the list since that is what CreateImage expects.
			if filepath.Dir(f.Name) == filepath.Base(f.Name) {
				fd, topErr = os.Open(path)
				if topErr != nil {
					break
				}
				files = append(files, fd)
			}
		} else {
			// We do not close fd here since we may want to use it to dockerize.
			fd, topErr = os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0644)
			if topErr != nil {
				break
			}

			var zipFd io.ReadCloser
			zipFd, topErr = f.Open()
			if topErr != nil {
				break
			}

			if _, topErr = io.Copy(fd, zipFd); topErr != nil {
				// OK to skip closing fd here.
				break
			}

			if err := zipFd.Close(); err != nil {
				return nil, err
			}

			// Only top-level files go into the list since that is what CreateImage expects.
			if filepath.Dir(f.Name) == "." {
				if _, topErr = fd.Seek(0, 0); topErr != nil {
					break
				}

				files = append(files, fd)
			} else {
				if err := fd.Close(); err != nil {
					return nil, err
				}
			}
		}
	}
	return
}

func getFunction(awsProfile, awsRegion, version, arn string) (*aws_lambda.GetFunctionOutput, error) {
	creds := aws_credentials.NewChainCredentials([]aws_credentials.Provider{
		&aws_credentials.EnvProvider{},
		&aws_credentials.SharedCredentialsProvider{
			Filename: "", // Look in default location.
			Profile:  awsProfile,
		},
	})

	conf := aws.NewConfig().WithCredentials(creds).WithCredentialsChainVerboseErrors(true).WithRegion(awsRegion)
	sess := aws_session.New(conf)
	conn := aws_lambda.New(sess)
	resp, err := conn.GetFunction(&aws_lambda.GetFunctionInput{
		FunctionName: aws.String(arn),
		Qualifier:    aws.String(version),
	})

	return resp, err
}

func create(c *cli.Context) error {
	args := c.Args()
	if len(args) < 4 {
		return fmt.Errorf("Expected at least 4 arguments, NAME RUNTIME HANDLER and file %d", len(args))
	}
	functionName := args[0]
	runtime := args[1]
	handler := args[2]
	fileNames := args[3:]

	files := make([]lambdaImpl.FileLike, 0, len(fileNames))
	opts := lambdaImpl.CreateImageOptions{
		Name:          functionName,
		Base:          fmt.Sprintf("iron/lambda-%s", runtime),
		Package:       "",
		Handler:       handler,
		OutputStream:  newdockerJSONWriter(os.Stdout),
		RawJSONStream: true,
	}

	if handler == "" {
		return errors.New("No handler specified.")
	}

	// For Java we allow only 1 file and it MUST be a JAR.
	if runtime == availableRuntimes[2] {
		if len(fileNames) != 1 {
			return errors.New("Java Lambda functions can only include 1 file and it must be a JAR file.")
		}

		if filepath.Ext(fileNames[0]) != ".jar" {
			return errors.New("Java Lambda function package must be a JAR file.")
		}

		opts.Package = filepath.Base(fileNames[0])
	}

	for _, fileName := range fileNames {
		file, err := os.Open(fileName)
		if err != nil {
			return err
		}
		defer file.Close()
		files = append(files, file)
	}

	return lambdaImpl.CreateImage(opts, files...)
}

func test(c *cli.Context) error {
	args := c.Args()
	if len(args) < 1 {
		return fmt.Errorf("Missing NAME argument")
	}
	functionName := args[0]

	exists, err := lambdaImpl.ImageExists(functionName)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("Function %s does not exist.", functionName)
	}

	payload := c.String("payload")
	// Redirect output to stdout.
	return lambdaImpl.RunImageWithPayload(functionName, payload)
}

func awsImport(c *cli.Context) error {
	args := c.Args()
	if len(args) < 3 {
		return fmt.Errorf("Missing arguments ARN, REGION and/or IMAGE")
	}

	version := c.String("version")
	downloadOnly := c.Bool("download-only")
	profile := c.String("profile")
	arn := args[0]
	region := args[1]
	image := args[2]

	function, err := getFunction(profile, region, version, arn)
	if err != nil {
		return err
	}
	functionName := *function.Configuration.FunctionName

	err = os.Mkdir(fmt.Sprintf("./%s", functionName), os.ModePerm)
	if err != nil {
		return err
	}

	tmpFileName, err := downloadToFile(*function.Code.Location)
	if err != nil {
		return err
	}
	defer os.Remove(tmpFileName)

	var files []lambdaImpl.FileLike

	if *function.Configuration.Runtime == availableRuntimes[2] {
		fmt.Println("Found Java Lambda function. Going to assume code is a single JAR file.")
		path := filepath.Join(functionName, "function.jar")
		if err := os.Rename(tmpFileName, path); err != nil {
			return err
		}
		fd, err := os.Open(path)
		if err != nil {
			return err
		}

		files = append(files, fd)
	} else {
		files, err = unzipAndGetTopLevelFiles(functionName, tmpFileName)
		if err != nil {
			return err
		}
	}

	if downloadOnly {
		// Since we are a command line program that will quit soon, it is OK to
		// let the OS clean `files` up.
		return err
	}

	opts := lambdaImpl.CreateImageOptions{
		Name:          functionName,
		Base:          fmt.Sprintf("iron/lambda-%s", *function.Configuration.Runtime),
		Package:       "",
		Handler:       *function.Configuration.Handler,
		OutputStream:  newdockerJSONWriter(os.Stdout),
		RawJSONStream: true,
	}

	if image != "" {
		opts.Name = image
	}

	if *function.Configuration.Runtime == availableRuntimes[2] {
		opts.Package = filepath.Base(files[0].(*os.File).Name())
	}

	return lambdaImpl.CreateImage(opts, files...)
}

func lambda() cli.Command {
	var flags []cli.Flag

	flags = append(flags, getFlags()...)

	return cli.Command{
		Name:      "lambda",
		Usage:     "create and publish lambda functions",
		ArgsUsage: "fnclt lambda",
		Subcommands: []cli.Command{
			{
				Name:      "create-function",
				Usage:     `create Docker image that can run your Lambda function, where files are the contents of the zip file to be uploaded to AWS Lambda.`,
				ArgsUsage: "name runtime handler /path [/paths...]",
				Action:    create,
				Flags:     flags,
			},
			{
				Name:      "test-function",
				Usage:     `runs local dockerized Lambda function and writes output to stdout.`,
				ArgsUsage: "name [--payload <value>]",
				Action:    test,
				Flags:     flags,
			},
			{
				Name:      "aws-import",
				Usage:     `converts an existing Lambda function to an image, where the function code is downloaded to a directory in the current working directory that has the same name as the Lambda function.`,
				ArgsUsage: "arn region image/name [--profile <aws profile>] [--version <version>] [--download-only]",
				Action:    awsImport,
				Flags:     flags,
			},
		},
	}
}
