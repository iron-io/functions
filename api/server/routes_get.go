package server

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (s *Server) handleRouteGet(ctx context.Context, r RequestController) {
	c := ctx.(*gin.Context)

	route := r.Route()
	err := r.Error()
	if err != nil {
		handleErrorResponse(c, r, err)
		return
	}

	c.JSON(http.StatusOK, routeResponse{"Successfully loaded route", route})
}
