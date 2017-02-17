package server

import (
	"github.com/gin-gonic/gin"
	"context"
	"github.com/iron-io/runner/common"
	"net/http"
	"github.com/iron-io/functions/api/models"
	"encoding/json"
)

func (s *Server) handleDockerLogin(c *gin.Context) {

	ctx := c.MustGet("ctx").(context.Context)
	log := common.Logger(ctx)

	dockerLogin := models.DockerLogin{}

	err := c.BindJSON(&dockerLogin)
	if err != nil {
		log.WithError(err).Debug(models.ErrInvalidJSON)
		c.JSON(http.StatusBadRequest, simpleError(models.ErrInvalidJSON))
		return
	}

	err = saveDockerCreds(ctx, s.Datastore, dockerLogin)

	c.JSON(http.StatusOK, nil)
}

func saveDockerCreds(ctx context.Context, ds models.Datastore, dockerLogin models.DockerLogin) error {

	val, err := json.Marshal(dockerLogin)
	if err != nil {
		return err
	}

	return ds.Put(ctx, []byte("dockerLogin"), val)
}