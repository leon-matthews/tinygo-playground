// Package main is a minimal TinyGo example
package main

import (
	"time"
)

func main() {
	for {
		time.Sleep(3 * time.Second)
		println("Hello, world!")
	}
}
