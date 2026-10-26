// Copyright the open-caption authors
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"net"
	"reflect"
	"strings"
	"testing"

	"github.com/tmshort/open-caption/internal/config"
)

func TestListenFallsBackWhenThePortIsTaken(t *testing.T) {
	taken, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Skip(err)
	}
	defer taken.Close()
	ln, err := listen(config.Listener{Addr: taken.Addr().String(), FallbackAddr: "127.0.0.1:0"})
	if err != nil {
		t.Fatalf("no fallback used: %v", err)
	}
	defer ln.Close()
	if ln.Addr().String() == taken.Addr().String() {
		t.Error("listened on the taken address")
	}
	// No fallback (or the same one): the error says what to try.
	for _, fb := range []string{"", taken.Addr().String()} {
		if _, err := listen(config.Listener{Addr: taken.Addr().String(), FallbackAddr: fb}); err == nil || !strings.Contains(err.Error(), "already using that port") {
			t.Errorf("fallback %q: %v", fb, err)
		}
	}
}

func TestURLsLeaveOutTheDefaultPort(t *testing.T) {
	for _, c := range []struct{ scheme, addr, want string }{
		{"http", "[::]:80", "http://localhost/"},
		{"http", "[::]:8080", "http://localhost:8080/"},
		{"https", "0.0.0.0:443", "https://localhost/"},
		{"https", "0.0.0.0:8443", "https://localhost:8443/"},
		{"http", "[::]:443", "http://localhost:443/"}, // 443 is not the default for http
	} {
		addr, _ := net.ResolveTCPAddr("tcp", strings.Replace(c.addr, "[::]", "0.0.0.0", 1))
		got := urls(c.scheme, addr)
		if len(got) == 0 || got[0] != c.want {
			t.Errorf("%s %s: %v, want first %s", c.scheme, c.addr, got, c.want)
		}
		for _, u := range got {
			if !strings.HasPrefix(u, c.scheme+"://") || !strings.HasSuffix(u, "/") {
				t.Errorf("%s: odd URL %q", c.addr, u)
			}
		}
	}
	if got := firstURL("http", "[::]:80"); got != "http://localhost/" {
		t.Errorf("%q", got)
	}
	if got := firstURL("https", "127.0.0.1:8443"); got != "https://localhost:8443/" {
		t.Errorf("%q", got)
	}
	if got := firstURL("http", "nonsense"); got != "http://localhost/" {
		t.Errorf("%q", got)
	}
}

func TestSmallHelpers(t *testing.T) {
	if got := splitList(" a, b ,,c ,"); !reflect.DeepEqual(got, []string{"a", "b", "c"}) {
		t.Errorf("%q", got)
	}
	if got := splitList("  "); len(got) != 0 {
		t.Errorf("%q", got)
	}
	for in, want := range map[string]bool{"yes": true, "YES": true, " true ": true, "on": true, "1": true, "no": false, "": false, "0": false, "off": false} {
		t.Setenv("OC_TEST_YES", in)
		if envYes("OC_TEST_YES") != want {
			t.Errorf("envYes(%q) = %v", in, !want)
		}
	}
	t.Setenv("OC_TEST_VAR", "")
	if env("OC_TEST_VAR", "fallback") != "fallback" {
		t.Error("an empty variable should give the default")
	}
	t.Setenv("OC_TEST_VAR", "set")
	if env("OC_TEST_VAR", "fallback") != "set" {
		t.Error("a set variable should win")
	}
	for in, want := range map[string]string{":8080": "127.0.0.1:8080", "0.0.0.0:9000": "127.0.0.1:9000", "garbage": "127.0.0.1:8080", "": "127.0.0.1:8080"} {
		if got := loopbackOf(in); got != want {
			t.Errorf("loopbackOf(%q) = %q, want %q", in, got, want)
		}
	}
}
