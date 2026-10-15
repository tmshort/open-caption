// Package live holds the authoritative live-show state: the cursor into the
// show file, what the audience currently sees, undo history, and the server-side
// loop timer. It has no network code; consumers subscribe through a sink.
// See docs/02-architecture.md ("Applying a cue").
package live

import (
	"errors"
	"sync"
	"time"

	"github.com/tmshort/open-caption/internal/showfile"
)

// Audience modes.
const (
	ModeIdle     = "idle"
	ModeCaptions = "captions"
	ModeBlank    = "blank"
)

// Event types sent to audience clients (docs/04-protocol-and-api.md).
const (
	EvSnapshot   = "snapshot"
	EvCaption    = "caption"    // appends; also dismisses fullscreen and un-blanks
	EvClear      = "clear"      // empties scrollback; dismisses fullscreen
	EvBlank      = "blank"      // {on}; "on" also dismisses fullscreen
	EvFullscreen = "fullscreen" // {item}
)

var (
	ErrStale   = errors.New("stale revision")
	ErrEnd     = errors.New("end of show file")
	ErrNoLoop  = errors.New("no loop is running")
	ErrNoUndo  = errors.New("nothing to undo")
	ErrBadCue  = errors.New("cue index out of range")
	historyMax = 100
)

// CaptionItem is one line on the scrolling surface.
type CaptionItem struct {
	ID      int    `json:"id"`
	Speaker string `json:"speaker,omitempty"`
	HTML    string `json:"html"`
	Text    string `json:"text"`
	Stage   bool   `json:"stage,omitempty"`
	Sung    bool   `json:"sung,omitempty"`
	// Parts is set when several voices are shown at the same time; HTML/Text then hold them joined.
	Parts []showfile.Part `json:"parts,omitempty"`
}

// State is the audience-visible state.
type State struct {
	Seq        int64          `json:"seq"`
	Mode       string         `json:"mode"`
	Scrollback []CaptionItem  `json:"scrollback"`
	Fullscreen *showfile.Item `json:"fullscreen"`
}

func (s State) clone() State {
	s.Scrollback = append([]CaptionItem(nil), s.Scrollback...)
	if s.Fullscreen != nil {
		f := *s.Fullscreen
		s.Fullscreen = &f
	}
	return s
}

// Event is one change published to subscribers. Seq increases by one per event.
type Event struct {
	Seq  int64
	Type string
	Data any
}

// LoopStatus describes a running loop.
type LoopStatus struct {
	Cue  int `json:"cue"`
	Item int `json:"item"`
}

// OperatorState is what the operator UI needs besides the cue list.
type OperatorState struct {
	Seq       int64       `json:"seq"`
	Rev       int64       `json:"rev"`
	Cursor    int         `json:"cursor"`    // index of the next cue (may be operator-only)
	NextIndex int         `json:"nextIndex"` // next audience-visible cue, -1 at the end
	Total     int         `json:"total"`
	Live      bool        `json:"live"`
	Mode      string      `json:"mode"`   // audience mode: idle | captions | blank
	LastAt    int64       `json:"lastAt"` // unix ms of the operator's last action (0 = none yet)
	Loop      *LoopStatus `json:"loop"`
}

// Config tunes the engine.
type Config struct {
	DefaultLoopSeconds float64
	ScrollbackMax      int
	// StartSeq is the sequence number the first event follows. A replacement
	// engine continues the previous engine's numbering so connected clients,
	// which drop events they believe they have seen, never miss any.
	StartSeq int64
}

type snapshot struct {
	st      State
	cursor  int
	loopCue int // -1: no loop was running
}

type loopRun struct {
	cue   int
	item  int
	gen   int
	timer *time.Timer
}

// Engine is safe for concurrent use. The sink is called with the engine lock
// held, so it must not block or call back into the engine.
type Engine struct {
	mu      sync.Mutex
	cues    []showfile.Cue
	cfg     Config
	sink    func(Event)
	st      State
	cursor  int
	rev     int64
	nextID  int
	gen     int
	history []snapshot
	loop    *loopRun
	touched bool  // a cue has been applied since activation/reset
	lastAt  int64 // unix ms of the last operator action (0 = none yet)
}

