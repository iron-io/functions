package auth

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"github.com/fsouza/go-dockerclient"
	"github.com/iron-io/functions/api/datastore"
	"github.com/iron-io/functions/api/models"
	"testing"
)

func TestDockerAuth_GetAuthConfiguration(t *testing.T) {

	dockerAuth := DockerAuth{
		Datastore: datastore.NewMock(nil, nil),
		Key:       []byte("A159B69FAF460F55C0966B6383CE0917"),
	}
	ctx := context.Background()

	authCfg := docker.AuthConfiguration{
		Username:      "testName",
		Password:      "password",
		Email:         "email",
		ServerAddress: "url",
	}
	//{"username": "string", "password": "string", "email": "string", "serveraddress" : "string", "auth": ""}
	bytes, err := json.Marshal(authCfg)
	if err != nil {
		t.Error(err)
	}
	authString := base64.StdEncoding.EncodeToString(bytes)

	creds := models.DockerCreds{
		Auth: authString,
	}
	dockerAuth.SaveDockerCredentials(ctx, creds)
	newAuthCfg, err := dockerAuth.GetAuthConfiguration(ctx)
	if err != nil {
		t.Error(err)
	}
	if newAuthCfg.Username != authCfg.Username {
		t.Fatalf("TestDockerAuth_GetAuthConfiguration: expected username `%v`, but it was `%v`", authCfg.Username, newAuthCfg.Username)
	}
	if newAuthCfg.Password != authCfg.Password {
		t.Fatalf("TestDockerAuth_GetAuthConfiguration: expected username `%v`, but it was `%v`", authCfg.Username, newAuthCfg.Username)
	}
	if newAuthCfg.Email != authCfg.Email {
		t.Fatalf("TestDockerAuth_GetAuthConfiguration: expected username `%v`, but it was `%v`", authCfg.Username, newAuthCfg.Username)
	}
	if newAuthCfg.ServerAddress != authCfg.ServerAddress {
		t.Fatalf("TestDockerAuth_GetAuthConfiguration: expected username `%v`, but it was `%v`", authCfg.Username, newAuthCfg.Username)
	}
}
