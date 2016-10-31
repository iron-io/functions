package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"

	functions "github.com/iron-io/functions_go"
	"github.com/urfave/cli"
)

func publish() cli.Command {
	cmd := publishcmd{
		commoncmd: &commoncmd{},
		RoutesApi: functions.NewRoutesApi(),
	}
	var flags []cli.Flag
	flags = append(flags, cmd.flags()...)
	flags = append(flags, cmd.commoncmd.flags()...)
	flags = append(flags, confFlags(&cmd.Configuration)...)
	return cli.Command{
		Name:   "publish",
		Usage:  "scan local directory for functions, build and publish them.",
		Flags:  flags,
		Action: cmd.scan,
	}
}

type publishcmd struct {
	*commoncmd
	*functions.RoutesApi

	dry      bool
	skippush bool
}

func (u *publishcmd) flags() []cli.Flag {
	return []cli.Flag{
		cli.BoolFlag{
			Name:        "skip-push",
			Usage:       "does not push Docker built images onto Docker Hub - useful for local development.",
			Destination: &u.skippush,
		},
	}
}

func (u *publishcmd) scan(c *cli.Context) error {
	scan(u.verbose, u.wd, u.walker)
	return nil
}

func (u *publishcmd) walker(path string, info os.FileInfo, err error, w io.Writer) error {
	walker(path, info, err, w, u.publish)
	return nil
}

// publish will take the found function and check for the presence of a
// Dockerfile, and run a three step process: parse functions file, build and
// push the container, and finally it will update function's route. Optionally,
// the route can be overriden inside the functions file.
func (u *publishcmd) publish(path string) error {
	fmt.Fprintln(verbwriter, "publishing", path)

	funcfile, err := buildfunc(path)
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

func extractAppNameRoute(path string) (appName, route string) {

	// The idea here is to extract the root-most directory name
	// as application name, it turns out that stdlib tools are great to
	// extract the deepest one. Thus, we revert the string and use the
	// stdlib as it is - and revert back to its normal content. Not fastest
	// ever, but it is simple.

	rpath := reverse(path)
	rroute, rappName := filepath.Split(rpath)
	route = filepath.Dir(reverse(rroute))
	return reverse(rappName), route
}

func reverse(s string) string {
	r := []rune(s)
	for i, j := 0, len(r)-1; i < len(r)/2; i, j = i+1, j-1 {
		r[i], r[j] = r[j], r[i]
	}
	return string(r)
}
