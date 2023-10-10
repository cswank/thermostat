package gpio

type (
	Printer interface {
		Print(s string)
		Off()
	}

	Waiter interface {
		Wait() error
		Status() map[string]bool
	}
)
