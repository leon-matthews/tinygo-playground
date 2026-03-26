TARGET ?= pico
export GOTOOLCHAIN := go1.25.8

.PHONY: build clean flash monitor

build: main.uf2

clean:
	rm -f main.elf main.uf2

flash: main.uf2
	tinygo flash -target $(TARGET) -monitor

monitor: main.uf2
	tinygo monitor -target $(TARGET)

main.elf: main.go pins_pico.go pins_waveshare_rp2040_zero.go
	tinygo build -size short -target $(TARGET) -o main.elf .

main.uf2: main.elf
	picotool uf2 convert main.elf main.uf2
