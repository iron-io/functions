package server

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/iron-io/runner/common"
)

func handleRouteDelete(c *gin.Context) {
	ctx := c.MustGet("ctx").(context.Context)
	log := common.Logger(ctx)

	appName := c.Param("app")
	routePath := c.Param("route")

	err := Api.Datastore.RemoveRoute(appName, routePath)
	if err != nil {
		log.Error(err)
		c.JSON(http.StatusInternalServerError, simpleError(err))
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Route deleted"})
}
