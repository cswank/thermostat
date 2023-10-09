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
	fmt.Printf("\r%s", s)
}

func (f Fake) Off() {
	fmt.Print("\r  ")
}

func (f Fake) Wait() error {
	ch := make(chan int)
	<-ch
	return nil
}

func (f Fake) Status() map[string]bool {
	return map[string]bool{}
}
