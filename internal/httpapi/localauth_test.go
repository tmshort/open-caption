// Copyright the open-caption authors
// SPDX-License-Identifier: Apache-2.0

package httpapi

import (
	"encoding/json"
	"github.com/tmshort/open-caption/internal/config"
	"net/http"
	"net/http/httptest"
	"testing"
)

func localServer(t *testing.T, mutate func(*Server)) *Server {
	t.Helper()
	cfg := testConfig()
	cfg.Access.LocalAutoLogin = true
	s := New(cfg, "test")
	t.Cleanup(s.Close)
	if mutate != nil {
		mutate(s)
	}
	return s
}

func probe(s *Server, path, remote, host string, hdr map[string]string) *httptest.ResponseRecorder {
	r := httptest.NewRequest("GET", "http://"+host+path, nil)
	r.RemoteAddr = remote
	r.Host = host
	for k, v := range hdr {
		r.Header.Set(k, v)
	}
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, r)
	return w
}

// On the computer that runs the program, Setup and the operator page need no password; nothing else does.
func TestBrowserOnThisComputerIsSignedInAsAdmin(t *testing.T) {
	s := localServer(t, nil)
	for _, host := range []string{"localhost:8080", "127.0.0.1", "127.0.0.1:8080", "[::1]:8080", "LOCALHOST"} {
		for _, remote := range []string{"127.0.0.1:50000", "[::1]:50000"} {
			w := probe(s, "/api/admin/ping", remote, host, nil)
			var m map[string]any
			json.Unmarshal(w.Body.Bytes(), &m)
			if w.Code != 200 || m["role"] != "admin" || m["auto"] != true || len(w.Result().Cookies()) != 1 {
				t.Errorf("%s from %s: %d %s", host, remote, w.Code, w.Body.String())
			}
		}
	}
	// the cookie it hands out keeps working (the operator page uses the same session) and says there is nothing to sign out of
	w := probe(s, "/api/admin/ping", "127.0.0.1:1", "localhost", nil)
	ck := w.Result().Cookies()[0]
	r := httptest.NewRequest("GET", "http://localhost/api/operator/state", nil)
	r.RemoteAddr, r.Host = "127.0.0.1:1", "localhost"
	r.AddCookie(ck)
	w2 := httptest.NewRecorder()
	s.Handler().ServeHTTP(w2, r)
	var st map[string]any
	json.Unmarshal(w2.Body.Bytes(), &st)
	if w2.Code != 200 || st["role"] != "admin" || st["auto"] != true {
		t.Errorf("operator state: %d %s", w2.Code, w2.Body.String())
	}
	if !ck.HttpOnly || ck.SameSite != http.SameSiteStrictMode {
		t.Errorf("cookie flags: %+v", ck)
	}
}

func TestAutomaticSignInIsRefusedWhenItCouldBeSomeoneElse(t *testing.T) {
	s := localServer(t, nil)
	refused := map[string]struct {
		remote, host string
		hdr          map[string]string
	}{
		"another computer on the network":        {"192.168.1.20:5000", "localhost", nil},
		"another computer using our address":     {"192.168.1.20:5000", "192.168.1.5", nil},
		"a web page rebinding its own name":      {"127.0.0.1:5000", "evil.example:8080", nil},
		"the computer's own name, not localhost": {"127.0.0.1:5000", "box-office.local", nil},
		"relayed by a proxy":                     {"127.0.0.1:5000", "localhost", map[string]string{"X-Forwarded-For": "203.0.113.9"}},
		"relayed (Forwarded)":                    {"127.0.0.1:5000", "localhost", map[string]string{"Forwarded": "for=203.0.113.9"}},
		"started by another website":             {"127.0.0.1:5000", "localhost", map[string]string{"Sec-Fetch-Site": "cross-site"}},
		"started by another port":                {"127.0.0.1:5000", "localhost", map[string]string{"Sec-Fetch-Site": "same-site"}},
	}
	for name, c := range refused {
		if w := probe(s, "/api/admin/ping", c.remote, c.host, c.hdr); w.Code != 401 {
			t.Errorf("%s: %d", name, w.Code)
		}
	}
	// a typed address (Sec-Fetch-Site none) and the program's own pages (same-origin) are fine
	for _, v := range []string{"none", "same-origin"} {
		if w := probe(s, "/api/admin/ping", "127.0.0.1:5000", "localhost", map[string]string{"Sec-Fetch-Site": v}); w.Code != 200 {
			t.Errorf("Sec-Fetch-Site %s: %d", v, w.Code)
		}
	}
}

func TestAutomaticSignInCanBeTurnedOffAndRespectsTheAccessList(t *testing.T) {
	off := localServer(t, func(s *Server) { s.cfg.Access.LocalAutoLogin = false })
	if w := probe(off, "/api/admin/ping", "127.0.0.1:1", "localhost", nil); w.Code != 401 {
		t.Errorf("turned off: %d", w.Code)
	}
	// the access list still decides who may even try: a list that leaves out 127.0.0.1 keeps it out
	restricted := localServer(t, func(s *Server) { s.adminAllow, _ = config.ParseAllowList([]string{"192.168.1.0/24"}) })
	if w := probe(restricted, "/api/admin/ping", "127.0.0.1:1", "localhost", nil); w.Code != 403 {
		t.Errorf("not on the access list: %d", w.Code)
	}
}

// A script on this computer that keeps no cookies still works on every call, without piling up sessions.
func TestCookielessLocalCallsAlwaysWorkAndDoNotPileUpSessions(t *testing.T) {
	s := localServer(t, nil)
	for i := 0; i < 3*maxAutoSessions; i++ {
		if w := probe(s, "/api/admin/ping", "127.0.0.1:50000", "localhost", nil); w.Code != 200 {
			t.Fatalf("call %d: %d", i, w.Code)
		}
	}
	s.auth.mu.Lock()
	n := len(s.auth.sessions)
	s.auth.mu.Unlock()
	if n > maxAutoSessions {
		t.Errorf("%d sessions kept, want at most %d", n, maxAutoSessions)
	}
}

// With the built-in defaults, Setup opens only on the computer that runs the program (access.admin_allow: [localhost]).
func TestSetupIsLocalOnlyByDefault(t *testing.T) {
	cfg := config.Default()
	cfg.AdminPassword, cfg.OperatorPassword = "adminpw", "operatorpw"
	s := New(cfg, "test")
	t.Cleanup(s.Close)
	if w := probe(s, "/api/admin/ping", "192.168.1.20:5000", "192.168.1.5", nil); w.Code != 403 {
		t.Errorf("another computer: %d", w.Code)
	}
	if w := probe(s, "/api/admin/ping", "127.0.0.1:5000", "localhost", nil); w.Code != 200 {
		t.Errorf("this computer: %d", w.Code)
	}
}