// New creates an engine in its initial state (auto-starting a leading loop).
func New(cues []showfile.Cue, cfg Config, sink func(Event)) *Engine {
	if cfg.DefaultLoopSeconds <= 0 {
		cfg.DefaultLoopSeconds = 10
	}
	if cfg.ScrollbackMax <= 0 {
		cfg.ScrollbackMax = 200
	}
	if sink == nil {
		sink = func(Event) {}
	}
	e := &Engine{cues: cues, cfg: cfg, sink: sink}
	e.st.Seq = cfg.StartSeq
	e.initialLocked()
	return e
}

// Close stops any running loop timer.
func (e *Engine) Close() {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.stopLoopLocked()
}

func (e *Engine) emit(typ string, data any) {
	e.st.Seq++
	e.sink(Event{Seq: e.st.Seq, Type: typ, Data: data})
}

func (e *Engine) emitSnapshot() { e.emit(EvSnapshot, e.st.clone()) }

// --- queries ---

// Snapshot returns the current audience state.
func (e *Engine) Snapshot() State {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.st.clone()
}

// Operator returns the operator-facing state.
func (e *Engine) Operator() OperatorState {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.operatorLocked()
}

func (e *Engine) operatorLocked() OperatorState {
	o := OperatorState{Seq: e.st.Seq, Rev: e.rev, Cursor: e.cursor, NextIndex: e.visibleFrom(e.cursor),
		Total: len(e.cues), Live: e.touched, Mode: e.st.Mode, LastAt: e.lastAt}
	if e.loop != nil {
		o.Loop = &LoopStatus{Cue: e.loop.cue, Item: e.loop.item}
	}
	return o
}

// Cues returns the current cue list (shared; do not modify).
func (e *Engine) Cues() []showfile.Cue {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.cues
}

// --- commands (all return the new operator state) ---

// Next sends the next audience-visible cue. rev < 0 skips the staleness check.
func (e *Engine) Next(rev int64) (OperatorState, error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	if err := e.checkLocked(rev); err != nil {
		return e.operatorLocked(), err
	}
	idx := e.visibleFrom(e.cursor)
	if idx < 0 {
		e.cursor = len(e.cues)
		return e.operatorLocked(), ErrEnd
	}
	e.pushLocked()
	e.applyLocked(e.cues[idx], idx)
	e.cursor = idx + 1
	e.touched = true
	e.bumpLocked()
	return e.operatorLocked(), nil
}

// Back undoes the last operator action. A loop that was running restarts at item 0.
func (e *Engine) Back(rev int64) (OperatorState, error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	if err := e.checkLocked(rev); err != nil {
		return e.operatorLocked(), err
	}
	n := len(e.history)
	if n == 0 {
		return e.operatorLocked(), ErrNoUndo
	}
	s := e.history[n-1]
	e.history = e.history[:n-1]
	e.stopLoopLocked()
	seq := e.st.Seq
	e.st = s.st.clone()
	e.st.Seq = seq
	e.cursor = s.cursor
	e.touched = e.cursor > e.initialCursor()
	if s.loopCue >= 0 && s.loopCue < len(e.cues) {
		e.beginLoopLocked(s.loopCue, false)
	}
	e.bumpLocked()
	e.emitSnapshot()
	return e.operatorLocked(), nil
}

// Stop ends a running loop and leaves a blank screen.
func (e *Engine) Stop(rev int64) (OperatorState, error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	if err := e.checkLocked(rev); err != nil {
		return e.operatorLocked(), err
	}
	if e.loop == nil {
		return e.operatorLocked(), ErrNoLoop
	}
	e.pushLocked()
	e.stopLoopLocked()
	e.st.Mode, e.st.Fullscreen = ModeBlank, nil
	e.emit(EvBlank, map[string]bool{"on": true})
	e.bumpLocked()
	return e.operatorLocked(), nil
}

// Jump sets the cursor to index, rebuilding the resulting state without
// replaying intermediate cues to clients (a single snapshot is sent).
func (e *Engine) Jump(index int, rev int64) (OperatorState, error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	if err := e.checkLocked(rev); err != nil {
		return e.operatorLocked(), err
	}
	if index < 0 || index > len(e.cues) {
		return e.operatorLocked(), ErrBadCue
	}
	e.pushLocked()
	st, cursor, loopCue := e.replayLocked(index)
	e.stopLoopLocked()
	seq := e.st.Seq
	e.st, e.cursor = st, cursor
	e.st.Seq = seq
	if loopCue >= 0 {
		e.beginLoopLocked(loopCue, false)
	}
	e.touched = cursor > e.initialCursor()
	e.bumpLocked()
	e.emitSnapshot()
	return e.operatorLocked(), nil
}

