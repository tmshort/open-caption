package httpapi

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/url"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/tmshort/open-caption/internal/config"
	"github.com/tmshort/open-caption/internal/hub"
	"github.com/tmshort/open-caption/internal/live"
	"github.com/tmshort/open-caption/internal/script"
)

// Theme is the venue look sent to audience clients (docs/04, LiveState.theme).
type Theme struct {
	BG          string `json:"bg"`
	FG          string `json:"fg"`
	Accent      string `json:"accent"`
	Font        string `json:"font"`
	ScalePct    int    `json:"scalePct"`
	Speakers    string `json:"speakers"` // "label" | "color" | "hidden"
	IdleMessage string `json:"idleMessage"`
}

// DefaultTheme is pure black on white, high contrast (REQ-045).
func DefaultTheme(idle string) Theme {
	return Theme{BG: "#000000", FG: "#ffffff", Accent: "#ffffff", Font: "system", ScalePct: 100, Speakers: "label", IdleMessage: idle}
}

// AudItem is a fullscreen item as the audience sees it (image names become URLs).
type AudItem struct {
	Kind    string `json:"kind"`
	HTML    string `json:"html,omitempty"`
	Src     string `json:"src,omitempty"`
	Caption string `json:"caption,omitempty"`
}

// AudState is the audience snapshot.
type AudState struct {
	Seq        int64              `json:"seq"`
	TS         int64              `json:"ts"`
	Lang       string             `json:"lang"`
	Mode       string             `json:"mode"`
	Scrollback []live.CaptionItem `json:"scrollback"`
	Fullscreen *AudItem           `json:"fullscreen"`
	Preload    []string           `json:"preload"`
	Speakers   []string           `json:"speakers"` // order of first appearance; the position picks a character's color
	Theme      Theme              `json:"theme"`
}

const preloadMax = 8

// Runtime owns the active show's engine and the audience hub.
type Runtime struct {
	cfg config.Config
	Hub *hub.Hub

	// OpHub carries operator-level state updates (/api/operator/stream).
	OpHub *hub.Hub

	mu          sync.RWMutex
	gen         int
	showID      string
	showName    string
	assetBase   string // URL prefix for image assets ("" = /assets/<showID>/)
	cuesVer     int
	assetsDir   string
	theme       Theme
	engine      *live.Engine
	lastPreload []string
	speakers    []string

	ctrlMu   sync.Mutex
	ctrl     string // session token of the controlling operator
	ctrlSeen time.Time
	streams  map[string]int

	opDirty   chan struct{}
	saveDirty chan struct{}
	livePath  string
	quit      chan struct{}
	wg        sync.WaitGroup
}

// controllerLease is how long a controller with no open stream stays in control.
const controllerLease = 30 * time.Second

// NewRuntime creates a runtime with no show loaded (audience sees the idle message).
func NewRuntime(cfg config.Config) *Runtime {
	rt := &Runtime{cfg: cfg, Hub: hub.New(cfg.Audience.MaxClients), OpHub: hub.New(0),
		theme: DefaultTheme(cfg.Defaults.IdleMessage), streams: map[string]int{},
		opDirty: make(chan struct{}, 1), saveDirty: make(chan struct{}, 1), quit: make(chan struct{})}
	rt.Hub.OnChange = rt.notify
	rt.SetShow("", "", "", nil, rt.theme)
	rt.wg.Add(1)
	go rt.opLoop()
	return rt
}

// Close stops background goroutines and the engine, flushing live state.
func (rt *Runtime) Close() {
	close(rt.quit)
	rt.wg.Wait()
	rt.Flush()
	rt.Engine().Close()
}

// notify marks operator state and saved state as dirty. It never blocks.
func (rt *Runtime) notify() {
	select {
	case rt.opDirty <- struct{}{}:
	default:
	}
	select {
	case rt.saveDirty <- struct{}{}:
	default:
	}
}

func (rt *Runtime) opLoop() {
	defer rt.wg.Done()
	var saveTimer <-chan time.Time
	for {
		select {
		case <-rt.quit:
			return
		case <-rt.opDirty:
			b, _ := json.Marshal(rt.OperatorPayload())
			rt.OpHub.Broadcast(hub.Frame{Data: hub.FormatEvent(-1, "opstate", b)})
		case <-rt.saveDirty:
			rt.mu.RLock()
			persist := rt.livePath != ""
			rt.mu.RUnlock()
			if persist && saveTimer == nil {
				saveTimer = time.After(250 * time.Millisecond) // debounce (docs/02)
			}
		case <-saveTimer:
			saveTimer = nil
			rt.Flush()
		}
	}
}

