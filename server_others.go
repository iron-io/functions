// +build !go1.8

package main

import (
	"context"

	"github.com/ccirello/supervisor"
	"github.com/iron-io/functions/api/models"
	"github.com/iron-io/functions/api/runner"
	"github.com/iron-io/functions/api/runner/task"
	"github.com/iron-io/functions/api/server"
)

func attachHTTP(
	svr *supervisor.Supervisor,
	ctx context.Context,
	ds models.Datastore,
	mq models.MessageQueue,
	rnr *runner.Runner,
	tasks chan task.Request,
	enqueue models.Enqueue,
) {
	svr.AddFunc(func(ctx context.Context) {
		srv := server.New(ctx, ds, mq, rnr, tasks, server.DefaultEnqueue)
		srv.Run()
		<-ctx.Done()
	})
}
