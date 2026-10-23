// Copyright the open-caption authors
// SPDX-License-Identifier: Apache-2.0

package report

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

type clock struct{ t time.Time }

func (c *clock) now() time.Time      { return c.t }
func (c *clock) add(d time.Duration) { c.t = c.t.Add(d) }
func newClock() *clock               { return &clock{t: time.Date(2026, 10, 20, 19, 30, 0, 0, time.UTC)} }

var kindOf = map[string]string{iphone: Phone, ipad: Tablet, pixel: Phone, androidT: Tablet, desktop: Computer}

const (
	iphone   = "Mozilla/5.0 (iPhone; CPU iPhone OS 17_0 like Mac OS X) AppleWebKit/605.1.15 Mobile/15E148 Safari/604.1"
	ipad     = "Mozilla/5.0 (iPad; CPU OS 17_0 like Mac OS X) AppleWebKit/605.1.15 Mobile/15E148"
	pixel    = "Mozilla/5.0 (Linux; Android 14; Pixel 8) AppleWebKit/537.36 Chrome/120 Mobile Safari/537.36"
	androidT = "Mozilla/5.0 (Linux; Android 13; SM-X700) AppleWebKit/537.36 Chrome/120 Safari/537.36"
	desktop  = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 Chrome/120 Safari/537.36"
)

// People who reconnect are one device, however often; people who arrive early are counted from the start.
func TestARunCountsDevicesNotConnections(t *testing.T) {
	c := newClock()
	tr := New(c.now)
	var closed []Report
	tr.OnClose = func(r Report) { closed = append(closed, r) }

	// Someone is already waiting in standby when the show goes on the air.
	earlyOut := tr.Connect("192.168.1.20", iphone, kindOf[iphone], "en", false)
	if _, ok := tr.Current(); ok {
		t.Fatal("nothing is counted before the show is on the air")
	}
	c.add(time.Minute)
	tr.Sync(true, "garden", "The Garden Party", nil)
	c.add(time.Minute)

	// A phone that drops off the WiFi three times, a Spanish tablet, a projector, a second tab on the first phone.
	for i := 0; i < 3; i++ {
		off := tr.Connect("192.168.1.30", pixel, kindOf[pixel], "en", false)
		c.add(10 * time.Second)
		off()
		c.add(5 * time.Second)
	}
	tablet := tr.Connect("192.168.1.31", ipad, kindOf[ipad], "es", false)
	proj := tr.Connect("192.168.1.2", desktop, kindOf[desktop], "en", true)
	tab2 := tr.Connect("192.168.1.20", iphone, kindOf[iphone], "en", false) // the early phone opens a second tab
	c.add(time.Minute)
	tab2()
	earlyOut()
	c.add(time.Minute)

	cur, ok := tr.Current()
	if !ok || !cur.Running || cur.Show != "The Garden Party" || cur.Totals.Devices != 4 {
		t.Fatalf("%+v ok=%v", cur.Totals, ok)
	}
	tot := cur.Totals
	if tot.Phones != 2 || tot.Tablets != 1 || tot.Projectors != 1 || tot.Computers != 0 {
		t.Errorf("kinds: %+v", tot)
	}
	if tot.Connections != 3+1+1+2 || tot.Reconnects != 2+0+0+1 {
		t.Errorf("connections %d reconnects %d", tot.Connections, tot.Reconnects)
	}
	if tot.Now != 2 {
		t.Errorf("still connected (tablet and projector): %d", tot.Now)
	}
	if cur.Languages["en"] != 3 || cur.Languages["es"] != 1 {
		t.Errorf("languages: %v", cur.Languages)
	}
	by := map[string]Device{}
	for _, d := range cur.Devices {
		by[d.Address] = d
	}
	early := by["192.168.1.20"]
	if early.Connections != 2 || early.Connected != Span(2*time.Minute+45*time.Second) || early.Longest != early.Connected || early.Now {
		t.Errorf("the early phone is counted from the start of the run, and its two tabs count once in time: %+v", early)
	}
	pix := by["192.168.1.30"]
	if pix.Connections != 3 || pix.Connected != Span(30*time.Second) || pix.Longest != Span(10*time.Second) || pix.Now {
		t.Errorf("a phone that reconnected: %+v", pix)
	}
	if tot.Peak != 3 || tot.PeakAt.IsZero() {
		t.Errorf("the busiest moment had three devices at once (the phone that kept dropping was never there with the others): %d", tot.Peak)
	}
	if !by["192.168.1.2"].Projector || by["192.168.1.31"].Languages[0] != "es" {
		t.Errorf("projector and language: %+v %+v", by["192.168.1.2"], by["192.168.1.31"])
	}

	// Going to standby ends the run with the people still there counted as connected at the end.
	tr.Sync(false, "garden", "The Garden Party", nil)
	if len(closed) != 1 || closed[0].Running || closed[0].Totals.Now != 2 || closed[0].Ended.IsZero() || closed[0].Totals.Devices != 4 {
		t.Fatalf("closed: %+v", closed)
	}
	if _, ok := tr.Current(); ok {
		t.Error("no run in progress after the show goes off the air")
	}
	tablet()
	proj()
	// Disconnecting after the run is harmless, and a new run starts afresh with whoever is connected then.
	late := tr.Connect("192.168.1.99", pixel, kindOf[pixel], "en", false)
	tr.Sync(true, "garden", "The Garden Party", nil)
	if cur, _ := tr.Current(); cur.Totals.Devices != 1 {
		t.Errorf("a new run starts afresh: %+v", cur.Totals)
	}
	late()
}

