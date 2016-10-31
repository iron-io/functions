package main

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"

	functions "github.com/iron-io/functions_go"
	"github.com/urfave/cli"
)


func publish() cli.Command {
	cmd := publishcmd{RoutesApi: functions.NewRoutesApi()}
	var flags []cli.Flag
	flags = append(flags, cmd.flags()...)
	flags = append(flags, confFlags(&cmd.Configuration)...)
	return cli.Command{
		Name:   "publish",
		Usage:  "scan local directory for functions, build and publish them.",
		Flags:  flags,
		Action: cmd.scan,
	}
}

type publishcmd struct {
	*functions.RoutesApi

	wd       string
	dry      bool
	skippush bool
	verbose  bool
}

func (u *publishcmd) flags() []cli.Flag {
	return []cli.Flag{
		cli.StringFlag{
			Name:        "d",
			Usage:       "working directory",
			Destination: &u.wd,
			EnvVar:      "WORK_DIR",
			Value:       "./",
		},
		cli.BoolFlag{
			Name:        "skip-push",
			Usage:       "does not push Docker built images onto Docker Hub - useful for local development.",
			Destination: &u.skippush,
		},
		cli.BoolFlag{
			Name:        "dry-run",
			Usage:       "display how update will proceed when executed",
			Destination: &u.dry,
		},
		cli.BoolFlag{
			Name:        "v",
			Usage:       "verbose mode",
			Destination: &u.verbose,
		},
	}
}

func (u *publishcmd) scan(c *cli.Context) error {
	scan(u.verbose, u.wd, u.walker)
	return nil
}

func (u *publishcmd) walker(path string, info os.FileInfo, err error, w io.Writer) error {
	if !isvalid(path, info) {
		return nil
	}

	fmt.Fprint(w, path, "\t")
	if u.dry {
		fmt.Fprintln(w, "dry-run")
	} else if err := u.update(path); err != nil {
		fmt.Fprintln(w, err)
	} else {
		fmt.Fprintln(w, "updated")
	}

	return nil
}

// update will take the found function and check for the presence of a Dockerfile,
// and run a three step process: parse functions file, build and push the
// container, and finally it will update function's route. Optionally, the route
// can be overriden inside the functions file.
func (u *publishcmd) update(path string) error {
	fmt.Fprintln(verbwriter, "deploying", path)

	funcfile, err := buildFunc(path)
	if err != nil {
		return err
	}

	if u.skippush {
		return nil
	}

	if err := u.dockerpush(funcfile.Image); err != nil {
		return err
	}

	if err := u.route(path, funcfile); err != nil {
		return err
	}

	return nil
}

func (publishcmd) dockerpush(image string) error {
	out, err := exec.Command("docker", "push", image).CombinedOutput()
	fmt.Fprintf(verbwriter, "%s\n", out)
	if err != nil {
		return fmt.Errorf("error running docker push: %v", err)
	}

	return nil
}

func (u *publishcmd) route(path string, ff *funcfile) error {
	resetBasePath(&u.Configuration)

	an, r := extractAppNameRoute(path)
	if ff.App == nil {
		ff.App = &an
	}
	if ff.Route == nil {
		ff.Route = &r
	}

	body := functions.RouteWrapper{
		Route: functions.Route{
			Path:  *ff.Route,
			Image: ff.Image,
		},
	}

	fmt.Fprintf(verbwriter, "updating API with appName: %s route: %s image: %s \n", *ff.App, *ff.Route, ff.Image)

	_, _, err := u.AppsAppRoutesPost(*ff.App, body)
	if err != nil {
		return fmt.Errorf("error getting routes: %v", err)
	}

	return nil
}
