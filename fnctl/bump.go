package main

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	bumper "github.com/giantswarm/semver-bump/bump"
	"github.com/giantswarm/semver-bump/storage"

	functions "github.com/iron-io/functions_go"
	"github.com/urfave/cli"
)

var (
	initialVersion = "0.0.1"

	errVersionFileNotFound = errors.New("no VERSION file found for this function")
)

func bump() cli.Command {
	cmd := bumpcmd{RoutesApi: functions.NewRoutesApi()}
	var flags []cli.Flag
	flags = append(flags, cmd.flags()...)
	flags = append(flags, confFlags(&cmd.Configuration)...)
	return cli.Command{
		Name:   "bump",
		Usage:  "bump function version",
		Flags:  flags,
		Action: cmd.scan,
	}
}

type bumpcmd struct {
	*functions.RoutesApi

	wd      string
	verbose bool
}

func (b *bumpcmd) flags() []cli.Flag {
	return []cli.Flag{
		cli.StringFlag{
			Name:        "d",
			Usage:       "working directory",
			Destination: &b.wd,
			EnvVar:      "WORK_DIR",
			Value:       "./",
		},
		cli.BoolFlag{
			Name:        "v",
			Usage:       "verbose mode",
			Destination: &b.verbose,
		},
	}
}

func (b *bumpcmd) scan(c *cli.Context) error {
	scan(b.verbose, b.wd, b.walker)
	return nil
}

func (b *bumpcmd) walker(path string, info os.FileInfo, err error, w io.Writer) error {
	if !isvalid(path, info) {
		return nil
	}

	fmt.Fprint(w, path, "\t")
	if err := b.bump(path); err != nil {
		fmt.Fprintln(w, err)
	} else {
		fmt.Fprintln(w, "bumped")
	}

	return nil
}

// bump will take the found valid function and bump its version
func (b *bumpcmd) bump(path string) error {
	fmt.Fprintln(verbwriter, "bumping", path)

	dir := filepath.Dir(path)
	versionfile := filepath.Join(dir, "VERSION")
	if _, err := os.Stat(versionfile); os.IsNotExist(err) {
		return errVersionFileNotFound
	}

	s, err := storage.NewVersionStorage("file", initialVersion)
	if err != nil {
		return err
	}

	version := bumper.NewSemverBumper(s, versionfile)
	newver, err := version.BumpPatchVersion("", "")
	if err != nil {
		return err
	}

	ioutil.WriteFile(versionfile, []byte(newver.String()), 0666)

	return nil
}
