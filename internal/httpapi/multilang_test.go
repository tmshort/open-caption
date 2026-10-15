package httpapi

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/tmshort/open-caption/internal/live"
	"github.com/tmshort/open-caption/internal/showfile"
)

func TestNegotiateLanguage(t *testing.T) {
	langs := []LangInfo{{"en", "English"}, {"es", "Español"}, {"pt-br", "Português (BR)"}, {"fr", "Français"}}
	cases := []struct{ query, accept, want string }{
		{"es", "", "es"},                           // explicit choice
		{"es", "fr", "es"},                         // explicit beats the browser
		{"zz", "fr", "fr"},                         // unknown explicit choice: fall back to the browser
		{"", "es-MX,es;q=0.9,en;q=0.8", "es"},      // regional falls back to the base language
		{"", "de;q=1.0, fr;q=0.8, en;q=0.5", "fr"}, // first one the show has, by quality
		{"", "en;q=0.2, fr;q=0.9", "fr"},           // order in the header does not matter, quality does
		{"", "pt", "pt-br"},                        // a plain request is served a regional variant
		{"", "pt-PT", "pt-br"},                     //
		{"", "FR-ca", "fr"},                        // case-insensitive
		{"", "ja,zh;q=0.9", "en"},                  // nothing matches: the show's main language
		{"", "", "en"},                             //
		{"", "*", "en"},                            //
		{"", "es;q=0", "en"},                       // q=0 means "not acceptable"
		{"", "garbage;;q=x,,,", "en"},              //
	}
	for _, c := range cases {
		if got := negotiateLang(c.query, c.accept, langs, "en"); got != c.want {
			t.Errorf("query %q accept %q: got %s, want %s", c.query, c.accept, got, c.want)
		}
	}
	if got := negotiateLang("", "es", []LangInfo{{"en", "English"}}, "en"); got != "en" {
		t.Errorf("single-language show: %s", got)
	}
}

func twoLanguageRuntime(t *testing.T, enSrc, esSrc string) (*Server, *httptest.Server) {
	t.Helper()
	cfg := testConfig()
	s := New(cfg, "test")
	t.Cleanup(s.Close)
	en, es := showfile.Parse(enSrc, showfile.Options{}), showfile.Parse(esSrc, showfile.Options{})
	if issues := showfile.CompareStructure(en.Cues, es.Cues, "English"); len(issues) > 0 {
		t.Fatalf("test show files must match: %+v", issues)
	}
	th := DefaultTheme("Soon")
	th.IdleMessages = map[string]string{"es": "Pronto"}
	s.Runtime().SetShowLangs("demo", "Demo", t.TempDir(),
		LangFile{"en", "English", en.Cues}, []LangFile{{"es", "Español", es.Cues}}, th)
	s.EnableDev()
	ts := httptest.NewServer(s.Handler())
	t.Cleanup(ts.Close)
	return s, ts
}

func connectLang(t *testing.T, base, query string, headers map[string]string) *sseReader {
	t.Helper()
	req, _ := http.NewRequest("GET", base+"/events"+query, nil)
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { resp.Body.Close() })
	r := &sseReader{resp: resp, ch: make(chan sseEvent, 100)}
	go readSSE(resp.Body, r.ch)
	return r
}

