package main

import "time"

// backoff waits longer after each failure, doubling from minWait up to maxWait.
//
// consecutive counts the failures so far, starting at zero.
func backoff(consecutive uint, minWait, maxWait time.Duration) time.Duration {
	const maxDoublings = 20 // Ensure we never overflow
	return min(minWait<<min(consecutive, maxDoublings), maxWait)
}

// connectBackoff spaces out attempts to bring the network up from scratch.
func connectBackoff(consecutive uint) time.Duration {
	return backoff(consecutive, 1*time.Second, 1*time.Minute)
}

// protocolBackoff spaces out retries of request/response protocols such as DHCP and NTP.
func protocolBackoff(consecutive uint) time.Duration {
	return backoff(consecutive, 100*time.Microsecond, 20*time.Millisecond)
}
