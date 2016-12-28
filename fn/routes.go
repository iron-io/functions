package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

	functions "github.com/iron-io/functions_go"
	"github.com/urfave/cli"
)

type routesCmd struct {
	*functions.RoutesApi
}

func routes() cli.Command {
	r := routesCmd{RoutesApi: functions.NewRoutesApi()}

	return cli.Command{
		Name:      "routes",
		Usage:     "operate routes",
		ArgsUsage: "fn routes",
		Subcommands: []cli.Command{
			{
				Name:      "call",
				Usage:     "call a route",
				ArgsUsage: "`app` /path",
				Action:    r.call,
				Flags:     runflags(),
			},
			{
				Name:      "list",
				Aliases:   []string{"l"},
				Usage:     "list routes for `app`",
				ArgsUsage: "`app`",
				Action:    r.list,
			},
			{
				Name:      "create",
				Aliases:   []string{"c"},
				Usage:     "create a route in an `app`",
				ArgsUsage: "`app` /path image/name",
				Action:    r.create,
				Flags: []cli.Flag{
					cli.Int64Flag{
						Name:  "memory,m",
						Usage: "memory in MiB",
						Value: 128,
					},
					cli.StringFlag{
						Name:  "type,t",
						Usage: "route type - sync or async",
						Value: "sync",
					},
					cli.StringSliceFlag{
						Name:  "config,c",
						Usage: "route configuration",
					},
					cli.StringFlag{
						Name:  "format,f",
						Usage: "hot container IO format - json or http",
						Value: "",
					},
					cli.IntFlag{
						Name:  "max-concurrency",
						Usage: "maximum concurrency for hot container",
						Value: 1,
					},
					cli.DurationFlag{
						Name:  "timeout",
						Usage: "route timeout",
						Value: 30 * time.Second,
					},
				},
			},
			{
				Name:      "delete",
				Aliases:   []string{"d"},
				Usage:     "delete a route from `app`",
				ArgsUsage: "`app` /path",
				Action:    r.delete,
			},
			{
				Name:  "config",
				Usage: "operate a route configuration set",
				Subcommands: []cli.Command{
					{
						Name:      "view",
						Aliases:   []string{"v"},
						Usage:     "view all configuration keys for this route",
						ArgsUsage: "`app` /path",
						Action:    r.configList,
						Flags: []cli.Flag{
							cli.BoolFlag{
								Name:  "shell,s",
								Usage: "output in shell format",
							},
							cli.BoolFlag{
								Name:  "json,j",
								Usage: "output in JSON format",
							},
						},
					},
					{
						Name:      "set",
						Aliases:   []string{"s"},
						Usage:     "store a configuration key for this route",
						ArgsUsage: "`app` /path <key> <value>",
						Action:    r.configSet,
					},
					{
						Name:      "unset",
						Aliases:   []string{"u"},
						Usage:     "remove a configuration key for this route",
						ArgsUsage: "`app` /path <key>",
						Action:    r.configUnset,
					},
				},
			},

			{
				Name:      "image",
				Usage:     "get/set route's image",
				ArgsUsage: "`app` /path [new image]",
				Action:    r.imageGetSet,
			},

			{
				Name:      "format",
				Usage:     "get/set route's format",
				ArgsUsage: "`app` /path [http|default]",
				Action:    r.formatGetSet,
			},

			{
				Name:      "type",
				Usage:     "get/set route's type",
				ArgsUsage: "`app` /path [sync|async]",
				Action:    r.typeGetSet,
			},

			{
				Name:      "maxconcurrency",
				Usage:     "get/set route's max concurrency (hot containers only)",
				ArgsUsage: "`app` /path [number of concurrent containers]",
				Action:    r.maxConcurrencyGetSet,
			},

			{
				Name:      "timeout",
				Usage:     "get/set route's execution timeout",
				ArgsUsage: "`app` /path [seconds]",
				Action:    r.timeoutGetSet,
			},
		},
	}
}

func call() cli.Command {
	r := routesCmd{RoutesApi: functions.NewRoutesApi()}

	return cli.Command{
		Name:      "call",
		Usage:     "call a remote function",
		ArgsUsage: "`app` /path",
		Flags:     runflags(),
		Action:    r.call,
	}
}

