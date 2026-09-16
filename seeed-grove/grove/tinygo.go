//go:build tinygo
package grove

import "machine"

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
