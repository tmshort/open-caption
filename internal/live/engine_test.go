package live

import (
	"fmt"
	"math/rand"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/tmshort/open-caption/internal/script"
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

func parse(src string) []script.Cue { return script.Parse(src, script.Options{}).Cues }

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
	item := r.ev[0].Data.(script.Item)
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
	// A script that shrank must not restore steps that point past its end.
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
