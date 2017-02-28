package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/fsouza/go-dockerclient"
	fnclient "github.com/iron-io/functions_go/client"
	dockerfnclient "github.com/iron-io/functions_go/client/docker"
	"github.com/iron-io/functions_go/models"
	"github.com/urfave/cli"
	"os"
)

type DockerLoginCmd struct {
	client *fnclient.Functions
}

func login() cli.Command {
	a := DockerLoginCmd{client: apiClient()}

	return cli.Command{
		Name:      "docker",
		ArgsUsage: "fn docker",
		Subcommands: []cli.Command{
			{
				Name:      "login",
				Usage:     "Storing docker repo credentials",
				ArgsUsage: "docker login -u -p -e -url",
				Action:    a.login,
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "username,u",
						Usage: "docker repo user name",
					},
					cli.StringFlag{
						Name:  "password,p",
						Usage: "docker repo password",
					},
					cli.StringFlag{
						Name:  "email,e",
						Usage: "docker repo user email",
					},
					cli.StringFlag{
						Name:  "url",
						Usage: "docker repo url, if you're using custom repo",
					},
				},
			},
		},
	}
}

func (l *DockerLoginCmd) login(c *cli.Context) error {
	fmt.Println("Storing docker repo credentials")

	authCfg := docker.AuthConfiguration{
		Username:      c.String("username"),
		Password:      c.String("password"),
		Email:         c.String("email"),
		ServerAddress: c.String("url"),
	}
	fmt.Fprintf(os.Stderr, "%v", authCfg)
	//{"username": "string", "password": "string", "email": "string", "serveraddress" : "string", "auth": ""}
	bytes, err := json.Marshal(authCfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error marshaling credentials to json: %v", err)
		return fmt.Errorf("unexpected error: %v", err)
	}
	authString := base64.StdEncoding.EncodeToString(bytes)

	params := &dockerfnclient.PostDockerLoginParams{
		Body: &models.DockerCreds{
			Auth: authString,
		},
		Context: context.Background(),
	}

	res, err := l.client.Docker.PostDockerLogin(params)
	fmt.Fprintf(os.Stdout, "res := %v", res)
	if err != nil {
		fmt.Fprintf(os.Stderr, "err: %v", err)
		return fmt.Errorf("unexpected error: %v", err)
	}
	fmt.Println(`Added docker repo credentials`)

	return nil
}