// SetBlank blanks or un-blanks the audience screen (ad hoc).
func (e *Engine) SetBlank(on bool, rev int64) (OperatorState, error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	if err := e.checkLocked(rev); err != nil {
		return e.operatorLocked(), err
	}
	if on == (e.st.Mode == ModeBlank) {
		return e.operatorLocked(), nil
	}
	e.pushLocked()
	e.stopLoopLocked()
	if on {
		e.st.Mode, e.st.Fullscreen = ModeBlank, nil
	} else {
		e.st.Mode = ModeCaptions
	}
	e.emit(EvBlank, map[string]bool{"on": on})
	e.touched = true
	e.bumpLocked()
	return e.operatorLocked(), nil
}

// Clear empties the scrolling captions (ad hoc).
func (e *Engine) Clear(rev int64) (OperatorState, error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	if err := e.checkLocked(rev); err != nil {
		return e.operatorLocked(), err
	}
	e.pushLocked()
	e.applyLocked(showfile.Cue{Kind: showfile.KindClear}, -1)
	e.touched = true
	e.bumpLocked()
	return e.operatorLocked(), nil
}

// Message shows one-off fullscreen text (ad hoc) until the next cue.
func (e *Engine) Message(text string, rev int64) (OperatorState, error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	if err := e.checkLocked(rev); err != nil {
		return e.operatorLocked(), err
	}
	h, p := showfile.RenderInline(text)
	e.pushLocked()
	e.applyLocked(showfile.Cue{Kind: showfile.KindFullscreen, Item: &showfile.Item{Kind: "text", HTML: h, Text: p}}, -1)
	e.touched = true
	e.bumpLocked()
	return e.operatorLocked(), nil
}

// Reset returns to the initial state (not live), auto-starting a leading loop.
func (e *Engine) Reset(rev int64) (OperatorState, error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	if err := e.checkLocked(rev); err != nil {
		return e.operatorLocked(), err
	}
	e.stopLoopLocked()
	e.history = nil
	e.st.Scrollback = nil
	e.initialLocked()
	e.bumpLocked()
	e.emitSnapshot()
	return e.operatorLocked(), nil
}

// ReplaceCues swaps in an edited show file during a live show. The cursor is
// kept by index (clamped); history is dropped. Reports whether the count changed.
func (e *Engine) ReplaceCues(cues []showfile.Cue) (countChanged bool) {
	e.mu.Lock()
	defer e.mu.Unlock()
	countChanged = len(cues) != len(e.cues)
	e.cues = cues
	e.history = nil
	if e.cursor > len(cues) {
		e.cursor = len(cues)
	}
	if e.loop != nil {
		lc := e.loop.cue
		e.stopLoopLocked()
		if lc < len(cues) && cues[lc].Kind == showfile.KindLoop {
			e.beginLoopLocked(lc, false)
		} else {
			e.st.Fullscreen = nil
		}
	}
	e.bumpLocked()
	e.emitSnapshot()
	return countChanged
}

// bumpLocked marks that an operator command was applied.
func (e *Engine) bumpLocked() {
	e.rev++
	e.lastAt = time.Now().UnixMilli()
}

func (e *Engine) checkLocked(rev int64) error {
	if rev >= 0 && rev != e.rev {
		return ErrStale
	}
	return nil
}

// --- internals ---

func (e *Engine) visibleFrom(i int) int {
	for ; i < len(e.cues); i++ {
		if !e.cues[i].Kind.OperatorOnly() {
			return i
		}
	}
	return -1
}

func (e *Engine) pushLocked() {
	lc := -1
	if e.loop != nil {
		lc = e.loop.cue
	}
	e.history = append(e.history, snapshot{st: e.st.clone(), cursor: e.cursor, loopCue: lc})
	if len(e.history) > historyMax {
		e.history = e.history[len(e.history)-historyMax:]
	}
}

