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
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/font/gofont/goregular"
	"golang.org/x/image/math/fixed"
)

type (
	OLED struct {
		bc   i2c.BusCloser
		dev  *ssd1306.Dev
		lock sync.Mutex
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

	return &OLED{bc: bc, dev: dev}, nil
}

func (o *OLED) Close() {
	o.bc.Close()
}

func (o *OLED) Print(tt, at int, state string) {
	o.lock.Lock()
	img := image1bit.NewVerticalLSB(o.dev.Bounds())

	fontTTF, _ := truetype.Parse(goregular.TTF)
	face := truetype.NewFace(fontTTF, &truetype.Options{
		Size: 24,
		DPI:  72,
	})

	smallFace := basicfont.Face7x13

	//y := img.Bounds().Dy() - 1 - f.Descent
	temperature := font.Drawer{
		Dst:  img,
		Src:  &image.Uniform{image1bit.On},
		Face: face,
		Dot:  fixed.P(0, 24),
	}

	legend := font.Drawer{
		Dst:  img,
		Src:  &image.Uniform{image1bit.On},
		Face: smallFace,
		Dot:  fixed.P(0, 34),
	}

	st := font.Drawer{
		Dst:  img,
		Src:  &image.Uniform{image1bit.On},
		Face: face,
		Dot:  fixed.P(0, 60),
	}

	if state == "Off" {
		temperature.DrawString(fmt.Sprintf("--         %02d", at))
	} else {
		temperature.DrawString(fmt.Sprintf("%02d         %02d", tt, at))
	}

	legend.DrawString("target     actual")

	st.DrawString(fmt.Sprintf("      %s", state))
	if err := o.dev.Draw(o.dev.Bounds(), img, image.Point{}); err != nil {
		log.Fatal(err)
	}
	o.lock.Unlock()
}
