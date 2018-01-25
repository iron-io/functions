package common

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/iron-io/functions/fn/langs"
	functions "github.com/iron-io/functions_go"
	"github.com/spf13/viper"
)

var (
	API_VERSION     string
	SSL_SKIP_VERIFY bool
	JWT_AUTH_KEY    string
	API_URL         string
	SCHEME          string
	INITIAL_VERSION string
	HOST            string
	BASE_PATH       string
)

func SetEnv() {
	viper.AutomaticEnv()
	API_VERSION = "/v1"
	SSL_SKIP_VERIFY = (os.Getenv("SSL_SKIP_VERIFY") == "true")
	JWT_AUTH_KEY = viper.GetString("jwt_auth_key")
	SCHEME = "http"
	INITIAL_VERSION = "0.0.1"
	viper.SetDefault("API_URL", "http://localhost:8080")
	API_URL = viper.GetString("API_URL")
	BASE_PATH = GetBasePath(API_VERSION)
}

func GetBasePath(version string) string {
	u, err := url.Parse(API_URL)
	if err != nil {
		log.Fatalln("Couldn't parse API URL:", err)
	}
	HOST = u.Host
	SCHEME = u.Scheme
	u.Path = version
	return u.String()
}

func ResetBasePath(c *functions.Configuration) error {
	c.BasePath = BASE_PATH
	return nil
}

func Verbwriter(verbose bool) io.Writer {
	verbwriter := ioutil.Discard
	if verbose {
		verbwriter = os.Stderr
	}
	return verbwriter
}

func Buildfunc(verbwriter io.Writer, fn string) (*Funcfile, error) {
	funcfile, err := ParseFuncfile(fn)
	if err != nil {
		return nil, err
	}

	if funcfile.Version == "" {
		err = funcfile.Bumpversion()
		if err != nil {
			return nil, err
		}
		if err := StoreFuncfile(fn, funcfile); err != nil {
			return nil, err
		}
		funcfile, err = ParseFuncfile(fn)
		if err != nil {
			return nil, err
		}
	}

	if err := localbuild(verbwriter, fn, funcfile.Build); err != nil {
		return nil, err
	}

	if err := dockerbuild(verbwriter, fn, funcfile); err != nil {
		return nil, err
	}

	return funcfile, nil
}

func localbuild(verbwriter io.Writer, path string, steps []string) error {
	for _, cmd := range steps {
		exe := exec.Command("/bin/sh", "-c", cmd)
		exe.Dir = filepath.Dir(path)
		exe.Stderr = verbwriter
		exe.Stdout = verbwriter
		if err := exe.Run(); err != nil {
			return fmt.Errorf("error running command %v (%v)", cmd, err)
		}
	}

	return nil
}

func dockerbuild(verbwriter io.Writer, path string, ff *Funcfile) error {
	dir := filepath.Dir(path)

	var helper langs.LangHelper
	dockerfile := filepath.Join(dir, "Dockerfile")
	if !Exists(dockerfile) {
		err := writeTmpDockerfile(dir, ff)
		defer os.Remove(filepath.Join(dir, "Dockerfile"))
		if err != nil {
			return err
		}
		helper = langs.GetLangHelper(*ff.Runtime)
		if helper == nil {
			return fmt.Errorf("Cannot build, no language helper found for %v", *ff.Runtime)
		}
		if helper.HasPreBuild() {
			err := helper.PreBuild()
			if err != nil {
				return err
			}
		}
	}

	fmt.Printf("Building image %v\n", ff.FullName())
	cmd := exec.Command("docker", "build", "-t", ff.FullName(), ".")
	cmd.Dir = dir
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("error running docker build: %v", err)
	}
	if helper != nil {
		err := helper.AfterBuild()
		if err != nil {
			return err
		}
	}
	return nil
}

func Exists(name string) bool {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

var AcceptableFnRuntimes = map[string]string{
	"elixir":           "iron/elixir",
	"erlang":           "iron/erlang",
	"gcc":              "iron/gcc",
	"go":               "iron/go",
	"java":             "iron/java",
	"leiningen":        "iron/leiningen",
	"mono":             "iron/mono",
	"node":             "iron/node",
	"perl":             "iron/perl",
	"php":              "iron/php",
	"python":           "iron/python:2",
	"ruby":             "iron/ruby",
	"scala":            "iron/scala",
	"rust":             "corey/rust-alpine",
	"dotnet":           "microsoft/dotnet:runtime",
	"lambda-nodejs4.3": "iron/functions-lambda:nodejs4.3",
}

const tplDockerfile = `FROM {{ .BaseImage }}
WORKDIR /function
ADD . /function/
{{ if ne .Entrypoint "" }} ENTRYPOINT [{{ .Entrypoint }}] {{ end }}
{{ if ne .Cmd "" }} CMD [{{ .Cmd }}] {{ end }}
`

func writeTmpDockerfile(dir string, ff *Funcfile) error {
	if ff.Entrypoint == "" && ff.Cmd == "" {
		return errors.New("entrypoint and cmd are missing, you must provide one or the other")
	}

	runtime, tag := ff.RuntimeTag()
	rt, ok := AcceptableFnRuntimes[runtime]
	if !ok {
		return fmt.Errorf("cannot use runtime %s", runtime)
	}

	if tag != "" {
		rt = fmt.Sprintf("%s:%s", rt, tag)
	}

	fd, err := os.Create(filepath.Join(dir, "Dockerfile"))
	if err != nil {
		return err
	}
	defer fd.Close()

	// convert entrypoint string to slice
	bufferEp := stringToSlice(ff.Entrypoint)
	bufferCmd := stringToSlice(ff.Cmd)

	t := template.Must(template.New("Dockerfile").Parse(tplDockerfile))
	err = t.Execute(fd, struct {
		BaseImage, Entrypoint, Cmd string
	}{rt, bufferEp.String(), bufferCmd.String()})

	return err
}

func stringToSlice(in string) bytes.Buffer {
	epvals := strings.Fields(in)
	var buffer bytes.Buffer
	for i, s := range epvals {
		if i > 0 {
			buffer.WriteString(", ")
		}
		buffer.WriteString("\"")
		buffer.WriteString(s)
		buffer.WriteString("\"")
	}
	return buffer
}

func ExtractEnvConfig(configs []string) map[string]string {
	c := make(map[string]string)
	for _, v := range configs {
		kv := strings.SplitN(v, "=", 2)
		if len(kv) == 2 {
			c[kv[0]] = os.ExpandEnv(kv[1])
		}
	}
	return c
}

func Dockerpush(ff *Funcfile) error {
	latestTag := ff.Name + ":latest"
	cmd := exec.Command("docker", "tag", ff.FullName(), latestTag)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("error tagging latest: %v", err)
	}
	cmd = exec.Command("docker", "push", ff.Name)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("error running docker push: %v", err)
	}
	return nil
}

func AppNamePath(img string) (string, string) {
	sep := strings.Index(img, "/")
	if sep < 0 {
		return "", ""
	}
	tag := strings.Index(img[sep:], ":")
	if tag < 0 {
		tag = len(img[sep:])
	}
	return img[:sep], img[sep : sep+tag]
}