// initialLocked sets the initial state: idle, or a running loop if the first
// audience-visible cue is a loop.
func (e *Engine) initialLocked() {
	st, cursor, loopCue := e.replayLocked(0)
	seq := e.st.Seq
	e.st, e.cursor, e.touched = st, cursor, false
	e.st.Seq = seq
	if loopCue >= 0 {
		e.beginLoopLocked(loopCue, false)
	}
}

func (e *Engine) initialCursor() int {
	if first := e.visibleFrom(0); first >= 0 && e.cues[first].Kind == showfile.KindLoop {
		return first + 1
	}
	return 0
}

// replayLocked computes the state with cues before `index` applied, without
// emitting events or touching timers. It returns the state, the cursor, and
// the loop cue that would be running (-1 if none).
func (e *Engine) replayLocked(index int) (State, int, int) {
	st := State{Mode: ModeIdle, Scrollback: []CaptionItem{}}
	cursor, loopCue := 0, -1
	first := e.visibleFrom(0)
	if first >= 0 && e.cues[first].Kind == showfile.KindLoop {
		loopCue, cursor = first, first+1
		st.Mode = ModeCaptions
		st.Fullscreen = e.loopItem(first, 0)
	}
	if index < cursor {
		index = cursor
	}
	for i := cursor; i < index && i < len(e.cues); i++ {
		c := e.cues[i]
		if c.Kind.OperatorOnly() {
			continue
		}
		if c.Kind == showfile.KindLoop {
			loopCue = i
			st.Mode, st.Fullscreen = ModeCaptions, e.loopItem(i, 0)
			continue
		}
		loopCue = -1
		e.applyTo(&st, c, nil)
	}
	if index > len(e.cues) {
		index = len(e.cues)
	}
	return st, index, loopCue
}

// applyLocked applies one cue for real: ends any loop, mutates state, emits.
func (e *Engine) applyLocked(c showfile.Cue, idx int) {
	e.stopLoopLocked()
	if c.Kind == showfile.KindLoop {
		e.beginLoopLocked(idx, true)
		return
	}
	e.applyTo(&e.st, c, e.emit)
}

// applyTo applies a non-loop cue to st; emit may be nil (replay).
func (e *Engine) applyTo(st *State, c showfile.Cue, emit func(string, any)) {
	out := func(t string, d any) {
		if emit != nil {
			emit(t, d)
		}
	}
	switch c.Kind {
	case showfile.KindCaption:
		e.nextID++
		it := CaptionItem{ID: e.nextID, Speaker: c.Speaker, HTML: c.HTML, Text: c.Text, Stage: c.Stage, Sung: c.Sung, Parts: c.Parts}
		st.Mode, st.Fullscreen = ModeCaptions, nil
		st.Scrollback = append(st.Scrollback, it)
		if over := len(st.Scrollback) - e.cfg.ScrollbackMax; over > 0 {
			st.Scrollback = append([]CaptionItem(nil), st.Scrollback[over:]...)
		}
		out(EvCaption, it)
	case showfile.KindClear:
		st.Mode, st.Fullscreen, st.Scrollback = ModeCaptions, nil, []CaptionItem{}
		out(EvClear, nil)
	case showfile.KindBlank:
		st.Mode, st.Fullscreen = ModeBlank, nil
		out(EvBlank, map[string]bool{"on": true})
	case showfile.KindImage, showfile.KindFullscreen:
		if c.Item == nil {
			return
		}
		it := *c.Item
		st.Mode, st.Fullscreen = ModeCaptions, &it
		out(EvFullscreen, it)
	}
}

// --- loop ---

func (e *Engine) loopItem(cueIdx, item int) *showfile.Item {
	it := e.cues[cueIdx].Items[item]
	return &it
}

func (e *Engine) itemDuration(cueIdx, item int) time.Duration {
	c := e.cues[cueIdx]
	secs := c.Items[item].Seconds
	if secs <= 0 {
		secs = c.Period
	}
	if secs <= 0 {
		secs = e.cfg.DefaultLoopSeconds
	}
	return time.Duration(secs * float64(time.Second))
}

// beginLoopLocked starts the loop at item 0. If announce, the first item is
// published as a fullscreen event (otherwise the caller sends a snapshot).
func (e *Engine) beginLoopLocked(cueIdx int, announce bool) {
	if cueIdx < 0 || cueIdx >= len(e.cues) || len(e.cues[cueIdx].Items) == 0 {
		return
	}
	e.gen++
	e.loop = &loopRun{cue: cueIdx, gen: e.gen}
	e.st.Mode, e.st.Fullscreen = ModeCaptions, e.loopItem(cueIdx, 0)
	if announce {
		e.emit(EvFullscreen, *e.st.Fullscreen)
	}
	e.scheduleLocked()
}

