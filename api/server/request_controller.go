package server

import (
	"net/http"

	"github.com/Sirupsen/logrus"
	"github.com/gin-gonic/gin"
	"github.com/iron-io/functions/api"
	"github.com/iron-io/runner/common"
)

type RequestController interface {
	AppName() string
	SetAppName(string)
	RoutePath() string
	SetRoutePath(string)
	Request() *http.Request
	Response() http.ResponseWriter
	Logger() logrus.FieldLogger
	SetLogger(logrus.FieldLogger)
}

type requestInfo struct {
	ctx *gin.Context
}

func (r *requestInfo) Response() http.ResponseWriter {
	return r.ctx.Writer
}

func (r *requestInfo) Request() *http.Request {
	return r.ctx.Request
}

func (r requestInfo) AppName() string {
	val, _ := r.ctx.Get(api.AppName)
	if v, ok := val.(string); ok {
		return v
	}
	return ""
}

func (r requestInfo) RoutePath() string {
	val, _ := r.ctx.Get(api.Path)
	if v, ok := val.(string); ok {
		return v
	}
	return ""
}

func (r *requestInfo) SetAppName(appname string) {
	r.ctx.Set(api.AppName, appname)
}

func (r *requestInfo) SetRoutePath(routepath string) {
	r.ctx.Set(api.Path, routepath)
}

func (r requestInfo) SetLogger(logger logrus.FieldLogger) {
	r.ctx.Set("logger", logger)
}

func (r requestInfo) Logger() logrus.FieldLogger {
	log, _ := r.ctx.Get("logger")
	if l, ok := log.(logrus.FieldLogger); ok {
		return l
	}
	return nil
}

func (s *Server) wrapHandler(f func(*gin.Context, RequestController)) gin.HandlerFunc {
	return func(c *gin.Context) {
		log := common.Logger(c)

		c.Set(api.AppName, c.Param(api.CApp))
		c.Set(api.Path, c.Param(api.CRoute))
		c.Set("logger", log.WithFields(extractFields(c)))

		info := &requestInfo{
			ctx: c,
		}
		fctx := &middlewareContextImpl{ctx: c}
		fctx.middlewares = append(s.middlewares, Middleware(MiddlewareFunc(func(ctx MiddlewareContext, r RequestController) error {
			f(c, r)
			return nil
		})))

		err := fctx.serveNext(info)
		if err != nil {
			c.AbortWithStatus(http.StatusInternalServerError)
		}
	}
}

func (s *Server) wrapRunHandler(f func(*gin.Context, RequestController)) gin.HandlerFunc {
	return func(c *gin.Context) {
		log := common.Logger(c)

		c.Set(api.AppName, c.Param(api.CApp))
		c.Set(api.Path, c.Param(api.CRoute))
		c.Set("logger", log.WithFields(extractFields(c)))

		info := &requestInfo{
			ctx: c,
		}
		fctx := &middlewareContextImpl{ctx: c}
		fctx.middlewares = append(s.middlewares, Middleware(MiddlewareFunc(func(ctx MiddlewareContext, r RequestController) error {
			f(c, r)
			return nil
		})))

		err := fctx.serveNext(info)
		if err != nil {
			c.AbortWithStatus(http.StatusInternalServerError)
		}
	}
}
