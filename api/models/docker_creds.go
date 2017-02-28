package models

import "errors"

type DockerCreds struct {

	// Docker credentials in base64 encoding
	Auth string `json:"auth,omitempty"`
}

var (
	ErrInvalidDockerCreds = errors.New("Invalid Docker Credentials")
)
