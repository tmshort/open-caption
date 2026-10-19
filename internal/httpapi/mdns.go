package httpapi

import (
	"net/http"
	"sync"
	"time"

	"github.com/tmshort/open-caption/internal/config"
	"github.com/tmshort/open-caption/internal/logx"
	"github.com/tmshort/open-caption/internal/mdns"
	"github.com/tmshort/open-caption/internal/netinfo"
)

// responder is a running mDNS announcement (internal/mdns), so tests can stand in for the network.
type responder interface {
	Name() string
	Host() string
	Close()
}

// MDNSStatus describes the local network name (http://open-caption.local/) for Setup.
type MDNSStatus struct {
	Enabled   bool   `json:"enabled"`
	Requested string `json:"requested"` // the name chosen
	Name      string `json:"name"`      // the name in use (differs if another computer already had the one chosen)
	Host      string `json:"host"`      // Name + ".local" while active
	Active    bool   `json:"active"`
	Starting  bool   `json:"starting"`
	Renamed   bool   `json:"renamed"`
	Error     string `json:"error,omitempty"`
}

type mdnsManager struct {
	mu     sync.Mutex
	gen    int
	cur    responder
	status MDNSStatus
	start  func(name string) (responder, error)
}

func newMDNSManager(cfg config.MDNS) *mdnsManager {
	return &mdnsManager{
		status: MDNSStatus{Enabled: cfg.Enabled, Requested: cfg.Name},
		start:  func(name string) (responder, error) { return mdns.Start(name, netinfo.Allowed) },
	}
}

// ApplyMDNS starts, changes or stops the announcement. It returns at once; the claim takes about a second
// (the name is first checked against the rest of the network), so poll MDNSStatus.
func (s *Server) ApplyMDNS(enabled bool, name string) {
	m := s.mdns
	m.mu.Lock()
	m.gen++
	gen := m.gen
	old := m.cur
	m.cur = nil
	m.status = MDNSStatus{Enabled: enabled, Requested: name, Starting: enabled}
	m.mu.Unlock()
	go func() {
		if old != nil {
			old.Close()
		}
		if !enabled {
			return
		}
		r, err := m.start(name)
		m.mu.Lock()
		defer m.mu.Unlock()
		if m.gen != gen { // changed again while this was starting
			if r != nil {
				go r.Close()
			}
			return
		}
		st := MDNSStatus{Enabled: true, Requested: name}
		if err != nil {
			st.Error = err.Error()
			logx.Warnf("mDNS: %s.local is not available: %v", name, err)
		} else {
			m.cur = r
			st.Active, st.Name, st.Host = true, r.Name(), r.Host()
			st.Renamed = st.Name != name
			if st.Renamed {
				logx.Warnf("mDNS: another computer already uses %s.local, so this one is %s", name, st.Host)
			} else {
				logx.Infof("mDNS: this computer can be reached at http://%s/", st.Host)
			}
		}
		m.status = st
	}()
}

// MDNSStatus returns the current state of the local network name.
func (s *Server) MDNSStatus() MDNSStatus {
	s.mdns.mu.Lock()
	defer s.mdns.mu.Unlock()
	return s.mdns.status
}

func (s *Server) stopMDNS() {
	s.mdns.mu.Lock()
	s.mdns.gen++
	old := s.mdns.cur
	s.mdns.cur = nil
	s.mdns.mu.Unlock()
	if old != nil {
		old.Close() // sends the goodbye
	}
}

// setMDNS is PUT /api/admin/settings/mdns: choose whether the server announces itself, and under what name.
func (s *Server) setMDNS(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Enabled bool   `json:"enabled"`
		Name    string `json:"name"`
	}
	if !decode(w, r, &req) {
		return
	}
	name, err := mdns.NormalizeName(req.Name)
	if err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	saved := false
	if s.cfg.Path != "" {
		if err := config.SetValues(s.cfg.Path, map[string]any{"mdns.enabled": req.Enabled, "mdns.name": name}); err != nil {
			writeErr(w, http.StatusInternalServerError, "Could not save the setting in config.yaml: "+err.Error())
			return
		}
		saved = true
	}
	s.ApplyMDNS(req.Enabled, name)
	deadline := time.Now().Add(6 * time.Second)
	for time.Now().Before(deadline) && s.MDNSStatus().Starting {
		time.Sleep(50 * time.Millisecond)
	}
	writeJSON(w, http.StatusOK, map[string]any{"mdns": s.MDNSStatus(), "saved": saved})
}
