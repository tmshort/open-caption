// Copyright the open-caption authors
// SPDX-License-Identifier: Apache-2.0

package httpapi

import (
	"bufio"
	"encoding/csv"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/tmshort/open-caption/internal/report"
	"github.com/tmshort/open-caption/internal/showfile"
)

const (
	uaPhone   = "Mozilla/5.0 (iPhone; CPU iPhone OS 17_4 like Mac OS X) AppleWebKit/605.1.15 Version/17.4 Mobile/15E148 Safari/604.1"
	uaTablet  = "Mozilla/5.0 (iPad; CPU OS 16_6 like Mac OS X) AppleWebKit/605.1.15 CriOS/120.0 Mobile/15E148 Safari/604.1"
	uaDesktop = "Mozilla/5.0 (Windows NT 10.0) Chrome/124.0 Safari/537.36"
)

// REQ-070: every device that used the audience page while the show was on the air is counted once, however often it
// reconnected, and the report is kept when the run ends.
func TestAudienceReportCountsDevicesAcrossReconnects(t *testing.T) {
	s, ts, c := standbyServer(t)
	_, r := c.do("POST", "/api/admin/shows", map[string]any{"sample": true})
	c.do("POST", "/api/admin/shows/"+r["id"].(string)+"/activate", map[string]any{}) // loaded in standby
	c.do("POST", "/api/operator/claim", map[string]any{"takeover": true})
	open := func(path, lang, ua string) *http.Response {
		req, _ := http.NewRequest("GET", ts.URL+path, nil)
		req.Header.Set("User-Agent", ua)
		req.Header.Set("Accept-Language", lang)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		bufio.NewReader(resp.Body).ReadString('\n') // the stream has started
		return resp
	}
	settle := func() { time.Sleep(150 * time.Millisecond) }

	// Two devices are waiting in standby: no report yet.
	phone := open("/events", "en", uaPhone)
	defer phone.Body.Close()
	proj := open("/events?display=projector", "es", uaDesktop)
	defer proj.Body.Close()
	settle()
	if code, _ := c.do("GET", "/api/operator/report", nil); code != 404 {
		t.Errorf("nothing is counted before the show is on the air: %d", code)
	}

	// On the air: the two waiting devices are counted from the start; a tablet joins; the phone drops and comes back.
	if code, _ := c.do("POST", "/api/operator/onair", nil); code != 200 {
		t.Fatal("on air")
	}
	tablet := open("/events", "en", uaTablet)
	defer tablet.Body.Close()
	phone.Body.Close()
	settle()
	phone2 := open("/events", "en", uaPhone)
	defer phone2.Body.Close()
	settle()

	code, body := c.do("GET", "/api/operator/report", nil)
	if code != 200 {
		t.Fatalf("%d %v", code, body)
	}
	rep := body["report"].(map[string]any)
	tot := rep["totals"].(map[string]any)
	if rep["running"] != true || tot["devices"].(float64) != 3 || tot["connections"].(float64) != 4 || tot["reconnects"].(float64) != 1 ||
		tot["phones"].(float64) != 1 || tot["tablets"].(float64) != 1 || tot["projectors"].(float64) != 1 || tot["now"].(float64) != 3 {
		t.Errorf("totals: %v", tot)
	}
	langs := rep["languages"].(map[string]any)
	if langs["en"].(float64) != 2 || langs["es"].(float64) != 1 || body["languageNames"].(map[string]any)["es"] != "Español" && body["languageNames"].(map[string]any)["es"] == "" {
		t.Errorf("languages: %v %v", langs, body["languageNames"])
	}
	// The pre-show loop that started with the show is in the report, under the name the show file gives it.
	if lp, _ := rep["loops"].([]any); len(lp) == 0 || lp[0].(map[string]any)["name"] != "Pre-show" || lp[0].(map[string]any)["kind"] != "preshow" {
		t.Errorf("loops: %v", rep["loops"])
	}
	if sr, _ := rep["series"].([]any); len(sr) < 2 {
		t.Errorf("the graph has points: %v", rep["series"])
	}
	if rep["show"] == "" {
		t.Errorf("the show is named: %v", rep)
	}

	// The CSV has a row for each device and needs a sign-in like the rest of the operator API.
	resp, err := c.hc.Get(ts.URL + "/api/operator/report.csv")
	if err != nil || resp.StatusCode != 200 || !strings.HasPrefix(resp.Header.Get("Content-Type"), "text/csv") || !strings.Contains(resp.Header.Get("Content-Disposition"), "audience-report-") {
		t.Fatalf("csv: %v %v", resp, err)
	}
	rows, err := csv.NewReader(strings.NewReader(strings.TrimPrefix(readAll(resp), "\xef\xbb\xbf"))).ReadAll()
	if err != nil || len(rows) != 4 || rows[0][0] != "Device" {
		t.Errorf("csv rows: %v %v", rows, err)
	}
	if anon, err := http.Get(ts.URL + "/api/operator/report"); err != nil || anon.StatusCode != 401 {
		t.Errorf("the report is for the operator: %v %v", anon, err)
	}
	if anon, err := http.Get(ts.URL + "/operator/report"); err != nil || anon.StatusCode == 404 {
		t.Errorf("the report page is served: %v %v", anon, err)
	}

	if code, _ := c.do("DELETE", "/api/admin/reports?id=current", nil); code != 409 {
		t.Errorf("the report of the run in progress cannot be deleted: %d", code)
	}

	// Back to standby ends the run: the final report is kept in the data folder and is what is shown afterwards.
	if code, _ := c.do("POST", "/api/operator/offair", nil); code != 200 {
		t.Fatal("off air")
	}
	_, body = c.do("GET", "/api/operator/report", nil)
	rep = body["report"].(map[string]any)
	if rep["running"] != false || rep["totals"].(map[string]any)["devices"].(float64) != 3 || rep["totals"].(map[string]any)["now"].(float64) != 3 {
		t.Errorf("final: %v", rep)
	}
	files, _ := os.ReadDir(filepath.Join(s.cfg.DataDir, "reports"))
	if len(files) != 1 || !strings.HasSuffix(files[0].Name(), ".json") {
		t.Errorf("the report is saved: %v", files)
	}
	_, list := c.do("GET", "/api/operator/reports", nil)
	if l := list["reports"].([]any); len(l) != 1 || l[0].(map[string]any)["running"] != false {
		t.Errorf("list: %v", list)
	}
	if code, _ := c.do("GET", "/api/operator/report?id=../config.yaml", nil); code != 404 {
		t.Errorf("a report is found only by the name of a saved one: %d", code)
	}
	phone2.Body.Close()

	// A saved report can be deleted (admin only); one that is not saved, or a path, is refused.
	name := list["reports"].([]any)[0].(map[string]any)["id"].(string)
	if code, _ := c.do("DELETE", "/api/admin/reports?id=../config.yaml", nil); code != 404 {
		t.Errorf("a path is not a report: %d", code)
	}
	if code, _ := c.do("DELETE", "/api/admin/reports?id="+name, nil); code != 200 {
		t.Errorf("delete a saved report: %d", code)
	}
	if code, _ := c.do("DELETE", "/api/admin/reports?id="+name, nil); code != 404 {
		t.Errorf("deleted twice: %d", code)
	}
	if code, _ := c.do("GET", "/api/operator/report?id="+name, nil); code != 404 {
		t.Errorf("a deleted report is gone: %d", code)
	}
}

