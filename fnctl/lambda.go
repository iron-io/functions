package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/docker/docker/pkg/jsonmessage"
	"github.com/iron-io/iron_go3/config"
	lambdaImpl "github.com/iron-io/lambda/lambda"
	"github.com/urfave/cli"
)

var availableRuntimes = []string{"nodejs", "python2.7", "java8"}

const (
	skipFunctionName = iota
	requireFunctionName
)

type LambdaFlags struct {
	*flag.FlagSet
}

func (lf *LambdaFlags) validateAllFlags(fnRequired int) error {
	fn := lf.Lookup("function-name")
	// Everything except import needs a function
	if fnRequired == requireFunctionName && (fn == nil || fn.Value.String() == "") {
		return errors.New(fmt.Sprintf("Please specify function-name."))
	}

	selectedRuntime := lf.Lookup("runtime")
	if selectedRuntime != nil {
		validRuntime := false
		for _, r := range availableRuntimes {
			if selectedRuntime.Value.String() == r {
				validRuntime = true
			}
		}

		if !validRuntime {
			return fmt.Errorf("Invalid runtime. Supported runtimes %s", availableRuntimes)
		}
	}

	return nil
}

func (lf *LambdaFlags) functionName() *string {
	return lf.String("function-name", "", "Name of function. This is usually follows Docker image naming conventions.")
}

func (lf *LambdaFlags) handler() *string {
	return lf.String("handler", "", "function/class that is the entrypoint for this function. Of the form <module name>.<function name> for nodejs/Python, <full class name>::<function name base> for Java.")
}

func (lf *LambdaFlags) runtime() *string {
	return lf.String("runtime", "", fmt.Sprintf("Runtime that your Lambda function depends on. Valid values are %s.", strings.Join(availableRuntimes, ", ")))
}

func (lf *LambdaFlags) clientContext() *string {
	return lf.String("client-context", "", "")
}

func (lf *LambdaFlags) payload() *string {
	return lf.String("payload", "", "Payload to pass to the Lambda function. This is usually a JSON object.")
}

func (lf *LambdaFlags) image() *string {
	return lf.String("image", "", "By default the name of the Docker image is the name of the Lambda function. Use this to set a custom name.")
}

func (lf *LambdaFlags) version() *string {
	return lf.String("version", "$LATEST", "Version of the function to import.")
}

func (lf *LambdaFlags) downloadOnly() *bool {
	return lf.Bool("download-only", false, "Only download the function into a directory. Will not create a Docker image.")
}

func (lf *LambdaFlags) awsProfile() *string {
	return lf.String("profile", "", "AWS Profile to load from credentials file.")
}

func (lf *LambdaFlags) awsRegion() *string {
	return lf.String("region", "us-east-1", "AWS region to use.")
}

type lambdaCmd struct {
	settings  config.Settings
	flags     *LambdaFlags
	token     *string
	projectID *string
}

type LambdaCreateCmd struct {
	lambdaCmd

	functionName *string
	runtime      *string
	handler      *string
	fileNames    []string
}

func (lcc *LambdaCreateCmd) Args() error {
	if lcc.flags.NArg() < 1 {
		return errors.New(`lambda create requires at least one file`)
	}

	for _, arg := range lcc.flags.Args() {
		lcc.fileNames = append(lcc.fileNames, arg)
	}

	return nil
}

func (lcc *LambdaCreateCmd) Usage() {
	fmt.Fprintln(os.Stderr, `usage: iron lambda create-function --function-name NAME --runtime RUNTIME --handler HANDLER file [files...]
Create Docker image that can run your Lambda function. The files are the contents of the zip file to be uploaded to AWS LambdaImpl.
`)
	lcc.flags.PrintDefaults()
}

func (lcc *LambdaCreateCmd) Config() error {
	return nil
}

