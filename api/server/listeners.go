package server

import (
	"context"
	"github.com/iron-io/functions/api/ifaces"
	"github.com/iron-io/functions/api/models"
)

// AddAppListener adds a listener that will be notified on App changes.
func (s *Server) AddAppListener(listener ifaces.AppListener) {
	s.AppListeners = append(s.AppListeners, listener)
}

func (s *Server) FireBeforeAppCreate(ctx context.Context, app *models.App) error {
	for _, l := range s.AppListeners {
		err := l.BeforeAppCreate(ctx, app)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *Server) FireAfterAppCreate(ctx context.Context, app *models.App) error {
	for _, l := range s.AppListeners {
		err := l.AfterAppCreate(ctx, app)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *Server) FireBeforeAppUpdate(ctx context.Context, app *models.App) error {
	for _, l := range s.AppListeners {
		err := l.BeforeAppUpdate(ctx, app)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *Server) FireAfterAppUpdate(ctx context.Context, app *models.App) error {
	for _, l := range s.AppListeners {
		err := l.AfterAppUpdate(ctx, app)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *Server) FireBeforeAppDelete(ctx context.Context, appName string) error {
	for _, l := range s.AppListeners {
		err := l.BeforeAppDelete(ctx, appName)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *Server) FireAfterAppDelete(ctx context.Context, appName string) error {
	for _, l := range s.AppListeners {
		err := l.AfterAppDelete(ctx, appName)
		if err != nil {
			return err
		}
	}
	return nil
}
