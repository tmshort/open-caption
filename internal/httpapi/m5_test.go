package httpapi

import (
	"archive/zip"
	"bytes"
	"github.com/tmshort/open-caption/internal/config"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/tmshort/open-caption/internal/logbuf"
)

func TestHelpNeedsASessionAndServesTopics(t *testing.T) {
	_, ts, admin := adminServer(t)
	anon := newClient(t, ts.URL)
	for _, p := range []string{"/api/help/topics", "/help/admin-images", "/help/sample-show.md", "/api/help/search?q=image"} {
		resp, _ := anon.hc.Get(ts.URL + p)
		resp.Body.Close()
		if resp.StatusCode != 401 {
			t.Errorf("anonymous %s: %d", p, resp.StatusCode)
		}
	}
	// The viewer page itself is a shell with no content; the audience must not get operator docs.
	if resp, _ := anon.hc.Get(ts.URL + "/help"); resp.StatusCode != 200 {
		t.Errorf("shell: %d", resp.StatusCode)
	}

	op := newClient(t, ts.URL)
	op.login("operator", "operatorpw")
	for _, c := range []*client{admin, op} {
		code, r := c.do("GET", "/api/help/topics", nil)
		if code != 200 || len(r["topics"].([]any)) < 10 {
			t.Fatalf("topics: %d %v", code, r)
		}
	}
	resp, _ := op.hc.Get(ts.URL + "/help/operator-keyboard")
	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	if resp.StatusCode != 200 || !strings.Contains(string(body), "<table>") || !strings.Contains(string(body), "Shift+Q") ||
		!strings.HasPrefix(resp.Header.Get("Content-Type"), "text/html") {
		t.Errorf("topic: %d %s", resp.StatusCode, body)
	}
	resp, _ = op.hc.Get(ts.URL + "/help/operator-keyboard?format=md")
	body, _ = io.ReadAll(resp.Body)
	resp.Body.Close()
	if !strings.HasPrefix(string(body), "# Keyboard") {
		t.Errorf("markdown: %s", body[:20])
	}
	if resp, _ = op.hc.Get(ts.URL + "/help/no-such-topic"); resp.StatusCode != 404 {
		t.Errorf("unknown topic: %d", resp.StatusCode)
	}
	resp, _ = op.hc.Get(ts.URL + "/help/sample-show.md")
	body, _ = io.ReadAll(resp.Body)
	resp.Body.Close()
	if resp.StatusCode != 200 || !strings.Contains(string(body), "[loop") {
		t.Errorf("sample show file download: %d", resp.StatusCode)
	}
	_, r := op.do("GET", "/api/help/search?q=spacebar", nil)
	if len(r["hits"].([]any)) == 0 {
		t.Error("search found nothing for spacebar")
	}
}

func TestCaptiveAPI(t *testing.T) {
	_, ts, _ := adminServer(t)
	resp, err := http.Get(ts.URL + "/captive-api")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	if resp.Header.Get("Content-Type") != "application/captive+json" ||
		!strings.Contains(string(b), `"captive":false`) || !strings.Contains(string(b), `"user-portal-url":"http://`+strings.TrimPrefix(ts.URL, "http://")+`/"`) {
		t.Errorf("%s %s", resp.Header.Get("Content-Type"), b)
	}
}

func TestSupportBundleHasNoPasswords(t *testing.T) {
	s, ts, c := adminServer(t)
	lb := logbuf.New(50)
	lb.Write([]byte("2026/01/01 started\n"))
	s.SetLogSource(lb)
	createShow(t, c, "Bundle show")

	resp, err := c.hc.Get(ts.URL + "/api/admin/support-bundle")
	if err != nil {
		t.Fatal(err)
	}
	data, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	if resp.StatusCode != 200 {
		t.Fatalf("status %d", resp.StatusCode)
	}
	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatal(err)
	}
	files := map[string]string{}
	for _, f := range zr.File {
		rc, _ := f.Open()
		b, _ := io.ReadAll(rc)
		rc.Close()
		files[f.Name] = string(b)
	}
	for _, name := range []string{"info.txt", "config.redacted.yaml", "log.txt", "shows.txt"} {
		if _, ok := files[name]; !ok {
			t.Errorf("bundle is missing %s", name)
		}
	}
	for name, content := range files {
		if strings.Contains(content, "adminpw") || strings.Contains(content, "operatorpw") {
			t.Errorf("%s leaks a password", name)
		}
	}
	if !strings.Contains(files["log.txt"], "started") || !strings.Contains(files["shows.txt"], "Bundle show") || !strings.Contains(files["info.txt"], "open-caption test") {
		t.Errorf("bundle contents: %v", files)
	}
	// Admin only.
	op := newClient(t, ts.URL)
	op.login("operator", "operatorpw")
	if code, _ := op.do("GET", "/api/admin/support-bundle", nil); code != 403 {
		t.Errorf("operator: %d", code)
	}
}

