package api

import (
	"fmt"
	"log"
	"os"
	"path"

	"github.com/Sirupsen/logrus"
	"github.com/gin-gonic/gin"
	"github.com/iron-io/functions/api/models"
	"github.com/iron-io/functions/api/server/datastore"
)

var Config = &models.Config{}
var Datastore models.Datastore = &datastore.Mock{}
var Router *gin.Engine

func extractFields(c *gin.Context) logrus.Fields {
	fields := logrus.Fields{"action": path.Base(c.HandlerName())}
	for _, param := range c.Params {
		fields[param.Key] = param.Value
	}
	return fields
}

func Start() {
	err := Config.Validate()
	if err != nil {
		logrus.WithError(err).Fatalln("Invalid config.")
	}
	log.Printf("config: %+v", Config)

	if Config.DatabaseURL == "" {
		cwd, _ := os.Getwd()
		Config.DatabaseURL = fmt.Sprintf("bolt://%s/bolt.db?bucket=funcs", cwd)
	}

	Datastore, err = datastore.New(Config.DatabaseURL)
	if err != nil {
		logrus.WithError(err).Fatalln("Invalid DB url.")
	}

	logrus.SetOutput(os.Stdout)
	logrus.SetLevel(logrus.DebugLevel)

	Router = gin.Default()

	Router.Use(func(c *gin.Context) {
		c.Set("log", logrus.WithFields(extractFields(c)))
		c.Next()
	})
}
