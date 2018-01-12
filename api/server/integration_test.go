// +build integration

package server

import (
	"context"
	"testing"
	"time"
	"os"

	"github.com/iron-io/functions/fn/app"
	"github.com/iron-io/functions/api/models"
	"github.com/spf13/viper"
)

var DB_FILE = "bolt:///tmp/bolt.db?bucket=funcs"

func TestIntegration(t *testing.T) {
	viper.Set(EnvDBURL, DB_FILE)
	defer os.Remove(DB_FILE)
	testIntegration(t)
}

func TestIntegrationWithAuth(t *testing.T) {
	viper.Set("jwt_auth_key", "test")
	viper.Set(EnvDBURL, DB_FILE)
	defer os.Remove(DB_FILE)
	testIntegration(t)
}

func testIntegration(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	funcServer := NewFromEnv(ctx)

	go funcServer.Start(ctx)
	time.Sleep(3 * time.Second)

	fn := app.NewFn()

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
	apps, err := funcServer.Datastore.GetApps(ctx, filter)

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

	apps, err = funcServer.Datastore.GetApps(ctx, filter)

	if len(apps) != 0 {
		t.Error("fn apps delete failed.")
	}
}
