// +build integration

package server

import (
	"context"
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
	viper.Set(EnvDBURL, DB_FILE)
	viper.Set(EnvMQURL, MQ_FILE)
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
	DB_FILE = "bolt:///tmp/bolt.db?bucket=funcs"
	MQ_FILE = "bolt:///tmp/bolt_mq.db"
	PORT = 8080
	API_URL = "http://localhost:8080"
	setupServer()
	setupCli()
	testIntegration(t)
	teardown()
}

func TestIntegrationWithAuth(t *testing.T) {
	viper.Set("jwt_auth_key", "test")
	DB_FILE = "bolt:///tmp/bolt_auth.db?bucket=funcs"
	MQ_FILE = "bolt:///tmp/bolt_auth_mq.db"
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
