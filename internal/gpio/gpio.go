package gpio

import "fmt"

type (
	Printer interface {
		Print(s string)
		Off()
	}

	Waiter interface {
		Wait() error
		Status() map[string]bool
	}

	Fake struct{}
)

func (f Fake) Print(s string) {
	fmt.Println(s)
}

func (f Fake) Off() {}

func (f Fake) Wait() error {
	ch := make(chan int)
	<-ch
	return nil
}

func (f Fake) Status() map[string]bool {
	return map[string]bool{}
}
