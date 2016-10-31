package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"text/tabwriter"

	functions "github.com/iron-io/functions_go"
	"github.com/urfave/cli"
)

func build() cli.Command {
	cmd := buildcmd{RoutesApi: functions.NewRoutesApi()}
	var flags []cli.Flag
	flags = append(flags, cmd.flags()...)
	flags = append(flags, confFlags(&cmd.Configuration)...)
	return cli.Command{
		Name:   "build",
		Usage:  "build function version",
		Flags:  flags,
		Action: cmd.scan,
	}
}

type buildcmd struct {
	*functions.RoutesApi

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
	if u.verbose {
		verbwriter = os.Stderr
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 8, 0, '\t', 0)
	fmt.Fprint(w, "path", "\t", "action", "\n")

	path := u.wd
	if !filepath.IsAbs(path) {
		cwd, _ := os.Getwd()
		path = filepath.Join(cwd, path)
	}
	os.Chdir(path)

	filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		return u.walker(path, info, err, w)
	})

	w.Flush()
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
		fmt.Fprintln(w, "build done")
	}

	return nil
}

// build will take the found valid function and build it
func (u *buildcmd) build(path string) error {
	fmt.Fprintln(verbwriter, "building", path)

	dir := filepath.Dir(path)
	dockerFile := filepath.Join(dir, "Dockerfile")
	if _, err := os.Stat(dockerFile); os.IsNotExist(err) {
		return errDockerFileNotFound
	}

	funcfile, err := parseFuncFile(path)
	if err != nil {
		return err
	}

	if funcfile.Build != nil {
		if err := localbuild(path, funcfile.Build); err != nil {
			return err
		}
	}

	if err := dockerbuild(path, funcfile.Image, true); err != nil {
		return err
	}

	return nil
}
