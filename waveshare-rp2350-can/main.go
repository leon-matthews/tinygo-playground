// Package main is an example for one or two Waveshare RP2350 CAN boards
package main

import (
	"fmt"
	"machine"
	"time"

	"tinygo.org/x/drivers/mcp2515"
)

func main() {
	// Onboard LED
	led := machine.GPIO25
	led.Configure(machine.PinConfig{Mode: machine.PinOutput})
	led.High()

	// Configure SPI for CAN
	cs := machine.GPIO9
	spi := machine.SPI1

	err := spi.Configure(machine.SPIConfig{
		Frequency: 1000000, // 1 MHz is safe for initial testing
		SCK:       machine.GPIO10,
		SDO:       machine.GPIO11,
		SDI:       machine.GPIO12,
		Mode:      0, // XL2515 relies on SPI Mode 0
	})
	if err != nil {
		fatal("Failed to configure SPI1:" + err.Error())
		return
	}

	// MCP2515 CAN driver
	can := mcp2515.New(spi, cs)
	can.Configure(mcp2515.Configuration{})
	err = can.Begin(mcp2515.CAN500kBps, mcp2515.Clock8MHz)
	if err != nil {
		fatal("Begin CAN: " + err.Error())
	}

	// Bidirectional TX/RX Test
	// LED on during real TX, LED off for loopback
	for {
		mode, err := can.Mode()
		if err != nil {
			fatal(err.Error())
		}

		if mode == mcp2515.ModeNormal {
			// Send "Hello" to the real world
			data := []byte("Hello")
			err = can.Tx(0x111, uint8(len(data)), data)
			if err != nil {
				fmt.Println("CAN: " + err.Error())
			}

			// Switch back to loopback
			led.Low()
			err = can.SetMode(mcp2515.ModeLoopback)
			if err != nil {
				fatal("Set mode: " + err.Error())
			}
		} else if mode == mcp2515.ModeLoopback {
			// Send "Loopback" to myself
			data := []byte("Loopback")
			err = can.Tx(0x100, uint8(len(data)), data)
			if err != nil {
				fmt.Println("CAN: " + err.Error())
			}

			// Switch back to normal mode
			led.High()
			err = can.SetMode(mcp2515.ModeNormal)
			if err != nil {
				fatal("Set mode: " + err.Error())
			}
		}

		// CAN RX?
		if can.Received() {
			msg, err := can.Rx()
			if err != nil {
				fmt.Println("CAN:" + err.Error())
			}
			fmt.Printf("CAN-ID: %03X dlc: %d data: %s\n", msg.ID, msg.Dlc, msg.Data)
		}

		time.Sleep(time.Second * 1)
	}
}

func fatal(msg string) {
	for {
		fmt.Println("FATAL " + msg)
		time.Sleep(1 * time.Second)
	}
}
