package common

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	bumper "github.com/giantswarm/semver-bump/bump"
	"github.com/giantswarm/semver-bump/storage"
	yaml "gopkg.in/yaml.v2"
)

var (
	Validfn = [...]string{
		"func.yaml",
		"func.yml",
		"func.json",
	}

	errUnexpectedFileFormat = errors.New("unexpected file format for function file")
)

type fftest struct {
	Name string            `yaml:"name,omitempty" json:"name,omitempty"`
	In   *string           `yaml:"in,omitempty" json:"in,omitempty"`
	Out  *string           `yaml:"out,omitempty" json:"out,omitempty"`
	Err  *string           `yaml:"err,omitempty" json:"err,omitempty"`
	Env  map[string]string `yaml:"env,omitempty" json:"env,omitempty"`
}

type Funcfile struct {
	Name           string            `yaml:"name,omitempty" json:"name,omitempty"`
	Version        string            `yaml:"version,omitempty" json:"version,omitempty"`
	Runtime        *string           `yaml:"runtime,omitempty" json:"runtime,omitempty"`
	Entrypoint     string            `yaml:"entrypoint,omitempty" json:"entrypoint,omitempty"`
	Cmd            string            `yaml:"cmd,omitempty" json:"cmd,omitempty"`
	Type           *string           `yaml:"type,omitempty" json:"type,omitempty"`
	Memory         *int64            `yaml:"memory,omitempty" json:"memory,omitempty"`
	Format         *string           `yaml:"format,omitempty" json:"format,omitempty"`
	Timeout        *time.Duration    `yaml:"timeout,omitempty" json:"timeout,omitempty"`
	IDLETimeout    *time.Duration    `yaml:"idle_timeout,omitempty" json:"idle_timeout,omitempty"`
	Headers        map[string]string `yaml:"headers,omitempty" json:"headers,omitempty"`
	Config         map[string]string `yaml:"config,omitempty" json:"config,omitempty"`
	Build          []string          `yaml:"build,omitempty" json:"build,omitempty"`
	Tests          []fftest          `yaml:"tests,omitempty" json:"tests,omitempty"`
	Path           *string           `yaml:"path,omitempty" json:"path,omitempty"`
	MaxConcurrency *int              `yaml:"max_concurrency,omitempty" json:"max_concurrency,omitempty"`
	JwtKey         *string           `yaml:"jwt_key,omitempty" json:"jwt_key,omitempty"`
}

func (ff *Funcfile) FullName() string {
	fname := ff.Name
	if ff.Version != "" {
		fname = fmt.Sprintf("%s:%s", fname, ff.Version)
	}
	return fname
}

func (ff *Funcfile) RuntimeTag() (runtime, tag string) {
	if ff.Runtime == nil {
		return "", ""
	}

	rt := *ff.Runtime
	tagpos := strings.Index(rt, ":")
	if tagpos == -1 {
		return rt, ""
	}

	return rt[:tagpos], rt[tagpos+1:]
}

func cleanImageName(name string) string {
	if i := strings.Index(name, ":"); i != -1 {
		name = name[:i]
	}

	return name
}

func (ff *Funcfile) Bumpversion() error {
	ff.Name = cleanImageName(ff.Name)
	if ff.Version == "" {
		ff.Version = INITIAL_VERSION
		return nil
	}

	s, err := storage.NewVersionStorage("local", ff.Version)
	if err != nil {
		return err
	}

	version := bumper.NewSemverBumper(s, "")
	newver, err := version.BumpPatchVersion("", "")
	if err != nil {
		return err
	}

	ff.Version = newver.String()
	return nil
}

func FindFuncfile(path string) (string, error) {
	for _, fn := range Validfn {
		fullfn := filepath.Join(path, fn)
		if Exists(fullfn) {
			return fullfn, nil
		}
	}
	return "", newNotFoundError("could not find function file")
}

func LoadFuncfile() (*Funcfile, error) {
	fn, err := FindFuncfile(".")
	if err != nil {
		return nil, err
	}
	return ParseFuncfile(fn)
}

func ParseFuncfile(path string) (*Funcfile, error) {
	ext := filepath.Ext(path)
	switch ext {
	case ".json":
		return decodeFuncfileJSON(path)
	case ".yaml", ".yml":
		return decodeFuncfileYAML(path)
	}
	return nil, errUnexpectedFileFormat
}

func StoreFuncfile(path string, ff *Funcfile) error {
	ext := filepath.Ext(path)
	switch ext {
	case ".json":
		return encodeFuncfileJSON(path, ff)
	case ".yaml", ".yml":
		return EncodeFuncfileYAML(path, ff)
	}
	return errUnexpectedFileFormat
}

func decodeFuncfileJSON(path string) (*Funcfile, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("could not open %s for parsing. Error: %v", path, err)
	}
	ff := new(Funcfile)
	err = json.NewDecoder(f).Decode(ff)
	return ff, err
}

func decodeFuncfileYAML(path string) (*Funcfile, error) {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("could not open %s for parsing. Error: %v", path, err)
	}
	ff := new(Funcfile)
	err = yaml.Unmarshal(b, ff)
	return ff, err
}

func encodeFuncfileJSON(path string, ff *Funcfile) error {
	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("could not open %s for encoding. Error: %v", path, err)
	}
	return json.NewEncoder(f).Encode(ff)
}

func EncodeFuncfileYAML(path string, ff *Funcfile) error {
	b, err := yaml.Marshal(ff)
	if err != nil {
		return fmt.Errorf("could not encode function file. Error: %v", err)
	}
	return ioutil.WriteFile(path, b, os.FileMode(0644))
}
