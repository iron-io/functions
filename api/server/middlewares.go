package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/iron-io/functions/api/models"
	"github.com/iron-io/functions/common"
	"github.com/spf13/viper"
)

func SetupJwtAuth(funcServer *Server) {
	// Add default JWT AUTH if env variable set
	if jwtAuthKey := viper.GetString("jwt_auth_key"); jwtAuthKey != "" {
		funcServer.AddMiddlewareFunc(func(ctx MiddlewareContext, w http.ResponseWriter, r *http.Request, app *models.App) error {
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

func (h *JwtAuthMiddleware) Serve(ctx MiddlewareContext, w http.ResponseWriter, r *http.Request, app *models.App) error {
	fmt.Println("JwtAuthMiddleware called")
	jwtAuthKey := viper.GetString("jwt_auth_key")

	if err := common.AuthJwt(jwtAuthKey, r); err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		m := map[string]string{"error": "Invalid API Authorization token."}
		json.NewEncoder(w).Encode(m)
		return errors.New("Invalid API authorization token.")
	}

	fmt.Println("auth succeeded!")
	return nil
}
