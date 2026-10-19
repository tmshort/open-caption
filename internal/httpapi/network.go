package httpapi

import (
	"fmt"
	"net"
	"net/http"
	"strings"

	"github.com/tmshort/open-caption/internal/config"
	"github.com/tmshort/open-caption/internal/logx"
	"github.com/tmshort/open-caption/internal/netinfo"
)

// ApplyNetwork makes the include and exclude lists (config network.include / network.exclude) the rule for which
// of this computer's addresses are offered. The lists must already be valid (config.Load checks them).
func (s *Server) ApplyNetwork(include, exclude []string, restrict bool) {
	f, err := config.NewAddressFilter(include, exclude)
	if err != nil {
		logx.Warnf("network: %v; every address is offered", err)
		f = config.AddressFilter{}
	}
	s.netMu.Lock()
	s.cfg.Network = config.Network{Include: include, Exclude: exclude, RestrictListening: restrict}
	s.netMu.Unlock()
	if f.Active() {
		netinfo.SetFilter(f.Allows)
	} else {
		netinfo.SetFilter(nil)
	}
	netinfo.SetRestrictListening(restrict)
}

// PinNetwork records that the lists were given on the command line (or environment) and so override config.yaml
// at every start.
func (s *Server) PinNetwork() { s.pinnedNetwork = true }

func (s *Server) networkSettings() config.Network {
	s.netMu.Lock()
	defer s.netMu.Unlock()
	return config.Network{Include: append([]string{}, s.cfg.Network.Include...), Exclude: append([]string{}, s.cfg.Network.Exclude...), RestrictListening: s.cfg.Network.RestrictListening}
}

type addressInfo struct {
	IP        string `json:"ip"`
	Interface string `json:"interface"`
	Shown     bool   `json:"shown"`
}

// networkAddresses lists every address of this computer's working network interfaces, and whether the rules offer it.
func (s *Server) networkAddresses() []addressInfo {
	n := s.networkSettings()
	f, _ := config.NewAddressFilter(n.Include, n.Exclude)
	out := []addressInfo{}
	for _, a := range netinfo.AllAddresses() {
		out = append(out, addressInfo{IP: a.IP.String(), Interface: a.Interface, Shown: f.Allows(a.IP)})
	}
	return out
}

// cleanList trims the entries and drops empty ones and repeats.
func cleanList(in []string) []string {
	out := []string{}
	seen := map[string]bool{}
	for _, e := range in {
		e = strings.TrimSpace(e)
		if e == "" || seen[strings.ToLower(e)] {
			continue
		}
		seen[strings.ToLower(e)] = true
		out = append(out, e)
	}
	return out
}

// setNetwork is PUT /api/admin/settings/network: choose which addresses are offered to the audience.
func (s *Server) setNetwork(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Include           []string `json:"include"`
		Exclude           []string `json:"exclude"`
		RestrictListening bool     `json:"restrictListening"`
	}
	if !decode(w, r, &req) {
		return
	}
	include, exclude := cleanList(req.Include), cleanList(req.Exclude)
	if len(include) > 50 || len(exclude) > 50 {
		writeErr(w, http.StatusBadRequest, "Use at most 50 lines in each box.")
		return
	}
	if _, err := config.NewAddressFilter(include, nil); err != nil {
		writeErr(w, http.StatusBadRequest, "A line in Show only these is not an address, a network like 192.168.1.0/24 or the word private.")
		return
	}
	if _, err := config.NewAddressFilter(nil, exclude); err != nil {
		writeErr(w, http.StatusBadRequest, "A line in Never show these is not an address, a network like 192.168.1.0/24 or the word private.")
		return
	}
	// Do not lock out the person who is changing it: refuse a restriction that would cut off the address this
	// very request arrived on (this computer itself, at localhost, is never cut off).
	if req.RestrictListening {
		f, _ := config.NewAddressFilter(include, exclude)
		if la, ok := r.Context().Value(http.LocalAddrContextKey).(net.Addr); ok {
			if ta, ok := la.(*net.TCPAddr); ok && !ta.IP.IsLoopback() && !f.Allows(ta.IP) {
				writeErr(w, http.StatusConflict, fmt.Sprintf("You are using Setup through %s, which this would refuse. Open Setup on this computer (at localhost) or through an address that stays shown, then try again.", ta.IP))
				return
			}
		}
	}
	saved := false
	if s.cfg.Path != "" {
		if err := config.SetValues(s.cfg.Path, map[string]any{"network.include": include, "network.exclude": exclude, "network.restrict_listening": req.RestrictListening}); err != nil {
			writeErr(w, http.StatusInternalServerError, "Could not save the setting in config.yaml: "+err.Error())
			return
		}
		saved = true
	}
	s.ApplyNetwork(include, exclude, req.RestrictListening)
	logx.Infof("network addresses changed in Setup: only %v, never %v, refuse connections to the others: %v", include, exclude, req.RestrictListening)
	resp := map[string]any{"saved": saved, "addresses": s.networkAddresses()}
	if s.pinnedNetwork {
		resp["note"] = "This copy was started with --network-include or --network-exclude (or the matching OPEN_CAPTION_ variable), which overrides the saved setting each time it starts. The new choice applies now, and lasts until the program is restarted."
	}
	writeJSON(w, http.StatusOK, resp)
}
