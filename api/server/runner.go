package server

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/gin-gonic/gin"
	"github.com/iron-io/functions/api/models"
	"github.com/iron-io/functions/api/runner"
	"github.com/iron-io/runner/common"
	uuid "github.com/satori/go.uuid"
)

func handleSpecial(c *gin.Context) {
	ctx := c.MustGet("ctx").(context.Context)
	log := common.Logger(ctx)

	err := Api.UseSpecialHandlers(c)
	if err != nil {
		log.WithError(err).Errorln("Error using special handler!")
		// todo: what do we do here? Should probably return a 500 or something
	}
}

func ToEnvName(envtype, name string) string {
	name = strings.ToUpper(strings.Replace(name, "-", "_", -1))
	return fmt.Sprintf("%s_%s", envtype, name)
}

func (s *Server) handleRequest(c *gin.Context, enqueue models.Enqueue) {
	if strings.HasPrefix(c.Request.URL.Path, "/v1") {
		c.Status(http.StatusNotFound)
		return
	}

	ctx := c.MustGet("ctx").(context.Context)

	reqID := uuid.NewV5(uuid.Nil, fmt.Sprintf("%s%s%d", c.Request.RemoteAddr, c.Request.URL.Path, time.Now().Unix())).String()
	c.Set("reqID", reqID) // todo: put this in the ctx instead of gin's

	ctx, log := common.LoggerWithFields(ctx, logrus.Fields{"call_id": reqID})

	var err error
	var payload io.Reader

	if c.Request.Method == "POST" || c.Request.Method == "PUT" {
		payload = c.Request.Body
		// Load complete body and close
		defer func() {
			io.Copy(ioutil.Discard, c.Request.Body)
			c.Request.Body.Close()
		}()
	} else if c.Request.Method == "GET" {
		reqPayload := c.Request.URL.Query().Get("payload")
		payload = strings.NewReader(reqPayload)
	}

	appName := c.Param("app")
	if appName == "" {
		// check context, app can be added via special handlers
		a, ok := c.Get("app")
		if ok {
			appName = a.(string)
		}
	}
	// if still no appName, we gotta exit
	if appName == "" {
		log.WithError(err).Error("Invalid app, blank")
		c.JSON(http.StatusBadRequest, simpleError(models.ErrAppsNotFound))
		return
	}
	rawroute := c.Param("route")
	if rawroute == "" {
		rawroute = c.Request.URL.Path
	}

	app, err := Api.Datastore.GetApp(appName)
	if err != nil || app == nil {
		log.WithError(err).Error(models.ErrAppsNotFound)
		c.JSON(http.StatusNotFound, simpleError(models.ErrAppsNotFound))
		return
	}

	// Theory of operation
	// The dynamic route matching happens in three phases: 1) LRU cache hit,
	// 2) Static Route Hit, ie, it assumes that the incoming request is
	// using a static route; 3) load app's routes and try matching each
	// one of them.

	log.WithFields(logrus.Fields{"app": appName, "path": rawroute}).Debug("Finding route on LRU cache")
	routes := s.loadcache(appName)
	if s.executeRoutes(routes, c, log, appName, app, rawroute, reqID, payload, enqueue) {
		return
	}

	log.WithFields(logrus.Fields{"app": appName, "path": rawroute}).Debug("Finding exact route on datastore")
	routes, err = Api.Datastore.GetRoutesByApp(appName, &models.RouteFilter{AppName: appName, Path: rawroute})
	if err != nil {
		log.WithError(err).Error(models.ErrRoutesList)
	}
	if s.executeRoutes(routes, c, log, appName, app, rawroute, reqID, payload, enqueue) {
		return
	}

	// TODO(ccirello): The problem here is that for every cold/cache-missed
	// request, we have to go through at least some subset of all routes for
	// this app. This will not scale. Some solutions have been proposed to
	// alleviate the issue, among them: count the number of slashes and use
	// it as an index to rule out routes that are either too long or too
	// short; register all apps routes into the gin's mux; for each matched
	// route, serve it once, and plug it to gin's mux for subsequent
	// requests. This a problem we cannot help. Perhaps, part of the
	// solution is to mention in the documentation that developers should
	// organize their applications in a way they don't have too many unused
	// routes. This is a potential point of DoS: a sufficiently high number
	// of non matched routes can overload the database and take the service
	// down.
	log.WithFields(logrus.Fields{"app": appName, "path": rawroute}).Debug("Finding route on datastore")
	routes, err = Api.Datastore.GetRoutesByApp(appName, &models.RouteFilter{AppName: appName})
	if err != nil {
		log.WithError(err).Error(models.ErrRoutesList)
		c.JSON(http.StatusInternalServerError, simpleError(models.ErrRoutesList))
		return
	}

	log.WithField("routes", routes).Debug("Got routes from datastore")

	if len(routes) == 0 {
		log.WithError(err).Error(models.ErrRunnerRouteNotFound)
		c.JSON(http.StatusNotFound, simpleError(models.ErrRunnerRouteNotFound))
		return
	}

	if !s.executeRoutes(routes, c, log, appName, app, rawroute, reqID, payload, enqueue) {
		log.Error(models.ErrRunnerRouteNotFound)
		c.JSON(http.StatusNotFound, simpleError(models.ErrRunnerRouteNotFound))
	}
}

