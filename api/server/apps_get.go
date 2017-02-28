package server

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (s *Server) handleAppGet(ctx context.Context, r RequestController) {
	c := ctx.(*gin.Context)

	app := r.App()
	if r.Error() != nil {
		handleErrorResponse(c, r, r.Error())
		return
	}

	c.JSON(http.StatusOK, appResponse{"Successfully loaded app", app})
}
