package server

import (
	"net/http"
	"path"

	"github.com/gin-gonic/gin"
	"github.com/iron-io/functions/api"
)

func (s *Server) handleRouteDelete(c *gin.Context, r RequestController) {
	appName := c.Param(api.CApp)
	routePath := path.Clean(c.Param(api.CRoute))

	if err := s.Datastore.RemoveRoute(c, appName, routePath); err != nil {
		handleErrorResponse(c, r, err)
		return
	}

	s.cachedelete(appName, routePath)
	c.JSON(http.StatusOK, gin.H{"message": "Route deleted"})
}
