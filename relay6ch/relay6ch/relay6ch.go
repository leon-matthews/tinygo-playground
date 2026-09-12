// Package relay6ch drives the relays of a Waveshare RP2350-Relay-6CH board.
//
// The board carries an RP2350B, so it needs a TinyGo target built for that chip
// rather than for the 30-pin RP2350A: relays five and six live on GPIO30 and
// GPIO31, which an RP2350A does not have. Build with -target=pico-plus2, which
// pairs the rp2350b chip support with this board's 16 MB of flash.
//
// Call New followed by Configure before using any other method:
//
//	board := relay6ch.New()
//	if err := board.Configure(relay6ch.Config{}); err != nil {
//		// handle error
//	}
//	board.Set(relay6ch.CH1, true)
package relay6ch

import (
	"errors"
	"machine"
)

// Pin assignments of the RP2350-Relay-6CH board.
const (
	Relay1Pin = machine.GPIO26
	Relay2Pin = machine.GPIO27
	Relay3Pin = machine.GPIO28
	Relay4Pin = machine.GPIO29
	Relay5Pin = machine.GPIO30
	Relay6Pin = machine.GPIO31

	// The buzzer is channel B of PWM slice 3, so pair it with machine.PWM3.
	// Driving it is the chime package's job, not this one's.
	BuzzerPin = machine.GPIO23

	// The board's RS485 transceiver sits on UART1, and the vendor firmware
	// talks Modbus RTU over it at 9600 8N1. This package leaves both pins
	// alone, so they are yours to configure. The transceiver switches
	// direction on its own; there is no driver-enable pin to manage.
	RS485TXPin = machine.GPIO24
	RS485RXPin = machine.GPIO25
)

// NumChannels is the number of relays on the board.
const NumChannels = 6

// Channel identifies one of the board's six relays.
type Channel uint8

// Relay channels, numbered as on the board's silkscreen.
const (
	CH1 Channel = iota
	CH2
	CH3
	CH4
	CH5
	CH6
)

// ErrInvalidChannel is returned for a channel outside the range CH1 to CH6.
var ErrInvalidChannel = errors.New("relay6ch: invalid channel")

// Config holds the board options applied by Configure.
//
// It carries no options yet, and exists so that adding one later leaves
// Configure's signature alone.
type Config struct{}

// Device is a handle on one board's relays.
type Device struct {
	relays [NumChannels]machine.Pin
	state  uint8 // one bit per relay, bit 0 is CH1
}

// New returns a Device for the board's fixed pin assignment.
//
// The hardware is untouched until Configure is called.
func New() *Device {
	return &Device{
		relays: [NumChannels]machine.Pin{
			Relay1Pin, Relay2Pin, Relay3Pin, Relay4Pin, Relay5Pin, Relay6Pin,
		},
	}
}

// Configure prepares the relay outputs, leaving every relay de-energised.
//
// The Config argument carries no options yet and is ignored.
func (d *Device) Configure(_ Config) error {
	for _, pin := range d.relays {
		pin.Configure(machine.PinConfig{Mode: machine.PinOutput})
		pin.Low()
	}
	d.state = 0
	return nil
}

// Set energises or de-energises one relay.
func (d *Device) Set(ch Channel, on bool) error {
	if ch >= NumChannels {
		return ErrInvalidChannel
	}
	d.set(ch, on)
	return nil
}

// Toggle flips one relay.
func (d *Device) Toggle(ch Channel) error {
	if ch >= NumChannels {
		return ErrInvalidChannel
	}
	d.set(ch, d.state&(1<<ch) == 0)
	return nil
}

// Get reports whether a relay is energised, and false for an invalid channel.
func (d *Device) Get(ch Channel) bool {
	if ch >= NumChannels {
		return false
	}
	return d.state&(1<<ch) != 0
}

// SetMask drives all six relays, bit 0 of mask holding the state of CH1.
//
// Bits above CH6 are ignored. The relays switch one at a time in channel
// order, not simultaneously.
func (d *Device) SetMask(mask uint8) {
	for ch := Channel(0); ch < NumChannels; ch++ {
		d.set(ch, mask&(1<<ch) != 0)
	}
}

// Mask returns the state of all six relays, bit 0 holding the state of CH1.
func (d *Device) Mask() uint8 {
	return d.state
}

// AllOn energises every relay.
func (d *Device) AllOn() {
	d.SetMask(1<<NumChannels - 1)
}

// AllOff de-energises every relay.
func (d *Device) AllOff() {
	d.SetMask(0)
}

// set drives one in-range relay and records its new state.
func (d *Device) set(ch Channel, on bool) {
	d.relays[ch].Set(on)
	if on {
		d.state |= 1 << ch
	} else {
		d.state &^= 1 << ch
	}
}
