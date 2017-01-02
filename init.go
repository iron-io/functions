package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/gin-gonic/gin"
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
		logrus.WithError(err).Fatalln("")
	}
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.SetDefault(envLogLevel, "info")
	viper.SetDefault(envMQ, fmt.Sprintf("bolt://%s/data/worker_mq.db", cwd))
	viper.SetDefault(envDB, fmt.Sprintf("bolt://%s/data/bolt.db?bucket=funcs", cwd))
	viper.SetDefault(envPort, 8080)
	viper.SetDefault(envAPIURL, fmt.Sprintf("http://127.0.0.1:%d", viper.GetInt(envPort)))
	viper.AutomaticEnv() // picks up env vars automatically
	logLevel, err := logrus.ParseLevel(viper.GetString(envLogLevel))
	if err != nil {
		logrus.WithError(err).Fatalln("Invalid log level.")
	}
	logrus.SetLevel(logLevel)

	gin.SetMode(gin.ReleaseMode)
	if logLevel == logrus.DebugLevel {
		gin.SetMode(gin.DebugMode)
	}
}

func contextWithSignal(ctx context.Context, signals ...os.Signal) context.Context {
	ctx, halt := context.WithCancel(context.Background())
	c := make(chan os.Signal, 1)
	signal.Notify(c, signals...)
	go func() {
		<-c
		logrus.Info("Halting...")
		halt()
	}()
	return ctx
}
