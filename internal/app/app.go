package app

import (
	"log"
	"os"
	"strconv"

	"github.com/cswank/gogadgets"
	"github.com/cswank/thermostat/internal/display"
	"github.com/cswank/thermostat/internal/ui"
)

var cfg = gogadgets.Config{
	Master: getenv("GOGADGETS_MASTER", "http://192.168.88.234:6111"),
	Host:   getenv("GOGADGETS_HOST", "http://192.168.88.64:6114"),
	Port:   6114,
}

func Start(debug bool) {
	a, b, c, d := deps()

	u := ui.New(a, b, c, d, cfg.Master, debug)
	app := gogadgets.New(&cfg, u)
	app.Start()
}

func deps() (*gogadgets.GPIO, *gogadgets.GPIO, *gogadgets.GPIO, *display.OLED) {
	g1, g2, g3 := newGPIO(18, "falling"), newGPIO(15, "both"), newGPIO(16, "both")

	d, err := display.New()
	if err != nil {
		log.Fatal(err)
	}

	return g1, g2, g3, d
}

func newGPIO(i int, dir string) *gogadgets.GPIO {
	g, err := gogadgets.NewGPIO(pin(i, dir))
	if err != nil {
		log.Fatal(err)
	}
	return g.(*gogadgets.GPIO)
}

func pin(i int, dir string) *gogadgets.Pin {
	return &gogadgets.Pin{
		Pin:       strconv.Itoa(i),
		Platform:  "rpi",
		Direction: "in",
		Edge:      dir,
		ActiveLow: "0",
	}
}

func getenv(key, def string) string {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	return v
}