func TestADifferentShowEndsTheRun(t *testing.T) {
	c := newClock()
	tr := New(c.now)
	var closed int
	tr.OnClose = func(Report) { closed++ }
	tr.Connect("10.0.0.1", iphone, kindOf[iphone], "en", false)
	tr.Sync(true, "a", "A", nil)
	tr.Sync(true, "a", "A renamed", nil)
	if closed != 0 {
		t.Error("the same show on the air keeps the run (a rename does not end it)")
	}
	if cur, _ := tr.Current(); cur.Show != "A renamed" {
		t.Errorf("the name follows a rename: %q", cur.Show)
	}
	tr.Sync(true, "b", "B", nil)
	cur, ok := tr.Current()
	if closed != 1 || !ok || cur.ShowID != "b" || cur.Totals.Devices != 1 {
		t.Errorf("a different show closes the run and starts another: closed=%d %+v", closed, cur)
	}
}

func TestDeviceIDIsStableAndAnonymous(t *testing.T) {
	a, b := ID("192.168.1.20", iphone), ID("192.168.1.20", iphone)
	if a != b || len(a) != 8 || strings.Contains(a, "192") || a == ID("192.168.1.21", iphone) || a == ID("192.168.1.20", pixel) {
		t.Errorf("%s %s", a, b)
	}
}

func TestStoreDeletesOnlyASavedReport(t *testing.T) {
	s := Store{Dir: t.TempDir()}
	r := Report{ShowID: "g", Started: time.Date(2026, 10, 20, 19, 0, 0, 0, time.UTC)}
	s.Save(r)
	other := filepath.Join(filepath.Dir(s.Dir), "keep.json")
	os.WriteFile(other, []byte("{}"), 0o600)
	for _, bad := range []string{"../keep.json", "a/b.json", ".hidden.json", "x.txt", "missing.json"} {
		if s.Delete(bad) == nil {
			t.Errorf("%q is not a saved report", bad)
		}
	}
	if _, err := os.Stat(other); err != nil {
		t.Error("a file outside the folder was touched")
	}
	if err := s.Delete(Name(r)); err != nil || len(s.Names()) != 0 {
		t.Errorf("delete: %v %v", err, s.Names())
	}
}

