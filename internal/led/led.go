package led

import (
	"strconv"

	"github.com/cswank/gogadgets"
)

var (
	chars = map[rune]uint8{
		//     abcdefg
		'0': 0b11111100,
		'1': 0b01100000,
		'2': 0b11011010,
		'3': 0b11110010,
		'4': 0b01100110,
		'5': 0b10110110,
		'6': 0b10111110,
		'7': 0b11100000,
		'8': 0b11111110,
		'9': 0b11110110,
		'H': 0b01101110,
		'E': 0b10011110,
		'C': 0b10011100,
		'L': 0b00011100,
		'F': 0b10001110,
	}
)

type LED struct {
	d1 digit
	d2 digit
}

func New(p1, p2 [7]int) (LED, error) {
	var l LED
	d1, err := newDigit(p1)
	if err != nil {
		return l, err
	}

	d2, err := newDigit(p1)
	if err != nil {
		return l, err
	}

	l.d1 = d1
	l.d2 = d2
	return l, nil
}

func (l LED) Print(s string) {
	l.d1.print(chars[rune(s[0])])
	l.d2.print(chars[rune(s[1])])
}

func (l LED) Off() {
	l.d1.print(0)
	l.d2.print(0)
}

func newDigit(pins [7]int) (digit, error) {
	var d digit
	for pin := range pins {
		g, err := gogadgets.NewGPIO(&gogadgets.Pin{
			Pin:       strconv.Itoa(pin),
			Platform:  "rpi",
			Direction: "out",
		})

		if err != nil {
			return d, err
		}

		d.pins[pin] = g.(*gogadgets.GPIO)
	}

	return d, nil
}

type digit struct {
	pins [7]*gogadgets.GPIO
}

func (d digit) print(j uint8) {
	for i := uint8(7); i > 0; i-- {
		if (j>>i | 1) > 0 {
			d.pins[i].On(nil)
		} else {
			d.pins[i].Off()
		}
	}
}
