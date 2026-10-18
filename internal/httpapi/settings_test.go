package httpapi

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/tmshort/open-caption/internal/logbuf"
	"github.com/tmshort/open-caption/internal/logx"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
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

func TestLogLevelsAndLogging(t *testing.T) {
	var buf bytes.Buffer
	log.SetOutput(&buf)
	log.SetFlags(0)
	t.Cleanup(func() { log.SetOutput(os.Stderr); log.SetFlags(log.LstdFlags); logx.SetLevel(logx.Info) })
	logx.SetLevel(logx.Debug)
	s, ts, c := serverWith(t, nil)
	_ = s
	id := createShow(t, c, "Logged")
	c.do("POST", "/api/admin/shows/"+id+"/activate", nil)
	aud := connect(t, ts.URL) // event streams still flush while the debug logger wraps the response
	aud.next(t)
	bad := newClient(t, ts.URL)
	bad.login("admin", "wrong-password")
	op := newClient(t, ts.URL)
	op.login("operator", "operatorpw")
	op.do("POST", "/api/operator/claim", map[string]bool{"takeover": true})
	op.do("POST", "/api/operator/next", nil)
	out := buf.String()
	for _, want := range []string{"INFO admin signed in", "WARN failed sign-in as admin", "INFO show \"Logged\" activated", "INFO operator control taken", "DEBUG operator command /api/operator/next", "DEBUG POST /api/admin/shows"} {
		if !strings.Contains(out, want) {
			t.Errorf("log lacks %q:\n%s", want, out)
		}
	}
	if strings.Contains(out, "adminpw") || strings.Contains(out, "wrong-password") {
		t.Error("a password reached the log")
	}
	// At warn level only problems are written.
	buf.Reset()
	logx.SetLevel(logx.Warn)
	op.do("POST", "/api/operator/next", nil)
	bad.login("admin", "wrong-again")
	if got := buf.String(); strings.Contains(got, "INFO") || strings.Contains(got, "DEBUG") || !strings.Contains(got, "WARN failed sign-in") {
		t.Errorf("warn level: %q", got)
	}
}

func TestArchivingAShow(t *testing.T) {
	_, _, c := serverWith(t, nil)
	keep, old := createShow(t, c, "Keep"), createShow(t, c, "Old")
	c.do("POST", "/api/admin/shows/"+keep+"/activate", nil)
	if code, r := c.do("POST", "/api/admin/shows/"+keep+"/archive", nil); code != 409 {
		t.Errorf("archiving the show on air: %d %v", code, r)
	}
	code, r := c.do("POST", "/api/admin/shows/"+old+"/archive", nil)
	if code != 200 || r["archived"] != true {
		t.Fatalf("archive: %d %v", code, r)
	}
	_, l := c.do("GET", "/api/admin/shows", nil)
	archived := 0
	for _, x := range l["shows"].([]any) {
		if x.(map[string]any)["archived"] == true {
			archived++
		}
	}
	if archived != 1 {
		t.Errorf("the list says which shows are archived: %v", l)
	}
	if code, r := c.do("POST", "/api/admin/shows/"+old+"/activate", nil); code != 409 || !strings.Contains(r["error"].(string), "archived") {
		t.Errorf("activating an archived show: %d %v", code, r)
	}
	if _, f := c.do("GET", "/api/admin/shows/"+old+"/showfile", nil); !strings.Contains(f["text"].(string), "archived: yes") {
		t.Errorf("%v", f["text"])
	}
	code, r = c.do("POST", "/api/admin/shows/"+old+"/unarchive", nil)
	if code != 200 || r["archived"] == true {
		t.Fatalf("restore: %d %v", code, r)
	}
	if code, _ := c.do("POST", "/api/admin/shows/"+old+"/activate", map[string]bool{"force": true}); code != 200 {
		t.Errorf("a restored show can go on air: %d", code)
	}
	// The show on air cannot be archived by typing the setting into its show file either.
	_, cur := c.do("GET", "/api/admin/shows/"+old+"/showfile", nil)
	code, r = c.do("PUT", "/api/admin/shows/"+old+"/showfile", map[string]string{"text": showfileWithArchived(cur["text"].(string))})
	if code != 409 || !strings.Contains(r["error"].(string), "on air") {
		t.Errorf("archiving the active show by editing: %d %v", code, r)
	}
	if _, g := c.do("GET", "/api/admin/shows/"+old+"/showfile", nil); strings.Contains(g["text"].(string), "archived") {
		t.Error("the refused save must not change the file")
	}
	// An archived show can be deleted.
	c.do("POST", "/api/admin/shows/"+keep+"/archive", nil)
	if code, _ := c.do("DELETE", "/api/admin/shows/"+keep, nil); code != 200 {
		t.Errorf("delete archived: %d", code)
	}
	if code, _ := c.do("POST", "/api/admin/shows/missing-0000/archive", nil); code != 404 {
		t.Errorf("unknown show: %d", code)
	}
}

