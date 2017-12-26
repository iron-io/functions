// +build full_stack

package server

import (
	"context"
	"testing"
	"os"
	"os/exec"
	"path"
	"fmt"
	"time"
)

func TestIntegration(t *testing.T) {
	testIntegration(t)
}

func testIntegration(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	funcServer := NewFromEnv(ctx)

	go funcServer.Start(ctx)

	fnTestBin := path.Join(os.TempDir(), "fn-test")

	time.Sleep(5 * time.Second)
	res, err := exec.Command("go", "run", "../../fn/main.go", "apps", "l").CombinedOutput()
	t.Error(string(res))
	fmt.Println(res)
	fmt.Println(err)
	os.Remove(fnTestBin)
}
