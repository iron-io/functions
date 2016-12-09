package server

import (
	"context"
	"github.com/iron-io/functions/api/models"
)

type RunnerListener interface {
	// BeforeABeforeRunppCreate called before a function run
	BeforeRun(ctx context.Context, route *models.Route) error
	// AfterRun called after a function run
	AfterRun(ctx context.Context, route *models.Route) error
}

// AddRunListeners adds a listener that will be fired before and after a function run.
func (s *Server) AddRunnerListener(listener RunnerListener) {
	s.runnerListeners = append(s.runnerListeners, listener)
}

func (s *Server) FireBeforeRun(ctx context.Context, route *models.Route) error {
	for _, l := range s.runnerListeners {
		err := l.BeforeRun(ctx, route)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *Server) FireAfterRun(ctx context.Context, route *models.Route) error {
	for _, l := range s.runnerListeners {
		err := l.AfterRun(ctx, route)
		if err != nil {
			return err
		}
	}
	return nil
}
