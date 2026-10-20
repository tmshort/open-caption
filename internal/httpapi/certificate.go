// Copyright the open-caption authors
// SPDX-License-Identifier: Apache-2.0

package httpapi

import (
	"net/http"

	"github.com/tmshort/open-caption/internal/acme"
)

// CertificateManager is what Setup needs from the automatic-certificate manager (internal/acme).
type CertificateManager interface {
	Status() acme.Status
	RenewNow() bool
}

// SetCertificates gives the server the certificate manager (nil when HTTPS gets its certificate another way).
func (s *Server) SetCertificates(m CertificateManager) {
	s.certMu.Lock()
	s.certs = m
	s.certMu.Unlock()
}

func (s *Server) certificates() CertificateManager {
	s.certMu.Lock()
	defer s.certMu.Unlock()
	return s.certs
}

// certificateInfo is GET /api/admin/settings/certificate: how HTTPS gets its certificate, and while ACME is on, its state.
func (s *Server) certificateInfo(w http.ResponseWriter, r *http.Request) {
	out := map[string]any{"https": s.cfg.HTTPS.Enabled, "providers": acme.Providers(), "available": acme.Available()}
	switch {
	case !s.cfg.HTTPS.Enabled:
		out["source"] = "none"
	case s.cfg.HTTPS.ACME.Enabled:
		out["source"] = "acme"
		out["directory"] = s.cfg.HTTPS.ACME.Directory
		out["provider"] = s.cfg.HTTPS.ACME.DNS.Provider
		if m := s.certificates(); m != nil {
			out["status"] = m.Status()
		}
	case s.cfg.HTTPS.CertFile != "":
		out["source"] = "file"
	default:
		out["source"] = "self-signed"
	}
	writeJSON(w, http.StatusOK, out)
}

// renewCertificate is POST /api/admin/settings/certificate/renew: get a new certificate now (it runs in the background).
func (s *Server) renewCertificate(w http.ResponseWriter, r *http.Request) {
	m := s.certificates()
	if m == nil {
		writeErr(w, http.StatusConflict, "Automatic certificates are not turned on.")
		return
	}
	if !m.RenewNow() {
		writeErr(w, http.StatusConflict, "A certificate is already being fetched.")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"status": m.Status()})
}
