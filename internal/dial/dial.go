package dial

import (
	"fmt"
	"log"

	"github.com/cswank/gogadgets"
)

func x() {
	p1, err := gogadgets.NewGPIO(&gogadgets.Pin{
		Pin:       "15",
		Platform:  "rpi",
		Direction: "in",
		Edge:      "falling",
		ActiveLow: "0",
	})

	if err != nil {
		log.Fatal(err)
	}

	p2, err := gogadgets.NewGPIO(&gogadgets.Pin{
		Pin:       "16",
		Platform:  "rpi",
		Direction: "in",
		Edge:      "falling",
		ActiveLow: "0",
	})

	ch1 := make(chan int)
	ch2 := make(chan int)

	go wait(p1.(*gogadgets.GPIO), ch1)
	go wait(p2.(*gogadgets.GPIO), ch2)

	if err != nil {
		log.Fatal(err)
	}

	for {
		select {
		case v1 := <-ch1:
			var v2 int
			if p2.Status()["gpio"] {
				v2 = 1
			}
			fmt.Printf("v1: %d%d\n", v1, v2)
		case v2 := <-ch2:
			var v1 int
			if p2.Status()["gpio"] {
				v1 = 1
			}
			fmt.Printf("v2: %d%d\n", v1, v2)
		}
	}
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
