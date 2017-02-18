package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/iron-io/functions/api"
	"github.com/iron-io/functions/api/models"
)

func (s *Server) handleRouteList(c *gin.Context, r RequestController) {
	filter := &models.RouteFilter{}

	if img := c.Query("image"); img != "" {
		filter.Image = img
	}

	var routes []*models.Route
	var err error
	appName := c.Param(api.CApp)
	if appName != "" {
		routes, err = s.Datastore.GetRoutesByApp(c, appName, filter)
	} else {
		routes, err = s.Datastore.GetRoutes(c, filter)
	}

	if err != nil {
		handleErrorResponse(c, r, err)
		return
	}

	c.JSON(http.StatusOK, routesResponse{"Sucessfully listed routes", routes})
}
