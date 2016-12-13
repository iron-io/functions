// +build go1.8

package main

import (
	"context"
	"fmt"
	"os"
	"plugin"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/ccirello/supervisor"
	"github.com/iron-io/functions/api/models"
	"github.com/iron-io/functions/api/runner"
	"github.com/iron-io/functions/api/runner/task"
	"github.com/iron-io/functions/api/server"
	"github.com/spf13/viper"
)

const (
	envPlugins = "plugins"
)

func loadPlugins(srv *server.Server) {
	plugins := strings.Split(viper.GetString(envPlugins), ",")
	for _, fn := range plugins {
		if fn == "" {
			continue
		}
		log.WithField("plugin", fn).Info("plugging in")
		if !exists(fn) {
			log.WithField("plugin", fn).Info("could not find plugin file")
			continue
		}

		p, err := plugin.Open(fn)
		if err != nil {
			log.WithField("plugin", fn).Info("could not load plugin")
			continue
		}

		plugRunnerListener(srv, fn, p)
		plugAppCreateListener(srv, fn, p)
		plugAppUpdateListener(srv, fn, p)
		plugAppDeleteListener(srv, fn, p)
		plugSpecialHandler(srv, fn, p)
	}
}

func exists(name string) bool {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

func plugRunnerListener(srv *server.Server, fn string, p *plugin.Plugin) {
	l, err := p.Lookup("RunnerListener")
	if err != nil {
		log.WithFields(log.Fields{
			"error":  err,
			"plugin": fn,
			"type":   "RunnerListener"},
		).Error("not plugged")
		return
	}

	rl, ok := l.(server.RunnerListener)
	if !ok {
		log.WithFields(log.Fields{
			"plugin": fn,
			"type":   fmt.Sprintf("%T", l)},
		).Error("wrong type")
		return
	}

	log.WithFields(log.Fields{
		"plugin": fn,
		"type":   "RunnerListener"},
	).Info("plugged")
	srv.AddRunnerListener(rl)
}

func plugAppCreateListener(srv *server.Server, fn string, p *plugin.Plugin) {
	l, err := p.Lookup("AppCreateListener")
	if err != nil {
		log.WithFields(log.Fields{
			"error":  err,
			"plugin": fn,
			"type":   "AppCreateListener"},
		).Error("not plugged")
		return
	}

	rl, ok := l.(server.AppCreateListener)
	if !ok {
		log.WithFields(log.Fields{
			"plugin": fn,
			"type":   fmt.Sprintf("%T", l)},
		).Error("wrong type")
		return
	}

	log.WithFields(log.Fields{
		"plugin": fn,
		"type":   "AppCreateListener"},
	).Info("plugged")
	srv.AddAppCreateListener(rl)
}

func plugAppUpdateListener(srv *server.Server, fn string, p *plugin.Plugin) {
	l, err := p.Lookup("AppUpdateListener")
	if err != nil {
		log.WithFields(log.Fields{
			"error":  err,
			"plugin": fn,
			"type":   "AppUpdateListener"},
		).Error("not plugged")
		return
	}

	rl, ok := l.(server.AppUpdateListener)
	if !ok {
		log.WithFields(log.Fields{
			"plugin": fn,
			"type":   fmt.Sprintf("%T", l)},
		).Error("wrong type")
		return
	}

	log.WithFields(log.Fields{
		"plugin": fn,
		"type":   "AppUpdateListener"},
	).Info("plugged")
	srv.AddAppUpdateListener(rl)
}

func plugAppDeleteListener(srv *server.Server, fn string, p *plugin.Plugin) {
	l, err := p.Lookup("AppDeleteListener")
	if err != nil {
		log.WithFields(log.Fields{
			"error":  err,
			"plugin": fn,
			"type":   "AppDeleteListener"},
		).Error("not plugged")
		return
	}

	rl, ok := l.(server.AppDeleteListener)
	if !ok {
		log.WithFields(log.Fields{
			"plugin": fn,
			"type":   fmt.Sprintf("%T", l)},
		).Error("wrong type")
		return
	}

	log.WithFields(log.Fields{
		"plugin": fn,
		"type":   "AppDeleteListener"},
	).Info("plugged")
	srv.AddAppDeleteListener(rl)
}

func plugSpecialHandler(srv *server.Server, fn string, p *plugin.Plugin) {
	l, err := p.Lookup("SpecialHandler")
	if err != nil {
		log.WithFields(log.Fields{
			"error":  err,
			"plugin": fn,
			"type":   "SpecialHandler"},
		).Error("not plugged")
		return
	}

	rl, ok := l.(server.SpecialHandler)
	if !ok {
		log.WithFields(log.Fields{
			"plugin": fn,
			"type":   fmt.Sprintf("%T", l)},
		).Error("wrong type")
		return
	}

	log.WithFields(log.Fields{
		"plugin": fn,
		"type":   "SpecialHandler"},
	).Info("plugged")
	srv.AddSpecialHandler(rl)
}

func attachHTTP(
	svr *supervisor.Supervisor,
	ctx context.Context,
	ds models.Datastore,
	mq models.MessageQueue,
	rnr *runner.Runner,
	tasks chan task.Request,
	enqueue models.Enqueue,
) {
	svr.AddFunc(func(ctx context.Context) {
		srv := server.New(ctx, ds, mq, rnr, tasks, server.DefaultEnqueue)
		loadPlugins(srv)
		srv.Run()
		<-ctx.Done()
	})
}
