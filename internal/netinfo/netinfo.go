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
	"os"
	"path/filepath"
	"sort"
	"sync/atomic"
	"time"
)

// Address is one of this computer's IPv4 addresses and the network interface it belongs to.
type Address struct {
	IP        net.IP
	Interface string
}

var filter atomic.Pointer[func(net.IP) bool]

// SetFilter limits LocalIPs to the addresses allow accepts (config network.include and network.exclude); nil
// offers every address. It can be changed at any time.
func SetFilter(allow func(net.IP) bool) {
	if allow == nil {
		filter.Store(nil)
		return
	}
	filter.Store(&allow)
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
		for _, ip := range usableIPs(ifc.Flags, addrs) {
			out = append(out, Address{IP: ip, Interface: ifc.Name})
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

// LocalIPs returns the addresses to offer: AllAddresses less what the filter leaves out.
func LocalIPs() []net.IP {
	var out []net.IP
	for _, a := range AllAddresses() {
		if Allowed(a.IP) {
			out = append(out, a.IP)
		}
	}
	return out
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
	if flags&net.FlagUp == 0 || flags&net.FlagLoopback != 0 {
		return nil
	}
	var out []net.IP
	for _, a := range addrs {
		if n, ok := a.(*net.IPNet); ok && !n.IP.IsLoopback() && n.IP.To4() != nil && n.IP.IsGlobalUnicast() {
			out = append(out, n.IP)
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
