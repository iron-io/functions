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

type lambdaCmd struct {
	functionName  string
	runtime       string
	handler       string
	fileNames     []string
	payload       string
	clientContext string
	arn           string
	version       string
	downloadOnly  bool
	awsProfile    string
	image         string
	awsRegion     string
}

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

func (lcc *lambdaCmd) getFlags() []cli.Flag {
	return []cli.Flag{
		cli.StringFlag{
			Name:        "payload",
			Usage:       "Payload to pass to the Lambda function. This is usually a JSON object.",
			Destination: &lcc.payload,
			Value:       "{}",
		},
		cli.StringFlag{
			Name:        "client-context",
			Usage:       "",
			Destination: &lcc.clientContext,
		},

		cli.StringFlag{
			Name:        "image",
			Usage:       "By default the name of the Docker image is the name of the Lambda function. Use this to set a custom name.",
			Destination: &lcc.image,
		},

		cli.StringFlag{
			Name:        "version",
			Usage:       "Version of the function to import.",
			Destination: &lcc.version,
			Value:       "$LATEST",
		},

		cli.BoolFlag{
			Name:        "download-only",
			Usage:       "Only download the function into a directory. Will not create a Docker image.",
			Destination: &lcc.downloadOnly,
		},

		cli.StringFlag{
			Name:        "profile",
			Usage:       "AWS Profile to load from credentials file.",
			Destination: &lcc.awsProfile,
		},

		cli.StringFlag{
			Name:        "region",
			Usage:       "AWS region to use.",
			Value:       "us-east-1",
			Destination: &lcc.awsRegion,
		},
	}
}

func (lcc *lambdaCmd) downloadToFile(url string) (string, error) {
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

func (lcc *lambdaCmd) unzipAndGetTopLevelFiles(dst, src string) (files []lambdaImpl.FileLike, topErr error) {
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

func (lcc *lambdaCmd) getFunction() (*aws_lambda.GetFunctionOutput, error) {
	creds := aws_credentials.NewChainCredentials([]aws_credentials.Provider{
		&aws_credentials.EnvProvider{},
		&aws_credentials.SharedCredentialsProvider{
			Filename: "", // Look in default location.
			Profile:  lcc.awsProfile,
		},
	})

	conf := aws.NewConfig().WithCredentials(creds).WithCredentialsChainVerboseErrors(true).WithRegion(lcc.awsRegion)
	sess := aws_session.New(conf)
	conn := aws_lambda.New(sess)
	resp, err := conn.GetFunction(&aws_lambda.GetFunctionInput{
		FunctionName: aws.String(lcc.arn),
		Qualifier:    aws.String(lcc.version),
	})

	return resp, err
}

func (lcc *lambdaCmd) init(c *cli.Context) {
	clientContext := c.String("client-context")
	payload := c.String("payload")
	version := c.String("version")
	downloadOnly := c.Bool("download-only")
	image := c.String("image")
	profile := c.String("profile")
	region := c.String("region")

	if c.Command.Name == "aws-import" {
		if len(c.Args()) > 0 {
			lcc.arn = c.Args()[0]
		}
	} else {
		lcc.fileNames = c.Args()
	}
	lcc.clientContext = clientContext
	lcc.payload = payload
	lcc.version = version
	lcc.downloadOnly = downloadOnly
	lcc.awsProfile = profile
	lcc.image = image
	lcc.awsRegion = region
}

func (lcc *lambdaCmd) create(c *cli.Context) error {
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

func (lcc *lambdaCmd) runTest(c *cli.Context) error {
	lcc.init(c)
	exists, err := lambdaImpl.ImageExists(lcc.functionName)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("Function %s does not exist.", lcc.functionName)
	}

	// Redirect output to stdout.
	return lambdaImpl.RunImageWithPayload(lcc.functionName, lcc.payload)
}

func (lcc *lambdaCmd) awsImport(c *cli.Context) error {
	lcc.init(c)
	function, err := lcc.getFunction()
	if err != nil {
		return err
	}
	functionName := *function.Configuration.FunctionName

	err = os.Mkdir(fmt.Sprintf("./%s", functionName), os.ModePerm)
	if err != nil {
		return err
	}

	tmpFileName, err := lcc.downloadToFile(*function.Code.Location)
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
		files, err = lcc.unzipAndGetTopLevelFiles(functionName, tmpFileName)
		if err != nil {
			return err
		}
	}

	if lcc.downloadOnly {
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

	if lcc.image != "" {
		opts.Name = lcc.image
	}

	if *function.Configuration.Runtime == availableRuntimes[2] {
		opts.Package = filepath.Base(files[0].(*os.File).Name())
	}

	return lambdaImpl.CreateImage(opts, files...)
}

func lambda() cli.Command {
	lcc := lambdaCmd{}
	var flags []cli.Flag

	flags = append(flags, lcc.getFlags()...)

	return cli.Command{
		Name:      "lambda",
		Usage:     "create and publish lambda functions",
		ArgsUsage: "fnclt lambda",
		Subcommands: []cli.Command{
			{
				Name:      "create-function",
				Usage:     `Create Docker image that can run your Lambda function. The files are the contents of the zip file to be uploaded to AWS Lambda.`,
				ArgsUsage: "NAME RUNTIME HANDLER file [files...]",
				Action:    lcc.create,
				Flags:     flags,
			},
			{
				Name:      "test-function",
				Usage:     `Runs local Dockerized Lambda function and writes output to stdout.`,
				ArgsUsage: "--function-name NAME [--client-context <value>] [--payload <value>]",
				Action:    lcc.runTest,
				Flags:     flags,
			},
			{
				Name:      "aws-import",
				Usage:     `Converts an existing Lambda function to an image. The function code is downloaded to a directory in the current working directory that has the same name as the Lambda function..`,
				ArgsUsage: "[--region <region>] [--profile <aws profile>] [--version <version>] [--download-only] [--image <name>] ARN",
				Action:    lcc.awsImport,
				Flags:     flags,
			},
		},
	}
}
