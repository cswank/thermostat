package ui

import (
	"fmt"

	"github.com/cswank/gogadgets"
	"github.com/cswank/thermostat/internal/button"
	"github.com/cswank/thermostat/internal/dial"
)

type (
	UI struct {
		dial dial.Dial
		btn  button.Button
		out  chan<- gogadgets.Message
		ti   chan int
		bi   chan button.State
	}
)

func New(btn, A, B int) (UI, error) {
	u := UI{
		ti: make(chan int),
		bi: make(chan button.State),
	}

	d, err := dial.New(A, B, u.temperatureInput)
	if err != nil {
		return u, err
	}

	b, err := button.New(btn, u.buttonInput)
	if err != nil {
		return u, err
	}

	u.dial = d
	u.btn = b

	go d.Start()
	go b.Start()
	go u.input()
	return u, nil
}

func (u *UI) Start(input <-chan gogadgets.Message, out chan<- gogadgets.Message) {
	u.out = out
	for msg := range input {
		fmt.Printf("%+v\n", msg)
	}

	u.btn.Close()
	u.dial.Close()
}

func (u UI) input() {
	for {
		select {
		case i := <-u.ti:
			fmt.Printf("temperature input: %d", i)
		case s := <-u.bi:
			fmt.Printf("temperature input: %s", s)
		}
	}
}

func (u UI) temperatureInput(i int) {
	u.ti <- i
}

func (u UI) buttonInput(b button.State) {
	u.bi <- b
}

func (u UI) GetUID() string {
	return "ui"
}

func (u UI) GetDirection() string {
	return "input"
}
