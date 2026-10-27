// Copyright the open-caption authors
// SPDX-License-Identifier: Apache-2.0

package live

import (
	"fmt"
	"math/rand"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/tmshort/open-caption/internal/showfile"
)

type rec struct {
	mu sync.Mutex
	ev []Event
}

func (r *rec) sink(e Event) { r.mu.Lock(); r.ev = append(r.ev, e); r.mu.Unlock() }
func (r *rec) types() []string {
	r.mu.Lock()
	defer r.mu.Unlock()
	var t []string
	for _, e := range r.ev {
		t = append(t, e.Type)
	}
	return t
}

func parse(src string) []showfile.Cue { return showfile.Parse(src, showfile.Options{}).Cues }

func newEngine(t *testing.T, src string) (*Engine, *rec) {
	t.Helper()
	r := &rec{}
	e := New(parse(src), Config{DefaultLoopSeconds: 10}, r.sink)
	t.Cleanup(e.Close)
	return e, r
}

func texts(s State) []string {
	out := []string{}
	for _, c := range s.Scrollback {
		out = append(out, c.Text)
	}
	return out
}

func TestNextAppendsCaptionsAndSkipsOperatorCues(t *testing.T) {
	e, r := newEngine(t, "A\n[section: S]\n[note: n]\nB\n")
	if st := e.Snapshot(); st.Mode != ModeIdle {
		t.Fatalf("initial mode %s", st.Mode)
	}
	mustNext(t, e)
	op := mustNext(t, e)
	if got := texts(e.Snapshot()); !reflect.DeepEqual(got, []string{"A", "B"}) {
		t.Fatalf("got %v", got)
	}
	if op.NextIndex != -1 || !op.Live {
		t.Errorf("%+v", op)
	}
	if _, err := e.Next(-1); err != ErrEnd {
		t.Errorf("err = %v, want ErrEnd", err)
	}
	if got := r.types(); !reflect.DeepEqual(got, []string{EvCaption, EvCaption}) {
		t.Errorf("events %v", got)
	}
}

func mustNext(t *testing.T, e *Engine) OperatorState {
	t.Helper()
	o, err := e.Next(-1)
	if err != nil {
		t.Fatal(err)
	}
	return o
}

func TestClearBlankImageFullscreen(t *testing.T) {
	e, _ := newEngine(t, "A\n[blank]\nB\n[image: x.jpg]\n[clear]\n[fullscreen: INTERMISSION]\nC")
	mustNext(t, e) // A
	mustNext(t, e) // blank
	if e.Snapshot().Mode != ModeBlank {
		t.Error("expected blank")
	}
	mustNext(t, e) // B un-blanks
	if st := e.Snapshot(); st.Mode != ModeCaptions || len(st.Scrollback) != 2 {
		t.Errorf("%+v", st)
	}
	mustNext(t, e) // image
	if st := e.Snapshot(); st.Fullscreen == nil || st.Fullscreen.Image != "x.jpg" || len(st.Scrollback) != 2 {
		t.Errorf("%+v", st)
	}
	mustNext(t, e) // clear
	if st := e.Snapshot(); st.Fullscreen != nil || len(st.Scrollback) != 0 {
		t.Errorf("%+v", st)
	}
	mustNext(t, e) // fullscreen text
	if st := e.Snapshot(); st.Fullscreen == nil || st.Fullscreen.Text != "INTERMISSION" {
		t.Errorf("%+v", st)
	}
	mustNext(t, e) // caption dismisses fullscreen
	if st := e.Snapshot(); st.Fullscreen != nil || texts(st)[0] != "C" {
		t.Errorf("%+v", st)
	}
}

func TestBackRestoresStateAndCursor(t *testing.T) {
	e, _ := newEngine(t, "A\nB\n[clear]\nC")
	for i := 0; i < 3; i++ {
		mustNext(t, e)
	}
	if len(e.Snapshot().Scrollback) != 0 {
		t.Fatal("clear should empty")
	}
	op, err := e.Back(-1)
	if err != nil {
		t.Fatal(err)
	}
	if got := texts(e.Snapshot()); !reflect.DeepEqual(got, []string{"A", "B"}) || op.Cursor != 2 {
		t.Errorf("%v %+v", got, op)
	}
	e.Back(-1)
	e.Back(-1)
	if _, err := e.Back(-1); err != ErrNoUndo {
		t.Errorf("err = %v", err)
	}
	if e.Snapshot().Mode != ModeIdle {
		t.Error("expected idle after undoing everything")
	}
}

func TestStaleRev(t *testing.T) {
	e, _ := newEngine(t, "A\nB")
	op := e.Operator()
	if _, err := e.Next(op.Rev); err != nil {
		t.Fatal(err)
	}
	if _, err := e.Next(op.Rev); err != ErrStale {
		t.Errorf("err = %v, want ErrStale", err)
	}
	if got := texts(e.Snapshot()); len(got) != 1 {
		t.Errorf("stale command applied: %v", got)
	}
}

