// Package netinfo finds the host's addresses and manages the self-signed
// HTTPS certificate (docs/06-deployment-and-network.md).
package netinfo

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"net"
	"net/netip"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// Kinds of network adapter, as far as the operating system tells.
const (
	KindEthernet = "ethernet"
	KindWiFi     = "wifi"
	KindOther    = "other" // virtual adapters, VPNs, Bluetooth, phones used as modems
)

// Address is one of this computer's IPv4 addresses and the network adapter it belongs to.
type Address struct {
	IP        net.IP
	Interface string       // the adapter's name, as the operating system calls it
	Kind      string       // KindEthernet, KindWiFi or KindOther
	Network   netip.Prefix // the network it is on, for example 192.168.1.0/24
}

var selected atomic.Pointer[[]string]

// SetInterfaces chooses which network interfaces are used (config network.interfaces): only their addresses are
// offered. An empty list means no choice was made: every interface is used. It can be changed at any time.
func SetInterfaces(names []string) {
	if len(names) == 0 {
		selected.Store(nil)
	} else {
		cp := append([]string(nil), names...)
		selected.Store(&cp)
	}
	offeredCache.reset()
	tableCache.reset()
}

// InterfaceSelected reports whether the interface called name is used.
func InterfaceSelected(name string) bool {
	p := selected.Load()
	if p == nil {
		return true
	}
	for _, n := range *p {
		if strings.EqualFold(strings.TrimSpace(n), name) {
			return true
		}
	}
	return false
}

// Allowed reports whether ip belongs to an interface that is used. An address that no interface is known to have
// (it just appeared, or is gone) is allowed: it cannot be told.
func Allowed(ip net.IP) bool {
	if selected.Load() == nil {
		return true
	}
	if name, ok := tableCache.interfaceOf(ip); ok {
		return InterfaceSelected(name)
	}
	return true
}

// InterfaceOf returns the name of the interface that has ip, if it is one of this computer's addresses.
func InterfaceOf(ip net.IP) (string, bool) { return tableCache.interfaceOf(ip) }

// AllAddresses returns the non-loopback unicast IPv4 addresses of the network interfaces that are up, whether
// or not the interface is used, the main connection first and the others in numeric order. The interfaces
// are asked, not the whole address table: on Windows that table also holds the last address of an adapter that
// is switched off or disconnected (a Wi-Fi card turned off still lists the address it had), and such an
// address must not be offered to the audience or put in the certificate.
func AllAddresses() []Address {
	ifs, _ := net.Interfaces()
	var out []Address
	for _, ifc := range ifs {
		addrs, _ := ifc.Addrs()
		nets := usableNets(ifc.Flags, addrs)
		if len(nets) == 0 {
			continue
		}
		kind := interfaceKind(ifc)
		for _, n := range nets {
			ip := n.IP.To4()
			ones, _ := n.Mask.Size()
			addr, _ := netip.AddrFromSlice(ip)
			out = append(out, Address{IP: ip, Interface: ifc.Name, Kind: kind, Network: netip.PrefixFrom(addr, ones).Masked()})
		}
	}
	ips := make([]net.IP, len(out))
	for i, a := range out {
		ips[i] = a.IP
	}
	ips = mainFirst(ips, mainIP())
	sorted := make([]Address, 0, len(out))
	for _, ip := range ips {
		for _, a := range out {
			if a.IP.Equal(ip) {
				sorted = append(sorted, a)
				break
			}
		}
	}
	return sorted
}

// usedAddresses is AllAddresses less those of interfaces that are not used.
func usedAddresses() []Address {
	var out []Address
	for _, a := range AllAddresses() {
		if InterfaceSelected(a.Interface) {
			out = append(out, a)
		}
	}
	return out
}

// LocalIPs returns the addresses to offer: those of the interfaces that are used, and only one for each network.
// A computer on one network through two adapters (a cable and Wi-Fi at once) has two addresses there, and the
// audience needs just one: the wired one, else the main connection's, else any other, else a Wi-Fi one.
func LocalIPs() []net.IP {
	var out []net.IP
	for _, a := range chooseOnePerNetwork(usedAddresses(), mainIP()) {
		out = append(out, a.IP)
	}
	return out
}

