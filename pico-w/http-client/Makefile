TARGET := pico-w

.PHONY: all clean

all: main.uf2

clean:
	rm -f main.elf main.uf2

main.elf:
	tinygo build -size short -target $(TARGET) main.go

main.uf2: main.elf
	picotool uf2 convert main.elf main.uf2
