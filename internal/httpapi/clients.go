// Copyright the open-caption authors
// SPDX-License-Identifier: Apache-2.0

package httpapi

import (
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/tmshort/open-caption/internal/hub"
	"github.com/tmshort/open-caption/internal/showfile"
)

// ClientRow is one connected audience device, as the Setup and operator pages list it.
type ClientRow struct {
	IP       string `json:"ip"`
	Kind     string `json:"kind"`    // phone | tablet | computer | projector | other
	OS       string `json:"os"`      // iPhone, Android, Windows, ... ("" if unknown)
	Browser  string `json:"browser"` // Safari, Chrome, ... ("" if unknown)
	Lang     string `json:"lang"`
	LangName string `json:"langName"`
	Since    int64  `json:"since"`          // unix ms
	Role     string `json:"role,omitempty"` // "viewer": a second device that only watches the Run page (/follow), not an audience device
}

// describeDevice guesses what a browser is from its User-Agent. It is only a label for a human reading the
// list, so a wrong guess is harmless; nothing depends on it.
func describeDevice(ua string, projector bool) (kind, os, browser string) {
	l := strings.ToLower(ua)
	has := func(s ...string) bool {
		for _, x := range s {
			if strings.Contains(l, x) {
				return true
			}
		}
		return false
	}
	switch {
	case has("iphone"):
		kind, os = "phone", "iPhone"
	case has("ipad"):
		kind, os = "tablet", "iPad"
	case has("android") && has("mobile"):
		kind, os = "phone", "Android"
	case has("android"):
		kind, os = "tablet", "Android"
	case has("cros"):
		kind, os = "computer", "ChromeOS"
	case has("windows"):
		kind, os = "computer", "Windows"
	case has("macintosh", "mac os"):
		kind, os = "computer", "Mac"
	case has("linux", "x11"):
		kind, os = "computer", "Linux"
	default:
		kind = "other"
	}
	switch {
	case has("edg/", "edga/", "edgios/"):
		browser = "Edge"
	case has("opr/", "opera"):
		browser = "Opera"
	case has("firefox/", "fxios/"):
		browser = "Firefox"
	case has("samsungbrowser"):
		browser = "Samsung Internet"
	case has("chrome/", "crios/"):
		browser = "Chrome"
	case has("safari/"):
		browser = "Safari"
	}
	if projector {
		kind = "projector"
	}
	return kind, os, browser
}

// addViewer records a page open only to watch the show, so the devices list can show it; the returned function removes it.
func (rt *Runtime) addViewer(ip, ua string) func() {
	kind, os, br := describeDevice(ua, false)
	rt.ctrlMu.Lock()
	rt.viewerN++
	id := rt.viewerN
	rt.viewers[id] = ClientRow{IP: ip, Kind: kind, OS: os, Browser: br, Role: "viewer", Since: time.Now().UnixMilli()}
	rt.ctrlMu.Unlock()
	return func() {
		rt.ctrlMu.Lock()
		delete(rt.viewers, id)
		rt.ctrlMu.Unlock()
	}
}

// Clients lists the connected audience devices (every language), oldest first within a language, languages in the
// show's order, then the pages that only watch (marked Role "viewer"), oldest first.
func (rt *Runtime) Clients() []ClientRow {
	rt.mu.RLock()
	order := append([]string(nil), rt.order...)
	names := map[string]string{}
	for c, l := range rt.langs {
		names[c] = l.name
	}
	hubs := make(map[string]*hub.Hub, len(rt.hubs))
	for k, h := range rt.hubs {
		hubs[k] = h
	}
	rt.mu.RUnlock()
	codes := append([]string(nil), order...)
	var extra []string
	for c := range hubs {
		if _, ok := names[c]; !ok {
			extra = append(extra, c) // a language the show has since dropped, still connected
		}
	}
	sort.Strings(extra)
	codes = append(codes, extra...)
	out := []ClientRow{}
	for _, code := range codes {
		h := hubs[code]
		if h == nil {
			continue
		}
		name := names[code]
		if name == "" {
			name = showfile.LanguageName(code)
		}
		for _, c := range h.Clients() {
			kind, os, br := describeDevice(c.UserAgent, c.Projector)
			out = append(out, ClientRow{IP: c.IP, Kind: kind, OS: os, Browser: br, Lang: code, LangName: name, Since: c.Since.UnixMilli()})
		}
	}
	rt.ctrlMu.Lock()
	watching := make([]ClientRow, 0, len(rt.viewers))
	for _, v := range rt.viewers {
		watching = append(watching, v)
	}
	rt.ctrlMu.Unlock()
	sort.Slice(watching, func(i, j int) bool { return watching[i].Since < watching[j].Since })
	return append(out, watching...)
}

// listClients is GET .../clients: who is connected to the audience page right now.
func (s *Server) listClients(w http.ResponseWriter, r *http.Request, rt *Runtime) {
	c := rt.Clients()
	audience := 0
	for _, x := range c {
		if x.Role == "" {
			audience++
		}
	}
	writeJSON(w, http.StatusOK, map[string]any{"clients": c, "total": audience})
}
