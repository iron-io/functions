package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"text/tabwriter"

	yaml "gopkg.in/yaml.v2"
)

var (
	validfn = [...]string{
		"functions.yaml",
		"functions.yml",
		"fn.yaml",
		"fn.yml",
		"functions.json",
		"fn.json",
	}

	errDockerFileNotFound   = errors.New("no Dockerfile found for this function")
	errUnexpectedFileFormat = errors.New("unexpected file format for function file")
	verbwriter              = ioutil.Discard
)

type funcfile struct {
	App   *string
	Image string
	Route *string
	Build []string
}

func parseFuncFile(path string) (*funcfile, error) {
	ext := filepath.Ext(path)
	switch ext {
	case ".json":
		return parseJSON(path)
	case ".yaml", ".yml":
		return parseYAML(path)
	}
	return nil, errUnexpectedFileFormat
}

func parseJSON(path string) (*funcfile, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("could not open %s for parsing. Error: %v", path, err)
	}
	ff := new(funcfile)
	err = json.NewDecoder(f).Decode(ff)
	return ff, err
}

func parseYAML(path string) (*funcfile, error) {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("could not open %s for parsing. Error: %v", path, err)
	}
	ff := new(funcfile)
	err = yaml.Unmarshal(b, ff)
	return ff, err
}

func buildFunc(path string) (*funcfile, error) {
	dir := filepath.Dir(path)
	dockerfile := filepath.Join(dir, "Dockerfile")
	if _, err := os.Stat(dockerfile); os.IsNotExist(err) {
		return nil, errDockerFileNotFound
	}

	funcfile, err := parseFuncFile(path)
	if err != nil {
		return nil, err
	}

	if err := localbuild(path, funcfile.Build); err != nil {
		return nil, err
	}

	if err := dockerbuild(path, funcfile.Image); err != nil {
		return nil, err
	}

	return funcfile, nil
}

func localbuild(path string, steps []string) error {
	for _, cmd := range steps {
		c := exec.Command("/bin/sh", "-c", cmd)
		c.Dir = filepath.Dir(path)
		out, err := c.CombinedOutput()
		fmt.Fprintf(verbwriter, "- %s:\n%s\n", cmd, out)
		if err != nil {
			return fmt.Errorf("error running command %v (%v)", cmd, err)
		}
	}

	return nil
}

func dockerbuild(path, image string) error {
	out, err := exec.Command("docker", "build", "-t", image, filepath.Dir(path)).CombinedOutput()
	fmt.Fprintf(verbwriter, "%s\n", out)
	if err != nil {
		return fmt.Errorf("error running docker build: %v", err)
	}

	return nil
}

func extractAppNameRoute(path string) (appName, route string) {

	// The idea here is to extract the root-most directory name
	// as application name, it turns out that stdlib tools are great to
	// extract the deepest one. Thus, we revert the string and use the
	// stdlib as it is - and revert back to its normal content. Not fastest
	// ever, but it is simple.

	rpath := reverse(path)
	rroute, rappName := filepath.Split(rpath)
	route = filepath.Dir(reverse(rroute))
	return reverse(rappName), route
}

func reverse(s string) string {
	r := []rune(s)
	for i, j := 0, len(r)-1; i < len(r)/2; i, j = i+1, j-1 {
		r[i], r[j] = r[j], r[i]
	}
	return string(r)
}

func scan(verbose bool, wd string, walker func(path string, info os.FileInfo, err error, w io.Writer) error) {
	if verbose {
		verbwriter = os.Stderr
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 8, 0, '\t', 0)
	fmt.Fprint(w, "path", "\t", "action", "\n")

	filepath.Walk(wd, func(path string, info os.FileInfo, err error) error {
		return walker(path, info, err, w)
	})

	w.Flush()
}

func isvalid(path string, info os.FileInfo) bool {
	if info.IsDir() {
		return false
	}

	basefn := filepath.Base(path)
	for _, fn := range validfn {
		if basefn == fn {
			return true
		}
	}

	return false
}
