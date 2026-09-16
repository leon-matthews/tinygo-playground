//go:build !tinygo
package grove

import "math"

var pinValues [256]bool

// analog returns an accessor function to read ADC input
func analog(i uint8) ADC {
    return func() uint16 {
        if pinValues[i] {
            return math.MaxUint16
        }
        return 0
    }
}

// pinout returns a mutator function to set a pin's output
func pinout(i uint8) PinOutput {
    return func(level bool) {
        pinValues[i] = level
    }
}

// pinInputPulldown returns an accessor function to a pin's value
func pinInputPulldown(i uint8) PinInput {
    return func() bool {
        return pinValues[i]
    }
}
