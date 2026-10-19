package netinfo

import (
	"net"
	"sync/atomic"
)

var restrict atomic.Bool

// SetRestrictListening makes LimitListener refuse connections that arrive on an address the filter does not
// offer (config network.restrict_listening). It can be changed at any time.
func SetRestrictListening(on bool) { restrict.Store(on) }

// RestrictingListening reports whether connections to addresses that are not offered are refused.
func RestrictingListening() bool { return restrict.Load() }

// ConnectionAllowed reports whether a connection that arrived on local (the address of this computer it was
// made to) may be served. Everything is allowed unless listening is restricted; this computer itself
// (loopback) always is, so Setup can still be reached to change the choice.
func ConnectionAllowed(local net.Addr) bool {
	if !restrict.Load() {
		return true
	}
	var ip net.IP
	switch a := local.(type) {
	case *net.TCPAddr:
		ip = a.IP
	case *net.IPAddr:
		ip = a.IP
	default:
		return true
	}
	if ip == nil || ip.IsLoopback() {
		return true
	}
	return Allowed(ip)
}

// LimitListener wraps a listener that is bound to every address (":80") so that connections to an address that is
// not offered are closed at once while listening is restricted. The rule is read for each connection, so a change
// applies without rebinding anything. refused, when not nil, is told the local address of each refused connection.
func LimitListener(ln net.Listener, refused func(local net.Addr)) net.Listener {
	return &limited{Listener: ln, refused: refused}
}

type limited struct {
	net.Listener
	refused func(local net.Addr)
}

func (l *limited) Accept() (net.Conn, error) {
	for {
		c, err := l.Listener.Accept()
		if err != nil {
			return nil, err
		}
		if ConnectionAllowed(c.LocalAddr()) {
			return c, nil
		}
		if l.refused != nil {
			l.refused(c.LocalAddr())
		}
		c.Close()
	}
}
