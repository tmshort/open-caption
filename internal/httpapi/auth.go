package httpapi

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"github.com/tmshort/open-caption/internal/logx"
	"net"
	"net/http"
	"net/netip"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// Roles. An admin session also satisfies operator checks (docs/04).
const (
	RoleOperator = "operator"
	RoleAdmin    = "admin"
)

const (
	cookieName      = "oc_session"
	sessionTTL      = 12 * time.Hour
	loginAttempts   = 5 // per IP per minute
	loginWindow     = time.Minute
	maxLoginBody    = 4 << 10
	maxSessionKeep  = 1000
	maxAutoSessions = 20 // sessions made without a password (a cookieless script on this computer makes one per call)
)

type session struct {
	role    string
	expires time.Time
	auto    bool // signed in without a password because the browser is on this computer (see localTrusted)
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
	Auto    bool   `json:"auto,omitempty"`
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
	a := &auth{adminPW: sha256.Sum256([]byte(adminPW)), operatorPW: sha256.Sum256([]byte(operatorPW)), path: path,
		sessions: map[string]session{}, attempts: map[string][]time.Time{}}
	a.fp = fingerprint(a.adminPW, a.operatorPW)
	a.load()
	return a
}

// fingerprint changes whenever either password does (it is derived from hashes, never the passwords).
func fingerprint(admin, operator [32]byte) string {
	h := sha256.Sum256(append(append([]byte("fp:"), admin[:]...), operator[:]...))
	return hex.EncodeToString(h[:])
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
			a.sessions[k] = session{role: s.Role, expires: exp, auto: s.Auto}
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
		if time.Now().Before(s.expires) && !s.auto {
			sv.Sessions[k] = savedSession{Role: s.role, Expires: s.expires.Unix(), Auto: s.auto}
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
	token, ok = a.issue(granted, false)
	return token, granted, ok
}

// issue creates a session with the given role and returns its token.
func (a *auth) issue(role string, auto bool) (string, bool) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", false
	}
	token := base64.RawURLEncoding.EncodeToString(b)
	a.mu.Lock()
	defer a.mu.Unlock()
	if len(a.sessions) >= maxSessionKeep {
		for k, s := range a.sessions {
			if time.Now().After(s.expires) {
				delete(a.sessions, k)
			}
		}
	}
	if auto {
		a.dropOldestAutoLocked()
	}
	a.sessions[tokenKey(token)] = session{role: role, expires: time.Now().Add(sessionTTL), auto: auto}
	if !auto { // an automatic sign-in is made again on demand, so it is not worth a disk write (or saving)
		a.saveLocked()
	}
	return token, true
}

// dropOldestAutoLocked makes room for one more automatic session (caller holds a.mu).
func (a *auth) dropOldestAutoLocked() {
	for {
		n, oldest, ok := 0, "", false
		var first time.Time
		for k, s := range a.sessions {
			if !s.auto {
				continue
			}
			n++
			if !ok || s.expires.Before(first) {
				oldest, first, ok = k, s.expires, true
			}
		}
		if n < maxAutoSessions {
			return
		}
		delete(a.sessions, oldest)
	}
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

// setPasswords switches to new passwords. Logins made with a password that changed are dropped, except
// keep (the person making the change, who stays signed in).
func (a *auth) setPasswords(adminPW, operatorPW, keep string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if adminPW != "" {
		a.adminPW = sha256.Sum256([]byte(adminPW))
	}
	if operatorPW != "" {
		a.operatorPW = sha256.Sum256([]byte(operatorPW))
	}
	a.fp = fingerprint(a.adminPW, a.operatorPW)
	for k, s := range a.sessions {
		if k == tokenKey(keep) {
			continue
		}
		if (adminPW != "" && s.role == RoleAdmin) || (operatorPW != "" && s.role == RoleOperator) {
			delete(a.sessions, k)
		}
	}
	a.saveLocked()
}

// prune forgets expired sessions.
func (a *auth) prune() {
	a.mu.Lock()
	defer a.mu.Unlock()
	for ip, ts := range a.attempts { // addresses that stopped trying need not be remembered
		if len(ts) == 0 || time.Since(ts[len(ts)-1]) >= loginWindow {
			delete(a.attempts, ip)
		}
	}
	changed := false
	for k, s := range a.sessions {
		if time.Now().After(s.expires) {
			delete(a.sessions, k)
			changed = true
		}
	}
	if changed {
		a.saveLocked()
	}
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
	Auto  bool // signed in automatically because the browser is on this computer: there is nothing to sign out of
}

func principal(r *http.Request) Principal {
	p, _ := r.Context().Value(ctxKey{}).(Principal)
	return p
}

// clientAddr is the address the request came from. Proxy headers are deliberately not trusted.
func clientAddr(r *http.Request) netip.Addr {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		host = r.RemoteAddr
	}
	a, err := netip.ParseAddr(host)
	if err != nil {
		return netip.Addr{}
	}
	return a.Unmap()
}

