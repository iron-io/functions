package runner

import (
	"context"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/iron-io/functions/api/runner"
	"github.com/iron-io/runner/drivers"
)

type TaskRequest struct {
	Ctx      *gin.Context
	Config   *runner.Config
	Response chan TaskResponse
}

type TaskResponse struct {
	Result drivers.RunResult
	Err    error
}

func Worker(ctx context.Context, n int) {
	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(ctx context.Context) {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					break
				case task := <-tasks:
					result, err := Api.Runner.Run(task.Ctx, task.Config)

					select {
					case task.Response <- TaskResponse{result, err}:
						close(task.Response)
					default:
					}
				}
			}
		}(ctx)
	}

	wg.Wait()
	<-ctx.Done()
}
