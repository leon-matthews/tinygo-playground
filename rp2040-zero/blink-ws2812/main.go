// Blink the onboard WS2812 LED
package main

import (
    "image/color"
    "machine"
    "time"

    "tinygo.org/x/drivers/ws2812"
)

var (
    red = color.RGBA{R: 0xff}
    blue = color.RGBA{B: 0xff}
    green = color.RGBA{G: 0xff}
    colors = []color.RGBA{red, blue, green}
)

func main() {
    var neo machine.Pin = machine.NEOPIXEL
    neo.Configure(machine.PinConfig{Mode: machine.PinOutput})
    ws := ws2812.NewWS2812(neo)
    for {
        for i := range colors {
            ws.WriteColors(colors[i:i+1])
            time.Sleep(1 * time.Second)
        }
    }
}