func TestOneKeypressSendsEveryLanguageTheSameLine(t *testing.T) {
	_, ts := twoLanguageRuntime(t,
		"[clear]\nHELLO\nHow are you?\n[fullscreen: INTERMISSION]\nWelcome back",
		"[clear]\nHOLA\n¿Cómo estás?\n[fullscreen: INTERMEDIO]\nBienvenidos de nuevo")

	en := connectLang(t, ts.URL, "?lang=en", nil)
	es := connectLang(t, ts.URL, "", map[string]string{"Accept-Language": "es-MX,es;q=0.9,en;q=0.5"})
	sEn, sEs := en.next(t), es.next(t)
	if sEn.Data["lang"] != "en" || sEs.Data["lang"] != "es" {
		t.Fatalf("languages: %v %v", sEn.Data["lang"], sEs.Data["lang"])
	}
	if sEs.Data["theme"].(map[string]any)["idleMessage"] != "Pronto" || sEn.Data["theme"].(map[string]any)["idleMessage"] != "Soon" {
		t.Errorf("idle messages: %v / %v", sEn.Data["theme"], sEs.Data["theme"])
	}
	if l := sEs.Data["languages"].([]any); len(l) != 2 || l[0].(map[string]any)["code"] != "en" || l[1].(map[string]any)["name"] != "Español" {
		t.Errorf("language list: %v", l)
	}

	post(t, ts.URL+"/dev/next") // [clear]
	post(t, ts.URL+"/dev/next") // HELLO / HOLA
	text := func(r *sseReader) string {
		for {
			ev := r.nextEvent(t)
			if ev.Type == "caption" {
				return ev.Data["item"].(map[string]any)["text"].(string)
			}
		}
	}
	if got := text(en); got != "HELLO" {
		t.Errorf("English: %q", got)
	}
	if got := text(es); got != "HOLA" {
		t.Errorf("Spanish: %q", got)
	}
	post(t, ts.URL+"/dev/next")
	post(t, ts.URL+"/dev/next") // fullscreen
	if got := text(es); got != "¿Cómo estás?" {
		t.Errorf("Spanish: %q", got)
	}
	var fsText string
	for {
		ev := es.nextEvent(t)
		if ev.Type == "fullscreen" {
			fsText = ev.Data["item"].(map[string]any)["html"].(string)
			break
		}
	}
	if fsText != "INTERMEDIO" {
		t.Errorf("Spanish fullscreen: %q", fsText)
	}

	// Back undoes it in every language at once; a late joiner sees the right scrollback in its own language.
	post(t, ts.URL+"/dev/back")
	late := connectLang(t, ts.URL, "?lang=es", nil)
	snap := late.next(t)
	sb := snap.Data["scrollback"].([]any)
	if len(sb) != 2 || sb[1].(map[string]any)["text"] != "¿Cómo estás?" || snap.Data["fullscreen"] != nil {
		t.Errorf("late joiner: %v", snap.Data)
	}
}

func TestLanguagesNeverDrift(t *testing.T) {
	s, ts := twoLanguageRuntime(t, "A\nB\n[loop: 0.03s]\nx\ny\n[/loop]\nC\nD", "a\nb\n[loop: 0.03s]\nx\ny\n[/loop]\nc\nd")
	rt := s.Runtime()
	cursors := func() (int, int) {
		rt.mu.RLock()
		defer rt.mu.RUnlock()
		return rt.langs["en"].engine.Operator().Cursor, rt.langs["es"].engine.Operator().Cursor
	}
	for i := 0; i < 3; i++ {
		post(t, ts.URL+"/dev/next")
	}
	if a, b := cursors(); a != 3 || b != 3 {
		t.Errorf("after 3 presses: %d %d", a, b)
	}
	post(t, ts.URL+"/dev/back")
	if a, b := cursors(); a != 2 || b != 2 {
		t.Errorf("after back: %d %d", a, b)
	}
	// A translation that somehow fell behind is put back on the operator's cue.
	rt.mu.RLock()
	es := rt.langs["es"].engine
	rt.mu.RUnlock()
	es.Jump(0, -1)
	post(t, ts.URL+"/dev/next")
	if a, b := cursors(); a != b {
		t.Errorf("languages drifted: %d vs %d", a, b)
	}
	// Both loops are running and show their own language's text.
	post(t, ts.URL+"/dev/reset")
	for _, code := range []string{"en", "es"} {
		rt.mu.RLock()
		op := rt.langs[code].engine.Operator()
		rt.mu.RUnlock()
		if op.Cursor != 0 {
			t.Errorf("%s cursor after reset: %d", code, op.Cursor)
		}
	}
}

