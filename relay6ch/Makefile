# The RP2350B needs an rp2350b target: pga2350 and metro-rp2350 also work.
TARGET ?= pico-plus2

# TinyGo 0.42 miscompiles goroutine wrappers for this target at the size
# optimisation levels, -opt=z and -opt=s, and -opt=z is the default. See the
# README. Optimising for speed instead dodges it, and is smaller here anyway.
OPT ?= -opt=2

.PHONY: build clean flash monitor

build: main.uf2

clean:
	rm -f main.elf main.uf2

flash:
	tinygo flash $(OPT) -target $(TARGET) -monitor .

monitor: main.uf2
	tinygo monitor -target $(TARGET)

main.elf: $(wildcard *.go relay6ch/*.go chime/*.go)
	tinygo build -size short $(OPT) -target $(TARGET) -o main.elf .

main.uf2: main.elf
	picotool uf2 convert main.elf main.uf2