// SetShow swaps in a show's cues (and resets live state), then pushes a
// snapshot to every connected client.
func (rt *Runtime) SetShow(id, name, assetsDir string, cues []script.Cue, theme Theme) {
	rt.mu.Lock()
	rt.gen++
	rt.cuesVer++
	gen := rt.gen
	old := rt.engine
	var startSeq int64
	if old != nil {
		startSeq = old.Snapshot().Seq + 1
	}
	rt.showID, rt.showName, rt.assetsDir, rt.theme = id, name, assetsDir, theme
	rt.speakers = script.SpeakersOf(cues)
	rt.engine = live.New(cues, live.Config{DefaultLoopSeconds: rt.cfg.Defaults.LoopItemSeconds,
		ScrollbackMax: rt.cfg.Audience.ScrollbackLines, StartSeq: startSeq}, func(ev live.Event) { rt.onEvent(gen, ev) })
	rt.lastPreload = nil
	rt.mu.Unlock()
	if old != nil {
		old.Close()
	}
	rt.broadcastSnapshot()
	rt.notify()
}

// Engine returns the current engine.
func (rt *Runtime) Engine() *live.Engine {
	rt.mu.RLock()
	defer rt.mu.RUnlock()
	return rt.engine
}

// ShowID returns the active show id.
func (rt *Runtime) ShowID() string {
	rt.mu.RLock()
	defer rt.mu.RUnlock()
	return rt.showID
}

// AssetsDir returns the active show's asset directory.
func (rt *Runtime) AssetsDir() string {
	rt.mu.RLock()
	defer rt.mu.RUnlock()
	return rt.assetsDir
}

// Do runs an operator command and then refreshes the audience's image preload list.
func (rt *Runtime) Do(fn func(*live.Engine) (live.OperatorState, error)) (live.OperatorState, error) {
	st, err := fn(rt.Engine())
	rt.refreshPreload()
	return st, err
}

func (rt *Runtime) item(it *script.Item) *AudItem {
	if it == nil {
		return nil
	}
	switch it.Kind {
	case "image":
		base := rt.assetBase
		if base == "" {
			base = "/assets/" + url.PathEscape(rt.showID) + "/"
		}
		return &AudItem{Kind: "image", Src: base + url.PathEscape(it.Image), Caption: it.Caption}
	case "text":
		return &AudItem{Kind: "text", HTML: it.HTML}
	}
	return &AudItem{Kind: "blank"}
}

// State builds the audience snapshot for the current engine state.
func (rt *Runtime) State() AudState {
	rt.mu.RLock()
	eng, theme := rt.engine, rt.theme
	rt.mu.RUnlock()
	st := eng.Snapshot()
	return rt.audState(st, theme, rt.preloadFor(eng))
}

func (rt *Runtime) audState(st live.State, theme Theme, preload []string) AudState {
	rt.mu.RLock()
	defer rt.mu.RUnlock()
	if st.Scrollback == nil {
		st.Scrollback = []live.CaptionItem{}
	}
	return AudState{Seq: st.Seq, TS: time.Now().UnixMilli(), Lang: rt.cfg.Defaults.Language, Mode: st.Mode,
		Scrollback: st.Scrollback, Fullscreen: rt.item(st.Fullscreen), Preload: preload, Speakers: rt.speakers, Theme: theme}
}

// onEvent runs under the engine lock: convert and broadcast without blocking.
func (rt *Runtime) onEvent(gen int, ev live.Event) {
	rt.mu.RLock()
	stale := gen != rt.gen
	var payload any
	ts := time.Now().UnixMilli()
	if !stale {
		switch ev.Type {
		case live.EvCaption:
			payload = map[string]any{"seq": ev.Seq, "ts": ts, "item": ev.Data}
		case live.EvFullscreen:
			it := ev.Data.(script.Item)
			payload = map[string]any{"seq": ev.Seq, "ts": ts, "item": rt.item(&it)}
		case live.EvBlank:
			payload = map[string]any{"seq": ev.Seq, "ts": ts, "on": ev.Data.(map[string]bool)["on"]}
		case live.EvSnapshot:
			st := ev.Data.(live.State)
			if st.Scrollback == nil {
				st.Scrollback = []live.CaptionItem{}
			}
			payload = AudState{Seq: st.Seq, TS: ts, Lang: rt.cfg.Defaults.Language, Mode: st.Mode, Scrollback: st.Scrollback,
				Fullscreen: rt.item(st.Fullscreen), Preload: rt.lastPreload, Speakers: rt.speakers, Theme: rt.theme}
		default:
			payload = map[string]any{"seq": ev.Seq, "ts": ts}
		}
	}
	rt.mu.RUnlock()
	if stale {
		return
	}
	b, err := json.Marshal(payload)
	if err != nil {
		return
	}
	rt.Hub.Broadcast(hub.Frame{Seq: ev.Seq, Data: hub.FormatEvent(ev.Seq, ev.Type, b)})
	rt.notify()
}