const loopScript = "[loop: 0.03s]\nOne\nTwo\n[/loop]\nA\n[loop]\n[fullscreen: Intermission | 0.03s]\n[/loop]\nB"

func TestLeadingLoopAutoStartsAndTicks(t *testing.T) {
	e, r := newEngine(t, loopScript)
	op := e.Operator()
	if op.Loop == nil || op.Cursor != 1 || op.Live {
		t.Fatalf("auto-start: %+v", op)
	}
	if st := e.Snapshot(); st.Fullscreen == nil || st.Fullscreen.Text != "One" {
		t.Fatalf("%+v", st.Fullscreen)
	}
	waitFor(t, func() bool { return len(r.types()) >= 3 })
	r.mu.Lock()
	item := r.ev[0].Data.(showfile.Item)
	r.mu.Unlock()
	if item.Text != "Two" {
		t.Errorf("first tick shows %q, want Two", item.Text)
	}
}

func waitFor(t *testing.T, f func() bool) {
	t.Helper()
	for i := 0; i < 200; i++ {
		if f() {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("timeout")
}

func TestNextEndsLoopStopBlanksBackRestarts(t *testing.T) {
	e, _ := newEngine(t, loopScript)
	mustNext(t, e) // A ends the pre-show loop
	if e.Operator().Loop != nil || texts(e.Snapshot())[0] != "A" || e.Snapshot().Fullscreen != nil {
		t.Fatalf("%+v", e.Snapshot())
	}
	mustNext(t, e) // starts intermission loop
	if e.Operator().Loop == nil || e.Operator().Loop.Cue != 2 {
		t.Fatalf("%+v", e.Operator())
	}
	op, err := e.Stop(-1)
	if err != nil || op.Loop != nil {
		t.Fatalf("%v %+v", err, op)
	}
	if st := e.Snapshot(); st.Mode != ModeBlank || st.Fullscreen != nil {
		t.Errorf("%+v", st)
	}
	if _, err := e.Stop(-1); err != ErrNoLoop {
		t.Errorf("err = %v", err)
	}
	// Back after STOP restarts the loop from item 0.
	op, _ = e.Back(-1)
	if op.Loop == nil || op.Loop.Item != 0 || e.Snapshot().Fullscreen.Text != "Intermission" {
		t.Errorf("%+v %+v", op, e.Snapshot())
	}
	// NEXT ends it and sends B.
	mustNext(t, e)
	if e.Operator().Loop != nil || texts(e.Snapshot())[1] != "B" {
		t.Errorf("%+v", e.Snapshot())
	}
	// Back "backs into" the loop, restarting it.
	op, _ = e.Back(-1)
	if op.Loop == nil || op.Loop.Item != 0 {
		t.Errorf("%+v", op)
	}
}

func TestAnnouncementIsReportedToTheOperatorUntilReplacedOrUndone(t *testing.T) {
	e, _ := newEngine(t, "A\nB")
	mustNext(t, e)
	if e.Operator().Announcement != "" {
		t.Fatal("no announcement yet")
	}
	e.Message("Back soon", -1)
	if got := e.Operator().Announcement; got != "Back soon" {
		t.Fatalf("announcement = %q", got)
	}
	e.Back(-1)
	if got := e.Operator().Announcement; got != "" {
		t.Errorf("undone, but the operator still sees %q", got)
	}
	e.Message("Again", -1)
	mustNext(t, e)
	if got := e.Operator().Announcement; got != "" {
		t.Errorf("the next line replaced it, but the operator still sees %q", got)
	}
}

func TestAdHocCommandsAreUndoable(t *testing.T) {
	e, _ := newEngine(t, "A")
	mustNext(t, e)
	e.SetBlank(true, -1)
	if e.Snapshot().Mode != ModeBlank {
		t.Fatal("blank")
	}
	e.Message("**Hold** please", -1)
	if st := e.Snapshot(); st.Fullscreen == nil || st.Fullscreen.HTML != "<strong>Hold</strong> please" {
		t.Fatalf("%+v", st)
	}
	e.Back(-1)
	if e.Snapshot().Mode != ModeBlank {
		t.Error("back should return to blank")
	}
	e.Back(-1)
	if e.Snapshot().Mode != ModeCaptions || len(e.Snapshot().Scrollback) != 1 {
		t.Errorf("%+v", e.Snapshot())
	}
	e.Clear(-1)
	if len(e.Snapshot().Scrollback) != 0 {
		t.Error("clear")
	}
}

func TestScrollbackBounded(t *testing.T) {
	src := ""
	for i := 0; i < 10; i++ {
		src += fmt.Sprintf("line %d\n", i)
	}
	r := &rec{}
	e := New(parse(src), Config{ScrollbackMax: 4}, r.sink)
	defer e.Close()
	for i := 0; i < 10; i++ {
		mustNext(t, e)
	}
	got := texts(e.Snapshot())
	if len(got) != 4 || got[0] != "line 6" || got[3] != "line 9" {
		t.Errorf("%v", got)
	}
}

func TestSeqIncreasesPerEvent(t *testing.T) {
	e, r := newEngine(t, "A\nB\n[clear]")
	for i := 0; i < 3; i++ {
		mustNext(t, e)
	}
	e.Back(-1)
	var prev int64
	for _, ev := range r.ev {
		if ev.Seq != prev+1 {
			t.Fatalf("seq %d after %d", ev.Seq, prev)
		}
		prev = ev.Seq
	}
	if e.Snapshot().Seq != prev {
		t.Errorf("snapshot seq %d != %d", e.Snapshot().Seq, prev)
	}
}

// Jumping to N must give the same audience state as sending cues one by one.
func TestJumpEqualsSequential(t *testing.T) {
	rng := rand.New(rand.NewSource(1))
	pieces := []string{"line\n", "[clear]\n", "[blank]\n", "[image: a.png]\n", "[fullscreen: FS]\n",
		"[section: s]\n", "[note: n]\n", "[loop: 5s]\nx\ny\n[/loop]\n"}
	for trial := 0; trial < 200; trial++ {
		src := ""
		for i := 0; i < 1+rng.Intn(12); i++ {
			src += pieces[rng.Intn(len(pieces))]
			src += fmt.Sprintf("caption %d\n", i)
		}
		cues := parse(src)
		seq := New(cues, Config{}, nil)
		jmp := New(cues, Config{}, nil)
		for n := 0; n <= len(cues); n++ {
			if n > 0 {
				if _, err := seq.Next(-1); err != nil && err != ErrEnd {
					t.Fatal(err)
				}
			}
			if seq.Operator().Cursor < n && n > 0 && seq.Operator().NextIndex == -1 {
				// sequential ran out of audience-visible cues earlier than n
			}
			c := seq.Operator().Cursor
			jmp.Jump(c, -1)
			a, b := seq.Snapshot(), jmp.Snapshot()
			a.Seq, b.Seq = 0, 0
			stripIDs(&a)
			stripIDs(&b)
			la, lb := seq.Operator().Loop != nil, jmp.Operator().Loop != nil
			if !reflect.DeepEqual(a, b) || la != lb {
				t.Fatalf("trial %d jump(%d) differs:\nseq  %+v loop=%v\njump %+v loop=%v\n%s", trial, c, a, la, b, lb, src)
			}
		}
		seq.Close()
		jmp.Close()
	}
}

func stripIDs(s *State) {
	for i := range s.Scrollback {
		s.Scrollback[i].ID = 0
	}
}

func TestJumpEmitsSingleSnapshotAndIsUndoable(t *testing.T) {
	e, r := newEngine(t, "A\nB\nC\nD")
	before := len(r.types())
	if _, err := e.Jump(3, -1); err != nil {
		t.Fatal(err)
	}
	if got := r.types()[before:]; !reflect.DeepEqual(got, []string{EvSnapshot}) {
		t.Errorf("events %v", got)
	}
	if got := texts(e.Snapshot()); !reflect.DeepEqual(got, []string{"A", "B", "C"}) {
		t.Errorf("%v", got)
	}
	e.Back(-1)
	if len(e.Snapshot().Scrollback) != 0 {
		t.Error("back should undo jump")
	}
	if _, err := e.Jump(99, -1); err != ErrBadCue {
		t.Errorf("err = %v", err)
	}
}

func TestResetAndLiveFlag(t *testing.T) {
	e, _ := newEngine(t, "[loop]\nx\n[/loop]\nA")
	if e.Operator().Live {
		t.Fatal("auto-started loop alone is not live")
	}
	mustNext(t, e)
	if !e.Operator().Live {
		t.Fatal("should be live")
	}
	op, _ := e.Reset(-1)
	if op.Live || op.Loop == nil || len(e.Snapshot().Scrollback) != 0 {
		t.Errorf("%+v", op)
	}
}

func TestPersistRoundTrip(t *testing.T) {
	cues := parse(loopScript)
	a := New(cues, Config{}, nil)
	defer a.Close()
	a.Next(-1)
	a.Next(-1) // intermission loop running
	p := a.Export()
	b := New(cues, Config{}, nil)
	defer b.Close()
	b.Restore(p)
	oa, ob := a.Operator(), b.Operator()
	if ob.Cursor != oa.Cursor || ob.Rev != oa.Rev || ob.Loop == nil || ob.Loop.Cue != 2 || ob.Loop.Item != 0 {
		t.Errorf("%+v vs %+v", oa, ob)
	}
	if got := texts(b.Snapshot()); !reflect.DeepEqual(got, []string{"A"}) {
		t.Errorf("%v", got)
	}
}

func TestReplaceCuesKeepsCursor(t *testing.T) {
	e, _ := newEngine(t, "A\nB\nC")
	mustNext(t, e)
	if !e.ReplaceCues(parse("A\nB\nC\nD")) {
		t.Error("count change not reported")
	}
	if e.Operator().Cursor != 1 || e.Operator().Total != 4 {
		t.Errorf("%+v", e.Operator())
	}
	e.ReplaceCues(parse("A"))
	if e.Operator().Cursor != 1 {
		t.Errorf("%+v", e.Operator())
	}
}

func TestBackStillWorksAfterARestart(t *testing.T) {
	cues := parse("A\nB\nC\nD")
	a := New(cues, Config{}, nil)
	defer a.Close()
	for i := 0; i < 3; i++ {
		mustNext(t, a)
	}
	b := New(cues, Config{}, nil)
	defer b.Close()
	b.Restore(a.Export())
	if got := texts(b.Snapshot()); !reflect.DeepEqual(got, []string{"A", "B", "C"}) {
		t.Fatalf("restored screen: %v", got)
	}
	if _, err := b.Back(-1); err != nil {
		t.Fatalf("Back after restart: %v", err)
	}
	if got := texts(b.Snapshot()); !reflect.DeepEqual(got, []string{"A", "B"}) || b.Operator().Cursor != 2 {
		t.Errorf("after Back: %v cursor %d", got, b.Operator().Cursor)
	}
	// Only the most recent steps are kept.
	long := parse(strings.Repeat("x\n", 40))
	c := New(long, Config{}, nil)
	defer c.Close()
	for i := 0; i < 30; i++ {
		mustNext(t, c)
	}
	if n := len(c.Export().History); n != persistedHistory {
		t.Errorf("history entries saved: %d, want %d", n, persistedHistory)
	}
	// A show file that shrank must not restore steps that point past its end.
	d := New(parse("A"), Config{}, nil)
	defer d.Close()
	d.Restore(c.Export())
	if len(d.Export().History) != 0 {
		t.Errorf("stale history kept: %d", len(d.Export().History))
	}
}

func TestLastActionTimeIsRecordedAndSurvivesRestart(t *testing.T) {
	a := New(parse("A\nB"), Config{}, nil)
	defer a.Close()
	if a.Operator().LastAt != 0 {
		t.Error("no action yet")
	}
	before := time.Now().UnixMilli()
	mustNext(t, a)
	at := a.Operator().LastAt
	if at < before || at > time.Now().UnixMilli() {
		t.Errorf("lastAt %d", at)
	}
	b := New(parse("A\nB"), Config{}, nil)
	defer b.Close()
	b.Restore(a.Export())
	if b.Operator().LastAt != at {
		t.Errorf("lastAt after restore: %d vs %d", b.Operator().LastAt, at)
	}
}

func TestConfigureChangesLoopTimeAndScrollbackOfARunningEngine(t *testing.T) {
	e := New(parse("[loop]\nx\n[/loop]\nA\n[loop]\ny\nz\n[/loop]\nB\nC\nD\nE"), Config{DefaultLoopSeconds: 60, ScrollbackMax: 10}, nil)
	defer e.Close()
	e.Configure(Config{DefaultLoopSeconds: 0.03}) // applies to loops started from now on
	mustNext(t, e)                                // A (ends the 60 s pre-show loop)
	mustNext(t, e)                                // starts the second loop on the new, short time
	waitFor(t, func() bool { return e.Operator().Loop != nil && e.Operator().Loop.Item == 1 })
	for i := 0; i < 4; i++ {
		mustNext(t, e)
	}
	e.Configure(Config{ScrollbackMax: 2})
	if got := texts(e.Snapshot()); len(got) != 2 || got[1] != "E" {
		t.Errorf("scrollback should shrink at once: %v", got)
	}
	e.Configure(Config{}) // zero values change nothing
	if got := texts(e.Snapshot()); len(got) != 2 {
		t.Errorf("%v", got)
	}
}

// An intermission timer counts from the moment the loop is entered (again each time), a time of day does not
// move, and the deadline travels on every item that shows it.
func TestLoopTimerDeadline(t *testing.T) {
	now := time.Date(2026, 10, 16, 20, 0, 0, 0, time.Local)
	clock := now
	r := &rec{}
	src := "Hello\n[loop: timer 15]\nINTERMISSION\nBack in {timer}\n[/loop]\nWelcome back\n"
	e := New(parse(src), Config{DefaultLoopSeconds: 60, Now: func() time.Time { return clock }}, r.sink)
	t.Cleanup(e.Close)
	if _, err := e.Next(-1); err != nil { // Hello
		t.Fatal(err)
	}
	clock = now.Add(2 * time.Minute)
	st, err := e.Next(-1) // enter the loop
	if err != nil || st.Loop == nil {
		t.Fatalf("%+v %v", st, err)
	}
	want := clock.Add(15 * time.Minute).UnixMilli()
	if st.Loop.Deadline != want {
		t.Errorf("deadline %d want %d", st.Loop.Deadline, want)
	}
	// the timer item carries it, the plain item does not
	if it := e.loopItemFor(0); it.Deadline != 0 {
		t.Errorf("plain item has a deadline: %+v", it)
	}
	if it := e.loopItemFor(1); it.Deadline != want || !it.Timer {
		t.Errorf("timer item: %+v", it)
	}
	// leave it (a caption) and come back with Back: the countdown starts again from that moment
	e.Next(-1)
	clock = now.Add(10 * time.Minute)
	st, _ = e.Back(-1)
	if st.Loop == nil || st.Loop.Deadline != clock.Add(15*time.Minute).UnixMilli() {
		t.Errorf("restart: %+v", st.Loop)
	}
}

func TestLoopTimerClockTimeAndSavingTheShowFile(t *testing.T) {
	now := time.Date(2026, 10, 16, 18, 50, 0, 0, time.Local)
	r := &rec{}
	src := "[loop: timer 7:30pm]\nStarting in {timer}\n[/loop]\nHello\n"
	e := New(parse(src), Config{Now: func() time.Time { return now }}, r.sink)
	t.Cleanup(e.Close)
	want := time.Date(2026, 10, 16, 19, 30, 0, 0, time.Local).UnixMilli()
	if st := e.Operator(); st.Loop == nil || st.Loop.Deadline != want {
		t.Fatalf("a leading loop starts with its time of day: %+v", st.Loop)
	}
	// An edit saved while the loop runs keeps the minutes countdown going (the clock one never moved).
	e2 := New(parse("[loop: timer 15]\nBack in {timer}\n[/loop]\nHello\n"), Config{Now: func() time.Time { return now }}, r.sink)
	t.Cleanup(e2.Close)
	first := e2.Operator().Loop.Deadline
	now = now.Add(5 * time.Minute)
	e2.ReplaceCues(parse("[loop: timer 15]\nBack in {timer}, folks\n[/loop]\nHello\n"))
	if got := e2.Operator().Loop.Deadline; got != first {
		t.Errorf("saving the show file restarted the countdown: %d -> %d", first, got)
	}
	e2.ReplaceCues(parse("[loop: timer 20]\nBack in {timer}\n[/loop]\nHello\n"))
	if got := e2.Operator().Loop.Deadline; got == first {
		t.Error("a different timer should start afresh")
	}
}

func TestLoopTimerSurvivesARestart(t *testing.T) {
	now := time.Date(2026, 10, 16, 20, 0, 0, 0, time.Local)
	src := "[loop: timer 15]\nBack in {timer}\n[/loop]\nHello\n"
	e := New(parse(src), Config{Now: func() time.Time { return now }}, nil)
	t.Cleanup(e.Close)
	p := e.Export()
	if p.LoopDeadline == 0 {
		t.Fatal("deadline not saved")
	}
	now = now.Add(4 * time.Minute)
	e2 := New(parse(src), Config{Now: func() time.Time { return now }}, nil)
	t.Cleanup(e2.Close)
	e2.Restore(p)
	if got := e2.Operator().Loop.Deadline; got != p.LoopDeadline {
		t.Errorf("after restart: %d want %d", got, p.LoopDeadline)
	}
}

// loopItemFor returns item i of the running loop as the audience would be sent it.
func (e *Engine) loopItemFor(i int) showfile.Item {
	e.mu.Lock()
	defer e.mu.Unlock()
	return *e.loopItem(e.loop.cue, i)
}

// ---- standby: the show is loaded but not on air ----

func newStandby(t *testing.T, src string) (*Engine, *rec) {
	t.Helper()
	r := &rec{}
	e := New(parse(src), Config{DefaultLoopSeconds: 60, Standby: true}, r.sink)
	t.Cleanup(e.Close)
	return e, r
}

func TestStandbyShowsNothingUntilStart(t *testing.T) {
	e, r := newStandby(t, "[loop: 5s]\nWelcome\n[/loop]\nHello\nWorld\n")
	st := e.Snapshot()
	if st.Mode != ModeStandby || st.Fullscreen != nil || len(st.Scrollback) != 0 {
		t.Fatalf("snapshot in standby: %+v", st)
	}
	o := e.Operator()
	if o.Live || !o.Standby || o.Loop != nil {
		t.Fatalf("operator state: %+v", o)
	}
	if o.NextIndex != 0 || o.Cursor != 1 {
		t.Errorf("START will run the pre-show loop, so that is what is next: %+v", o)
	}
	for name, fn := range map[string]func() error{
		"next":  func() error { _, err := e.Next(-1); return err },
		"back":  func() error { _, err := e.Back(-1); return err },
		"jump":  func() error { _, err := e.Jump(2, -1); return err },
		"blank": func() error { _, err := e.SetBlank(true, -1); return err },
		"clear": func() error { _, err := e.Clear(-1); return err },
		"msg":   func() error { _, err := e.Message("hi", -1); return err },
	} {
		if err := fn(); err != ErrStandby {
			t.Errorf("%s in standby: %v", name, err)
		}
	}
	if len(r.types()) != 0 {
		t.Errorf("nothing may be sent to the audience in standby: %v", r.types())
	}
	// Start: on air, the pre-show loop begins, and the first Next sends the first line.
	o, err := e.Start(-1)
	if err != nil || !o.Live || o.Standby || o.Loop == nil {
		t.Fatalf("start: %+v %v", o, err)
	}
	if st := e.Snapshot(); st.Mode != ModeCaptions || st.Fullscreen == nil || !strings.Contains(st.Fullscreen.Text, "Welcome") {
		t.Errorf("on air: %+v", st)
	}
	if o, _ = e.Next(-1); o.Cursor != 2 {
		t.Errorf("next: %+v", o)
	}
	if st := e.Snapshot(); len(st.Scrollback) != 1 || st.Scrollback[0].Text != "Hello" {
		t.Errorf("scrollback: %+v", st.Scrollback)
	}
}

func TestStandbyWithoutALeadingLoop(t *testing.T) {
	e, _ := newStandby(t, "Hello\nWorld\n")
	if st := e.Snapshot(); st.Mode != ModeStandby {
		t.Fatalf("%+v", st)
	}
	e.Start(-1)
	if st := e.Snapshot(); st.Mode != ModeIdle { // the show's own message, until the first line
		t.Errorf("on air before the first line: %+v", st)
	}
	if o := e.Operator(); !o.Live || o.Cursor != 0 || o.NextIndex != 0 {
		t.Errorf("%+v", o)
	}
}

func TestOffAirKeepsThePositionAndResumes(t *testing.T) {
	e, r := newStandby(t, "Hello\n[loop: 5s]\nINTERMISSION\n[/loop]\nWorld\n")
	e.Start(-1)
	e.Next(-1)
	e.Next(-1) // into the loop
	if o := e.Operator(); o.Loop == nil {
		t.Fatalf("loop should be running: %+v", o)
	}
	before := len(r.types())
	o, err := e.OffAir(-1)
	if err != nil || o.Live || !o.Standby || o.Loop != nil || o.Cursor != 2 {
		t.Fatalf("off air: %+v %v", o, err)
	}
	if st := e.Snapshot(); st.Mode != ModeStandby || len(st.Scrollback) != 0 || st.Fullscreen != nil {
		t.Errorf("the audience must see nothing: %+v", st)
	}
	if got := r.types()[before:]; len(got) != 1 || got[0] != EvSnapshot {
		t.Errorf("events off air: %v", got)
	}
	if _, err := e.Next(-1); err != ErrStandby {
		t.Errorf("next off air: %v", err)
	}
	o, _ = e.Start(-1)
	if !o.Live || o.Cursor != 2 || o.Loop == nil {
		t.Errorf("back on air, loop running again: %+v", o)
	}
	if st := e.Snapshot(); len(st.Scrollback) != 1 || st.Scrollback[0].Text != "Hello" {
		t.Errorf("the lines are still there: %+v", st.Scrollback)
	}
	// Going off air twice, or starting twice, changes nothing.
	e.Start(-1)
	e.OffAir(-1)
	if o, _ = e.OffAir(-1); !o.Standby {
		t.Error("still off air")
	}
}

func TestResetReturnsToStandbyAndStandbySurvivesARestart(t *testing.T) {
	src := "[loop: 5s]\nWelcome\n[/loop]\nHello\n"
	e, _ := newStandby(t, src)
	e.Start(-1)
	e.Next(-1)
	if o, _ := e.Reset(-1); !o.Standby || o.Live || o.Cursor != 1 {
		t.Errorf("reset: %+v", o)
	}
	if st := e.Snapshot(); st.Mode != ModeStandby || len(st.Scrollback) != 0 {
		t.Errorf("%+v", st)
	}
	p := e.Export()
	e2, _ := newStandby(t, src)
	e2.Restore(p)
	if o := e2.Operator(); !o.Standby {
		t.Errorf("a show in standby stays in standby after a restart: %+v", o)
	}
	e2.Start(-1)
	if o := e2.Operator(); o.Loop == nil {
		t.Errorf("and its pre-show still starts: %+v", o)
	}
	// on air stays on air
	e2.Next(-1)
	p = e2.Export()
	e3, _ := newStandby(t, src)
	e3.Restore(p)
	if o := e3.Operator(); o.Standby || !o.Live {
		t.Errorf("on air stays on air: %+v", o)
	}
}

// There is nothing to restart in standby before anything was sent (a fresh show, or one just restarted).
func TestAtStartMeansNothingToRestart(t *testing.T) {
	e, _ := newStandby(t, "[loop: 5s]\nWelcome\n[/loop]\nHello\nWorld\n")
	if !e.Operator().AtStart {
		t.Error("a freshly loaded show is at the start")
	}
	e.Start(-1)
	if e.Operator().AtStart {
		t.Error("on air is not standby")
	}
	e.Next(-1)
	e.OffAir(-1)
	if o := e.Operator(); !o.Standby || o.AtStart {
		t.Errorf("off air with lines sent can be restarted: %+v", o)
	}
	e.Reset(-1)
	if !e.Operator().AtStart {
		t.Error("just restarted: back at the start")
	}
	// started and taken straight off air with nothing sent: still nothing to restart
	e.Start(-1)
	e.OffAir(-1)
	if !e.Operator().AtStart {
		t.Error("nothing was sent, so nothing to restart")
	}
}

func TestAlignPositionsAStandbyEngineWithoutGoingOnAir(t *testing.T) {
	e, _ := newStandby(t, "One\nTwo\nThree\n")
	if _, err := e.Jump(2, -1); err != ErrStandby {
		t.Fatalf("Jump in standby: %v, want ErrStandby", err)
	}
	if err := e.Align(2); err != nil {
		t.Fatal(err)
	}
	op := e.Operator()
	if op.Cursor != 2 || op.Live || !op.Standby {
		t.Fatalf("after Align: cursor %d, live %v, standby %v; want 2, false, true", op.Cursor, op.Live, op.Standby)
	}
	if s := e.Snapshot(); s.Mode != ModeStandby || len(s.Scrollback) != 0 {
		t.Fatalf("the audience must still see standby: %+v", s)
	}
	if _, err := e.Start(-1); err != nil {
		t.Fatal(err)
	}
	if got := e.Operator().Cursor; got != 2 {
		t.Fatalf("after Start the cursor is %d, want 2", got)
	}
	if err := e.Align(99); err != ErrBadCue {
		t.Fatalf("Align past the end: %v", err)
	}
}

func TestItemDurationIsCappedSoATimerCannotOverflow(t *testing.T) {
	e := New([]showfile.Cue{{Kind: showfile.KindLoop, Period: 1e12, Items: []showfile.Item{{Kind: "text", Seconds: 1e12}}}},
		Config{DefaultLoopSeconds: 60}, func(Event) {})
	defer e.Close()
	if d := e.itemDuration(0, 0); d != time.Hour {
		t.Fatalf("duration %v, want %v", d, time.Hour)
	}
}

func TestSkipPassesTheNextLineWithoutSendingIt(t *testing.T) {
	e, r := newEngine(t, "A\n[note: n]\nB\nC\n")
	mustNext(t, e)
	before := e.Operator()
	op, err := e.Skip(-1)
	if err != nil {
		t.Fatal(err)
	}
	if op.Cursor != 3 || op.NextIndex != 3 || op.Rev == before.Rev || op.LastAt != before.LastAt {
		t.Errorf("after the skip: %+v (before %+v)", op, before)
	}
	if got := texts(e.Snapshot()); !reflect.DeepEqual(got, []string{"A"}) {
		t.Errorf("the skipped line was shown: %v", got)
	}
	mustNext(t, e)
	if got := texts(e.Snapshot()); !reflect.DeepEqual(got, []string{"A", "C"}) {
		t.Errorf("the line after the skipped one comes next: %v", got)
	}
	if got := r.types(); !reflect.DeepEqual(got, []string{EvCaption, EvCaption}) {
		t.Errorf("a skip tells the audience nothing: %v", got)
	}
	// Back undoes the send, then the skip.
	e.Back(-1)
	if op, _ = e.Back(-1); op.NextIndex != 2 {
		t.Errorf("Back undoes the skip: %+v", op)
	}
	// Skipping the last line leaves the end of the show.
	e.Skip(-1)
	e.Skip(-1)
	if _, err := e.Skip(-1); err != ErrEnd {
		t.Errorf("err = %v, want ErrEnd", err)
	}
}

func TestSkipNeedsTheShowOnAir(t *testing.T) {
	e := New(parse("A\nB\n"), Config{DefaultLoopSeconds: 10, Standby: true}, func(Event) {})
	t.Cleanup(e.Close)
	if _, err := e.Skip(-1); err != ErrStandby {
		t.Errorf("err = %v, want ErrStandby", err)
	}
	if op := e.Operator(); op.NextIndex != 0 {
		t.Errorf("a refused skip moved the show: %+v", op)
	}
}

func TestSkippedLinesAreRemembered(t *testing.T) {
	e, _ := newEngine(t, "A\nB\nC\nD\nE\n")
	mustNext(t, e) // A
	op, _ := e.Skip(-1)
	if !reflect.DeepEqual(op.Skipped, []int{1}) {
		t.Fatalf("skipped %v", op.Skipped)
	}
	mustNext(t, e) // C
	e.Skip(-1)     // D
	if op = e.Operator(); !reflect.DeepEqual(op.Skipped, []int{1, 3}) {
		t.Fatalf("skipped %v", op.Skipped)
	}
	// They survive a restart.
	e2, _ := newEngine(t, "A\nB\nC\nD\nE\n")
	e2.Restore(e.Export())
	if got := e2.Operator().Skipped; !reflect.DeepEqual(got, []int{1, 3}) {
		t.Errorf("after a restart: %v", got)
	}
	// Back takes the last skip away, and going back past one makes the line unsent again.
	if op, _ = e.Back(-1); !reflect.DeepEqual(op.Skipped, []int{1}) {
		t.Errorf("Back: %v", op.Skipped)
	}
	e.Skip(-1)
	if op, _ = e.Jump(1, -1); len(op.Skipped) != 0 {
		t.Errorf("a jump back to the skipped line: %v", op.Skipped)
	}
	e.Skip(-1)
	if op, _ = e.Reset(-1); len(op.Skipped) != 0 {
		t.Errorf("Reset: %v", op.Skipped)
	}
}

// SentAt is when the line now on screen was sent: Next and Jump set it, Back brings the earlier one back, other actions
// (blank, clear) leave it, going on air restarts it, and it survives a restart of the program.
func TestSentAtIsWhenTheLineOnScreenWasSent(t *testing.T) {
	e, _ := newEngine(t, "A\nB\nC\n")
	if e.Operator().SentAt != 0 {
		t.Fatal("nothing sent yet")
	}
	pause := func() { time.Sleep(3 * time.Millisecond) }
	mustNext(t, e)
	a := e.Operator().SentAt
	if a == 0 || a != e.Operator().LastAt {
		t.Fatalf("Next sets it to the moment of the action: %d %d", a, e.Operator().LastAt)
	}
	pause()
	if _, err := e.SetBlank(true, -1); err != nil {
		t.Fatal(err)
	}
	pause()
	if o := e.Operator(); o.SentAt != a || o.LastAt <= a {
		t.Errorf("blanking is an action but does not send a line: sent %d, last %d, was %d", o.SentAt, o.LastAt, a)
	}
	pause()
	mustNext(t, e)
	b := e.Operator().SentAt
	if b <= a {
		t.Errorf("the next line is sent later: %d %d", a, b)
	}
	p := e.Export()
	e2, _ := newEngine(t, "A\nB\nC\n")
	e2.Restore(p)
	if e2.Operator().SentAt != b {
		t.Errorf("kept over a restart: %d, want %d", e2.Operator().SentAt, b)
	}
	if _, err := e2.Back(-1); err != nil {
		t.Fatal(err)
	}
	if got := e2.Operator().SentAt; got != a {
		t.Errorf("Back (even after a restart) brings back when the earlier line was sent: %d", got)
	}
	if _, err := e.Back(-1); err != nil {
		t.Fatal(err)
	}
	if _, err := e.Back(-1); err != nil {
		t.Fatal(err)
	}
	pause()
	if _, err := e.Jump(2, -1); err != nil {
		t.Fatal(err)
	}
	if j := e.Operator().SentAt; j <= b {
		t.Errorf("a jump puts a line on screen now: %d", j)
	}
	if _, err := e.Reset(-1); err != nil || e.Operator().SentAt != 0 {
		t.Errorf("Reset forgets it: %v %d", err, e.Operator().SentAt)
	}
}

func TestSentAtRestartsWhenTheShowGoesBackOnAir(t *testing.T) {
	e, _ := newStandby(t, "A\nB\n")
	if _, err := e.Start(-1); err != nil {
		t.Fatal(err)
	}
	mustNext(t, e)
	first := e.Operator().SentAt
	if _, err := e.OffAir(-1); err != nil {
		t.Fatal(err)
	}
	time.Sleep(3 * time.Millisecond)
	if _, err := e.Start(-1); err != nil {
		t.Fatal(err)
	}
	if got := e.Operator().SentAt; got <= first {
		t.Errorf("back on air, the line on screen counts from now: %d, was %d", got, first)
	}
}
