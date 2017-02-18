package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/iron-io/functions/api/models"
)

func (s *Server) handleAppCreate(c *gin.Context, r RequestController) {
	log := r.Logger()

	var wapp models.AppWrapper

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

	if err := wapp.Validate(); err != nil {
		log.Error(err)
		c.JSON(http.StatusInternalServerError, simpleError(err))
		return
	}

	err = s.FireBeforeAppCreate(c, wapp.App)
	if err != nil {
		log.WithError(err).Error(models.ErrAppsCreate)
		c.JSON(http.StatusInternalServerError, simpleError(err))
		return
	}

	app, err := s.Datastore.InsertApp(c, wapp.App)
	if err != nil {
		handleErrorResponse(c, r, err)
		return
	}

	err = s.FireAfterAppCreate(c, wapp.App)
	if err != nil {
		log.WithError(err).Error(models.ErrAppsCreate)
		c.JSON(http.StatusInternalServerError, simpleError(err))
		return
	}

	c.JSON(http.StatusOK, appResponse{"App successfully created", app})
}
