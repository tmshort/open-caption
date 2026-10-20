// Copyright the open-caption authors
// SPDX-License-Identifier: Apache-2.0

package mdns

import (
	"context"
	"errors"
	"fmt"
	"net"
	"sync"

	"golang.org/x/net/ipv4"
)

var group = &net.UDPAddr{IP: net.IPv4(224, 0, 0, 251), Port: 5353}

// udpTransport speaks mDNS over IPv4 multicast.
type udpTransport struct {
	raw   net.PacketConn
	p     *ipv4.PacketConn
	port  int
	allow func(net.IP) bool // which addresses may be announced (nil: all)

	mu     sync.Mutex // guards joined and serialises sends (SetMulticastInterface is per socket)
	joined map[int]bool
}

func newUDPTransport(port int, allow func(net.IP) bool) (*udpTransport, error) {
	lc := net.ListenConfig{Control: shareAddress} // other mDNS programs on this computer use the same port
	raw, err := lc.ListenPacket(context.Background(), "udp4", fmt.Sprintf("0.0.0.0:%d", port))
	if err != nil {
		return nil, fmt.Errorf("cannot listen for multicast DNS on UDP port %d: %w", port, err)
	}
	t := &udpTransport{raw: raw, p: ipv4.NewPacketConn(raw), port: port, allow: allow, joined: map[int]bool{}}
	t.p.SetControlMessage(ipv4.FlagInterface, true) // not available everywhere: then the interface is unknown (0)
	t.p.SetMulticastTTL(255)
	t.p.SetMulticastLoopback(true)
	t.Refresh()
	if len(t.Interfaces()) == 0 {
		raw.Close()
		return nil, errors.New("no network interface with an IPv4 address that can do multicast was found")
	}
	return t, nil
}

func usable(i net.Interface, allow func(net.IP) bool) bool {
	return i.Flags&net.FlagUp != 0 && i.Flags&net.FlagMulticast != 0 && i.Flags&net.FlagLoopback == 0 && len(ipv4Addrs(i.Index, allow)) > 0
}

func (t *udpTransport) Refresh() {
	ifs, _ := net.Interfaces()
	t.mu.Lock()
	defer t.mu.Unlock()
	for _, i := range ifs {
		if usable(i, t.allow) && !t.joined[i.Index] {
			ifc := i
			if err := t.p.JoinGroup(&ifc, &net.UDPAddr{IP: group.IP}); err == nil {
				t.joined[i.Index] = true
			}
		}
	}
	for idx := range t.joined { // forget interfaces that went away
		if ifc, err := net.InterfaceByIndex(idx); err != nil || !usable(*ifc, t.allow) {
			delete(t.joined, idx)
		}
	}
}

func (t *udpTransport) Interfaces() []int {
	t.mu.Lock()
	defer t.mu.Unlock()
	out := make([]int, 0, len(t.joined))
	for idx := range t.joined {
		out = append(out, idx)
	}
	return out
}

func (t *udpTransport) Recv() ([]byte, int, *net.UDPAddr, error) {
	buf := make([]byte, 9000)
	n, cm, src, err := t.p.ReadFrom(buf)
	if err != nil {
		return nil, 0, nil, err
	}
	idx := 0
	if cm != nil {
		idx = cm.IfIndex
	}
	u, _ := src.(*net.UDPAddr)
	return buf[:n], idx, u, nil
}

func (t *udpTransport) Multicast(pkt []byte, ifIndex int) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	var first error
	for idx := range t.joined {
		if ifIndex != 0 && idx != ifIndex {
			continue
		}
		ifc, err := net.InterfaceByIndex(idx)
		if err != nil {
			continue
		}
		if err := t.p.SetMulticastInterface(ifc); err != nil {
			continue
		}
		if _, err := t.p.WriteTo(pkt, nil, &net.UDPAddr{IP: group.IP, Port: t.port}); err != nil && first == nil {
			first = err
		}
	}
	return first
}

func (t *udpTransport) Unicast(pkt []byte, dst *net.UDPAddr) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	_, err := t.raw.WriteTo(pkt, dst)
	return err
}

func (t *udpTransport) Close() error { return t.raw.Close() }

// ipv4Addrs returns an interface's usable IPv4 addresses that allow (when not nil) accepts.
func ipv4Addrs(ifIndex int, allow func(net.IP) bool) []net.IP {
	ifc, err := net.InterfaceByIndex(ifIndex)
	if err != nil {
		return nil
	}
	addrs, _ := ifc.Addrs()
	var out []net.IP
	for _, a := range addrs {
		if n, ok := a.(*net.IPNet); ok {
			if ip := n.IP.To4(); ip != nil && ip.IsGlobalUnicast() && (allow == nil || allow(ip)) {
				out = append(out, ip)
			}
		}
	}
	return out
}

func realEnvironment(allow func(net.IP) bool) environment {
	return environment{
		addrs: func(ifIndex int) []net.IP { return ipv4Addrs(ifIndex, allow) },
		isOurs: func(ip net.IP) bool {
			addrs, _ := net.InterfaceAddrs()
			for _, a := range addrs {
				if n, ok := a.(*net.IPNet); ok && n.IP.Equal(ip) {
					return true
				}
			}
			return false
		},
	}
}
