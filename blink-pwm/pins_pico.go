//go:build pico || pico2

package main

import "machine"

const ledPin = machine.LED

var ledPWM = machine.PWM4
