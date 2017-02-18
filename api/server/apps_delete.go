package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/iron-io/functions/api"
	"github.com/iron-io/functions/api/models"
)

func (s *Server) handleAppDelete(c *gin.Context, r RequestController) {
	log := r.Logger()

	app := &models.App{Name: c.Param(api.CApp)}

	routes, err := s.Datastore.GetRoutesByApp(c, app.Name, &models.RouteFilter{})
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

	err = s.FireBeforeAppDelete(c, app)
	if err != nil {
		log.WithError(err).Error(models.ErrAppsRemoving)
		c.JSON(http.StatusInternalServerError, simpleError(ErrInternalServerError))
		return
	}

	err = s.Datastore.RemoveApp(c, app.Name)
	if err != nil {
		handleErrorResponse(c, r, err)
		return
	}

	err = s.FireAfterAppDelete(c, app)
	if err != nil {
		log.WithError(err).Error(models.ErrAppsRemoving)
		c.JSON(http.StatusInternalServerError, simpleError(ErrInternalServerError))
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "App deleted"})
}
