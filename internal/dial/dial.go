package dial

import (
	"fmt"
	"log"
	"strconv"

	"github.com/cswank/gogadgets"
)

type (
	Dial struct {
		a     *gogadgets.GPIO
		b     *gogadgets.GPIO
		f     func(i int)
		close chan bool
	}
)

func New(a, b int, f func(i int)) (Dial, error) {
	p1, err := gogadgets.NewGPIO(&gogadgets.Pin{
		Pin:       strconv.Itoa(a),
		Platform:  "rpi",
		Direction: "in",
		Edge:      "falling",
		ActiveLow: "0",
	})

	if err != nil {
		return Dial{}, err
	}

	p2, err := gogadgets.NewGPIO(&gogadgets.Pin{
		Pin:       strconv.Itoa(b),
		Platform:  "rpi",
		Direction: "in",
		Edge:      "falling",
		ActiveLow: "0",
	})

	return Dial{
		f:     f,
		a:     p1.(*gogadgets.GPIO),
		b:     p2.(*gogadgets.GPIO),
		close: make(chan bool),
	}, err
}

func (d Dial) Start() {
	ch1 := make(chan int)
	ch2 := make(chan int)

	go wait(d.a, ch1)
	go wait(d.b, ch2)

	var stop bool
	var i int
	for !stop {
		select {
		case v1 := <-ch1:
			var v2 int
			if d.b.Status()["gpio"] {
				v2 = 1
			}
			fmt.Printf("v1: %d%d\n", v1, v2)
		case v2 := <-ch2:
			var v1 int
			if d.a.Status()["gpio"] {
				v1 = 1
			}
			fmt.Printf("v2: %d%d\n", v1, v2)
		case <-d.close:
			stop = true
		}
		d.f(i)
	}
}

func (d Dial) Close() {
	go func() {
		d.close <- true
	}()
}

func wait(p *gogadgets.GPIO, ch chan int) {
	for {
		if err := p.Wait(); err != nil {
			log.Println("unable to wait for gpio pin")
		}
		var i int
		if p.Status()["gpio"] {
			i = 1
		}
		ch <- i
	}
}
