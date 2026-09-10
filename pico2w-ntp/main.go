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
	syncEvery = 60 * time.Minute

	// serialWait gives the USB serial monitor time to attach before we log.
	serialWait = 2 * time.Second

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

	stack, err := connectNetwork(logger)
	if err != nil {
		panic("network setup failed:" + err.Error())
	}
	syncClockForever(stack, logger)
}

// connectNetwork joins the wifi network and acquires a DHCP lease.
//
// The returned stack is ready to use; a goroutine keeps its frames moving.
func connectNetwork(logger *slog.Logger) (xnet.StackRetrying, error) {
	w, err := joinWifi(wifiSSID, wifiPassword, hostname, logger)
	if err != nil {
		return xnet.StackRetrying{}, err
	}

	// Nothing on the network works until frames are being moved.
	go w.exchangeFramesForever()

	dhcp, err := w.setupDHCP()
	if err != nil {
		return xnet.StackRetrying{}, err
	}
	logger.Info("DHCP complete",
		slog.String("addr", ipv4.String(dhcp.AssignedAddr4)),
		slog.Any("dns", dhcp.DNSServers),
	)
	return w.stack.StackRetrying(backoff), nil
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
func queryNTP(stack xnet.StackRetrying) (server netip.Addr, offset time.Duration, err error) {
	addrs, err := stack.DoLookupIP(ntpHost, queryTimeout, queryRetries)
	if err != nil {
		return server, 0, errors.New("DNS lookup failed: " + err.Error())
	}
	server = addrs[0]
	offset, err = stack.DoNTP(server, queryTimeout, queryRetries)
	if err != nil {
		return server, 0, errors.New("NTP request failed for " + server.String() + ": " + err.Error())
	}
	return server, offset, nil
}
