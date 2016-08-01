package router

import (
	"net/http"

	"github.com/Sirupsen/logrus"
	"github.com/gin-gonic/gin"
	"github.com/iron-io/functions/api/models"
	"github.com/iron-io/functions/api/server"
)

func handleRouteList(c *gin.Context) {
	log := c.MustGet("log").(logrus.FieldLogger)

	appName := c.Param("app")

	filter := &models.RouteFilter{
		AppName: appName,
	}

	routes, err := api.Datastore.GetRoutes(filter)
	if err != nil {
		log.WithError(err).Error(models.ErrRoutesGet)
		c.JSON(http.StatusInternalServerError, simpleError(models.ErrRoutesGet))
		return
	}

	log.WithFields(logrus.Fields{"routes": routes}).Debug("Got routes")

	c.JSON(http.StatusOK, &models.RoutesWrapper{Routes: routes})
}
