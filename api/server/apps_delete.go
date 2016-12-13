package server

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/iron-io/functions/api/models"
	"github.com/iron-io/runner/common"
)

func (s *Server) handleAppDelete(c *gin.Context) {
	ctx := c.MustGet("ctx").(context.Context)
	log := common.Logger(ctx)

	appName := ctx.Value("appName").(string)

	routes, err := s.Datastore.GetRoutesByApp(ctx, appName, &models.RouteFilter{})
	if err != nil {
		log.WithError(err).Error(models.ErrAppsRemoving)
		c.JSON(http.StatusInternalServerError, simpleError(ErrInternalServerError))
		return
	}

	if len(routes) > 0 {
		log.WithError(err).Debug(models.ErrDeleteAppsWithRoutes)
		c.JSON(http.StatusBadRequest, simpleError(models.ErrDeleteAppsWithRoutes))
		return
	}

	err = s.FireBeforeAppDelete(ctx, appName)
	if err != nil {
		log.WithError(err).Error(models.ErrAppsRemoving)
		c.JSON(http.StatusInternalServerError, simpleError(ErrInternalServerError))
		return
	}

	err = s.Datastore.RemoveApp(ctx, appName)
	if err != nil {
		if err == models.ErrAppsNotFound {
			log.WithError(err).Debug(models.ErrAppsRemoving)
			c.JSON(http.StatusNotFound, simpleError(err))
			return
		}

		log.WithError(err).Error(models.ErrAppsRemoving)
		c.JSON(http.StatusInternalServerError, simpleError(ErrInternalServerError))
		return
	}

	err = s.FireAfterAppDelete(ctx, appName)
	if err != nil {
		log.WithError(err).Error(models.ErrAppsRemoving)
		c.JSON(http.StatusInternalServerError, simpleError(ErrInternalServerError))
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "App deleted"})
}
