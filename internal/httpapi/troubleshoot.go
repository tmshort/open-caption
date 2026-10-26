// Copyright the open-caption authors
// SPDX-License-Identifier: Apache-2.0

package httpapi

import (
	"net/http"
	"os"
	"strings"

	"github.com/tmshort/open-caption/internal/logx"
	"github.com/tmshort/open-caption/internal/netinfo"
)

// A finding is one thing Setup's Troubleshoot panel reports: what is wrong or fine, as an id and the numbers or names
// that go with it. The words are in Setup (so they are translated), and Setup knows where each one is fixed.
type finding struct {
	ID     string         `json:"id"`
	Status string         `json:"status"` // ok | info | warn | bad
	Params map[string]any `json:"params,omitempty"`
}

// troubleshoot is GET /api/admin/troubleshoot: what the program knows about why people might not be able to reach
// it, in the order a person would look. ?brief=1 returns only the count of problems logged (for the button's badge).
func (s *Server) troubleshoot(w http.ResponseWriter, r *http.Request) {
	problems := 0
	if p, ok := s.logs.(interface{ Problems() int }); ok {
		problems = p.Problems()
	}
	if r.URL.Query().Get("brief") != "" {
		writeJSON(w, http.StatusOK, map[string]any{"problems": problems})
		return
	}
	var cs []finding
	add := func(id, status string, kv ...any) {
		c := finding{ID: id, Status: status}
		if len(kv) > 0 {
			c.Params = map[string]any{}
			for i := 0; i+1 < len(kv); i += 2 {
				c.Params[kv[i].(string)] = kv[i+1]
			}
		}
		cs = append(cs, c)
	}

	// 1. an address to give people
	if ips := netinfo.LocalIPs(); len(ips) == 0 {
		add("address", "bad")
	} else {
		add("address", "ok", "ip", ips[0].String())
	}

	// 2. the port the program really got
	s.lmu.Lock()
	l := s.listeners
	s.lmu.Unlock()
	if l.HTTPAddr != "" {
		if l.HTTPConfigured != "" && !samePort(l.HTTPConfigured, l.HTTPAddr) {
			add("port", "warn", "wanted", l.HTTPConfigured, "actual", l.HTTPAddr)
		} else {
			add("port", "ok", "addr", l.HTTPAddr)
		}
	}

	if s.cfg.HTTPS.Enabled && l.HTTPSAddr != "" {
		if l.HTTPSConfigured != "" && !samePort(l.HTTPSConfigured, l.HTTPSAddr) {
			add("port-https", "warn", "wanted", l.HTTPSConfigured, "actual", l.HTTPSAddr)
		} else {
			add("port-https", "ok", "addr", l.HTTPSAddr)
		}
	}

	// 3. HTTPS and its certificate
	switch {
	case !s.cfg.HTTPS.Enabled:
		add("https", "info", "state", "off")
	case s.cfg.HTTPS.ACME.Enabled:
		if m := s.certificates(); m == nil {
			add("https", "info", "state", "acme-starting")
		} else {
			st := m.Status()
			switch {
			case st.State == "failed":
				add("https", "bad", "state", "acme-failed", "domain", st.Domain, "error", st.LastError)
			case st.State == "ok" && st.DaysLeft <= 14:
				add("https", "warn", "state", "acme-expiring", "domain", st.Domain, "days", st.DaysLeft)
			case st.State == "ok":
				add("https", "ok", "state", "acme", "domain", st.Domain, "days", st.DaysLeft)
			default:
				add("https", "info", "state", "acme-working", "domain", st.Domain)
			}
		}
	case s.cfg.HTTPS.CertFile != "":
		add("https", "ok", "state", "file")
	default:
		add("https", "info", "state", "self-signed")
	}

	// 4. the .local name
	switch ms := s.MDNSStatus(); {
	case !ms.Enabled:
		add("name", "info", "state", "off")
	case ms.Error != "":
		add("name", "warn", "state", "error", "error", ms.Error)
	case ms.Renamed:
		add("name", "warn", "state", "renamed", "requested", ms.Requested+".local", "host", ms.Host)
	case ms.Active:
		add("name", "ok", "state", "active", "host", ms.Host)
	default:
		add("name", "info", "state", "starting")
	}

	// 5. the settings file can be saved (Setup changes it)
	if s.cfg.Path != "" {
		if f, err := os.OpenFile(s.cfg.Path, os.O_WRONLY, 0); err != nil {
			add("settings", "bad", "path", s.cfg.Path, "error", err.Error())
		} else {
			f.Close()
			add("settings", "ok")
		}
	}

	// 6. the show
	if id := s.rt.ShowID(); id == "" {
		add("show", "info", "state", "none")
	} else if ld, err := s.loadShow(id); err == nil && len(ld.skipped) > 0 {
		add("show", "warn", "state", "languages", "list", strings.Join(sortedKeys(ld.skipped), ", "))
	} else {
		add("show", "ok", "state", "loaded", "name", s.rt.ShowName())
	}

	// 7. who is connected
	add("devices", "info", "n", s.rt.AudienceCount())

	// 8. what the log says
	n := 0
	if s.logs != nil {
		for _, ln := range s.logs.Lines() {
			if m := logLevelRe.FindStringSubmatch(ln); m != nil && (m[1] == "WARN" || m[1] == "ERROR") {
				n++
			}
		}
	}
	if n > 0 {
		add("problems", "warn", "n", n)
	} else {
		add("problems", "ok")
	}
	if logx.Current() == logx.Debug {
		add("debug", "info")
	}
	writeJSON(w, http.StatusOK, map[string]any{"checks": cs, "problems": problems, "level": logx.Current().String()})
}

// samePort reports whether two listen addresses have the same port. The address asked for (":80") and the one the
// system reports ("[::]:80") are written differently when the program got what it asked for; only a different port
// means it fell back. A port of 0 asked for any free one.
func samePort(wanted, actual string) bool {
	return portOf(wanted) == "0" || portOf(wanted) == portOf(actual)
}
