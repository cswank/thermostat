package ui

import (
	"fmt"
	"math"
	"time"

	"github.com/cswank/gogadgets"
	"github.com/cswank/thermostat/internal/button"
	"github.com/cswank/thermostat/internal/dial"
	"github.com/cswank/thermostat/internal/gpio"
)

type (
	printer interface {
		Print(tt, at int, state string)
	}

	UI struct {
		dial        dial.Dial
		btn         button.Button
		display     printer
		out         chan<- gogadgets.Message
		ti          chan int
		bi          chan button.State
		debug       bool
		state       button.State
		furnace     string
		temperature struct {
			target int
			actual int
		}
	}
)

func New(btn, A, B gpio.Waiter, p printer, furnaceAddress string, debug bool) *UI {
	ti := make(chan int)
	bi := make(chan button.State)
	u := UI{
		ti:      ti,
		bi:      bi,
		dial:    dial.New(A, B, temperatureInput(ti)),
		btn:     button.New(btn, buttonInput(bi)),
		display: p,
		furnace: furnaceAddress,
		debug:   debug,
	}

	go u.dial.Start()
	go u.btn.Start()
	go u.input()
	return &u
}

func (u *UI) Start(input <-chan gogadgets.Message, out chan<- gogadgets.Message) {
	fmt.Println(&out)
	u.out = out
	for msg := range input {
		if msg.Type != "update" {
			continue
		}

		switch msg.Sender {
		case "home temperature":
			fmt.Println("cmd", msg.Value)
			v, ok := msg.Value.ToFloat()
			if ok {
				i := int(math.Round(v))
				if i != u.temperature.actual {
					u.display.Print(u.temperature.target, i, u.state.String())
				}
				u.temperature.actual = int(v)
			}
		case "home furnace":
			if msg.TargetValue == nil {
				continue
			}

			fmt.Println("cmd", msg.Value.Cmd)
			switch msg.Value.Cmd {
			case "turn off furnace":
				u.state = button.State(button.Off)
				u.display.Print(u.temperature.target, u.temperature.actual, u.state.String())
			case "heat home":
				u.state = button.State(button.Heat)
				f, _ := msg.TargetValue.ToFloat()
				u.temperature.target = int(math.Round(f))
				u.display.Print(u.temperature.target, u.temperature.actual, u.state.String())
			case "cool home":
				u.state = button.State(button.Cool)
				f, _ := msg.TargetValue.ToFloat()
				u.temperature.target = int(math.Round(f))
				u.display.Print(u.temperature.target, u.temperature.actual, u.state.String())
			}
		}
	}

	u.btn.Close()
	u.dial.Close()
}

func (u *UI) input() {
	var tk <-chan time.Time
	presses := int(-1)
	//TODO:  set this at startup
	u.temperature.target = 70
	for {
		select {
		case i := <-u.ti:
			if u.state != button.Off {
				u.temperature.target += i
				u.display.Print(u.temperature.target, u.temperature.actual, u.state.String())
				tk = time.After(2 * time.Second)
				presses = 2
			}
		case <-u.bi:
			if presses > -1 {
				u.state.Next()
			}
			presses++
			u.display.Print(u.temperature.target, u.temperature.actual, u.state.String())
			tk = time.After(2 * time.Second)
		case <-tk:
			if presses > 0 {
				u.command()
			}
			presses = -1
		}
	}
}

func (u *UI) command() {
	var cmd string
	switch u.state {
	case button.Cool:
		cmd = fmt.Sprintf("cool home to %d F", u.temperature.target)
	case button.Heat:
		cmd = fmt.Sprintf("heat home to %d F", u.temperature.target)
	case button.Off:
		cmd = "turn off furnace"
	}

	fmt.Printf("%s\n", cmd)
	u.out <- gogadgets.Message{
		UUID:   gogadgets.GetUUID(),
		Type:   gogadgets.COMMAND,
		Sender: "thermostat",
		Host:   u.furnace,
		Body:   cmd,
	}
}

func temperatureInput(ch chan int) func(i int) {
	return func(i int) {
		ch <- i
	}
}

func buttonInput(ch chan button.State) func() {
	return func() {
		ch <- 1
	}
}

func (u UI) GetUID() string {
	return "ui"
}

func (u UI) GetDirection() string {
	return "input"
}
