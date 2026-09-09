package main

import (
	"fmt"
	"time"

	"tinygo.org/x/drivers/mcp2515"

	"machine"
)

func main() {
	// Onboard LED
	led := machine.GPIO25
	led.Configure(machine.PinConfig{Mode: machine.PinOutput})

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
		failMessage("Failed to configure SPI1:" + err.Error())
		return
	}

	// MCP2515 CAN driver
	can := mcp2515.New(spi, cs)
	can.Configure(mcp2515.Configuration{})
	err = can.Begin(mcp2515.CAN500kBps, mcp2515.Clock8MHz)
	if err != nil {
		failMessage("Begin CAN: " + err.Error())
	}

	// Read CNF1 register from CAN controller
	// LED on during CAN TX & RX
	for {
		// LED on
		led.High()

		// CAN TX
		err = can.Tx(0x111, 8, []byte{0x00, 0xAA, 0x55, 0xAA, 0x55, 0xAA, 0x55, 0xAA})
		if err != nil {
			failMessage("CAN: " + err.Error())
		}

		// CAN RX?
		if can.Received() {
			msg, err := can.Rx()
			if err != nil {
				failMessage("CAN:" + err.Error())
			}
			fmt.Printf("CAN-ID: %03X dlc: %d data: ", msg.ID, msg.Dlc)
			for _, b := range msg.Data {
				fmt.Printf("%02X ", b)
			}
			fmt.Println()
		}

		// LED off
		led.Low()

		time.Sleep(time.Second * 1)
	}
}

func failMessage(msg string) {
	for {
		println(msg)
		time.Sleep(1 * time.Second)
	}
}
