# Note that `#<` is an automatic variable holding the first dependency
TARGET ?= xiao-esp32c3

.PHONY: hello
hello: 00-hello/hello.go
	go build -o /dev/null $<
	tinygo flash --monitor --target=$(TARGET) $<

.PHONY: buzzer
buzzer: 01-buzzer/buzzer.go
	go build -o /dev/null $<
	tinygo flash --monitor --target=$(TARGET) $<

.PHONY: touch
touch: 02-touch/touch.go
	go build -o /dev/null $<
	tinygo flash --monitor --target=$(TARGET) $<

.PHONY: light
light: 03-light/light.go
	go build -o /dev/null $<
	tinygo flash --monitor --target=$(TARGET) $<
