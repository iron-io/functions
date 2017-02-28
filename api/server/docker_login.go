package server

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"github.com/fsouza/go-dockerclient"
	"github.com/gin-gonic/gin"
	"github.com/iron-io/functions/api/models"
	"github.com/iron-io/runner/common"
	"net/http"
)

func (s *Server) handleDockerLogin(c *gin.Context) {

	ctx := c.MustGet("ctx").(context.Context)
	log := common.Logger(ctx)

	dockerLogin := models.DockerCreds{}

	err := c.BindJSON(&dockerLogin)
	if err != nil {
		log.WithError(err).Debug(models.ErrInvalidJSON)
		c.JSON(http.StatusBadRequest, simpleError(models.ErrInvalidJSON))
		return
	}
	log.Infoln(dockerLogin)

	err = s.Datastore.SaveDockerCredentials(ctx, dockerLogin)

	c.JSON(http.StatusOK, nil)
}

func getAuthConfiguration(s *Server, ctx context.Context) (*docker.AuthConfiguration, error) {
	creds, err := s.Datastore.GetDockerCredentials(ctx)
	if err != nil {
		return nil, err
	}
	if creds == nil {
		return &docker.AuthConfiguration{}, nil
	}

	authCfg := &docker.AuthConfiguration{}
	data, err := base64.StdEncoding.DecodeString(creds.Auth)

	err = json.Unmarshal(data, authCfg)
	if err != nil {
		return nil, err
	}
	return authCfg, nil
}
