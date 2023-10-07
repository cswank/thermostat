package main

import (
	"fmt"

	"github.com/cswank/gogadgets"
)

var cfg = gogadgets.Config{
	Master: "http://192.168.88.234:6111",
	Host:   "http://192.168.88.254:6114",
	Port:   6114,
}

func main() {
	u := ui{}
	app := gogadgets.New(&cfg, &u)
	app.Start()
}

type ui struct {
	out chan<- gogadgets.Message
}

func (u *ui) Start(input <-chan gogadgets.Message, out chan<- gogadgets.Message) {
	for msg := range input {
		fmt.Printf("%+v\n", msg)
	}
}

func (u ui) GetUID() string {
	return "ui"
}

func (u ui) GetDirection() string {
	return "input"
}
