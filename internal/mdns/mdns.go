// Copyright the open-caption authors
// SPDX-License-Identifier: Apache-2.0

// Package mdns announces this computer on the local network as <name>.local (multicast DNS, RFC 6762), so
// devices can open http://open-caption.local/ instead of a numeric address. It answers address queries for
// that one name, checks first that no other computer already uses it, and says goodbye when it stops. It does
// not need the internet, a DNS server or any router setting, but some guest WiFi networks block multicast.
package mdns

import (
	"errors"
	"fmt"
	"net"
	"regexp"
	"strings"
	"sync"
	"time"

	"golang.org/x/net/dns/dnsmessage"

	"github.com/tmshort/open-caption/internal/logx"
)

// DefaultName is the label used when none is chosen: open-caption.local.
const DefaultName = "open-caption"

const (
	recordTTL   = 120 // seconds an announced address may be cached
	cacheFlush  = 0x8000
	maxAttempts = 9 // open-caption, open-caption-2 ... open-caption-9
)

var labelRe = regexp.MustCompile(`^[a-z0-9]([a-z0-9-]{0,61}[a-z0-9])?$`)

// NormalizeName lowercases a name, drops a trailing ".local" and checks that what is left is one valid DNS
// label: letters, digits and hyphens, not starting or ending with a hyphen, at most 63 characters.
func NormalizeName(s string) (string, error) {
	n := strings.ToLower(strings.TrimSpace(s))
	n = strings.TrimSuffix(strings.TrimSuffix(n, "."), ".local")
	if n == "" {
		return "", errors.New("a name is needed (for example open-caption)")
	}
	if !labelRe.MatchString(n) {
		return "", fmt.Errorf("%q is not a usable name: use letters, digits and hyphens only (no spaces, dots or underscores), starting and ending with a letter or digit, at most 63 characters", s)
	}
	return n, nil
}

// transport is the network, so the protocol can be tested without one.
type transport interface {
	// Recv blocks for the next packet; ifIndex is the interface it arrived on (0 if unknown).
	Recv() (pkt []byte, ifIndex int, src *net.UDPAddr, err error)
	// Multicast sends pkt to the mDNS group on one interface (ifIndex 0: on every interface).
	Multicast(pkt []byte, ifIndex int) error
	// Unicast sends pkt to a single address (for old-style resolvers that ask from another port).
	Unicast(pkt []byte, dst *net.UDPAddr) error
	// Interfaces lists the interfaces to announce on. Refresh re-joins after the network changes.
	Interfaces() []int
	Refresh()
	Close() error
}

// environment is what the responder asks of the computer.
type environment struct {
	// addrs returns the IPv4 addresses to publish for an interface.
	addrs func(ifIndex int) []net.IP
	// isOurs reports whether ip belongs to this computer (to ignore our own looped-back packets).
	isOurs func(ip net.IP) bool
}

type timing struct {
	probes, probeGap, announceGap, watch time.Duration
}

var defaultTiming = timing{probes: 3, probeGap: 250 * time.Millisecond, announceGap: time.Second, watch: 30 * time.Second}

// Responder is a running announcement.
type Responder struct {
	tr  transport
	env environment
	tm  timing

	mu       sync.Mutex
	label    string // the name actually claimed
	wanted   string
	ready    bool
	conflict bool   // set by the receive loop while probing
	known    string // the addresses last announced, to notice a change

	quit chan struct{}
	done sync.WaitGroup
}

// Start claims name.local (or name-2.local ... if another computer has it) and keeps answering for it.
// It returns once the name is claimed, or an error if multicast cannot be used on this computer. allow, when not
// nil, decides which of this computer's addresses are announced (it is asked again each time, so it may change).
func Start(name string, allow func(net.IP) bool) (*Responder, error) {
	n, err := NormalizeName(name)
	if err != nil {
		return nil, err
	}
	tr, err := newUDPTransport(5353, allow)
	if err != nil {
		return nil, err
	}
	return startWith(n, tr, realEnvironment(allow), defaultTiming)
}

func startWith(name string, tr transport, env environment, tm timing) (*Responder, error) {
	r := &Responder{tr: tr, env: env, tm: tm, wanted: name, quit: make(chan struct{})}
	r.done.Add(1)
	go r.receive()
	label, err := r.claim(name)
	if err != nil {
		r.Close()
		return nil, err
	}
	r.mu.Lock()
	r.label, r.ready = label, true
	r.mu.Unlock()
	r.announce(recordTTL)
	r.done.Add(1)
	go r.maintain()
	return r, nil
}

// Name is the label being answered for (it differs from the one asked for if that was taken).
func (r *Responder) Name() string { r.mu.Lock(); defer r.mu.Unlock(); return r.label }

// Host is the full name, for example open-caption.local.
func (r *Responder) Host() string { return r.Name() + ".local" }