func (a *routesCmd) list(c *cli.Context) error {
	if c.Args().First() == "" {
		return errors.New("error: routes listing takes one argument, an app name")
	}

	if err := resetBasePath(a.Configuration); err != nil {
		return fmt.Errorf("error setting endpoint: %v", err)
	}

	appName := c.Args().Get(0)
	wrapper, _, err := a.AppsAppRoutesGet(appName)
	if err != nil {
		return fmt.Errorf("error getting routes: %v", err)
	}

	if msg := wrapper.Error_.Message; msg != "" {
		return errors.New(msg)
	}

	baseURL, err := url.Parse(a.Configuration.BasePath)
	if err != nil {
		return fmt.Errorf("error parsing base path: %v", err)
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 8, 0, '\t', 0)
	fmt.Fprint(w, "path", "\t", "image", "\t", "endpoint", "\n")
	for _, route := range wrapper.Routes {
		u, err := url.Parse("../")
		u.Path = path.Join(u.Path, "r", appName, route.Path)
		if err != nil {
			return fmt.Errorf("error parsing functions route path: %v", err)
		}

		fmt.Fprint(w, route.Path, "\t", route.Image, "\t", baseURL.ResolveReference(u).String(), "\n")
	}
	w.Flush()

	return nil
}

func (a *routesCmd) call(c *cli.Context) error {
	if c.Args().Get(0) == "" || c.Args().Get(1) == "" {
		return errors.New("error: routes listing takes three arguments: an app name and a route")
	}

	if err := resetBasePath(a.Configuration); err != nil {
		return fmt.Errorf("error setting endpoint: %v", err)
	}

	appName := c.Args().Get(0)
	route := c.Args().Get(1)

	baseURL, err := url.Parse(a.Configuration.BasePath)
	if err != nil {
		return fmt.Errorf("error parsing base path: %v", err)
	}

	u, err := url.Parse("../")
	u.Path = path.Join(u.Path, "r", appName, route)
	content := stdin()

	return callfn(baseURL.ResolveReference(u).String(), content, os.Stdout, c.StringSlice("e"))
}

