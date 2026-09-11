// Command pico2w-ntp keeps the Pico 2 W system clock in step with an NTP server.
//
// WARNING: default -scheduler=cores unsupported, compile with -scheduler=tasks set!
package main

import (
	"errors"
	"log/slog"
	"machine"
	"net/netip"
	"runtime"
	"time"

	"github.com/soypat/cyw43439"
	"github.com/soypat/lneto/ipv4"
	"github.com/soypat/lneto/x/xnet"
)

// Build-time variables, supplied by the Makefile.
// tinygo build -ldflags="-X 'main.wifiPassword=swordfish'"
var (
	wifiSSID     string
	wifiPassword string
)

const (
	hostname  = "ntp-pico"
	ntpHost   = "pool.ntp.org"
	syncEvery = 15 * time.Minute

	// serialWait gives the USB serial monitor time to attach before we log.
	serialWait = 1 * time.Second

	// Both DNS and NTP are single request/response exchanges over UDP, so the
	// same patience suits each of them.
	queryTimeout = 5 * time.Second
	queryRetries = 3
)

func main() {
	logger := slog.New(slog.NewTextHandler(machine.Serial, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	time.Sleep(serialWait)
	logger.Info("starting NTP client")

	// The radio claims PIO program space, a state machine and a DMA channel that
	// it never gives back, so the device is made exactly once per boot.
	dev := cyw43439.NewPicoWDevice()
	stack := connectNetwork(dev, logger)

	// Start NTP goroutine
	go syncClockForever(stack, logger)

	// Don't exit!
	select {}
}

// connectNetwork joins the wifi network and acquires a DHCP lease, retrying
// each step until it succeeds.
//
// The returned stack is ready to use; a goroutine keeps its frames moving. dev
// is reset in place between join attempts rather than remade, because a second
// device would claim PIO and DMA resources the first one still holds.
func connectNetwork(dev *cyw43439.Device, logger *slog.Logger) xnet.StackRetrying {
	var w *wifi
	for attempt := uint(0); ; attempt++ {
		var err error
		w, err = joinWifi(dev, wifiSSID, wifiPassword, hostname, logger)
		if err != nil {
			wait := connectBackoff(attempt)
			logger.Error("wifi join failed",
				slog.String("err", err.Error()),
				slog.Duration("retry", wait),
			)

			// The radio is left mid-handshake, and Init in joinWifi is the way back.
			dev.Reset()
			time.Sleep(wait)
			continue
		}
		break
	}

	// Nothing on the network works until frames are being moved. This runs for
	// the life of the program: the join above is the last step that can fail in
	// a way that would strand it.
	go w.exchangeFramesForever()

	for attempt := uint(0); ; attempt++ {
		dhcp, err := w.setupDHCP()
		if err != nil {
			wait := connectBackoff(attempt)
			logger.Error("DHCP failed",
				slog.String("err", err.Error()),
				slog.Duration("retry", wait),
			)
			time.Sleep(wait)
			continue
		}
		logger.Info("DHCP complete",
			slog.String("addr", ipv4.String(dhcp.AssignedAddr4)),
			slog.Any("dns", dhcp.DNSServers),
		)
		break
	}
	return w.stack.StackRetrying(protocolBackoff)
}

// syncClockForever corrects the system clock from ntpHost every syncEvery.
//
// Failures are logged and left for the next cycle. A clock that is briefly
// stale beats one that stops being corrected at all.
func syncClockForever(stack xnet.StackRetrying, logger *slog.Logger) {
	for {
		server, offset, err := queryNTP(stack)
		if err != nil {
			logger.Error("clock sync failed", slog.String("err", err.Error()))
		} else {
			runtime.AdjustTimeOffset(int64(offset))
			logger.Info("clock synced",
				slog.Duration("offset", offset),
				slog.String("from", server.String()),
			)
		}
		time.Sleep(syncEvery)
	}
}

// queryNTP asks ntpHost how far the system clock has drifted, and who answered.
//
// The host is resolved on every call rather than once at startup: pool.ntp.org
// hands out a different set of servers each time, which is how the pool
// spreads load and routes around servers that have gone away.
func queryNTP(stack xnet.StackRetrying) (netip.Addr, time.Duration, error) {
	addrs, err := stack.DoLookupIP(ntpHost, queryTimeout, queryRetries)
	if err != nil {
		return netip.Addr{}, 0, errors.New("DNS lookup failed: " + err.Error())
	}
	server := addrs[0]
	offset, err := stack.DoNTP(server, queryTimeout, queryRetries)
	if err != nil {
		return netip.Addr{}, 0, errors.New("NTP request failed for " + server.String() + ": " + err.Error())
	}
	return server, offset, nil
}
