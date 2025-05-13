package rainbow

import (
    "iter"
)

// Rainbow generates all hexadecimal strings of the given length, zero-padded.
func Rainbow() iter.Seq[uint32] {

    // I have swapped red and green for the LED on the Waveshare RP2040 Zero
    pack := func(r, g, b uint8) uint32 {
        colour := uint32(r)<<24 | uint32(g)<<16 | uint32(b)<<8
        return colour
    }

	return func(yield func(uint32) bool) {
        for {
            // Start with red
            var (
                r uint8 = 255
                g, b uint8
            )

            // Red to yellow
            println("Red")
            for _ = range 255 {
                g++
                if !yield(pack(r, g, b)) {
                    return
                }
            }

            // Yellow to green
            println("Yellow")
            for _ = range 255 {
                r--
                if !yield(pack(r, g, b)) {
                    return
                }
            }

            // Green -> cyan
            println("Green")
            for _ = range 255 {
                b++
                if !yield(pack(r, g, b)) {
                    return
                }
            }

            // Cyan -> blue
            println("Cyan")
            for _ = range 255 {
                g--
                if !yield(pack(r, g, b)) {
                    return
                }
            }

            // Blue -> Purple
            println("Blue")
            for _ = range 255 {
                r++
                if !yield(pack(r, g, b)) {
                    return
                }
            }

            // Purple -> Red
            println("Purple")
            for _ = range 255 {
                b--
                if !yield(pack(r, g, b)) {
                    return
                }
            }
        }
    }
}
