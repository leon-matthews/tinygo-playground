package main

import (
    "time"

    "local.dev/grove/grove"
)

var shield grove.ShieldXiao

func main() {
    // Touch sensor pulls data line high while a finger is present
    touch := shield.Connector(0).PinInputPulldown()

    lastPrint := time.Now()
    last := false
    for {
        beingTouched := touch.GetLevel()
        if beingTouched && time.Since(lastPrint) > time.Second {
            // Print status only once a second
            println("Somebody is touching me!")
            lastPrint = time.Now()
        } else if last && !beingTouched {
            // First loop (only) after letting go
            println("I'm getting cold, please touch me.")
        }

        last = beingTouched
        time.Sleep(10 * time.Millisecond)
    }
}