func TestReportsStartedWithinOneSecondHaveDistinctFiles(t *testing.T) {
	s := Store{Dir: t.TempDir()}
	start := time.Date(2026, 10, 20, 19, 0, 0, 0, time.UTC)
	for _, at := range []time.Time{start, start.Add(time.Nanosecond)} {
		if err := s.Save(Report{ShowID: "garden", Started: at}); err != nil {
			t.Fatal(err)
		}
	}
	if got := s.Names(); len(got) != 2 || got[0] == got[1] {
		t.Fatalf("same-second reports overwritten: %v", got)
	}
	first := Report{ShowID: "garden party/../x", Started: start}
	second := Report{ShowID: first.ShowID, Started: start.Add(time.Nanosecond)}
	if Name(first) == Name(second) || DownloadName(first) != DownloadName(second) {
		t.Fatalf("stored names must differ, readable downloads may match: %q, %q; %q, %q", Name(first), Name(second), DownloadName(first), DownloadName(second))
	}
	if got, want := DownloadName(first), "audience-report-"+start.Local().Format("20060102-150405")+"-garden-party-x.csv"; got != want {
		t.Fatalf("unreadable or unsafe CSV name: %q", got)
	}
}

func TestStoreSavesLoadsAndKeepsTheNewest(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "reports")
	s := Store{Dir: dir}
	base := time.Date(2026, 10, 20, 19, 0, 0, 0, time.UTC)
	for i := 0; i < Keep+3; i++ {
		if err := s.Save(Report{ShowID: "garden party/../x", Show: "G", Started: base.Add(time.Duration(i) * time.Hour)}); err != nil {
			t.Fatal(err)
		}
	}
	names := s.Names()
	if len(names) != Keep {
		t.Fatalf("%d reports kept", len(names))
	}
	r, name, ok := s.Latest()
	if !ok || !r.Started.Equal(base.Add(time.Duration(Keep+2)*time.Hour)) || name != names[0] {
		t.Errorf("the newest is the latest: %v %s", r.Started, name)
	}
	if strings.Contains(name, "/") || strings.Contains(name, "..") {
		t.Errorf("a show id cannot make a path: %s", name)
	}
	// Saving the same run again (it was saved while going on) replaces the file.
	run := Report{ShowID: "g", Started: base.Add(1000 * time.Hour), Running: true}
	s.Save(run)
	run.Running, run.Ended = false, run.Started.Add(time.Hour)
	s.Save(run)
	got, _ := s.Load(Name(run))
	if got.Running || got.Ended.IsZero() {
		t.Errorf("the final report replaced the one saved while it ran: %+v", got)
	}
	for _, bad := range []string{"../x.json", "a/b.json", ".hidden.json", "x.txt", "missing.json"} {
		if _, err := s.Load(bad); err == nil {
			t.Errorf("%q must not load", bad)
		}
	}
	if st, err := os.Stat(filepath.Join(dir, Name(run))); err != nil || runtimeWindows() == false && st.Mode().Perm() != 0o600 {
		t.Errorf("the file is private: %v %v", st, err)
	}
}

func runtimeWindows() bool { return os.PathSeparator == '\\' }

func TestCSVHasARowForEachDevice(t *testing.T) {
	r := Report{Devices: []Device{{ID: "ab12cd34", Address: "192.168.1.20", Kind: Phone, Languages: []string{"en", "es"}, First: time.Date(2026, 10, 20, 19, 31, 0, 0, time.UTC), Last: time.Date(2026, 10, 20, 20, 1, 5, 0, time.UTC),
		Connections: 3, Connected: Span(29*time.Minute + 5*time.Second), Longest: Span(20 * time.Minute), Now: true}}}
	var b bytes.Buffer
	if err := WriteCSV(&b, r); err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSpace(b.String()), "\n")
	if len(lines) != 2 || !strings.HasPrefix(lines[0], "Device,Address,Type") || !strings.Contains(lines[1], "ab12cd34,192.168.1.20,phone,device,en es,") || !strings.Contains(lines[1], ",3,0:29:05,0:20:00,yes") {
		t.Errorf("%q", lines)
	}
}

