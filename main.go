package main

import (
	"flag"
	"log"
	"strconv"

	"github.com/cswank/gogadgets"
	"github.com/cswank/thermostat/internal/gpio"
	"github.com/cswank/thermostat/internal/led"
	"github.com/cswank/thermostat/internal/ui"
)

var (
	cfg = gogadgets.Config{
		Master: "http://192.168.88.234:6111",
		Host:   "http://192.168.88.254:6114",
		Port:   6114,
	}

	fk = flag.Bool("fake", false, "don't connect to real GPIO")
)

func main() {
	flag.Parse()

	var a, b, c gpio.Waiter
	var l gpio.Printer
	if *fk {
		a, b, c, l = fake()
	} else {
		a, b, c, l = real()

	}

	u := ui.New(a, b, c, l)
	app := gogadgets.New(&cfg, &u)
	app.Start()
}

func fake() (*gpio.Fake, *gpio.Fake, *gpio.Fake, *gpio.Fake) {
	return &gpio.Fake{}, &gpio.Fake{}, &gpio.Fake{}, &gpio.Fake{}
}

func real() (*gogadgets.GPIO, *gogadgets.GPIO, *gogadgets.GPIO, led.LED) {
	p1, err := gogadgets.NewGPIO(pin(14))
	if err != nil {
		log.Fatal(err)
	}

	p2, err := gogadgets.NewGPIO(pin(15))
	if err != nil {
		log.Fatal(err)
	}

	p3, err := gogadgets.NewGPIO(pin(16))
	if err != nil {
		log.Fatal(err)
	}

	l, err := led.New([7]int{0, 1, 2, 3, 4, 5, 6}, [7]int{0, 1, 2, 3, 4, 5, 6})
	if err != nil {
		log.Fatal(err)
	}

	return p1.(*gogadgets.GPIO), p2.(*gogadgets.GPIO), p3.(*gogadgets.GPIO), l
}

func pin(i int) *gogadgets.Pin {
	return &gogadgets.Pin{
		Pin:       strconv.Itoa(i),
		Platform:  "rpi",
		Direction: "in",
		Edge:      "falling",
		ActiveLow: "0",
	}
}
