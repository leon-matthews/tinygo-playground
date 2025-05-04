package main

import (
	"time"

	"machine"

	pio "github.com/tinygo-org/pio/rp2-pio"
	"github.com/tinygo-org/pio/rp2-pio/piolib"

    "blink/rainbow"
)

func main() {
    led := newNeoPixel(machine.NEOPIXEL)

    for {
        for colour := range rainbow.Rainbow() {
            led.PutRaw(colour)
            time.Sleep(time.Millisecond * 50)
        }
    }
}


func newNeoPixel(pin machine.Pin) *piolib.WS2812B {
    sm, err := pio.PIO0.ClaimStateMachine()
    if err != nil {
		panic(err.Error())
	}

	ws, err := piolib.NewWS2812B(sm, pin)
	if err != nil {
		panic(err.Error())
	}

    return ws
}
