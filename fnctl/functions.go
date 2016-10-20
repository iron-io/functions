package main

import (
	"errors"
	"fmt"
	"os"

	functions "github.com/iron-io/functions_go"
	"github.com/urfave/cli"
)

type AppGet struct { // TODO
	functions.AppsApi
}

func newAppGet() cli.Command {
	var a AppGet
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

	fmt.Println(a)

	name := c.Args().Get(0)
	wrapper, _, err := a.AppsAppGet(name)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error getting app: %v", err)
		return nil // TODO return error instead?
	}
	fmt.Println(wrapper.App)
	return nil
}

func confFlags(c *functions.Configuration) []cli.Flag {
	// TODO full support... need more stable swagger/discuss some fields
	//UserName     string            `json:"userName,omitempty"`
	//Password     string            `json:"password,omitempty"`
	//APIKeyPrefix map[string]string `json:"APIKeyPrefix,omitempty"`
	//APIKey       map[string]string `json:"APIKey,omitempty"`

	//DebugFile     string            `json:"debugFile,omitempty"`
	//OAuthToken    string            `json:"oAuthToken,omitempty"`
	//Timeout       int               `json:"timeout,omitempty"`
	//BasePath      string            `json:"basePath,omitempty"`
	//Host          string            `json:"host,omitempty"`
	//Scheme        string            `json:"scheme,omitempty"`
	//AccessToken   string            `json:"accessToken,omitempty"`
	//DefaultHeader map[string]string `json:"defaultHeader,omitempty"`
	//UserAgent     string            `json:"userAgent,omitempty"`
	//APIClient     APIClient         `json:"APIClient,omitempty"`

	return []cli.Flag{
		cli.StringFlag{
			Name:        "username",
			Usage:       "your user name",
			Destination: &c.UserName,
			EnvVar:      "USERNAME",
		},
		cli.StringFlag{
			Name:        "password",
			Usage:       "password",
			Destination: &c.Password,
			EnvVar:      "PASSWORD",
		},
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

//func(a *AppGet) Flags(...string) error {
//return nil
//}

//func(a *AppGet) Args() error           {
//if s.flags.NArg() < 1 {
//return errors.New("error: app get takes one arg, an app name")
//}

//a.Name = s.flags.Arg(0)
//return nil
//}

//func(a *AppGet) Config() error         {
//var c functions.Configuration
//envconfig.Init(&c)
//return nil
//}
//func(a *AppGet) Usage()                { }
//func(a *AppGet) Run()                  { }

func main() {
	app := cli.NewApp()
	app.Name = "fnctl"
	app.Version = "0.0.1"
	app.Authors = []cli.Author{cli.Author{Name: "iron.io"}}
	app.Usage = "IronFunctions command line tools"
	// app.Flags = ???
	app.CommandNotFound = func(c *cli.Context, cmd string) { fmt.Fprintf(os.Stderr, "command not found: %v\n", cmd) }
	app.Commands = []cli.Command{
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

//type AppGet struct { // TODO
//}
type AppPut struct { // TODO
}
type AppsList struct { // TODO
}
type AppsPost struct { // TODO
}
type AppRoutesList struct { // TODO
}
type AppRoutesDelete struct { // TODO
}
type AppRoutesGet struct { // TODO
}
