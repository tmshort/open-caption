// Copyright the open-caption authors
// SPDX-License-Identifier: Apache-2.0

// Package report counts the devices that used the audience page during one run of a show (REQ-070): from the
// moment the operator puts the show on air until it goes back to standby, is restarted or replaced. A device is
// told apart by its address and browser type, so a phone that drops off the WiFi and comes back is one device
// however often it reconnects, and the report says how many people used the service regardless of how long
// each stayed. It knows nothing of the network or the show: the program tells it when a device connects or
// disconnects and when the show goes on and off the air.
package report

import (
	"crypto/sha256"
	"encoding/hex"
	"sort"
	"strconv"
	"sync"
	"time"
)

// Device kinds (as the program's device list names them; anything else counts with the computers).
const (
	Phone    = "phone"
	Tablet   = "tablet"
	Computer = "computer"
)

// Span is a length of time that is written as a whole number of seconds in JSON (the saved reports and the API).
type Span time.Duration

func (s Span) MarshalJSON() ([]byte, error) {
	return []byte(strconv.FormatInt(int64(time.Duration(s).Round(time.Second)/time.Second), 10)), nil
}

func (s *Span) UnmarshalJSON(b []byte) error {
	n, err := strconv.ParseInt(string(b), 10, 64)
	*s = Span(time.Duration(n) * time.Second)
	return err
}

// Device is one device (an address and a browser type) seen during the run.
type Device struct {
	ID          string    `json:"id"` // a short anonymous code made from the address and browser type
	Address     string    `json:"address"`
	Kind        string    `json:"kind"`
	Projector   bool      `json:"projector,omitempty"` // opened as a projector / shared screen
	Languages   []string  `json:"languages"`
	First       time.Time `json:"first"`
	Last        time.Time `json:"last"`
	Connections int       `json:"connections"`
	Connected   Span      `json:"connected"` // time with at least one connection open
	Longest     Span      `json:"longest"`   // the longest stay without a break
	Now         bool      `json:"now"`       // still connected (at the end of the run, or at this moment)
}

// Point is how many devices were connected at one moment; the count stays the same until the next point.
type Point struct {
	T time.Time `json:"t"`
	N int       `json:"n"`
}

// Loop kinds, by where a loop is in the show file.
const (
	LoopPreShow      = "preshow"      // before the first line the audience sees
	LoopIntermission = "intermission" // between lines
	LoopPostShow     = "postshow"     // after the last line
	LoopOther        = "loop"         // a show with no lines of its own
)

// Loop is a stretch of the run during which a loop (a pre-show, an intermission) was running.
type Loop struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
	Kind  string    `json:"kind"`           // what it is taken for by where it stands (LoopPreShow ...)
	Name  string    `json:"name,omitempty"` // what the show file calls it (name Intermission); better than Kind when there
	Cue   int       `json:"cue"`            // the loop's place in the show file
}

// LoopInfo says which loop is running.
type LoopInfo struct {
	Cue  int
	Kind string
	Name string
}

// Totals sum the run up.
type Totals struct {
	Devices      int       `json:"devices"`
	Phones       int       `json:"phones"`
	Tablets      int       `json:"tablets"`
	Computers    int       `json:"computers"`
	Projectors   int       `json:"projectors"`
	Peak         int       `json:"peak"` // the most devices connected at once
	PeakAt       time.Time `json:"peakAt"`
	Now          int       `json:"now"` // devices still connected (at the end of the run, or at this moment)
	Connections  int       `json:"connections"`
	Reconnects   int       `json:"reconnects"` // connections beyond each device's first
	AvgConnected Span      `json:"avgConnected"`
	Longest      Span      `json:"longest"`
}

// Report is the outcome of one run.
type Report struct {
	ShowID    string         `json:"showId"`
	Show      string         `json:"show"`
	Started   time.Time      `json:"started"`
	Ended     time.Time      `json:"ended"` // zero while the run is going on
	Running   bool           `json:"running"`
	Totals    Totals         `json:"totals"`
	Languages map[string]int `json:"languages"` // devices that used each audience language
	Devices   []Device       `json:"devices"`
	Series    []Point        `json:"series"` // devices connected over time, from the start to the end (or now)
	Loops     []Loop         `json:"loops"`  // when a loop was running
}

