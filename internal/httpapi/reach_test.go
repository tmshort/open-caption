// Copyright the open-caption authors
// SPDX-License-Identifier: Apache-2.0

package httpapi

import (
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
)

// The connection test says which addresses answer and why the others do not.
func TestProbeURLsSaysWhichAddressesAnswer(t *testing.T) {
	up := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { fmt.Fprintln(w, "ok") }))
	defer up.Close()
	other := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { http.NotFound(w, r) }))
	defer other.Close()
	tlsUp := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { fmt.Fprintln(w, "ok") }))
	defer tlsUp.Close()
	ln, _ := net.Listen("tcp", "127.0.0.1:0")
	closed := "http://" + ln.Addr().String() + "/"
	ln.Close()

	got := probeURLs([]string{up.URL + "/", tlsUp.URL + "/", other.URL + "/", closed})
	if !got[0].OK || got[0].Error != "" {
		t.Errorf("a running server should answer: %+v", got[0])
	}
	if !got[1].OK {
		t.Errorf("a self-signed https server should answer (the certificate is not checked): %+v", got[1])
	}
	if got[2].OK || got[2].Error == "" {
		t.Errorf("a server that is not this program should not count: %+v", got[2])
	}
	if got[3].OK || got[3].Error != "refused" {
		t.Errorf("a closed port is refused: %+v", got[3])
	}
}

func TestReachTestWithNothingListeningHasNoResults(t *testing.T) {
	_, _, c := serverWith(t, nil)
	code, r := c.do("POST", "/api/admin/troubleshoot/reach", nil)
	if code != 200 {
		t.Fatalf("%d %v", code, r)
	}
	if rs, _ := r["results"].([]any); len(rs) != 0 {
		t.Errorf("no listeners are known to the test server, so there is nothing to ask: %v", r)
	}
}