// Close says goodbye (so other computers forget the name at once) and stops.
func (r *Responder) Close() {
	select {
	case <-r.quit:
		return
	default:
	}
	if r.Name() != "" {
		r.announce(0)
	}
	close(r.quit)
	r.tr.Close()
	r.done.Wait()
}

// claim probes for name and returns the first free variant of it.
func (r *Responder) claim(name string) (string, error) {
	for i := 1; i <= maxAttempts; i++ {
		cand := name
		if i > 1 {
			cand = fmt.Sprintf("%s-%d", name, i)
		}
		if r.probe(cand) {
			return cand, nil
		}
		logx.Infof("mDNS: the name %s.local is already in use on this network", cand)
	}
	return "", fmt.Errorf("the name %s.local and its variants are all in use on this network", name)
}

// probe asks the network three times whether anyone has name.local; it is free if nobody answers.
func (r *Responder) probe(label string) bool {
	r.mu.Lock()
	r.label, r.conflict, r.ready = label, false, false
	r.mu.Unlock()
	for i := 0; i < int(r.tm.probes); i++ {
		for _, idx := range r.tr.Interfaces() {
			if pkt, err := r.probePacket(label, idx); err == nil {
				r.tr.Multicast(pkt, idx)
			}
		}
		select {
		case <-time.After(r.tm.probeGap):
		case <-r.quit:
			return false
		}
		r.mu.Lock()
		c := r.conflict
		r.mu.Unlock()
		if c {
			r.mu.Lock()
			r.label = ""
			r.mu.Unlock()
			return false
		}
	}
	return true
}

func (r *Responder) probePacket(label string, ifIndex int) ([]byte, error) {
	name := dnsmessage.MustNewName(label + ".local.")
	b := dnsmessage.NewBuilder(nil, dnsmessage.Header{})
	b.EnableCompression()
	if err := b.StartQuestions(); err != nil {
		return nil, err
	}
	b.Question(dnsmessage.Question{Name: name, Type: dnsmessage.TypeALL, Class: dnsmessage.ClassINET | cacheFlush}) // QU: unicast reply wanted
	b.StartAuthorities()
	for _, ip := range r.env.addrs(ifIndex) {
		if a := ip.To4(); a != nil {
			var ar dnsmessage.AResource
			copy(ar.A[:], a)
			b.AResource(dnsmessage.ResourceHeader{Name: name, Class: dnsmessage.ClassINET, TTL: recordTTL}, ar)
		}
	}
	return b.Finish()
}

// announce sends our records unsolicited on every interface (ttl 0 is the goodbye).
func (r *Responder) announce(ttl uint32) {
	for _, idx := range r.tr.Interfaces() {
		if pkt, ok := r.answerPacket(0, idx, dnsmessage.TypeALL, ttl, false); ok {
			r.tr.Multicast(pkt, idx)
		}
	}
	r.mu.Lock()
	r.known = r.addrKey()
	r.mu.Unlock()
}

func (r *Responder) addrKey() string {
	var parts []string
	for _, idx := range r.tr.Interfaces() {
		for _, ip := range r.env.addrs(idx) {
			parts = append(parts, fmt.Sprintf("%d:%s", idx, ip))
		}
	}
	return strings.Join(parts, ",")
}

// answerPacket builds our reply for a question of type qt on one interface. legacy replies (to resolvers that do
// not speak mDNS properly) copy the query id, carry the question and use a short ttl without the cache flush bit.
func (r *Responder) answerPacket(id uint16, ifIndex int, qt dnsmessage.Type, ttl uint32, legacy bool) ([]byte, bool) {
	label := r.Name()
	if label == "" {
		return nil, false
	}
	name := dnsmessage.MustNewName(label + ".local.")
	var v4 []net.IP
	for _, ip := range r.env.addrs(ifIndex) {
		if ip.To4() != nil {
			v4 = append(v4, ip)
		}
	}
	class := dnsmessage.ClassINET | cacheFlush
	if legacy {
		class = dnsmessage.ClassINET
		if ttl > 10 {
			ttl = 10
		}
	}
	b := dnsmessage.NewBuilder(nil, dnsmessage.Header{ID: id, Response: true, Authoritative: true})
	b.EnableCompression()
	if legacy {
		b.StartQuestions()
		b.Question(dnsmessage.Question{Name: name, Type: qt, Class: dnsmessage.ClassINET})
	}
	b.StartAnswers()
	switch {
	case (qt == dnsmessage.TypeA || qt == dnsmessage.TypeALL) && len(v4) > 0:
		for _, ip := range v4 {
			var ar dnsmessage.AResource
			copy(ar.A[:], ip.To4())
			b.AResource(dnsmessage.ResourceHeader{Name: name, Class: class, TTL: ttl}, ar)
		}
	case qt == dnsmessage.TypeAAAA || len(v4) == 0:
		// We have no IPv6 address: say so, so the asker does not wait for one (RFC 6762 section 6.1).
		b.UnknownResource(dnsmessage.ResourceHeader{Name: name, Class: class, TTL: ttl}, dnsmessage.UnknownResource{Type: 47, Data: nsec(name, qt == dnsmessage.TypeAAAA)})
	}
	if qt == dnsmessage.TypeALL && len(v4) > 0 {
		b.StartAdditionals()
		b.UnknownResource(dnsmessage.ResourceHeader{Name: name, Class: class, TTL: ttl}, dnsmessage.UnknownResource{Type: 47, Data: nsec(name, true)})
	}
	pkt, err := b.Finish()
	return pkt, err == nil
}

