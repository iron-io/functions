package server

import (
	"context"
	"net/http"
	"path"

	"github.com/gin-gonic/gin"
)

func (s *Server) handleRouteDelete(ctx context.Context, r RequestController) {
	c := ctx.(*gin.Context)

	route := r.Route()
	appName := route.AppName
	routePath := path.Clean(route.Path)

	if err := s.Datastore.RemoveRoute(c, appName, routePath); err != nil {
		handleErrorResponse(c, r, err)
		return
	}

	s.cachedelete(appName, routePath)
	c.JSON(http.StatusOK, gin.H{"message": "Route deleted"})
}
