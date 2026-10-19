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

var filter atomic.Pointer[func(net.IP) bool]

// SetFilter limits LocalIPs to the addresses allow accepts (config network.include and network.exclude); nil
// offers every address. It can be changed at any time.
func SetFilter(allow func(net.IP) bool) {
	if allow == nil {
		filter.Store(nil)
	} else {
		filter.Store(&allow)
	}
	offeredCache.reset()
}

var prefer atomic.Pointer[[]string]

// SetPrefer sets which adapter is used when two are on the same network (config network.prefer): the first entry
// that matches an adapter's name, or its kind ("ethernet", "wifi"), wins over every other rule.
func SetPrefer(entries []string) {
	if len(entries) == 0 {
		prefer.Store(nil)
	} else {
		cp := append([]string(nil), entries...)
		prefer.Store(&cp)
	}
	offeredCache.reset()
}

// preferRank is the position of the first prefer entry that matches the address's adapter, or -1.
func preferRank(a Address) int {
	p := prefer.Load()
	if p == nil {
		return -1
	}
	for i, e := range *p {
		e = strings.ToLower(strings.TrimSpace(e))
		switch {
		case strings.EqualFold(e, a.Interface):
			return i
		case (e == KindEthernet || e == "wired") && a.Kind == KindEthernet, (e == KindWiFi || e == "wi-fi" || e == "wireless") && a.Kind == KindWiFi:
			return i
		}
	}
	return -1
}

// Allowed reports whether the filter (SetFilter) offers ip; with no filter, every address is offered.
func Allowed(ip net.IP) bool {
	allow := filter.Load()
	return allow == nil || (*allow)(ip)
}

// AllAddresses returns the non-loopback unicast IPv4 addresses of the network interfaces that are up, whether
// or not the filter would offer them, the main connection first and the others in numeric order. The interfaces
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

// LocalIPs returns the addresses to offer: those the filter allows, and only one for each network. A computer
// on one network through two adapters (a cable and Wi-Fi at once) has two addresses there, and the audience
// needs just one: the main connection's, else the wired one, else a Wi-Fi one.
func LocalIPs() []net.IP {
	var allowed []Address
	for _, a := range AllAddresses() {
		if Allowed(a.IP) {
			allowed = append(allowed, a)
		}
	}
	var out []net.IP
	for _, a := range chooseOnePerNetwork(allowed, mainIP()) {
		out = append(out, a.IP)
	}
	return out
}

// rank orders the addresses of one network: an adapter chosen with network.prefer, then the main connection, then
// wired, then anything else, then Wi-Fi.
func rank(a Address, main net.IP) int {
	if r := preferRank(a); r >= 0 {
		return r - 1000 // a chosen adapter beats everything below
	}
	switch {
	case main != nil && a.IP.Equal(main):
		return 0
	case a.Kind == KindEthernet:
		return 1
	case a.Kind == KindWiFi:
		return 3
	}
	return 2
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

// Status says what the rules do with one address.
type Status struct {
	Address
	Allowed bool   // network.include / network.exclude allow it
	Offered bool   // it is the one given to the audience for its network
	Instead net.IP // when allowed but not offered: the address offered for the same network
}

// Statuses lists every address with what the rules do with it, in the order AllAddresses gives them.
func Statuses() []Status {
	all := AllAddresses()
	var allowed []Address
	for _, a := range all {
		if Allowed(a.IP) {
			allowed = append(allowed, a)
		}
	}
	chosen := map[netip.Prefix]net.IP{}
	for _, a := range chooseOnePerNetwork(allowed, mainIP()) {
		chosen[a.Network] = a.IP
	}
	out := make([]Status, 0, len(all))
	for _, a := range all {
		st := Status{Address: a, Allowed: Allowed(a.IP)}
		if st.Allowed {
			if c := chosen[a.Network]; c.Equal(a.IP) {
				st.Offered = true
			} else {
				st.Instead = c
			}
		}
		out = append(out, st)
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
