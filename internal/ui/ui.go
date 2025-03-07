package ui

import (
	"fmt"
	"log"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/cswank/gogadgets"
	"github.com/cswank/thermostat/internal/button"
	"github.com/cswank/thermostat/internal/gpio"
)

type (
	printer interface {
		Print(tt, at int, state string)
		Clear()
		Message(s string)
	}

	UI struct {
		btn         button.Button
		plus        button.Button
		minus       button.Button
		display     printer
		out         chan<- gogadgets.Message
		ti          chan int
		bi          chan button.State
		external    chan string
		debug       bool
		state       button.State
		furnace     string
		cmd         string
		temperature struct {
			target int
			actual int
		}
	}
)

func New(btn, plus, minus gpio.Waiter, p printer, furnaceAddress string, debug bool) *UI {
	ti := make(chan int)
	bi := make(chan button.State)
	return &UI{
		ti:       ti,
		bi:       bi,
		external: make(chan string),
		btn:      button.New(btn, buttonInput(bi)),
		plus:     button.New(plus, temperatureUp(ti)),
		minus:    button.New(minus, temperatureDown(ti)),
		display:  p,
		furnace:  furnaceAddress,
		debug:    debug,
	}
}

func (u *UI) Start(input <-chan gogadgets.Message, out chan<- gogadgets.Message) {
	go u.btn.Start()
	go u.plus.Start()
	go u.minus.Start()
	go u.input()

	u.temperature.target = 70
	u.out = out

	var stop bool
	for !stop {
		select {
		case msg := <-input:
			u.handleUpdate(msg)
		}
	}

	u.btn.Close()
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
			u.external <- "turn off furnace"
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
	u.state = st
	f, _ := v.ToFloat()
	u.temperature.target = int(math.Round(f))
	u.external <- v.Cmd
}

func (u *UI) input() {
	presses := int(-1)
	var bye int
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
		case cmd := <-u.external:
			bye = 1
			u.display.Message(u.shorten(cmd))
			off.reset(4)
		case <-off.t.C:
			if presses > -1 {
				continue
			}

			off.recv = true
			switch bye {
			case 0:
				bye = 1
				u.display.Message(u.shorten(u.cmd))
				off.reset(4)
			case 1:
				bye = 2
				u.display.Message("bye")
				off.reset(2)
			default:
				bye = 0
				u.display.Clear()
			}
		}
	}
}

func (u UI) shorten(s string) string {
	return strings.Replace(strings.Replace(s, "home ", "", 1), " furnace", "", 1)
}

func (u *UI) command() {
	switch u.state {
	case button.Cool:
		u.cmd = fmt.Sprintf("cool home to %d F", u.temperature.target)
	case button.Heat:
		u.cmd = fmt.Sprintf("heat home to %d F", u.temperature.target)
	case button.Off:
		u.cmd = "turn off furnace"
	}

	u.doCommand(u.cmd)
}

func (u *UI) doCommand(cmd string) {
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

func (u *UI) Handlers() []gogadgets.HTTPHandler {
	return []gogadgets.HTTPHandler{
		&handler{verb: "GET", path: "/", handler: u.status},
		&handler{verb: "POST", path: "/settings", handler: u.settings},
	}
}

func (u *UI) status(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	var changed bool
	switch q.Get("t") {
	case "1":
		u.temperature.target += 1
		changed = true
	case "-1":
		u.temperature.target -= 1
		changed = true
	}

	if q.Get("s") == "next" {
		u.state.Next()
		changed = true
	}

	minus := `<a href="?t=-1">-</a>`
	plus := `<a href="?t=1">+</a>`

	t := strconv.Itoa(u.temperature.target)
	if u.state == button.Off {
		t, minus, plus = "--", "", ""
	}

	fmt.Fprintf(w, fmt.Sprintf(`<!DOCTYPE html>
<html>
<h2><a href="?s=next">%s</a></h2>
<h3>Temperature: %d</h3>
<h3>Target: %s %s %s</h3>
<div><a href="/">refresh</a></div>
</html>`, u.state, u.temperature.actual, minus, t, plus))

	if changed {
		u.command()
	}
}

func (u *UI) settings(w http.ResponseWriter, r *http.Request) {
	log.Println("settings POST")
}

func (u *UI) Verb() string {
	return "GET"
}

func (u *UI) Path() string {
	return "/"
}

func temperatureInput(ch chan int) func(i int) {
	return func(i int) {
		ch <- i
	}
}

func temperatureUp(ch chan int) func() {
	return func() {
		ch <- 1
	}
}

func temperatureDown(ch chan int) func() {
	return func() {
		ch <- -1
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

type handler struct {
	verb    string
	path    string
	handler func(w http.ResponseWriter, r *http.Request)
}

func (h handler) Verb() string {
	return h.verb
}

func (h handler) Path() string {
	return h.path
}

func (h handler) Handler(w http.ResponseWriter, r *http.Request) {
	h.handler(w, r)
}
