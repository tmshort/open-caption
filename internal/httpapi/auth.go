package httpapi

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Roles. An admin session also satisfies operator checks (docs/04).
const (
	RoleOperator = "operator"
	RoleAdmin    = "admin"
)

const (
	cookieName     = "oc_session"
	sessionTTL     = 12 * time.Hour
	loginAttempts  = 5 // per IP per minute
	loginWindow    = time.Minute
	maxLoginBody   = 4 << 10
	maxSessionKeep = 1000
)

type session struct {
	role    string
	expires time.Time
}

// savedSessions is the on-disk form. Only a hash of each token is stored, and
// the file is discarded if either password changed since it was written.
type savedSessions struct {
	Fingerprint string                  `json:"fingerprint"`
	Sessions    map[string]savedSession `json:"sessions"`
}

type savedSession struct {
	Role    string `json:"role"`
	Expires int64  `json:"expires"`
}

func tokenKey(token string) string {
	h := sha256.Sum256([]byte("session:" + token))
	return hex.EncodeToString(h[:])
}

// auth keeps sessions and login rate limits in memory.
type auth struct {
	adminPW, operatorPW [32]byte // sha256 of the configured passwords

	path string // where sessions are saved ("" = memory only)
	fp   string

	mu       sync.Mutex
	sessions map[string]session // keyed by tokenKey(token)
	attempts map[string][]time.Time
}

// newAuth creates the session store. With a path, sessions survive a restart,
// so an operator is not signed out if the program is restarted mid-show.
func newAuth(adminPW, operatorPW, path string) *auth {
	fp := sha256.Sum256([]byte(adminPW + "\x00" + operatorPW))
	a := &auth{adminPW: sha256.Sum256([]byte(adminPW)), operatorPW: sha256.Sum256([]byte(operatorPW)),
		path: path, fp: hex.EncodeToString(fp[:]),
		sessions: map[string]session{}, attempts: map[string][]time.Time{}}
	a.load()
	return a
}

func (a *auth) load() {
	if a.path == "" {
		return
	}
	b, err := os.ReadFile(a.path)
	if err != nil {
		return
	}
	var sv savedSessions
	if json.Unmarshal(b, &sv) != nil || sv.Fingerprint != a.fp {
		return // unreadable, or a password changed: everyone signs in again
	}
	for k, s := range sv.Sessions {
		if exp := time.Unix(s.Expires, 0); time.Now().Before(exp) && (s.Role == RoleAdmin || s.Role == RoleOperator) {
			a.sessions[k] = session{role: s.Role, expires: exp}
		}
	}
}

// saveLocked writes the sessions (caller holds a.mu).
func (a *auth) saveLocked() {
	if a.path == "" {
		return
	}
	sv := savedSessions{Fingerprint: a.fp, Sessions: map[string]savedSession{}}
	for k, s := range a.sessions {
		if time.Now().Before(s.expires) {
			sv.Sessions[k] = savedSession{Role: s.role, Expires: s.expires.Unix()}
		}
	}
	b, err := json.Marshal(sv)
	if err != nil {
		return
	}
	os.MkdirAll(filepath.Dir(a.path), 0o755)
	tmp := a.path + ".tmp"
	if os.WriteFile(tmp, b, 0o600) == nil {
		os.Rename(tmp, a.path)
	}
}

func pwEqual(a [32]byte, pw string) bool {
	b := sha256.Sum256([]byte(pw))
	return subtle.ConstantTimeCompare(a[:], b[:]) == 1
}

// allow records a login attempt from ip and reports whether it is within the limit.
func (a *auth) allow(ip string) bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	now := time.Now()
	recent := a.attempts[ip][:0]
	for _, t := range a.attempts[ip] {
		if now.Sub(t) < loginWindow {
			recent = append(recent, t)
		}
	}
	if len(recent) >= loginAttempts {
		a.attempts[ip] = recent
		return false
	}
	a.attempts[ip] = append(recent, now)
	return true
}

