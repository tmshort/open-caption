// Copyright the open-caption authors
// SPDX-License-Identifier: Apache-2.0

package httpapi

import (
	"bufio"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/tmshort/open-caption/internal/config"
	"github.com/tmshort/open-caption/internal/showfile"
)

type sseEvent struct {
	ID, Type string
	Data     map[string]any
}

type sseReader struct {
	resp *http.Response
	ch   chan sseEvent
}

func connect(t *testing.T, url string) *sseReader {
	t.Helper()
	resp, err := http.Get(url + "/events")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != 200 {
		t.Fatalf("status %d", resp.StatusCode)
	}
	r := &sseReader{resp: resp, ch: make(chan sseEvent, 100)}
	go readSSE(resp.Body, r.ch)
	t.Cleanup(func() { resp.Body.Close() })
	return r
}

func (r *sseReader) next(t *testing.T) sseEvent {
	t.Helper()
	select {
	case ev, ok := <-r.ch:
		if !ok {
			t.Fatal("stream closed")
		}
		return ev
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for event")
	}
	return sseEvent{}
}

// nextEvent returns the next sequenced event, skipping preload hints (their timing is not part of the contract).
func (r *sseReader) nextEvent(t *testing.T) sseEvent {
	t.Helper()
	for {
		if ev := r.next(t); ev.Type != "preload" {
			return ev
		}
	}
}

func devServer(t *testing.T, src string, maxClients int) (*Server, *httptest.Server, string) {
	t.Helper()
	cfg := config.Default()
	cfg.Audience.MaxClients = maxClients
	s := New(cfg, "test")
	t.Cleanup(s.Close)
	s.EnableDev()
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "curtain.jpg"), []byte("\xff\xd8\xffjpeg"), 0o644)
	os.WriteFile(filepath.Join(dir, "notes.txt"), []byte("secret"), 0o644)
	s.Runtime().SetShow("demo", "Demo", dir, showfile.Parse(src, showfile.Options{}).Cues, DefaultTheme("Soon"))
	ts := httptest.NewServer(s.Handler())
	t.Cleanup(ts.Close)
	return s, ts, dir
}

func post(t *testing.T, url string) {
	t.Helper()
	resp, err := http.Post(url, "", nil)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
}

func TestSnapshotThenCaptions(t *testing.T) {
	_, ts, _ := devServer(t, "HAMLET: To be\nOr **not**\n[image: curtain.jpg]\nEnd", 0)
	c := connect(t, ts.URL)
	snap := c.next(t)
	if snap.Type != "snapshot" || snap.Data["mode"] != "idle" || snap.Data["theme"].(map[string]any)["idleMessage"] != "Soon" {
		t.Fatalf("%+v", snap)
	}
	post(t, ts.URL+"/dev/next")
	ev := c.nextEvent(t)
	item := ev.Data["item"].(map[string]any)
	if ev.Type != "caption" || item["speaker"] != "HAMLET" || item["html"] != "To be" {
		t.Fatalf("%+v", ev)
	}
	post(t, ts.URL+"/dev/next")
	post(t, ts.URL+"/dev/next") // image
	if ev = c.nextEvent(t); ev.Data["item"].(map[string]any)["html"] != "Or <strong>not</strong>" {
		t.Fatalf("%+v", ev)
	}
	ev = c.nextEvent(t)
	fs := ev.Data["item"].(map[string]any)
	if ev.Type != "fullscreen" || fs["kind"] != "image" || fs["src"] != "/assets/demo/curtain.jpg" {
		t.Fatalf("%+v", ev)
	}
}

func TestLateJoinerSeesScrollbackAndPreload(t *testing.T) {
	_, ts, _ := devServer(t, "A\nB\n[image: curtain.jpg]\nC", 0)
	post(t, ts.URL+"/dev/next")
	post(t, ts.URL+"/dev/next")
	c := connect(t, ts.URL)
	snap := c.next(t)
	sb := snap.Data["scrollback"].([]any)
	pre := snap.Data["preload"].([]any)
	if len(sb) != 2 || len(pre) != 1 || pre[0] != "/assets/demo/curtain.jpg" {
		t.Fatalf("%+v", snap.Data)
	}
}

