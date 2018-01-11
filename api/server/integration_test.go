// +build integration

package server

import (
	"context"
	"testing"
	"time"

	"github.com/iron-io/functions/fn/app"
	"github.com/iron-io/functions/api/models"
)

func TestIntegration(t *testing.T) {
	testIntegration(t)
}

func testIntegration(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	funcServer := NewFromEnv(ctx)

	go funcServer.Start(ctx)
	time.Sleep(10 * time.Second)

	fn := app.NewFn()
	err := fn.Run([]string{"fn", "apps", "l"})
	if err != nil {
		t.Error(err)
	}

	err = fn.Run([]string{"fn", "apps", "delete", "test"})
	if err != nil {
		t.Error(err)
	}

	// err = fn.Run([]string{"fn", "apps", "c", "test"})
	// if err != nil {
	// 	t.Error(err)
	// }

	filter := &models.AppFilter{}
	apps, err := funcServer.Datastore.GetApps(ctx, filter)

	if len(apps) != 1 {
		t.Error("fn apps create failed.")
	}

	err = fn.Run([]string{"fn", "invalid"})
	if err != nil {
		t.Error(err)
	}

	//	res, err := exec.Command("go", "run", "../../fn/main.go", "apps", "l").CombinedOutput()
	//	fmt.Println(res)
	//	fmt.Println(err)
	//	os.Remove(fnTestBin)
}
