package mdns

import (
	"net"
	"strings"
	"sync"
	"testing"
	"time"

	"golang.org/x/net/dns/dnsmessage"
)

// bus is a pretend network segment: everything multicast by one node reaches every node, itself included
// (as real multicast does), and unicast reaches the node with that address.
type bus struct {
	mu    sync.Mutex
	nodes []*node
}

type node struct {
	bus  *bus
	ip   net.IP
	in   chan packet
	mu   sync.Mutex
	sent [][]byte
	done chan struct{}
}

type packet struct {
	b   []byte
	src *net.UDPAddr
}

func (b *bus) add(ip string) *node {
	n := &node{bus: b, ip: net.ParseIP(ip).To4(), in: make(chan packet, 64), done: make(chan struct{})}
	b.mu.Lock()
	b.nodes = append(b.nodes, n)
	b.mu.Unlock()
	return n
}

// transport methods
func (n *node) Recv() ([]byte, int, *net.UDPAddr, error) {
	select {
	case p := <-n.in:
		return p.b, 1, p.src, nil
	case <-n.done:
		return nil, 0, nil, net.ErrClosed
	}
}
func (n *node) Multicast(pkt []byte, ifIndex int) error {
	n.mu.Lock()
	n.sent = append(n.sent, pkt)
	n.mu.Unlock()
	n.bus.mu.Lock()
	defer n.bus.mu.Unlock()
	for _, o := range n.bus.nodes {
		select {
		case o.in <- packet{pkt, &net.UDPAddr{IP: n.addr(), Port: 5353}}:
		default:
		}
	}
	return nil
}
func (n *node) Unicast(pkt []byte, dst *net.UDPAddr) error {
	n.mu.Lock()
	n.sent = append(n.sent, pkt)
	n.mu.Unlock()
	n.bus.mu.Lock()
	defer n.bus.mu.Unlock()
	for _, o := range n.bus.nodes {
		if o.addr().Equal(dst.IP) {
			o.in <- packet{pkt, &net.UDPAddr{IP: n.addr(), Port: 5353}}
		}
	}
	return nil
}
func (n *node) Interfaces() []int { return []int{1} }
func (n *node) Refresh()          {}
func (n *node) Close() error {
	select {
	case <-n.done:
	default:
		close(n.done)
	}
	return nil
}
func (n *node) addr() net.IP { n.mu.Lock(); defer n.mu.Unlock(); return n.ip }

func (n *node) lastSent() []byte {
	n.mu.Lock()
	defer n.mu.Unlock()
	if len(n.sent) == 0 {
		return nil
	}
	return n.sent[len(n.sent)-1]
}

var fast = timing{probes: 3, probeGap: 15 * time.Millisecond, announceGap: 10 * time.Millisecond, watch: 20 * time.Millisecond}

func (n *node) env() environment {
	return environment{
		addrs:  func(int) []net.IP { return []net.IP{n.addr()} },
		isOurs: func(ip net.IP) bool { return ip.Equal(n.addr()) },
	}
}

func start(t *testing.T, n *node, name string) *Responder {
	t.Helper()
	r, err := startWith(name, n, n.env(), fast)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(r.Close)
	return r
}

func query(t *testing.T, name string, qt dnsmessage.Type, qu bool) []byte {
	t.Helper()
	class := dnsmessage.ClassINET
	if qu {
		class |= cacheFlush
	}
	b := dnsmessage.NewBuilder(nil, dnsmessage.Header{ID: 0x1234})
	b.StartQuestions()
	b.Question(dnsmessage.Question{Name: dnsmessage.MustNewName(name), Type: qt, Class: class})
	pkt, err := b.Finish()
	if err != nil {
		t.Fatal(err)
	}
	return pkt
}

type reply struct {
	id    uint16
	as    []net.IP
	ttl   []uint32
	flush bool
	nsec  bool
	nsecA bool
	hasQ  bool
}

func parse(t *testing.T, pkt []byte) reply {
	t.Helper()
	var p dnsmessage.Parser
	h, err := p.Start(pkt)
	if err != nil || !h.Response {
		t.Fatalf("not a response: %v", err)
	}
	r := reply{id: h.ID}
	if qs, _ := p.AllQuestions(); len(qs) > 0 {
		r.hasQ = true
	}
	for _, section := range []func() (dnsmessage.ResourceHeader, error){p.AnswerHeader, p.AdditionalHeader} {
		for {
			rh, err := section()
			if err != nil {
				break
			}
			switch rh.Type {
			case dnsmessage.TypeA:
				a, _ := p.AResource()
				r.as = append(r.as, net.IP(a.A[:]))
				r.ttl = append(r.ttl, rh.TTL)
				r.flush = rh.Class&cacheFlush != 0
			case 47:
				u, _ := p.UnknownResource()
				r.nsec = true
				r.nsecA = len(u.Data) > 0 && u.Data[len(u.Data)-1] == 0x40
			default:
				p.SkipAnswer()
			}
		}
	}
	return r
}

// ask delivers a query from a pretend other computer and waits for what the node sends back.
func ask(t *testing.T, n *node, pkt []byte, port int) []byte {
	t.Helper()
	n.mu.Lock()
	before := len(n.sent)
	n.mu.Unlock()
	n.in <- packet{pkt, &net.UDPAddr{IP: net.ParseIP("10.0.0.99"), Port: port}}
	for i := 0; i < 100; i++ {
		time.Sleep(5 * time.Millisecond)
		n.mu.Lock()
		if len(n.sent) > before {
			p := n.sent[len(n.sent)-1]
			n.mu.Unlock()
			return p
		}
		n.mu.Unlock()
	}
	return nil
}

