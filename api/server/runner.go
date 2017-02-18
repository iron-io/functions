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
	fcommon "github.com/iron-io/functions/api/common"
	"github.com/iron-io/functions/api/models"
	"github.com/iron-io/functions/api/runner"
	"github.com/iron-io/functions/api/runner/task"
	"github.com/iron-io/runner/common"
	uuid "github.com/satori/go.uuid"
)

type runnerResponse struct {
	RequestID string            `json:"request_id,omitempty"`
	Error     *models.ErrorBody `json:"error,omitempty"`
}

func toEnvName(envtype, name string) string {
	name = strings.ToUpper(strings.Replace(name, "-", "_", -1))
	return fmt.Sprintf("%s_%s", envtype, name)
}

func (s *Server) handleRequest(c *gin.Context, r RequestController) {
	reqID := uuid.NewV5(uuid.Nil, fmt.Sprintf("%s%s%d", c.Request.RemoteAddr, c.Request.URL.Path, time.Now().Unix())).String()
	ctx, log := fcommon.LoggerWithStack(c, "handleAppRequest")
	_, log = common.LoggerWithFields(ctx, logrus.Fields{"call_id": reqID})
	r.SetLogger(log)

	var err error
	var payload io.Reader

	if c.Request.Method == "POST" {
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

	reqRoute := &models.Route{
		AppName: r.AppName(),
		Path:    r.RoutePath(),
	}

	s.FireBeforeDispatch(c, reqRoute)

	if reqRoute.AppName == "" {
		c.JSON(http.StatusNotFound, models.ErrAppsNotFound)
		return
	}

	app, err := s.Datastore.GetApp(c, reqRoute.AppName)
	if err != nil || app == nil {
		if err != nil {
			log.WithError(err).Error("error getting app from datastore")
		}
		c.JSON(http.StatusNotFound, simpleError(models.ErrAppsNotFound))
		return
	}

	log.WithFields(logrus.Fields{"app": reqRoute.AppName, "path": reqRoute.Path}).Debug("Finding route on datastore")
	routes, err := s.loadroutes(c, models.RouteFilter{AppName: reqRoute.AppName, Path: reqRoute.Path})
	if err != nil {
		log.WithError(err).Error(models.ErrRoutesList)
		c.JSON(http.StatusInternalServerError, simpleError(models.ErrRoutesList))
		return
	}

	if len(routes) == 0 {
		log.WithError(err).Error("route not found in db:", reqRoute.Path)
		c.JSON(http.StatusNotFound, simpleError(models.ErrRunnerRouteNotFound))
		return
	}

	log.WithField("routes", len(routes)).Debug("Got routes from datastore")
	route := routes[0]
	log = log.WithFields(logrus.Fields{"app": reqRoute.AppName, "path": route.Path, "image": route.Image})

	if s.serve(c, reqRoute.AppName, route, app, reqRoute.Path, reqID, payload, s.Enqueue) {
		s.FireAfterDispatch(c, reqRoute)
		return
	}

	log.Error(models.ErrRunnerRouteNotFound)
	c.JSON(http.StatusNotFound, simpleError(models.ErrRunnerRouteNotFound))
}

func (s *Server) loadroutes(ctx context.Context, filter models.RouteFilter) ([]*models.Route, error) {
	if route, ok := s.cacheget(filter.AppName, filter.Path); ok {
		return []*models.Route{route}, nil
	}
	resp, err := s.singleflight.do(
		filter,
		func() (interface{}, error) {
			return s.Datastore.GetRoutesByApp(ctx, filter.AppName, &filter)
		},
	)
	return resp.([]*models.Route), err
}

// TODO: Should remove *gin.Context from these functions, should use only context.Context
func (s *Server) serve(c *gin.Context, appName string, found *models.Route, app *models.App, route, reqID string, payload io.Reader, enqueue models.Enqueue) (ok bool) {
	ctx, log := fcommon.LoggerWithStack(c, "serve")
	_, log = common.LoggerWithFields(ctx, logrus.Fields{"app": appName, "route": found.Path, "image": found.Image})
	c.Set("log", log)
	log.Infoln("serving ", appName, found.Path)

	params, match := matchRoute(found.Path, route)
	if !match {
		return false
	}

	var stdout bytes.Buffer // TODO: should limit the size of this, error if gets too big. akin to: https://golang.org/pkg/io/#LimitReader

	envVars := map[string]string{
		"METHOD":      c.Request.Method,
		"ROUTE":       found.Path,
		"REQUEST_URL": c.Request.URL.String(),
	}

	// app config
	for k, v := range app.Config {
		envVars[toEnvName("", k)] = v
	}
	for k, v := range found.Config {
		envVars[toEnvName("", k)] = v
	}

	// params
	for _, param := range params {
		envVars[toEnvName("PARAM", param.Key)] = param.Value
	}

	// headers
	for header, value := range c.Request.Header {
		envVars[toEnvName("HEADER", header)] = strings.Join(value, " ")
	}

	cfg := &task.Config{
		AppName:        appName,
		Path:           found.Path,
		Env:            envVars,
		Format:         found.Format,
		ID:             reqID,
		Image:          found.Image,
		MaxConcurrency: found.MaxConcurrency,
		Memory:         found.Memory,
		Stdin:          payload,
		Stdout:         &stdout,
		Timeout:        time.Duration(found.Timeout) * time.Second,
	}

	s.Runner.Enqueue()
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
		enqueue(c, s.MQ, task)
		log.Info("Added new task to queue")
		c.JSON(http.StatusAccepted, map[string]string{"call_id": task.ID})

	default:
		result, err := runner.RunTask(s.tasks, ctx, cfg)
		if err != nil {
			break
		}
		for k, v := range found.Headers {
			c.Header(k, v[0])
		}

		switch result.Status() {
		case "success":
			c.Data(http.StatusOK, "", stdout.Bytes())
		case "timeout":
			c.JSON(http.StatusGatewayTimeout, runnerResponse{
				RequestID: cfg.ID,
				Error: &models.ErrorBody{
					Message: models.ErrRunnerTimeout.Error(),
				},
			})
		default:
			c.JSON(http.StatusInternalServerError, runnerResponse{
				RequestID: cfg.ID,
				Error: &models.ErrorBody{
					Message: result.Error(),
				},
			})
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
