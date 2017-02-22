package server

import (
	"context"
	"net/http"
	"path"

	"github.com/gin-gonic/gin"
	"github.com/iron-io/functions/api"
)

func (s *Server) handleRouteGet(c *gin.Context) {
	ctx := c.MustGet("ctx").(context.Context)

	appName := c.Param(api.CApp)
	routePath := path.Clean(c.Param(api.CRoute))

	route, err := s.Datastore.GetRoute(ctx, appName, routePath)
	if err != nil {
		handleErrorResponse(c, err)
		return
	}

	c.JSON(http.StatusOK, routeResponse{"Successfully loaded route", route})
}