func TestSnapshotCarriesSpeakerOrderForColors(t *testing.T) {
	_, ts, _ := devServer(t, "MARGARET: hi\nEDWIN: hello\nMARGARET: again", 0)
	c := connect(t, ts.URL)
	snap := c.next(t)
	sp, _ := snap.Data["speakers"].([]any)
	if len(sp) != 2 || sp[0] != "MARGARET" || sp[1] != "EDWIN" {
		t.Errorf("speakers: %v", snap.Data["speakers"])
	}
	resp, _ := http.Get(ts.URL + "/static/shared/speakers.js")
	resp.Body.Close()
	if resp.StatusCode != 200 {
		t.Errorf("shared speakers.js: %d", resp.StatusCode)
	}
}

// Every font the audience page's stylesheet asks for must actually be served (the program works with no internet).
func TestBundledFontsAreServed(t *testing.T) {
	_, ts, c := adminServer(t)
	css, _ := http.Get(ts.URL + "/static/audience/audience.css")
	body, _ := io.ReadAll(css.Body)
	css.Body.Close()
	urls := regexp.MustCompile(`url\((/[^)]+\.woff2)\)`).FindAllStringSubmatch(string(body), -1)
	if len(urls) < 8 {
		t.Fatalf("expected the OpenDyslexic and Atkinson @font-face rules, found %d", len(urls))
	}
	for _, m := range urls {
		resp, err := http.Get(ts.URL + m[1])
		if err != nil {
			t.Fatal(err)
		}
		b, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		if resp.StatusCode != 200 || len(b) < 10000 || string(b[:4]) != "wOF2" ||
			resp.Header.Get("Content-Type") != "font/woff2" || resp.Header.Get("X-Content-Type-Options") != "nosniff" ||
			!strings.Contains(resp.Header.Get("Cache-Control"), "max-age") {
			t.Errorf("%s: %d %s %v (%d bytes)", m[1], resp.StatusCode, resp.Header.Get("Content-Type"), resp.Header, len(b))
		}
	}
	// The font can be chosen for a show and reaches audiences in the theme.
	id := createShow(t, c, "Fonts")
	if code, _ := c.do("PUT", "/api/admin/shows/"+id, map[string]any{"theme": map[string]any{"font": "dyslexic"}}); code != 200 {
		t.Errorf("dyslexic font: %d", code)
	}
	if code, _ := c.do("PUT", "/api/admin/shows/"+id, map[string]any{"theme": map[string]any{"font": "comic-sans"}}); code != 400 {
		t.Errorf("unknown font: %d", code)
	}
	// The license travels with the font.
	for _, lic := range []string{"OFL-OpenDyslexic.txt", "OFL-Atkinson.txt"} {
		if resp, _ := http.Get(ts.URL + "/static/shared/fonts/" + lic); resp.StatusCode != 200 {
			t.Errorf("%s: %d", lic, resp.StatusCode)
		}
	}
	if code, _ := c.do("PUT", "/api/admin/shows/"+id, map[string]any{"theme": map[string]any{"font": "hyperlegible"}}); code != 200 {
		t.Errorf("hyperlegible font: %d", code)
	}
}

func TestSweepDropsRehearsalsOfExpiredSessions(t *testing.T) {
	s, _, c := adminServer(t)
	id := createShow(t, c, "Janitor")
	if code, _ := c.do("POST", "/api/admin/rehearsal/start", map[string]string{"showId": id}); code != 200 {
		t.Fatal("start rehearsal")
	}
	s.rmu.Lock()
	n := len(s.rehearsals)
	s.rmu.Unlock()
	if n != 1 {
		t.Fatalf("rehearsals: %d", n)
	}
	s.sweep() // still signed in: kept
	s.rmu.Lock()
	n = len(s.rehearsals)
	s.rmu.Unlock()
	if n != 1 {
		t.Errorf("a live session's rehearsal was dropped")
	}
	s.auth.mu.Lock()
	for k, sess := range s.auth.sessions { // simulate the login expiring
		sess.expires = time.Now().Add(-time.Minute)
		s.auth.sessions[k] = sess
	}
	s.auth.mu.Unlock()
	s.sweep()
	s.rmu.Lock()
	n = len(s.rehearsals)
	s.rmu.Unlock()
	if n != 0 {
		t.Errorf("rehearsal of an expired session should be dropped, have %d", n)
	}
}

