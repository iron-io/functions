package server

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/iron-io/functions/api"
	"github.com/iron-io/functions/api/models"
	"github.com/iron-io/functions/api/runner/task"
	"github.com/iron-io/runner/common"
)

func (s *Server) handleRouteUpdate(ctx context.Context, r RequestController) {
	log := common.Logger(ctx)
	c := ctx.(*gin.Context)

	var wroute models.RouteWrapper

	wroute.Route = r.Route()
	err := r.Error()
	if err != nil {
		log.WithError(err).Debug(models.ErrInvalidJSON)
		c.JSON(http.StatusBadRequest, simpleError(models.ErrInvalidJSON))
		return
	}

	if wroute.Route == nil {
		log.Debug(models.ErrRoutesMissingNew)
		c.JSON(http.StatusBadRequest, simpleError(models.ErrRoutesMissingNew))
		return
	}

	wroute.Route.AppName = c.Param(api.CApp)

	if wroute.Route.Path != "" {
		log.Debug(models.ErrRoutesPathImmutable)
		c.JSON(http.StatusBadRequest, simpleError(models.ErrRoutesPathImmutable))
		return
	}

	wroute.Route.Path = c.Param(api.CRoute)

	if wroute.Route.Image != "" {
		err = s.Runner.EnsureImageExists(c, &task.Config{
			Image: wroute.Route.Image,
		})
		if err != nil {
			log.WithError(err).Debug(models.ErrRoutesUpdate)
			c.JSON(http.StatusBadRequest, simpleError(models.ErrUsableImage))
			return
		}
	}

	route, err := s.Datastore.UpdateRoute(c, wroute.Route)
	if err != nil {
		handleErrorResponse(c, r, err)
		return
	}

	s.cacherefresh(route)

	c.JSON(http.StatusOK, routeResponse{"Route successfully updated", route})
}