func (lcc *LambdaCreateCmd) Flags(args ...string) error {
	flags := flag.NewFlagSet("commands", flag.ContinueOnError)
	flags.Usage = func() {}
	lcc.flags = &LambdaFlags{flags}

	lcc.functionName = lcc.flags.functionName()
	lcc.handler = lcc.flags.handler()
	lcc.runtime = lcc.flags.runtime()

	if err := lcc.flags.Parse(args); err != nil {
		return err
	}

	return lcc.flags.validateAllFlags(requireFunctionName)
}

type DockerJsonWriter struct {
	under io.Writer
	w     io.Writer
}

func NewDockerJsonWriter(under io.Writer) *DockerJsonWriter {
	r, w := io.Pipe()
	go func() {
		err := jsonmessage.DisplayJSONMessagesStream(r, under, 1, true, nil)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}()
	return &DockerJsonWriter{under, w}
}

func (djw *DockerJsonWriter) Write(p []byte) (int, error) {
	return djw.w.Write(p)
}

func (lcc *LambdaCreateCmd) Run(c *cli.Context) {
	lcc.fileNames = c.Args()
	files := make([]lambdaImpl.FileLike, 0, len(lcc.fileNames))
	opts := lambdaImpl.CreateImageOptions{
		Name:          *lcc.functionName,
		Base:          fmt.Sprintf("iron/lambda-%s", *lcc.runtime),
		Package:       "",
		Handler:       *lcc.handler,
		OutputStream:  NewDockerJsonWriter(os.Stdout),
		RawJSONStream: true,
	}

	if *lcc.handler == "" {
		fmt.Fprintln(os.Stderr, "No handler specified.")
		os.Exit(1)
	}

	// For Java we allow only 1 file and it MUST be a JAR.
	if *lcc.runtime == "java8" {
		if len(lcc.fileNames) != 1 {
			fmt.Fprintln(os.Stderr, "Java Lambda functions can only include 1 file and it must be a JAR file.")
			os.Exit(1)
		}

		if filepath.Ext(lcc.fileNames[0]) != ".jar" {
			fmt.Fprintln(os.Stderr, "Java Lambda function package must be a JAR file.")
			os.Exit(1)
		}

		opts.Package = filepath.Base(lcc.fileNames[0])
	}

	for _, fileName := range lcc.fileNames {
		file, err := os.Open(fileName)
		defer file.Close()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		files = append(files, file)
	}

	err := lambdaImpl.CreateImage(opts, files...)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func (lcc *LambdaCreateCmd) getFlags() []cli.Flag {
	return []cli.Flag{
		cli.StringFlag{
			Name:        "function-name",
			Usage:       "Name of function. This is usually follows Docker image naming conventions.",
			Destination: lcc.functionName,
		},
		cli.StringFlag{
			Name:        "runtime",
			Usage:       fmt.Sprintf("Runtime that your Lambda function depends on. Valid values are %s.", strings.Join(availableRuntimes, ", ")),
			Destination: lcc.runtime,
		},
		cli.StringFlag{
			Name:        "handler",
			Usage:       "function/class that is the entrypoint for this function. Of the form <module name>.<function name> for nodejs/Python, <full class name>::<function name base> for Java.",
			Destination: lcc.handler,
		},
	}
}

func lambda() cli.Command {
	lcc := LambdaCreateCmd{}
	var flags []cli.Flag

	//init flags
	flag.Parse()
	lcc.Flags()
	lcc.Args()

	flags = append(flags, lcc.getFlags()...)
	return cli.Command{
		Name:      "lambda",
		Usage:     "create and publish lambda functions",
		ArgsUsage: "fnclt lambda",
		Subcommands: []cli.Command{
			{
				Name:      "create-function",
				Usage:     `Create Docker image that can run your Lambda function. The files are the contents of the zip file to be uploaded to AWS LambdaImpl.`,
				ArgsUsage: "--function-name NAME --runtime RUNTIME --handler HANDLER file [files...]",
				Action:    lcc.Run,
				Flags:     flags,
			},
		},
	}
}
