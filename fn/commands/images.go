package commands

import (
	image_commands "github.com/iron-io/functions/fn/commands/images"
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
			image_commands.Build(),
			image_commands.Deploy(),
			image_commands.Bump(),
			Call(),
			image_commands.Push(),
			image_commands.Run(),
			testfn(),
		},
	}
}
