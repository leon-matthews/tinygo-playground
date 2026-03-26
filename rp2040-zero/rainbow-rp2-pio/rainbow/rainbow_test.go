package rainbow_test

import (
    "fmt"
    "testing"

    "hello/rainbow"
)


func TestRainbow(t *testing.T) {
    t.Run("first few values", func(t *testing.T) {
        expected := 1530
        colours := make([]uint32, 0, expected)

        for colour := range rainbow.Rainbow() {
            fmt.Printf("%032b\n", colour)
            colours = append(colours, colour)
            if len(colours) == expected {
                break
            }
        }
    })
}
