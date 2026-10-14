package httpapi

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/tmshort/open-caption/internal/config"
)

func TestHealthz(t *testing.T) {
	h := New(config.Default(), "test").Handler()
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest("GET", "/healthz", nil))
	if rr.Code != http.StatusOK || strings.TrimSpace(rr.Body.String()) != "ok" {
		t.Errorf("%d %q", rr.Code, rr.Body.String())
	}
}
