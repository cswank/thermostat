package main

import (
	"flag"

	"github.com/cswank/thermostat/internal/app"
)

var (
	fake = flag.Bool("fake", false, "don't connect to real GPIO")
)

func main() {
	flag.Parse()
	app.Start(*fake)
}
