package dial

import (
	"fmt"
	"log"

	"github.com/cswank/thermostat/internal/gpio"
)

type (
	Dial struct {
		a     gpio.Waiter
		b     gpio.Waiter
		f     func(i int)
		close chan bool
	}
)

func New(p1, p2 gpio.Waiter, f func(i int)) Dial {
	return Dial{
		f:     f,
		a:     p1,
		b:     p2,
		close: make(chan bool),
	}
}

func (d Dial) Start() {
	ch1 := make(chan uint8)
	ch2 := make(chan uint8)

	go wait(d.a, ch1)
	go wait(d.b, ch2)
	var q queue

	var stop bool
	for !stop {
		select {
		case a := <-ch1:
			v := uint16(1)
			if a == 1 {
				v = 3
			}
			if i := q.add(v); i != 0 {
				d.f(i)
			}
		case b := <-ch2:
			v := uint16(2)
			if b == 1 {
				v = 4
			}
			if i := q.add(v); i != 0 {
				d.f(i)
			}
		case <-d.close:
			stop = true
		}
	}
}

var (
	// 1, 2, 3, 4
	left = uint16(0b001010011100)
	// 2, 1, 4, 3
	right = uint16(0b010001100011)
)

type queue uint16

func (q *queue) add(i uint16) int {
	v := uint16(*q)
	v = v<<3 | i
	*q = queue(v)
	fmt.Printf("  %03b  %012b\n", i, v&0b111111111111)
	if v&0b111111111111 == left {
		return -1
	}
	if v&0b111111111111 == right {
		return 1
	}
	return 0
}

func (d Dial) Close() {
	go func() {
		d.close <- true
	}()
}

func wait(p gpio.Waiter, ch chan uint8) {
	f, err := p.Open()
	buf := make([]byte, 2)
	if err != nil {
		log.Fatal(err)
	}
	var v uint8
	for {
		if err := p.Wait(); err != nil {
			log.Println("unable to wait for gpio pin")
		}
		f.Seek(0, 0)
		f.Read(buf)
		v = 0
		if string(buf) == "1\n" {
			v = 1
		}
		ch <- v
	}
}
