package httpapi

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"image"
	"image/png"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func adminServer(t *testing.T) (*Server, *httptest.Server, *client) {
	t.Helper()
	cfg := testConfig()
	cfg.DataDir = t.TempDir()
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

func pngData(t *testing.T) []byte {
	var b bytes.Buffer
	png.Encode(&b, image.NewRGBA(image.Rect(0, 0, 4, 4)))
	return b.Bytes()
}

func (c *client) upload(id string, files map[string][]byte) (int, map[string]any) {
	c.t.Helper()
	var body bytes.Buffer
	mw := multipart.NewWriter(&body)
	for name, data := range files {
		fw, _ := mw.CreateFormFile("files", name)
		fw.Write(data)
	}
	mw.Close()
	req, _ := http.NewRequest("POST", c.base+"/api/admin/shows/"+id+"/assets", &body)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	resp, err := c.hc.Do(req)
	if err != nil {
		c.t.Fatal(err)
	}
	defer resp.Body.Close()
	var out map[string]any
	json.NewDecoder(resp.Body).Decode(&out)
	return resp.StatusCode, out
}

func createShow(t *testing.T, c *client, name string) string {
	t.Helper()
	code, r := c.do("POST", "/api/admin/shows", map[string]any{"name": name})
	if code != 201 {
		t.Fatalf("create: %d %v", code, r)
	}
	return r["id"].(string)
}

func TestAdminRequiresAdminSession(t *testing.T) {
	_, ts, _ := adminServer(t)
	anon := newClient(t, ts.URL)
	op := newClient(t, ts.URL)
	op.login("operator", "operatorpw")
	for _, p := range []string{"/api/admin/shows", "/api/admin/network", "/api/admin/settings"} {
		if code, _ := anon.do("GET", p, nil); code != 401 {
			t.Errorf("anon %s: %d", p, code)
		}
		if code, _ := op.do("GET", p, nil); code != 403 {
			t.Errorf("operator %s: %d", p, code)
		}
	}
	if code, _ := op.do("POST", "/api/admin/rehearsal/start", map[string]string{"showId": "x"}); code != 403 {
		t.Errorf("operator rehearsal: %d", code)
	}
}

func TestShowLifecycleScriptAndAssets(t *testing.T) {
	_, _, c := adminServer(t)
	id := createShow(t, c, "Garden Party")

	script := "[image: curtian.png]\n[image: used.png]\nHello **world**\n<b>x</b>\n"
	code, r := c.do("PUT", "/api/admin/shows/"+id+"/script", map[string]string{"text": script})
	if code != 200 {
		t.Fatalf("put script: %d %v", code, r)
	}
	a := r["analysis"].(map[string]any)
	if len(a["cues"].([]any)) != 4 || len(a["missing"].([]any)) != 2 {
		t.Errorf("analysis %v", a)
	}

	if code, r := c.upload(id, map[string][]byte{"used.png": pngData(t), "curtain.png": pngData(t), "bad.png": []byte("<svg/>"), "x.svg": []byte("<svg/>")}); code != 200 ||
		len(r["saved"].([]any)) != 2 || len(r["failed"].([]any)) != 2 {
		t.Errorf("upload: %d %v", code, r)
	}
	_, r = c.do("GET", "/api/admin/shows/"+id+"/script", nil)
	a = r["analysis"].(map[string]any)
	diags := toJSON(a["diagnostics"])
	if !strings.Contains(diags, `Did you mean \"curtain.png\"`) {
		t.Errorf("fuzzy suggestion missing: %s", diags)
	}
	if len(a["used"].(map[string]any)["used.png"].([]any)) != 1 {
		t.Errorf("used: %v", a["used"])
	}
	if code, _ := c.do("DELETE", "/api/admin/shows/"+id+"/assets/curtain.png", nil); code != 200 {
		t.Errorf("delete asset: %d", code)
	}
	if code, _ := c.do("DELETE", "/api/admin/shows/"+id+"/assets/..%2Fshow.yaml", nil); code == 200 {
		t.Error("traversal delete succeeded")
	}
	// validate does not save
	c.do("POST", "/api/admin/shows/"+id+"/script/validate", map[string]string{"text": "totally different"})
	_, r = c.do("GET", "/api/admin/shows/"+id+"/script", nil)
	if r["text"] != script {
		t.Error("validate must not save")
	}
	// bad ids
	if code, _ := c.do("GET", "/api/admin/shows/NOPE", nil); code != 400 {
		t.Errorf("bad id: %d", code)
	}
	if code, _ := c.do("GET", "/api/admin/shows/missing-0000", nil); code != 404 {
		t.Errorf("unknown show: %d", code)
	}
}

func TestThemeUpdateValidatesAndReachesAudience(t *testing.T) {
	s, ts, c := adminServer(t)
	id := createShow(t, c, "T")
	if code, _ := c.do("PUT", "/api/admin/shows/"+id, map[string]any{"theme": map[string]any{"bg": "red"}}); code != 400 {
		t.Errorf("bad theme: %d", code)
	}
	if code, _ := c.do("POST", "/api/admin/shows/"+id+"/activate", nil); code != 200 {
		t.Fatal("activate")
	}
	aud := connect(t, ts.URL)
	aud.next(t)
	code, _ := c.do("PUT", "/api/admin/shows/"+id, map[string]any{"theme": map[string]any{"bg": "#101010", "fg": "#ffee00", "scalePct": 140}})
	if code != 200 {
		t.Fatalf("theme: %d", code)
	}
	ev := aud.next(t)
	th := ev.Data["theme"].(map[string]any)
	if ev.Type != "snapshot" || th["bg"] != "#101010" || th["scalePct"].(float64) != 140 {
		t.Errorf("%+v", ev)
	}
	if s.rt.Theme().FG != "#ffee00" {
		t.Error("runtime theme not updated")
	}
}

func TestActivateAndLiveGuards(t *testing.T) {
	s, ts, c := adminServer(t)
	a, b := createShow(t, c, "First"), createShow(t, c, "Second")
	c.do("PUT", "/api/admin/shows/"+a+"/script", map[string]string{"text": "One\nTwo\nThree"})
	c.do("PUT", "/api/admin/shows/"+b+"/script", map[string]string{"text": "Other"})

	aud := connect(t, ts.URL)
	aud.next(t)
	if code, _ := c.do("POST", "/api/admin/shows/"+a+"/activate", nil); code != 200 {
		t.Fatal("activate a")
	}
	if ev := aud.next(t); ev.Type != "snapshot" {
		t.Errorf("activation should push a snapshot: %+v", ev)
	}
	op := newClient(t, ts.URL)
	op.login("operator", "operatorpw")
	op.do("POST", "/api/operator/claim", map[string]bool{"takeover": true})
	op.do("POST", "/api/operator/next", nil)
	aud.nextEvent(t) // the caption

	// Live: switching shows needs force; editing the active script needs the show name.
	if code, r := c.do("POST", "/api/admin/shows/"+b+"/activate", nil); code != 409 || r["live"] != true {
		t.Errorf("activate while live: %d %v", code, r)
	}
	code, r := c.do("PUT", "/api/admin/shows/"+a+"/script", map[string]string{"text": "One\nTWO\nThree\nFour"})
	if code != 409 || r["needConfirm"] != true || r["showName"] != "First" {
		t.Errorf("live edit without confirm: %d %v", code, r)
	}
	code, r = c.do("PUT", "/api/admin/shows/"+a+"/script", map[string]string{"text": "One\nTWO\nThree\nFour", "confirm": "First"})
	if code != 200 || r["countChanged"] != true || r["wasLive"] != true {
		t.Errorf("confirmed live edit: %d %v", code, r)
	}
	if op := s.rt.Engine().Operator(); op.Cursor != 1 || op.Total != 4 {
		t.Errorf("cursor should be kept across a live edit: %+v", op)
	}
	if code, _ := c.do("DELETE", "/api/admin/shows/"+a, nil); code != 409 {
		t.Errorf("delete live show: %d", code)
	}
	if code, _ := c.do("POST", "/api/admin/shows/"+b+"/activate", map[string]bool{"force": true}); code != 200 {
		t.Errorf("forced activate: %d", code)
	}
	if s.rt.ShowID() != b || s.rt.Engine().Operator().Live {
		t.Error("new show should be active and not live")
	}
}

func TestSampleShowAndExportImportRoundTrip(t *testing.T) {
	_, ts, c := adminServer(t)
	code, r := c.do("POST", "/api/admin/shows", map[string]any{"sample": true})
	if code != 201 {
		t.Fatalf("sample: %d %v", code, r)
	}
	id := r["id"].(string)
	_, r = c.do("GET", "/api/admin/shows/"+id+"/script", nil)
	a := r["analysis"].(map[string]any)
	for _, d := range a["diagnostics"].([]any) {
		if d.(map[string]any)["severity"] != "info" {
			t.Errorf("sample show should be clean: %v", d)
		}
	}
	if len(a["assets"].([]any)) != 3 {
		t.Errorf("sample assets: %v", a["assets"])
	}

	resp, err := c.hc.Get(ts.URL + "/api/admin/shows/" + id + "/export")
	if err != nil {
		t.Fatal(err)
	}
	zbytes, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	zr, err := zip.NewReader(bytes.NewReader(zbytes), int64(len(zbytes)))
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	for _, f := range zr.File {
		names = append(names, f.Name)
	}
	if !strings.Contains(strings.Join(names, ","), "show.yaml") || !strings.Contains(strings.Join(names, ","), "assets/welcome.png") {
		t.Errorf("zip contents: %v", names)
	}

	var body bytes.Buffer
	mw := multipart.NewWriter(&body)
	fw, _ := mw.CreateFormFile("file", "show.zip")
	fw.Write(zbytes)
	mw.Close()
	req, _ := http.NewRequest("POST", ts.URL+"/api/admin/shows/import", &body)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	resp, err = c.hc.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	var imp map[string]any
	json.NewDecoder(resp.Body).Decode(&imp)
	resp.Body.Close()
	if resp.StatusCode != 201 {
		t.Fatalf("import: %d %v", resp.StatusCode, imp)
	}
	nid := imp["show"].(map[string]any)["id"].(string)
	if nid == id || !strings.Contains(imp["show"].(map[string]any)["name"].(string), "(imported)") {
		t.Errorf("import should create a new show: %v", imp)
	}
	_, r = c.do("GET", "/api/admin/shows/"+nid+"/script", nil)
	if len(r["analysis"].(map[string]any)["assets"].([]any)) != 3 {
		t.Errorf("imported assets: %v", r["analysis"])
	}
	// Garbage is rejected.
	body.Reset()
	mw = multipart.NewWriter(&body)
	fw, _ = mw.CreateFormFile("file", "x.zip")
	fw.Write([]byte("not a zip"))
	mw.Close()
	req, _ = http.NewRequest("POST", ts.URL+"/api/admin/shows/import", &body)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	resp, _ = c.hc.Do(req)
	resp.Body.Close()
	if resp.StatusCode != 400 {
		t.Errorf("garbage import: %d", resp.StatusCode)
	}
}

func TestNetworkAndQR(t *testing.T) {
	s, ts, c := adminServer(t)
	s.SetListeners(Listeners{HTTPAddr: "[::]:8080", HTTPConfigured: ":80", HTTPSAddr: "[::]:443", HTTPSConfigured: ":443"})
	code, r := c.do("GET", "/api/admin/network", nil)
	if code != 200 {
		t.Fatal(code)
	}
	urls := r["urls"].([]any)
	var sawLocal8080, sawHTTPS bool
	for _, u := range urls {
		url := u.(map[string]any)["url"].(string)
		sawLocal8080 = sawLocal8080 || url == "http://localhost:8080/"
		sawHTTPS = sawHTTPS || url == "https://localhost/"
	}
	if !sawLocal8080 || !sawHTTPS {
		t.Errorf("urls: %v", urls)
	}
	w := toJSON(r["warnings"])
	if !strings.Contains(w, "Port 80 was not available") || !strings.Contains(w, "self-signed") {
		t.Errorf("warnings: %s", w)
	}
	resp, _ := c.hc.Get(ts.URL + "/api/admin/network/qr.png?url=http://localhost:8080/")
	head := make([]byte, 4)
	resp.Body.Read(head)
	resp.Body.Close()
	if resp.StatusCode != 200 || string(head[1:4]) != "PNG" {
		t.Errorf("qr: %d %q", resp.StatusCode, head)
	}
	resp, _ = c.hc.Get(ts.URL + "/api/admin/network/qr.png?url=http://evil.example/")
	resp.Body.Close()
	if resp.StatusCode != 400 {
		t.Errorf("foreign QR url: %d", resp.StatusCode)
	}
}

func TestSettingsRedactsPasswords(t *testing.T) {
	_, _, c := adminServer(t)
	_, r := c.do("GET", "/api/admin/settings", nil)
	if strings.Contains(toJSON(r), "adminpw") || strings.Contains(toJSON(r), "operatorpw") {
		t.Errorf("passwords leaked: %v", r)
	}
}

func TestRehearsalIsIsolatedFromTheLiveShow(t *testing.T) {
	s, ts, c := adminServer(t)
	id := createShow(t, c, "Reh")
	c.do("PUT", "/api/admin/shows/"+id+"/script", map[string]string{"text": "One\nTwo\n[image: a.png]"})
	c.upload(id, map[string][]byte{"a.png": pngData(t)})
	c.do("POST", "/api/admin/shows/"+id+"/activate", nil)

	live := connect(t, ts.URL)
	live.next(t)

	if code, _ := c.do("GET", "/api/admin/rehearsal/state", nil); code != 404 {
		t.Errorf("no rehearsal yet: %d", code)
	}
	if code, _ := c.do("POST", "/api/admin/rehearsal/start", map[string]string{"showId": id}); code != 200 {
		t.Fatal("start rehearsal")
	}
	reh, err := c.hc.Get(ts.URL + "/api/admin/rehearsal/events")
	if err != nil {
		t.Fatal(err)
	}
	defer reh.Body.Close()
	if reh.StatusCode != 200 {
		t.Fatalf("rehearsal events: %d", reh.StatusCode)
	}

	if code, _ := c.do("POST", "/api/admin/rehearsal/next", nil); code != 200 {
		t.Fatalf("rehearsal next")
	}
	c.do("POST", "/api/admin/rehearsal/next", nil)
	c.do("POST", "/api/admin/rehearsal/next", nil)
	if op := s.rt.Engine().Operator(); op.Cursor != 0 || op.Live {
		t.Errorf("the live show must not move: %+v", op)
	}
	select {
	case ev := <-live.ch:
		t.Errorf("live audience received a rehearsal event: %+v", ev)
	default:
	}
	_, st := c.do("GET", "/api/admin/rehearsal/state", nil)
	if st["payload"].(map[string]any)["op"].(map[string]any)["cursor"].(float64) != 3 {
		t.Errorf("rehearsal state: %v", st["payload"])
	}
	// The live show's own operator API is untouched by the rehearsal controller.
	if code, _ := c.do("GET", "/api/operator/state", nil); code != 200 {
		t.Errorf("admin on operator api: %d", code)
	}
	// Logging out drops the rehearsal.
	c.do("POST", "/api/logout", nil)
	s.rmu.Lock()
	n2 := len(s.rehearsals)
	s.rmu.Unlock()
	if n2 != 0 {
		t.Errorf("rehearsal should be dropped on logout: %d", n2)
	}
}

func TestBootActiveRestoresShowAndPosition(t *testing.T) {
	cfg := testConfig()
	cfg.DataDir = t.TempDir()
	s1 := New(cfg, "t")
	s1.Runtime().EnablePersistence(cfg.DataDir + "/live.json")
	ts1 := httptest.NewServer(s1.Handler())
	c := newClient(t, ts1.URL)
	c.login("admin", "adminpw")
	id := createShow(t, c, "Persist")
	c.do("PUT", "/api/admin/shows/"+id+"/script", map[string]string{"text": "A\nB\nC"})
	c.do("POST", "/api/admin/shows/"+id+"/activate", nil)
	c.do("POST", "/api/operator/claim", map[string]bool{"takeover": true})
	c.do("POST", "/api/operator/next", nil)
	c.do("POST", "/api/operator/next", nil)
	ts1.Close()
	s1.Close()

	s2 := New(cfg, "t")
	defer s2.Close()
	s2.Runtime().EnablePersistence(cfg.DataDir + "/live.json")
	if !s2.BootActive() || !s2.Runtime().Restore() {
		t.Fatal("expected to resume")
	}
	if op := s2.Runtime().Engine().Operator(); op.Cursor != 2 || s2.Runtime().ShowID() != id {
		t.Errorf("%+v show=%s", op, s2.Runtime().ShowID())
	}
}
