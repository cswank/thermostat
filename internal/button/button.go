package button

import (
	"log"

	"github.com/cswank/thermostat/internal/gpio"
)

type (
	State int

	Button struct {
		gpio  gpio.Waiter
		f     func(s State)
		close chan bool
	}
)

const (
	Off  State = 0
	Heat State = 1
	Cool State = 2
)

func New(g gpio.Waiter, f func(State)) Button {
	return Button{
		gpio:  g,
		f:     f,
		close: make(chan bool),
	}
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

func wait(p gpio.Waiter, ch chan struct{}) {
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
