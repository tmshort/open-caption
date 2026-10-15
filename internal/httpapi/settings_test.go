package httpapi

import (
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/tmshort/open-caption/internal/config"
)

func serverWith(t *testing.T, mut func(*config.Config)) (*Server, *httptest.Server, *client) {
	t.Helper()
	cfg := testConfig()
	cfg.DataDir = t.TempDir()
	if mut != nil {
		mut(&cfg)
	}
	s := New(cfg, "test")
	t.Cleanup(s.Close)
	ts := httptest.NewServer(s.Handler())
	t.Cleanup(ts.Close)
	c := newClient(t, ts.URL)
	if c.login("admin", "adminpw") != 200 {
		t.Fatal("admin login")
	}
	return s, ts, c
}

func putFile(t *testing.T, c *client, id, lang, text string) map[string]any {
	t.Helper()
	q := ""
	if lang != "" {
		q = "?lang=" + lang
	}
	code, r := c.do("PUT", "/api/admin/shows/"+id+"/showfile"+q, map[string]string{"text": text})
	if code != 200 {
		t.Fatalf("save %s: %d %v", lang, code, r)
	}
	return r
}

func theme(t *testing.T, ev sseEvent) map[string]any {
	t.Helper()
	th, _ := ev.Data["theme"].(map[string]any)
	if th == nil {
		t.Fatalf("no theme in %+v", ev)
	}
	return th
}

func TestTheShowFileSetsTheLookAndTheTitle(t *testing.T) {
	s, ts, c := serverWith(t, nil)
	id := createShow(t, c, "Draft")
	putFile(t, c, id, "", "[settings]\ntitle: Layered Show\nfont: dyslexic\ntext size: 150%\nspeaker names: color\nbackground: #101010\nidle message: See you soon\n[/settings]\nHello\nWorld")
	_, r := c.do("GET", "/api/admin/shows/"+id, nil)
	th := r["theme"].(map[string]any)
	if r["name"] != "Layered Show" || th["font"] != "dyslexic" || th["scalePct"].(float64) != 150 || th["bg"] != "#101010" || r["lookSaved"] != true {
		t.Fatalf("the title and look come from the show file: %v", r)
	}
	if th["fg"] != "#ffffff" || th["scrollbackLines"].(float64) != 200 {
		t.Errorf("anything the show file leaves out uses the defaults: %v", th)
	}
	c.do("POST", "/api/admin/shows/"+id+"/activate", nil)
	aud := connect(t, ts.URL)
	snap := aud.next(t)
	at := theme(t, snap)
	if at["font"] != "dyslexic" || at["scalePct"].(float64) != 150 || at["idleMessage"] != "See you soon" || at["speakers"] != "color" {
		t.Errorf("audience theme: %v", at)
	}
	if p := s.rt.OperatorPayload(); p.Show.Name != "Layered Show" {
		t.Errorf("operator sees the title: %+v", p.Show)
	}

	// Editing the show file while it is on air changes the look for everyone straight away.
	_, cur := c.do("GET", "/api/admin/shows/"+id+"/showfile", nil)
	text := cur["text"].(string)
	putFile(t, c, id, "", strings.Replace(strings.Replace(text, "font: dyslexic", "font: serif", 1), "title: Layered Show", "title: Renamed Live", 1))
	ev := aud.nextEvent(t)
	for ev.Type != "snapshot" || theme(t, ev)["font"] != "serif" {
		ev = aud.nextEvent(t)
	}
	if p := s.rt.OperatorPayload(); p.Show.Name != "Renamed Live" {
		t.Errorf("a changed title reaches the operator: %+v", p.Show)
	}
}

func TestConfigDefaultsThenShowFile(t *testing.T) {
	_, _, c := serverWith(t, func(cfg *config.Config) {
		cfg.Defaults.Look = config.Look{Background: "#112233", TextColor: "#eeeeee", SpeakerNameColor: "#ffcc00", Font: "serif", TextSizePercent: 120, SpeakerNames: "color"}
		cfg.Defaults.IdleMessage = "Configured idle"
		cfg.Defaults.LoopItemSeconds = 7
		cfg.Audience.ScrollbackLines = 300
	})
	plain := createShow(t, c, "Plain")
	_, r := c.do("GET", "/api/admin/shows/"+plain, nil)
	th := r["theme"].(map[string]any)
	if th["bg"] != "#112233" || th["font"] != "serif" || th["scalePct"].(float64) != 120 || th["idleMessage"] != "Configured idle" ||
		th["loopSeconds"].(float64) != 7 || th["scrollbackLines"].(float64) != 300 || r["lookSaved"] != false {
		t.Errorf("a show that says nothing uses config.yaml's defaults: %v", th)
	}
	putFile(t, c, plain, "", "[settings]\ntitle: Plain\nfont: mono\nloop seconds: 3\n[/settings]\nHi")
	_, r = c.do("GET", "/api/admin/shows/"+plain, nil)
	th = r["theme"].(map[string]any)
	if th["font"] != "mono" || th["loopSeconds"].(float64) != 3 || th["bg"] != "#112233" || th["scrollbackLines"].(float64) != 300 {
		t.Errorf("the show file overrides only what it sets: %v", th)
	}
}

