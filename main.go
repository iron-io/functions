package main

import (
	"context"

	"github.com/iron-io/functions/api/server"
)

func main() {
	ctx := context.Background()

	funcServer := server.New(ctx)
	// Setup your custom extensions, listeners, etc here
	funcServer.Start(ctx)
}
