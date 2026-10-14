package httpapi

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/tmshort/open-caption/internal/config"
	"github.com/tmshort/open-caption/internal/live"
	"github.com/tmshort/open-caption/internal/script"
)

func testConfig() config.Config {
	c := config.Default()
	c.AdminPassword, c.OperatorPassword = "adminpw", "operatorpw"
	return c
}

func opServer(t *testing.T, src string) (*Server, *httptest.Server) {
	t.Helper()
	s := New(testConfig(), "test")
	t.Cleanup(s.Close)
	s.Runtime().SetShow("demo", "Demo Show", t.TempDir(), script.Parse(src, script.Options{}).Cues, DefaultTheme("Soon"))
	ts := httptest.NewServer(s.Handler())
	t.Cleanup(ts.Close)
	return s, ts
}

type client struct {
	t    *testing.T
	base string
	hc   *http.Client
}

func newClient(t *testing.T, base string) *client {
	jar, _ := cookiejar.New(nil)
	return &client{t: t, base: base, hc: &http.Client{Jar: jar}}
}

func (c *client) do(method, path string, body any) (int, map[string]any) {
	c.t.Helper()
	var rdr *bytes.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		rdr = bytes.NewReader(b)
	} else {
		rdr = bytes.NewReader(nil)
	}
	req, _ := http.NewRequest(method, c.base+path, rdr)
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.hc.Do(req)
	if err != nil {
		c.t.Fatal(err)
	}
	defer resp.Body.Close()
	var out map[string]any
	json.NewDecoder(resp.Body).Decode(&out)
	return resp.StatusCode, out
}

func (c *client) login(role, pw string) int {
	code, _ := c.do("POST", "/api/login", map[string]string{"role": role, "password": pw})
	return code
}

func TestLoginRolesAndPermissions(t *testing.T) {
	_, ts := opServer(t, "A\nB")
	anon := newClient(t, ts.URL)
	if code, _ := anon.do("GET", "/api/operator/state", nil); code != 401 {
		t.Errorf("anonymous: %d", code)
	}
	if code, _ := anon.do("POST", "/api/operator/next", nil); code != 401 {
		t.Errorf("anonymous next: %d", code)
	}

	opc := newClient(t, ts.URL)
	if opc.login("operator", "wrong") != 401 {
		t.Error("wrong password accepted")
	}
	if opc.login("operator", "operatorpw") != 200 {
		t.Fatal("operator login failed")
	}
	if code, _ := opc.do("GET", "/api/operator/state", nil); code != 200 {
		t.Errorf("operator state: %d", code)
	}
	if code, _ := opc.do("GET", "/api/admin/ping", nil); code != 403 {
		t.Errorf("operator on admin api: %d, want 403", code)
	}

	adm := newClient(t, ts.URL)
	if adm.login("admin", "operatorpw") != 401 {
		t.Error("admin role must not accept the operator password")
	}
	if adm.login("admin", "adminpw") != 200 {
		t.Fatal("admin login failed")
	}
	if code, _ := adm.do("GET", "/api/admin/ping", nil); code != 200 {
		t.Errorf("admin ping: %d", code)
	}
	if code, _ := adm.do("GET", "/api/operator/state", nil); code != 200 {
		t.Errorf("admin on operator api: %d", code)
	}
	both := newClient(t, ts.URL)
	if both.login("operator", "adminpw") != 200 {
		t.Error("the operator login should accept the admin password")
	}
	if code, _ := both.do("GET", "/api/admin/ping", nil); code != 200 {
		t.Errorf("admin password used at /operator should give an admin session: %d", code)
	}

	opc.do("POST", "/api/logout", nil)
	if code, _ := opc.do("GET", "/api/operator/state", nil); code != 401 {
		t.Errorf("after logout: %d", code)
	}
}

func TestSessionCookieFlags(t *testing.T) {
	_, ts := opServer(t, "A")
	b, _ := json.Marshal(map[string]string{"role": "operator", "password": "operatorpw"})
	resp, _ := http.Post(ts.URL+"/api/login", "application/json", bytes.NewReader(b))
	resp.Body.Close()
	c := resp.Cookies()
	if len(c) != 1 || !c[0].HttpOnly || c[0].SameSite != http.SameSiteStrictMode || c[0].Secure {
		t.Errorf("plain-HTTP cookie flags: %+v", c)
	}
	tsTLS := httptest.NewTLSServer(ts.Config.Handler)
	defer tsTLS.Close()
	resp, err := tsTLS.Client().Post(tsTLS.URL+"/api/login", "application/json", bytes.NewReader(b))
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if c := resp.Cookies(); len(c) != 1 || !c[0].Secure {
		t.Errorf("HTTPS cookie must be Secure: %+v", c)
	}
	// No Basic-auth challenge anywhere.
	r2, _ := http.Get(ts.URL + "/api/operator/state")
	r2.Body.Close()
	if r2.Header.Get("WWW-Authenticate") != "" {
		t.Error("must not use HTTP Basic auth")
	}
}