// rank orders the addresses of one network: a wired (Ethernet) interface first, whatever its address, the one with
// the main connection first among those; then the main connection of any other kind, then anything else, then Wi-Fi.
func rank(a Address, main net.IP) int {
	isMain := main != nil && a.IP.Equal(main)
	switch {
	case a.Kind == KindEthernet && isMain:
		return 0
	case a.Kind == KindEthernet:
		return 1
	case isMain:
		return 2
	case a.Kind == KindWiFi:
		return 4
	}
	return 3
}

// chooseOnePerNetwork keeps, for each network, the address with the best rank (the first of equals), in the
// order given.
func chooseOnePerNetwork(addrs []Address, main net.IP) []Address {
	best := map[netip.Prefix]int{}
	for i, a := range addrs {
		if j, ok := best[a.Network]; !ok || rank(a, main) < rank(addrs[j], main) {
			best[a.Network] = i
		}
	}
	var out []Address
	for i, a := range addrs {
		if best[a.Network] == i {
			out = append(out, a)
		}
	}
	return out
}

// Interface is one working network interface and what is done with it, for Setup's list.
type Interface struct {
	Name      string
	Kind      string   // KindEthernet, KindWiFi or KindOther
	Addresses []net.IP // its IPv4 addresses
	Selected  bool     // it is ticked: its addresses may be offered
	Used      bool     // at least one of its addresses is offered
	Instead   string   // when ticked but not used: the interface whose address is offered for the same network
	InsteadIP net.IP   // ... and that address
}

// Interfaces lists the working network interfaces (those that are up and have an address), the one with the main
// connection first, with what is done with each.
func Interfaces() []Interface { return buildInterfaces(AllAddresses(), mainIP()) }

func buildInterfaces(all []Address, main net.IP) []Interface {
	var used []Address
	for _, a := range all {
		if InterfaceSelected(a.Interface) {
			used = append(used, a)
		}
	}
	chosen := map[netip.Prefix]Address{}
	for _, a := range chooseOnePerNetwork(used, main) {
		chosen[a.Network] = a
	}
	var out []Interface
	index := map[string]int{}
	for _, a := range all {
		i, ok := index[a.Interface]
		if !ok {
			i = len(out)
			index[a.Interface] = i
			out = append(out, Interface{Name: a.Interface, Kind: a.Kind, Selected: InterfaceSelected(a.Interface)})
		}
		it := &out[i]
		it.Addresses = append(it.Addresses, a.IP)
		if !it.Selected {
			continue
		}
		if c := chosen[a.Network]; c.IP.Equal(a.IP) {
			it.Used = true
		} else if it.Instead == "" {
			it.Instead, it.InsteadIP = c.Interface, c.IP
		}
	}
	for i := range out {
		if out[i].Used {
			out[i].Instead, out[i].InsteadIP = "", nil // used through another of its addresses
		}
	}
	return out
}

// Offered reports whether ip is one of LocalIPs. The answer is kept for a couple of seconds, so it is cheap to
// ask for every packet (the .local announcement does).
func Offered(ip net.IP) bool { return offeredCache.has(ip) }

var offeredCache cache

type cache struct {
	mu  sync.Mutex
	at  time.Time
	ips []net.IP
}

func (c *cache) reset() { c.mu.Lock(); c.at = time.Time{}; c.mu.Unlock() }

func (c *cache) has(ip net.IP) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	if time.Since(c.at) > 2*time.Second {
		c.ips, c.at = LocalIPs(), time.Now()
	}
	for _, o := range c.ips {
		if o.Equal(ip) {
			return true
		}
	}
	return false
}

// tableCache remembers which interface has which address for a couple of seconds, for Allowed (asked for every
// connection and every packet).
var tableCache addrTable

type addrTable struct {
	mu sync.Mutex
	at time.Time
	by map[string]string
}

func (t *addrTable) reset() { t.mu.Lock(); t.at = time.Time{}; t.mu.Unlock() }

func (t *addrTable) interfaceOf(ip net.IP) (string, bool) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if time.Since(t.at) > 2*time.Second {
		t.by = map[string]string{}
		for _, a := range AllAddresses() {
			t.by[a.IP.String()] = a.Interface
		}
		t.at = time.Now()
	}
	name, ok := t.by[ip.String()]
	return name, ok
}

