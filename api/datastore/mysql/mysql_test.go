package mysql

import (
	"bytes"
	"database/sql"
	"fmt"
	"net/url"
	"os/exec"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/iron-io/functions/api/datastore/internal/datastoretest"
)

const tmpMysql = "mysql:secret@tcp(localhost:3306)/funcs"

func prepareMysqlTest(logf, fatalf func(string, ...interface{})) (func(), func()) {
	fmt.Println("initializing mysql for test")
	tryRun(logf, "remove old mysql container", exec.Command("docker", "rm", "-f", "iron-mysql-test"))
	mustRun(fatalf, "start mysql container", exec.Command("docker", "run", "--name", "iron-mysql-test", "-p", "3307:3306", "-d", "mysql"))

	maxRetries := 3
	currentRetry := 0
	wait := 1 * time.Second
	var db *sql.DB
	var err error
	for {
		db, err = sql.Open("mysql", "mysql:secret@/")
		if err != nil {
			if currentRetry == maxRetries {
				fatalf("failed to connect to mysql after %d retries", maxRetries)
				break
			}
			fmt.Println("failed to connect to mysql:", err)
			fmt.Println("retrying in:", wait)
			time.Sleep(wait)
			wait = 2 * wait
			currentRetry++
			continue
		} else {
			break
		}
	}
	_, err = db.Exec(`CREATE DATABASE funcs;`)
	if err != nil {
		fmt.Println("failed to create database:", err)
	}
	_, err = db.Exec(`GRANT ALL PRIVILEGES ON funcs.* TO mysql@localhost WITH GRANT OPTION;`)
	if err != nil {
		fmt.Println("failed to grant priviledges to user 'mysql:", err)
	}

	fmt.Println("mysql for test ready")
	return func() {
			db, err := sql.Open("mysql", "mysql:secret@/funcs")
			if err != nil {
				fatalf("failed to connect for truncation: %s\n", err)
			}
			for _, table := range []string{"routes", "apps", "extras"} {
				_, err = db.Exec(`TRUNCATE TABLE ` + table)
				if err != nil {
					fatalf("failed to truncate table %q: %s\n", table, err)
				}
			}
		},
		func() {
			tryRun(logf, "stop mysql container", exec.Command("docker", "rm", "-f", "iron-mysql-test"))
		}
}

func TestDatastore(t *testing.T) {
	_, close := prepareMysqlTest(t.Logf, t.Fatalf)
	defer close()

	u, err := url.Parse(tmpMysql)
	if err != nil {
		t.Fatalf("failed to parse url: %s\n", err)
	}
	ds, err := New(u)
	if err != nil {
		t.Fatalf("failed to create mysql datastore: %s\n", err)
	}

	datastoretest.Test(t, ds)
}

func tryRun(logf func(string, ...interface{}), desc string, cmd *exec.Cmd) {
	var b bytes.Buffer
	cmd.Stderr = &b
	if err := cmd.Run(); err != nil {
		logf("failed to %s: %s", desc, b.String())
	}
}

func mustRun(fatalf func(string, ...interface{}), desc string, cmd *exec.Cmd) {
	var b bytes.Buffer
	cmd.Stderr = &b
	if err := cmd.Run(); err != nil {
		fatalf("failed to %s: %s", desc, b.String())
	}
}
