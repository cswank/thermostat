package main

import (
	"flag"

	"github.com/cswank/thermostat/internal/app"
)

var (
	debug = flag.Bool("debug", false, "write more stuff to led/stdout")
)

func main() {
	flag.Parse()
	app.Start(*debug)
}