// ID is the short code for a device: the same address and browser type always give the same one.
func ID(address, agent string) string {
	h := sha256.Sum256([]byte(address + "|" + agent))
	return hex.EncodeToString(h[:])[:8]
}

type device struct {
	Device
	open  int       // connections open now
	since time.Time // when the current stay began
	langs map[string]bool
}

type conn struct {
	id        string
	address   string
	agent     string
	kind      string
	lang      string
	projector bool
}

// Tracker follows the audience connections and records the run in progress.
type Tracker struct {
	mu  sync.Mutex
	now func() time.Time

	conns map[uint64]*conn // every audience connection now, in a run or not
	next  uint64

	run *run
	// OnClose is called, outside the lock, with the finished report when a run ends.
	OnClose func(Report)
}

type run struct {
	showID, show string
	started      time.Time
	devs         map[string]*device
	peak         int
	peakAt       time.Time
	series       []Point
	loops        []Loop // the last one is still running when open is true
	open         bool
}

// maxSeries bounds the points kept for the graph; beyond it every second point is dropped, which keeps the shape.
const maxSeries = 6000

// New makes a tracker that reads the clock from now (nil: time.Now).
func New(now func() time.Time) *Tracker {
	if now == nil {
		now = time.Now
	}
	return &Tracker{now: now, conns: map[uint64]*conn{}}
}

// Connect records that a device opened the audience page, and returns the function to call when it goes away.
// kind is what the program made of the browser's user-agent text: phone, tablet, computer or something else.
func (t *Tracker) Connect(address, agent, kind, lang string, projector bool) (disconnect func()) {
	t.mu.Lock()
	t.next++
	id := t.next
	c := &conn{id: ID(address, agent), address: address, agent: agent, kind: kind, lang: lang, projector: projector}
	t.conns[id] = c
	if t.run != nil {
		t.addLocked(c, t.now())
	}
	t.mu.Unlock()
	var once sync.Once
	return func() {
		once.Do(func() {
			t.mu.Lock()
			defer t.mu.Unlock()
			delete(t.conns, id)
			if t.run != nil {
				t.dropLocked(c, t.now())
			}
		})
	}
}

// Sync tells the tracker whether the show is on the air, and which one. Going on the air starts a run (counting the
// devices already connected); going off the air, or a different show, ends it.
func (t *Tracker) Sync(onAir bool, showID, show string, loop *LoopInfo) {
	t.mu.Lock()
	var closed *Report
	now := t.now()
	if t.run != nil && (!onAir || t.run.showID != showID) {
		t.markLocked(now)
		r := t.reportLocked(now, false)
		closed = &r
		t.run = nil
	}
	if onAir && t.run == nil {
		t.run = &run{showID: showID, show: show, started: now, devs: map[string]*device{}}
		ids := make([]uint64, 0, len(t.conns))
		for id := range t.conns {
			ids = append(ids, id)
		}
		sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
		for _, id := range ids {
			t.addLocked(t.conns[id], now)
		}
		t.markLocked(now)
	} else if t.run != nil {
		t.run.show = show
	}
	if t.run != nil {
		t.loopLocked(now, loop)
	}
	cb := t.OnClose
	t.mu.Unlock()
	if closed != nil && cb != nil {
		cb(*closed)
	}
}

func (t *Tracker) addLocked(c *conn, now time.Time) {
	r := t.run
	d := r.devs[c.id]
	if d == nil {
		d = &device{Device: Device{ID: c.id, Address: c.address, Kind: c.kind, First: now}, langs: map[string]bool{}}
		r.devs[c.id] = d
	}
	d.Connections++
	d.Last = now
	if c.projector {
		d.Projector = true
	}
	if c.lang != "" {
		d.langs[c.lang] = true
	}
	if d.open == 0 {
		d.since = now
	}
	d.open++
	if n := r.connectedLocked(); n > r.peak {
		r.peak, r.peakAt = n, now
	}
	t.markLocked(now)
}

