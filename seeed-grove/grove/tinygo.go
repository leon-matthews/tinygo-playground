//go:build tinygo
package grove

import (
    "sync"
    "time"

    "machine"
)

var onceADC sync.Once

// analog returns an accessor function to read ADC input
func analog(p uint8) ADC {
    // Initialise the ADC periphial... once.
    onceADC.Do(machine.InitADC)

    var sensor machine.ADC
    sensor.Pin = devicePins[p]
    err := sensor.Configure(machine.ADCConfig{
        Reference: 0,
        Samples: 0,
        SampleTime: 0,
    })
    if err != nil {
        time.Sleep(1 * time.Second)
        println(err)
        panic(err)
    }

    return sensor.Get
}

// pinout returns device specific pin state mutator function
func pinout(p uint8) PinOutput {
	pin := devicePins[p]
	pin.Configure(machine.PinConfig{Mode: machine.PinOutput})
	return pin.Set
}

// pinInputPulldown configures input put with a pull-down
func pinInputPulldown(p uint8) PinInput {
    pin := devicePins[p]
    pin.Configure(machine.PinConfig{Mode: machine.PinInputPulldown})
    return pin.Get
}
