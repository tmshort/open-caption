// Copyright the open-caption authors
// SPDX-License-Identifier: Apache-2.0

package httpapi

import (
	"crypto/tls"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// Every cookie the server sets (automatic sign-in on this computer, sign-out) is Secure over HTTPS and not over plain
// http, where a Secure cookie would be dropped by the browser and nobody could stay signed in. (Password sign-in has
// its own check in TestSessionCookieFlags.)
func TestSessionCookiesAreSecureOnlyOverHTTPS(t *testing.T) {
	s := localServer(t, nil)
	do := func(method, path string, secure bool) *http.Cookie {
		t.Helper()
		r := httptest.NewRequest(method, "http://localhost:8080"+path, strings.NewReader(""))
		r.RemoteAddr, r.Host = "127.0.0.1:50000", "localhost:8080"
		if secure {
			r.TLS = &tls.ConnectionState{}
		}
		w := httptest.NewRecorder()
		s.Handler().ServeHTTP(w, r)
		c := w.Result().Cookies()
		if len(c) != 1 {
			t.Fatalf("%s %s (secure=%v): %d cookies", method, path, secure, len(c))
		}
		return c[0]
	}
	for _, secure := range []bool{false, true} {
		auto := do("GET", "/api/admin/ping", secure) // the browser on this computer is signed in without a password
		if auto.Secure != secure || !auto.HttpOnly || auto.SameSite != http.SameSiteStrictMode || auto.MaxAge <= 0 {
			t.Errorf("automatic sign-in cookie (https=%v): %+v", secure, auto)
		}
		out := do("POST", "/api/logout", secure)
		if out.Secure != secure || !out.HttpOnly || out.MaxAge >= 0 || out.Value != "" {
			t.Errorf("sign-out cookie (https=%v): %+v", secure, out)
		}
	}
}
