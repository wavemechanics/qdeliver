package main

import (
	"os"

	"github.com/wavemechanics/qdeliver/app"
)

func main() {
	os.Exit(app.Run(os.Args[1:]))
}
