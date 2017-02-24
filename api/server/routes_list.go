package server

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/iron-io/functions/api/models"
)

func (s *Server) handleRouteList(ctx context.Context, r RequestController) {
	c := ctx.(*gin.Context)

	filter := &models.RouteFilter{}

	if img := c.Query("image"); img != "" {
		filter.Image = img
	}

	var routes []*models.Route
	var err error

	app := r.App()
	err = r.Error()
	if err != nil {
		handleErrorResponse(c, r, err)
		return
	}

	appName := app.Name
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