// nsec is the "this name has no records of the other kind" record: the name itself, then a bitmap of the
// types that do exist (A only, or none when we have no address at all).
func nsec(name dnsmessage.Name, withA bool) []byte {
	var out []byte
	for _, label := range strings.Split(strings.TrimSuffix(name.String(), "."), ".") {
		out = append(out, byte(len(label)))
		out = append(out, label...)
	}
	out = append(out, 0)
	if withA {
		return append(out, 0, 1, 0x40) // window 0, 1 byte, bit for type 1 (A)
	}
	return append(out, 0, 1, 0x00)
}

// receive handles every packet that arrives until Close.
func (r *Responder) receive() {
	defer r.done.Done()
	for {
		pkt, ifIndex, src, err := r.tr.Recv()
		if err != nil {
			select {
			case <-r.quit:
				return
			default:
			}
			if errors.Is(err, net.ErrClosed) {
				return
			}
			time.Sleep(100 * time.Millisecond)
			continue
		}
		r.handle(pkt, ifIndex, src)
	}
}

// handle reacts to one packet: answers questions for our name, and notices another computer claiming it.
func (r *Responder) handle(pkt []byte, ifIndex int, src *net.UDPAddr) {
	if src != nil && r.env.isOurs(src.IP) {
		return // our own packet coming back
	}
	var p dnsmessage.Parser
	h, err := p.Start(pkt)
	if err != nil {
		return
	}
	label := r.Name()
	if label == "" {
		return
	}
	host := label + ".local."
	if h.Response {
		// Someone else answering for our name with addresses that are not ours is a conflict.
		p.SkipAllQuestions()
		for {
			rh, err := p.AnswerHeader()
			if err != nil {
				return
			}
			if rh.Type == dnsmessage.TypeA && strings.EqualFold(rh.Name.String(), host) {
				if a, err := p.AResource(); err == nil && !r.env.isOurs(net.IP(a.A[:])) {
					r.markConflict()
				}
				continue
			}
			if p.SkipAnswer() != nil {
				return
			}
		}
	}
	qs, err := p.AllQuestions()
	if err != nil {
		return
	}
	asked := false
	var qt dnsmessage.Type
	for _, q := range qs {
		if !strings.EqualFold(q.Name.String(), host) {
			continue
		}
		cls := q.Class &^ cacheFlush
		if cls != dnsmessage.ClassINET && cls != dnsmessage.ClassANY {
			continue
		}
		switch q.Type {
		case dnsmessage.TypeA, dnsmessage.TypeAAAA, dnsmessage.TypeALL:
			asked, qt = true, q.Type
		}
	}
	if !asked {
		return
	}
	// A question for our name that carries authority records is another computer probing for it.
	hasAuth := false
	if p.SkipAllAnswers() == nil {
		_, aerr := p.AuthorityHeader()
		hasAuth = aerr == nil
	}
	if hasAuth && !r.isReady() {
		r.markConflict()
		return
	}
	if !r.isReady() {
		return
	}
	legacy := src != nil && src.Port != 5353
	if legacy {
		if pkt, ok := r.answerPacket(h.ID, ifIndex, qt, recordTTL, true); ok {
			r.tr.Unicast(pkt, src)
		}
		return
	}
	if ifIndex == 0 { // the interface is unknown: answer on each, with that interface's own addresses
		for _, idx := range r.tr.Interfaces() {
			if pkt, ok := r.answerPacket(0, idx, qt, recordTTL, false); ok {
				r.tr.Multicast(pkt, idx)
			}
		}
		return
	}
	if pkt, ok := r.answerPacket(0, ifIndex, qt, recordTTL, false); ok {
		r.tr.Multicast(pkt, ifIndex)
	}
}

func (r *Responder) markConflict() { r.mu.Lock(); r.conflict = true; r.mu.Unlock() }
func (r *Responder) isReady() bool { r.mu.Lock(); defer r.mu.Unlock(); return r.ready }

// maintain re-announces when this computer's addresses change (WiFi changed, cable plugged in).
func (r *Responder) maintain() {
	defer r.done.Done()
	t := time.NewTicker(r.tm.watch)
	defer t.Stop()
	for {
		select {
		case <-r.quit:
			return
		case <-t.C:
			r.tr.Refresh()
			r.mu.Lock()
			changed := r.known != r.addrKey()
			r.mu.Unlock()
			if changed {
				logx.Infof("mDNS: this computer's addresses changed; announcing %s again", r.Host())
				r.announce(recordTTL)
			}
		}
	}
}
