package main

import "C"

import (
	"context"
	"io"
	"os"

	"github.com/davecgh/go-spew/spew"
	"github.com/iron-io/functions/api/models"
)

var RunnerListener requestdump

func init() {
	f, err := os.Create("dump.log")
	if err != nil {
		panic(err)
	}
	RunnerListener = requestdump{
		log: f,
	}
}

type requestdump struct {
	log io.Writer
}

func (c *requestdump) BeforeDispatch(ctx context.Context, route *models.Route) error {
	return nil
}

func (c *requestdump) AfterDispatch(ctx context.Context, route *models.Route) error {
	spew.Fdump(c.log, route)
	return nil
}
