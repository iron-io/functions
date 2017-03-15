package models

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"github.com/fsouza/go-dockerclient"
)

type DockerCreds struct {
	// Docker credentials in base64 encoding
	Auth string `json:"auth,omitempty"`
}

var (
	ErrDockerCredsMissing = errors.New("Missing Docker Credentials")
	ErrDockerCredsInvalid = errors.New("Invalid Docker Credentials")
)

func (dc DockerCreds) ToDockerAuthentication() (*docker.AuthConfiguration, error) {

	data, err := base64.StdEncoding.DecodeString(dc.Auth)

	authCfg := &docker.AuthConfiguration{}

	err = json.Unmarshal(data, authCfg)
	if err != nil {
		return nil, err
	}
	return authCfg, nil
}
func (dc DockerCreds) Validate() error {
	if dc.Auth == "" {
		return ErrDockerCredsMissing
	}
	if _, err := dc.ToDockerAuthentication(); err != nil {
		return ErrDockerCredsInvalid
	}
	return nil
}
