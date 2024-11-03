package app

import (
	"log"
	"os"
	"strconv"
	"time"

	"github.com/cswank/gogadgets"
	"github.com/cswank/thermostat/internal/display"
	"github.com/cswank/thermostat/internal/ui"
)

const (
	stateBtn = 15
	plusBtn  = 37
	minusBtn = 13
)

var (
	cfg = gogadgets.Config{
		Host: getenv("GOGADGETS_HOST", "http://192.168.88.64:80"),
		Port: 80,
		Gadgets: []gogadgets.GadgetConfig{
			{
				Name:     "temperature",
				Location: "home",
				Pin: gogadgets.Pin{
					Type:      "thermometer",
					OneWireId: "",
					Units:     "F",
					Sleep:     15 * time.Second,
				},
			},
			{
				Location: "home",
				Name:     "furnace",
				Pin: gogadgets.Pin{
					Type: "thermostat",
					Pins: map[string]gogadgets.Pin{
						"heat": {
							Type:      "gpio",
							Platform:  "rpi",
							Pin:       "35",
							Direction: "out",
						},
						"cool": {
							Type:      "gpio",
							Platform:  "rpi",
							Pin:       "40",
							Direction: "out",
						},
						"fan": {
							Type:      "gpio",
							Platform:  "rpi",
							Pin:       "11",
							Direction: "out",
						},
					},
					Args: map[string]any{
						"sensor":     "home temperature",
						"hysteresis": 3.0,
					},
				},
			},
			{
				Type: "cron",
				Args: map[string]any{
					"jobs": []any{},
				},
			},
		},
	}

	winter = []any{
		"0   22   *  *  *  heat home to 66 F",
		"0   6    *  *  *  heat home to 72 F",
	}

	summer = []any{
		"0   6    *  *  *  cool home to 78 F",
		"0   8    *  *  *  cool home to 82 F",
	}
)

func Start(debug bool, season string, hysteresis float64, w1 string) {
	switch season {
	case "summer":
		cfg.Gadgets[2].Args["jobs"] = summer
	case "winter":
		cfg.Gadgets[2].Args["jobs"] = winter
	}

	btn, plus, minus, d := deps()

	cfg.Gadgets[1].Pin.Args["hysteresis"] = hysteresis
	cfg.Gadgets[0].Pin.OneWireId = w1

	u := ui.New(btn, plus, minus, d, cfg.Master, debug)
	cfg.Endpoints = u.Handlers()
	app := gogadgets.New(&cfg, u)
	app.Start()
}

func deps() (*gogadgets.GPIO, *gogadgets.GPIO, *gogadgets.GPIO, *display.OLED) {
	btn, plus, minus := newGPIO(stateBtn, "falling"), newGPIO(plusBtn, "falling"), newGPIO(minusBtn, "falling")

	d, err := display.New()
	if err != nil {
		log.Fatal(err)
	}

	return btn, plus, minus, d
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
