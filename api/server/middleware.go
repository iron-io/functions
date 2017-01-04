// TODO: it would be nice to move these into the top level folder so people can use these with the "functions" package, eg: functions.AddMiddleware(...)
package server

import (
	"context"
	"net/http"

	"github.com/Sirupsen/logrus"
	"github.com/gin-gonic/gin"
	"github.com/iron-io/functions/api/models"
)

type Middleware interface {
	ServeHTTP(w http.ResponseWriter, r *http.Request, app *models.App) error
}

type MiddlewareFunc func(w http.ResponseWriter, r *http.Request, app *models.App) error

// ServeHTTP calls f(w, r).
func (f MiddlewareFunc) ServeHTTP(w http.ResponseWriter, r *http.Request, app *models.App) error {
	return f(w, r, app)
}

func (s *Server) middlewareWrapperFunc(ctx context.Context) gin.HandlerFunc {
	// todo: after calling middleware, test that the context is cancelled
	return func(c *gin.Context) {
		ctx = c.MustGet("ctx").(context.Context)
		ctx, cancel := context.WithCancel(ctx)
		// replace context so it's cancelable
		r := c.Request.WithContext(ctx)
		for _, l := range s.middlewares {
			// we could pass the CancelFunc into this method instead of returning err so the middleware can can cancel at this level
			err := l.ServeHTTP(c.Writer, r, nil)
			if err != nil {
				logrus.WithError(err).Warnln("Middleware error")
				cancel() // I know this is unecessary right now
				// todo: might be a good idea to check if anything is written yet, and if not, output the error: simpleError(err)
				// see: http://stackoverflow.com/questions/39415827/golang-http-check-if-responsewriter-has-been-written
				c.Abort()
				return
			}
		}
	}
}

// AddAppEndpoint adds an endpoints to /v1/apps/:app/x
func (s *Server) AddMiddleware(m Middleware) {
	s.middlewares = append(s.middlewares, m)
}

// AddAppEndpoint adds an endpoints to /v1/apps/:app/x
func (s *Server) AddMiddlewareFunc(m func(w http.ResponseWriter, r *http.Request, app *models.App) error) {
	s.AddMiddleware(MiddlewareFunc(m))
}
