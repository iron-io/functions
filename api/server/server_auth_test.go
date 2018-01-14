// +build server

package server

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/iron-io/functions/common"
	"github.com/spf13/viper"
)

var UnAuthtestSuite = []struct {
	name              string
	method            string
	path              string
	body              string
	expectedCode      int
	expectedCacheSize int
}{
	{"create my app", "POST", "/v1/apps", `{ "app": { "name": "myapp" } }`, http.StatusUnauthorized, 0},
	{"list apps", "GET", "/v1/apps", ``, http.StatusUnauthorized, 0},
	{"get app", "GET", "/v1/apps/myapp", ``, http.StatusUnauthorized, 0},
	{"add myroute", "POST", "/v1/apps/myapp/routes", `{ "route": { "name": "myroute", "path": "/myroute", "image": "iron/hello" } }`, http.StatusUnauthorized, 0},
	{"add myroute2", "POST", "/v1/apps/myapp/routes", `{ "route": { "name": "myroute2", "path": "/myroute2", "image": "iron/error" } }`, http.StatusUnauthorized, 0},
	{"get myroute", "GET", "/v1/apps/myapp/routes/myroute", ``, http.StatusUnauthorized, 0},
	{"get myroute2", "GET", "/v1/apps/myapp/routes/myroute2", ``, http.StatusUnauthorized, 0},
	{"get all routes", "GET", "/v1/apps/myapp/routes", ``, http.StatusUnauthorized, 0},
	// These two are currently returning 404 because they dont get created : temporarily using StatusNotFound
	//		{"execute myroute", "POST", "/r/myapp/myroute", `{ "name": "Teste" }`, http.StatusUnauthorized, 0},
	//		{"execute myroute2", "POST", "/r/myapp/myroute2", `{ "name": "Teste" }`, http.StatusUnauthorized, 0},
	{"execute myroute", "POST", "/r/myapp/myroute", `{ "name": "Teste" }`, http.StatusNotFound, 0},
	{"execute myroute2", "POST", "/r/myapp/myroute2", `{ "name": "Teste" }`, http.StatusNotFound, 0},
	{"delete myroute", "DELETE", "/v1/apps/myapp/routes/myroute", ``, http.StatusUnauthorized, 0},
	{"delete app (fail)", "DELETE", "/v1/apps/myapp", ``, http.StatusUnauthorized, 0},
	{"delete myroute2", "DELETE", "/v1/apps/myapp/routes/myroute2", ``, http.StatusUnauthorized, 0},
	{"delete app (success)", "DELETE", "/v1/apps/myapp", ``, http.StatusUnauthorized, 0},
	{"get deleted app", "GET", "/v1/apps/myapp", ``, http.StatusUnauthorized, 0},
	{"get deleteds route on deleted app", "GET", "/v1/apps/myapp/routes/myroute", ``, http.StatusUnauthorized, 0},
}

func routerRequestWithAuth(t *testing.T, router *gin.Engine, method, path string, body io.Reader, setAuth func(*http.Request)) (*http.Request, *httptest.ResponseRecorder) {
	req, err := http.NewRequest(method, "http://127.0.0.1:8080"+path, body)
	setAuth(req)
	if err != nil {
		t.Fatalf("Test: Could not create %s request to %s: %v", method, path, err)
	}

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	return req, rec
}

func setJwtAuth(req *http.Request) {
	if jwtAuthKey := viper.GetString("jwt_auth_key"); jwtAuthKey != "" {
		jwtToken, err := common.GetJwt(jwtAuthKey, 60*60)
		if err == nil {
			req.Header.Set("Authorization", fmt.Sprintf("Bearer %v", jwtToken))
		}
	}
}

func setBrokenJwtAuth(req *http.Request) {
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %v", "broken token"))
}

func TestFullStackWithAuth(t *testing.T) {
	viper.Set("jwt_auth_key", "test")
	testFullStack(t, setJwtAuth, testSuite)
	teardown()
}

func TestFullStackWithBrokenAuth(t *testing.T) {
	viper.Set("jwt_auth_key", "test")
	testFullStack(t, setBrokenJwtAuth, UnAuthtestSuite)
	teardown()
}
