package main

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"path"
	"text/tabwriter"

	"github.com/iron-io/functions_go"
	"github.com/urfave/cli"
)

type routesCmd struct {
	*functions.RoutesApi
}

func routes() cli.Command {
	r := routesCmd{RoutesApi: functions.NewRoutesApi()}

	return cli.Command{
		Name:      "routes",
		Usage:     "list routes",
		ArgsUsage: "fnclt routes",
		Flags:     append(confFlags(&r.Configuration), []cli.Flag{}...),
		Action:    r.list,
		Subcommands: []cli.Command{
			{
				Name:      "create",
				Usage:     "create a route",
				ArgsUsage: "appName /path image/name",
				Action:    r.create,
			},
			{
				Name:      "delete",
				Usage:     "delete a route",
				ArgsUsage: "appName /path",
				Action:    r.delete,
			},
		},
	}
}

func (a *routesCmd) list(c *cli.Context) error {
	if c.Args().First() == "" {
		return errors.New("error: routes listing takes one argument, an app name")
	}

	resetBasePath(&a.Configuration)

	appName := c.Args().Get(0)
	wrapper, _, err := a.AppsAppRoutesGet(appName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error getting routes: %v", err)
		return nil // TODO return error instead?
	}

	baseURL, err := url.Parse(a.Configuration.BasePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error parsing base path: %v", err)
		return nil // TODO return error instead?
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 8, 0, '\t', 0)
	fmt.Fprint(w, "path", "\t", "image", "\t", "endpoint", "\n")
	for _, route := range wrapper.Routes {
		u, err := url.Parse("../")
		u.Path = path.Join(u.Path, "r", appName, route.Path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error parsing functions route path: %v", err)
			return nil // TODO return error instead?
		}

		fmt.Fprint(w, route.Path, "\t", route.Image, "\t", baseURL.ResolveReference(u).String(), "\n")
	}
	w.Flush()

	return nil
}

func (a *routesCmd) create(c *cli.Context) error {
	if c.Args().Get(0) == "" || c.Args().Get(1) == "" || c.Args().Get(2) == "" {
		return errors.New("error: routes listing takes three arguments: an app name, a path and an image")
	}

	resetBasePath(&a.Configuration)

	name := c.Args().Get(0)
	path := c.Args().Get(1)
	image := c.Args().Get(2)
	body := functions.RouteWrapper{
		Route: functions.Route{
			AppName: name,
			Path:    path,
			Image:   image,
		},
	}
	wrapper, _, err := a.AppsAppRoutesPost(name, body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error getting routes: %v", err)
		return nil // TODO return error instead?
	}

	fmt.Println(wrapper.Route.Path, "created with", wrapper.Route.Image)
	return nil
}

func (a *routesCmd) delete(c *cli.Context) error {
	if c.Args().Get(0) == "" || c.Args().Get(1) == "" {
		return errors.New("error: routes listing takes three arguments: an app name and a path")
	}

	resetBasePath(&a.Configuration)

	name := c.Args().Get(0)
	path := c.Args().Get(1)
	_, err := a.AppsAppRoutesRouteDelete(name, path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error getting routes: %v", err)
		return nil // TODO return error instead?
	}

	fmt.Println(path, "deleted")
	return nil
}