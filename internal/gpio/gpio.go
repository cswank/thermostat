package gpio

import "os"

type (
	Waiter interface {
		Wait() error
		Status() map[string]bool
		Open() (*os.File, error)
	}
)
