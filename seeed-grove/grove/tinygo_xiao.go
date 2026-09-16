//go:build xiao_esp32c3 || xiao_rp2350

package grove

import "machine"

// devicePins maps a Xiao DX connector index to its board-specific [machine.Pin].
var devicePins = [...]machine.Pin{
    machine.D0,
    machine.D1,
    machine.D2,
    machine.D3,
    machine.D4,
    machine.D5,
	machine.D6,
	machine.D7,
	machine.D8,
	machine.D9,
	machine.D10,
}