func TestLookSavedIsRememberedOnTheServer(t *testing.T) {
	_, _, c := adminServer(t)
	id := createShow(t, c, "Look")
	_, r := c.do("GET", "/api/admin/shows/"+id, nil)
	if r["lookSaved"] != false {
		t.Errorf("new show: %v", r["lookSaved"])
	}
	c.do("PUT", "/api/admin/shows/"+id, map[string]any{"name": "Renamed"})
	if _, r = c.do("GET", "/api/admin/shows/"+id, nil); r["lookSaved"] != false {
		t.Error("renaming must not count as saving the look")
	}
	c.do("PUT", "/api/admin/shows/"+id, map[string]any{"theme": map[string]any{"scalePct": 120}})
	if _, r = c.do("GET", "/api/admin/shows/"+id, nil); r["lookSaved"] != true {
		t.Error("saving the look should be remembered")
	}
}

// ---- access list and forged-request protection ----

func accessServer(t *testing.T, admin, operator []string) (*Server, *httptest.Server) {
	t.Helper()
	cfg := testConfig()
	cfg.DataDir = t.TempDir()
	cfg.Access.AdminAllow, cfg.Access.OperatorAllow = admin, operator
	s := New(cfg, "t")
	t.Cleanup(s.Close)
	ts := httptest.NewServer(s.Handler()) // requests arrive from 127.0.0.1
	t.Cleanup(ts.Close)
	return s, ts
}