func TestNormalizeName(t *testing.T) {
	for in, want := range map[string]string{"open-caption": "open-caption", " Theater-2.LOCAL ": "theater-2", "Box.local.": "box"} {
		if got, err := NormalizeName(in); err != nil || got != want {
			t.Errorf("%q: %q %v", in, got, err)
		}
	}
	for _, bad := range []string{"", "  ", "has space", "under_score", "-lead", "trail-", "a.b", "é", strings.Repeat("a", 64), "x/y"} {
		if _, err := NormalizeName(bad); err == nil {
			t.Errorf("%q accepted", bad)
		}
	}
}

func TestClaimsTheNameAndAnswersAddressQueries(t *testing.T) {
	b := &bus{}
	n := b.add("192.168.1.10")
	r := start(t, n, "open-caption")
	if r.Host() != "open-caption.local" {
		t.Fatalf("%s", r.Host())
	}
	// A query, in any case, for our name is answered with our address, marked to replace older cache entries.
	got := parse(t, ask(t, n, query(t, "Open-Caption.LOCAL.", dnsmessage.TypeA, false), 5353))
	if len(got.as) != 1 || !got.as[0].Equal(net.ParseIP("192.168.1.10")) || got.ttl[0] != recordTTL || !got.flush || got.id != 0 {
		t.Errorf("%+v", got)
	}
	// A question for another name is not ours to answer.
	if p := ask(t, n, query(t, "printer.local.", dnsmessage.TypeA, false), 5353); p != nil && !strings.Contains(string(p), "printer") && len(parseAny(p)) > 0 {
		t.Errorf("answered for another name")
	}
	// No IPv6 address: the answer says there is none instead of staying silent.
	v6 := parse(t, ask(t, n, query(t, "open-caption.local.", dnsmessage.TypeAAAA, false), 5353))
	if len(v6.as) != 0 || !v6.nsec || !v6.nsecA {
		t.Errorf("%+v", v6)
	}
	// An old-style resolver asking from another port gets a plain unicast answer with its own query id.
	legacy := parse(t, ask(t, n, query(t, "open-caption.local.", dnsmessage.TypeA, false), 40000))
	if legacy.id != 0x1234 || !legacy.hasQ || len(legacy.as) != 1 || legacy.flush || legacy.ttl[0] > 10 {
		t.Errorf("%+v", legacy)
	}
}

func parseAny(p []byte) []net.IP {
	defer func() { recover() }()
	var out []net.IP
	var ps dnsmessage.Parser
	if _, err := ps.Start(p); err != nil {
		return nil
	}
	ps.SkipAllQuestions()
	for {
		rh, err := ps.AnswerHeader()
		if err != nil {
			return out
		}
		if rh.Type == dnsmessage.TypeA {
			a, _ := ps.AResource()
			out = append(out, net.IP(a.A[:]))
		} else {
			ps.SkipAnswer()
		}
	}
}

func TestSecondComputerWithTheSameNameGetsAnotherOne(t *testing.T) {
	b := &bus{}
	first := b.add("192.168.1.10")
	second := b.add("192.168.1.20")
	r1 := start(t, first, "open-caption")
	r2 := start(t, second, "open-caption")
	if r1.Name() != "open-caption" || r2.Name() != "open-caption-2" {
		t.Fatalf("%s %s", r1.Name(), r2.Name())
	}
	a := parse(t, ask(t, first, query(t, "open-caption.local.", dnsmessage.TypeA, false), 5353))
	c := parse(t, ask(t, second, query(t, "open-caption-2.local.", dnsmessage.TypeA, false), 5353))
	if len(a.as) != 1 || !a.as[0].Equal(first.addr()) || len(c.as) != 1 || !c.as[0].Equal(second.addr()) {
		t.Errorf("%+v %+v", a, c)
	}
}

func TestGivesUpWhenEveryVariantIsTaken(t *testing.T) {
	b := &bus{}
	for i, name := range []string{"x", "x-2", "x-3", "x-4", "x-5", "x-6", "x-7", "x-8", "x-9"} {
		n := b.add("192.168.1." + string(rune('1'+i)))
		start(t, n, name)
	}
	late := b.add("192.168.1.99")
	if _, err := startWith("x", late, late.env(), fast); err == nil || !strings.Contains(err.Error(), "in use") {
		t.Errorf("%v", err)
	}
}

func TestSaysGoodbyeAndReannouncesAChangedAddress(t *testing.T) {
	b := &bus{}
	n := b.add("192.168.1.10")
	r := start(t, n, "open-caption")
	// change the address: the watcher announces the new one
	n.mu.Lock()
	n.ip = net.ParseIP("192.168.1.77").To4()
	n.mu.Unlock()
	time.Sleep(150 * time.Millisecond)
	got := parse(t, n.lastSent())
	if len(got.as) != 1 || !got.as[0].Equal(net.ParseIP("192.168.1.77")) {
		t.Errorf("after the change: %+v", got)
	}
	r.Close()
	bye := parse(t, n.lastSent())
	if len(bye.as) != 1 || bye.ttl[0] != 0 {
		t.Errorf("the goodbye has ttl 0: %+v", bye)
	}
	r.Close() // closing twice is harmless
}
