package main

import (
	"time"

	"local.dev/grove/grove"
)

const position = 2

var (
	light  grove.LightSensor
	shield grove.ShieldXiao
)

func main() {
	// Light sensor returns an analog voltage
	const numSamples = 32
	light.Configure(shield.Connector(position).ADC(), numSamples)

	for {
		lux, saturated := light.Lux()
		if saturated {
			println(">", grove.MaxLux, "lux")
		} else {
			println(lux, "lux")
		}
		time.Sleep(3 * time.Second)
	}
}
