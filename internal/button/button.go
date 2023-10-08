package button

import (
	"log"
	"strconv"

	"github.com/cswank/gogadgets"
)

type (
	State int

	Button struct {
		gpio  *gogadgets.GPIO
		f     func(s State)
		close chan bool
	}
)

const (
	Off  State = 0
	Heat State = 1
	Cool State = 2
)

func New(pin int, f func(State)) (Button, error) {
	g, err := gogadgets.NewGPIO(&gogadgets.Pin{
		Pin:       strconv.Itoa(pin),
		Platform:  "rpi",
		Direction: "in",
		Edge:      "falling",
		ActiveLow: "0",
	})

	return Button{
		gpio:  g.(*gogadgets.GPIO),
		f:     f,
		close: make(chan bool),
	}, err
}

func (b Button) Start() {
	ch := make(chan struct{})
	go wait(b.gpio, ch)

	var st State
	var stop bool
	for !stop {
		select {
		case <-ch:
			st = st.next()
			b.f(st)
		case <-b.close:
			stop = true
		}
	}
}

func (b Button) Close() {
	go func() {
		b.close <- true
	}()
}

func wait(p *gogadgets.GPIO, ch chan struct{}) {
	for {
		if err := p.Wait(); err != nil {
			log.Println("unable to wait for gpio pin")
		}
		ch <- struct{}{}
	}
}

func (s State) next() State {
	s += 1
	if s > 2 {
		s = 0
	}
	return s
}

func (s State) String() string {
	switch s {
	case Cool:
		return "cool"
	case Heat:
		return "heat"
	default:
		return "off"
	}
}