func TestEarlierVersionsOfAShowFile(t *testing.T) {
	_, _, c := serverWith(t, nil)
	id := createShow(t, c, "Versions")
	putFile(t, c, id, "", titled("Versions", "First draft"))
	putFile(t, c, id, "", titled("Versions", "Second draft"))
	_, r := c.do("GET", "/api/admin/shows/"+id+"/showfile/history", nil)
	vs := r["versions"].([]any)
	if len(vs) != 2 {
		t.Fatalf("%v", r)
	}
	newest := vs[0].(map[string]any)["id"].(string)
	_, v := c.do("GET", "/api/admin/shows/"+id+"/showfile/history/"+newest, nil)
	if !strings.Contains(v["text"].(string), "First draft") {
		t.Errorf("%v", v)
	}
	if code, _ := c.do("GET", "/api/admin/shows/"+id+"/showfile/history/..%2Fshow.md", nil); code != 404 && code != 400 {
		t.Errorf("traversal: %d", code)
	}
	if code, _ := c.do("GET", "/api/admin/shows/missing-0000/showfile/history", nil); code != 404 {
		t.Errorf("unknown show: %d", code)
	}
}

func showfileWithArchived(text string) string {
	return "[settings]\narchived: yes\n[/settings]\n" + text
}

func TestLogLevelCanBeChangedInSetupAndIsKept(t *testing.T) {
	t.Cleanup(func() { logx.SetLevel(logx.Info) })
	cfgPath := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(cfgPath, []byte("# kept\nadmin_password: adminpw\noperator_password: operatorpw\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	_, _, c := serverWith(t, func(cfg *config.Config) { cfg.Path = cfgPath })
	_, g := c.do("GET", "/api/admin/settings", nil)
	if g["logLevel"] != "info" {
		t.Errorf("%v", g["logLevel"])
	}
	code, r := c.do("PUT", "/api/admin/settings/log-level", map[string]string{"level": "debug"})
	if code != 200 || r["level"] != "debug" || r["saved"] != true || logx.Current() != logx.Debug {
		t.Fatalf("%d %v", code, r)
	}
	if b, _ := os.ReadFile(cfgPath); !strings.Contains(string(b), "# kept") || !strings.Contains(string(b), "level: debug") {
		t.Errorf("config.yaml:\n%s", b)
	}
	for _, bad := range []string{"loud", ""} {
		if code, _ := c.do("PUT", "/api/admin/settings/log-level", map[string]string{"level": bad}); code != 400 {
			t.Errorf("%q: %d", bad, code)
		}
	}
	if logx.Current() != logx.Debug {
		t.Error("a refused change must not change the level")
	}
}

func TestSetupSaysWhenALogLevelFlagOverridesTheSavedOne(t *testing.T) {
	t.Cleanup(func() { logx.SetLevel(logx.Info) })
	cfgPath := filepath.Join(t.TempDir(), "config.yaml")
	os.WriteFile(cfgPath, []byte("admin_password: adminpw\noperator_password: operatorpw\n"), 0o600)
	s, _, c := serverWith(t, func(cfg *config.Config) { cfg.Path = cfgPath })
	_, g := c.do("GET", "/api/admin/settings", nil)
	if g["logLevelPinned"] != "" {
		t.Errorf("nothing pinned yet: %v", g["logLevelPinned"])
	}
	if _, r := c.do("PUT", "/api/admin/settings/log-level", map[string]string{"level": "warn"}); r["note"] != nil {
		t.Errorf("no note without a flag: %v", r)
	}
	s.PinLogLevel("debug")
	_, g = c.do("GET", "/api/admin/settings", nil)
	if g["logLevelPinned"] != "debug" {
		t.Errorf("%v", g["logLevelPinned"])
	}
	_, r := c.do("PUT", "/api/admin/settings/log-level", map[string]string{"level": "info"})
	if note, _ := r["note"].(string); !strings.Contains(note, "--log-level debug") || !strings.Contains(note, "restarted") || r["level"] != "info" {
		t.Errorf("%v", r)
	}
}

func TestRecentLogViewer(t *testing.T) {
	s, ts, c := serverWith(t, nil)
	lb := logbuf.New(100)
	for _, ln := range []string{
		"2026/10/15 10:00:00 INFO admin signed in from 10.0.0.2",
		"2026/10/15 10:00:01 DEBUG GET /healthz -> 200 (0s) from 10.0.0.2",
		"2026/10/15 10:00:02 WARN failed sign-in as admin from 10.0.0.9: wrong password",
		"2026/10/15 10:00:03 ERROR config: something broke",
		"a line from another writer",
	} {
		lb.Write([]byte(ln + "\n"))
	}
	s.SetLogSource(lb)
	lines := func(q string) []any {
		code, r := c.do("GET", "/api/admin/log"+q, nil)
		if code != 200 {
			t.Fatalf("%s: %d %v", q, code, r)
		}
		return r["lines"].([]any)
	}
	if got := lines(""); len(got) != 5 {
		t.Errorf("everything by default: %v", got)
	}
	if got := lines("?level=info"); len(got) != 4 {
		t.Errorf("info and up (a line with no level counts as info): %v", got)
	}
	if got := lines("?level=warn"); len(got) != 2 || !strings.Contains(got[0].(string), "failed sign-in") {
		t.Errorf("problems only: %v", got)
	}
	if got := lines("?level=error"); len(got) != 1 {
		t.Errorf("errors only: %v", got)
	}
	if got := lines("?lines=2"); len(got) != 2 || got[1] != "a line from another writer" {
		t.Errorf("the newest lines: %v", got)
	}
	if code, _ := c.do("GET", "/api/admin/log?level=loud", nil); code != 400 {
		t.Errorf("bad level: %d", code)
	}
	// Only Setup may read it.
	op := newClient(t, ts.URL)
	op.login("operator", "operatorpw")
	if code, _ := op.do("GET", "/api/admin/log", nil); code != 403 && code != 401 {
		t.Errorf("the operator must not read the log: %d", code)
	}
	if code, _ := newClient(t, ts.URL).do("GET", "/api/admin/log", nil); code != 401 {
		t.Errorf("signed out: %d", code)
	}
}

// fakeResponder stands in for the network so tests never send multicast packets.
type fakeResponder struct {
	name   string
	closed *bool
}

func (f fakeResponder) Name() string { return f.name }
func (f fakeResponder) Host() string { return f.name + ".local" }
func (f fakeResponder) Close()       { *f.closed = true }

func TestLocalNetworkNameCanBeChosenInSetup(t *testing.T) {
	cfgPath := filepath.Join(t.TempDir(), "config.yaml")
	os.WriteFile(cfgPath, []byte("# kept\nadmin_password: adminpw\noperator_password: operatorpw\nhttp:\n  addr: \":80\"\n"), 0o600)
	s, _, c := serverWith(t, func(cfg *config.Config) { cfg.Path = cfgPath })
	var starts []string
	var firstClosed bool
	s.mdns.start = func(name string) (responder, error) {
		starts = append(starts, name)
		if name == "taken" {
			return fakeResponder{name: "taken-2", closed: new(bool)}, nil
		}
		if name == "broken" {
			return nil, errors.New("no multicast here")
		}
		return fakeResponder{name: name, closed: &firstClosed}, nil
	}
	s.SetListeners(Listeners{HTTPAddr: "[::]:80", HTTPConfigured: ":80"})

	code, r := c.do("PUT", "/api/admin/settings/mdns", map[string]any{"enabled": true, "name": " Box-Office.LOCAL "})
	st := r["mdns"].(map[string]any)
	if code != 200 || st["active"] != true || st["host"] != "box-office.local" || r["saved"] != true {
		t.Fatalf("%d %v", code, r)
	}
	if b, _ := os.ReadFile(cfgPath); !strings.Contains(string(b), "# kept") || !strings.Contains(string(b), "name: box-office") || !strings.Contains(string(b), "enabled: true") {
		t.Errorf("config.yaml:\n%s", b)
	}
	// It shows in the list of audience addresses, first, and a QR code may be made for it.
	_, n := c.do("GET", "/api/admin/network", nil)
	first := n["urls"].([]any)[0].(map[string]any)
	if first["url"] != "http://box-office.local/" || first["name"] != true {
		t.Errorf("%v", first)
	}
	if code, _ := c.do("GET", "/api/admin/network/qr.png?url="+url.QueryEscape("http://box-office.local/"), nil); code != 200 {
		t.Errorf("qr for the local name: %d", code)
	}
	if code, _ := c.do("GET", "/api/admin/network/qr.png?url="+url.QueryEscape("http://evil.local/"), nil); code != 400 {
		t.Errorf("qr for someone else's name: %d", code)
	}
	_, g := c.do("GET", "/api/admin/settings", nil)
	if g["mdns"].(map[string]any)["host"] != "box-office.local" {
		t.Errorf("%v", g["mdns"])
	}

	// Changing the name stops the old announcement; a name already in use elsewhere is reported.
	_, r = c.do("PUT", "/api/admin/settings/mdns", map[string]any{"enabled": true, "name": "taken"})
	st = r["mdns"].(map[string]any)
	if !firstClosed || st["renamed"] != true || st["host"] != "taken-2.local" {
		t.Errorf("closed=%v %v", firstClosed, st)
	}
	_, r = c.do("PUT", "/api/admin/settings/mdns", map[string]any{"enabled": true, "name": "broken"})
	if st = r["mdns"].(map[string]any); st["active"] != false || !strings.Contains(st["error"].(string), "no multicast") {
		t.Errorf("%v", st)
	}
	_, r = c.do("PUT", "/api/admin/settings/mdns", map[string]any{"enabled": false, "name": "box-office"})
	if st = r["mdns"].(map[string]any); st["enabled"] != false || st["active"] != false {
		t.Errorf("%v", st)
	}
	if b, _ := os.ReadFile(cfgPath); !strings.Contains(string(b), "enabled: false") {
		t.Errorf("the choice is kept:\n%s", b)
	}
	// Bad names are refused and change nothing.
	for _, bad := range []string{"has space", "", "a.b", "-x"} {
		if code, _ := c.do("PUT", "/api/admin/settings/mdns", map[string]any{"enabled": true, "name": bad}); code != 400 {
			t.Errorf("%q: %d", bad, code)
		}
	}
	if len(starts) != 3 {
		t.Errorf("the network was started only for valid names: %v", starts)
	}
}

func TestHelpIsAvailableInEveryLanguage(t *testing.T) {
	_, _, c := serverWith(t, nil)
	titles := func(lang string) map[string]string {
		code, r := c.do("GET", "/api/help/topics?lang="+lang, nil)
		if code != 200 {
			t.Fatalf("%s: %d", lang, code)
		}
		out := map[string]string{}
		for _, x := range r["topics"].([]any) {
			m := x.(map[string]any)
			out[m["id"].(string)] = m["title"].(string)
		}
		return out
	}
	en := titles("")
	for lang, want := range map[string]string{"es": "Primeros pasos", "fr": "Premiers pas", "de": "Erste Schritte", "it": "Primi passi", "pt": "Primeiros passos"} {
		got := titles(lang)
		if len(got) != len(en) || got["admin-getting-started"] != want {
			t.Errorf("%s: %q (%d topics, English has %d)", lang, got["admin-getting-started"], len(got), len(en))
		}
		if got["admin-usage-guide"] == en["admin-usage-guide"] || got["admin-venue-network-guide"] == en["admin-venue-network-guide"] {
			t.Errorf("%s: the guides are not translated", lang)
		}
	}
	if got := titles("klingon"); got["admin-getting-started"] != en["admin-getting-started"] {
		t.Errorf("an unknown language falls back to English: %q", got["admin-getting-started"])
	}
	// A topic renders in the language asked for, and search looks in that language.
	resp, _ := c.hc.Get(c.base + "/help/admin-images?lang=fr")
	b, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	if resp.StatusCode != 200 || !strings.Contains(string(b), "Téléversez des images") || strings.Contains(string(b), "Upload pictures") {
		t.Errorf("French topic: %d %.200s", resp.StatusCode, b)
	}
	_, hits := c.do("GET", "/api/help/search?lang=de&q=Leertaste", nil)
	if len(hits["hits"].([]any)) == 0 {
		t.Error("German search finds nothing for Leertaste")
	}
	_, hits = c.do("GET", "/api/help/search?q=Leertaste", nil)
	if len(hits["hits"].([]any)) != 0 {
		t.Error("English search must not find German words")
	}
}

func TestNamingAndKeepingVersionsOverTheAPI(t *testing.T) {
	_, _, c := serverWith(t, nil)
	id := createShow(t, c, "Names")
	putFile(t, c, id, "", titled("Names", "First"))
	putFile(t, c, id, "", titled("Names", "Second"))
	_, r := c.do("GET", "/api/admin/shows/"+id+"/showfile/history", nil)
	v0 := r["versions"].([]any)[0].(map[string]any)["id"].(string)
	if code, _ := c.do("PUT", "/api/admin/shows/"+id+"/showfile/history/"+v0+"/label", map[string]string{"label": "Dress rehearsal"}); code != 200 {
		t.Fatalf("label: %d", code)
	}
	code, k := c.do("POST", "/api/admin/shows/"+id+"/showfile/history", map[string]string{"label": "Opening night"})
	if code != 201 || k["label"] != "Opening night" {
		t.Fatalf("keep: %d %v", code, k)
	}
	_, r = c.do("GET", "/api/admin/shows/"+id+"/showfile/history", nil)
	labels := toJSON(r["versions"])
	if !strings.Contains(labels, "Dress rehearsal") || !strings.Contains(labels, "Opening night") {
		t.Errorf("%s", labels)
	}
	if code, _ := c.do("POST", "/api/admin/shows/"+id+"/showfile/history", map[string]string{"label": ""}); code != 400 {
		t.Errorf("empty name: %d", code)
	}
	if code, _ := c.do("PUT", "/api/admin/shows/"+id+"/showfile/history/20260101T000000.000Z/label", map[string]string{"label": "x"}); code != 404 {
		t.Errorf("unknown version: %d", code)
	}
}

// A small zip can unpack to a great deal; an import whose files together are too big is refused.
func TestImportRefusesAZipThatUnpacksToTooMuch(t *testing.T) {
	_, ts, c := serverWith(t, nil)
	var zb bytes.Buffer
	zw := zip.NewWriter(&zb)
	chunk := make([]byte, 1<<20)
	for i := 0; i < 21; i++ {
		f, _ := zw.Create(fmt.Sprintf("assets/big%d.png", i))
		for j := 0; j < 10; j++ {
			f.Write(chunk)
		}
	}
	zw.Close()
	if zb.Len() > maxZipBody {
		t.Fatalf("test zip is not small: %d", zb.Len())
	}
	body, ctype := multipartZip(t, zb.Bytes())
	code, out := postRaw(t, c, ts.URL+"/api/admin/shows/import", ctype, body)
	if code != http.StatusRequestEntityTooLarge {
		t.Errorf("%d %v, want 413", code, out)
	}
}