func readAll(resp *http.Response) string {
	defer resp.Body.Close()
	var sb strings.Builder
	sc := bufio.NewScanner(resp.Body)
	for sc.Scan() {
		sb.WriteString(sc.Text() + "\n")
	}
	return sb.String()
}

// A rehearsal (an administrator's private run-through) is not an audience: its viewers are not counted.
func TestRehearsalViewersAreNotCounted(t *testing.T) {
	s, ts, c := standbyServer(t)
	_, r := c.do("POST", "/api/admin/shows", map[string]any{"sample": true})
	id := r["id"].(string)
	c.do("POST", "/api/admin/shows/"+id+"/activate", map[string]any{})
	c.do("POST", "/api/operator/claim", map[string]any{"takeover": true})
	if code, _ := c.do("POST", "/api/operator/onair", nil); code != 200 {
		t.Fatal("on air")
	}
	if code, _ := c.do("POST", "/api/admin/rehearsal/start", map[string]string{"showId": id}); code != 200 {
		t.Fatal("start rehearsal")
	}
	reh, err := c.hc.Get(ts.URL + "/api/admin/rehearsal/events")
	if err != nil || reh.StatusCode != 200 {
		t.Fatalf("%v %v", reh, err)
	}
	defer reh.Body.Close()
	bufio.NewReader(reh.Body).ReadString('\n')
	time.Sleep(100 * time.Millisecond)
	s.syncReport()
	cur, ok := s.tracker.Current()
	if !ok || cur.Totals.Devices != 0 {
		t.Errorf("the show is on the air with nobody in the audience: %v %+v", ok, cur.Totals)
	}
}

func TestLoopsAreNamedByWhereTheyStand(t *testing.T) {
	cue := func(k showfile.CueKind) showfile.Cue { return showfile.Cue{Kind: k} }
	cues := []showfile.Cue{cue(showfile.KindLoop), cue(showfile.KindNote), cue(showfile.KindCaption), cue(showfile.KindSection), cue(showfile.KindLoop), cue(showfile.KindImage), cue(showfile.KindLoop)}
	for at, want := range map[int]string{0: report.LoopPreShow, 4: report.LoopIntermission, 6: report.LoopPostShow} {
		if got := loopKind(cues, at); got != want {
			t.Errorf("loop at %d: %s, want %s", at, got, want)
		}
	}
	if got := loopKind([]showfile.Cue{cue(showfile.KindLoop)}, 0); got != report.LoopOther {
		t.Errorf("a show of only loops: %s", got)
	}
}
