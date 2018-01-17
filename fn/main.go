package main

import (
	"github.com/iron-io/functions/fn/app"
	"os"
)

func main() {
	fn := app.NewFn()
	fn.Run(os.Args)
}
