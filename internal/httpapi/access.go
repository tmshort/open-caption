// Copyright the open-caption authors
// SPDX-License-Identifier: Apache-2.0

package httpapi

import (
	"net/http"

	"github.com/tmshort/open-caption/internal/config"
	"github.com/tmshort/open-caption/internal/logx"
)

// maxAccessEntries is the most entries one list may hold.
const maxAccessEntries = 100

// allowLists returns the lists that say which computers may open Setup, the operator page and (with the
// viewer-only password) the Follow page. Setup can change them while the program runs.
func (s *Server) allowLists() (admin, operator, viewer config.AllowList) {
	s.accessMu.RLock()
	defer s.accessMu.RUnlock()
	return s.adminAllow, s.operatorAllow, s.viewerAllow
}

func (s *Server) adminRestricted() bool    { a, _, _ := s.allowLists(); return a.Restricted() }
func (s *Server) operatorRestricted() bool { _, o, _ := s.allowLists(); return o.Restricted() }

// localAutoLogin reports whether a browser on this computer is signed in without a password (access.local_auto_login).
func (s *Server) localAutoLogin() bool { return s.snapshot().Access.LocalAutoLogin }

// accessView is the access settings as Setup shows them.
func (s *Server) accessView() map[string]any {
	a := s.snapshot().Access
	_, _, v := s.allowLists()
	return map[string]any{
		"adminAllow": nonNil(a.AdminAllow), "operatorAllow": nonNil(a.OperatorAllow), "viewerAllow": nonNil(a.ViewerAllow),
		"viewerRestricted": v.Restricted(), "localAutoLogin": a.LocalAutoLogin,
	}
}

func nonNil(l []string) []string {
	if l == nil {
		return []string{}
	}
	return l
}

// setAccess is PUT /api/admin/settings/access: who may open Setup, the operator page and the Follow page (the three
// lists), and whether a browser on this computer is signed in without a password. It is kept in config.yaml and applies
// at once. A change that would shut out the computer making it is refused, so Setup cannot be locked away by a slip.
func (s *Server) setAccess(w http.ResponseWriter, r *http.Request) {
	var req struct {
		AdminAllow     []string `json:"adminAllow"`
		OperatorAllow  []string `json:"operatorAllow"`
		ViewerAllow    []string `json:"viewerAllow"`
		LocalAutoLogin bool     `json:"localAutoLogin"`
	}
	if !decode(w, r, &req) {
		return
	}
	admin, operator, viewer := cleanList(req.AdminAllow), cleanList(req.OperatorAllow), cleanList(req.ViewerAllow)
	if len(admin) > maxAccessEntries || len(operator) > maxAccessEntries || len(viewer) > maxAccessEntries {
		writeErr(w, http.StatusBadRequest, "Too many entries in a list.")
		return
	}
	var lists [3]config.AllowList
	for i, l := range [][]string{admin, operator, viewer} {
		p, err := config.ParseAllowList(l)
		if err != nil {
			writeErr(w, http.StatusBadRequest, "An entry is not an address, a network like 192.168.1.0/24, “localhost” or “private”.")
			return
		}
		lists[i] = p
	}
	if ip := clientAddr(r); ip.IsValid() && !lists[0].Allows(ip) {
		writeErr(w, http.StatusConflict, "This would stop the computer you are using from opening Setup. Keep it in the list for Setup (tick “This computer” if you are using Setup on it), then try again.")
		return
	}
	saved := false
	if s.cfg.Path != "" {
		kv := map[string]any{"access.admin_allow": admin, "access.operator_allow": operator, "access.viewer_allow": viewer, "access.local_auto_login": req.LocalAutoLogin}
		if err := config.SetValues(s.cfg.Path, kv); err != nil {
			writeErr(w, http.StatusInternalServerError, "Could not save the setting in config.yaml: "+err.Error())
			return
		}
		saved = true
	}
	s.cfgMu.Lock()
	s.cfg.Access = config.Access{AdminAllow: admin, OperatorAllow: operator, ViewerAllow: viewer, LocalAutoLogin: req.LocalAutoLogin}
	s.cfgMu.Unlock()
	s.accessMu.Lock()
	s.adminAllow, s.operatorAllow, s.viewerAllow = lists[0], lists[1], lists[2]
	s.accessMu.Unlock()
	logx.Infof("who may open Setup, the operator page and the Follow page changed in Setup: %v, %v, %v; local sign-in without a password: %v", admin, operator, viewer, req.LocalAutoLogin)
	writeJSON(w, http.StatusOK, map[string]any{"saved": saved, "access": s.accessView(), "adminRestricted": lists[0].Restricted(), "operatorRestricted": lists[1].Restricted()})
}
