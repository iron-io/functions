package commands

import (
	"fmt"
	"os"

	"github.com/iron-io/functions/fn/common"
	"github.com/urfave/cli"
)

func Build() cli.Command {
	cmd := Buildcmd{}
	flags := append([]cli.Flag{}, cmd.flags()...)
	return cli.Command{
		Name:   "build",
		Usage:  "build function version",
		Flags:  flags,
		Action: cmd.Build,
	}
}

type Buildcmd struct {
	Verbose bool
}

func (b *Buildcmd) flags() []cli.Flag {
	return []cli.Flag{
		cli.BoolFlag{
			Name:        "v",
			Usage:       "verbose mode",
			Destination: &b.Verbose,
		},
	}
}

// build will take the found valid function and build it
func (b *Buildcmd) Build(c *cli.Context) error {
	verbwriter := common.Verbwriter(b.Verbose)

	path, err := os.Getwd()
	if err != nil {
		return err
	}
	fn, err := common.FindFuncfile(path)
	if err != nil {
		return err
	}

	fmt.Fprintln(verbwriter, "building", fn)
	ff, err := common.Buildfunc(verbwriter, fn)
	if err != nil {
		return err
	}

	fmt.Printf("Function %v built successfully.\n", ff.FullName())
	return nil
}
