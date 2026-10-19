package netinfo

import (
	"crypto/x509"
	"net"
	"strings"
	"testing"
)

func TestSelfSignedCreatedReusedAndRenewedForNewIP(t *testing.T) {
	dir := t.TempDir()
	ip1, ip2 := net.ParseIP("192.168.1.10"), net.ParseIP("10.0.0.7")
	c1, err := EnsureSelfSigned(dir, "stagepc", []net.IP{ip1})
	if err != nil {
		t.Fatal(err)
	}
	leaf, _ := x509.ParseCertificate(c1.Certificate[0])
	for _, h := range []string{"192.168.1.10", "localhost", "stagepc", "127.0.0.1"} {
		if err := leaf.VerifyHostname(h); err != nil {
			t.Errorf("cert should cover %s: %v", h, err)
		}
	}
	c2, _ := EnsureSelfSigned(dir, "stagepc", []net.IP{ip1})
	if string(c2.Certificate[0]) != string(c1.Certificate[0]) {
		t.Error("certificate should be reused when nothing changed")
	}
	c3, _ := EnsureSelfSigned(dir, "stagepc", []net.IP{ip1, ip2})
	if string(c3.Certificate[0]) == string(c1.Certificate[0]) {
		t.Error("certificate should be reissued when a new address appears")
	}
}

func TestLocalIPsAreNotLoopback(t *testing.T) {
	for _, ip := range LocalIPs() {
		if ip.IsLoopback() || ip.To4() == nil {
			t.Errorf("unexpected %v", ip)
		}
	}
}

func TestUsableIPsSkipsInterfacesThatAreDown(t *testing.T) {
	addrs := []net.Addr{
		&net.IPNet{IP: net.ParseIP("192.168.1.138"), Mask: net.CIDRMask(24, 32)},
		&net.IPNet{IP: net.ParseIP("fe80::1"), Mask: net.CIDRMask(64, 128)},    // not IPv4
		&net.IPNet{IP: net.ParseIP("169.254.3.4"), Mask: net.CIDRMask(16, 32)}, // link-local
		&net.IPNet{IP: net.ParseIP("127.0.0.1"), Mask: net.CIDRMask(8, 32)},
	}
	if got := usableIPs(net.FlagUp|net.FlagMulticast, addrs); len(got) != 1 || got[0].String() != "192.168.1.138" {
		t.Errorf("an interface that is up offers only its IPv4 unicast address: %v", got)
	}
	// Windows keeps the last address of a Wi-Fi adapter that was turned off or unplugged.
	if got := usableIPs(net.FlagMulticast, addrs); len(got) != 0 {
		t.Errorf("an interface that is down offers nothing: %v", got)
	}
	if got := usableIPs(net.FlagUp|net.FlagLoopback, addrs); len(got) != 0 {
		t.Errorf("loopback offers nothing: %v", got)
	}
}

func TestMainFirstPutsTheMainConnectionFirstThenNumericOrder(t *testing.T) {
	ips := func(l ...string) []net.IP {
		var out []net.IP
		for _, s := range l {
			out = append(out, net.ParseIP(s))
		}
		return out
	}
	str := func(l []net.IP) string {
		var s []string
		for _, ip := range l {
			s = append(s, ip.String())
		}
		return strings.Join(s, " ")
	}
	if got := str(mainFirst(ips("192.168.1.138", "192.168.1.185", "10.0.0.5"), net.ParseIP("192.168.1.185"))); got != "192.168.1.185 10.0.0.5 192.168.1.138" {
		t.Errorf("main first, then numeric order: %s", got)
	}
	if got := str(mainFirst(ips("192.168.1.100", "192.168.1.9"), nil)); got != "192.168.1.9 192.168.1.100" {
		t.Errorf("without a main connection the order is numeric (not text): %s", got)
	}
	if got := str(mainFirst(ips("192.168.1.9", "192.168.1.100"), net.ParseIP("172.16.0.1"))); got != "192.168.1.9 192.168.1.100" {
		t.Errorf("a main address that is not in the list changes nothing: %s", got)
	}
}
