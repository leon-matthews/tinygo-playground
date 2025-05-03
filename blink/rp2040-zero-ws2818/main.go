package main

import (
	"image/color"
	"time"

	"machine"

	"tinygo.org/x/drivers/ws2812"
)

func main() {
	neo := ws2812.NewWS2812(machine.NEOPIXEL)
	neo.Pin.Configure(machine.PinConfig{Mode: machine.PinOutput})

	red := color.RGBA{R: 0xff, G: 0x00, B: 0x00}
	green := color.RGBA{R: 0x00, G: 0xff, B: 0x00}
	blue := color.RGBA{R: 0x00, G: 0x00, B: 0xff}

	colors := make([]color.RGBA, 1, 1)

	for {
        println("CPU frequency:", machine.CPUFrequency()/1e6, "MHz")

		println("Red!")
		colors[0] = red
		neo.WriteColors(colors)
		time.Sleep(time.Second)

		println("Green!")
		colors[0] = green
		neo.WriteColors(colors)
		time.Sleep(time.Second)

		println("Blue!")
		colors[0] = blue
		neo.WriteColors(colors)
		time.Sleep(time.Second)
	}
}
