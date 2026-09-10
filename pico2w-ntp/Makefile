TARGET ?= pico2-w

# WiFi credentials, passed into the firmware at link time.
LDFLAGS := -X 'main.wifiSSID=$(WIFI_SSID)' -X 'main.wifiPassword=$(WIFI_PASSWORD)'

# Abort the build unless the named variable holds a non-empty value.
require = $(if $(strip $($1)),,$(error $1 is not set - try: export $1=...))

TINYGO_FLAGS ?= -target $(TARGET) -scheduler=tasks -ldflags="$(LDFLAGS)"
#export GOTOOLCHAIN := go1.27.1

.PHONY: build check-credentials clean flash monitor

build: main.uf2

check-credentials:
	@$(call require,WIFI_SSID)
	@$(call require,WIFI_PASSWORD)

clean:
	rm -f main.elf main.uf2

flash: check-credentials
	tinygo flash $(TINYGO_FLAGS) -monitor .

monitor: main.uf2
	tinygo monitor -target $(TARGET)

main.elf: $(wildcard *.go) | check-credentials
	tinygo build $(TINYGO_FLAGS) -size short -o main.elf .

main.uf2: main.elf
	picotool uf2 convert main.elf main.uf2
