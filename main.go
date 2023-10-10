package main

import (
	"flag"

	"github.com/cswank/thermostat/internal/app"
)

var (
	fake  = flag.Bool("fake", false, "don't connect to real GPIO")
	debug = flag.Bool("debug", false, "write more stuff to led/stdout")
)

func main() {
	flag.Parse()
	app.Start(*fake, *debug)
}
