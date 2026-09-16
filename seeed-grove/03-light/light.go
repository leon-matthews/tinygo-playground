package main

import (
	"time"

	"local.dev/grove/grove"
)

const position = 2

var shield grove.ShieldXiao

func main() {
    // Light sensor returns an analog voltage
    adc := shield.Connector(2).ADC()

	for {
		println("Light?", adc.ReadAnalogValue())
		time.Sleep(3 * time.Second)
	}
}