// login checks a password for the requested role and returns a session token.
// Role "operator" accepts the operator or the admin password; "admin" only the admin one.
func (a *auth) login(role, password string) (token, granted string, ok bool) {
	switch {
	case role == RoleAdmin && pwEqual(a.adminPW, password):
		granted = RoleAdmin
	case role == RoleOperator && pwEqual(a.adminPW, password):
		granted = RoleAdmin
	case role == RoleOperator && pwEqual(a.operatorPW, password):
		granted = RoleOperator
	default:
		return "", "", false
	}
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", "", false
	}
	token = base64.RawURLEncoding.EncodeToString(b)
	a.mu.Lock()
	defer a.mu.Unlock()
	if len(a.sessions) >= maxSessionKeep {
		for k, s := range a.sessions {
			if time.Now().After(s.expires) {
				delete(a.sessions, k)
			}
		}
	}
	a.sessions[tokenKey(token)] = session{role: granted, expires: time.Now().Add(sessionTTL)}
	a.saveLocked()
	return token, granted, true
}

func (a *auth) lookup(token string) (session, bool) {
	a.mu.Lock()
	defer a.mu.Unlock()
	k := tokenKey(token)
	s, ok := a.sessions[k]
	if !ok || time.Now().After(s.expires) {
		delete(a.sessions, k)
		return session{}, false
	}
	return s, true
}

func (a *auth) revoke(token string) {
	a.mu.Lock()
	delete(a.sessions, tokenKey(token))
	a.saveLocked()
	a.mu.Unlock()
}

type ctxKey struct{}

// Principal is the authenticated caller.
type Principal struct {
	Token string
	Role  string
}

func principal(r *http.Request) Principal {
	p, _ := r.Context().Value(ctxKey{}).(Principal)
	return p
}

// require wraps h so only sessions with at least role may call it.
func (s *Server) require(role string, h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c, err := r.Cookie(cookieName)
		if err != nil {
			writeErr(w, http.StatusUnauthorized, "login required")
			return
		}
		sess, ok := s.auth.lookup(c.Value)
		if !ok || (role == RoleAdmin && sess.role != RoleAdmin) {
			code := http.StatusUnauthorized
			if ok {
				code = http.StatusForbidden
			}
			writeErr(w, code, "not allowed")
			return
		}
		h(w, r.WithContext(context.WithValue(r.Context(), ctxKey{}, Principal{Token: c.Value, Role: sess.role})))
	}
}

func writeErr(w http.ResponseWriter, code int, msg string) {
	writeJSON(w, code, map[string]any{"error": msg})
}

func clientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

func (s *Server) authRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/login", func(w http.ResponseWriter, r *http.Request) {
		var req struct{ Role, Password string }
		if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxLoginBody)).Decode(&req); err != nil {
			writeErr(w, http.StatusBadRequest, "bad request")
			return
		}
		if !s.auth.allow(clientIP(r)) {
			w.Header().Set("Retry-After", "60")
			writeErr(w, http.StatusTooManyRequests, "too many attempts; wait a minute")
			return
		}
		token, role, ok := s.auth.login(req.Role, req.Password)
		if !ok {
			writeErr(w, http.StatusUnauthorized, "wrong password")
			return
		}
		http.SetCookie(w, &http.Cookie{Name: cookieName, Value: token, Path: "/", HttpOnly: true,
			SameSite: http.SameSiteStrictMode, Secure: r.TLS != nil, MaxAge: int(sessionTTL.Seconds())})
		writeJSON(w, http.StatusOK, map[string]any{"role": role})
	})
	mux.HandleFunc("POST /api/logout", func(w http.ResponseWriter, r *http.Request) {
		if c, err := r.Cookie(cookieName); err == nil {
			s.auth.revoke(c.Value)
			s.dropRehearsal(c.Value)
		}
		http.SetCookie(w, &http.Cookie{Name: cookieName, Value: "", Path: "/", HttpOnly: true, MaxAge: -1,
			SameSite: http.SameSiteStrictMode, Secure: r.TLS != nil})
		writeJSON(w, http.StatusOK, map[string]any{"ok": true})
	})
	mux.HandleFunc("GET /api/me", s.require(RoleOperator, func(w http.ResponseWriter, r *http.Request) {
		p := principal(r)
		writeJSON(w, http.StatusOK, map[string]any{"role": p.Role, "sid": SID(p.Token)})
	}))
}
