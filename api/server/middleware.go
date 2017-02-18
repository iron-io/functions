// TODO: it would be nice to move these into the top level folder so people can use these with the "functions" package, eg: functions.AddMiddleware(...)
package server

import (
	"context"
	"reflect"

	"github.com/Sirupsen/logrus"
	"github.com/gin-gonic/gin"
	fcommon "github.com/iron-io/functions/api/common"
	"time"
)

// Middleware is the interface required for implementing functions middlewar
type Middleware interface {
	// Serve is what the Middleware must implement. Can modify the request, write output, etc.
	// todo: should we abstract the HTTP out of this?  In case we want to support other protocols.
	Serve(ctx MiddlewareContext, r RequestController) error
}

// MiddlewareFunc func form of Middleware
type MiddlewareFunc func(ctx MiddlewareContext, r RequestController) error

// Serve wrapper
func (f MiddlewareFunc) Serve(ctx MiddlewareContext, r RequestController) error {
	return f(ctx, r)
}

// AddMiddleware adds middleware to all /v1/* routes
func (s *Server) AddMiddleware(m Middleware) {
	s.middlewares = append(s.middlewares, m)
}

// AddAppEndpoint adds middleware to all /v1/* routes
func (s *Server) AddMiddlewareFunc(m func(ctx MiddlewareContext, r RequestController) error) {
	s.AddMiddleware(MiddlewareFunc(m))
}

// AddRunMiddleware adds middleware to the user functions routes, not the API
func (s *Server) AddRunMiddleware(m Middleware) {
	s.runMiddlewares = append(s.runMiddlewares, m)
}

// AddRunMiddleware adds middleware to the user functions routes, not the API
func (s *Server) AddRunMiddlewareFunc(m func(ctx MiddlewareContext, r RequestController) error) {
	s.AddRunMiddleware(MiddlewareFunc(m))
}

// MiddlewareContext extends context.Context for Middleware
type MiddlewareContext interface {
	context.Context

	// Middleware can call Next() explicitly to call the next middleware in the chain. If Next() is not called and an error is not returned, Next() will automatically be called.
	Next(ctx MiddlewareContext, r RequestController)
	// Index returns the index of where we're at in the chain
	Index() int
	// WithValue same behavior as context.WithValue, but returns MiddlewareContext
	WithValue(key, val interface{}) MiddlewareContext
}

type middlewareContextImpl struct {
	ctx *gin.Context

	nextCalled  bool
	index       int
	middlewares []Middleware
}

func (c *middlewareContextImpl) Next(ctx MiddlewareContext, r RequestController) {
	_, log := fcommon.LoggerWithStack(c, "Next")
	log.Infoln("Next called", ctx.Index())
	c2 := ctx.(*middlewareContextImpl)
	c2.nextCalled = true
	c2.index++
	c2.ctx.Set("logger", log)
	c2.serveNext(r)
}

func (c *middlewareContextImpl) serveNext(r RequestController) error {
	_, log := fcommon.LoggerWithStack(c, "serveNext")
	log.Infoln("serving middleware", c.Index())
	if c.Index() >= len(c.middlewares) {
		return nil
	}
	c.ctx.Set("logger", log)
	// make shallow copy:
	fctx2 := *c
	fctx2.nextCalled = false
	c.ctx.Request = c.ctx.Request.WithContext(&fctx2)
	nextM := c.middlewares[c.Index()]
	err := nextM.Serve(&fctx2, r)
	if err != nil {
		logrus.WithError(err).Warnln("Middleware error")
		// todo: might be a good idea to check if anything has been written yet, and if not, output the error: simpleError(err)
		// see: http://stackoverflow.com/questions/39415827/golang-http-check-if-responsewriter-has-been-written
		return err
	}
	// this will be true if the user called Next() explicitly. If not, let's call it here.
	// if !fctx2.nextCalled {
	// 	// then we automatically call next
	// 	fctx2.Next(c, c.ginContext.Writer, r, fctx2.app)
	// }

	return nil
}

func (c *middlewareContextImpl) Index() int {
	return c.index
}

// WithValue is essentially the same as context.Context, but returns the MiddlewareContext
func (c *middlewareContextImpl) WithValue(key, val interface{}) MiddlewareContext {
	if key == nil {
		panic("nil key")
	}
	if !reflect.TypeOf(key).Comparable() {
		panic("key is not comparable")
	}
	if keyAsString, ok := key.(string); ok {
		mc2 := &middlewareContextImpl{nextCalled: c.nextCalled, index: c.index, middlewares: c.middlewares}
		newctx := *c.ctx
		newctx.Set(keyAsString, val)
		*mc2.ctx = newctx
		return mc2
	}
	return c
}

/** context.Context interface methods */
func (c *middlewareContextImpl) Value(key interface{}) interface{} {
	return c.ctx.Value(key)
}

func (c *middlewareContextImpl) Deadline() (deadline time.Time, ok bool) {
	return
}

func (c *middlewareContextImpl) Done() <-chan struct{} {
	return nil
}

func (c *middlewareContextImpl) Err() error {
	return nil
}
