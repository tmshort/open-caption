package netinfo

import (
	"sync"
	"time"
)

// portCache remembers which hardware port (Wi-Fi, Ethernet ...) each device has, for the systems where that is a
// command to run (macOS). The list is reloaded every minute, and at once, at most every few seconds, when a device
// is asked about that is not in it: an adapter plugged in after the program started is not in the old list.
type portCache struct {
	mu   sync.Mutex
	at   time.Time
	m    map[string]string
	load func() map[string]string
	now  func() time.Time
}

const (
	portsMaxAge     = time.Minute
	portsMissMinAge = 3 * time.Second
)

func (c *portCache) port(device string) (string, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	now := time.Now
	if c.now != nil {
		now = c.now
	}
	refresh := func() { c.m, c.at = c.load(), now() }
	if c.m == nil || now().Sub(c.at) > portsMaxAge {
		refresh()
	}
	p, ok := c.m[device]
	if !ok && now().Sub(c.at) > portsMissMinAge {
		refresh()
		p, ok = c.m[device]
	}
	return p, ok
}
