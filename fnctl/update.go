package main

import (
	"fmt"
	"os"

	"github.com/urfave/cli"
)

func update() cli.Command {
	return cli.Command{
		Name:      "update",
		Usage:     "scan local directory for automatic functions update",
		ArgsUsage: "fnclt update",
		Action: func(c *cli.Context) error {
			fmt.Fprintln(os.Stderr, "update not implemented")
			os.Exit(1)
			return nil
		},
	}
}
