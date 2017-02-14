package postgres

import (
	"bytes"
	"database/sql"
	"fmt"
	"net/url"
	"os/exec"
	"testing"
	"time"

	"github.com/iron-io/functions/api/datastore/internal/datastoretest"
	"github.com/iron-io/functions/api/models"
)

const tmpPostgres = "postgres://postgres@127.0.0.1:15432/funcs?sslmode=disable"

func preparePostgresTest(logf, fatalf func(string, ...interface{})) (func(), func()) {
	fmt.Println("initializing postgres for test")
	tryRun(logf, "remove old postgres container", exec.Command("docker", "rm", "-f", "iron-postgres-test"))
	mustRun(fatalf, "start postgres container", exec.Command("docker", "run", "--name", "iron-postgres-test", "-p", "15432:5432", "-d", "postgres"))

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
		db, err := sql.Open("postgres", tmpPostgres)
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
		tryRun(logf, "stop postgres container", exec.Command("docker", "rm", "-f", "iron-postgres-test"))
	}
}

func TestPathRegexp(t *testing.T) {
	for _, test := range []struct{ path, expected string }{
		{`/`, `^(/|/\*)$`},
		{`/blogs`, `^/(\*|((:|blogs)))$`},
		{`/blogs/`, `^/(\*|((:|blogs)/))$`},
		{`/blogs/123`, `^/(\*|((:|blogs)/(\*|((:|123)))))$`},
		{`/blogs/123/comments`, `^/(\*|((:|blogs)/(\*|((:|123)/(\*|((:|comments)))))))$`},
		{`/blogs/123/comments/456`, `^/(\*|((:|blogs)/(\*|((:|123)/(\*|((:|comments)/(\*|((:|456)))))))))$`},
	} {
		got := pathRegexp(test.path)
		if got != test.expected {
			t.Errorf("%q - expected %q but got %q", test.path, test.expected, got)
		}
	}
}

func TestConflictRegexp(t *testing.T) {
	for _, test := range []struct{ input, expected string }{
		{`/`, `^/$`},
		{`/:`, `^/(([^:](.*))|(:))$`},
		{`/*`, `^/((.+))$`},
		{`/test`, `^/(\*|:(.*)|((:|test)))$`},
		{`/test/`, `^/(\*|:(.*)|((:|test)/))$`},
		{`/test/test`, `^/(\*|:(.*)|((:|test)/(\*|:(.*)|((:|test)))))$`},
		{`/test/:`, `^/(\*|:(.*)|((:|test)/(([^:](.*))|(:))))$`},
		{`/test/*`, `^/(\*|:(.*)|((:|test)/((.+))))$`},
		{`/test/:/`, `^/(\*|:(.*)|((:|test)/(([^:](.*))|(:/))))$`},
		{`/test/:/test`, `^/(\*|:(.*)|((:|test)/(([^:](.*))|(:/(\*|:(.*)|((:|test)))))))$`},
	} {
		got := conflictRegexp(test.input)
		if got != test.expected {
			t.Errorf("%q - expected %q but got %q", test.input, test.expected, got)
		}
	}
}

func TestDatastore(t *testing.T) {
	_, close := preparePostgresTest(t.Logf, t.Fatalf)
	defer close()

	u, err := url.Parse(tmpPostgres)
	if err != nil {
		t.Fatalf("failed to parse url:", err)
	}
	ds, err := New(u)
	if err != nil {
		t.Fatalf("failed to create postgres datastore:", err)
	}

	datastoretest.Test(t, ds)
}

// Note: Running all of these at once may exceed the default timeout of 10m.
func BenchmarkDatastore(b *testing.B) {
	u, err := url.Parse(tmpPostgres)
	if err != nil {
		b.Fatalf("failed to parse url:", err)
	}

	truncate, close := preparePostgresTest(b.Logf, b.Fatalf)
	defer close()

	ds, err := New(u)
	if err != nil {
		b.Fatalf("failed to create postgres datastore:", err)
	}

	datastoretest.Benchmark(b, func(b *testing.B) (models.Datastore, func()) {
		truncate()
		return ds, func(){}
	})
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
