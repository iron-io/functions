package server

import (
	"net/http"
	"path"

	"github.com/gin-gonic/gin"
	"github.com/iron-io/functions/api"
)

func (s *Server) handleRouteGet(c *gin.Context, r RequestController) {
	appName := c.Param(api.CApp)
	routePath := path.Clean(c.Param(api.CRoute))

	route, err := s.Datastore.GetRoute(c, appName, routePath)
	if err != nil {
		handleErrorResponse(c, r, err)
		return
	}

	c.JSON(http.StatusOK, routeResponse{"Successfully loaded route", route})
}
