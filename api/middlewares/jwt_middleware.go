package middlewares

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/iron-io/functions/api/models"
	"github.com/iron-io/functions/api/server"
	"github.com/spf13/viper"
)

func SetupJwtAuth(funcServer *server.Server) {
	// Add default JWT AUTH if env variable set
	if jwt_auth_key := viper.Get("jwt_auth_key"); jwt_auth_key != nil {
		funcServer.AddMiddlewareFunc(func(ctx server.MiddlewareContext, w http.ResponseWriter, r *http.Request, app *models.App) error {
			start := time.Now()
			fmt.Println("JwtAuthMiddlewareFunc called at:", start)
			ctx.Next()
			fmt.Println("Duration:", (time.Now().Sub(start)))
			return nil
		})
		funcServer.AddMiddleware(&JwtAuthMiddleware{})
	}
}

type JwtAuthMiddleware struct {
}

func (h *JwtAuthMiddleware) Serve(ctx server.MiddlewareContext, w http.ResponseWriter, r *http.Request, app *models.App) error {
	fmt.Println("JwtAuthMiddleware called")

	// check auth header
	tokenHeader := strings.SplitN(r.Header.Get("Authorization"), " ", 3)
	if len(tokenHeader) < 2 || tokenHeader[1] != "KlaatuBaradaNikto" {
		w.WriteHeader(http.StatusUnauthorized)
		m := map[string]string{"error": "Invalid Authorization token. Sorry!"}
		json.NewEncoder(w).Encode(m)
		return errors.New("Invalid authorization token.")
	}
	fmt.Println("auth succeeded!")
	return nil
}
