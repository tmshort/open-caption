// Copyright the open-caption authors
// SPDX-License-Identifier: Apache-2.0

package mdns

import (
	"net"
	"os"
	"testing"
	"time"

	"golang.org/x/net/dns/dnsmessage"
)

// TestRealSocketsAnswerAQuestionFromThisComputer runs the responder on real multicast sockets (on a spare port, so it
// does not disturb the computer's own mDNS service) and asks it from this same computer. The fake network of the other
// tests cannot show a problem with the sockets, the multicast join, or a question that comes from the computer's own
// address, which is what a browser here sends.
func TestRealSocketsAnswerAQuestionFromThisComputer(t *testing.T) {
	var ifc *net.Interface
	var ip net.IP
	ifs, _ := net.Interfaces()
	for i := range ifs {
		if usable(ifs[i], nil) {
			ifc, ip = &ifs[i], ipv4Addrs(ifs[i].Index, nil)[0]
			break
		}
	}
	if ifc == nil {
		t.Skip("no network interface that can do IPv4 multicast")
	}
	spare, err := net.ListenPacket("udp4", "0.0.0.0:0")
	if err != nil {
		t.Skip(err)
	}
	port := spare.LocalAddr().(*net.UDPAddr).Port
	spare.Close()

	tr, err := newUDPTransport(port, nil)
	if err != nil {
		t.Skip("cannot use multicast here: ", err)
	}
	r, err := startWith("oc-realtest", tr, realEnvironment(nil), timing{probes: 2, probeGap: 20 * time.Millisecond, announceGap: 20 * time.Millisecond, watch: time.Hour})
	if err != nil {
		tr.Close()
		t.Fatal(err)
	}
	defer r.Close()

	// An old-style resolver: it asks from an ordinary port, so the answer comes back to it directly.
	c, err := net.ListenUDP("udp4", &net.UDPAddr{IP: ip})
	if err != nil {
		t.Skip(err)
	}
	defer c.Close()
	b := dnsmessage.NewBuilder(nil, dnsmessage.Header{ID: 0x4321})
	b.StartQuestions()
	b.Question(dnsmessage.Question{Name: dnsmessage.MustNewName("oc-realtest.local."), Type: dnsmessage.TypeA, Class: dnsmessage.ClassINET})
	q, _ := b.Finish()

	buf := make([]byte, 1500)
	deadline := time.Now().Add(4 * time.Second)
	for time.Now().Before(deadline) { // ask again: the first question can arrive before the group is joined
		c.SetWriteDeadline(time.Now().Add(time.Second))
		if _, err := c.WriteToUDP(q, &net.UDPAddr{IP: group.IP, Port: port}); err != nil {
			t.Skip("cannot send multicast here: ", err)
		}
		c.SetReadDeadline(time.Now().Add(500 * time.Millisecond))
		n, _, err := c.ReadFromUDP(buf)
		if err != nil {
			continue
		}
		got := parseAny(buf[:n])
		if len(got) == 0 {
			continue // not an answer: the responder's own probe or announcement, which Windows also delivers here
		}
		for _, g := range got {
			if g.Equal(ip) {
				return
			}
		}
		t.Fatalf("answered with %v, want %v", got, ip)
	}
	if os.Getenv("CI") != "" {
		t.Skip("multicast does not loop back on this runner")
	}
	t.Fatalf("no answer from the responder to a question sent from %v on %s", ip, ifc.Name)
}
