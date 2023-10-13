package ui

import (
	"fmt"
	"time"

	"github.com/cswank/gogadgets"
	"github.com/cswank/thermostat/internal/button"
	"github.com/cswank/thermostat/internal/dial"
	"github.com/cswank/thermostat/internal/gpio"
)

type (
	UI struct {
		dial        dial.Dial
		btn         button.Button
		led         gpio.Printer
		out         chan<- gogadgets.Message
		ti          chan int
		bi          chan button.State
		debug       bool
		temperature struct {
			set    int
			actual int
		}
	}
)

func New(btn, A, B gpio.Waiter, led gpio.Printer, debug bool) UI {
	ti := make(chan int)
	bi := make(chan button.State)
	u := UI{
		ti:    ti,
		bi:    bi,
		dial:  dial.New(A, B, temperatureInput(ti)),
		btn:   button.New(btn, buttonInput(bi)),
		led:   led,
		debug: debug,
	}

	go u.dial.Start()
	go u.btn.Start()
	go u.input()
	return u
}

func (u *UI) Start(input <-chan gogadgets.Message, out chan<- gogadgets.Message) {
	u.out = out
	for msg := range input {
		if msg.Type != "update" {
			continue
		}

		switch msg.Sender {
		case "home temperature":
			v, ok := msg.Value.ToFloat()
			if ok {
				u.temperature.actual = int(v)
				if u.debug {
					u.ti <- int(v)
				}
			}
		case "home furnace":
			if msg.TargetValue == nil {
				continue
			}

			switch msg.Value.Cmd {
			case "turn off furnace":
				u.bi <- button.Off
			case "heat home":
				u.bi <- button.Heat
				f, _ := msg.TargetValue.ToFloat()
				u.temperature.set = int(f)
			case "cool home":
				u.bi <- button.Cool
				f, _ := msg.TargetValue.ToFloat()
				u.temperature.set = int(f)
			}
		}
	}

	u.btn.Close()
	u.dial.Close()
}

func (u *UI) input() {
	var tk <-chan time.Time
	for {
		select {
		case i := <-u.ti:
			u.led.Print(fmt.Sprintf("%d", i))
			tk = time.After(2 * time.Second)
		case s := <-u.bi:
			u.led.Print(s.String())
			tk = time.After(2 * time.Second)
		case <-tk:
			u.led.Off()
		}
	}
}

func temperatureInput(ch chan int) func(i int) {
	return func(i int) {
		ch <- i
	}
}

func buttonInput(ch chan button.State) func(b button.State) {
	return func(b button.State) {
		ch <- b
	}
}

func (u UI) GetUID() string {
	return "ui"
}

func (u UI) GetDirection() string {
	return "input"
}
