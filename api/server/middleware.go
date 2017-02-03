// TODO: it would be nice to move these into the top level folder so people can use these with the "functions" package, eg: functions.AddMiddleware(...)
package server

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/iron-io/functions/api"
)

// Middleware is the interface required for implementing functions middleware
type Middleware func(*MiddlewareContext)

type MiddlewareContext struct {
	AppName   string
	RoutePath string

	context.Context
	*http.Request
	http.ResponseWriter

	Next  func(*MiddlewareContext)
	Index int
}

func runMiddlewares(rootCtx *MiddlewareContext, middlewares []Middleware) {
	index := -1
	next := func(mc *MiddlewareContext) {
		index++
		mdw := middlewares[index]
		mcc := *mc
		mcc.Index = index
		mdw(&mcc)
		mc.AppName = mcc.AppName
		mc.RoutePath = mcc.RoutePath
	}

	rootCtx.Next = next

	for index < len(middlewares)-1 {
		next(rootCtx)
	}
}

func (s *Server) middlewareWrapperFunc(ctx context.Context) gin.HandlerFunc {
	return func(c *gin.Context) {
		if len(s.middlewares) == 0 {
			return
		}

		appName := c.Param(api.AppName)
		routePath := c.Param(api.Path)

		ctx := c.MustGet("ctx").(context.Context)

		mc := MiddlewareContext{
			Context:        ctx,
			Request:        c.Request,
			ResponseWriter: c.Writer,

			AppName:   appName,
			RoutePath: routePath,
		}

		runMiddlewares(&mc, s.middlewares)

		c.Set(api.AppName, mc.AppName)
		c.Set(api.Path, mc.RoutePath)
	}
}

// AddAppEndpoint adds an endpoints to /v1/apps/:app/x
func (s *Server) AddMiddleware(m Middleware) {
	s.middlewares = append(s.middlewares, m)
}

// AddAppEndpoint adds an endpoints to /v1/apps/:app/x
func (s *Server) AddMiddlewareFunc(m http.HandlerFunc) {
	s.AddMiddleware(func(c *MiddlewareContext) {
		m.ServeHTTP(c.ResponseWriter, c.Request)
	})
}
