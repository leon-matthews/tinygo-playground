// Package chime plays short tunes on a PWM-driven buzzer.
//
// Build a Player from the PWM slice and pin the buzzer sits on, then call the
// tune methods on their own goroutine:
//
//	player, err := chime.New(machine.PWM3, machine.GPIO23)
//	if err != nil {
//		// handle error
//	}
//	go player.Sunrise()
//
// Only one thing sounds at a time. Starting a tune while another is playing
// replaces it, so the most recent event is always the one heard rather than
// the one that has to wait its turn.
package chime

import (
	"machine"
	"sync"
	"sync/atomic"
	"time"

	"tinygo.org/x/drivers/tone"
)

// DefaultBeepHz is the frequency Beep sounds.
//
// Small piezo transducers resonate somewhere near this, and it is the
// frequency the Waveshare RP2350-Relay-6CH vendor firmware uses, so it is
// about as loud as such a buzzer gets.
const DefaultBeepHz = 3276

// Player sounds tunes on one buzzer, one at a time.
//
// The play methods block for as long as the sound lasts, so call them on their
// own goroutine. A Player is safe to use from several goroutines at once: each
// new sound cancels the one playing at its next note boundary.
type Player struct {
	speaker tone.Speaker
	mu      sync.Mutex

	gen       uint32 // generation of the most recent request
	transpose int32  // semitones added to every note of a tune
}

// New returns a Player driving a buzzer on the given PWM slice and pin.
//
// The pin must be one of that slice's channels. The buzzer is left silent.
func New(pwm tone.PWM, pin machine.Pin) (*Player, error) {
	speaker, err := tone.New(pwm, pin)
	if err != nil {
		return nil, err
	}
	speaker.Stop()
	return &Player{speaker: speaker}, nil
}

// SetTranspose shifts every note of every tune by a number of semitones.
//
// A piezo buzzer is loudest at its resonant frequency and can be very quiet an
// octave below it, so a positive offset often carries much further than the
// written pitch. Notes are clamped to the range the tone package covers. The
// offset applies from the next tune onwards.
func (p *Player) SetTranspose(semitones int8) {
	atomic.StoreInt32(&p.transpose, int32(semitones))
}

// Play sounds a tune, blocking until it finishes.
//
// It returns early, leaving the new sound to take over, if another is started
// in the meantime.
func (p *Player) Play(t Tune) {
	gen := atomic.AddUint32(&p.gen, 1)

	p.mu.Lock()
	defer p.mu.Unlock()

	semitones := atomic.LoadInt32(&p.transpose)
	for i := 0; i < t.Steps(); i++ {
		// Checked per note rather than per tick, which keeps cancellation
		// free of timers and so of any allocation.
		if atomic.LoadUint32(&p.gen) != gen {
			return
		}
		note, d := t.Step(i)
		if note == 0 {
			p.speaker.Stop()
		} else {
			p.speaker.SetNote(shiftNote(note, semitones))
		}
		time.Sleep(d)
	}
	p.speaker.Stop()
}

// Beep sounds the buzzer at DefaultBeepHz for the given duration.
func (p *Player) Beep(d time.Duration) {
	p.Tone(DefaultBeepHz, d)
}

// Tone sounds one frequency in hertz for the given duration, blocking until it
// finishes.
//
// SetTranspose does not apply, the frequency being given outright. A frequency
// of zero is silence.
func (p *Player) Tone(hz uint32, d time.Duration) {
	gen := atomic.AddUint32(&p.gen, 1)

	p.mu.Lock()
	defer p.mu.Unlock()

	if atomic.LoadUint32(&p.gen) != gen {
		return
	}
	if hz == 0 {
		p.speaker.Stop()
	} else {
		p.speaker.SetPeriod(uint64(1e9) / uint64(hz))
	}
	time.Sleep(d)
	p.speaker.Stop()
}

// Stop silences the buzzer and cancels whatever is playing.
//
// It waits for the cancelled tune to reach its next note boundary, so it can
// block for as long as that note lasts.
func (p *Player) Stop() {
	atomic.AddUint32(&p.gen, 1)

	p.mu.Lock()
	defer p.mu.Unlock()

	p.speaker.Stop()
}

// Sunrise plays the power-on jingle.
func (p *Player) Sunrise() { p.Play(Sunrise) }

// RelayOn marks one relay energising.
func (p *Player) RelayOn() { p.Play(RelayOn) }

// RelayOff marks one relay releasing.
func (p *Player) RelayOff() { p.Play(RelayOff) }

// AllOn marks every relay energising.
func (p *Player) AllOn() { p.Play(AllOn) }

// AllOff marks every relay releasing.
func (p *Player) AllOff() { p.Play(AllOff) }

// LinkUp marks the network coming up.
func (p *Player) LinkUp() { p.Play(LinkUp) }

// LinkLost marks the network going away.
func (p *Player) LinkLost() { p.Play(LinkLost) }

// Ack confirms an accepted command.
func (p *Player) Ack() { p.Play(Ack) }

// Rejected refuses a command.
func (p *Player) Rejected() { p.Play(Rejected) }

// Warning reports a recoverable fault.
func (p *Player) Warning() { p.Play(Warning) }

// Fault reports an unrecoverable fault.
func (p *Player) Fault() { p.Play(Fault) }

// shiftNote shifts a note, clamped to the range the tone package covers.
func shiftNote(n tone.Note, semitones int32) tone.Note {
	shifted := int32(n) + semitones
	if shifted < int32(tone.A0) {
		return tone.A0
	}
	if shifted > int32(tone.B8) {
		return tone.B8
	}
	return tone.Note(shifted)
}
