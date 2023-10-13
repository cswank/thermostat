package gpio

import "os"

type (
	Printer interface {
		Print(s string)
		Off()
	}

	Waiter interface {
		Wait() error
		Status() map[string]bool
		Open() (*os.File, error)
	}
)
