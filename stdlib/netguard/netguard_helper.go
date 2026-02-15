package netguard

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"syscall"
	"time"
)

// NewAllow creates a Guard that only permits connections to the listed CIDRs.
func NewAllow(cidrs []string) (Guard, error) {
	nets, err := parseCIDRs(cidrs)
	if err != nil {
		return Guard{}, err
	}
	return Guard{networks: nets, mode: "allow"}, nil
}

// NewBlock creates a Guard that blocks connections to the listed CIDRs.
func NewBlock(cidrs []string) (Guard, error) {
	nets, err := parseCIDRs(cidrs)
	if err != nil {
		return Guard{}, err
	}
	return Guard{networks: nets, mode: "block"}, nil
}

// NewSSRFGuard creates a Guard that blocks all private, loopback, link-local,
// CGN, multicast, and reserved IP ranges — the standard SSRF protection set.
func NewSSRFGuard() Guard {
	cidrs := []string{
		"10.0.0.0/8",
		"172.16.0.0/12",
		"192.168.0.0/16",
		"127.0.0.0/8",
		"169.254.0.0/16",
		"::1/128",
		"fc00::/7",
		"fe80::/10",
		"0.0.0.0/8",
		"100.64.0.0/10",
		"192.0.0.0/24",
		"192.0.2.0/24",
		"198.18.0.0/15",
		"198.51.100.0/24",
		"203.0.113.0/24",
		"224.0.0.0/4",
		"240.0.0.0/4",
	}
	nets, _ := parseCIDRs(cidrs) // known-good CIDRs, cannot fail
	return Guard{networks: nets, blockPrivate: true, mode: "block"}
}

// Check validates a single IP string against the guard policy.
// Returns true if the IP is allowed, false if blocked.
func Check(g Guard, ipStr string) bool {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return false // unparseable IPs are never allowed
	}
	return checkIP(g, ip)
}

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

// parseCIDRs parses a list of CIDR strings into []*net.IPNet.
func parseCIDRs(cidrs []string) ([]*net.IPNet, error) {
	nets := make([]*net.IPNet, 0, len(cidrs))
	for _, cidr := range cidrs {
		_, ipNet, err := net.ParseCIDR(cidr)
		if err != nil {
			return nil, fmt.Errorf("netguard: invalid CIDR %q: %w", cidr, err)
		}
		nets = append(nets, ipNet)
	}
	return nets, nil
}

// checkIP tests whether an IP is allowed by the guard policy.
func checkIP(g Guard, ip net.IP) bool {
	matched := false
	for _, n := range g.networks {
		if n.Contains(ip) {
			matched = true
			break
		}
	}

	switch g.mode {
	case "allow":
		return matched // only listed CIDRs are permitted
	case "block":
		return !matched // listed CIDRs are blocked
	default:
		return false
	}
}
