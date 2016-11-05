// Copyright 2016 Iron.io
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package runner

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/gin-gonic/gin"
	"github.com/iron-io/functions/api/models"
	"github.com/iron-io/functions/api/mqs"
	"github.com/iron-io/runner/drivers"
)

func setLogBuffer() *bytes.Buffer {
	var buf bytes.Buffer
	buf.WriteByte('\n')
	logrus.SetOutput(&buf)
	gin.DefaultErrorWriter = &buf
	gin.DefaultWriter = &buf
	log.SetOutput(&buf)
	return &buf
}

func getMockTask() models.Task {
	priority := int32(0)
	image := fmt.Sprintf("Image-%d", rand.Int31()%1000)
	task := &models.Task{}
	task.Image = &image
	task.ID = fmt.Sprintf("ID-%d", rand.Int31()%1000)
	task.AppName = fmt.Sprintf("RouteName-%d", rand.Int31()%1000)
	task.Priority = &priority
	return *task
}

func getTestServer(mockTasks []*models.Task) *httptest.Server {
	ctx := context.TODO()

	mq, err := mqs.New("memory://test")
	if err != nil {
		panic(err)
	}

	for _, mt := range mockTasks {
		mq.Push(ctx, mt)
	}

	getHandler := func(c *gin.Context) {
		task, err := mq.Reserve(ctx)
		if err != nil {
			logrus.WithError(err)
			c.JSON(http.StatusInternalServerError, err)
			return
		}
		c.JSON(http.StatusAccepted, task)
	}

	delHandler := func(c *gin.Context) {
		body, err := ioutil.ReadAll(c.Request.Body)
		if err != nil {
			logrus.WithError(err)
			c.JSON(http.StatusInternalServerError, err.Error())
			return
		}
		var task models.Task
		if err = json.Unmarshal(body, &task); err != nil {
			logrus.WithError(err)
			c.JSON(http.StatusInternalServerError, err.Error())
			return
		}

		if err := mq.Delete(ctx, &task); err != nil {
			logrus.WithError(err)
			c.JSON(http.StatusInternalServerError, err.Error())
			return
		}
		c.JSON(http.StatusAccepted, task)
	}

	r := gin.Default()
	r.GET("/tasks", getHandler)
	r.DELETE("/tasks", delHandler)
	return httptest.NewServer(r)
}

var helloImage = "iron/hello"

func TestRunTask(t *testing.T) {
	mockTask := getMockTask()
	mockTask.Image = &helloImage

	result, err := runTask(context.Background(), &mockTask)
	if err != nil {
		t.Error(err)
	}

	if result.Status() != "success" {
		t.Errorf("TestRunTask failed to execute runTask")
	}
}

func TestGetTask(t *testing.T) {
	buf := setLogBuffer()
	mockTask := getMockTask()

	ts := getTestServer([]*models.Task{&mockTask})
	defer ts.Close()

	url := ts.URL + "/tasks"
	task, err := getTask(context.Background(), url)
	if err != nil {
		t.Log(buf.String())
		t.Error("expected no error, got", err)
	}
	if task.ID != mockTask.ID {
		t.Log(buf.String())
		t.Errorf("expected task ID '%s', got '%s'", task.ID, mockTask.ID)
	}
}

func TestGetTaskError(t *testing.T) {
	buf := setLogBuffer()

	tests := []map[string]interface{}{
		{
			"url":   "/invalid",
			"task":  getMockTask(),
			"error": "invalid character 'p' after top-level value",
		},
	}

	var tasks []*models.Task
	for _, v := range tests {
		task := v["task"].(models.Task)
		tasks = append(tasks, &task)
	}

	ts := getTestServer(tasks)
	defer ts.Close()

	for i, test := range tests {
		url := ts.URL + test["url"].(string)
		_, err := getTask(context.Background(), url)
		if err == nil {
			t.Log(buf.String())
			t.Errorf("expected error '%s'", test["error"].(string))
		}
		if err.Error() != test["error"].(string) {
			t.Log(buf.String())
			t.Errorf("test %d: expected error '%s', got '%s'", i, test["error"].(string), err)
		}
	}
}

func TestDeleteTask(t *testing.T) {
	buf := setLogBuffer()
	mockTask := getMockTask()

	ts := getTestServer([]*models.Task{&mockTask})
	defer ts.Close()

	url := ts.URL + "/tasks"
	err := deleteTask(url, &mockTask)
	if err == nil {
		t.Log(buf.String())
		t.Error("expected error 'Not reserver', got", err)
	}

	_, err = getTask(context.Background(), url)
	if err != nil {
		t.Log(buf.String())
		t.Error("expected no error, got", err)
	}

	err = deleteTask(url, &mockTask)
	if err != nil {
		t.Log(buf.String())
		t.Error("expected no error, got", err)
	}
}

func TestTasksrvURL(t *testing.T) {
	tests := []struct {
		in, out string
	}{
		{"//localhost:8081", "http://localhost:8081/tasks"},
		{"//localhost:8081/", "http://localhost:8081/tasks"},
		{"http://localhost:8081", "http://localhost:8081/tasks"},
		{"http://localhost:8081/", "http://localhost:8081/tasks"},
		{"http://localhost:8081/endpoint", "http://localhost:8081/endpoint"},
	}

	for _, tt := range tests {
		if got, _ := tasksrvURL(tt.in); got != tt.out {
			t.Errorf("tasksrv: %s\texpected: %s\tgot: %s\t", tt.in, tt.out, got)
		}
	}
}

func TestAsyncRunnersGracefulShutdown(t *testing.T) {
	buf := setLogBuffer()
	mockTask := getMockTask()
	ts := getTestServer([]*models.Task{&mockTask})
	defer ts.Close()

	ctx, _ := context.WithTimeout(context.Background(), 2*time.Second)
	var wg sync.WaitGroup
	wg.Add(1)
	go startAsyncRunners(ctx, &wg, 0, ts.URL+"/tasks", func(ctx context.Context, task *models.Task) (drivers.RunResult, error) {
		return nil, nil
	})
	wg.Wait()

	if err := ctx.Err(); err != context.DeadlineExceeded {
		t.Log(buf.String())
		t.Errorf("async runners stopped unexpectedly. context error: %v", err)
	}
}
