package server

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/iron-io/functions/api/models"
	"github.com/iron-io/runner/common"
	"net/http"
)

func (s *Server) handleDockerLogin(c *gin.Context) {

	ctx := c.MustGet("ctx").(context.Context)
	log := common.Logger(ctx)

	dockerCreds := models.DockerCreds{}

	err := c.BindJSON(&dockerCreds)
	if err != nil {
		log.WithError(err).Debug(models.ErrInvalidJSON)
		c.JSON(http.StatusBadRequest, simpleError(models.ErrInvalidJSON))
		return
	}

	if err := dockerCreds.Validate(); err != nil {
		log.Error(err)
		c.JSON(http.StatusBadRequest, simpleError(err))
		return
	}

	err = s.DockerAuth.SaveDockerCredentials(ctx, dockerCreds)

	c.JSON(http.StatusOK, nil)
}