func TestBackSendsSnapshot(t *testing.T) {
	_, ts, _ := devServer(t, "A\nB", 0)
	c := connect(t, ts.URL)
	c.next(t)
	post(t, ts.URL+"/dev/next")
	post(t, ts.URL+"/dev/next")
	c.nextEvent(t)
	c.nextEvent(t)
	post(t, ts.URL+"/dev/back")
	ev := c.nextEvent(t)
	if ev.Type != "snapshot" || len(ev.Data["scrollback"].([]any)) != 1 {
		t.Fatalf("%+v", ev)
	}
}

func TestLoopTicksReachClient(t *testing.T) {
	_, ts, _ := devServer(t, "[loop: 0.05s]\nOne\nTwo\n[/loop]\nA", 0)
	c := connect(t, ts.URL)
	snap := c.next(t)
	if snap.Data["fullscreen"].(map[string]any)["html"] != "One" {
		t.Fatalf("%+v", snap.Data)
	}
	ev := c.next(t)
	if ev.Type != "fullscreen" || ev.Data["item"].(map[string]any)["html"] != "Two" {
		t.Fatalf("%+v", ev)
	}
}

func TestMaxClients503(t *testing.T) {
	_, ts, _ := devServer(t, "A", 1)
	connect(t, ts.URL).next(t)
	resp, err := http.Get(ts.URL + "/events")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 503 || resp.Header.Get("Retry-After") == "" {
		t.Errorf("status %d", resp.StatusCode)
	}
}

func TestAssetServing(t *testing.T) {
	_, ts, _ := devServer(t, "A", 0)
	cases := []struct {
		path string
		code int
	}{
		{"/assets/demo/curtain.jpg", 200},
		{"/assets/demo/notes.txt", 404},       // not an image type
		{"/assets/demo/..%2Fsecret.jpg", 404}, // traversal
		{"/assets/other/curtain.jpg", 404},    // not the active show
		{"/assets/demo/missing.png", 404},
	}
	for _, c := range cases {
		resp, err := http.Get(ts.URL + c.path)
		if err != nil {
			t.Fatal(err)
		}
		resp.Body.Close()
		if resp.StatusCode != c.code {
			t.Errorf("%s: %d, want %d", c.path, resp.StatusCode, c.code)
		}
		if c.code == 200 && resp.Header.Get("X-Content-Type-Options") != "nosniff" {
			t.Error("missing nosniff")
		}
	}
}

func TestAudiencePageAndStatic(t *testing.T) {
	_, ts, _ := devServer(t, "A", 0)
	resp, _ := http.Get(ts.URL + "/")
	if resp.StatusCode != 200 || !strings.Contains(resp.Header.Get("Content-Security-Policy"), "default-src 'self'") {
		t.Errorf("page %d %v", resp.StatusCode, resp.Header)
	}
	resp.Body.Close()
	for _, p := range []string{"/static/audience/audience.js", "/static/audience/audience.css"} {
		resp, _ = http.Get(ts.URL + p)
		if resp.StatusCode != 200 {
			t.Errorf("%s: %d", p, resp.StatusCode)
		}
		resp.Body.Close()
	}
	resp, _ = http.Get(ts.URL + "/nope")
	if resp.StatusCode != 404 {
		t.Errorf("unknown path: %d", resp.StatusCode)
	}
	resp.Body.Close()
}

func TestDevEndpointsOffByDefault(t *testing.T) {
	s := New(config.Default(), "t")
	ts := httptest.NewServer(s.Handler())
	defer ts.Close()
	resp, _ := http.Post(ts.URL+"/dev/next", "", nil)
	if resp.StatusCode == 200 {
		t.Error("dev endpoint must not exist unless enabled")
	}
	resp.Body.Close()
}

func readSSE(body io.Reader, out chan sseEvent) {
	defer close(out)
	sc := bufio.NewScanner(body)
	var ev sseEvent
	for sc.Scan() {
		line := sc.Text()
		switch {
		case line == "":
			if ev.Type != "" {
				out <- ev
			}
			ev = sseEvent{}
		case strings.HasPrefix(line, "id: "):
			ev.ID = line[4:]
		case strings.HasPrefix(line, "event: "):
			ev.Type = line[7:]
		case strings.HasPrefix(line, "data: "):
			json.Unmarshal([]byte(line[6:]), &ev.Data)
		}
	}
}