// broadcastSnapshot pushes a full snapshot to every client. The frame carries
// Seq 0 so the per-connection "already covered by my snapshot" filter never drops it.
func (rt *Runtime) broadcastSnapshot() {
	s := rt.State()
	b, _ := json.Marshal(s)
	rt.Hub.Broadcast(hub.Frame{Seq: 0, Data: hub.FormatEvent(s.Seq, live.EvSnapshot, b)})
}

// SetAssetBase sets the URL prefix for image assets (used by rehearsal runtimes).
func (rt *Runtime) SetAssetBase(base string) {
	rt.mu.Lock()
	rt.assetBase = base
	rt.mu.Unlock()
}

// SetTheme changes the look and pushes it to every connected client.
func (rt *Runtime) SetTheme(t Theme) {
	rt.mu.Lock()
	rt.theme = t
	rt.mu.Unlock()
	rt.broadcastSnapshot()
}

// ReplaceCues swaps in an edited script. If the show is not live it is reset
// (so a leading loop starts correctly); if live, the cursor is kept by index.
// It reports whether the cue count changed.
func (rt *Runtime) ReplaceCues(cues []script.Cue) (countChanged, wasLive bool) {
	eng := rt.Engine()
	rt.mu.Lock()
	rt.speakers = script.SpeakersOf(cues)
	rt.mu.Unlock()
	wasLive = eng.Operator().Live
	if wasLive {
		countChanged = eng.ReplaceCues(cues)
	} else {
		countChanged = len(cues) != len(eng.Cues())
		eng.ReplaceCues(cues)
		eng.Reset(-1)
	}
	rt.mu.Lock()
	rt.cuesVer++
	rt.mu.Unlock()
	rt.refreshPreload()
	rt.notify()
	return countChanged, wasLive
}

// SavedShowID returns the show id recorded in live.json, if any.
func (rt *Runtime) SavedShowID() string {
	rt.mu.RLock()
	path := rt.livePath
	rt.mu.RUnlock()
	b, err := os.ReadFile(path)
	if path == "" || err != nil {
		return ""
	}
	var sv savedLive
	if json.Unmarshal(b, &sv) != nil {
		return ""
	}
	return sv.ShowID
}

// Theme returns the current theme.
func (rt *Runtime) Theme() Theme {
	rt.mu.RLock()
	defer rt.mu.RUnlock()
	return rt.theme
}

// preloadFor lists the next few image URLs at or after the cursor.
func (rt *Runtime) preloadFor(eng *live.Engine) []string {
	cues, cursor := eng.Cues(), eng.Operator().Cursor
	seen := map[string]bool{}
	out := []string{}
	add := func(it script.Item) {
		if it.Kind != "image" || len(out) >= preloadMax {
			return
		}
		if u := rt.item(&it).Src; !seen[u] {
			seen[u] = true
			out = append(out, u)
		}
	}
	for i := cursor; i < len(cues) && len(out) < preloadMax; i++ {
		c := cues[i]
		if c.Item != nil {
			add(*c.Item)
		}
		for _, it := range c.Items {
			add(it)
		}
	}
	return out
}

// refreshPreload broadcasts the preload hint (no seq) if it changed.
func (rt *Runtime) refreshPreload() {
	eng := rt.Engine()
	list := rt.preloadFor(eng)
	rt.mu.Lock()
	same := len(list) == len(rt.lastPreload)
	for i := 0; same && i < len(list); i++ {
		same = list[i] == rt.lastPreload[i]
	}
	if same {
		rt.mu.Unlock()
		return
	}
	rt.lastPreload = list
	rt.mu.Unlock()
	b, _ := json.Marshal(map[string]any{"urls": list})
	rt.Hub.Broadcast(hub.Frame{Seq: 0, Data: hub.FormatEvent(-1, "preload", b)})
}

// --- operator control (REQ-034): one controlling operator at a time ---

