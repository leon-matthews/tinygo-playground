# relay6ch

A TinyGo demo for the Waveshare RP2350-Relay-6CH board. `main.go` walks the relays in
turn, chiming as each one switches. The `relay6ch/` package drives the hardware and the
`chime/` package plays short tunes on the buzzer.

## Building

The board carries an **RP2350B**, not the 30-pin RP2350A found on a Pico 2. Two of the
relays sit on GPIO30 and GPIO31, which an RP2350A does not have, so `-target=pico2` fails
to compile:

    relay6ch/relay6ch.go:29:22: undefined: machine.GPIO30

Use `pico-plus2` instead. It inherits TinyGo's `rp2350b` chip support and declares the
16 MB of flash this board has:

    tinygo build -opt=2 -target=pico-plus2 -o main.elf .
    tinygo flash -opt=2 -target=pico-plus2 .

Any other `rp2350b` target works as well if you prefer one: `pga2350`, `metro-rp2350`.

`-opt=2` is not optional here. TinyGo 0.42 miscompiles goroutine wrappers for this target
at both size optimisation levels, and `-opt=z` is the default, so a plain `tinygo build`
of this program fails in LLVM. See [Toolchain](#toolchain) below. The `Makefile` passes
the flag for you.

## Hardware

| Function      | Pin              | Notes                                          |
| ------------- | ---------------- | ---------------------------------------------- |
| Relays CH1-6  | GPIO26 to GPIO31 | Push-pull output, active high, open at reset   |
| Buzzer        | GPIO23           | PWM slice 3, channel B                         |
| RS485         | GPIO24 / GPIO25  | UART1, not driven by this package              |

The RS485 transceiver switches direction by itself, so there is no driver-enable pin to
manage. The vendor firmware speaks Modbus RTU over it at 9600 8N1, listening as slave
`0x06` and accepting function `0x05` writes to coils `0x0001` through `0x0006`. This
package leaves UART1 alone; configure it yourself if you need it.

### Relay drive circuit

From the [board schematic][schematic], every channel is the same chain, and nothing
power-related hangs off the GPIO itself:

    GPIO --4K7-- SS8050 --> PC817X3 optocoupler --4K7-- SS8050 --> HLS8L-DC5V coil

The GPIO sources about 0.55 mA into the base of an SS8050, which sinks the LED of a
PC817X3CSP9F optocoupler along with the channel's red status LED. On the far side of the
barrier a second SS8050 switches the 5 V relay coil, which has an LL4148 flyback diode
across it. Contacts reach a 3-way terminal block, NO/COM/NC, rated 10 A at 250 VAC or
30 VDC.

The coil side runs on an isolated 5 V rail from a B0505S-3WR2 DC-DC converter, and shares
that domain with the RS485 transceiver. Three consequences are worth knowing:

- A 100K base-emitter pull-down holds the driver off whenever the GPIO is high impedance,
  which is its state from power-on until Configure runs. The relays cannot close during
  start-up.
- The red channel LED sits on the microcontroller side of the barrier, so it shows what
  was commanded rather than whether the coil actually pulled in.
- Six coils at once draw roughly 0.45 A of the converter's 0.6 A, shared with the RS485
  transceiver. AllOn is within budget, but it is the first thing to suspect if switching
  everything together upsets the RS485 link.

[schematic]: https://files.waveshare.com/wiki/RP2350-Relay-6CH/RP2350-Relay-6CH.pdf

## Usage

```go
import "local.dev/relay6ch/relay6ch"

board := relay6ch.New()
if err := board.Configure(relay6ch.Config{}); err != nil {
    // handle error
}

board.Set(relay6ch.CH1, true)     // close one relay
board.Toggle(relay6ch.CH2)        // flip one relay
board.SetMask(0b010101)           // drive all six at once, bit 0 is CH1
board.AllOff()
```

Relay state is tracked in the `Device` rather than read back from the pads, so `Toggle`
and `Get` stay correct and cost nothing. `Set` and `Toggle` return `ErrInvalidChannel`
for a channel outside CH1 to CH6; the mask methods ignore bits above CH6.

`Config` carries no options. It is there so that adding one later does not change
`Configure`'s signature.

## Chimes

The buzzer belongs to the `chime/` package, which plays short tunes on it through TinyGo's
`tone` driver. Build a player from the PWM slice and pin, then call a tune on its own
goroutine:

```go
import "local.dev/relay6ch/chime"

player, err := chime.New(machine.PWM3, relay6ch.BuzzerPin)
if err != nil {
    // handle error
}

go player.Sunrise()               // power-on jingle
go player.RelayOn()               // one relay energised
go player.Fault()                 // unrecoverable fault
player.Beep(100 * time.Millisecond)
```

Only one sound plays at a time, and starting another replaces the one playing at its next
note boundary, so the most recent event is always the one heard rather than the one that
waited its turn. The tunes share a vocabulary: rising means energised or connected and
falling means released or lost, two notes is one channel where three is all of them, a gap
in the middle marks a network event rather than a relay, and a pair said twice is a fault.

Every tune is a string constant of two bytes per step, a MIDI note number and a duration
in units of 5 ms, so the whole set folds into read-only data at compile time and costs no
RAM. Nothing is allocated while playing either.

The notes run from 698 Hz to 2093 Hz, but the fitted transducer resonates near 3276 Hz,
which is where the vendor firmware drives it. Everything here is therefore below the
loudest part of its range, the lowest notes furthest of all. `SetTranspose` shifts the
whole set in semitones if the tunes sound thin on the bench; `+7` puts the top note on the
resonant peak.

## Toolchain

`go vet` cannot be run against this code: pointing `cmd/go` at TinyGo's GOROOT makes vet
reject TinyGo's own runtime for using `//go:inline` in the standard library. Type
checking comes from `tinygo build` instead, with `revive` and `gofumpt` on top.

TinyGo 0.42 cannot build this program at the size optimisation levels. With
`-scheduler=tasks`, which is the default for `pico-plus2`, and either `-opt=z` or `-opt=s`,
LLVM rejects the module it is given:

    invalid llvm.used member
    i32 ptrtoint (ptr @"(*local.dev/relay6ch/chime.Player).AllOff$gowrapper" to i32)
    LLVM ERROR: Broken module found, compilation aborted!

The failure is not deterministic. The same source also aborts with `std::bad_alloc` or a
segmentation fault inside the compiler, which is what memory corruption tends to look
like from outside. It needs several `go player.Method()` call sites to show up, and goes
away if any one of them is removed or spelled differently, so it is a codegen bug rather
than anything about the Go being compiled.

`-opt=1` and `-opt=2` both build cleanly, and `-opt=2` is the smaller of the two here, so
the `Makefile` uses it. The other two schedulers are not a way out: `-scheduler=asyncify`
and `-scheduler=threads` both fail to link on this target for unrelated reasons.
