# Writing and Testing Drivers

This primer gives the local conventions of this repository. It assumes that you know Go well,
and that you have read [Driver Design for TinyGo](https://tinygo.org/docs/guides/driver-design)
for the general theory. This document tells you what the code in this repository actually does.

## Contents

1. [Keep the machine package out of your driver](#1-keep-the-machine-package-out-of-your-driver)
2. [Package layout and API shape](#2-package-layout-and-api-shape)
3. [Build tags, when you cannot avoid machine](#3-build-tags-when-you-cannot-avoid-machine)
4. [Tests](#4-tests)
5. [Examples and the smoke test](#5-examples-and-the-smoke-test)
6. [Repository conventions](#6-repository-conventions)
7. [Checklist](#7-checklist)

## 1. Keep the machine package out of your driver

The `machine` package exists only in TinyGo baremetal builds. The host Go toolchain cannot
compile it. This one fact controls the design of every driver here. If your driver imports
`machine` from a file that has no build tag, then two things break. `go test` cannot build the
package, and no one can use your driver from a host program.

The solution is to accept small interfaces. The root package declares them.

| Hardware   | Type to accept                  | Declared in    |
| ---------- | ------------------------------- | -------------- |
| I2C bus    | `drivers.I2C`                   | `i2c.go`       |
| SPI bus    | `drivers.SPI`                   | `spi.go`       |
| UART       | `drivers.UART`                  | `uart.go`      |
| Output pin | `pin.Output` or `pin.OutputFunc` | `internal/pin` |
| Input pin  | `pin.Input` or `pin.InputFunc`   | `internal/pin` |

`machine.I2C` and `machine.SPI` satisfy these interfaces structurally. The user gives you
`machine.I2C0`, and your driver never names the type. `drivers.I2C` has one method, and
`drivers.SPI` has two.

### Pins need a hardware abstraction layer

`machine.Pin` is a numeric value type, not an interface, so you cannot abstract it with a type
assertion. Use the pin HAL in `internal/pin`, which gives you two styles of the same idea:

```go
type Output interface{ Set(level bool) }
type OutputFunc func(level bool)
```

The function type is convenient for the caller, because a method value converts directly:

```go
cs := machine.D10
cs.Configure(machine.PinConfig{Mode: machine.PinOutput})
dev := mydriver.New(machine.SPI0, pin.OutputFunc(cs.Set))
```

Use the interface type when you want the caller to pass an object, as `mcp2515.New` does
(`mcp2515/mcp2515.go:57`). Use a struct of function fields when one pin must change mode during
operation. `unoqmatrix.CharlieplexPin` (`unoqmatrix/matrix.go:65`) holds a `Set` and a `Float`
field for exactly this reason, because a charlieplexed pin must also float.

`internal/pin` is an internal package, so only drivers in this repository can import it. Code
outside cannot. Some drivers declare an equivalent interface in their own package instead, such
as `sharpmem.Pin` (`sharpmem/sharpmem.go:29`) and `gdew0154m09.OutputPin`. Both choices are
correct. Use `internal/pin` for a driver in this repository, because it keeps the type consistent
across drivers.

### Do not configure pins in a driver

`internal/legacy/pinconfig.go:9` states the rule:

> It was observed this way of developing drivers was non-portable and unusable on "big" Go
> projects so future projects should NOT configure pins in driver code. Users must configure pins
> before passing them as arguments to drivers.

The `legacy.ConfigurePinOut` family is deprecated. It exists to keep old drivers compatible. Do
not call it from new code. Document in your `New` function which mode each pin needs.

The two I2C helpers in the same package are different, and you can use them. They are thin
wrappers on `drivers.I2C.Tx` for the register pattern, and 48 files use them:

```go
legacy.ReadRegister(bus, address, reg, buf)  // Tx(addr, []byte{reg}, buf)
legacy.WriteRegister(bus, address, reg, buf) // Tx(addr, append([]byte{reg}, buf...), nil)
```

## 2. Package layout and API shape

One device gets one directory, and the package name matches the directory name. Do not import
another driver package. `drivers.go` explains why. Each driver must stay independent, to keep the
compiled size of the user program small.

The usual API shape:

```go
type Device struct{ ... }

// New returns a new device for the given bus. It does no IO.
func New(bus drivers.SPI, cs pin.Output) *Device

// Configure sets up the device for use.
func (d *Device) Configure(cfg Config) error

// Connected returns whether the device responds on the bus.
func (d *Device) Connected() bool
```

Points to hold to:

- `New` allocates and stores. It must do no IO, because the bus is often not ready yet.
- `Configure` does the IO. Give it a `Config` struct when the device has options, and treat the
  zero value as the sensible default.
- Add `Connected` when the device has an identity register. 32 drivers have it.
- Put the register addresses in `registers.go` as unexported constants. 77 drivers do this.
- Declare errors as package-level variables, so that callers can compare them.
- Implement `drivers.Sensor` when the device measures something, and `drivers.Displayer` when it
  draws. Both are in the root package.

## 3. Build tags, when you cannot avoid machine

Some drivers must have `machine`, for pin timing or for board-specific wiring. 63 packages import
it. Isolate it in its own file with a build tag, and keep the protocol logic and the arithmetic in
files that have no tag. Three tag styles are in use:

| Tag                    | Meaning                    | Example                         |
| ---------------------- | -------------------------- | ------------------------------- |
| `//go:build tinygo`    | Needs the machine package  | `dht/thermometer.go`            |
| `//go:build baremetal` | Needs machine and hardware | `unoqmatrix/matrix_tinygo.go`   |
| `//go:build pico`      | Board or CPU specific      | `waveshare-epd/.../dev_pico.go` |

`dht` is the clearest example. All four of its files that import `machine` carry
`//go:build tinygo`, and it also splits the timing loop between `highfreq.go` and `lowfreq.go` by
CPU. The pure bit decoding stays in an untagged file, which is why `dht` can have a host test.

For the reverse direction, look at `internal/legacy/pinconfig_go.go`. It carries
`//go:build !tinygo` and gives empty stub functions. The host build links, and the calls do
nothing. Use this pattern when a driver must call into hardware setup that has no host meaning.

Two rules follow from this:

- Never put a build tag on a `_test.go` file. No test file in this repository has one, and the
  test runner does not set any tag.
- A package that imports `machine` from an untagged file cannot have a test. The one such package
  with a test file, `waveshare-epd/epd2in66b`, is switched off by hand in `Makefile:22`. Treat
  that as a warning, not as a pattern.

## 4. Tests

`make test` runs three steps: `fmt-check`, `unit-test`, and `smoke-test`.

`unit-test` finds the packages by itself. `Makefile:20` collects every directory that holds a
`*_test.go` file, then subtracts `EXCLUDE_TESTS`. So when you add your first test file, CI starts
to compile your package with the host toolchain. Nothing else opts you in.

### I2C devices: use the tester package

`tester` gives you a mock bus and two register-file devices. Nine packages use it.
`tester.I2CDevice8` keeps registers in an array, and `tester.I2CDevice16` keeps them in a map and
fails the test when the driver touches a register that you did not populate. That map behaviour
is useful, because it catches an accidental read of the wrong address.

`tester.Failer` needs only a `Fatalf` method, so a plain `*testing.T` works:

```go
func TestReadTemperature(t *testing.T) {
	bus := tester.NewI2CBus(t)
	fake := tester.NewI2CDevice16(t, Address)
	fake.Registers = map[uint8]uint16{RegTemp: 0x1234}
	bus.AddDevice(fake)

	dev := New(bus)
	got, err := dev.ReadTemperature()
	if err != nil {
		t.Fatalf("ReadTemperature() error = %v", err)
	}
	if got != want {
		t.Errorf("ReadTemperature() = %d, want %d", got, want)
	}
}
```

Write your own I2C fake when the device does not use a simple register map. `scd30/scd30_test.go`
shows the alternative. It holds a list of expected transactions, checks the write bytes of each
one, and reports any transaction that the driver skipped or added.

### SPI devices: write your own fake

There is no shared SPI helper. Each package writes one, at one of two levels of detail.

An operation log is enough for a display, where you only need to prove the byte sequence.
`sharpmem` and `gdew0154m09` append every byte to a slice and compare it.

A register file is better for a controller with read and write instructions, because the driver
reads back what it wrote. `mcp2515/mcp2515_test.go` decodes the instruction bytes and keeps the
register state:

```go
type fakeSPI struct {
	t     *testing.T
	cs    *fakeOutputPin
	regs  [128]byte
	frame []byte
}

func (s *fakeSPI) Transfer(value byte) (byte, error) {
	s.t.Helper()
	s.frame = append(s.frame, value)
	if s.frame[0] == cmdWrite && len(s.frame) == 3 {
		s.regs[s.frame[1]] = s.frame[2]
	}
	return 0, nil
}
```

The chip select pin gives the fake its frame boundaries. Give your fake output pin a callback, and
reset the frame buffer on the falling edge. Then one fake can serve a whole sequence of
instructions.

Make the fake assert the protocol as well. `gdew0154m09` calls `t.Error` when chip select is high
during a transfer, and when the data pin is wrong for the transfer type. A test double that knows
the protocol finds bugs that no check of a return value can find.

### What to assert

Assert the bytes on the wire, and the values that the driver returns. Do not assert unexported
fields of the device. A test that reads the frame buffer keeps working after you rewrite the
internals, and it documents the datasheet at the same time.

Cover the error paths too. Both `mcp2515` and `gdew0154m09` make the fake fail at the Nth
operation, which proves that the driver returns the error instead of hiding it.

### Test library

13 test files use `github.com/frankban/quicktest`, and 16 use only `testing`. The recent tests
(`mcp2515`, `scd30`, `unoqmatrix`) use the standard library. Prefer the standard library for new
tests. Keep quicktest when you edit a file that already uses it.

## 5. Examples and the smoke test

Every driver needs a program in `examples/<driver>/main.go`, and a line in `smoketest.sh`:

```
tinygo build -size short -o ./build/test.hex -target=<board> ./examples/<driver>/main.go
```

`smoketest.sh` holds 162 of these builds, and `make smoke-test` runs them in parallel. This is the
only compile check that your machine-tagged files get, so name a target board that you used for
the real hardware test. The example also shows the user how to configure the pins, which the
driver no longer does for them.

## 6. Repository conventions

- Branch from `dev`, and send the pull request to `dev`. The `release` branch holds the last
  release.
- Open a GitHub issue before you write a new driver, to prevent two people doing the same work.
- `AGENTS.md` asks for ASD-STE-100 Simplified Technical English in comments, pull request text,
  and issues.
- `AGENTS.md` forbids coding tool attributions anywhere.
- Keep comments short, two lines at most. When a magic value needs an explanation, cite the
  datasheet with the section and page number, or give a URL.
- `gofmt` must be clean. `make fmt-check` fails the build on any unformatted file.
- Do not edit `CHANGELOG.md`. The maintainer writes it at release time.

## 7. Checklist

- [ ] The package imports `machine` only from files that have a build tag.
- [ ] `New` takes `drivers.I2C` or `drivers.SPI`, plus pin HAL types. It does no IO.
- [ ] `Configure` does the IO, and the zero `Config` gives usable defaults.
- [ ] The driver does not configure any pin.
- [ ] Register constants are unexported, and in `registers.go`.
- [ ] The driver implements `drivers.Sensor` or `drivers.Displayer` when that applies.
- [ ] Tests use `tester` for a register-mapped I2C device, or a local fake.
- [ ] Tests cover at least one error path.
- [ ] `examples/<driver>/main.go` exists, and `smoketest.sh` has a line for it.
- [ ] `make test` passes.
- [ ] You tested the driver on real hardware, and the pull request says which board.
