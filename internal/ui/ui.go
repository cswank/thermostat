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
		temperature struct {
			set    int
			actual int
		}
	}
)

func New(btn, A, B gpio.Waiter, led gpio.Printer) UI {
	ti := make(chan int)
	bi := make(chan button.State)
	u := UI{
		ti:   ti,
		bi:   bi,
		dial: dial.New(A, B, temperatureInput(ti)),
		btn:  button.New(btn, buttonInput(bi)),
		led:  led,
	}

	go u.dial.Start()
	go u.btn.Start()
	go u.input()
	return u
}

// Start is called by the gogadgets app
func (u *UI) Start(input <-chan gogadgets.Message, out chan<- gogadgets.Message) {
	u.out = out
	for msg := range input {
		if msg.Sender != "home temperature" {
			continue
		}

		switch msg.Type {
		case "update":
			v, ok := msg.Value.ToFloat()
			if ok {
				u.temperature.actual = int(v)
			}
		case "command":
			f, _, _ := gogadgets.ParseCommand(msg.Body)
			u.temperature.set = int(f)
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
			u.led.Print(fmt.Sprintf("%02d", i))
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
