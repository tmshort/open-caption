package httpapi

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"net"
	"net/http"
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

// auth keeps sessions and login rate limits in memory.
type auth struct {
	adminPW, operatorPW [32]byte // sha256 of the configured passwords

	mu       sync.Mutex
	sessions map[string]session
	attempts map[string][]time.Time
}

func newAuth(adminPW, operatorPW string) *auth {
	return &auth{adminPW: sha256.Sum256([]byte(adminPW)), operatorPW: sha256.Sum256([]byte(operatorPW)),
		sessions: map[string]session{}, attempts: map[string][]time.Time{}}
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
	a.sessions[token] = session{role: granted, expires: time.Now().Add(sessionTTL)}
	return token, granted, true
}

func (a *auth) lookup(token string) (session, bool) {
	a.mu.Lock()
	defer a.mu.Unlock()
	s, ok := a.sessions[token]
	if !ok || time.Now().After(s.expires) {
		delete(a.sessions, token)
		return session{}, false
	}
	return s, true
}

func (a *auth) revoke(token string) {
	a.mu.Lock()
	delete(a.sessions, token)
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
