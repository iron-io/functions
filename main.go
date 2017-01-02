package main

import (
	"context"

	log "github.com/Sirupsen/logrus"
	"github.com/iron-io/functions/api/datastore"
	"github.com/iron-io/functions/api/mqs"
	"github.com/iron-io/functions/api/server"
	"github.com/spf13/viper"
)

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
	funcServer.Start(ctx)
}