func (s *Server) executeRoutes(routes []*models.Route, c *gin.Context, log logrus.FieldLogger, appName string, app *models.App, rawroute, reqID string, payload io.Reader, enqueue models.Enqueue) (found bool) {
	for _, r := range routes {
		if ok := s.processRoute(c, log, appName, r, app, rawroute, reqID, payload, enqueue); ok {
			s.refreshcache(appName, r)
			return true
		}
	}
	return false
}

func (s *Server) processRoute(c *gin.Context, log logrus.FieldLogger, appName string, found *models.Route, app *models.App, route, reqID string, payload io.Reader, enqueue models.Enqueue) (ok bool) {
	log = log.WithFields(logrus.Fields{"app": appName, "route": found.Path, "image": found.Image})

	params, match := matchRoute(found.Path, route)
	if !match {
		return false
	}

	var stdout bytes.Buffer // TODO: should limit the size of this, error if gets too big. akin to: https://golang.org/pkg/io/#LimitReader
	stderr := runner.NewFuncLogger(appName, route, found.Image, reqID)

	envVars := map[string]string{
		"METHOD":      c.Request.Method,
		"ROUTE":       found.Path,
		"REQUEST_URL": c.Request.URL.String(),
	}

	// app config
	for k, v := range app.Config {
		envVars[ToEnvName("", k)] = v
	}
	for k, v := range found.Config {
		envVars[ToEnvName("", k)] = v
	}

	// params
	for _, param := range params {
		envVars[ToEnvName("PARAM", param.Key)] = param.Value
	}

	// headers
	for header, value := range c.Request.Header {
		envVars[ToEnvName("HEADER", header)] = strings.Join(value, " ")
	}

	cfg := &runner.Config{
		Image:   found.Image,
		Timeout: 30 * time.Second,
		ID:      reqID,
		AppName: appName,
		Stdout:  &stdout,
		Stderr:  stderr,
		Env:     envVars,
		Memory:  found.Memory,
		Stdin:   payload,
	}

	switch found.Type {
	case "async":
		// Read payload
		pl, err := ioutil.ReadAll(cfg.Stdin)
		if err != nil {
			log.WithError(err).Error(models.ErrInvalidPayload)
			c.JSON(http.StatusBadRequest, simpleError(models.ErrInvalidPayload))
			return true
		}

		// Create Task
		priority := int32(0)
		task := &models.Task{}
		task.Image = &cfg.Image
		task.ID = cfg.ID
		task.Path = found.Path
		task.AppName = cfg.AppName
		task.Priority = &priority
		task.EnvVars = cfg.Env
		task.Payload = string(pl)
		// Push to queue
		enqueue(task)
		log.Info("Added new task to queue")

	default:
		result, err := runner.RunTask(s.tasks, c, cfg)
		if err != nil {
			break
		}
		for k, v := range found.Headers {
			c.Header(k, v[0])
		}

		if result.Status() == "success" {
			c.Data(http.StatusOK, "", stdout.Bytes())
		} else {
			c.AbortWithStatus(http.StatusInternalServerError)
		}

	}

	return true
}

var fakeHandler = func(http.ResponseWriter, *http.Request, Params) {}

func matchRoute(baseRoute, route string) (Params, bool) {
	tree := &node{}
	tree.addRoute(baseRoute, fakeHandler)
	handler, p, _ := tree.getValue(route)
	if handler == nil {
		return nil, false
	}

	return p, true
}
