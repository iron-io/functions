package server

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/iron-io/functions/api/models"
)

func (s *Server) handleAppList(ctx context.Context, r RequestController) {
	c := ctx.(*gin.Context)

	filter := &models.AppFilter{}

	apps, err := s.Datastore.GetApps(c, filter)
	if err != nil {
		handleErrorResponse(c, r, err)
		return
	}

	c.JSON(http.StatusOK, appsResponse{"Successfully listed applications", apps})
}
