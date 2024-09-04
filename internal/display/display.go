package display

import (
	"fmt"
	"image"
	"log"

	"sync"

	"github.com/golang/freetype/truetype"
	"periph.io/x/conn/v3/i2c"
	"periph.io/x/conn/v3/i2c/i2creg"
	"periph.io/x/devices/v3/ssd1306"
	"periph.io/x/devices/v3/ssd1306/image1bit"
	"periph.io/x/host/v3"

	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/goregular"
	"golang.org/x/image/math/fixed"
)

type (
	OLED struct {
		bc     i2c.BusCloser
		dev    *ssd1306.Dev
		font   *truetype.Font
		bounds image.Rectangle
		lock   sync.Mutex
	}
)

func New() (*OLED, error) {
	_, err := host.Init()
	if err != nil {
		return nil, err
	}

	bc, err := i2creg.Open("")
	if err != nil {
		return nil, err
	}

	dev, err := ssd1306.NewI2C(bc, &ssd1306.DefaultOpts)
	if err != nil {
		return nil, err
	}

	font, err := truetype.Parse(goregular.TTF)
	if err != nil {
		return nil, err
	}

	return &OLED{
		bc:     bc,
		dev:    dev,
		font:   font,
		bounds: dev.Bounds(),
	}, nil
}

func (o *OLED) Close() {
	o.bc.Close()
}

func (o *OLED) Clear() {
	o.dev.Halt()
}

func (o *OLED) Message(msg string) {
	o.print(msg, 24, 38)
}

func (o *OLED) Print(target, actual int, state string) {
	o.print(o.temperature(target, actual, state), 38, 36)
	o.print(state, 24, 60)
}

func (o *OLED) temperature(target, actual int, state string) string {
	if state == "Off" {
		return fmt.Sprintf("--    %02d", actual)
	}
	return fmt.Sprintf("%02d    %02d", target, actual)
}

func (o *OLED) print(msg string, size float64, y int) {
	o.lock.Lock()

	img := image1bit.NewVerticalLSB(o.bounds)

	d := font.Drawer{
		Dst:  img,
		Src:  &image.Uniform{image1bit.On},
		Face: truetype.NewFace(o.font, &truetype.Options{Size: size, DPI: 72}),
	}

	rec, _ := d.BoundString(msg)
	d.Dot = fixed.P(64-rec.Max.X.Ceil()/2, y)

	d.DrawString(msg)
	if err := o.dev.Draw(o.dev.Bounds(), img, image.Point{}); err != nil {
		log.Fatal(err)
	}

	o.lock.Unlock()
}
