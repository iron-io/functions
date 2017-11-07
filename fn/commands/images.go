package commands

import (
	"github.com/iron-io/functions_go"
	"github.com/urfave/cli"
)

type imagesCmd struct {
	*functions.AppsApi
}

func Images() cli.Command {
	return cli.Command{
		Name:  "images",
		Usage: "manage function images",
		Subcommands: []cli.Command{
			Build(),
			Deploy(),
			Bump(),
			Call(),
			Push(),
			Run(),
			testfn(),
		},
	}
}
