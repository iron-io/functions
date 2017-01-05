package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"

	"github.com/iron-io/functions_go"
	"github.com/jmoiron/jsonq"
	"github.com/urfave/cli"
	"strings"
)

type appsCmd struct {
	*functions.AppsApi
}

func apps() cli.Command {
	a := appsCmd{AppsApi: functions.NewAppsApi()}

	return cli.Command{
		Name:      "apps",
		Usage:     "operate applications",
		ArgsUsage: "fn apps",
		Subcommands: []cli.Command{
			{
				Name:      "create",
				Aliases:   []string{"c"},
				Usage:     "create a new app",
				ArgsUsage: "`app`",
				Action:    a.create,
				Flags: []cli.Flag{
					cli.StringSliceFlag{
						Name:  "config",
						Usage: "application configuration",
					},
				},
			},
			{
				Name:      "inspect",
				Aliases:   []string{"i"},
				Usage:     "retrieve one or all apps properties",
				ArgsUsage: "`app` [property.[key]]",
				Action:    a.inspect,
			},
			{
				Name:      "update",
				Aliases:   []string{"u"},
				Usage:     "update an `app`",
				ArgsUsage: "`app`",
				Action:    a.update,
				Flags: []cli.Flag{
					cli.StringSliceFlag{
						Name:  "config,c",
						Usage: "route configuration",
					},
				},
			},
			{
				Name:    "list",
				Aliases: []string{"l"},
				Usage:   "list all apps",
				Action:  a.list,
			},
			{
				Name:   "delete",
				Usage:  "delete an app",
				Action: a.delete,
			},
		},
	}
}

func (a *appsCmd) list(c *cli.Context) error {
	if err := resetBasePath(a.Configuration); err != nil {
		return fmt.Errorf("error setting endpoint: %v", err)
	}

	wrapper, _, err := a.AppsGet()
	if err != nil {
		return fmt.Errorf("error getting app: %v", err)
	}

	if msg := wrapper.Error_.Message; msg != "" {
		return errors.New(msg)
	}

	if len(wrapper.Apps) == 0 {
		fmt.Println("no apps found")
		return nil
	}

	for _, app := range wrapper.Apps {
		fmt.Println(app.Name)
	}

	return nil
}

func (a *appsCmd) create(c *cli.Context) error {
	if c.Args().First() == "" {
		return errors.New("error: missing app name after create command")
	}

	if err := resetBasePath(a.Configuration); err != nil {
		return fmt.Errorf("error setting endpoint: %v", err)
	}

	body := functions.AppWrapper{App: functions.App{
		Name:   c.Args().Get(0),
		Config: extractEnvConfig(c.StringSlice("config")),
	}}
	wrapper, _, err := a.AppsPost(body)
	if err != nil {
		return fmt.Errorf("error creating app: %v", err)
	}

	if msg := wrapper.Error_.Message; msg != "" {
		return errors.New(msg)
	}

	fmt.Println("app", wrapper.App.Name, "created")
	return nil
}

func (a *appsCmd) update(c *cli.Context) error {
	if c.Args().First() == "" {
		return errors.New("error: missing app name after update command")
	}

	if err := resetBasePath(a.Configuration); err != nil {
		return fmt.Errorf("error setting endpoint: %v", err)
	}

	appName := c.Args().First()

	patchedApp := &functions.App{
		Config: extractEnvConfig(c.StringSlice("config")),
	}

	err := a.patchApp(appName, patchedApp)
	if err != nil {
		return err
	}

	fmt.Println("app", appName, "updated")
	return nil
}

func (a *appsCmd) patchApp(appName string, app *functions.App) error {
	wrapper, _, err := a.AppsAppGet(appName)
	if err != nil {
		return fmt.Errorf("error loading app: %v", err)
	}

	if msg := wrapper.Error_.Message; msg != "" {
		return errors.New(msg)
	}

	wrapper.App.Name = ""
	if app != nil {
		if app.Config != nil {
			for k, v := range app.Config {
				if v == "" {
					delete(wrapper.App.Config, k)
					continue
				}
				wrapper.App.Config[k] = v
			}
		}
	}

	if wrapper, _, err = a.AppsAppPatch(appName, *wrapper); err != nil {
		return fmt.Errorf("error updating app: %v", err)
	}

	if msg := wrapper.Error_.Message; msg != "" {
		return errors.New(msg)
	}

	return nil
}

func (a *appsCmd) inspect(c *cli.Context) error {
	if c.Args().Get(0) == "" {
		return errors.New("error: missing app name after the inspect command")
	}

	if err := resetBasePath(a.Configuration); err != nil {
		return fmt.Errorf("error setting endpoint: %v", err)
	}

	appName := c.Args().First()
	prop := c.Args().Get(1)

	wrapper, resp, err := a.AppsAppGet(appName)
	if err != nil {
		return fmt.Errorf("error retrieving app: %v", err)
	}

	if msg := wrapper.Error_.Message; msg != "" {
		return errors.New(msg)
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "\t")

	if prop == "" {
		enc.Encode(wrapper.App)
		return nil
	}

	var inspect struct{ App map[string]interface{} }
	err = json.Unmarshal(resp.Payload, &inspect)
	if err != nil {
		return fmt.Errorf("error inspect app: %v", err)
	}

	jq := jsonq.NewQuery(inspect.App)
	field, err := jq.Interface(strings.Split(prop, ".")...)
	if err != nil {
		return errors.New("failed to inspect the property")
	}
	enc.Encode(field)

	return nil
}

func (a *appsCmd) delete(c *cli.Context) error {
	appName := c.Args().First()
	if appName == "" {
		return errors.New("error: deleting an app takes one argument, an app name")
	}

	if err := resetBasePath(a.Configuration); err != nil {
		return fmt.Errorf("error setting endpoint: %v", err)
	}

	resp, err := a.AppsAppDelete(appName)
	if err != nil {
		return fmt.Errorf("error deleting app: %v", err)
	}

	if resp.StatusCode == http.StatusBadRequest {
		return errors.New("could not delete this application - pending routes")
	}

	fmt.Println("app", appName, "deleted")
	return nil
}
