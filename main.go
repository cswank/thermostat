package main

import (
	"flag"
	"log"

	"github.com/cswank/thermostat/internal/app"
)

var (
	debug      = flag.Bool("debug", false, "write more stuff to led/stdout")
	season     = flag.String("season", "winter", "winter or summer")
	hysteresis = flag.Float64("hysteresis", 5.0, "hysteresis")
)

func main() {
	flag.Parse()
	app.Start(*debug, *season, *hysteresis)
	log.Println("exit")
}
