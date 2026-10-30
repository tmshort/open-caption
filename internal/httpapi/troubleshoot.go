// Copyright the open-caption authors
// SPDX-License-Identifier: Apache-2.0

package httpapi

import (
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/tmshort/open-caption/internal/config"
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

	// 5b. the data folder holds the shows, reports and live state: it must be writable and have room
	if dir := s.cfg.DataDir; dir != "" {
		if f, err := os.CreateTemp(dir, ".write-test-*"); err != nil {
			add("data", "bad", "state", "unwritable", "path", dir, "error", err.Error())
		} else {
			f.Close()
			os.Remove(f.Name())
			if free, ok := freeBytes(dir); ok && free < lowDiskBytes {
				add("data", "warn", "state", "low", "mb", free>>20)
			} else if ok {
				add("data", "ok", "state", "ok", "mb", free>>20)
			} else {
				add("data", "ok", "state", "writable")
			}
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

	// 5c. the clock: a certificate is valid for a period and the report is stamped with the time, so a clock that is far off
	// (a computer that has lost power, a wrong time zone setting) causes confusing trouble. The browser says what time it thinks it is.
	if ms, err := strconv.ParseInt(r.URL.Query().Get("now"), 10, 64); err == nil && ms > 0 {
		if d := time.Since(time.UnixMilli(ms)); d > clockSlack || d < -clockSlack {
			add("clock", "warn", "minutes", int(d.Round(time.Minute)/time.Minute)) // positive: the program's clock is ahead of the browser's
		}
	}

	// 6a. the show file itself: errors stop a line being shown as written, a missing image shows nothing
	if id := s.rt.ShowID(); id != "" {
		if text, err := s.store.ReadShowFile(id, ""); err == nil {
			a, _ := analyze(text, mustAssets(s.store, id), false)
			errs, warns := 0, 0
			for _, d := range a.Diagnostics {
				switch d.Severity {
				case "error":
					errs++
				case "warning":
					warns++
				}
			}
			var names []string
			for _, m := range a.Missing {
				names = append(names, m.Name)
			}
			switch {
			case errs > 0:
				add("showfile", "bad", "state", "errors", "id", id, "n", errs)
			case len(names) > 0:
				add("showfile", "warn", "state", "missing", "id", id, "n", len(names), "list", strings.Join(names, ", "))
			case warns > 0:
				add("showfile", "info", "state", "warnings", "id", id, "n", warns)
			}
		}
	}

	// 6b. what the audience sees: the standby screen until the operator starts the show, then the captions
	if s.rt.ShowID() != "" {
		if s.rt.OperatorPayload().Op.Live {
			add("onair", "info", "state", "live", "name", s.rt.ShowName())
		} else {
			add("onair", "info", "state", "standby", "name", s.rt.ShowName())
		}
	}

	// 6c. who may open the pages: the lists that turn people away, when they are likely to surprise
	adm, op, vw := s.allowLists()
	acc := s.snapshot()
	switch {
	case acc.ViewerPassword != "" && op.Restricted() && !vw.Restricted():
		add("access", "warn", "state", "follow-open")
	case onlyThisComputer(acc.Access.OperatorAllow):
		add("access", "info", "state", "operator-local")
	case onlyThisComputer(acc.Access.AdminAllow):
		add("access", "info", "state", "setup-local")
	case adm.Restricted() || op.Restricted() || vw.Restricted():
		add("access", "ok", "state", "limited")
	}

	// 6d. the passwords: only what is worth a person's attention (never a value)
	switch {
	case config.SamePassword(acc.AdminPassword, acc.OperatorPassword):
		add("passwords", "warn", "state", "same")
	case !config.IsHashed(acc.AdminPassword) || !config.IsHashed(acc.OperatorPassword):
		add("passwords", "info", "state", "plain")
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

// clockSlack is how far the program's clock may differ from the browser's before Troubleshoot says so.
const clockSlack = 5 * time.Minute

// lowDiskBytes is the free space below which Troubleshoot warns: saved reports, versions and uploaded images need room.
const lowDiskBytes = 500 << 20

// samePort reports whether two listen addresses have the same port. The address asked for (":80") and the one the
// system reports ("[::]:80") are written differently when the program got what it asked for; only a different port
// means it fell back. A port of 0 asked for any free one.
func samePort(wanted, actual string) bool {
	return portOf(wanted) == "0" || portOf(wanted) == portOf(actual)
}

// onlyThisComputer reports whether an access list admits nothing but this computer (localhost, the default for Setup).
func onlyThisComputer(list []string) bool {
	if len(list) == 0 {
		return false
	}
	for _, e := range list {
		switch strings.ToLower(strings.TrimSpace(e)) {
		case "localhost", "127.0.0.1", "::1", "[::1]":
		default:
			return false
		}
	}
	return true
}
