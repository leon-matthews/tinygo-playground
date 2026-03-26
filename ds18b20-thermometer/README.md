# DS18B20 Thermometer

Reads one or more DS18B20 1-Wire temperature sensors and prints readings over
the USB serial debug connection.

## Parts

- Raspberry Pi Pico
- One or more DS18B20 sensors
- 4.7 kΩ resistor

## Wiring

```
DS18B20        Pico
-------        ----
VDD    ------> 3.3V (pin 36)
GND    ------> GND  (pin 38)
DQ     ------> GP26 (pin 31)
```

Place the 4.7 kΩ resistor between DQ and 3.3V (pin 36).

For multiple sensors, connect all DQ pins together on the same line — 1-Wire
supports multiple devices on a single data pin. One pull-up resistor is
sufficient regardless of how many sensors are connected.

## Build & flash

```
make flash
```

On startup the serial monitor will print the ROM ID of each sensor found, then
print temperature readings every ~2 seconds.
