package lb

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/Sirupsen/logrus"
	"github.com/gin-gonic/gin"
	"github.com/iron-io/functions/api/datastore"
	"github.com/iron-io/functions/api/models"
	"github.com/iron-io/functions/api/mqs"
	"github.com/iron-io/functions/api/server"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

var testRouteJSON, _ = json.Marshal(&models.RouteWrapper{Route: &models.Route{Path: "/hello", AppName: "myapp", Image: "iron/hello"}})

func setLogBuffer() *bytes.Buffer {
	var buf bytes.Buffer
	buf.WriteByte('\n')
	logrus.SetOutput(&buf)
	gin.DefaultErrorWriter = &buf
	gin.DefaultWriter = &buf
	log.SetOutput(&buf)
	return &buf
}

func startNodes(t testing.TB, n int) ([]string, func()) {
	nodes := []string{}
	ctx, cancel := context.WithCancel(context.Background())

	fds, err := ioutil.TempFile(os.TempDir(), "functions-test-ds-lb")
	if err != nil {
		t.Fatalf("Failed to create temp ds db: %v", err)
	}

	fmq, err := ioutil.TempFile(os.TempDir(), "functions-test-ds-lb")
	if err != nil {
		t.Fatalf("Failed to create temp mq db: %v", err)
	}

	ds, err := datastore.New(fmt.Sprintf("bolt://%s?bucket=funcs", fds.Name()))
	if err != nil {
		t.Fatalf("Invalid DB url: %v", err)
	}

	mq, err := mqs.New(fmt.Sprintf("bolt://%s", fmq.Name()))
	if err != nil {
		t.Fatalf("Error on init MQ: %v", err)
	}

	var addr string
	for i := 0; i < n; i++ {
		addr = fmt.Sprintf("127.0.0.1:%d", 8090+i)
		funcServer := server.New(ctx, ds, mq, fmt.Sprintf("http://%s", addr))
		go funcServer.Start(ctx)
		nodes = append(nodes, addr)
	}

	time.Sleep(1 * time.Second)
	// create an test route '/myapp/hello' using the function 'iron/hello'
	http.Post(fmt.Sprintf("http://%s/v1/apps/myapp/routes", addr), "application/json", bytes.NewBuffer(testRouteJSON))

	return nodes, func() {
		ds.Close()
		mq.Close()
		fds.Close()
		fmq.Close()
		cancel()
	}
}

func TestConsistentHashReverseProxy(t *testing.T) {
	setLogBuffer()
	nodes, stop := startNodes(t, 3)
	defer stop()

	p := ConsistentHashReverseProxy(context.Background(), DefaultDirector, nodes)
	req, err := http.NewRequest("GET", "/v1/apps", nil)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	rec := httptest.NewRecorder()
	p.ServeHTTP(rec, req)
	b, _ := ioutil.ReadAll(rec.Body)
	fmt.Println(string(b))
}

func BenchmarkLB_3Nodes(b *testing.B) {
	b.StopTimer()
	setLogBuffer()
	nodes, stop := startNodes(b, 3)
	defer stop()
	p := ConsistentHashReverseProxy(context.Background(), DefaultDirector, nodes)
	rec := httptest.NewRecorder()
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		req, _ := http.NewRequest("GET", fmt.Sprintf("/r/myapp/hello?i=%d", i), nil)
		p.ServeHTTP(rec, req)
	}
}

func BenchmarkLB_15Nodes(b *testing.B) {
	b.StopTimer()
	setLogBuffer()
	nodes, stop := startNodes(b, 15)
	defer stop()
	p := ConsistentHashReverseProxy(context.Background(), DefaultDirector, nodes)
	rec := httptest.NewRecorder()
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		req, _ := http.NewRequest("GET", fmt.Sprintf("/r/myapp/hello?i=%d", i), nil)
		p.ServeHTTP(rec, req)
	}
}

func BenchmarkLB_30Nodes(b *testing.B) {
	b.StopTimer()
	setLogBuffer()
	nodes, stop := startNodes(b, 30)
	defer stop()
	p := ConsistentHashReverseProxy(context.Background(), DefaultDirector, nodes)
	rec := httptest.NewRecorder()
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		req, _ := http.NewRequest("GET", fmt.Sprintf("/r/myapp/hello?i=%d", i), nil)
		p.ServeHTTP(rec, req)
	}
}

func BenchmarkLB_100Nodes(b *testing.B) {
	b.StopTimer()
	setLogBuffer()
	nodes, stop := startNodes(b, 100)
	defer stop()
	p := ConsistentHashReverseProxy(context.Background(), DefaultDirector, nodes)
	rec := httptest.NewRecorder()
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		req, _ := http.NewRequest("GET", fmt.Sprintf("/r/myapp/hello?i=%d", i), nil)
		p.ServeHTTP(rec, req)
	}
}

func BenchmarkLB_200Nodes(b *testing.B) {
	b.StopTimer()
	setLogBuffer()
	nodes, stop := startNodes(b, 200)
	defer stop()
	p := ConsistentHashReverseProxy(context.Background(), DefaultDirector, nodes)
	rec := httptest.NewRecorder()
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		req, _ := http.NewRequest("GET", fmt.Sprintf("/r/myapp/hello?i=%d", i), nil)
		p.ServeHTTP(rec, req)
	}
}

func BenchmarkLB_500Nodes(b *testing.B) {
	b.StopTimer()
	setLogBuffer()
	nodes, stop := startNodes(b, 500)
	defer stop()
	p := ConsistentHashReverseProxy(context.Background(), DefaultDirector, nodes)
	rec := httptest.NewRecorder()
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		req, _ := http.NewRequest("GET", fmt.Sprintf("/r/myapp/hello?i=%d", i), nil)
		p.ServeHTTP(rec, req)
	}
}

func BenchmarkLB_1000Nodes(b *testing.B) {
	b.StopTimer()
	setLogBuffer()
	nodes, stop := startNodes(b, 1000)
	defer stop()
	p := ConsistentHashReverseProxy(context.Background(), DefaultDirector, nodes)
	rec := httptest.NewRecorder()
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		req, _ := http.NewRequest("GET", fmt.Sprintf("/r/myapp/hello?i=%d", i), nil)
		p.ServeHTTP(rec, req)
	}
}
