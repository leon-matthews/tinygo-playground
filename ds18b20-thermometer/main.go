package main

// DS18B20 1-Wire temperature sensor example.
//
// Connect one or more DS18B20 sensors to the oneWirePin defined for your target
// (see pins_*.go) with a 4.7 kΩ pull-up resistor to 3.3 V. Temperature
// readings are sent over the USB serial debug connection.

import (
	"fmt"
	"time"

	"tinygo.org/x/drivers/ds18b20"
	"tinygo.org/x/drivers/onewire"
)

func main() {
	wire := onewire.New(oneWirePin)
	wire.Configure(onewire.Config{})

	sensor := ds18b20.New(wire)
	sensor.Configure()

	fmt.Println("Searching for DS18B20 sensors...")

	devices, err := wire.Search(onewire.SEARCH_ROM)
	if err != nil || len(devices) == 0 {
		fmt.Println("No sensors found. Check wiring and pull-up resistor.")
		return
	}

	fmt.Printf("Found %d sensor(s):\n", len(devices))
	for i, rom := range devices {
		fmt.Printf("  Sensor %d: %X\n", i, rom)
	}

	for {
		// Broadcast conversion request — all sensors convert simultaneously.
		sensor.RequestTemperature(nil)
		time.Sleep(750 * time.Millisecond) // worst-case 12-bit conversion time

		for i, rom := range devices {
			temp, err := sensor.ReadTemperature(rom)
			if err != nil {
				fmt.Printf("Sensor %d: read error\n", i)
				continue
			}
			// temp is millidegrees Celsius; split into whole and fractional parts.
			whole := temp / 1000
			frac := temp % 1000
			if frac < 0 {
				frac = -frac
			}
			fmt.Printf("Sensor %d: %d.%03d°C\n", i, whole, frac)
		}

		time.Sleep(time.Second)
	}
}
