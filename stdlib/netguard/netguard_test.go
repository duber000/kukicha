package netguard

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewAllowValidCIDRs(t *testing.T) {
	g, err := NewAllow([]string{"10.0.0.0/8", "192.168.0.0/16"})
	if err != nil {
		t.Fatalf("NewAllow: %v", err)
	}
	if g.mode != "allow" {
		t.Errorf("mode = %q, want %q", g.mode, "allow")
	}
	if len(g.networks) != 2 {
		t.Errorf("len(networks) = %d, want 2", len(g.networks))
	}
}

func TestNewBlockValidCIDRs(t *testing.T) {
	g, err := NewBlock([]string{"10.0.0.0/8"})
	if err != nil {
		t.Fatalf("NewBlock: %v", err)
	}
	if g.mode != "block" {
		t.Errorf("mode = %q, want %q", g.mode, "block")
	}
}

func TestNewAllowInvalidCIDR(t *testing.T) {
	_, err := NewAllow([]string{"not-a-cidr"})
	if err == nil {
		t.Error("expected error for invalid CIDR, got nil")
	}
}

func TestNewBlockInvalidCIDR(t *testing.T) {
	_, err := NewBlock([]string{"256.0.0.0/8"})
	if err == nil {
		t.Error("expected error for invalid CIDR, got nil")
	}
}

func TestCheckAllowMode(t *testing.T) {
	g, err := NewAllow([]string{"93.184.216.0/24"})
	if err != nil {
		t.Fatalf("NewAllow: %v", err)
	}

	tests := []struct {
		ip   string
		want bool
	}{
		{"93.184.216.34", true},
		{"93.184.216.255", true},
		{"93.184.217.1", false},
		{"10.0.0.1", false},
		{"not-an-ip", false},
	}
	for _, tt := range tests {
		got := Check(g, tt.ip)
		if got != tt.want {
			t.Errorf("Check(allow, %q) = %v, want %v", tt.ip, got, tt.want)
		}
	}
}

func TestCheckBlockMode(t *testing.T) {
	g, err := NewBlock([]string{"10.0.0.0/8", "192.168.0.0/16"})
	if err != nil {
		t.Fatalf("NewBlock: %v", err)
	}

	tests := []struct {
		ip   string
		want bool
	}{
		{"10.0.0.1", false},
		{"192.168.1.1", false},
		{"8.8.8.8", true},
		{"93.184.216.34", true},
	}
	for _, tt := range tests {
		got := Check(g, tt.ip)
		if got != tt.want {
			t.Errorf("Check(block, %q) = %v, want %v", tt.ip, got, tt.want)
		}
	}
}

func TestSSRFGuardBlocksPrivateIPs(t *testing.T) {
	g := NewSSRFGuard()

	blocked := []string{
		"10.0.0.1",
		"10.255.255.255",
		"172.16.0.1",
		"172.31.255.255",
		"192.168.0.1",
		"192.168.255.255",
		"127.0.0.1",
		"127.255.255.255",
		"169.254.1.1",
		"0.0.0.1",
		"100.64.0.1",
		"192.0.0.1",
		"192.0.2.1",
		"198.18.0.1",
		"198.51.100.1",
		"203.0.113.1",
		"224.0.0.1",
		"240.0.0.1",
		"::1",
		"fc00::1",
		"fe80::1",
	}
	for _, ip := range blocked {
		if Check(g, ip) {
			t.Errorf("SSRFGuard should block %q but allowed it", ip)
		}
	}
}

func TestSSRFGuardAllowsPublicIPs(t *testing.T) {
	g := NewSSRFGuard()

	allowed := []string{
		"8.8.8.8",
		"1.1.1.1",
		"93.184.216.34",
		"2606:4700:4700::1111",
	}
	for _, ip := range allowed {
		if !Check(g, ip) {
			t.Errorf("SSRFGuard should allow %q but blocked it", ip)
		}
	}
}

func TestDialContextAllowLocalhost(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "ok")
	}))
	defer ts.Close()

	g, err := NewAllow([]string{"127.0.0.0/8"})
	if err != nil {
		t.Fatalf("NewAllow: %v", err)
	}

	client := HTTPClient(g)
	resp, err := client.Get(ts.URL)
	if err != nil {
		t.Fatalf("HTTPClient.Get: %v", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if string(body) != "ok" {
		t.Errorf("body = %q, want %q", string(body), "ok")
	}
}

func TestDialContextBlockLocalhost(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "should not reach")
	}))
	defer ts.Close()

	g, err := NewBlock([]string{"127.0.0.0/8"})
	if err != nil {
		t.Fatalf("NewBlock: %v", err)
	}

	client := HTTPClient(g)
	_, err = client.Get(ts.URL)
	if err == nil {
		t.Error("expected error when dialing blocked localhost, got nil")
	}
}

func TestHTTPClientIntegration(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "guarded-response")
	}))
	defer ts.Close()

	// Allow localhost so we can reach the test server
	g, err := NewAllow([]string{"127.0.0.0/8"})
	if err != nil {
		t.Fatalf("NewAllow: %v", err)
	}

	client := HTTPClient(g)
	resp, err := client.Get(ts.URL)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("ReadAll: %v", err)
	}
	if string(body) != "guarded-response" {
		t.Errorf("body = %q, want %q", string(body), "guarded-response")
	}
	if resp.StatusCode != 200 {
		t.Errorf("status = %d, want 200", resp.StatusCode)
	}
}

func TestHTTPTransportReturnsTransport(t *testing.T) {
	g := NewSSRFGuard()
	tr := HTTPTransport(g)
	if tr == nil {
		t.Fatal("HTTPTransport returned nil")
	}
}

func TestSSRFGuardBlocksLocalhostServer(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "ssrf-leak")
	}))
	defer ts.Close()

	g := NewSSRFGuard()
	client := HTTPClient(g)
	_, err := client.Get(ts.URL)
	if err == nil {
		t.Error("SSRFGuard should block requests to localhost test server")
	}
}
