package ui

import (
	"fmt"
	"log"
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
		Clear()
		Message(s string)
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
	return &UI{
		ti:      ti,
		bi:      bi,
		dial:    dial.New(A, B, temperatureInput(ti)),
		btn:     button.New(btn, buttonInput(bi)),
		display: p,
		furnace: furnaceAddress,
		debug:   debug,
	}
}

func (u *UI) Start(input <-chan gogadgets.Message, out chan<- gogadgets.Message) {
	go u.dial.Start()
	go u.btn.Start()
	go u.input()

	u.out = out

	go func() {
		time.Sleep(time.Second)
		out <- gogadgets.Message{
			UUID:   gogadgets.GetUUID(),
			Type:   gogadgets.COMMAND,
			Sender: "thermostat",
			Host:   u.furnace,
			Body:   "update",
		}
	}()

	reconnect := time.After(10 * time.Minute)

	var stop bool
	for !stop {
		select {
		case msg := <-input:
			u.handleUpdate(msg)
			reconnect = time.After(15 * time.Minute)
		case <-reconnect:
			log.Println("reconnect")
			stop = true
		}
	}

	u.btn.Close()
	u.dial.Close()
	log.Println("exit")
}

func (u *UI) handleUpdate(msg gogadgets.Message) {
	if msg.Type != "update" {
		return
	}

	switch msg.Sender {
	case "home temperature":
		v, ok := msg.Value.ToFloat()
		if ok {
			i := int(math.Round(v))
			u.temperature.actual = int(i)
		}
	case "home furnace":
		if msg.Value.Cmd != "" {
			switch msg.Value.Cmd {
			case "turn off furnace":
				u.state = button.State(button.Off)
			case "heat home":
				u.state = button.State(button.Heat)
				if msg.TargetValue != nil {
					f, _ := msg.TargetValue.ToFloat()
					u.temperature.target = int(math.Round(f))
				}
			case "cool home":
				u.state = button.State(button.Cool)
				if msg.TargetValue != nil {
					f, _ := msg.TargetValue.ToFloat()
					u.temperature.target = int(math.Round(f))
				}
			}
		} else {
			if msg.Value.Output["cool"] {
				u.state = button.State(button.Cool)
			} else if msg.Value.Output["heat"] {
				u.state = button.State(button.Heat)
				if msg.TargetValue != nil {
					f, _ := msg.TargetValue.ToFloat()
					u.temperature.target = int(math.Round(f))
				}
			} else {
				u.state = button.State(button.Off)
				if msg.TargetValue != nil {
					f, _ := msg.TargetValue.ToFloat()
					u.temperature.target = int(math.Round(f))
				}
			}
		}
	}
}

func (u *UI) input() {
	var cmd <-chan time.Time
	presses := int(-1)
	u.temperature.target = 70
	bye := true
	u.display.Message("hi")
	off := time.After(1 * time.Second)
	for {
		select {
		case i := <-u.ti:
			if u.state != button.Off {
				u.temperature.target += i
				u.display.Print(u.temperature.target, u.temperature.actual, u.state.String())
				cmd = time.After(2 * time.Second)
				presses = 2
			}
		case <-u.bi:
			if presses > -1 {
				u.state.Next()
			}
			presses++
			u.display.Print(u.temperature.target, u.temperature.actual, u.state.String())
			cmd = time.After(2 * time.Second)
		case <-cmd:
			if presses > 0 {
				u.command()
			}
			presses = -1
			off = time.After(5 * time.Second)
		case <-off:
			if !bye {
				bye = true
				u.display.Message("bye")
				off = time.After(1 * time.Second)
			} else {
				bye = false
				u.display.Clear()
			}
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
