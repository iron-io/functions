package main

import (
	"errors"
	"fmt"
	"net/url"
	"os"

	functions "github.com/iron-io/functions_go"
	"github.com/urfave/cli"
)

func main() {
	app := cli.NewApp()
	app.Name = "fnctl"
	app.Version = "0.0.1"
	app.Authors = []cli.Author{cli.Author{Name: "iron.io"}}
	app.Usage = "IronFunctions command line tools"
	app.CommandNotFound = func(c *cli.Context, cmd string) { fmt.Fprintf(os.Stderr, "command not found: %v\n", cmd) }
	app.Commands = []cli.Command{
		newAppsGet(),
		newAppGet(),
		//newAppPut(),
		//newAppsList(),
		//newAppsPost(),
		//newAppRoutesList(),
		//newAppRoutesDelete(),
		//newAppRoutesGet(),
	}
	app.Run(os.Args)
}

type AppsGet struct { // TODO
	*functions.AppsApi
}

func newAppsGet() cli.Command {
	a := AppsGet{AppsApi: functions.NewAppsApi()}

	return cli.Command{
		Name:      "apps",
		Usage:     "list all apps",
		ArgsUsage: "fnclt apps",
		Flags:     append(confFlags(&a.Configuration), []cli.Flag{}...),
		Action:    a.appsGet,
	}
}

func (a *AppsGet) appsGet(c *cli.Context) error {
	resetBasePath(&a.Configuration)

	wrapper, _, err := a.AppsGet()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error getting app: %v", err)
		return nil // TODO return error instead?
	}

	for _, app := range wrapper.Apps {
		fmt.Println(app.Name)
	}
	return nil
}

type AppGet struct { // TODO
	*functions.AppsApi
}

func newAppGet() cli.Command {
	a := AppGet{AppsApi: functions.NewAppsApi()}

	return cli.Command{
		Name:      "getapp",
		Usage:     "get information about an app",
		ArgsUsage: "fnclt getapp [app]",
		Flags:     append(confFlags(&a.Configuration), []cli.Flag{}...),
		Action:    a.appGet,
	}
}

func (a *AppGet) appGet(c *cli.Context) error {
	if c.Args().First() == "" {
		return errors.New("error: app get takes one arg, an app name")
	}

	resetBasePath(&a.Configuration)

	name := c.Args().Get(0)
	wrapper, _, err := a.AppsAppGet(name)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error getting app: %v", err)
		return nil // TODO return error instead?
	}

	fmt.Println(wrapper.App.Name)
	return nil
}

func resetBasePath(c *functions.Configuration) {
	var u url.URL
	u.Scheme = c.Scheme
	u.Host = c.Host
	u.Path = "/v1"
	c.BasePath = u.String()
}

func confFlags(c *functions.Configuration) []cli.Flag {
	return []cli.Flag{
		cli.StringFlag{
			Name:        "host",
			Usage:       "raw host path to functions api, e.g. functions.iron.io",
			Destination: &c.Host,
			EnvVar:      "HOST",
			Value:       "127.0.0.1:8080",
		},
		cli.StringFlag{
			Name:        "scheme",
			Usage:       "http/https",
			Destination: &c.Scheme,
			EnvVar:      "SCHEME",
			Value:       "http",
		},
	}
}
