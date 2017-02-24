package server

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/iron-io/functions/api"
	"github.com/iron-io/functions/api/models"
	"github.com/iron-io/runner/common"
)

func (s *Server) handleRouteCreate(ctx context.Context, r RequestController) {
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
		log.WithError(err).Debug(models.ErrInvalidJSON)
		c.JSON(http.StatusBadRequest, simpleError(models.ErrRoutesMissingNew))
		return
	}

	wroute.Route.AppName = c.Param(api.CApp)

	if err := wroute.Validate(); err != nil {
		log.WithError(err).Debug(models.ErrRoutesCreate)
		c.JSON(http.StatusBadRequest, simpleError(err))
		return
	}

	if wroute.Route.Image == "" {
		log.WithError(models.ErrRoutesValidationMissingImage).Debug(models.ErrRoutesCreate)
		c.JSON(http.StatusBadRequest, simpleError(models.ErrRoutesValidationMissingImage))
		return
	}

	// err = s.Runner.EnsureImageExists(ctx, &task.Config{
	// 	Image: wroute.Route.Image,
	// })
	// if err != nil {
	// 	c.JSON(http.StatusBadRequest, simpleError(models.ErrUsableImage))
	// 	return
	// }

	app, err := s.Datastore.GetApp(c, wroute.Route.AppName)
	if err != nil && err != models.ErrAppsNotFound {
		log.WithError(err).Error(models.ErrAppsGet)
		c.JSON(http.StatusInternalServerError, simpleError(models.ErrAppsGet))
		return
	} else if app == nil {
		// Create a new application and add the route to that new application
		newapp := &models.App{Name: wroute.Route.AppName}
		if err := newapp.Validate(); err != nil {
			log.Error(err)
			c.JSON(http.StatusInternalServerError, simpleError(err))
			return
		}

		err = s.FireBeforeAppCreate(c, newapp)
		if err != nil {
			log.WithError(err).Error(models.ErrAppsCreate)
			c.JSON(http.StatusInternalServerError, simpleError(ErrInternalServerError))
			return
		}

		_, err = s.Datastore.InsertApp(c, newapp)
		if err != nil {
			log.WithError(err).Error(models.ErrRoutesCreate)
			c.JSON(http.StatusInternalServerError, simpleError(ErrInternalServerError))
			return
		}

		err = s.FireAfterAppCreate(c, newapp)
		if err != nil {
			log.WithError(err).Error(models.ErrRoutesCreate)
			c.JSON(http.StatusInternalServerError, simpleError(ErrInternalServerError))
			return
		}

	}

	route, err := s.Datastore.InsertRoute(c, wroute.Route)
	if err != nil {
		handleErrorResponse(c, r, err)
		return
	}

	s.cacherefresh(route)

	c.JSON(http.StatusOK, routeResponse{"Route successfully created", route})
}
