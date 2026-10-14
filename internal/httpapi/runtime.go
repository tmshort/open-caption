package httpapi

import (
	"encoding/json"
	"net/url"
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
	Theme      Theme              `json:"theme"`
}

const preloadMax = 8

// Runtime owns the active show's engine and the audience hub.
type Runtime struct {
	cfg config.Config
	Hub *hub.Hub

	mu          sync.RWMutex
	gen         int
	showID      string
	assetsDir   string
	theme       Theme
	engine      *live.Engine
	lastPreload []string
}

// NewRuntime creates a runtime with no show loaded (audience sees the idle message).
func NewRuntime(cfg config.Config) *Runtime {
	rt := &Runtime{cfg: cfg, Hub: hub.New(cfg.Audience.MaxClients), theme: DefaultTheme(cfg.Defaults.IdleMessage)}
	rt.SetShow("", "", nil, rt.theme)
	return rt
}

// SetShow swaps in a show's cues (and resets live state), then pushes a
// snapshot to every connected client.
func (rt *Runtime) SetShow(id, assetsDir string, cues []script.Cue, theme Theme) {
	rt.mu.Lock()
	rt.gen++
	gen := rt.gen
	old := rt.engine
	rt.showID, rt.assetsDir, rt.theme = id, assetsDir, theme
	rt.engine = live.New(cues, live.Config{DefaultLoopSeconds: rt.cfg.Defaults.LoopItemSeconds,
		ScrollbackMax: rt.cfg.Audience.ScrollbackLines}, func(ev live.Event) { rt.onEvent(gen, ev) })
	rt.lastPreload = nil
	rt.mu.Unlock()
	if old != nil {
		old.Close()
	}
	rt.broadcastSnapshot()
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
		return &AudItem{Kind: "image", Src: "/assets/" + url.PathEscape(rt.showID) + "/" + url.PathEscape(it.Image), Caption: it.Caption}
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
		Scrollback: st.Scrollback, Fullscreen: rt.item(st.Fullscreen), Preload: preload, Theme: theme}
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
				Fullscreen: rt.item(st.Fullscreen), Preload: rt.lastPreload, Theme: rt.theme}
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
}

func (rt *Runtime) broadcastSnapshot() {
	s := rt.State()
	b, _ := json.Marshal(s)
	rt.Hub.Broadcast(hub.Frame{Seq: s.Seq, Data: hub.FormatEvent(s.Seq, live.EvSnapshot, b)})
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
