// +build integration

package server

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/iron-io/functions/api/models"
	"github.com/iron-io/functions/fn/app"
	"github.com/spf13/viper"
	"github.com/urfave/cli"
)

var DB_FILE string
var MQ_FILE string
var API_URL string
var PORT int
var funcServer *Server
var Cancel context.CancelFunc
var Ctx context.Context
var fn *cli.App

func setupServer() {
	viper.Set(EnvDBURL, fmt.Sprintf("bolt://%s?bucket=funcs", DB_FILE))
	viper.Set(EnvMQURL, fmt.Sprintf("bolt://%s", MQ_FILE))
	viper.Set(EnvPort, PORT)
	Ctx, Cancel = context.WithCancel(context.Background())
	funcServer = NewFromEnv(Ctx)
	go funcServer.Start(Ctx)
	time.Sleep(2 * time.Second)
}

func setupCli() {
	viper.Set("API_URL", API_URL)
	fn = app.NewFn()
}

func teardown() {
	os.Remove(DB_FILE)
	os.Remove(MQ_FILE)
	Cancel()
	time.Sleep(2 * time.Second)
}

func TestIntegration(t *testing.T) {
	DB_FILE = "/tmp/bolt_fn_db.db"
	MQ_FILE = "/tmp/bolt_fn_mq.db"
	PORT = 8080
	API_URL = "http://localhost:8080"
	setupServer()
	setupCli()
	testIntegration(t)
	teardown()
}

func TestIntegrationWithAuth(t *testing.T) {
	viper.Set("jwt_auth_key", "test")
	DB_FILE = "/tmp/bolt_fn_auth_db.db"
	MQ_FILE = "/tmp/bolt_fn_auth_mq.db"
	PORT = 8081
	API_URL = "http://localhost:8081"
	setupServer()
	setupCli()
	testIntegration(t)
	teardown()
}

func testIntegration(t *testing.T) {
	// Test list

	err := fn.Run([]string{"fn", "apps", "l"})
	if err != nil {
		t.Error(err)
	}

	// Test create app

	err = fn.Run([]string{"fn", "apps", "c", "test"})
	if err != nil {
		t.Error(err)
	}

	filter := &models.AppFilter{}
	apps, err := funcServer.Datastore.GetApps(Ctx, filter)

	if len(apps) != 1 {
		t.Error("fn apps create failed.")
	}

	if apps[0].Name != "test" {
		t.Error("fn apps create failed. - name doesnt match")
	}

	// Test create route

	err = fn.Run([]string{"fn", "routes", "c", "test", "/new-route", "--jwt-key", "route_key"})
	if err != nil {
		t.Error(err)
	}

	routeFilter := &models.RouteFilter{}
	routes, err := funcServer.Datastore.GetRoutes(Ctx, routeFilter)

	if len(routes) != 1 {
		t.Error("fn routes create failed.")
	}

	if routes[0].Path != "/new-route" {
		t.Error("fn routes create failed. - path doesnt match")
	}

	// Test call route

	err = fn.Run([]string{"fn", "routes", "call", "test", "/new-route"})
	if err != nil {
		t.Error(err)
	}

	// Test delete route

	err = fn.Run([]string{"fn", "routes", "delete", "test", "/new-route"})
	if err != nil {
		t.Error(err)
	}

	routes, err = funcServer.Datastore.GetRoutes(Ctx, routeFilter)

	if len(routes) != 0 {
		t.Error("fn routes delete failed.")
	}

	// Test delete app

	err = fn.Run([]string{"fn", "apps", "delete", "test"})
	if err != nil {
		t.Error(err)
	}

	apps, err = funcServer.Datastore.GetApps(Ctx, filter)

	if len(apps) != 0 {
		t.Error("fn apps delete failed.")
	}
}
