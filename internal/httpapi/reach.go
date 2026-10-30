// Copyright the open-caption authors
// SPDX-License-Identifier: Apache-2.0

package httpapi

import (
	"crypto/tls"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/tmshort/open-caption/internal/netinfo"
)

// reachResult is the answer for one address in Troubleshoot's connection test.
type reachResult struct {
	URL   string `json:"url"`
	OK    bool   `json:"ok"`
	Ms    int    `json:"ms,omitempty"`
	Error string `json:"error,omitempty"`
}

// reachTimeout is how long one address may take to answer; a firewall that drops packets shows up as this.
const reachTimeout = 3 * time.Second

// probeURLs asks each address for /healthz at the same time and says which answered. The certificate is not checked
// (a self-signed one is the usual case): this is about whether the server can be reached, not whether it is trusted.
func probeURLs(urls []string) []reachResult {
	client := &http.Client{
		Timeout:       reachTimeout,
		Transport:     &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, DisableKeepAlives: true}, //nolint:gosec // reachability only
		CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse },
	}
	out := make([]reachResult, len(urls))
	var wg sync.WaitGroup
	for i, u := range urls {
		wg.Add(1)
		go func() {
			defer wg.Done()
			out[i] = reachResult{URL: u}
			start := time.Now()
			resp, err := client.Get(u + "healthz")
			if err != nil {
				out[i].Error = simplifyErr(err)
				return
			}
			defer resp.Body.Close()
			b, _ := io.ReadAll(io.LimitReader(resp.Body, 16))
			out[i].Ms = int(time.Since(start).Milliseconds())
			out[i].OK = resp.StatusCode == http.StatusOK && strings.TrimSpace(string(b)) == "ok"
			if !out[i].OK {
				out[i].Error = resp.Status
			}
		}()
	}
	wg.Wait()
	return out
}

// simplifyErr keeps the reason a connection failed ("connection refused", "timeout") without the request details.
func simplifyErr(err error) string {
	s := err.Error()
	switch {
	case strings.Contains(s, "deadline exceeded") || strings.Contains(s, "Timeout") || strings.Contains(s, "timeout"):
		return "timeout"
	case strings.Contains(s, "refused"):
		return "refused"
	}
	if i := strings.LastIndex(s, ": "); i >= 0 {
		return s[i+2:]
	}
	return s
}

// reachTest is POST /api/admin/troubleshoot/reach: this computer asks for the program's own web server on each network
// address it gives out (HTTP, and HTTPS while it is on) and reports which answered and how fast. It proves the server is
// listening on that address, not that another device can get through the venue's network or a firewall.
func (s *Server) reachTest(w http.ResponseWriter, r *http.Request) {
	s.lmu.Lock()
	l := s.listeners
	s.lmu.Unlock()
	var urls []string
	for _, ip := range netinfo.LocalIPs() {
		if l.HTTPAddr != "" {
			urls = append(urls, hostURL("http", ip.String(), l.HTTPAddr, "80"))
		}
		if s.cfg.HTTPS.Enabled && l.HTTPSAddr != "" {
			urls = append(urls, hostURL("https", ip.String(), l.HTTPSAddr, "443"))
		}
	}
	writeJSON(w, http.StatusOK, map[string]any{"results": probeURLs(urls)})
}
