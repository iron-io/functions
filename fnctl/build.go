package main

import (
	"fmt"
	"io"
	"os"

	"github.com/urfave/cli"
)

func build() cli.Command {
	cmd := buildcmd{}
	var flags []cli.Flag
	flags = append(flags, cmd.flags()...)
	return cli.Command{
		Name:   "build",
		Usage:  "build function version",
		Flags:  flags,
		Action: cmd.scan,
	}
}

type buildcmd struct {
	wd      string
	verbose bool
}

func (u *buildcmd) flags() []cli.Flag {
	return []cli.Flag{
		cli.StringFlag{
			Name:        "d",
			Usage:       "working directory",
			Destination: &u.wd,
			EnvVar:      "WORK_DIR",
			Value:       "./",
		},
		cli.BoolFlag{
			Name:        "v",
			Usage:       "verbose mode",
			Destination: &u.verbose,
		},
	}
}

func (u *buildcmd) scan(c *cli.Context) error {
	scan(u.verbose, u.wd, u.walker)
	return nil
}

func (u *buildcmd) walker(path string, info os.FileInfo, err error, w io.Writer) error {
	if !isvalid(path, info) {
		return nil
	}

	fmt.Fprint(w, path, "\t")
	if err := u.build(path); err != nil {
		fmt.Fprintln(w, err)
	} else {
		fmt.Fprintln(w, "built")
	}

	return nil
}

// build will take the found valid function and build it
func (u *buildcmd) build(path string) error {
	fmt.Fprintln(verbwriter, "building", path)
	_, err := buildFunc(path)
	return err
}
