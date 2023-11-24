package main

import (
	"flag"
	"log"

	"github.com/cswank/thermostat/internal/app"
)

var (
	debug = flag.Bool("debug", false, "write more stuff to led/stdout")
)

func main() {
	flag.Parse()
	app.Start(*debug)
	log.Println("exit")
}