func TestLoopTimeAndScrollbackComeFromTheShowFile(t *testing.T) {
	_, ts, c := serverWith(t, nil)
	id := createShow(t, c, "Timing")
	var lines []string
	for i := 1; i <= 30; i++ {
		lines = append(lines, "Line "+strings.Repeat("x", i%3+1))
	}
	putFile(t, c, id, "", "[settings]\ntitle: Timing\nloop seconds: 0.05\nscrollback lines: 20\n[/settings]\n[loop]\nOne\nTwo\n[/loop]\n"+strings.Join(lines, "\n"))
	c.do("POST", "/api/admin/shows/"+id+"/activate", nil)
	aud := connect(t, ts.URL)
	aud.next(t)
	// The pre-show loop steps on after 50 ms (the file's loop seconds), not the 10 s default.
	deadline := time.After(2 * time.Second)
	for stepped := false; !stepped; {
		select {
		case ev := <-aud.ch:
			stepped = ev.Type == "fullscreen"
		case <-deadline:
			t.Fatal("the loop did not advance on the show file's time")
		}
	}
	op := newClient(t, ts.URL)
	op.login("operator", "operatorpw")
	op.do("POST", "/api/operator/claim", map[string]bool{"takeover": true})
	for i := 0; i < 26; i++ {
		op.do("POST", "/api/operator/next", nil)
	}
	late := connect(t, ts.URL)
	if n := len(late.next(t).Data["scrollback"].([]any)); n != 20 {
		t.Errorf("scrollback lines from the show file: %d", n)
	}
}

func TestLookStepWritesTheShowFileAndTranslationsKeepTheirOwnIdleMessage(t *testing.T) {
	_, _, c := serverWith(t, nil)
	id := createShow(t, c, "Look")
	putFile(t, c, id, "", "[language: en | English]\n[settings]\ntitle: Look\n[/settings]\nHello")
	c.do("POST", "/api/admin/shows/"+id+"/languages", map[string]any{"code": "es", "copyFromMain": true})
	code, r := c.do("PUT", "/api/admin/shows/"+id, map[string]any{"theme": map[string]any{
		"bg": "#000011", "fg": "#ffffcc", "font": "hyperlegible", "scalePct": 130, "speakers": "color",
		"idleMessage": "Back in a moment", "idleMessages": map[string]string{"es": "Un momento"}, "loopSeconds": 8, "scrollbackLines": 100}})
	if code != 200 {
		t.Fatalf("%d %v", code, r)
	}
	_, f := c.do("GET", "/api/admin/shows/"+id+"/showfile", nil)
	main := f["text"].(string)
	for _, want := range []string{"[language: en | English]", "[settings]", "title: Look", "background: #000011", "text color: #ffffcc", "font: hyperlegible", "text size: 130%", "speaker names: color", "idle message: Back in a moment", "loop seconds: 8", "scrollback lines: 100", "Hello"} {
		if !strings.Contains(main, want) {
			t.Errorf("the main show file should say %q:\n%s", want, main)
		}
	}
	_, f = c.do("GET", "/api/admin/shows/"+id+"/showfile?lang=es", nil)
	es := f["text"].(string)
	if !strings.Contains(es, "idle message: Un momento") || strings.Contains(es, "background") || strings.Contains(es, "font:") {
		t.Errorf("a translation holds only its own idle message:\n%s", es)
	}
	_, r = c.do("GET", "/api/admin/shows/"+id, nil)
	th := r["theme"].(map[string]any)
	if th["idleMessages"].(map[string]any)["es"] != "Un momento" || th["font"] != "hyperlegible" {
		t.Errorf("%v", th)
	}
	// An invalid look is refused and the file is untouched.
	if code, _ := c.do("PUT", "/api/admin/shows/"+id, map[string]any{"theme": map[string]any{"font": "comic"}}); code != 400 {
		t.Errorf("bad font: %d", code)
	}
	if _, f = c.do("GET", "/api/admin/shows/"+id+"/showfile", nil); f["text"] != main {
		t.Error("a refused look must not change the file")
	}
}

