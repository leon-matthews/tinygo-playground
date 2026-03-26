package main

import (
	"machine"
	"time"
)

var period uint64 = 1e9 / 500

func main() {
	pwm := ledPWM

	// Configure the PWM with the given period.
	pwm.Configure(machine.PWMConfig{
		Period: period,
	})

	ch, err := pwm.Channel(ledPin)
	if err != nil {
		println(err.Error())
		return
	}

	for { 
		for i := 1; i < 255; i++ {
            // This performs a stylish fade-out blink
			pwm.Set(ch, pwm.Top()/uint32(i))
			time.Sleep(time.Millisecond * 5)
		}
	}
}
