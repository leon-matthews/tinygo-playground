// Command relay6ch walks the relays of a Waveshare RP2350-Relay-6CH board in
// turn, chiming as each one switches, then switches all six together.
//
//	tinygo flash -target=pico-plus2 .
package main

import (
	"machine"
	"time"

	"local.dev/relay6ch/chime"
	"local.dev/relay6ch/relay6ch"
)

func main() {
	board := relay6ch.New()
	if err := board.Configure(relay6ch.Config{}); err != nil {
		fail("relay6ch", err)
	}

	player, err := chime.New(machine.PWM3, relay6ch.BuzzerPin)
	if err != nil {
		fail("chime", err)
	}

	// The fitted transducer resonates near 3.3 kHz, above everything these
	// tunes play, so raise this if they sound thin on the bench.
	player.SetTranspose(0)

	go player.Sunrise()
	time.Sleep(2 * time.Second)

	for {
		// Close one relay at a time, using the mask so that the previous
		// channel opens in the same call.
		for ch := relay6ch.CH1; ch <= relay6ch.CH6; ch++ {
			board.SetMask(1 << ch)
			go player.RelayOn()
			time.Sleep(time.Second)
		}

		board.AllOn()
		go player.AllOn()
		time.Sleep(time.Second)

		board.AllOff()
		go player.AllOff()
		time.Sleep(time.Second)
	}
}

// fail reports a start-up error for ever, there being nowhere else to send it.
func fail(what string, err error) {
	for {
		println(what+":", err.Error())
		time.Sleep(time.Second)
	}
}
