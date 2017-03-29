package runner

import (
	"context"
	"encoding/json"
	"io"
	"os"
	"time"

	docker "github.com/fsouza/go-dockerclient"
	"github.com/iron-io/functions/api/runner/task"
	"github.com/iron-io/runner/drivers"
)

type containerTask struct {
	ctx    context.Context
	cfg    *task.Config
	canRun chan bool
}

func (t *containerTask) Command() string { return "" }

func (t *containerTask) EnvVars() map[string]string {
	return t.cfg.Env
}
func (t *containerTask) Input() io.Reader {
	return t.cfg.Stdin
}

func (t *containerTask) Labels() map[string]string {
	return map[string]string{
		"LogName": t.cfg.AppName,
	}
}

func (t *containerTask) Id() string                         { return t.cfg.ID }
func (t *containerTask) Route() string                      { return "" }
func (t *containerTask) Image() string                      { return t.cfg.Image }
func (t *containerTask) Timeout() time.Duration             { return t.cfg.Timeout }
func (t *containerTask) IdleTimeout() time.Duration   { return t.cfg.IdleTimeout }
func (t *containerTask) Logger() (io.Writer, io.Writer)     { return t.cfg.Stdout, t.cfg.Stderr }
func (t *containerTask) Volumes() [][2]string               { return [][2]string{} }
func (t *containerTask) WorkDir() string                    { return "" }

func (t *containerTask) Close()                 {}
func (t *containerTask) WriteStat(drivers.Stat) {}

// Implementing the docker.AuthConfiguration interface.  Pulling in
// the docker repo password from environment variables
func (t *containerTask) DockerAuth() (docker.AuthConfiguration, error) {
	reg, _, _ := drivers.ParseImage(t.Image())

	// Default to the nil configuration
	authconfig := docker.AuthConfiguration{}

	// Attempt to fetch config from built in environment
	regsettings := t.EnvVars()["DOCKER_REPOSITORY_AUTHS"]

	if regsettings == "" {
		// Attempt to fetch it from an environment variable
		regsettings = os.Getenv("DOCKER_REPOSITORY_AUTHS")
	}

	// If we have settings, unmarshal them
	if regsettings != "" {
		registries := make(dockerRegistries, 0)
		if err := json.Unmarshal([]byte(regsettings), &registries); err != nil {
			return authconfig, err
		}

		if customAuth := registries.Find(reg); customAuth != nil {
			authconfig = docker.AuthConfiguration{
				Password:      customAuth.Password,
				ServerAddress: customAuth.Name,
				Username:      customAuth.Username,
			}
		}
	}

	return authconfig, nil
}
