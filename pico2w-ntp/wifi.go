package main

import (
	"errors"
	"log/slog"
	"time"

	"github.com/soypat/cyw43439"
	"github.com/soypat/lneto"
	"github.com/soypat/lneto/ethernet"
	"github.com/soypat/lneto/x/xnet"
)

const (
	// idleSleep is how long the frame exchange waits when it has nothing to send.
	idleSleep = 5 * time.Millisecond

	// stackMTU is the largest ethernet payload the stack will emit. This is a
	// hard ceiling: lneto rejects any larger MTU when the stack is configured.
	stackMTU = ethernet.MaxMTU

	// txFrameSize holds one outgoing frame: payload plus header, VLAN tag and
	// CRC. Received frames use a buffer inside the radio driver, not this one.
	txFrameSize = stackMTU + ethernet.MaxOverheadSize

	// DHCP is several round trips with a server that may be slow to answer.
	dhcpTimeout = 3 * time.Second
	dhcpRetries = 3

	// Resolving the gateway is one exchange with a device on the same subnet.
	arpTimeout = 500 * time.Millisecond
	arpRetries = 4
)

// wifi is a CYW43439 radio joined to a network, with a TCP/IP stack bound to it.
type wifi struct {
	dev    *cyw43439.Device
	stack  *xnet.StackAsync // Read by connectNetwork to run DNS and NTP.
	logger *slog.Logger
}

// joinWifi initialises the CYW43439 radio and joins the named network.
//
// The stack has no address yet; call [wifi.setupDHCP] for that. No frames move
// until [wifi.exchangeFramesForever] is running.
func joinWifi(ssid, password, hostname string, logger *slog.Logger) (*wifi, error) {
	start := time.Now()
	dev := cyw43439.NewPicoWDevice()
	devcfg := cyw43439.DefaultWifiConfig()
	devcfg.Logger = logger
	err := dev.Init(devcfg)
	if err != nil {
		return nil, err
	}
	err = dev.Join(ssid, cyw43439.JoinOptions{Passphrase: password})
	if err != nil {
		return nil, err
	}
	mac, err := dev.HardwareAddr6()
	if err != nil {
		return nil, err
	}

	// Joining takes an unpredictable few hundred milliseconds, so its duration
	// serves as the entropy the stack needs for port and transaction numbers.
	stack := new(xnet.StackAsync)
	err = stack.Reset(xnet.StackConfig{
		HardwareAddress: mac,
		Hostname:        hostname,
		MTU:             stackMTU,
		RandSeed:        time.Since(start).Nanoseconds(),
		Logger:          logger,
	})
	if err != nil {
		return nil, err
	}

	// The driver and the stack meet here: both speak raw ethernet frames.
	//
	// A frame that isn't addressed to us is discarded inside the stack, which
	// reports that as ErrPacketDrop. It is a filtering decision rather than a
	// failure, so return nil and keep it out of the error logs. The AP forwards
	// every broadcast and multicast on the LAN to us, so this is most frames.
	dev.RecvEthHandle(func(pkt []byte) error {
		err := stack.IngressEthernet(pkt)
		if errors.Is(err, lneto.ErrPacketDrop) {
			return nil
		}
		return err
	})
	return &wifi{dev: dev, stack: stack, logger: logger}, nil
}

// exchangeFramesForever moves ethernet frames between the radio and the stack.
//
// Run it in its own goroutine before using the stack. Nothing in lneto has a
// timer of its own, so every protocol only makes progress while this loop runs.
func (w *wifi) exchangeFramesForever() {
	buf := make([]byte, txFrameSize)
	for {
		// PollOne passes any frame it reads to the handler set in joinWifi. It
		// cannot tell us whether it read one: a data frame and an empty queue
		// both report false, so we pace this loop on transmission alone.
		_, err := w.dev.PollOne()
		if err != nil {
			w.logger.Error("poll failed", slog.String("err", err.Error()))
		}
		sent, err := w.stack.EgressEthernet(buf)
		if err != nil {
			w.logger.Error("egress failed", slog.String("err", err.Error()))
		}
		if sent > 0 {
			err = w.dev.SendEth(buf[:sent])
			if err != nil {
				w.logger.Error("send failed", slog.String("err", err.Error()))
			}
		}
		if sent == 0 {
			time.Sleep(idleSleep)
		}
	}
}

// setupDHCP acquires a lease for the stack and resolves the gateway.
//
// The server chooses the address; it is reported in the returned results.
func (w *wifi) setupDHCP() (*xnet.DHCPResults, error) {
	retrying := w.stack.StackRetrying(backoff)
	results, err := retrying.DoDHCPv4([4]byte{}, dhcpTimeout, dhcpRetries)
	if err != nil {
		return nil, err
	}
	err = w.stack.AssimilateDHCPResults(results)
	if err != nil {
		return nil, err
	}

	// Traffic leaving the subnet is addressed to the router's MAC, so we need it.
	// The 6 in the call below is the MAC length; ARP itself is IPv4-only.
	gateway, err := retrying.DoResolveHardwareAddress6(results.Router, arpTimeout, arpRetries)
	if err != nil {
		return nil, err
	}
	w.stack.SetGatewayHardwareAddr(gateway)
	return results, nil
}

// backoff spaces out protocol retries, doubling each time up to 20ms.
//
// Suited to request/response protocols such as DHCP and NTP, not to TCP.
func backoff(consecutive uint) time.Duration {
	const (
		minWait  = 100 * time.Microsecond
		maxWait  = 20 * time.Millisecond
		maxShift = 15
	)
	return min(minWait<<min(consecutive, maxShift), maxWait)
}
