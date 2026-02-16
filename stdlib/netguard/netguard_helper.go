package netguard

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"syscall"
	"time"
)

// DialContext resolves the address, validates ALL resolved IPs against the
// guard, then dials the first allowed IP directly (preventing DNS rebinding).
// Uses net.Dialer Control as defense-in-depth to re-check at syscall level.
func DialContext(g Guard, ctx context.Context, network, addr string) (net.Conn, error) {
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return nil, fmt.Errorf("netguard: invalid address %q: %w", addr, err)
	}

	// Resolve all IPs for the host
	ips, err := net.DefaultResolver.LookupIPAddr(ctx, host)
	if err != nil {
		return nil, fmt.Errorf("netguard: dns lookup %q: %w", host, err)
	}

	// Find the first allowed IP
	var dialIP net.IPAddr
	found := false
	for _, ip := range ips {
		if checkIP(g, ip.IP) {
			dialIP = ip
			found = true
			break
		}
	}
	if !found {
		return nil, fmt.Errorf("netguard: all resolved IPs for %q are blocked", host)
	}

	// Dial with defense-in-depth Control function
	dialer := net.Dialer{
		Timeout: 30 * time.Second,
		Control: func(network, address string, c syscall.RawConn) error {
			connHost, _, _ := net.SplitHostPort(address)
			connIP := net.ParseIP(connHost)
			if connIP != nil && !checkIP(g, connIP) {
				return fmt.Errorf("netguard: connection to %s blocked by policy", connHost)
			}
			return nil
		},
	}

	dialAddr := net.JoinHostPort(dialIP.IP.String(), port)
	return dialer.DialContext(ctx, network, dialAddr)
}

// HTTPTransport returns an *http.Transport that uses the guarded DialContext.
func HTTPTransport(g Guard) *http.Transport {
	return &http.Transport{
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			return DialContext(g, ctx, network, addr)
		},
	}
}

// HTTPClient returns an *http.Client using the guarded transport.
func HTTPClient(g Guard) *http.Client {
	return &http.Client{
		Transport: HTTPTransport(g),
	}
}
