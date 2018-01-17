package commands

import (
	"errors"
	"fmt"
	"github.com/iron-io/functions/fn/common"
	"github.com/urfave/cli"
	"io"
	"os"
	"os/exec"
	"strings"
)

func Run() cli.Command {
	r := runCmd{}

	return cli.Command{
		Name:      "run",
		Usage:     "run a function locally",
		ArgsUsage: "[username/image:tag]",
		Flags:     append(Runflags(), []cli.Flag{}...),
		Action:    r.run,
	}
}

type runCmd struct{}

func Runflags() []cli.Flag {
	return []cli.Flag{
		cli.StringSliceFlag{
			Name:  "e",
			Usage: "select environment variables to be sent to function",
		},
		cli.StringSliceFlag{
			Name:  "link",
			Usage: "select container links for the function",
		},
		cli.StringFlag{
			Name:  "method",
			Usage: "http method for function",
		},
	}
}

func (r *runCmd) run(c *cli.Context) error {
	image := c.Args().First()
	if image == "" {
		ff, err := common.LoadFuncfile()
		if err != nil {
			if _, ok := err.(*common.NotFoundError); ok {
				return errors.New("error: image name is missing or no function file found")
			}
			return err
		}
		image = ff.FullName()
	}

	return Runff(image, Stdin(), os.Stdout, os.Stderr, c.String("method"), c.StringSlice("e"), c.StringSlice("link"))
}

func Runff(image string, stdin io.Reader, stdout, stderr io.Writer, method string, restrictedEnv []string, links []string) error {
	sh := []string{"docker", "run", "--rm", "-i"}

	var env []string
	detectedEnv := os.Environ()
	if len(restrictedEnv) > 0 {
		detectedEnv = restrictedEnv
	}

	if method == "" {
		if stdin == nil {
			method = "GET"
		} else {
			method = "POST"
		}
	}
	sh = append(sh, "-e", kvEq("METHOD", method))

	for _, e := range detectedEnv {
		shellvar, envvar := extractEnvVar(e)
		sh = append(sh, shellvar...)
		env = append(env, envvar)
	}

	for _, l := range links {
		sh = append(sh, "--link", l)
	}

	dockerenv := []string{"DOCKER_TLS_VERIFY", "DOCKER_HOST", "DOCKER_CERT_PATH", "DOCKER_MACHINE_NAME"}
	for _, e := range dockerenv {
		env = append(env, fmt.Sprint(e, "=", os.Getenv(e)))
	}

	sh = append(sh, image)
	cmd := exec.Command(sh[0], sh[1:]...)
	cmd.Stdin = stdin
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	cmd.Env = env
	return cmd.Run()
}

func extractEnvVar(e string) ([]string, string) {
	kv := strings.Split(e, "=")
	name := toEnvName("HEADER", kv[0])
	sh := []string{"-e", name}
	var v string
	if len(kv) > 1 {
		v = kv[1]
	} else {
		v = os.Getenv(kv[0])
	}
	return sh, kvEq(name, v)
}

func kvEq(k, v string) string {
	return fmt.Sprintf("%s=%s", k, v)
}

// From server.toEnvName()
func toEnvName(envtype, name string) string {
	name = strings.ToUpper(strings.Replace(name, "-", "_", -1))
	return fmt.Sprintf("%s_%s", envtype, name)
}
