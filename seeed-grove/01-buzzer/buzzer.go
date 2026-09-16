// Package main demonstrates the 'O' in 'GPIO' by toggling a Grove buzzer
package main

import (
	"time"

	"local.dev/grove/grove"
)

const (
	position = 1
)

var shield grove.ShieldXiao

func main() {
	buzzer := shield.Connector(position).PinOutput()

	for {
        beepBeep(buzzer)
        println("Beep Beep!")
        time.Sleep(5 * time.Second)
    }

}


// The Road Runner?
// 880Hz for 120ms, 60ms gap, then 1047Hz for 150ms
func beepBeep(pin grove.PinOutput) {
    beep(pin, halfPeriod(676), 145 * time.Millisecond)
    time.Sleep(75 * time.Millisecond)
    beep(pin, halfPeriod(720), 145 * time.Millisecond)
}

// halfPeriod calculates a sleep duration from a desired frequency
func halfPeriod(hz int) time.Duration {
    return time.Second / time.Duration(2*hz)
}

// beep bit-bangs an output with a simple sleep-driven loop
func beep(pin grove.PinOutput, halfPeriod, duration time.Duration) {
    for elapsed := time.Duration(0); elapsed < duration; elapsed += 2 * halfPeriod {
		pin.High()
		time.Sleep(halfPeriod)
		pin.Low()
		time.Sleep(halfPeriod)
    }
}
