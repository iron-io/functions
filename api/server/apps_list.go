package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/iron-io/functions/api/models"
)

func (s *Server) handleAppList(c *gin.Context, r RequestController) {
	filter := &models.AppFilter{}

	apps, err := s.Datastore.GetApps(c, filter)
	if err != nil {
		handleErrorResponse(c, r, err)
		return
	}

	c.JSON(http.StatusOK, appsResponse{"Successfully listed applications", apps})
}