func TestSpansAreWholeSecondsInJSON(t *testing.T) {
	b, err := json.Marshal(Device{Connected: Span(90*time.Second + 400*time.Millisecond), Longest: Span(time.Hour)})
	if err != nil || !strings.Contains(string(b), `"connected":90`) || !strings.Contains(string(b), `"longest":3600`) {
		t.Fatalf("%s %v", b, err)
	}
	var d Device
	if err := json.Unmarshal(b, &d); err != nil || d.Connected != Span(90*time.Second) || d.Longest != Span(time.Hour) {
		t.Errorf("%+v %v", d, err)
	}
}

// The graph: how many devices were connected at each moment, and when a loop was running.
func TestTheGraphFollowsDevicesAndLoops(t *testing.T) {
	c := newClock()
	tr := New(c.now)
	at := func(s int) time.Time {
		return time.Date(2026, 10, 20, 19, 30, 0, 0, time.UTC).Add(time.Duration(s) * time.Second)
	}
	pre := &LoopInfo{Cue: 0, Kind: LoopPreShow}
	interm := &LoopInfo{Cue: 9, Kind: LoopIntermission, Name: "Act 1 break"}

	early := tr.Connect("10.0.0.1", iphone, Phone, "en", false)
	tr.Sync(true, "s", "S", pre) // on air, with the pre-show loop running and one device waiting
	c.add(10 * time.Second)
	b := tr.Connect("10.0.0.2", pixel, Phone, "en", false)
	c.add(10 * time.Second)
	tr.Sync(true, "s", "S", nil) // the pre-show ends at 0:20
	c.add(40 * time.Second)
	b()
	c.add(10 * time.Second)
	tr.Sync(true, "s", "S", interm) // an intermission starts at 1:10
	c.add(30 * time.Second)

	cur, _ := tr.Current()
	want := []Point{{at(0), 1}, {at(10), 2}, {at(60), 1}, {at(100), 1}}
	if len(cur.Series) != len(want) {
		t.Fatalf("series: %+v", cur.Series)
	}
	for i, p := range want {
		if !cur.Series[i].T.Equal(p.T) || cur.Series[i].N != p.N {
			t.Errorf("point %d: %+v, want %+v", i, cur.Series[i], p)
		}
	}
	if len(cur.Loops) != 2 || cur.Loops[0].Kind != LoopPreShow || !cur.Loops[0].Start.Equal(at(0)) || !cur.Loops[0].End.Equal(at(20)) ||
		cur.Loops[1].Kind != LoopIntermission || cur.Loops[1].Name != "Act 1 break" || cur.Loops[0].Name != "" || !cur.Loops[1].Start.Equal(at(70)) || !cur.Loops[1].End.Equal(at(100)) {
		t.Errorf("loops: %+v", cur.Loops)
	}
	// Ending the run closes the open loop, and the graph ends there.
	var final Report
	tr.OnClose = func(r Report) { final = r }
	c.add(20 * time.Second)
	tr.Sync(false, "s", "S", nil)
	early()
	if final.Loops[1].End != at(120) || !final.Series[len(final.Series)-1].T.Equal(at(120)) || final.Running {
		t.Errorf("final: %+v %+v", final.Loops, final.Series)
	}
	// A loop that is still the same one is not split in two.
	tr.Sync(true, "s", "S", pre)
	tr.Sync(true, "s", "S", pre)
	if cur, _ := tr.Current(); len(cur.Loops) != 1 {
		t.Errorf("one loop: %+v", cur.Loops)
	}
}

func TestALongRunKeepsTheGraphSmall(t *testing.T) {
	c := newClock()
	tr := New(c.now)
	tr.Sync(true, "s", "S", nil)
	for i := 0; i < 5*maxSeries; i++ {
		off := tr.Connect("10.0.0.1", iphone, Phone, "en", false)
		c.add(time.Second)
		off()
		c.add(time.Second)
	}
	cur, _ := tr.Current()
	if len(cur.Series) > maxSeries+2 || len(cur.Series) < maxSeries/4 {
		t.Errorf("%d points", len(cur.Series))
	}
	for i := 1; i < len(cur.Series); i++ {
		if cur.Series[i].T.Before(cur.Series[i-1].T) {
			t.Fatalf("points out of order at %d", i)
		}
	}
}
