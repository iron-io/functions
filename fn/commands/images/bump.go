package commands

import (
	"fmt"
	"github.com/iron-io/functions/fn/common"
	"github.com/urfave/cli"
	"os"
)

var (
	initialVersion = common.INITIAL_VERSION
)

func Bump() cli.Command {
	cmd := bumpcmd{}
	flags := append([]cli.Flag{}, cmd.flags()...)
	return cli.Command{
		Name:   "bump",
		Usage:  "bump function version",
		Flags:  flags,
		Action: cmd.bump,
	}
}

type bumpcmd struct {
	verbose bool
}

func (b *bumpcmd) flags() []cli.Flag {
	return []cli.Flag{
		cli.BoolFlag{
			Name:        "v",
			Usage:       "verbose mode",
			Destination: &b.verbose,
		},
	}
}

// bump will take the found valid function and bump its version
func (b *bumpcmd) bump(c *cli.Context) error {
	verbwriter := common.Verbwriter(b.verbose)

	path, err := os.Getwd()
	if err != nil {
		return err
	}
	fn, err := common.FindFuncfile(path)
	if err != nil {
		return err
	}

	fmt.Fprintln(verbwriter, "bumping version for", fn)

	funcfile, err := common.ParseFuncfile(fn)
	if err != nil {
		return err
	}

	err = funcfile.Bumpversion()
	if err != nil {
		return err
	}

	if err := common.StoreFuncfile(fn, funcfile); err != nil {
		return err
	}

	fmt.Println("Bumped to version", funcfile.Version)
	return nil
}
