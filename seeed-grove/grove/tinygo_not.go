//go:build !tinygo
package grove

var pinValues [256]bool

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
