package chime

import (
	"time"

	"tinygo.org/x/drivers/tone"
)

// Tune is a sequence of notes, each held for a fixed time.
//
// The encoding is two bytes per step: a MIDI note number, then a duration in
// units of 5 ms. A note of zero is a rest. Every note the tone package defines
// is below 128, so each byte is a one-byte string constant and a whole tune
// folds at compile time into read-only data, costing no RAM.
type Tune string

// unitMS is the duration quantum of an encoded step, in milliseconds.
const unitMS = 5

// Note bytes, named for the pitch they encode.
const (
	rest = string(rune(0))

	c5  = string(rune(tone.C5))  // 523 Hz
	e5  = string(rune(tone.E5))  // 659 Hz
	f5  = string(rune(tone.F5))  // 698 Hz
	g5  = string(rune(tone.G5))  // 784 Hz
	a5  = string(rune(tone.A5))  // 880 Hz
	c6  = string(rune(tone.C6))  // 1047 Hz
	cs6 = string(rune(tone.CS6)) // 1109 Hz
	e6  = string(rune(tone.E6))  // 1319 Hz
	f6  = string(rune(tone.F6))  // 1397 Hz
	g6  = string(rune(tone.G6))  // 1568 Hz
	a6  = string(rune(tone.A6))  // 1760 Hz
	c7  = string(rune(tone.C7))  // 2093 Hz
)

// Duration bytes, named for the milliseconds they encode.
const (
	d35  = string(rune(35 / unitMS))
	d40  = string(rune(40 / unitMS))
	d45  = string(rune(45 / unitMS))
	d50  = string(rune(50 / unitMS))
	d60  = string(rune(60 / unitMS))
	d70  = string(rune(70 / unitMS))
	d80  = string(rune(80 / unitMS))
	d90  = string(rune(90 / unitMS))
	d110 = string(rune(110 / unitMS))
	d140 = string(rune(140 / unitMS))
	d170 = string(rune(170 / unitMS))
	d300 = string(rune(300 / unitMS))
	d320 = string(rune(320 / unitMS))
)

// The tunes share one vocabulary, so the board can be understood without being
// watched: rising means energised or connected and falling means released or
// lost, two notes is one channel where three is all of them, a gap in the
// middle marks a network event rather than a relay, and a pair said twice is a
// fault. Every pitch sits between 698 Hz and 2093 Hz.
const (
	// Sunrise is the power-on jingle, a C major arpeggio over two octaves.
	Sunrise = Tune(c5 + d70 + e5 + d70 + g5 + d70 + c6 + d70 +
		e6 + d70 + g6 + d70 + rest + d50 + c7 + d320)

	// RelayOn marks one relay energising, on a rising perfect fifth.
	RelayOn = Tune(c6 + d40 + g6 + d90)

	// RelayOff marks one relay releasing, and is RelayOn reversed.
	RelayOff = Tune(g6 + d40 + c6 + d90)

	// AllOn marks every relay energising, RelayOn with the octave on top.
	AllOn = Tune(c6 + d40 + g6 + d40 + c7 + d140)

	// AllOff marks every relay releasing, and is AllOn reversed.
	AllOff = Tune(c7 + d40 + g6 + d40 + c6 + d140)

	// LinkUp marks the network coming up, the gap saying no relay moved.
	LinkUp = Tune(g6 + d60 + rest + d45 + c7 + d170)

	// LinkLost marks the network going away, and is LinkUp reversed.
	LinkLost = Tune(c7 + d60 + rest + d45 + g6 + d170)

	// Ack confirms an accepted command, and is the shortest tune here.
	Ack = Tune(e6 + d35)

	// Rejected refuses a command, on two flat notes that go nowhere.
	Rejected = Tune(f5 + d45 + rest + d35 + f5 + d45)

	// Warning reports a recoverable fault, repeating a falling minor third.
	Warning = Tune(a6 + d80 + f6 + d80 + rest + d70 + a6 + d80 + f6 + d80)

	// Fault reports an unrecoverable one, on the only tritone in the set.
	Fault = Tune(g6 + d110 + cs6 + d110 + rest + d70 + g6 + d110 + cs6 + d300)
)

// Steps returns the number of notes and rests in the tune.
func (t Tune) Steps() int {
	return len(t) / 2
}

// Step returns the note and duration of step i, counting from zero.
//
// The note is zero for a rest. Step panics if i is out of range.
func (t Tune) Step(i int) (tone.Note, time.Duration) {
	return tone.Note(t[i*2]), time.Duration(t[i*2+1]) * unitMS * time.Millisecond
}

// Duration returns how long the tune takes to play.
func (t Tune) Duration() time.Duration {
	var total time.Duration
	for i := 0; i < t.Steps(); i++ {
		_, d := t.Step(i)
		total += d
	}
	return total
}
