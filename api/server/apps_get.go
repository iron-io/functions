package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/iron-io/functions/api"
)

func (s *Server) handleAppGet(c *gin.Context, r RequestController) {
	appName := c.Param(api.CApp)
	app, err := s.Datastore.GetApp(c, appName)
	if err != nil {
		handleErrorResponse(c, r, err)
		return
	}

	c.JSON(http.StatusOK, appResponse{"Successfully loaded app", app})
}
