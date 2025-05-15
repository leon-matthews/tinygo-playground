TARGET := waveshare-rp2040-zero

.PHONY: build clean flash monitor

build: main.elf

clean:
	rm -f main.elf main.uf2

flash: main.uf2
	tinygo flash -target $(TARGET)

monitor: main.uf2
	tinygo monitor -target $(TARGET)

main.elf: main.go
	tinygo build -size short -target $(TARGET) main.go

main.uf2: main.elf
	picotool uf2 convert main.elf main.uf2
