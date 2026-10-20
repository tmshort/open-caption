package netinfo

import (
	"crypto/x509"
	"net"
	"net/netip"
	"strings"
	"testing"
	"time"
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

type fakeConn struct {
	net.Conn
	local  net.Addr
	closed *bool
}

func (c fakeConn) LocalAddr() net.Addr { return c.local }
func (c fakeConn) Close() error        { *c.closed = true; return nil }

type fakeListener struct {
	net.Listener
	conns []net.Conn
}

func (l *fakeListener) Accept() (net.Conn, error) {
	if len(l.conns) == 0 {
		return nil, net.ErrClosed
	}
	c := l.conns[0]
	l.conns = l.conns[1:]
	return c, nil
}

// fakeTable makes Allowed see a computer with a cable (192.168.1.185) and Wi-Fi (192.168.1.138).
func fakeTable() {
	tableCache.mu.Lock()
	tableCache.by = map[string]string{"192.168.1.185": "Ethernet", "192.168.1.138": "Wi-Fi"}
	tableCache.at = time.Now()
	tableCache.mu.Unlock()
}

func TestRestrictedListeningRefusesAddressesOfInterfacesThatAreNotUsed(t *testing.T) {
	t.Cleanup(func() { SetInterfaces(nil); SetRestrictListening(false) })
	tcp := func(s string) *net.TCPAddr { return &net.TCPAddr{IP: net.ParseIP(s), Port: 80} }
	SetInterfaces([]string{"ethernet"}) // the name matches in any case
	fakeTable()

	// Not restricted: the ticks only choose what is offered.
	if !ConnectionAllowed(tcp("192.168.1.138")) {
		t.Error("without the restriction every address is served")
	}
	SetRestrictListening(true)
	for addr, want := range map[string]bool{"192.168.1.185": true, "192.168.1.138": false, "127.0.0.1": true, "::1": true, "10.9.9.9": true /* not an address of any interface: cannot be told */} {
		if got := ConnectionAllowed(tcp(addr)); got != want {
			t.Errorf("%s: %v, want %v", addr, got, want)
		}
	}
	// A connection made to an address of an interface that is not used is closed and never reaches the server;
	// the next one is served.
	var bad, good bool
	var refusedLocal net.Addr
	ln := LimitListener(&fakeListener{conns: []net.Conn{fakeConn{local: tcp("192.168.1.138"), closed: &bad}, fakeConn{local: tcp("192.168.1.185"), closed: &good}}}, func(a net.Addr) { refusedLocal = a })
	c, err := ln.Accept()
	if err != nil || c.LocalAddr().String() != "192.168.1.185:80" || !bad || good {
		t.Fatalf("conn=%v err=%v refusedClosed=%v servedClosed=%v", c, err, bad, good)
	}
	if refusedLocal == nil || refusedLocal.String() != "192.168.1.138:80" {
		t.Errorf("the refusal is reported: %v", refusedLocal)
	}
	// No choice made yet (an empty list): every interface is used.
	SetInterfaces(nil)
	if !ConnectionAllowed(tcp("192.168.1.138")) || !InterfaceSelected("anything") {
		t.Error("an empty list means every interface")
	}
}

func TestInterfaceSelectionIsByNameInAnyCase(t *testing.T) {
	t.Cleanup(func() { SetInterfaces(nil) })
	SetInterfaces([]string{" Ethernet ", "en4"})
	for name, want := range map[string]bool{"ethernet": true, "ETHERNET": true, "en4": true, "Wi-Fi": false, "en0": false} {
		if got := InterfaceSelected(name); got != want {
			t.Errorf("%s: %v, want %v", name, got, want)
		}
	}
}

func TestOnlyOneAddressIsOfferedForEachNetwork(t *testing.T) {
	mk := func(ip, iface, kind, netw string) Address {
		return Address{IP: net.ParseIP(ip), Interface: iface, Kind: kind, Network: netip.MustParsePrefix(netw)}
	}
	wifi := mk("192.168.1.138", "Wi-Fi", KindWiFi, "192.168.1.0/24")
	cable := mk("192.168.1.185", "Ethernet", KindEthernet, "192.168.1.0/24")
	vpn := mk("10.8.0.2", "tun0", KindOther, "10.8.0.0/24")
	str := func(l []Address) string {
		var s []string
		for _, a := range l {
			s = append(s, a.IP.String())
		}
		return strings.Join(s, " ")
	}
	// The wired adapter wins over Wi-Fi even though Wi-Fi sorts first; another network keeps its own address.
	if got := str(chooseOnePerNetwork([]Address{wifi, cable, vpn}, nil)); got != "192.168.1.185 10.8.0.2" {
		t.Errorf("wired over Wi-Fi: %s", got)
	}
	// The main connection (the one that reaches the gateway) wins over everything, also over the cable.
	if got := str(chooseOnePerNetwork([]Address{cable, wifi}, net.ParseIP("192.168.1.138"))); got != "192.168.1.138" {
		t.Errorf("the main connection wins: %s", got)
	}
	// Equals keep the first (numeric order); an "other" adapter beats Wi-Fi but not a cable.
	other := mk("192.168.1.50", "bridge0", KindOther, "192.168.1.0/24")
	if got := str(chooseOnePerNetwork([]Address{wifi, other}, nil)); got != "192.168.1.50" {
		t.Errorf("other beats Wi-Fi: %s", got)
	}
	if got := str(chooseOnePerNetwork([]Address{cable, other}, nil)); got != "192.168.1.185" {
		t.Errorf("a cable beats other: %s", got)
	}
}

func TestAdapterKindFromItsName(t *testing.T) {
	for name, want := range map[string]string{
		"Wi-Fi": KindWiFi, "wlan0": KindWiFi, "wlp3s0": KindWiFi, "Wireless Network Connection": KindWiFi,
		"Ethernet": KindEthernet, "Ethernet 2": KindEthernet, "eth0": KindEthernet, "enp0s31f6": KindEthernet,
		"docker0": KindOther, "tun0": KindOther, "vEthernet (WSL)": KindEthernet, "utun3": KindOther,
	} {
		if got := nameKind(name); got != want {
			t.Errorf("%q: %s, want %s", name, got, want)
		}
	}
}

func TestInterfacesSayWhichOneIsUsedForANetwork(t *testing.T) {
	t.Cleanup(func() { SetInterfaces(nil) })
	mk := func(ip, iface, kind, netw string) Address {
		return Address{IP: net.ParseIP(ip), Interface: iface, Kind: kind, Network: netip.MustParsePrefix(netw)}
	}
	all := []Address{mk("192.168.1.138", "Wi-Fi", KindWiFi, "192.168.1.0/24"), mk("192.168.1.185", "Ethernet", KindEthernet, "192.168.1.0/24"), mk("10.0.0.4", "vEthernet", KindEthernet, "10.0.0.0/24")}
	find := func(l []Interface, name string) Interface {
		for _, i := range l {
			if i.Name == name {
				return i
			}
		}
		t.Fatalf("no %s in %v", name, l)
		return Interface{}
	}
	// Everything ticked (no choice yet): the cable is used for the shared network, Wi-Fi says so; another network is its own.
	l := buildInterfaces(all, nil)
	if w := find(l, "Wi-Fi"); !w.Selected || w.Used || w.Instead != "Ethernet" || w.InsteadIP.String() != "192.168.1.185" {
		t.Errorf("Wi-Fi is ticked but not used: %+v", w)
	}
	if e := find(l, "Ethernet"); !e.Used || e.Instead != "" {
		t.Errorf("%+v", e)
	}
	if v := find(l, "vEthernet"); !v.Used {
		t.Errorf("a different network keeps its own address: %+v", v)
	}
	// Unticking the cable lets Wi-Fi take over; an unticked interface is neither used nor "instead".
	SetInterfaces([]string{"Wi-Fi", "vEthernet"})
	l = buildInterfaces(all, nil)
	if w, e := find(l, "Wi-Fi"), find(l, "Ethernet"); !w.Used || e.Selected || e.Used || e.Instead != "" {
		t.Errorf("wifi=%+v cable=%+v", w, e)
	}
	// The main connection wins over the cable when both are ticked.
	SetInterfaces(nil)
	l = buildInterfaces(all, net.ParseIP("192.168.1.138"))
	if w := find(l, "Wi-Fi"); !w.Used {
		t.Errorf("the main connection is used: %+v", w)
	}
}
