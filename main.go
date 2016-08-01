package main

import (
	"os"

	"github.com/hollowcms/hollow/api/router"
	"github.com/iron-io/functions/api/server"
)

func main() {
	api.Config.DatabaseURL = os.Getenv("DB")
	api.Start()
	router.Start(api.Router)
	api.Router.Run()
}
