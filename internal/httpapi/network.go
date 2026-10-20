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

// ApplyNetwork makes the ticked interfaces (config network.interfaces) and the restriction the rule for which of
// this computer's addresses are offered. The values must already be valid (config.Load checks them).
func (s *Server) ApplyNetwork(n config.Network) {
	s.netMu.Lock()
	s.cfg.Network = n
	s.netMu.Unlock()
	netinfo.SetInterfaces(n.Interfaces)
	netinfo.SetRestrictListening(n.RestrictListening)
}

// PinNetwork records that the interfaces were given on the command line (or environment) and so override
// config.yaml at every start.
func (s *Server) PinNetwork() { s.pinnedNetwork = true }

func (s *Server) networkSettings() config.Network {
	s.netMu.Lock()
	defer s.netMu.Unlock()
	return config.Network{Interfaces: append([]string{}, s.cfg.Network.Interfaces...), RestrictListening: s.cfg.Network.RestrictListening}
}

// interfaceInfo is one network interface for Setup's list: a tick, its kind, its addresses and what is done with it.
type interfaceInfo struct {
	Name      string   `json:"name"`
	Kind      string   `json:"kind"` // "ethernet", "wifi" or "other"
	Addresses []string `json:"addresses"`
	Selected  bool     `json:"selected"`          // ticked
	Used      bool     `json:"used"`              // an address of it is offered
	Instead   string   `json:"instead,omitempty"` // ticked but not used: the interface offered for the same network
	InsteadIP string   `json:"insteadIp,omitempty"`
	Current   bool     `json:"current"` // the interface this request arrived on
}

// currentInterface is the interface of this computer that the request came in on; "" for this computer itself
// (localhost) or when it cannot be told.
func currentInterface(r *http.Request) string {
	if la, ok := r.Context().Value(http.LocalAddrContextKey).(net.Addr); ok {
		if ta, ok := la.(*net.TCPAddr); ok && ta.IP != nil && !ta.IP.IsLoopback() {
			if name, ok := netinfo.InterfaceOf(ta.IP); ok {
				return name
			}
		}
	}
	return ""
}

// networkInterfaces lists this computer's working network interfaces and what is done with each.
func (s *Server) networkInterfaces(r *http.Request) []interfaceInfo {
	cur := currentInterface(r)
	out := []interfaceInfo{}
	for _, it := range netinfo.Interfaces() {
		x := interfaceInfo{Name: it.Name, Kind: it.Kind, Addresses: []string{}, Selected: it.Selected, Used: it.Used, Instead: it.Instead, Current: cur != "" && strings.EqualFold(cur, it.Name)}
		for _, ip := range it.Addresses {
			x.Addresses = append(x.Addresses, ip.String())
		}
		if it.InsteadIP != nil {
			x.InsteadIP = it.InsteadIP.String()
		}
		out = append(out, x)
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

// setNetwork is PUT /api/admin/settings/network: choose which network interfaces are used.
func (s *Server) setNetwork(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Interfaces        []string `json:"interfaces"`
		RestrictListening bool     `json:"restrictListening"`
	}
	if !decode(w, r, &req) {
		return
	}
	names := cleanList(req.Interfaces)
	if len(names) == 0 {
		writeErr(w, http.StatusBadRequest, "Tick at least one network interface.")
		return
	}
	if len(names) > 100 {
		writeErr(w, http.StatusBadRequest, "Too many network interfaces.")
		return
	}
	// Do not cut off the interface Setup is being used through. This computer itself (localhost) has none to protect.
	if cur := currentInterface(r); cur != "" {
		ticked := false
		for _, n := range names {
			if strings.EqualFold(n, cur) {
				ticked = true
			}
		}
		if !ticked {
			writeErr(w, http.StatusConflict, fmt.Sprintf("You are using Setup through the interface %s, which this would turn off. Open Setup on this computer (at localhost) or through an interface that stays on, then try again.", cur))
			return
		}
	}
	saved := false
	if s.cfg.Path != "" {
		if err := config.SetValues(s.cfg.Path, map[string]any{"network.interfaces": names, "network.restrict_listening": req.RestrictListening}); err != nil {
			writeErr(w, http.StatusInternalServerError, "Could not save the setting in config.yaml: "+err.Error())
			return
		}
		saved = true
	}
	s.ApplyNetwork(config.Network{Interfaces: names, RestrictListening: req.RestrictListening})
	logx.Infof("network interfaces changed in Setup: %v; refuse connections to the others: %v", names, req.RestrictListening)
	resp := map[string]any{"saved": saved, "interfaces": s.networkInterfaces(r)}
	if s.pinnedNetwork {
		resp["note"] = "This copy was started with --network-interfaces (or OPEN_CAPTION_NETWORK_INTERFACES), which overrides the saved setting each time it starts. The new choice applies now, and lasts until the program is restarted."
	}
	writeJSON(w, http.StatusOK, resp)
}