func callfn(u string, content io.Reader, output io.Writer, env []string) error {
	req, err := http.NewRequest("POST", u, content)
	if err != nil {
		return fmt.Errorf("error running route: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")

	if len(env) > 0 {
		envAsHeader(req, env)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("error running route: %v", err)
	}

	io.Copy(output, resp.Body)
	return nil
}

func envAsHeader(req *http.Request, selectedEnv []string) {
	detectedEnv := os.Environ()
	if len(selectedEnv) > 0 {
		detectedEnv = selectedEnv
	}

	for _, e := range detectedEnv {
		kv := strings.Split(e, "=")
		name := kv[0]
		req.Header.Set(name, os.Getenv(name))
	}
}

func (a *routesCmd) create(c *cli.Context) error {
	if c.Args().Get(0) == "" {
		return errors.New("error: routes creation takes at least one argument: an app name")
	}

	if err := resetBasePath(a.Configuration); err != nil {
		return fmt.Errorf("error setting endpoint: %v", err)
	}

	appName := c.Args().Get(0)
	route := c.Args().Get(1)
	image := c.Args().Get(2)
	var (
		format  string
		maxC    int
		timeout time.Duration
	)
	if image == "" {
		ff, err := loadFuncfile()
		if err != nil {
			if _, ok := err.(*notFoundError); ok {
				return errors.New("error: image name is missing or no function file found")
			} else {
				return err
			}
		}
		image = ff.FullName()
		if ff.Format != nil {
			format = *ff.Format
		}
		if ff.MaxConcurrency != nil {
			maxC = *ff.MaxConcurrency
		}
		if ff.Timeout != nil {
			timeout = *ff.Timeout
		}
		if route == "" && ff.Path != nil {
			route = *ff.Path
		}
	}

	if route == "" {
		return errors.New("error: route path is missing")
	}
	if image == "" {
		return errors.New("error: function image name is missing")
	}

	if f := c.String("format"); f != "" {
		format = f
	}
	if m := c.Int("max-concurrency"); m > 0 {
		maxC = m
	}
	if t := c.Duration("timeout"); t > 0 {
		timeout = t
	}

	body := functions.RouteWrapper{
		Route: functions.Route{
			Path:           route,
			Image:          image,
			Memory:         c.Int64("memory"),
			Type_:          c.String("type"),
			Config:         extractEnvConfig(c.StringSlice("config")),
			Format:         format,
			MaxConcurrency: int32(maxC),
			Timeout:        int32(timeout.Seconds()),
		},
	}

	wrapper, _, err := a.AppsAppRoutesPost(appName, body)
	if err != nil {
		return fmt.Errorf("error creating route: %v", err)
	}

	if msg := wrapper.Error_.Message; msg != "" {
		return errors.New(msg)
	}

	fmt.Println(wrapper.Route.Path, "created with", wrapper.Route.Image)
	return nil
}

func (a *routesCmd) delete(c *cli.Context) error {
	if c.Args().Get(0) == "" || c.Args().Get(1) == "" {
		return errors.New("error: routes listing takes three arguments: an app name and a path")
	}

	if err := resetBasePath(a.Configuration); err != nil {
		return fmt.Errorf("error setting endpoint: %v", err)
	}

	appName := c.Args().Get(0)
	route := c.Args().Get(1)

	resp, err := a.AppsAppRoutesRouteDelete(appName, route)
	if err != nil {
		return fmt.Errorf("error deleting route: %v", err)
	}

	if resp.StatusCode >= http.StatusBadRequest {
		return fmt.Errorf("route not found: %s", route)
	}

	fmt.Println(route, "deleted")
	return nil
}

func (a *routesCmd) configList(c *cli.Context) error {
	if c.Args().Get(0) == "" || c.Args().Get(1) == "" {
		return errors.New("error: route configuration description takes two arguments: an app name and a route")
	}

	if err := resetBasePath(a.Configuration); err != nil {
		return fmt.Errorf("error setting endpoint: %v", err)
	}

	appName := c.Args().Get(0)
	route := c.Args().Get(1)
	wrapper, _, err := a.AppsAppRoutesRouteGet(appName, route)
	if err != nil {
		return fmt.Errorf("error loading route information: %v", err)
	}

	if msg := wrapper.Error_.Message; msg != "" {
		return errors.New(msg)
	}

	config := wrapper.Route.Config
	if len(config) == 0 {
		return errors.New("this route has no configurations")
	}

	if c.Bool("json") {
		if err := json.NewEncoder(os.Stdout).Encode(config); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	} else if c.Bool("shell") {
		for k, v := range config {
			fmt.Print("export ", k, "=", v, "\n")
		}
	} else {
		fmt.Println(appName, wrapper.Route.Path, "configuration:")
		w := tabwriter.NewWriter(os.Stdout, 0, 8, 1, ' ', 0)
		for k, v := range config {
			fmt.Fprint(w, k, ":\t", v, "\n")
		}
		w.Flush()
	}
	return nil
}

func (a *routesCmd) patchRoute(appName, routePath string, r *functions.Route) error {
	wrapper, _, err := a.AppsAppRoutesRouteGet(appName, routePath)
	if err != nil {
		return fmt.Errorf("error loading route: %v", err)
	}

	if msg := wrapper.Error_.Message; msg != "" {
		return errors.New(msg)
	}

	wrapper.Route.Path = ""
	if r != nil {
		if r.Config != nil {
			for k, v := range r.Config {
				if v == "" {
					delete(r.Config, k)
					continue
				}
				wrapper.Route.Config[k] = v
			}
		}
		if r.Headers != nil {
			for k, v := range r.Headers {
				if v[0] == "" {
					delete(r.Headers, k)
					continue
				}
				wrapper.Route.Headers[k] = v
			}
		}
		if r.Image != "" {
			wrapper.Route.Image = r.Image
		}
		if r.Format != "" {
			wrapper.Route.Format = r.Format
		}
		if r.MaxConcurrency > 0 {
			wrapper.Route.MaxConcurrency = r.MaxConcurrency
		}
		if r.Memory > 0 {
			wrapper.Route.Memory = r.Memory
		}
		if r.Timeout > 0 {
			wrapper.Route.Timeout = r.Timeout
		}
	}

	if wrapper, _, err = a.AppsAppRoutesRoutePatch(appName, routePath, *wrapper); err != nil {
		return fmt.Errorf("error updating route configuration: %v", err)
	}

	if msg := wrapper.Error_.Message; msg != "" {
		return errors.New(msg)
	}

	return nil
}

func (a *routesCmd) configSet(c *cli.Context) error {
	if c.Args().Get(0) == "" || c.Args().Get(1) == "" || c.Args().Get(2) == "" {
		return errors.New("error: route configuration setting takes four arguments: an app name, a route, a key and a value")
	}

	if err := resetBasePath(a.Configuration); err != nil {
		return fmt.Errorf("error setting endpoint: %v", err)
	}

	appName := c.Args().Get(0)
	route := c.Args().Get(1)
	key := c.Args().Get(2)
	value := c.Args().Get(3)

	config := map[string]string{}
	config[key] = value

	err := a.patchRoute(appName, route, &functions.Route{
		Config: config,
	})
	if err != nil {
		return err
	}

	fmt.Println(appName, route, "updated", key, "with", value)

	return nil
}

func (a *routesCmd) configUnset(c *cli.Context) error {
	if c.Args().Get(0) == "" || c.Args().Get(1) == "" || c.Args().Get(2) == "" {
		return errors.New("error: route configuration setting takes four arguments: an app name, a route and a key")
	}

	if err := resetBasePath(a.Configuration); err != nil {
		return fmt.Errorf("error setting endpoint: %v", err)
	}

	appName := c.Args().Get(0)
	route := c.Args().Get(1)
	key := c.Args().Get(2)

	config := map[string]string{}
	config[key] = ""

	err := a.patchRoute(appName, route, &functions.Route{
		Config: config,
	})
	if err != nil {
		return err
	}

	fmt.Println(appName, route, "removed", key)
	return nil
}

func (a *routesCmd) imageGetSet(c *cli.Context) error {
	if c.Args().Get(0) == "" || c.Args().Get(1) == "" {
		return errors.New("error: missing two required arguments: an app name and a route path")
	}

	if err := resetBasePath(a.Configuration); err != nil {
		return fmt.Errorf("error setting endpoint: %v", err)
	}

	appName := c.Args().Get(0)
	route := c.Args().Get(1)
	newimage := c.Args().Get(2)

	if newimage == "" {
		wrapper, _, err := a.AppsAppRoutesRouteGet(appName, route)
		if err != nil {
			return fmt.Errorf("error loading route: %v", err)
		}

		if msg := wrapper.Error_.Message; msg != "" {
			return errors.New(msg)
		}

		fmt.Printf("`%s%s` image: %s \n", appName, route, wrapper.Route.Image)
		return nil
	}

	err := a.patchRoute(appName, route, &functions.Route{
		Image: newimage,
	})
	if err != nil {
		return err
	}

	fmt.Printf("`%s%s` image set to %s \n", appName, route, newimage)
	return nil
}

func (a *routesCmd) formatGetSet(c *cli.Context) error {
	if c.Args().Get(0) == "" || c.Args().Get(1) == "" {
		return errors.New("error: missing two required arguments: an app name and a route path")
	}

	if err := resetBasePath(a.Configuration); err != nil {
		return fmt.Errorf("error setting endpoint: %v", err)
	}

	appName := c.Args().Get(0)
	route := c.Args().Get(1)
	newformat := c.Args().Get(2)

	if newformat == "" {
		wrapper, _, err := a.AppsAppRoutesRouteGet(appName, route)
		if err != nil {
			return fmt.Errorf("error loading route: %v", err)
		}

		if msg := wrapper.Error_.Message; msg != "" {
			return errors.New(msg)
		}

		fmt.Printf("`%s%s` format: %s \n", appName, route, wrapper.Route.Format)
		return nil
	}

	err := a.patchRoute(appName, route, &functions.Route{
		Format: newformat,
	})
	if err != nil {
		return err
	}

	fmt.Printf("`%s%s` format set to %s \n", appName, route, newformat)
	return nil
}

func (a *routesCmd) typeGetSet(c *cli.Context) error {
	if c.Args().Get(0) == "" || c.Args().Get(1) == "" {
		return errors.New("error: missing two required arguments: an app name and a route path")
	}

	if err := resetBasePath(a.Configuration); err != nil {
		return fmt.Errorf("error setting endpoint: %v", err)
	}

	appName := c.Args().Get(0)
	route := c.Args().Get(1)
	newtype := c.Args().Get(2)

	if newtype == "" {
		wrapper, _, err := a.AppsAppRoutesRouteGet(appName, route)
		if err != nil {
			return fmt.Errorf("error loading route: %v", err)
		}

		if msg := wrapper.Error_.Message; msg != "" {
			return errors.New(msg)
		}

		fmt.Printf("`%s%s` type: %s \n", appName, route, wrapper.Route.Type_)
		return nil
	}

	err := a.patchRoute(appName, route, &functions.Route{
		Type_: newtype,
	})
	if err != nil {
		return err
	}

	fmt.Printf("`%s%s` type set to %s \n", appName, route, newtype)
	return nil
}

func (a *routesCmd) maxConcurrencyGetSet(c *cli.Context) error {
	if c.Args().Get(0) == "" || c.Args().Get(1) == "" {
		return errors.New("error: missing two required arguments: an app name and a route path")
	}

	if err := resetBasePath(a.Configuration); err != nil {
		return fmt.Errorf("error setting endpoint: %v", err)
	}

	appName := c.Args().Get(0)
	route := c.Args().Get(1)
	maxstr := c.Args().Get(2)

	newmax, _ := strconv.ParseInt(maxstr, 10, 64)

	if newmax > 0 {
		wrapper, _, err := a.AppsAppRoutesRouteGet(appName, route)
		if err != nil {
			return fmt.Errorf("error loading route: %v", err)
		}

		if msg := wrapper.Error_.Message; msg != "" {
			return errors.New(msg)
		}

		fmt.Printf("`%s%s` max concurrency: %d \n", appName, route, wrapper.Route.MaxConcurrency)
		return nil
	}

	err := a.patchRoute(appName, route, &functions.Route{
		MaxConcurrency: int32(newmax),
	})
	if err != nil {
		return err
	}

	fmt.Printf("`%s%s` max set to %d \n", appName, route, newmax)
	return nil
}

func (a *routesCmd) timeoutGetSet(c *cli.Context) error {
	if c.Args().Get(0) == "" || c.Args().Get(1) == "" {
		return errors.New("error: missing two required arguments: an app name and a route path")
	}

	if err := resetBasePath(a.Configuration); err != nil {
		return fmt.Errorf("error setting endpoint: %v", err)
	}

	appName := c.Args().Get(0)
	route := c.Args().Get(1)
	timeoutstr := c.Args().Get(2)

	newtimeout, _ := strconv.ParseInt(timeoutstr, 10, 64)

	if newtimeout > 0 {
		wrapper, _, err := a.AppsAppRoutesRouteGet(appName, route)
		if err != nil {
			return fmt.Errorf("error loading route: %v", err)
		}

		if msg := wrapper.Error_.Message; msg != "" {
			return errors.New(msg)
		}

		fmt.Printf("`%s%s` timeout: %d \n", appName, route, wrapper.Route.Timeout)
		return nil
	}

	err := a.patchRoute(appName, route, &functions.Route{
		Timeout: int32(newtimeout),
	})
	if err != nil {
		return err
	}

	fmt.Printf("`%s%s` format set to %d \n", appName, route, newtimeout)
	return nil
}

func (a *routesCmd) headersList(c *cli.Context) error {
	if c.Args().Get(0) == "" || c.Args().Get(1) == "" {
		return errors.New("error: route configuration description takes two arguments: an app name and a route")
	}

	if err := resetBasePath(a.Configuration); err != nil {
		return fmt.Errorf("error setting endpoint: %v", err)
	}

	appName := c.Args().Get(0)
	route := c.Args().Get(1)
	wrapper, _, err := a.AppsAppRoutesRouteGet(appName, route)
	if err != nil {
		return fmt.Errorf("error loading route information: %v", err)
	}

	if msg := wrapper.Error_.Message; msg != "" {
		return errors.New(msg)
	}

	headers := wrapper.Route.Headers
	if len(headers) == 0 {
		return errors.New("this route has no headers")
	}

	fmt.Println(appName, wrapper.Route.Path, "headers:")
	w := tabwriter.NewWriter(os.Stdout, 0, 8, 1, ' ', 0)
	for k, v := range headers {
		fmt.Fprint(w, k, ":\t", v, "\n")
	}
	w.Flush()
	return nil
}

func (a *routesCmd) headersSet(c *cli.Context) error {
	if c.Args().Get(0) == "" || c.Args().Get(1) == "" || c.Args().Get(2) == "" {
		return errors.New("error: route configuration setting takes four arguments: an app name, a route, a key and a value")
	}

	if err := resetBasePath(a.Configuration); err != nil {
		return fmt.Errorf("error setting endpoint: %v", err)
	}

	appName := c.Args().Get(0)
	route := c.Args().Get(1)
	key := c.Args().Get(2)
	value := c.Args().Get(3)

	headers := map[string][]string{}
	headers[key] = strings.Split(value, ";")
	for k, v := range headers[key] {
		headers[key][k] = strings.TrimSpace(v)
	}

	err := a.patchRoute(appName, route, &functions.Route{
		Headers: headers,
	})
	if err != nil {
		return err
	}

	fmt.Println(appName, route, "headers updated", key, "with", value)
	return nil
}

func (a *routesCmd) headersUnset(c *cli.Context) error {
	if c.Args().Get(0) == "" || c.Args().Get(1) == "" || c.Args().Get(2) == "" {
		return errors.New("error: route configuration setting takes four arguments: an app name, a route and a key")
	}

	if err := resetBasePath(a.Configuration); err != nil {
		return fmt.Errorf("error setting endpoint: %v", err)
	}

	appName := c.Args().Get(0)
	route := c.Args().Get(1)
	key := c.Args().Get(2)

	headers := map[string][]string{}
	headers[key] = []string{""}

	err := a.patchRoute(appName, route, &functions.Route{
		Headers: headers,
	})
	if err != nil {
		return err
	}

	fmt.Println(appName, route, "removed header", key)
	return nil
}