func TestLoginRateLimit(t *testing.T) {
	_, ts := opServer(t, "A")
	c := newClient(t, ts.URL)
	for i := 0; i < loginAttempts; i++ {
		if code := c.login("operator", "nope"); code != 401 {
			t.Fatalf("attempt %d: %d", i, code)
		}
	}
	if code := c.login("operator", "operatorpw"); code != 429 {
		t.Errorf("sixth attempt: %d, want 429", code)
	}
}

func TestOperatorFlow(t *testing.T) {
	s, ts := opServer(t, "[loop: 5s]\nWelcome\n[/loop]\nA\nB\n[image: pic.png]")
	c := newClient(t, ts.URL)
	c.login("operator", "operatorpw")

	code, st := c.do("GET", "/api/operator/state", nil)
	if code != 200 {
		t.Fatal(code)
	}
	cues := st["cues"].([]any)
	if len(cues) != 4 {
		t.Fatalf("cues %v", cues)
	}
	img := cues[3].(map[string]any)
	if img["kind"] != "image" || img["label"] != "pic.png" || strings.Contains(toJSON(img), "/assets/") {
		t.Errorf("operator must see only the image name: %v", img)
	}
	if code, r := c.do("POST", "/api/operator/claim", map[string]bool{"takeover": false}); code != 200 || r["ok"] != true {
		t.Fatalf("claim: %d %v", code, r)
	}
	p := st["payload"].(map[string]any)
	op := p["op"].(map[string]any)
	if op["loop"] == nil || op["live"] != false {
		t.Fatalf("pre-show loop should auto-start: %v", op)
	}

	rev := func(r map[string]any) int64 {
		return int64(r["payload"].(map[string]any)["op"].(map[string]any)["rev"].(float64))
	}
	_, r := c.do("POST", "/api/operator/next", map[string]any{"rev": rev(map[string]any{"payload": p})})
	op = r["payload"].(map[string]any)["op"].(map[string]any)
	if op["loop"] != nil || op["live"] != true || op["cursor"].(float64) != 2 {
		t.Fatalf("NEXT should end loop and send A: %v", op)
	}
	if got := s.Runtime().Engine().Snapshot().Scrollback; len(got) != 1 || got[0].Text != "A" {
		t.Errorf("%+v", got)
	}

	// Stale rev is rejected and changes nothing.
	code, r = c.do("POST", "/api/operator/next", map[string]any{"rev": 0})
	if code != 409 || len(s.Runtime().Engine().Snapshot().Scrollback) != 1 {
		t.Errorf("stale: %d %v", code, r)
	}
	// No rev = no check.
	if code, _ = c.do("POST", "/api/operator/next", nil); code != 200 {
		t.Errorf("next without rev: %d", code)
	}
	// Back, jump, message, blank, clear, reset.
	if code, _ = c.do("POST", "/api/operator/back", nil); code != 200 {
		t.Errorf("back: %d", code)
	}
	if code, _ = c.do("POST", "/api/operator/jump", map[string]any{"index": 3}); code != 200 {
		t.Errorf("jump: %d", code)
	}
	if code, _ = c.do("POST", "/api/operator/jump", map[string]any{"index": 99}); code != 400 {
		t.Errorf("bad jump: %d", code)
	}
	if code, _ = c.do("POST", "/api/operator/message", map[string]any{"text": ""}); code != 400 {
		t.Errorf("empty message: %d", code)
	}
	if code, _ = c.do("POST", "/api/operator/message", map[string]any{"text": "Please **stay** seated"}); code != 200 {
		t.Errorf("message: %d", code)
	}
	if fs := s.Runtime().Engine().Snapshot().Fullscreen; fs == nil || fs.HTML != "Please <strong>stay</strong> seated" {
		t.Errorf("%+v", fs)
	}
	if code, _ = c.do("POST", "/api/operator/blank", map[string]any{"on": true}); code != 200 {
		t.Errorf("blank: %d", code)
	}
	if code, r = c.do("POST", "/api/operator/stop", nil); code != 200 || r["notice"] != "No loop is running" {
		t.Errorf("stop with no loop: %d %v", code, r)
	}
	if code, _ = c.do("POST", "/api/operator/reset", nil); code != 200 {
		t.Errorf("reset: %d", code)
	}
	if s.Runtime().Engine().Operator().Loop == nil {
		t.Error("reset should auto-start the leading loop")
	}
}

func toJSON(v any) string { b, _ := json.Marshal(v); return string(b) }

