package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"strings"

	"github.com/urfave/cli"
)

type initFnCmd struct {
	force bool
}

func initFn() cli.Command {
	a := initFnCmd{}

	return cli.Command{
		Name:      "init",
		Usage:     "create a local function.yaml file",
		ArgsUsage: "fnctl init <entrypoint>",
		Action:    a.init,
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:        "f",
				Usage:       "overwrite existing function.yaml",
				Destination: &a.force,
			},
		},
	}
}

func (a *initFnCmd) init(c *cli.Context) error {
	if !a.force {
		for _, fn := range validfn {
			if _, err := os.Stat(fn); !os.IsNotExist(err) {
				return errors.New("function file already exists")
			}
		}
	}

	entrypoint := c.Args().First()
	if entrypoint == "" {
		return errors.New("entrypoint is missing")
	}

	scores := map[string]uint{
		"": 0,
	}
	pwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("error detecting current working directory: %s\n", err)
	}
	err = filepath.Walk(pwd, func(_ string, info os.FileInfo, _ error) error {
		if info.IsDir() {
			return nil
		}

		ext := strings.ToLower(filepath.Ext(info.Name()))
		if ext == "" {
			return nil
		}
		scores[ext]++
		return nil
	})
	if err != nil {
		return fmt.Errorf("file walk error: %s\n", err)
	}

	biggest := ""
	for ext, score := range scores {
		if score > scores[biggest] {
			biggest = ext
		}
	}

	runtime, ok := fileExtToRuntime[biggest]
	if !ok {
		return fmt.Errorf("could not detect language of this function: %s\n", biggest)
	}
	ff := &funcfile{
		Image: runtime,
		// Runtime:   runtime,
		Version: initialVersion,
		// Entrypoint: c.Args().First(),
	}

	encodeFuncfileYAML("function.yaml", ff)
	return nil
}

var fileExtToRuntime = map[string]string{
	".c":     "gcc",
	".class": "java",
	".clj":   "leiningen",
	".cpp":   "gcc",
	".erl":   "erlang",
	".ex":    "elixir",
	".go":    "go",
	".h":     "gcc",
	".java":  "java",
	".js":    "node",
	".php":   "php",
	".pl":    "perl",
	".py":    "python",
	".scala": "scala",
}
