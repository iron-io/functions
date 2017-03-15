package server

import (
	"bytes"
	"github.com/iron-io/functions/api/auth"
	"github.com/iron-io/functions/api/datastore"
	"github.com/iron-io/functions/api/models"
	"github.com/iron-io/functions/api/mqs"
	"net/http"
	"strings"
	"testing"
)

func TestServer_DockerHandle(t *testing.T) {

	buf := setLogBuffer()
	tasks := mockTasksConduit()
	defer close(tasks)

	for i, test := range []struct {
		ds            models.Datastore
		path          string
		body          string
		expectedCode  int
		expectedError error
	}{
		// errors
		{datastore.NewMock(), "/v1/docker/login", ``, http.StatusBadRequest, models.ErrInvalidJSON},
		{datastore.NewMock(), "/v1/docker/login", `{}`, http.StatusBadRequest, models.ErrDockerCredsMissing},
		{datastore.NewMock(), "/v1/docker/login", `{ "auth": "someInvalidData"}`, http.StatusBadRequest, models.ErrDockerCredsInvalid},

		// success
		{datastore.NewMock(), "/v1/docker/login", `{ "auth": "eyJ1c2VybmFtZSI6InRlc3ROYW1lIiwicGFzc3dvcmQiOiJwYXNzd29yZCIsImVtYWlsIjoiZW1haWwiLCJzZXJ2ZXJhZGRyZXNzIjoidXJsIn0=" }`, http.StatusOK, nil},
	} {
		rnr, cancel := testRunner(t)
		srv := testServer(test.ds, auth.NewDockerMock(test.ds), &mqs.Mock{}, rnr, tasks)
		router := srv.Router

		body := bytes.NewBuffer([]byte(test.body))
		_, rec := routerRequest(t, router, "POST", test.path, body)

		if rec.Code != test.expectedCode {
			t.Log(buf.String())
			t.Errorf("Test %d: Expected status code to be %d but was %d",
				i, test.expectedCode, rec.Code)
		}

		if test.expectedError != nil {
			resp := getErrorResponse(t, rec)

			if !strings.Contains(resp.Error.Message, test.expectedError.Error()) {
				t.Log(buf.String())
				t.Errorf("Test %d: Expected error message to have `%s`",
					i, test.expectedError.Error())
			}
		}
		cancel()
	}

}
