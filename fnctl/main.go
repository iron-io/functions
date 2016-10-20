package main

/*
	# list apps
	iron apps
	# create app
	iron apps create

	# create route
	iron routes --app APP_NAME create
	# list routes
	iron routes --app APP_NAME
*/

import (
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
	app.Authors = []cli.Author{{Name: "iron.io"}}
	app.Usage = "IronFunctions command line tools"
	app.CommandNotFound = func(c *cli.Context, cmd string) { fmt.Fprintf(os.Stderr, "command not found: %v\n", cmd) }
	app.Commands = []cli.Command{
		apps(),
		routes(),
	}
	app.Run(os.Args)
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
