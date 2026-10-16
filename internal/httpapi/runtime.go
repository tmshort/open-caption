package httpapi

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/tmshort/open-caption/internal/config"
	"github.com/tmshort/open-caption/internal/hub"
	"github.com/tmshort/open-caption/internal/live"
	"github.com/tmshort/open-caption/internal/showfile"
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
	// IdleMessages are the pre-show messages in other languages (by code). Clients are sent only the
	// one for their language, in IdleMessage.
	IdleMessages map[string]string `json:"-"`
	// LoopSeconds and ScrollbackLines are show-wide settings that travel with the theme but are not sent
	// to audience devices.
	LoopSeconds     float64 `json:"-"`
	ScrollbackLines int     `json:"-"`
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
	// Deadline (unix ms) is when the countdown in a loop's text reaches zero; the page fills in each
	// [data-timer] element from it, using the snapshot's ts to allow for a device clock that is wrong.
	Deadline int64 `json:"deadline,omitempty"`
}

// StandbyView is what the standby screen shows: the venue's message and logo (both optional; with no message
// the page uses its built-in words in the viewer's language).
type StandbyView struct {
	Message string `json:"message,omitempty"`
	Logo    string `json:"logo,omitempty"` // URL of the logo image
}

// LangInfo names a language a show is available in.
type LangInfo struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

// LangFile is one language's parsed show file.
type LangFile struct {
	Code, Name string
	Cues       []showfile.Cue
}

// AudState is the audience snapshot.
type AudState struct {
	Seq        int64              `json:"seq"`
	TS         int64              `json:"ts"`
	Lang       string             `json:"lang"`
	Languages  []LangInfo         `json:"languages"`
	Mode       string             `json:"mode"`
	Scrollback []live.CaptionItem `json:"scrollback"`
	Fullscreen *AudItem           `json:"fullscreen"`
	Preload    []string           `json:"preload"`
	Speakers   []string           `json:"speakers"` // order of first appearance; the position picks a character's color
	Theme      Theme              `json:"theme"`
	Standby    *StandbyView       `json:"standby,omitempty"` // only while the show is not on air (mode "standby")
}

const preloadMax = 8

// langRT is one language of the active show: its own engine (kept in step with the others by Do).
type langRT struct {
	code, name  string
	engine      *live.Engine
	speakers    []string
	lastPreload []string
}

// Runtime owns the active show: one engine per language, driven together by the operator, and one
// audience hub per language.
type Runtime struct {
	cfg config.Config

	// OpHub carries operator-level state updates (/api/operator/stream).
	OpHub *hub.Hub

	mu        sync.RWMutex
	gen       int
	showID    string
	showName  string
	assetBase string // URL prefix for image assets ("" = /assets/<showID>/)
	cuesVer   int
	assetsDir string
	theme     Theme
	primary   string              // the language the operator sees
	order     []string            // language codes, primary first
	langs     map[string]*langRT  // languages of the active show
	hubs      map[string]*hub.Hub // one per language code; kept across shows so connected devices stay put

	standbyMode bool        // shows start in standby (see live.Config.Standby)
	standby     StandbyView // the venue's standby screen

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
	rt := &Runtime{cfg: cfg, OpHub: hub.New(0), hubs: map[string]*hub.Hub{}, langs: map[string]*langRT{},
		theme: DefaultTheme(cfg.Defaults.IdleMessage), streams: map[string]int{},
		opDirty: make(chan struct{}, 1), saveDirty: make(chan struct{}, 1), quit: make(chan struct{})}
	rt.SetShow("", "", "", nil, rt.theme)
	rt.wg.Add(1)
	go rt.opLoop()
	return rt
}

