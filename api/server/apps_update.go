package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/iron-io/functions/api"
	"github.com/iron-io/functions/api/models"
)

func (s *Server) handleAppUpdate(c *gin.Context, r RequestController) {
	log := r.Logger()

	wapp := models.AppWrapper{}

	err := c.BindJSON(&wapp)
	if err != nil {
		log.WithError(err).Debug(models.ErrInvalidJSON)
		c.JSON(http.StatusBadRequest, simpleError(models.ErrInvalidJSON))
		return
	}

	if wapp.App == nil {
		log.Debug(models.ErrAppsMissingNew)
		c.JSON(http.StatusBadRequest, simpleError(models.ErrAppsMissingNew))
		return
	}

	if wapp.App.Name != "" {
		log.Debug(models.ErrAppsNameImmutable)
		c.JSON(http.StatusBadRequest, simpleError(models.ErrAppsNameImmutable))
		return
	}

	wapp.App.Name = c.Param(api.CApp)

	err = s.FireAfterAppUpdate(c, wapp.App)
	if err != nil {
		log.WithError(err).Error(models.ErrAppsUpdate)
		c.JSON(http.StatusInternalServerError, simpleError(ErrInternalServerError))
		return
	}

	app, err := s.Datastore.UpdateApp(c, wapp.App)
	if err != nil {
		handleErrorResponse(c, r, err)
		return
	}

	err = s.FireAfterAppUpdate(c, wapp.App)
	if err != nil {
		log.WithError(err).Error(models.ErrAppsUpdate)
		c.JSON(http.StatusInternalServerError, simpleError(ErrInternalServerError))
		return
	}

	c.JSON(http.StatusOK, appResponse{"App successfully updated", app})
}
