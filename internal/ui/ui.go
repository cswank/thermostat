package ui

import (
	"fmt"
	"log"
	"math"
	"strings"
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
	log.Println(furnaceAddress)
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
		time.Sleep(1 * time.Second)
		out <- gogadgets.Message{
			UUID:   gogadgets.GetUUID(),
			Type:   gogadgets.COMMAND,
			Sender: "thermostat",
			Host:   u.furnace,
			Body:   "update",
		}
	}()

	reconnect := time.NewTicker(15 * time.Minute)

	var stop bool
	var lastUpdate time.Time
	for !stop {
		select {
		case msg := <-input:
			u.handleUpdate(msg)
			lastUpdate = time.Now()
		case <-reconnect.C:
			if time.Now().Sub(lastUpdate) > (15 * time.Minute) {
				go u.reconnect(out)
			}
		}
	}

	u.btn.Close()
	u.dial.Close()
}

func (u *UI) handleUpdate(msg gogadgets.Message) {
	if msg.Type != "update" {
		return
	}

	switch msg.Sender {
	case "home temperature":
		u.updateActual(msg.Value)
	case "home furnace":
		if msg.TargetValue != nil {
			switch firstTwoWords(msg.TargetValue.Cmd) {
			case "heat home":
				u.updateState(msg.TargetValue, button.Heat)
			case "cool home":
				u.updateState(msg.TargetValue, button.Cool)
			}
		} else {
			u.state = button.State(button.Off)
		}
	}
}

func (u *UI) updateActual(val gogadgets.Value) {
	v, ok := val.ToFloat()
	if ok {
		i := int(math.Round(v))
		u.temperature.actual = int(i)
	}
}

func (u *UI) updateState(v *gogadgets.Value, st button.State) {
	u.state = button.State(st)
	f, _ := v.ToFloat()
	u.temperature.target = int(math.Round(f))
}

func (u *UI) reconnect(out chan<- gogadgets.Message) {
	log.Println("reconnect")
	out <- gogadgets.Message{
		UUID:   gogadgets.GetUUID(),
		Type:   gogadgets.COMMAND,
		Host:   u.furnace,
		Sender: "thermostat",
		Body:   "reconnect",
	}
	out <- gogadgets.Message{
		UUID:   gogadgets.GetUUID(),
		Type:   gogadgets.COMMAND,
		Sender: "thermostat",
		Host:   u.furnace,
		Body:   "update",
	}
	log.Println("reconnected")
}

func (u *UI) input() {
	presses := int(-1)
	bye := true
	u.display.Message("hi")
	off := newTimer(1 * time.Second)
	cmd := newTimer(-1)
	for {
		select {
		case i := <-u.ti:
			if u.state != button.Off {
				u.temperature.target += i
				u.display.Print(u.temperature.target, u.temperature.actual, u.state.String())
				cmd.reset(3)
				presses = 2
			}
		case <-u.bi:
			if presses > -1 {
				u.state.Next()
			}
			presses++
			u.display.Print(u.temperature.target, u.temperature.actual, u.state.String())
			cmd.reset(3)
		case <-cmd.t.C:
			cmd.recv = true
			if presses > 0 {
				u.command()
			}
			presses = -1
			off.reset(5)
		case <-off.t.C:
			off.recv = true
			if !bye {
				bye = true
				u.display.Message("bye")
				off.reset(1)
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

func (u UI) GetUID() string {
	return "ui"
}

func (u UI) GetDirection() string {
	return "input"
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

func firstTwoWords(s string) string {
	p := strings.Split(s, " ")
	if len(p) < 2 {
		return ""
	}
	return fmt.Sprintf("%s %s", p[0], p[1])
}

type timer struct {
	t    *time.Timer
	recv bool
}

func newTimer(seconds time.Duration) *timer {
	return &timer{t: time.NewTimer(seconds * time.Second)}
}

func (t *timer) reset(seconds time.Duration) {
	if !t.t.Stop() {
		if !t.recv {
			<-t.t.C
		}
	}
	t.t.Reset(seconds * time.Second)
}