// UseStandby makes every show start in standby: nothing of it reaches the audience until the operator starts
// it. Call it before a show is loaded; it also puts the (empty) current show into standby.
func (rt *Runtime) UseStandby() {
	rt.mu.Lock()
	rt.standbyMode = true
	id, name, dir, theme := rt.showID, rt.showName, rt.assetsDir, rt.theme
	rt.mu.Unlock()
	if id == "" {
		rt.SetShow("", name, dir, nil, theme)
	}
}

// SetStandbyScreen sets the venue's standby screen and shows the change to devices that are looking at it.
func (rt *Runtime) SetStandbyScreen(v StandbyView) {
	rt.mu.Lock()
	rt.standby = v
	hubs := make(map[string]*hub.Hub, len(rt.hubs))
	for c, h := range rt.hubs {
		hubs[c] = h
	}
	l := rt.langs[rt.primary]
	rt.mu.Unlock()
	if l != nil && l.engine.Snapshot().Mode == live.ModeStandby {
		rt.pushToAll(hubs)
	}
}

// Close stops background goroutines and the engines, flushing live state.
func (rt *Runtime) Close() {
	close(rt.quit)
	rt.wg.Wait()
	rt.Flush()
	rt.mu.RLock()
	var engs []*live.Engine
	for _, l := range rt.langs {
		engs = append(engs, l.engine)
	}
	rt.mu.RUnlock()
	for _, e := range engs {
		e.Close()
	}
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

// hubLocked returns the hub for a language, creating it on first use (caller holds rt.mu for writing).
func (rt *Runtime) hubLocked(code string) *hub.Hub {
	h := rt.hubs[code]
	if h == nil {
		h = hub.New(rt.cfg.Audience.MaxClients)
		h.OnChange = rt.notify
		rt.hubs[code] = h
	}
	return h
}

// HubFor returns the audience hub for a language code.
func (rt *Runtime) HubFor(code string) *hub.Hub {
	rt.mu.Lock()
	defer rt.mu.Unlock()
	return rt.hubLocked(code)
}

// AudienceCount is the number of connected audience devices across all languages.
func (rt *Runtime) AudienceCount() int {
	n := 0
	for _, c := range rt.AudienceByLang() {
		n += c
	}
	return n
}

// AudienceByLang counts connected devices per language code.
func (rt *Runtime) AudienceByLang() map[string]int {
	rt.mu.RLock()
	hubs := make(map[string]*hub.Hub, len(rt.hubs))
	for k, h := range rt.hubs {
		hubs[k] = h
	}
	rt.mu.RUnlock()
	out := map[string]int{}
	for k, h := range hubs {
		if c := h.Count(); c > 0 {
			out[k] = c
		}
	}
	return out
}

func (rt *Runtime) newEngine(gen int, code string, cues []showfile.Cue, startSeq int64) *live.Engine {
	return live.New(cues, rt.engineConfigLocked(startSeq), func(ev live.Event) { rt.onEvent(gen, code, ev) })
}

// engineConfigLocked is the engine configuration from the show's settings, falling back to config.yaml.
func (rt *Runtime) engineConfigLocked(startSeq int64) live.Config {
	c := live.Config{DefaultLoopSeconds: rt.cfg.Defaults.LoopItemSeconds, ScrollbackMax: rt.cfg.Audience.ScrollbackLines, StartSeq: startSeq, Standby: rt.standbyMode}
	if rt.theme.LoopSeconds > 0 {
		c.DefaultLoopSeconds = rt.theme.LoopSeconds
	}
	if rt.theme.ScrollbackLines > 0 {
		c.ScrollbackMax = rt.theme.ScrollbackLines
	}
	return c
}

// oldSeqs reads each language's current sequence number. It is done before taking rt.mu for writing
// because an engine may be holding its own lock while waiting for rt.mu in onEvent.
func (rt *Runtime) oldSeqs() map[string]int64 {
	rt.mu.RLock()
	engs := map[string]*live.Engine{}
	for c, l := range rt.langs {
		engs[c] = l.engine
	}
	rt.mu.RUnlock()
	out := map[string]int64{}
	for c, e := range engs {
		out[c] = e.Snapshot().Seq + 1
	}
	return out
}

// SetShow swaps in a one-language show (and resets live state), then pushes a snapshot to every client.
func (rt *Runtime) SetShow(id, name, assetsDir string, cues []showfile.Cue, theme Theme) {
	code := rt.cfg.Defaults.Language
	if code == "" {
		code = "en"
	}
	rt.SetShowLangs(id, name, assetsDir, LangFile{Code: code, Name: showfile.LanguageName(code), Cues: cues}, nil, theme)
}

// SetShowLangs swaps in a show with a primary language and any translations (which must already have
// been checked to match the primary's structure), resets live state, and pushes a snapshot to every client.
func (rt *Runtime) SetShowLangs(id, name, assetsDir string, primary LangFile, others []LangFile, theme Theme) {
	seqs := rt.oldSeqs()
	rt.mu.Lock()
	rt.gen++
	rt.cuesVer++
	gen := rt.gen
	old := rt.langs
	rt.showID, rt.showName, rt.assetsDir, rt.theme = id, name, assetsDir, theme
	rt.primary = primary.Code
	rt.langs = map[string]*langRT{}
	rt.order = nil
	for _, ls := range append([]LangFile{primary}, others...) {
		if _, dup := rt.langs[ls.Code]; dup {
			continue
		}
		rt.hubLocked(ls.Code)
		rt.order = append(rt.order, ls.Code)
		rt.langs[ls.Code] = &langRT{code: ls.Code, name: ls.Name, speakers: showfile.SpeakersOf(ls.Cues),
			engine: rt.newEngine(gen, ls.Code, ls.Cues, seqs[ls.Code])}
	}
	hubs := make(map[string]*hub.Hub, len(rt.hubs))
	for c, h := range rt.hubs {
		hubs[c] = h
	}
	rt.mu.Unlock()
	for _, l := range old {
		l.engine.Close()
	}
	rt.pushToAll(hubs)
	rt.notify()
}

// pushToAll sends every language hub a fresh snapshot, or tells devices whose language no longer exists
// in the show to reload (they then pick a language again).
func (rt *Runtime) pushToAll(hubs map[string]*hub.Hub) {
	for code, h := range hubs {
		rt.mu.RLock()
		_, has := rt.langs[code]
		rt.mu.RUnlock()
		if !has {
			h.Broadcast(hub.Frame{Seq: 0, Data: hub.FormatEvent(-1, "reload", []byte("{}"))})
			continue
		}
		s := rt.State(code)
		b, _ := json.Marshal(s)
		h.Broadcast(hub.Frame{Seq: 0, Data: hub.FormatEvent(s.Seq, live.EvSnapshot, b)})
	}
}

// Engine returns the primary language's engine (what the operator sees and controls).
func (rt *Runtime) Engine() *live.Engine {
	rt.mu.RLock()
	defer rt.mu.RUnlock()
	return rt.langs[rt.primary].engine
}

// Languages lists the active show's languages, the primary first.
func (rt *Runtime) Languages() []LangInfo {
	rt.mu.RLock()
	defer rt.mu.RUnlock()
	return rt.languagesLocked()
}

func (rt *Runtime) languagesLocked() []LangInfo {
	out := make([]LangInfo, 0, len(rt.order))
	for _, c := range rt.order {
		out = append(out, LangInfo{Code: c, Name: rt.langs[c].name})
	}
	return out
}

// PrimaryLang is the code of the language the operator follows.
func (rt *Runtime) PrimaryLang() string {
	rt.mu.RLock()
	defer rt.mu.RUnlock()
	return rt.primary
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

// Do runs an operator command on the primary language and then on every translation, so all languages
// move to the same cue together; then it refreshes the audience's image preload lists.
func (rt *Runtime) Do(fn func(*live.Engine) (live.OperatorState, error)) (live.OperatorState, error) {
	rt.mu.RLock()
	primary := rt.langs[rt.primary]
	var others []*langRT
	for _, c := range rt.order[1:] {
		others = append(others, rt.langs[c])
	}
	rt.mu.RUnlock()

	st, err := fn(primary.engine)
	if err == live.ErrStale || len(others) == 0 {
		rt.refreshPreload()
		return st, err
	}
	for _, l := range others {
		fn(l.engine) // their errors (such as end of show file) mirror the primary's
		// A translation must never drift from the primary: put it on the same cue if it did.
		if l.engine.Operator().Cursor != st.Cursor {
			l.engine.Jump(st.Cursor, -1)
		}
	}
	rt.refreshPreload()
	return st, err
}

func (rt *Runtime) item(it *showfile.Item) *AudItem {
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
		return &AudItem{Kind: "text", HTML: it.HTML, Deadline: it.Deadline}
	}
	return &AudItem{Kind: "blank"}
}

// resolveLang maps a requested code to one the show has (exact, then its base language), or the primary.
func (rt *Runtime) resolveLang(code string) string {
	rt.mu.RLock()
	defer rt.mu.RUnlock()
	return rt.resolveLangLocked(code)
}

func (rt *Runtime) resolveLangLocked(code string) string {
	if _, ok := rt.langs[code]; ok {
		return code
	}
	return rt.primary
}

// State builds the audience snapshot for a language (the primary if the show lacks it).
func (rt *Runtime) State(lang string) AudState {
	rt.mu.RLock()
	lang = rt.resolveLangLocked(lang)
	l := rt.langs[lang]
	eng := l.engine
	rt.mu.RUnlock()
	st := eng.Snapshot()
	return rt.audState(lang, st, rt.preloadFor(l))
}

func (rt *Runtime) audState(lang string, st live.State, preload []string) AudState {
	rt.mu.RLock()
	defer rt.mu.RUnlock()
	if st.Scrollback == nil {
		st.Scrollback = []live.CaptionItem{}
	}
	return AudState{Seq: st.Seq, TS: time.Now().UnixMilli(), Lang: lang, Languages: rt.languagesLocked(), Mode: st.Mode,
		Scrollback: st.Scrollback, Fullscreen: rt.item(st.Fullscreen), Preload: preload,
		Speakers: rt.langs[lang].speakers, Theme: rt.themeForLocked(lang), Standby: rt.standbyForLocked(st.Mode)}
}

func (rt *Runtime) standbyForLocked(mode string) *StandbyView {
	if mode != live.ModeStandby {
		return nil
	}
	v := rt.standby
	return &v
}

// themeForLocked is the theme with the pre-show message in the viewer's language.
func (rt *Runtime) themeForLocked(lang string) Theme {
	t := rt.theme
	if m := t.IdleMessages[lang]; m != "" {
		t.IdleMessage = m
	}
	return t
}

// onEvent runs under an engine lock: convert and broadcast without blocking.
func (rt *Runtime) onEvent(gen int, lang string, ev live.Event) {
	rt.mu.RLock()
	l := rt.langs[lang]
	stale := gen != rt.gen || l == nil
	var payload any
	var h *hub.Hub
	ts := time.Now().UnixMilli()
	if !stale {
		h = rt.hubs[lang]
		switch ev.Type {
		case live.EvCaption:
			payload = map[string]any{"seq": ev.Seq, "ts": ts, "item": ev.Data}
		case live.EvFullscreen:
			it := ev.Data.(showfile.Item)
			payload = map[string]any{"seq": ev.Seq, "ts": ts, "item": rt.item(&it)}
		case live.EvBlank:
			payload = map[string]any{"seq": ev.Seq, "ts": ts, "on": ev.Data.(map[string]bool)["on"]}
		case live.EvSnapshot:
			st := ev.Data.(live.State)
			if st.Scrollback == nil {
				st.Scrollback = []live.CaptionItem{}
			}
			payload = AudState{Seq: st.Seq, TS: ts, Lang: lang, Languages: rt.languagesLocked(), Mode: st.Mode, Scrollback: st.Scrollback,
				Fullscreen: rt.item(st.Fullscreen), Preload: l.lastPreload, Speakers: l.speakers, Theme: rt.themeForLocked(lang), Standby: rt.standbyForLocked(st.Mode)}
		default:
			payload = map[string]any{"seq": ev.Seq, "ts": ts}
		}
	}
	rt.mu.RUnlock()
	if stale || h == nil {
		return
	}
	b, err := json.Marshal(payload)
	if err != nil {
		return
	}
	h.Broadcast(hub.Frame{Seq: ev.Seq, Data: hub.FormatEvent(ev.Seq, ev.Type, b)})
	rt.notify()
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
	hubs := make(map[string]*hub.Hub, len(rt.hubs))
	for c, h := range rt.hubs {
		hubs[c] = h
	}
	cfg := rt.engineConfigLocked(0)
	var engs []*live.Engine
	for _, l := range rt.langs {
		engs = append(engs, l.engine)
	}
	rt.mu.Unlock()
	for _, e := range engs { // loop time and scrollback size apply to the show on air
		e.Configure(cfg)
	}
	rt.pushToAll(hubs)
}

// ReplaceLangs swaps in edited show files for the active show. If the show is not live it is reset (so a
// leading loop starts correctly); if live, every language keeps the operator's place by cue number.
// It reports whether the primary show file's cue count changed.
func (rt *Runtime) ReplaceLangs(primary LangFile, others []LangFile) (countChanged, wasLive bool) {
	eng := rt.Engine()
	oldCount := len(eng.Cues())
	wasLive = eng.Operator().Live
	countChanged = oldCount != len(primary.Cues)
	if !wasLive {
		rt.mu.RLock()
		id, name, dir, theme := rt.showID, rt.showName, rt.assetsDir, rt.theme
		rt.mu.RUnlock()
		rt.SetShowLangs(id, name, dir, primary, others, theme)
		return
	}

	cursor := eng.Operator().Cursor
	all := append([]LangFile{primary}, others...)
	rt.mu.Lock()
	gen := rt.gen
	rt.cuesVer++
	rt.primary = primary.Code
	var removed []*langRT
	for c, l := range rt.langs {
		keep := false
		for _, ls := range all {
			keep = keep || ls.Code == c
		}
		if !keep {
			removed = append(removed, l)
			delete(rt.langs, c)
		}
	}
	var added []*langRT
	type replace struct {
		eng  *live.Engine
		cues []showfile.Cue
	}
	var replaces []replace
	rt.order = nil
	for _, ls := range all {
		if _, dup := rt.langsOrderHas(ls.Code); dup {
			continue
		}
		rt.order = append(rt.order, ls.Code)
		if l, ok := rt.langs[ls.Code]; ok {
			l.name, l.speakers = ls.Name, showfile.SpeakersOf(ls.Cues)
			replaces = append(replaces, replace{l.engine, ls.Cues}) // applied after unlocking: the engine emits a snapshot, which calls onEvent
		} else {
			l := &langRT{code: ls.Code, name: ls.Name, speakers: showfile.SpeakersOf(ls.Cues)}
			l.engine = rt.newEngine(gen, ls.Code, ls.Cues, 0)
			rt.langs[ls.Code] = l
			rt.hubLocked(ls.Code)
			added = append(added, l)
		}
	}
	reload := make([]*hub.Hub, 0, len(removed))
	for _, l := range removed {
		reload = append(reload, rt.hubs[l.code])
	}
	rt.mu.Unlock()
	for i, l := range removed {
		l.engine.Close()
		if reload[i] != nil {
			reload[i].Broadcast(hub.Frame{Seq: 0, Data: hub.FormatEvent(-1, "reload", []byte("{}"))})
		}
	}
	for _, r := range replaces {
		r.eng.ReplaceCues(r.cues)
	}
	for _, l := range added {
		l.engine.Jump(cursor, -1)
	}
	rt.refreshPreload()
	rt.notify()
	return
}

func (rt *Runtime) langsOrderHas(code string) (int, bool) {
	for i, c := range rt.order {
		if c == code {
			return i, true
		}
	}
	return -1, false
}

// ReplaceCues replaces the primary language's show file only (kept for single-language callers).
func (rt *Runtime) ReplaceCues(cues []showfile.Cue) (countChanged, wasLive bool) {
	rt.mu.RLock()
	code, name := rt.primary, rt.langs[rt.primary].name
	var others []LangFile
	for _, c := range rt.order[1:] {
		others = append(others, LangFile{Code: c, Name: rt.langs[c].name, Cues: rt.langs[c].engine.Cues()})
	}
	rt.mu.RUnlock()
	return rt.ReplaceLangs(LangFile{Code: code, Name: name, Cues: cues}, others)
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
func (rt *Runtime) preloadFor(l *langRT) []string {
	cues, cursor := l.engine.Cues(), l.engine.Operator().Cursor
	seen := map[string]bool{}
	out := []string{}
	add := func(it showfile.Item) {
		if it.Kind != "image" || len(out) >= preloadMax {
			return
		}
		rt.mu.RLock()
		u := rt.item(&it).Src
		rt.mu.RUnlock()
		if !seen[u] {
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

// refreshPreload broadcasts each language's preload hint (no seq) if it changed.
func (rt *Runtime) refreshPreload() {
	rt.mu.RLock()
	ls := make([]*langRT, 0, len(rt.order))
	for _, c := range rt.order {
		ls = append(ls, rt.langs[c])
	}
	rt.mu.RUnlock()
	for _, l := range ls {
		list := rt.preloadFor(l)
		rt.mu.Lock()
		same := len(list) == len(l.lastPreload)
		for i := 0; same && i < len(list); i++ {
			same = list[i] == l.lastPreload[i]
		}
		if same {
			rt.mu.Unlock()
			continue
		}
		l.lastPreload = list
		h := rt.hubs[l.code]
		rt.mu.Unlock()
		b, _ := json.Marshal(map[string]any{"urls": list})
		h.Broadcast(hub.Frame{Seq: 0, Data: hub.FormatEvent(-1, "preload", b)})
	}
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
	Op             live.OperatorState `json:"op"`
	Audience       int                `json:"audience"`
	AudienceByLang map[string]int     `json:"audienceByLang"`
	Languages      []LangInfo         `json:"languages"`
	Controller     string             `json:"controller"`
	CuesVer        int                `json:"cuesVer"`
	Show           struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"show"`
}

// OperatorPayload snapshots the operator-visible state.
func (rt *Runtime) OperatorPayload() OperatorPayload {
	rt.mu.RLock()
	eng, ver, id, name, langs := rt.langs[rt.primary].engine, rt.cuesVer, rt.showID, rt.showName, rt.languagesLocked()
	rt.mu.RUnlock()
	by := rt.AudienceByLang()
	total := 0
	for _, c := range by {
		total += c
	}
	p := OperatorPayload{Op: eng.Operator(), Audience: total, AudienceByLang: by, Languages: langs, Controller: rt.ControllerSID(), CuesVer: ver}
	p.Show.ID, p.Show.Name = id, name
	return p
}

// --- persistence (REQ-012) ---

type savedLive struct {
	ShowID string                    `json:"showId"`
	State  live.Persisted            `json:"state"` // the primary language (older files have only this)
	Langs  map[string]live.Persisted `json:"langs,omitempty"`
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
	path, id := rt.livePath, rt.showID
	engs := map[string]*live.Engine{}
	for c, l := range rt.langs {
		engs[c] = l.engine
	}
	primary := rt.primary
	rt.mu.RUnlock()
	if path == "" || id == "" {
		return
	}
	sv := savedLive{ShowID: id, State: engs[primary].Export(), Langs: map[string]live.Persisted{}}
	for c, e := range engs {
		sv.Langs[c] = e.Export()
	}
	b, err := json.Marshal(sv)
	if err != nil {
		return
	}
	os.MkdirAll(filepath.Dir(path), 0o755)
	tmp := path + ".tmp"
	if os.WriteFile(tmp, b, 0o600) == nil {
		os.Rename(tmp, path)
	}
}

// Restore reloads saved live state if it belongs to the currently loaded show. A language with no saved
// state (added since) is moved to the primary's cue.
func (rt *Runtime) Restore() bool {
	rt.mu.RLock()
	path, id, primary := rt.livePath, rt.showID, rt.primary
	engs := map[string]*live.Engine{}
	for c, l := range rt.langs {
		engs[c] = l.engine
	}
	rt.mu.RUnlock()
	b, err := os.ReadFile(path)
	if path == "" || err != nil {
		return false
	}
	var sv savedLive
	if json.Unmarshal(b, &sv) != nil || sv.ShowID != id {
		return false
	}
	engs[primary].Restore(sv.State)
	for c, e := range engs {
		if c == primary {
			continue
		}
		if p, ok := sv.Langs[c]; ok {
			e.Restore(p)
		} else {
			e.Jump(sv.State.Cursor, -1)
		}
	}
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

// negotiateLang picks the language for an audience request: an explicit ?lang= wins; otherwise the
// browser's Accept-Language header (quality values respected, regional tags falling back to the base
// language, e.g. es-MX to es); otherwise the show's primary language.
func negotiateLang(query, acceptLanguage string, langs []LangInfo, primary string) string {
	has := func(code string) (string, bool) {
		for _, l := range langs {
			if l.Code == code {
				return l.Code, true
			}
		}
		return "", false
	}
	// match finds a show language for one requested tag: exact, then its base language, then a regional
	// variant of the base (a request for "pt" is served "pt-br" if that is all there is).
	match := func(tag string) (string, bool) {
		tag = strings.ToLower(strings.TrimSpace(tag))
		if tag == "" || tag == "*" {
			return "", false
		}
		if c, ok := has(tag); ok {
			return c, true
		}
		base := tag
		if i := strings.IndexByte(tag, '-'); i >= 0 {
			base = tag[:i]
			if c, ok := has(base); ok {
				return c, true
			}
		}
		for _, l := range langs {
			if strings.HasPrefix(l.Code, base+"-") {
				return l.Code, true
			}
		}
		return "", false
	}
	if c, ok := match(query); ok {
		return c
	}
	type cand struct {
		tag string
		q   float64
		pos int
	}
	var cands []cand
	for i, part := range strings.Split(acceptLanguage, ",") {
		tag, params, _ := strings.Cut(strings.TrimSpace(part), ";")
		q := 1.0
		if _, v, ok := strings.Cut(strings.TrimSpace(params), "q="); ok {
			if f, err := strconv.ParseFloat(strings.TrimSpace(v), 64); err == nil {
				q = f
			}
		}
		if q > 0 {
			cands = append(cands, cand{tag, q, i})
		}
	}
	sort.SliceStable(cands, func(i, j int) bool { return cands[i].q > cands[j].q })
	for _, c := range cands {
		if code, ok := match(c.tag); ok {
			return code
		}
	}
	return primary
}

// Negotiate chooses the language for an audience request to this runtime.
func (rt *Runtime) Negotiate(query, acceptLanguage string) string {
	rt.mu.RLock()
	defer rt.mu.RUnlock()
	return negotiateLang(query, acceptLanguage, rt.languagesLocked(), rt.primary)
}
