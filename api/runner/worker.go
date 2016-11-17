package runner

import (
	"context"
	"sync"

	"github.com/iron-io/runner/drivers"
)

type TaskRequest struct {
	Ctx      context.Context
	Config   *Config
	Response chan TaskResponse
}

type TaskResponse struct {
	Result drivers.RunResult
	Err    error
}

func StartWorkers(ctx context.Context, rnr *Runner, tasks <-chan TaskRequest) {
	var wg sync.WaitGroup

	wg.Add(1)
	go func(ctx context.Context) {
		defer wg.Done()
		for {
			select {
			case <-ctx.Done():
				return
			case task := <-tasks:
				wg.Add(1)
				go func(task TaskRequest) {
					defer wg.Done()
					result, err := rnr.Run(task.Ctx, task.Config)
					select {
					case task.Response <- TaskResponse{result, err}:
						close(task.Response)
					default:
					}
				}(task)
			}
		}
	}(ctx)

	<-ctx.Done()
	wg.Wait()
}
