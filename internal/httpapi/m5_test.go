package httpapi

import (
	"archive/zip"
	"bytes"
	"io"
	"net/http"
	"regexp"
	"strings"
	"testing"

	"github.com/tmshort/open-caption/internal/logbuf"
)

func TestHelpNeedsASessionAndServesTopics(t *testing.T) {
	_, ts, admin := adminServer(t)
	anon := newClient(t, ts.URL)
	for _, p := range []string{"/api/help/topics", "/help/admin-images", "/help/sample-script.md", "/api/help/search?q=image"} {
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
	resp, _ = op.hc.Get(ts.URL + "/help/sample-script.md")
	body, _ = io.ReadAll(resp.Body)
	resp.Body.Close()
	if resp.StatusCode != 200 || !strings.Contains(string(body), "[loop") {
		t.Errorf("sample script download: %d", resp.StatusCode)
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
