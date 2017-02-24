package server

import (
	"context"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/iron-io/functions/api"
	"github.com/iron-io/functions/api/models"
	"github.com/iron-io/runner/common"
)

type RequestController interface {
	App() *models.App
	SetApp(models.App)
	Route() *models.Route
	SetRoute(models.Route)
	Request() *http.Request
	Response() http.ResponseWriter
	Error() error
}

type requestInfo struct {
	ctx   *gin.Context
	app   *models.App
	route *models.Route
	err   error

	appOnce       sync.Once
	lazyLoadApp   func() (*models.App, error)
	routeOnce     sync.Once
	lazyLoadRoute func() (*models.Route, error)
}

func (r *requestInfo) Response() http.ResponseWriter {
	return r.ctx.Writer
}

func (r *requestInfo) Request() *http.Request {
	return r.ctx.Request
}

func (r *requestInfo) App() *models.App {
	var app *models.App

	if r.app == nil {
		r.appOnce.Do(func() {
			r.app, r.err = r.lazyLoadApp()
		})
	}

	if r.app != nil {
		a := *r.app
		app = &a
	}

	return app
}

func (r *requestInfo) Route() *models.Route {
	var route *models.Route

	if r.route == nil {
		r.routeOnce.Do(func() {
			r.route, r.err = r.lazyLoadRoute()
		})
	}

	if r.route != nil {
		rt := *r.route
		route = &rt
	}

	return route
}

func (r *requestInfo) SetApp(app models.App) {
	r.app = &app
}

func (r *requestInfo) SetRoute(route models.Route) {
	r.route = &route
}

func (r *requestInfo) Error() error {
	return r.err
}

func (s *Server) wrapHandler(f func(context.Context, RequestController)) gin.HandlerFunc {
	return func(c *gin.Context) {
		log := common.Logger(c)

		c.Set("logger", log.WithFields(extractFields(c)))
		appName := c.Param(api.CApp)
		routePath := c.Param(api.CRoute)

		info := &requestInfo{
			ctx: c,
			lazyLoadApp: func() (*models.App, error) {
				if c.Request.Method == "GET" {
					return s.Datastore.GetApp(c, appName)
				} else if c.Request.Method == "POST" || c.Request.Method == "PATCH" {
					var wapp models.AppWrapper
					err := c.BindJSON(&wapp)
					return wapp.App, err
				}
				return &models.App{Name: appName}, nil
			},
			lazyLoadRoute: func() (*models.Route, error) {
				if c.Request.Method == "GET" {
					return s.Datastore.GetRoute(c, appName, routePath)
				} else if c.Request.Method == "POST" || c.Request.Method == "PATCH" {
					var wroute models.RouteWrapper
					err := c.BindJSON(&wroute)
					return wroute.Route, err
				}
				return &models.Route{Path: routePath, AppName: appName}, nil
			},
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

func (s *Server) wrapRunHandler(f func(context.Context, RequestController)) gin.HandlerFunc {
	return func(c *gin.Context) {
		log := common.Logger(c)

		c.Set("logger", log.WithFields(extractFields(c)))
		appName := c.Param(api.CApp)
		routePath := c.Param(api.CRoute)

		info := &requestInfo{
			ctx: c,
			lazyLoadApp: func() (*models.App, error) {
				return s.Datastore.GetApp(c, appName)
			},
			lazyLoadRoute: func() (*models.Route, error) {
				return &models.Route{Path: routePath, AppName: appName}, nil
			},
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
