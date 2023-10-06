package main

import (
	"fmt"

	"github.com/cswank/gogadgets"
)

var cfg = gogadgets.Config{
	Master: "http://192.168.88.234:6111",
	Host:   "http://192.168.88.254",
	Port:   6114,
}

func main() {
	app := gogadgets.New(cfg)
	app.Start()
}

type ui struct {
	out chan<- gogadgets.Message
}

func (u *ui) Start(input <-chan gogadgets.Message, out chan<- gogadgets.Message) {
	for {
		msg := <-input
		fmt.Printf("%+v\n", msg)
	}
}
