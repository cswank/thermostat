package button

import (
	"log"
	"time"

	"github.com/cswank/thermostat/internal/gpio"
)

type (
	State int

	Button struct {
		gpio  gpio.Waiter
		f     func()
		close chan bool
	}
)

const (
	Off  State = 0
	Heat State = 1
	Cool State = 2
)

func New(g gpio.Waiter, f func()) Button {
	return Button{
		gpio:  g,
		f:     f,
		close: make(chan bool),
	}
}

func (b Button) Start() {
	ch := make(chan struct{})
	go wait(b.gpio, ch)

	var stop bool
	for !stop {
		select {
		case <-ch:
			b.f()
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
		time.Sleep(200 * time.Millisecond)
	}
}

func (s *State) Next() State {
	*s += 1
	if *s > 2 {
		*s = 0
	}
	return *s
}

func (s *State) Prev() State {
	*s -= 1
	if *s < 0 {
		*s = 2
	}
	return *s
}

func (s State) String() string {
	switch s {
	case Cool:
		return "Cool"
	case Heat:
		return "Heat"
	default:
		return "Off"
	}
}