// allowedFor reports whether this computer may use the Setup (RoleAdmin) or operator pages. An admin
// address may always use the operator pages too.
func (s *Server) allowedFor(role string, r *http.Request) bool {
	ip := clientAddr(r)
	if !ip.IsValid() {
		return !s.adminAllow.Restricted() && !s.operatorAllow.Restricted()
	}
	if role == RoleAdmin {
		return s.adminAllow.Allows(ip)
	}
	return s.operatorAllow.Allows(ip) || s.adminAllow.Allows(ip)
}

const notFromHere = "That page is not available from this computer. It can only be opened from the computers listed under access in config.yaml."

// gate protects a page (HTML or its JavaScript) with the access list.
func (s *Server) gate(role string, h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !s.allowedFor(role, r) {
			http.Error(w, notFromHere, http.StatusForbidden)
			return
		}
		h(w, r)
	}
}

// originGuard refuses state-changing /api requests that a browser says came from another website or
// another port of this computer (a forged request), whatever cookies it carries. Requests with no
// Origin header (command-line tools, the program's own tests) are allowed.
func (s *Server) originGuard(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if logx.Enabled(logx.Debug) {
			rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
			start := time.Now()
			defer func() {
				logx.Debugf("%s %s -> %d (%s) from %s", r.Method, r.URL.Path, rec.status, time.Since(start).Round(time.Millisecond), clientIP(r))
			}()
			w = rec
		}
		if strings.HasPrefix(r.URL.Path, "/api/") && r.Method != http.MethodGet && r.Method != http.MethodHead && r.Method != http.MethodOptions {
			site := r.Header.Get("Sec-Fetch-Site")
			bad := site == "cross-site" || site == "same-site"
			if o := r.Header.Get("Origin"); o != "" && o != "null" {
				u, err := url.Parse(o)
				bad = bad || err != nil || u.Host != r.Host
			} else if o == "null" {
				bad = true
			}
			if bad {
				logx.Warnf("refused a %s %s from %s: it did not come from the program's own pages (Origin %q)", r.Method, r.URL.Path, clientIP(r), r.Header.Get("Origin"))
				writeErr(w, http.StatusForbidden, "this request did not come from the program's own pages")
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}

// localHosts are the names a browser on this computer uses to reach it.
var localHosts = map[string]bool{"localhost": true, "127.0.0.1": true, "::1": true}

// proxyHeaders show that a request was relayed by something else (a reverse proxy, a tunnel).
var proxyHeaders = []string{"X-Forwarded-For", "X-Forwarded-Host", "X-Forwarded-Proto", "Forwarded", "X-Real-Ip"}

// localTrusted reports whether the request comes from a browser on this very computer, in which case Setup and
// the operator page open without a password (access.local_auto_login). It is deliberately strict, because
// "the request came from 127.0.0.1" alone is not proof of a person at the keyboard:
//   - the connection must be from a loopback address;
//   - the address in the browser's bar (the Host header) must be localhost, 127.0.0.1 or [::1], which defeats
//     a web page that makes the browser reach this server under its own name (DNS rebinding);
//   - nothing may have relayed it (a reverse proxy on this computer makes every visitor look local);
//   - a browser that says the request was started by another website or another port is refused.
func (s *Server) localTrusted(r *http.Request) bool {
	if !s.cfg.Access.LocalAutoLogin {
		return false
	}
	if ip := clientAddr(r); !ip.IsValid() || !ip.IsLoopback() {
		return false
	}
	host := r.Host
	if h, _, err := net.SplitHostPort(host); err == nil {
		host = h
	}
	if !localHosts[strings.ToLower(strings.Trim(host, "[]"))] {
		return false
	}
	for _, h := range proxyHeaders {
		if r.Header.Get(h) != "" {
			return false
		}
	}
	switch r.Header.Get("Sec-Fetch-Site") {
	case "", "same-origin", "none":
		return true
	}
	return false
}

// require wraps h so only sessions with at least role may call it. A browser on this computer is signed in as
// admin automatically (localTrusted); everyone else needs a session from /api/login.
func (s *Server) require(role string, h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !s.allowedFor(role, r) {
			writeErr(w, http.StatusForbidden, "not available from this computer")
			return
		}
		var sess session
		var token string
		ok := false
		if c, err := r.Cookie(cookieName); err == nil {
			sess, ok = s.auth.lookup(c.Value)
			token = c.Value
		}
		if (!ok || (role == RoleAdmin && sess.role != RoleAdmin)) && s.localTrusted(r) {
			if t, issued := s.auth.issue(RoleAdmin, true); issued {
				http.SetCookie(w, &http.Cookie{Name: cookieName, Value: t, Path: "/", HttpOnly: true,
					SameSite: http.SameSiteStrictMode, Secure: r.TLS != nil, MaxAge: int(sessionTTL.Seconds())})
				logx.Infof("signed in as admin without a password: the browser is on this computer")
				token, sess, ok = t, session{role: RoleAdmin, auto: true}, true
			}
		}
		if !ok {
			writeErr(w, http.StatusUnauthorized, "login required")
			return
		}
		if role == RoleAdmin && sess.role != RoleAdmin {
			writeErr(w, http.StatusForbidden, "not allowed")
			return
		}
		h(w, r.WithContext(context.WithValue(r.Context(), ctxKey{}, Principal{Token: token, Role: sess.role, Auto: sess.auto})))
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
		if (req.Role == RoleAdmin || req.Role == RoleOperator) && !s.allowedFor(req.Role, r) {
			logx.Warnf("sign-in as %s refused from %s: not on the access list", req.Role, clientIP(r))
			writeErr(w, http.StatusForbidden, "signing in is not available from this computer")
			return
		}
		if !s.auth.allow(clientIP(r)) {
			logx.Warnf("too many sign-in attempts from %s", clientIP(r))
			w.Header().Set("Retry-After", "60")
			writeErr(w, http.StatusTooManyRequests, "too many attempts; wait a minute")
			return
		}
		token, role, ok := s.auth.login(req.Role, req.Password)
		if !ok {
			logx.Warnf("failed sign-in as %s from %s: wrong password", req.Role, clientIP(r))
			writeErr(w, http.StatusUnauthorized, "wrong password")
			return
		}
		http.SetCookie(w, &http.Cookie{Name: cookieName, Value: token, Path: "/", HttpOnly: true,
			SameSite: http.SameSiteStrictMode, Secure: r.TLS != nil, MaxAge: int(sessionTTL.Seconds())})
		logx.Infof("%s signed in from %s", role, clientIP(r))
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

// statusRecorder notes the status code for the debug log. It passes flushing through, which the
// event streams need.
type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(code int) { r.status = code; r.ResponseWriter.WriteHeader(code) }
func (r *statusRecorder) Flush() {
	if f, ok := r.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}
func (r *statusRecorder) Unwrap() http.ResponseWriter { return r.ResponseWriter }
