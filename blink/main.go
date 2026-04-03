package main

import (
	"time"

    "machine"
)


func main() {
    led.Configure(machine.PinConfig{Mode: machine.PinOutput})

    for {
        led.High()
        time.Sleep(time.Second)

        led.Low()
        time.Sleep(time.Second)
	}
}
