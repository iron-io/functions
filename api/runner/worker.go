package runner

import (
	"context"
	"sync"

	"github.com/iron-io/runner/drivers"
)

type taskPriority int

func (t taskPriority) String() string {
	if t == High {
		return "High"
	}
	return "Low"
}

const (
	High taskPriority = iota
	Low
)

type TaskRequest struct {
	Prio     taskPriority
	Ctx      context.Context
	Config   *Config
	Response chan TaskResponse
}

type TaskResponse struct {
	Result drivers.RunResult
	Err    error
}

// StartWorkers handle incoming tasks and spawns self-regulating container
// workers. Internally it works with a priority queue, used to favor the
// execution of Sync tasks over Async ones.
func StartWorkers(ctx context.Context, rnr *Runner, tasks <-chan TaskRequest) {
	var wg sync.WaitGroup
	prioQ := &priotasksqueue{cond: &sync.Cond{L: &sync.Mutex{}}}

	go func() {
		for {
			select {
			case <-ctx.Done():
				prioQ.close()
				return
			case task := <-tasks:
				if task.Prio == High {
					prioQ.pushHigh(task)
				} else if task.Prio == Low {
					prioQ.pushLow(task)
				}
			}
		}
	}()

	wg.Add(1)
	go func(ctx context.Context) {
		defer wg.Done()
		for {
			task := prioQ.shift()
			if task == nil {
				return
			}

			wg.Add(1)
			go func(task TaskRequest) {
				defer wg.Done()
				result, err := rnr.Run(task.Ctx, task.Config)
				select {
				case task.Response <- TaskResponse{result, err}:
					close(task.Response)
				default:
				}
			}(*task)
		}
	}(ctx)

	<-ctx.Done()
	wg.Wait()
}

type priotasksqueue struct {
	high, low []TaskRequest
	cond      *sync.Cond
	closed    bool
}

func (p *priotasksqueue) shift() *TaskRequest {
	p.cond.L.Lock()
	if len(p.high) == 0 && len(p.low) == 0 && !p.closed {
		p.cond.Wait()
	}
	if p.closed {
		p.cond.L.Unlock()
		return nil
	}

	var v TaskRequest
	if len(p.high) > 0 {
		v, p.high = p.high[0], p.high[1:]
	} else {
		v, p.low = p.low[0], p.low[1:]
	}
	p.cond.L.Unlock()
	return &v
}

func (p *priotasksqueue) pushHigh(task TaskRequest) {
	p.cond.L.Lock()
	p.high = append(p.high, task)
	p.cond.L.Unlock()
	p.cond.Signal()
}

func (p *priotasksqueue) pushLow(task TaskRequest) {
	p.cond.L.Lock()
	p.low = append(p.low, task)
	p.cond.L.Unlock()
	p.cond.Signal()
}

func (p *priotasksqueue) close() {
	p.cond.L.Lock()
	p.closed = true
	p.cond.L.Unlock()
	p.cond.Signal()
}
