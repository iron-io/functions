package server

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMiddleware(t *testing.T) {
	req, _ := http.NewRequest("GET", "/", nil)

	mc := MiddlewareContext{
		Context:        context.Background(),
		Request:        req,
		ResponseWriter: httptest.NewRecorder(),

		AppName:   "test",
		RoutePath: "test",
	}

	runMiddlewares(&mc, []Middleware{
		Middleware(func(c *MiddlewareContext) {
			c.AppName += "e"
			c.RoutePath += "e"
		}),

		Middleware(func(c *MiddlewareContext) {
			c.AppName += "d"
			c.RoutePath += "d"
		}),
	})

	if mc.AppName != "tested" || mc.RoutePath != "tested" {
		t.Fatal("Expected AppName and RoutePath to be 'tested'")
	}
}