func TestAudienceAndOperatorCountsPerLanguage(t *testing.T) {
	s, ts := twoLanguageRuntime(t, "A", "a")
	connectLang(t, ts.URL, "?lang=en", nil).next(t)
	connectLang(t, ts.URL, "?lang=es", nil).next(t)
	connectLang(t, ts.URL, "?lang=es", nil).next(t)
	connectLang(t, ts.URL, "?lang=zz", nil).next(t) // unknown: served the main language
	deadline := time.Now().Add(2 * time.Second)
	var p OperatorPayload
	for time.Now().Before(deadline) {
		p = s.Runtime().OperatorPayload()
		if p.Audience == 4 {
			break
		}
		time.Sleep(20 * time.Millisecond)
	}
	if p.Audience != 4 || p.AudienceByLang["en"] != 2 || p.AudienceByLang["es"] != 2 {
		t.Errorf("%+v", p.AudienceByLang)
	}
	if len(p.Languages) != 2 || p.Languages[0].Code != "en" {
		t.Errorf("languages: %+v", p.Languages)
	}
}

func TestSwitchingShowsSendsAnUnavailableLanguageAReload(t *testing.T) {
	s, ts := twoLanguageRuntime(t, "A", "a")
	es := connectLang(t, ts.URL, "?lang=es", nil)
	es.next(t)
	en := connectLang(t, ts.URL, "?lang=en", nil)
	en.next(t)
	s.Runtime().SetShow("one-lang", "One", "", showfile.Parse("X\nY", showfile.Options{}).Cues, DefaultTheme("Soon"))
	if ev := es.nextEvent(t); ev.Type != "reload" {
		t.Errorf("a device on a language the new show lacks must be told to reload, got %s", ev.Type)
	}
	if ev := en.nextEvent(t); ev.Type != "snapshot" || ev.Data["lang"] != "en" {
		t.Errorf("English devices keep their stream and get the new show: %+v", ev)
	}
}

func TestLiveStateOfEveryLanguageSurvivesARestart(t *testing.T) {
	cfg := testConfig()
	dir := t.TempDir() + "/live.json"
	en, es := showfile.Parse("A\nB\nC", showfile.Options{}), showfile.Parse("a\nb\nc", showfile.Options{})
	mk := func() *Server {
		s := New(cfg, "t")
		s.Runtime().EnablePersistence(dir)
		s.Runtime().SetShowLangs("demo", "Demo", "", LangFile{"en", "English", en.Cues}, []LangFile{{"es", "Español", es.Cues}}, DefaultTheme(""))
		return s
	}
	s1 := mk()
	for i := 0; i < 2; i++ {
		s1.Runtime().Do(func(e *live.Engine) (live.OperatorState, error) { return e.Next(-1) })
	}
	s1.Close()
	s2 := mk()
	defer s2.Close()
	if !s2.Runtime().Restore() {
		t.Fatal("restore")
	}
	for code, want := range map[string][]string{"en": {"A", "B"}, "es": {"a", "b"}} {
		st := s2.Runtime().State(code)
		var got []string
		for _, c := range st.Scrollback {
			got = append(got, c.Text)
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("%s: %v want %v", code, got, want)
		}
	}
	if _, err := s2.Runtime().Do(func(e *live.Engine) (live.OperatorState, error) { return e.Back(-1) }); err != nil {
		t.Fatal(err)
	}
	for code, want := range map[string]int{"en": 1, "es": 1} {
		if n := len(s2.Runtime().State(code).Scrollback); n != want {
			t.Errorf("%s after Back: %d lines", code, n)
		}
	}
}