func TestOnlyOneControllerAndTakeover(t *testing.T) {
	_, ts := opServer(t, "A\nB\nC")
	a, b := newClient(t, ts.URL), newClient(t, ts.URL)
	a.login("operator", "operatorpw")
	b.login("operator", "operatorpw")

	if code, _ := a.do("POST", "/api/operator/claim", map[string]bool{"takeover": false}); code != 200 {
		t.Fatal("a should get control")
	}
	if code, r := b.do("POST", "/api/operator/claim", map[string]bool{"takeover": false}); code != 409 || r["ok"] != false {
		t.Errorf("b claim without takeover: %d %v", code, r)
	}
	if code, _ := b.do("POST", "/api/operator/next", nil); code != 409 {
		t.Errorf("view-only operator must not advance: %d", code)
	}
	if code, _ := a.do("POST", "/api/operator/next", nil); code != 200 {
		t.Errorf("controller next: %d", code)
	}
	if code, _ := b.do("POST", "/api/operator/claim", map[string]bool{"takeover": true}); code != 200 {
		t.Error("takeover failed")
	}
	if code, _ := a.do("POST", "/api/operator/next", nil); code != 409 {
		t.Errorf("old controller after takeover: %d", code)
	}
	if code, _ := b.do("POST", "/api/operator/next", nil); code != 200 {
		t.Errorf("new controller next: %d", code)
	}
}

func TestOperatorStreamPushesChanges(t *testing.T) {
	_, ts := opServer(t, "A\nB")
	c := newClient(t, ts.URL)
	c.login("operator", "operatorpw")
	c.do("POST", "/api/operator/claim", map[string]bool{"takeover": false})

	resp, err := c.hc.Get(ts.URL + "/api/operator/stream")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	events := make(chan map[string]any, 20)
	go func() {
		buf := make([]byte, 0, 4096)
		tmp := make([]byte, 1024)
		for {
			n, err := resp.Body.Read(tmp)
			buf = append(buf, tmp[:n]...)
			for {
				i := strings.Index(string(buf), "\n\n")
				if i < 0 {
					break
				}
				block := string(buf[:i])
				buf = buf[i+2:]
				for _, ln := range strings.Split(block, "\n") {
					if strings.HasPrefix(ln, "data: ") {
						var m map[string]any
						if json.Unmarshal([]byte(ln[6:]), &m) == nil && m["op"] != nil {
							events <- m
						}
					}
				}
			}
			if err != nil {
				return
			}
		}
	}()
	wait := func(pred func(map[string]any) bool) map[string]any {
		deadline := time.After(2 * time.Second)
		for {
			select {
			case m := <-events:
				if pred(m) {
					return m
				}
			case <-deadline:
				t.Fatal("timed out waiting for operator stream update")
			}
		}
	}
	wait(func(m map[string]any) bool { return true }) // initial state
	c.do("POST", "/api/operator/next", nil)
	m := wait(func(m map[string]any) bool { return m["op"].(map[string]any)["cursor"].(float64) == 1 })
	if m["controller"] == "" {
		t.Error("controller should be reported")
	}
	// An audience device connecting changes the viewer count.
	aud := connect(t, ts.URL)
	aud.next(t)
	wait(func(m map[string]any) bool { return m["audience"].(float64) == 1 })
}

func TestLiveStatePersistsAcrossRestart(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "live.json")
	cues := script.Parse("A\nB\nC", script.Options{}).Cues

	s1 := New(testConfig(), "t")
	s1.Runtime().EnablePersistence(path)
	s1.Runtime().SetShow("demo", "Demo", "", cues, DefaultTheme(""))
	s1.Runtime().Do(func(e *liveEngine) (liveOpState, error) { return e.Next(-1) })
	s1.Runtime().Do(func(e *liveEngine) (liveOpState, error) { return e.Next(-1) })
	s1.Close() // flushes

	s2 := New(testConfig(), "t")
	defer s2.Close()
	s2.Runtime().EnablePersistence(path)
	s2.Runtime().SetShow("demo", "Demo", "", cues, DefaultTheme(""))
	if !s2.Runtime().Restore() {
		t.Fatal("expected restore")
	}
	if op := s2.Runtime().Engine().Operator(); op.Cursor != 2 {
		t.Errorf("cursor %d", op.Cursor)
	}
	if sb := s2.Runtime().Engine().Snapshot().Scrollback; len(sb) != 2 {
		t.Errorf("%+v", sb)
	}

	// A different show must not pick up the saved position.
	s3 := New(testConfig(), "t")
	defer s3.Close()
	s3.Runtime().EnablePersistence(path)
	s3.Runtime().SetShow("other", "Other", "", cues, DefaultTheme(""))
	if s3.Runtime().Restore() {
		t.Error("state from another show must not be restored")
	}
}

func TestPages(t *testing.T) {
	_, ts := opServer(t, "A")
	for _, p := range []string{"/operator", "/admin", "/static/operator/operator.js", "/static/admin/admin.js"} {
		resp, _ := http.Get(ts.URL + p)
		resp.Body.Close()
		if resp.StatusCode != 200 {
			t.Errorf("%s: %d", p, resp.StatusCode)
		}
	}
	resp, _ := http.Get(ts.URL + "/operator")
	resp.Body.Close()
	if !strings.Contains(resp.Header.Get("Content-Security-Policy"), "default-src 'self'") || resp.Header.Get("Cache-Control") != "no-store" {
		t.Errorf("headers: %v", resp.Header)
	}
}

// Short aliases so tests can call Runtime.Do without importing live twice.
type (
	liveEngine  = live.Engine
	liveOpState = live.OperatorState
)
