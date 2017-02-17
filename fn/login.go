package main

import (
	"github.com/urfave/cli"
	"fmt"
	"os"
	"encoding/base64"
	"encoding/json"
	fnclient "github.com/iron-io/functions_go/client"
	"github.com/iron-io/functions_go/models"
	"github.com/iron-io/functions_go/client/docker"
)

type DockerLoginCmd struct {
	Email         *string `json:"email"`
	Username      *string `json:"username"`
	Password      *string `json:"password"`
	Serveraddress *string `json:"serveraddress"`
	client *fnclient.Functions
}

func login() cli.Command {
	a := DockerLoginCmd{client: apiClient()}

	return cli.Command{
		Name:        "docker",
		ArgsUsage:   "fn docker",
		Subcommands: []cli.Command{
			{
				Name:      "login",
				Usage:     "Storing docker repo credentials",
				ArgsUsage: "docker login -u -p -e -url",
				Action:    a.login,
				Flags:     []cli.Flag{
					cli.StringFlag{
						Name:        "username,u",
						Usage:       "docker repo user name",
						Destination: a.Username,
					},
					cli.StringFlag{
						Name:        "password,p",
						Usage:       "docker repo password",
						Destination: a.Password,
					},
					cli.StringFlag{
						Name:        "email,e",
						Usage:       "docker repo user email",
						Destination: a.Email,
					},
					cli.StringFlag{
						Name:        "url",
						Usage:       "docker repo url, if you're using custom repo",
						Destination: a.Serveraddress,
					},
				},
			},
		},
	}
}

func (l *DockerLoginCmd) login() {
	fmt.Println("Storing docker repo credentials")

	//{"username": "string", "password": "string", "email": "string", "serveraddress" : "string", "auth": ""}
	bytes, err := json.Marshal(*l)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error marshaling credentials to json: %v", err)
		return
	}
	authString := base64.StdEncoding.EncodeToString(bytes)

	params := &docker.PostDockerLoginParams{
		Body: &models.DockerCreds{
			Auth: authString,
		},
	}

	_, err = l.client.Docker.PostDockerLogin(params)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
	fmt.Println(`Added docker repo credentials`)
}