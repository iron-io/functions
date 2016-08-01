package router

import (
	"net/http"

	"github.com/Sirupsen/logrus"
	"github.com/gin-gonic/gin"
	"github.com/iron-io/functions/api/models"
	"github.com/iron-io/functions/api/server"
)

func handleAppList(c *gin.Context) {
	log := c.MustGet("log").(logrus.FieldLogger)

	filter := &models.AppFilter{}

	apps, err := api.Datastore.GetApps(filter)
	if err != nil {
		log.WithError(err).Debug(models.ErrAppsList)
		c.JSON(http.StatusInternalServerError, simpleError(models.ErrAppsList))
		return
	}

	c.JSON(http.StatusOK, &models.AppsWrapper{apps})
}
