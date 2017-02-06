package postgres_test

import (
	"bytes"
	"database/sql"
	"fmt"
	"os/exec"
	"testing"
	"time"

	"github.com/iron-io/functions/api/datastore"
	"github.com/iron-io/functions/api/datastore/internal/datastoretest"
)

const tmpPostgres = "postgres://postgres@127.0.0.1:15432/funcs?sslmode=disable"

func preparePostgresTest(t *testing.T) func() {
	fmt.Println("initializing postgres for test")
	tryRun(t, "remove old postgres container", exec.Command("docker", "rm", "-f", "iron-postgres-test"))
	mustRun(t, "start postgres container", exec.Command("docker", "run", "--name", "iron-postgres-test", "-p", "15432:5432", "-d", "postgres"))

	wait := 1 * time.Second
	for {
		db, err := sql.Open("postgres", "postgres://postgres@127.0.0.1:15432?sslmode=disable")
		if err != nil {
			fmt.Println("failed to connect to postgres:", err)
			fmt.Println("retrying in:", wait)
			time.Sleep(wait)
			wait = 2 * wait
			continue
		}

		_, err = db.Exec(`CREATE DATABASE funcs;`)
		if err != nil {
			fmt.Println("failed to create database:", err)
			fmt.Println("retrying in:", wait)
			time.Sleep(wait)
			wait = 2 * wait
			continue
		}
		_, err = db.Exec(`GRANT ALL PRIVILEGES ON DATABASE funcs TO postgres;`)
		if err == nil {
			break
		}
		fmt.Println("failed to grant privileges:", err)
		fmt.Println("retrying in:", wait)
		time.Sleep(wait)
		wait = 2 * wait
	}
	fmt.Println("postgres for test ready")
	return func() {
		tryRun(t, "stop postgres container", exec.Command("docker", "rm", "-f", "iron-postgres-test"))
	}
}

func TestPostgres(t *testing.T) {
	close := preparePostgresTest(t)
	defer close()

	datastoretest.New(datastore.New(tmpPostgres))(t)
}

func tryRun(t *testing.T, desc string, cmd *exec.Cmd) {
	var b bytes.Buffer
	cmd.Stderr = &b
	if err := cmd.Run(); err != nil {
		t.Logf("failed to %s: %s", desc, b.String())
	}
}

func mustRun(t *testing.T, desc string, cmd *exec.Cmd) {
	var b bytes.Buffer
	cmd.Stderr = &b
	if err := cmd.Run(); err != nil {
		t.Fatalf("failed to %s: %s", desc, b.String())
	}
}