func status(t *testing.T, method, url string, hdr map[string]string, body string) int {
	t.Helper()
	req, _ := http.NewRequest(method, url, strings.NewReader(body))
	for k, v := range hdr {
		req.Header.Set(k, v)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	return resp.StatusCode
}

func TestUnrestrictedByDefault(t *testing.T) {
	_, ts := accessServer(t, nil, nil)
	for _, p := range []string{"/admin", "/operator", "/static/admin/admin.js", "/static/operator/operator.js"} {
		if code := status(t, "GET", ts.URL+p, nil, ""); code != 200 {
			t.Errorf("%s: %d", p, code)
		}
	}
}

func TestSetupLimitedToTheListedComputers(t *testing.T) {
	// Setup only from "elsewhere": this test computer (127.0.0.1) is not on the list.
	_, ts := accessServer(t, []string{"10.0.0.0/8"}, nil)
	for _, p := range []string{"/admin", "/static/admin/admin.js"} {
		if code := status(t, "GET", ts.URL+p, nil, ""); code != 403 {
			t.Errorf("%s from an unlisted computer: %d", p, code)
		}
	}
	// Login for the admin role is refused; the operator role is unaffected (its list is empty).
	if code := status(t, "POST", ts.URL+"/api/login", nil, `{"role":"admin","password":"adminpw"}`); code != 403 {
		t.Errorf("admin login: %d", code)
	}
	if code := status(t, "POST", ts.URL+"/api/login", nil, `{"role":"operator","password":"operatorpw"}`); code != 200 {
		t.Errorf("operator login: %d", code)
	}
	if code := status(t, "GET", ts.URL+"/operator", nil, ""); code != 200 {
		t.Errorf("operator page: %d", code)
	}
	// The audience page is never limited.
	for _, p := range []string{"/", "/healthz", "/static/audience/audience.js"} {
		if code := status(t, "GET", ts.URL+p, nil, ""); code != 200 {
			t.Errorf("%s: %d", p, code)
		}
	}
	// An unusable setting denies everything instead of allowing everyone.
	_, ts2 := accessServer(t, []string{"not an address"}, nil)
	if code := status(t, "GET", ts2.URL+"/admin", nil, ""); code != 403 {
		t.Errorf("bad list should deny: %d", code)
	}
}

func TestAListedComputerCanUseSetupAndTheOperatorPage(t *testing.T) {
	_, ts := accessServer(t, []string{"localhost"}, []string{"10.0.0.0/8"})
	c := newClient(t, ts.URL)
	if c.login("admin", "adminpw") != 200 {
		t.Fatal("admin login from the listed computer")
	}
	if code, _ := c.do("GET", "/api/admin/ping", nil); code != 200 {
		t.Errorf("admin api: %d", code)
	}
	// This address is not in operator_allow but is in admin_allow, so the admin may still run the show.
	if code, _ := c.do("GET", "/api/operator/state", nil); code != 200 {
		t.Errorf("an admin address may use the operator pages: %d", code)
	}
	if code := status(t, "GET", ts.URL+"/operator", nil, ""); code != 200 {
		t.Errorf("operator page from an admin address: %d", code)
	}
}

func TestOperatorListAloneKeepsSetupOpenOnlyToAdminAddresses(t *testing.T) {
	_, ts := accessServer(t, nil, []string{"10.0.0.0/8"})
	// Operator pages: refused for this computer (not on operator_allow, and admin_allow is unrestricted so it IS allowed).
	if code := status(t, "GET", ts.URL+"/operator", nil, ""); code != 200 {
		t.Errorf("unrestricted admin list lets this computer in: %d", code)
	}
	_, ts2 := accessServer(t, []string{"10.0.0.0/8"}, []string{"10.0.0.0/8"})
	if code := status(t, "GET", ts2.URL+"/operator", nil, ""); code != 403 {
		t.Errorf("neither list allows this computer: %d", code)
	}
	c := newClient(t, ts2.URL)
	if code := c.login("operator", "operatorpw"); code != 403 {
		t.Errorf("operator login: %d", code)
	}
}

func TestSessionCookieDoesNotWorkFromAnotherComputer(t *testing.T) {
	s, ts := accessServer(t, nil, nil)
	c := newClient(t, ts.URL)
	c.login("admin", "adminpw")
	if code, _ := c.do("GET", "/api/admin/ping", nil); code != 200 {
		t.Fatal("setup")
	}
	// The list is tightened (as if the cookie had been taken to another computer): the valid cookie no longer works.
	s.adminAllow, _ = config.ParseAllowList([]string{"10.0.0.0/8"})
	if code, _ := c.do("GET", "/api/admin/ping", nil); code != 403 {
		t.Errorf("a stolen cookie must not work from an unlisted computer: %d", code)
	}
}

func TestForgedRequestsAreRefused(t *testing.T) {
	_, ts := accessServer(t, nil, nil)
	c := newClient(t, ts.URL)
	c.login("admin", "adminpw")
	host := strings.TrimPrefix(ts.URL, "http://")
	cookie := ""
	for _, ck := range c.hc.Jar.Cookies(mustURL(ts.URL)) {
		cookie = ck.String()
	}
	if cookie == "" {
		t.Fatal("no session cookie")
	}
	post := func(h map[string]string) int {
		h["Cookie"] = cookie
		return status(t, "POST", ts.URL+"/api/admin/shows", h, `{"name":"x"}`)
	}
	cases := []struct {
		name string
		hdr  map[string]string
		want int
	}{
		{"no origin (command line)", map[string]string{}, 201},
		{"same origin", map[string]string{"Origin": "http://" + host, "Sec-Fetch-Site": "same-origin"}, 201},
		{"another website", map[string]string{"Origin": "http://evil.example"}, 403},
		{"another port on this computer", map[string]string{"Origin": "http://" + strings.Split(host, ":")[0] + ":9999"}, 403},
		{"browser says cross-site", map[string]string{"Sec-Fetch-Site": "cross-site"}, 403},
		{"browser says same-site", map[string]string{"Sec-Fetch-Site": "same-site"}, 403},
		{"opaque origin", map[string]string{"Origin": "null"}, 403},
		{"user typed it (none)", map[string]string{"Sec-Fetch-Site": "none"}, 201},
	}
	for _, tc := range cases {
		if got := post(tc.hdr); got != tc.want {
			t.Errorf("%s: %d, want %d", tc.name, got, tc.want)
		}
	}
	// Reads and the open audience stream are not affected.
	if code := status(t, "GET", ts.URL+"/healthz", map[string]string{"Origin": "http://evil.example"}, ""); code != 200 {
		t.Errorf("GET: %d", code)
	}
	// Login itself is also protected.
	if code := status(t, "POST", ts.URL+"/api/login", map[string]string{"Origin": "http://evil.example"}, `{"role":"admin","password":"adminpw"}`); code != 403 {
		t.Errorf("forged login: %d", code)
	}
}

func mustURL(s string) *url.URL { u, _ := url.Parse(s); return u }

func TestSettingsReportTheAccessRules(t *testing.T) {
	_, ts := accessServer(t, []string{"localhost"}, nil)
	c := newClient(t, ts.URL)
	c.login("admin", "adminpw")
	_, r := c.do("GET", "/api/admin/settings", nil)
	if r["adminRestricted"] != true || r["operatorRestricted"] != false || toJSON(r["adminAllow"]) != `["localhost"]` {
		t.Errorf("%v", r)
	}
}

func TestEveryPageLinksTheFavicon(t *testing.T) {
	_, ts, _ := adminServer(t)
	resp, _ := http.Get(ts.URL + "/static/shared/favicon.svg")
	if resp.StatusCode != 200 || !strings.HasPrefix(resp.Header.Get("Content-Type"), "image/svg+xml") {
		t.Fatalf("favicon: %d %s", resp.StatusCode, resp.Header.Get("Content-Type"))
	}
	for _, p := range []string{"/", "/admin", "/operator"} {
		r, _ := http.Get(ts.URL + p)
		b, _ := io.ReadAll(r.Body)
		r.Body.Close()
		if !strings.Contains(string(b), `href="/static/shared/favicon.svg"`) {
			t.Errorf("%s does not link the favicon", p)
		}
	}
}
