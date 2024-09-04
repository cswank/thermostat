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
		bounds image.Rectangle
		lock   sync.Mutex
		small  font.Face
		large  font.Face
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
		small:  truetype.NewFace(font, &truetype.Options{Size: 24, DPI: 72}),
		large:  truetype.NewFace(font, &truetype.Options{Size: 38, DPI: 72}),
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
	o.print(msg, 38, o.small)
}

func (o *OLED) Print(target, actual int, state string) {
	o.print(o.temperature(target, actual, state), 36, o.large)
	o.print(state, 60, o.small)
}

func (o *OLED) temperature(target, actual int, state string) string {
	if state == "Off" {
		return fmt.Sprintf("--    %02d", actual)
	}
	return fmt.Sprintf("%02d    %02d", target, actual)
}

func (o *OLED) print(msg string, y int, face font.Face) {
	o.lock.Lock()

	img := image1bit.NewVerticalLSB(o.bounds)

	d := font.Drawer{
		Dst:  img,
		Src:  &image.Uniform{C: image1bit.On},
		Face: face,
	}

	rec, _ := d.BoundString(msg)
	d.Dot = fixed.P(64-rec.Max.X.Ceil()/2, y)

	d.DrawString(msg)
	if err := o.dev.Draw(o.bounds, img, image.Point{}); err != nil {
		log.Fatal(err)
	}

	o.lock.Unlock()
}
