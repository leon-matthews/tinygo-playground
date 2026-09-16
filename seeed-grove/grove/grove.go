// Package grove supports for the Seeed Grove ecosystem on XIAO boards with TinyGo.
package grove

import "math"

// ADC is a function that returns a pin's analog value
type ADC func() uint16
func (a ADC) ReadAnalogValue() uint16 { return a() }
func (a ADC) ReadAnalogFraction() float32 { return float32(a()) / math.MaxUint16 }

// PinInput is a function that returns the pin's current level
type PinInput func() bool
func (p PinInput) GetLevel() bool { return p() }

// PinOutput is a function that sets a pin's value
type PinOutput func(val bool)
func (p PinOutput) High() { p(true) }
func (p PinOutput) Low() { p(false) }

// Connector represents a Grove connector
type Connector struct {
	yellow uint8 // Highest data/GPIO pin
	white  uint8 // Next to yellow, towards centre
}

func (c Connector) ADC() ADC {
    return analog(c.yellow)
}

func (c Connector) PinOutput() PinOutput {
	return pinout(c.yellow)
}

func (c Connector) PinInputPulldown() PinInput {
    return pinInputPulldown(c.yellow)
}

// ShieldXiao represents the Seed Grove shield for its Xiao platform
type ShieldXiao struct{}

// Connector sets the yellow and white pin values for given position
func (s ShieldXiao) Connector(position uint8) Connector {
	c := Connector{}
	switch position {
	case 0, 1, 2:
	    // Digital & Analog IO
	    c.yellow, c.white = position, position + 1
	case 3, 4:
	    // I2C (Shared: SCL=5, SDA=4)
	    c.yellow, c.white = 5, 4
	case 5:
	    // UART (TX=6, RX=7)
	    c.yellow, c.white =  7, 6
    case 6, 7:
        // Digital IO
	    c.yellow, c.white = position + 2, position + 3
	default:
	    // No pins
	    c.yellow, c.white = 0xff, 0xff
	}
	return c
}