func TestAdminTranslationWorkflow(t *testing.T) {
	s, ts, c := adminServer(t)
	id := createShow(t, c, "Bilingual")
	c.do("PUT", "/api/admin/shows/"+id+"/showfile", map[string]string{"text": "[language: en | English]\nHello\n[clear]\n[song]\nla la\n[/song]"})

	if code, r := c.do("POST", "/api/admin/shows/"+id+"/languages", map[string]any{"code": "ES", "copyFromMain": true}); code != 201 || r["code"] != "es" || r["name"] != "Español" {
		t.Fatalf("add: %d %v", code, r)
	}
	if code, _ := c.do("POST", "/api/admin/shows/"+id+"/languages", map[string]any{"code": "es"}); code != 409 {
		t.Errorf("duplicate language: %d", code)
	}
	if code, _ := c.do("POST", "/api/admin/shows/"+id+"/languages", map[string]any{"code": "en"}); code != 409 {
		t.Errorf("main language again: %d", code)
	}
	for _, bad := range []string{"", "english", "../x", "e"} {
		if code, _ := c.do("POST", "/api/admin/shows/"+id+"/languages", map[string]any{"code": bad}); code != 400 {
			t.Errorf("code %q: %d", bad, code)
		}
	}
	_, r := c.do("GET", "/api/admin/shows/"+id+"/showfile?lang=es", nil)
	text := r["text"].(string)
	if !strings.HasPrefix(text, "[language: es | Español]\n") || strings.Contains(text, "[language: en") || !strings.Contains(text, "Hello") {
		t.Errorf("copy for translating:\n%s", text)
	}
	if len(r["analysis"].(map[string]any)["diagnostics"].([]any)) > 1 { // just the summary line
		t.Logf("diagnostics: %v", r["analysis"].(map[string]any)["diagnostics"])
	}

	// The copy matches, so the translation is available.
	_, ls := c.do("GET", "/api/admin/shows/"+id+"/languages", nil)
	langs := ls["languages"].([]any)
	if len(langs) != 2 || langs[0].(map[string]any)["primary"] != true || langs[1].(map[string]any)["available"] != true || len(ls["known"].([]any)) < 20 {
		t.Fatalf("languages: %v", ls)
	}

	// Translate: same shape, different words.
	es := "[language: es | Español]\nHola\n[clear]\n[song]\nla la\n[/song]"
	if code, _ := c.do("PUT", "/api/admin/shows/"+id+"/showfile?lang=es", map[string]string{"text": es}); code != 200 {
		t.Fatal("save translation")
	}
	// Break the structure: the save is accepted (work is never refused) but the problem is reported and the language held back.
	code, r := c.do("PUT", "/api/admin/shows/"+id+"/showfile?lang=es", map[string]string{"text": "[language: es | Español]\nHola\nExtra line\n[clear]\n[song]\nla la\n[/song]"})
	if code != 200 {
		t.Fatalf("broken translation save: %d", code)
	}
	diags := toJSON(r["analysis"].(map[string]any)["diagnostics"])
	if !strings.Contains(diags, "does not match the English show file") {
		t.Errorf("structure problem not shown: %s", diags)
	}
	_, ls = c.do("GET", "/api/admin/shows/"+id+"/languages", nil)
	es0 := ls["languages"].([]any)[1].(map[string]any)
	if es0["available"] != false || len(es0["structure"].([]any)) == 0 {
		t.Errorf("a mismatched translation must not be available: %v", es0)
	}
	code, r = c.do("POST", "/api/admin/shows/"+id+"/activate", nil)
	if code != 200 || toJSON(r["skippedLanguages"]) != `["es"]` {
		t.Errorf("activation should skip the broken translation: %d %v", code, r)
	}
	if got := s.rt.Languages(); len(got) != 1 {
		t.Errorf("languages on air: %v", got)
	}

	// Fix it, and saving while the show is on air adds the language without restarting the show.
	c.do("PUT", "/api/admin/shows/"+id+"/showfile?lang=es", map[string]string{"text": es})
	if got := s.rt.Languages(); len(got) != 2 || got[1].Code != "es" || got[1].Name != "Español" {
		t.Errorf("languages on air after the fix: %v", got)
	}
	// Export includes every language; deleting works and updates the air.
	resp, _ := c.hc.Get(ts.URL + "/api/admin/shows/" + id + "/export")
	b := make([]byte, 1<<16)
	n, _ := resp.Body.Read(b)
	resp.Body.Close()
	if !strings.Contains(string(b[:n]), "show.es.md") {
		t.Error("export should include the Spanish show file")
	}
	if code, _ := c.do("DELETE", "/api/admin/shows/"+id+"/languages/en", nil); code != 400 {
		t.Errorf("deleting the main language: %d", code)
	}
	if code, _ := c.do("DELETE", "/api/admin/shows/"+id+"/languages/es", nil); code != 200 {
		t.Errorf("delete translation: %d", code)
	}
	if got := s.rt.Languages(); len(got) != 1 {
		t.Errorf("after delete: %v", got)
	}
}