func (e *Engine) scheduleLocked() {
	l := e.loop
	gen := l.gen
	l.timer = time.AfterFunc(e.itemDuration(l.cue, l.item), func() { e.tick(gen) })
}

func (e *Engine) tick(gen int) {
	e.mu.Lock()
	defer e.mu.Unlock()
	l := e.loop
	if l == nil || l.gen != gen {
		return
	}
	if n := len(e.cues[l.cue].Items); n > 1 {
		l.item = (l.item + 1) % n
		e.st.Fullscreen = e.loopItem(l.cue, l.item)
		e.emit(EvFullscreen, *e.st.Fullscreen)
	}
	e.scheduleLocked()
}

func (e *Engine) stopLoopLocked() {
	if e.loop != nil {
		if e.loop.timer != nil {
			e.loop.timer.Stop()
		}
		e.loop = nil
	}
}

// --- persistence ---

// Persisted is the engine state saved across restarts. History is not saved.
type Persisted struct {
	Cursor     int            `json:"cursor"`
	Rev        int64          `json:"rev"`
	NextID     int            `json:"nextId"`
	Seq        int64          `json:"seq"`
	Mode       string         `json:"mode"`
	Scrollback []CaptionItem  `json:"scrollback"`
	Fullscreen *showfile.Item `json:"fullscreen,omitempty"`
	LoopCue    int            `json:"loopCue"` // -1: none
	Touched    bool           `json:"touched"`
	LastAt     int64          `json:"lastAt,omitempty"`
	// History holds the most recent undo steps so Back still works after a restart.
	History []HistoryEntry `json:"history,omitempty"`
}

// HistoryEntry is one saved undo step.
type HistoryEntry struct {
	State   State `json:"state"`
	Cursor  int   `json:"cursor"`
	LoopCue int   `json:"loopCue"`
}

// persistedHistory is how many undo steps survive a restart (kept small: each holds the screen's captions).
const persistedHistory = 10

// Export captures the state to save.
func (e *Engine) Export() Persisted {
	e.mu.Lock()
	defer e.mu.Unlock()
	p := Persisted{Cursor: e.cursor, Rev: e.rev, NextID: e.nextID, Seq: e.st.Seq, Mode: e.st.Mode,
		Scrollback: append([]CaptionItem(nil), e.st.Scrollback...), Fullscreen: e.st.Fullscreen,
		LoopCue: -1, Touched: e.touched, LastAt: e.lastAt}
	from := max(0, len(e.history)-persistedHistory)
	for _, h := range e.history[from:] {
		p.History = append(p.History, HistoryEntry{State: h.st.clone(), Cursor: h.cursor, LoopCue: h.loopCue})
	}
	if e.loop != nil {
		p.LoopCue = e.loop.cue
	}
	return p
}

// Restore loads saved state. A saved loop restarts from its first item.
func (e *Engine) Restore(p Persisted) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.stopLoopLocked()
	e.history = nil
	e.lastAt = p.LastAt
	for _, h := range p.History {
		if h.Cursor >= 0 && h.Cursor <= len(e.cues) && h.LoopCue < len(e.cues) {
			e.history = append(e.history, snapshot{st: h.State.clone(), cursor: h.Cursor, loopCue: h.LoopCue})
		}
	}
	if p.Cursor < 0 || p.Cursor > len(e.cues) {
		p.Cursor = len(e.cues)
	}
	e.cursor, e.rev, e.nextID, e.touched = p.Cursor, p.Rev, p.NextID, p.Touched
	e.st = State{Seq: p.Seq + 1, Mode: p.Mode, Scrollback: append([]CaptionItem{}, p.Scrollback...), Fullscreen: p.Fullscreen}
	if e.st.Mode == "" {
		e.st.Mode = ModeIdle
	}
	if p.LoopCue >= 0 && p.LoopCue < len(e.cues) && e.cues[p.LoopCue].Kind == showfile.KindLoop {
		e.beginLoopLocked(p.LoopCue, false)
	}
	e.emitSnapshot()
}
