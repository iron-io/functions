package main

import "C"

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/iron-io/functions/api/models"
)

var RunnerListener commonlog

func init() {
	f, err := os.Create("access.log")
	if err != nil {
		panic(err)
	}
	RunnerListener = commonlog{
		log: f,
	}
}

type commonlog struct {
	log io.Writer
}

func (c *commonlog) BeforeDispatch(ctx context.Context, route *models.Route) error {
	return nil
}

func (c *commonlog) AfterDispatch(ctx context.Context, route *models.Route) error {
	req := fmt.Sprint(`"`, route.Type, " ", route.Path, " HTTP/1.0", `"`)
	fmt.Fprintln(c.log, "functions", "-", "-", time.Now().Format("02/Jan/2006:15:04:05 -0700"), req, 200, "-")
	return nil
}