// mainIP is the address this computer uses to reach the rest of the network (the one on the route to the
// default gateway), or nil when there is none. Connecting a UDP socket only asks the operating system which
// address it would send from: no packet is sent, so it also works on a venue network without internet.
func mainIP() net.IP {
	c, err := net.Dial("udp4", "192.0.2.1:9") // a documentation address (RFC 5737): never reached
	if err != nil {
		return nil
	}
	defer c.Close()
	if a, ok := c.LocalAddr().(*net.UDPAddr); ok {
		return a.IP
	}
	return nil
}

// mainFirst sorts the addresses numerically and puts main (if it is one of them) first, so the address given
// to the audience is the one most likely to work when a computer has a cable and Wi-Fi at once.
func mainFirst(ips []net.IP, main net.IP) []net.IP {
	sort.Slice(ips, func(i, j int) bool { return bytes.Compare(ips[i].To4(), ips[j].To4()) < 0 })
	for i, ip := range ips {
		if main != nil && ip.Equal(main) {
			copy(ips[1:i+1], ips[:i])
			ips[0] = ip
			break
		}
	}
	return ips
}

// usableIPs picks the addresses worth offering from one interface: none unless it is up.
func usableIPs(flags net.Flags, addrs []net.Addr) []net.IP {
	var out []net.IP
	for _, n := range usableNets(flags, addrs) {
		out = append(out, n.IP)
	}
	return out
}

// usableNets is usableIPs with each address's network mask.
func usableNets(flags net.Flags, addrs []net.Addr) []*net.IPNet {
	if flags&net.FlagUp == 0 || flags&net.FlagLoopback != 0 {
		return nil
	}
	var out []*net.IPNet
	for _, a := range addrs {
		if n, ok := a.(*net.IPNet); ok && !n.IP.IsLoopback() && n.IP.To4() != nil && n.IP.IsGlobalUnicast() {
			out = append(out, n)
		}
	}
	return out
}

const (
	certValidity = 365 * 24 * time.Hour
	renewBefore  = 30 * 24 * time.Hour
)

// EnsureSelfSigned returns a usable certificate in dir, creating or renewing it
// when missing, close to expiry, or no longer covering this host's addresses.
func EnsureSelfSigned(dir string, hostname string, ips []net.IP) (tls.Certificate, error) {
	certPath, keyPath := filepath.Join(dir, "cert.pem"), filepath.Join(dir, "key.pem")
	if cert, err := tls.LoadX509KeyPair(certPath, keyPath); err == nil {
		if leaf, err := x509.ParseCertificate(cert.Certificate[0]); err == nil && covers(leaf, hostname, ips) {
			cert.Leaf = leaf
			return cert, nil
		}
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return tls.Certificate{}, err
	}
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return tls.Certificate{}, err
	}
	serial, _ := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 100))
	tmpl := &x509.Certificate{
		SerialNumber: serial,
		Subject:      pkix.Name{CommonName: "open-caption", Organization: []string{"open-caption (self-signed)"}},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(certValidity),
		KeyUsage:     x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		DNSNames:     dnsNames(hostname),
		IPAddresses:  append([]net.IP{net.ParseIP("127.0.0.1"), net.IPv6loopback}, ips...),
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	if err != nil {
		return tls.Certificate{}, err
	}
	kb, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		return tls.Certificate{}, err
	}
	if err := os.WriteFile(certPath, pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der}), 0o644); err != nil {
		return tls.Certificate{}, err
	}
	if err := os.WriteFile(keyPath, pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: kb}), 0o600); err != nil {
		return tls.Certificate{}, err
	}
	return tls.LoadX509KeyPair(certPath, keyPath)
}

func dnsNames(hostname string) []string {
	names := []string{"localhost"}
	if hostname != "" && hostname != "localhost" {
		names = append(names, hostname)
	}
	return names
}

func covers(leaf *x509.Certificate, hostname string, ips []net.IP) bool {
	if time.Until(leaf.NotAfter) < renewBefore {
		return false
	}
	for _, ip := range ips {
		if leaf.VerifyHostname(ip.String()) != nil {
			return false
		}
	}
	for _, n := range dnsNames(hostname) {
		if leaf.VerifyHostname(n) != nil {
			return false
		}
	}
	return true
}
