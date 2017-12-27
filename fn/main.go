package main

import (
	"os"
	"github.com/iron-io/functions/fn/app"
)

func main() {
	fn := app.NewFn()
	fn.Run(os.Args)
}
