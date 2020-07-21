package main

import (
	"os"
	"github.com/wavemechanics/deliver/app"
)

func main() {
	os.Exit(app.Run(os.Args[1:]))
}