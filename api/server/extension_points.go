// TODO: it would be nice to move these into the top level folder so people can use these with the "functions" package, eg: functions.ApiHandler
package server

import (
	"context"

	"github.com/gin-gonic/gin"
)

type ApiHandler interface {
	Handle(ctx context.Context)
}

type apiHandlerWrapper struct {
	// apiHandler
}

func apiHandlerWrapperFunc(apiHandler ApiHandler) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.MustGet("ctx").(context.Context)
		apiHandler.Handle(ctx)
	}
}

// AddEndpoint adds an endpoint to /v1/x
func (s *Server) AddEndpoint(method, path string, handler ApiHandler) {
	v1 := s.Router.Group("/v1")
	// v1.GET("/apps/:app/log", logHandler(cfg))
	v1.Handle(method, path, apiHandlerWrapperFunc(handler))
}

// AddAppEndpoint adds an endpoints to /v1/apps/:app/x
func (s *Server) AddAppEndpoint(method, path string, handler ApiHandler) {
	v1 := s.Router.Group("/v1")
	v1.Handle(method, "/apps/:app"+path, apiHandlerWrapperFunc(handler))
}