// markLocked adds a point to the graph when the number of connected devices changed.
func (t *Tracker) markLocked(now time.Time) {
	r := t.run
	n := r.connectedLocked()
	if k := len(r.series); k > 0 {
		last := r.series[k-1]
		if last.N == n {
			return
		}
		if last.T.Equal(now) {
			r.series[k-1].N = n // several changes at one moment: keep the last
			return
		}
	}
	r.series = append(r.series, Point{T: now, N: n})
	if len(r.series) > maxSeries {
		kept := r.series[:0:0]
		for i, p := range r.series {
			if i%2 == 0 || i == len(r.series)-1 {
				kept = append(kept, p)
			}
		}
		r.series = kept
	}
}

// loopLocked follows which loop is running: a new one opens a stretch, none (or another) closes the one before.
func (t *Tracker) loopLocked(now time.Time, loop *LoopInfo) {
	r := t.run
	if r.open {
		cur := &r.loops[len(r.loops)-1]
		if loop != nil && loop.Cue == cur.Cue {
			return
		}
		cur.End, r.open = now, false
	}
	if loop != nil {
		r.loops = append(r.loops, Loop{Start: now, Kind: loop.Kind, Name: loop.Name, Cue: loop.Cue})
		r.open = true
	}
}

func (t *Tracker) dropLocked(c *conn, now time.Time) {
	d := t.run.devs[c.id]
	if d == nil || d.open == 0 {
		return // connected before this run began and gone before it: not part of it
	}
	d.open--
	d.Last = now
	if d.open == 0 {
		t.endStay(d, now)
	}
	t.markLocked(now)
}

func (t *Tracker) endStay(d *device, now time.Time) {
	stay := now.Sub(d.since)
	d.Connected += Span(stay)
	if Span(stay) > d.Longest {
		d.Longest = Span(stay)
	}
}

func (r *run) connectedLocked() int {
	n := 0
	for _, d := range r.devs {
		if d.open > 0 {
			n++
		}
	}
	return n
}

// Current is the report of the run in progress, or false when there is none.
func (t *Tracker) Current() (Report, bool) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.run == nil {
		return Report{}, false
	}
	return t.reportLocked(t.now(), true), true
}

// reportLocked builds the report as of now; the stays still open are counted up to now.
func (t *Tracker) reportLocked(now time.Time, running bool) Report {
	r := t.run
	rep := Report{ShowID: r.showID, Show: r.show, Started: r.started, Running: running, Languages: map[string]int{}, Devices: []Device{}}
	if !running {
		rep.Ended = now
	}
	tot := &rep.Totals
	var sum Span
	ids := make([]string, 0, len(r.devs))
	for id := range r.devs {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool {
		a, b := r.devs[ids[i]], r.devs[ids[j]]
		if !a.First.Equal(b.First) {
			return a.First.Before(b.First)
		}
		return a.ID < b.ID
	})
	for _, id := range ids {
		d := r.devs[id]
		out := d.Device
		out.Connected, out.Longest = d.Connected, d.Longest
		if d.open > 0 {
			out.Now = true
			stay := Span(now.Sub(d.since))
			out.Connected += stay
			if stay > out.Longest {
				out.Longest = stay
			}
			out.Last = now
			tot.Now++
		}
		for l := range d.langs {
			out.Languages = append(out.Languages, l)
			rep.Languages[l]++
		}
		sort.Strings(out.Languages)
		tot.Devices++
		switch {
		case d.Projector:
			tot.Projectors++
		case d.Kind == Phone:
			tot.Phones++
		case d.Kind == Tablet:
			tot.Tablets++
		default:
			tot.Computers++
		}
		tot.Connections += d.Connections
		tot.Reconnects += d.Connections - 1
		sum += out.Connected
		if out.Longest > tot.Longest {
			tot.Longest = out.Longest
		}
		rep.Devices = append(rep.Devices, out)
	}
	if tot.Devices > 0 {
		tot.AvgConnected = sum / Span(tot.Devices)
	}
	tot.Peak, tot.PeakAt = r.peak, r.peakAt
	rep.Series = append([]Point{}, r.series...)
	if n := len(rep.Series); n == 0 || !rep.Series[n-1].T.Equal(now) {
		rep.Series = append(rep.Series, Point{T: now, N: r.connectedLocked()}) // the graph runs on to now (or the end)
	}
	rep.Loops = append([]Loop{}, r.loops...)
	if r.open {
		rep.Loops[len(rep.Loops)-1].End = now // still running
	}
	return rep
}
