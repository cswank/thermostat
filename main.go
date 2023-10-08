package main

import (
	"log"

	"github.com/cswank/gogadgets"
	"github.com/cswank/thermostat/internal/ui"
)

var cfg = gogadgets.Config{
	Master: "http://192.168.88.234:6111",
	Host:   "http://192.168.88.254:6114",
	Port:   6114,
}

func main() {
	u, err := ui.New(14, 15, 16)
	if err != nil {
		log.Fatal(err)
	}

	app := gogadgets.New(&cfg, &u)
	app.Start()
}
