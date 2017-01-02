package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/gin-gonic/gin"
	"github.com/iron-io/functions/api/datastore"
	"github.com/iron-io/functions/api/mqs"
	"github.com/iron-io/functions/api/server"
	"github.com/spf13/viper"
)

const (
	envLogLevel = "log_level"
	envMQ       = "mq_url"
	envDB       = "db_url"
	envPort     = "port" // be careful, Gin expects this variable to be "port"
	envAPIURL   = "api_url"
)

func init() {
	cwd, err := os.Getwd()
	if err != nil {
		log.WithError(err).Fatalln("")
	}
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.SetDefault(envLogLevel, "info")
	viper.SetDefault(envMQ, fmt.Sprintf("bolt://%s/data/worker_mq.db", cwd))
	viper.SetDefault(envDB, fmt.Sprintf("bolt://%s/data/bolt.db?bucket=funcs", cwd))
	viper.SetDefault(envPort, 8080)
	viper.SetDefault(envAPIURL, fmt.Sprintf("http://127.0.0.1:%d", viper.GetInt(envPort)))
	viper.AutomaticEnv() // picks up env vars automatically
	logLevel, err := log.ParseLevel(viper.GetString("log_level"))
	if err != nil {
		log.WithError(err).Fatalln("Invalid log level.")
	}
	log.SetLevel(logLevel)

	gin.SetMode(gin.ReleaseMode)
	if logLevel == log.DebugLevel {
		gin.SetMode(gin.DebugMode)
	}
}

func main() {
	ctx := context.Background()
	ds, err := datastore.New(viper.GetString(envDB))
	if err != nil {
		log.WithError(err).Fatalln("Invalid DB url.")
	}

	mq, err := mqs.New(viper.GetString(envMQ))
	if err != nil {
		log.WithError(err).Fatal("Error on init MQ")
	}

	apiURL := viper.GetString(envAPIURL)

	funcServer := server.New(ctx, ds, mq, apiURL)
	// Setup your custom extensions, listeners, etc here
	funcServer.Start()
}
