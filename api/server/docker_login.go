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

	dockerLogin := models.DockerCreds{}

	err := c.BindJSON(&dockerLogin)
	if err != nil {
		log.WithError(err).Debug(models.ErrInvalidJSON)
		c.JSON(http.StatusBadRequest, simpleError(models.ErrInvalidJSON))
		return
	}
	log.Infoln(dockerLogin)

	err = s.DockerAuth.SaveDockerCredentials(ctx, dockerLogin)

	c.JSON(http.StatusOK, nil)
}