func TestSettingProblemsAreShownWithLineNumbers(t *testing.T) {
	_, _, c := serverWith(t, nil)
	id := createShow(t, c, "Problems")
	r := putFile(t, c, id, "", "[settings]\ntitle: Problems\nfont: comic sans\ncolour: red\n[/settings]\nHello")
	d := toJSON(r["analysis"].(map[string]any)["diagnostics"])
	for _, want := range []string{`"line":3`, "is not a font", `Unknown setting`, `"line":4`} {
		if !strings.Contains(d, want) {
			t.Errorf("missing %q in %s", want, d)
		}
	}
	c.do("POST", "/api/admin/shows/"+id+"/languages", map[string]any{"code": "es"})
	r = putFile(t, c, id, "es", "[language: es | Español]\n[settings]\nfont: serif\nidle message: Pronto\n[/settings]\nHola")
	d = toJSON(r["analysis"].(map[string]any)["diagnostics"])
	if !strings.Contains(d, "set once for the whole show") || strings.Contains(d, "Pronto") {
		t.Errorf("a translation may only set the idle message: %s", d)
	}
	// A file with no title is flagged (it would be listed as "Untitled show").
	r = putFile(t, c, id, "", "Hello")
	if !strings.Contains(toJSON(r["analysis"].(map[string]any)["diagnostics"]), "no title") {
		t.Errorf("%v", r["analysis"])
	}
	if _, g := c.do("GET", "/api/admin/shows/"+id, nil); g["name"] != "Untitled show" {
		t.Errorf("%v", g["name"])
	}
}

func TestShowsSavedInTheOlderLayoutAreConvertedWhenTheServerStarts(t *testing.T) {
	cfg := testConfig()
	cfg.DataDir = t.TempDir()
	sd := filepath.Join(cfg.DataDir, "shows", "garden-1234")
	os.MkdirAll(filepath.Join(sd, "assets"), 0o755)
	os.WriteFile(filepath.Join(sd, "show.yaml"), []byte("name: Garden\nlanguage: en\nlook_saved: true\ntheme:\n  bg: \"#202020\"\n  fg: \"#ffffff\"\n  accent: \"#ffffff\"\n  font: mono\n  scale_pct: 110\n  speakers: label\n  idle_message: Hi\n"), 0o644)
	os.WriteFile(filepath.Join(sd, "script.en.md"), []byte("[clear]\nHello"), 0o644)
	s := New(cfg, "t")
	defer s.Close()
	ts := httptest.NewServer(s.Handler())
	defer ts.Close()
	c := newClient(t, ts.URL)
	c.login("admin", "adminpw")
	_, r := c.do("GET", "/api/admin/shows/garden-1234", nil)
	th, _ := r["theme"].(map[string]any)
	if r["name"] != "Garden" || th["font"] != "mono" || th["scalePct"].(float64) != 110 || th["bg"] != "#202020" {
		t.Fatalf("%v", r)
	}
	_, f := c.do("GET", "/api/admin/shows/garden-1234/showfile", nil)
	if !strings.Contains(f["text"].(string), "title: Garden") || !strings.Contains(f["text"].(string), "Hello") {
		t.Errorf("%v", f["text"])
	}
	if _, err := os.Stat(filepath.Join(sd, "old-format", "show.yaml")); err != nil {
		t.Error("the old files are kept")
	}
}

func TestExportAndImportKeepEverythingTheShowFileSays(t *testing.T) {
	_, ts, c := serverWith(t, nil)
	id := createShow(t, c, "Portable")
	putFile(t, c, id, "", "[settings]\ntitle: Portable\nfont: dyslexic\nbackground: #123456\n[/settings]\nHello")
	resp, _ := c.hc.Get(ts.URL + "/api/admin/shows/" + id + "/export")
	b := make([]byte, 1<<16)
	n, _ := resp.Body.Read(b)
	resp.Body.Close()
	if !strings.Contains(string(b[:n]), "show.md") || strings.Contains(string(b[:n]), "show.yaml") {
		t.Errorf("export names: %q", b[:200])
	}
	// import it again
	body, ctype := multipartZip(t, b[:n])
	code, r := postRaw(t, c, ts.URL+"/api/admin/shows/import", ctype, body)
	if code != 201 {
		t.Fatalf("import: %d %v", code, r)
	}
	show := r["show"].(map[string]any)
	th := show["theme"].(map[string]any)
	if show["name"] != "Portable (imported)" || th["font"] != "dyslexic" || th["bg"] != "#123456" {
		t.Errorf("%v", show)
	}
}

func multipartZip(t *testing.T, zipBytes []byte) (*strings.Reader, string) {
	t.Helper()
	var sb strings.Builder
	mw := newMultipart(&sb)
	fw, _ := mw.CreateFormFile("file", "show.zip")
	fw.Write(zipBytes)
	mw.Close()
	return strings.NewReader(sb.String()), mw.FormDataContentType()
}

func newMultipart(w io.Writer) *multipart.Writer { return multipart.NewWriter(w) }

func postRaw(t *testing.T, c *client, url, ctype string, body io.Reader) (int, map[string]any) {
	t.Helper()
	req, _ := http.NewRequest("POST", url, body)
	req.Header.Set("Content-Type", ctype)
	resp, err := c.hc.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	var out map[string]any
	json.NewDecoder(resp.Body).Decode(&out)
	return resp.StatusCode, out
}