// SID is the public identifier of a session token (safe to send to browsers).
func SID(token string) string {
	h := sha256.Sum256([]byte("sid:" + token))
	return hex.EncodeToString(h[:6])
}

func (rt *Runtime) ctrlLiveLocked() bool {
	return rt.ctrl != "" && (rt.streams[rt.ctrl] > 0 || time.Since(rt.ctrlSeen) < controllerLease)
}

// Claim makes token the controller if nobody live holds control, or if takeover is set.
func (rt *Runtime) Claim(token string, takeover bool) bool {
	rt.ctrlMu.Lock()
	ok := false
	if rt.ctrl == token || !rt.ctrlLiveLocked() || takeover {
		rt.ctrl, rt.ctrlSeen, ok = token, time.Now(), true
	}
	rt.ctrlMu.Unlock()
	rt.notify()
	return ok
}

// IsController reports whether token currently controls the show, refreshing its lease.
func (rt *Runtime) IsController(token string) bool {
	rt.ctrlMu.Lock()
	defer rt.ctrlMu.Unlock()
	if rt.ctrl != token {
		return false
	}
	rt.ctrlSeen = time.Now()
	return true
}

func (rt *Runtime) streamOpen(token string) {
	rt.ctrlMu.Lock()
	rt.streams[token]++
	rt.ctrlMu.Unlock()
}

func (rt *Runtime) streamClose(token string) {
	rt.ctrlMu.Lock()
	if rt.streams[token]--; rt.streams[token] <= 0 {
		delete(rt.streams, token)
	}
	if token == rt.ctrl {
		rt.ctrlSeen = time.Now()
	}
	rt.ctrlMu.Unlock()
	rt.notify()
}

// ControllerSID returns the live controller's public id, or "".
func (rt *Runtime) ControllerSID() string {
	rt.ctrlMu.Lock()
	defer rt.ctrlMu.Unlock()
	if rt.ctrlLiveLocked() {
		return SID(rt.ctrl)
	}
	return ""
}

// OperatorPayload is what operator streams carry on every change.
type OperatorPayload struct {
	Op         live.OperatorState `json:"op"`
	Audience   int                `json:"audience"`
	Controller string             `json:"controller"`
	CuesVer    int                `json:"cuesVer"`
	Show       struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"show"`
}

// OperatorPayload snapshots the operator-visible state.
func (rt *Runtime) OperatorPayload() OperatorPayload {
	rt.mu.RLock()
	eng, ver, id, name := rt.engine, rt.cuesVer, rt.showID, rt.showName
	rt.mu.RUnlock()
	p := OperatorPayload{Op: eng.Operator(), Audience: rt.Hub.Count(), Controller: rt.ControllerSID(), CuesVer: ver}
	p.Show.ID, p.Show.Name = id, name
	return p
}

// --- persistence (REQ-012) ---

type savedLive struct {
	ShowID string         `json:"showId"`
	State  live.Persisted `json:"state"`
}

// EnablePersistence turns on debounced saving of live state to path.
func (rt *Runtime) EnablePersistence(path string) {
	rt.mu.Lock()
	rt.livePath = path
	rt.mu.Unlock()
}

// Flush writes live state now (no-op if persistence is off or no show is loaded).
func (rt *Runtime) Flush() {
	rt.mu.RLock()
	path, id, eng := rt.livePath, rt.showID, rt.engine
	rt.mu.RUnlock()
	if path == "" || id == "" {
		return
	}
	b, err := json.Marshal(savedLive{ShowID: id, State: eng.Export()})
	if err != nil {
		return
	}
	os.MkdirAll(filepath.Dir(path), 0o755)
	tmp := path + ".tmp"
	if os.WriteFile(tmp, b, 0o600) == nil {
		os.Rename(tmp, path)
	}
}

// Restore reloads saved live state if it belongs to the currently loaded show.
func (rt *Runtime) Restore() bool {
	rt.mu.RLock()
	path, id, eng := rt.livePath, rt.showID, rt.engine
	rt.mu.RUnlock()
	b, err := os.ReadFile(path)
	if path == "" || err != nil {
		return false
	}
	var sv savedLive
	if json.Unmarshal(b, &sv) != nil || sv.ShowID != id {
		return false
	}
	eng.Restore(sv.State)
	rt.refreshPreload()
	return true
}

// SetName changes the displayed show name without resetting anything.
func (rt *Runtime) SetName(name string) {
	rt.mu.Lock()
	rt.showName = name
	rt.mu.Unlock()
	rt.notify()
}
