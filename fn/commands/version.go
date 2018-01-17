package commands

import (
	"fmt"
	vers "github.com/iron-io/functions/api/version"
	"github.com/iron-io/functions/fn/common"
	functions "github.com/iron-io/functions_go"
	"github.com/urfave/cli"
)

func Version() cli.Command {
	r := versionCmd{VersionApi: functions.NewVersionApi()}
	return cli.Command{
		Name:   "version",
		Usage:  "displays fn and functions daemon versions",
		Action: r.version,
	}
}

type versionCmd struct {
	*functions.VersionApi
}

func (r *versionCmd) version(c *cli.Context) error {
	r.Configuration.BasePath = common.GetBasePath("")

	fmt.Println("Client version:", vers.Version)
	v, _, err := r.VersionGet()
	if err != nil {
		return err
	}
	fmt.Println("Server version", v.Version)
	return nil
}
