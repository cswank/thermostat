package app

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/cswank/gogadgets"
	"github.com/cswank/thermostat/internal/gpio"
	"github.com/cswank/thermostat/internal/led"
	"github.com/cswank/thermostat/internal/ui"
)

var cfg = gogadgets.Config{
	Master: getenv("GOGADGETS_MASTER", "http://192.168.88.234:6111"),
	Host:   getenv("GOGADGETS_HOST", "http://192.168.88.254:6114"),
	Port:   6114,
}

func Start(fakeDeps bool) {
	var a, b, c gpio.Waiter
	var l gpio.Printer
	var debug bool
	if fakeDeps {
		a, b, c, l = &fake{}, &fake{}, &fake{}, &fake{}
		debug = true
	} else {
		a, b, c, l = real()
	}

	u := ui.New(a, b, c, l, debug)
	app := gogadgets.New(&cfg, &u)
	app.Start()
}

func real() (*gogadgets.GPIO, *gogadgets.GPIO, *gogadgets.GPIO, led.LED) {
	g1, g2, g3 := newGPIO(14), newGPIO(15), newGPIO(16)

	// TODO: real pins
	l, err := led.New(
		[7]int{0, 1, 2, 3, 4, 5, 6},
		[7]int{0, 1, 2, 3, 4, 5, 6})
	if err != nil {
		log.Fatal(err)
	}

	return g1, g2, g3, l
}

func newGPIO(i int) *gogadgets.GPIO {
	g, err := gogadgets.NewGPIO(pin(i))
	if err != nil {
		log.Fatal(err)
	}
	return g.(*gogadgets.GPIO)
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

type fake struct{}

func (f fake) Print(s string) {
	fmt.Printf("\r%s", s)
}

func (f fake) Off() {
	fmt.Print("\r  ")
}

func (f fake) Wait() error {
	ch := make(chan int)
	<-ch
	return nil
}

func (f fake) Status() map[string]bool {
	return map[string]bool{}
}

func getenv(key, def string) string {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	return v
}
