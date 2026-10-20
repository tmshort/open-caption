// Copyright the open-caption authors
// SPDX-License-Identifier: Apache-2.0

package showfile

import (
	"strings"
	"testing"
	"time"
)

func TestParseTimer(t *testing.T) {
	ok := map[string]Timer{
		"15":         {Minutes: 15},
		"15 min":     {Minutes: 15},
		"15 Minutes": {Minutes: 15},
		"7.5m":       {Minutes: 7.5},
		"7:30pm":     {Clock: "19:30"},
		"7:30 PM":    {Clock: "19:30"},
		"7pm":        {Clock: "19:00"},
		"12am":       {Clock: "00:00"},
		"12:15 p.m.": {Clock: "12:15"},
		"19:30":      {Clock: "19:30"},
		"7:05":       {Clock: "07:05"},
	}
	for in, want := range ok {
		got, err := ParseTimer(in)
		if err != nil || got.Minutes != want.Minutes || got.Clock != want.Clock {
			t.Errorf("%q: %+v %v, want %+v", in, got, err, want)
		}
	}
	for _, in := range []string{"", "soon", "25:00", "13pm", "0", "7:75pm", "-5", "601"} {
		if _, err := ParseTimer(in); err == nil {
			t.Errorf("%q should be refused", in)
		}
	}
}

func TestTimerDeadline(t *testing.T) {
	now := time.Date(2026, 10, 16, 18, 45, 10, 0, time.Local)
	d, _ := ParseTimer("15")
	if got := d.Deadline(now); got != now.Add(15*time.Minute).UnixMilli() {
		t.Errorf("minutes: %d", got)
	}
	c, _ := ParseTimer("7:30pm")
	if got, want := c.Deadline(now), time.Date(2026, 10, 16, 19, 30, 0, 0, time.Local).UnixMilli(); got != want {
		t.Errorf("clock: %d want %d", got, want)
	}
	late := time.Date(2026, 10, 16, 20, 0, 0, 0, time.Local)
	if c.Deadline(late) >= late.UnixMilli() {
		t.Error("a time already past must stay in the past")
	}
}

func TestLoopTimerParsing(t *testing.T) {
	src := "[loop: 5s, timer 15 minutes]\nINTERMISSION\nBack in {timer}\n[image: a.png]\n[/loop]\n"
	r := Parse(src, Options{Assets: []string{"a.png"}})
	if len(r.Cues) != 1 || r.Cues[0].Timer == nil || r.Cues[0].Timer.Minutes != 15 || r.Cues[0].Period != 5 {
		t.Fatalf("cue: %+v %v", r.Cues, r.Diagnostics)
	}
	its := r.Cues[0].Items
	if its[0].Timer || !its[1].Timer || !strings.Contains(its[1].HTML, `data-timer`) || strings.Contains(its[1].HTML, "{timer}") {
		t.Errorf("items: %+v", its)
	}
	for _, d := range r.Diagnostics {
		if d.Severity != SevInfo {
			t.Errorf("unexpected %s: %s", d.Severity, d.Message)
		}
	}
	// the other ways to write it
	for _, src := range []string{"[loop: timer 7:30pm]\nGo {timer}\n[/loop]", "[loop: Timer: 15]\nx {timer}\n[/loop]", "[loop: countdown 10]\n[fullscreen: {timer}]\n[/loop]"} {
		r := Parse(src, Options{})
		if len(r.Cues) != 1 || r.Cues[0].Timer == nil {
			t.Errorf("%q: %+v", src, r.Diagnostics)
		}
	}
}

func TestLoopTimerProblems(t *testing.T) {
	cases := map[string]string{
		"[loop: timer soon]\nx {timer}\n[/loop]": "is not a number of minutes",
		"[loop: timer 9999]\nx {timer}\n[/loop]": "at most 600 minutes",
		"[loop: timer]\nx {timer}\n[/loop]":      "is not a number of minutes",
		"[loop]\nBack in {timer}\n[/loop]":       "needs a timer on its loop",
		"[loop: timer 15]\nplain text\n[/loop]":  "never shown",
		"[loop: sometimes]\nplain text\n[/loop]": "expected a time like 10s",
	}
	for src, want := range cases {
		r := Parse(src, Options{})
		found := false
		for _, d := range r.Diagnostics {
			found = found || (d.Severity == SevWarning && strings.Contains(d.Message, want))
		}
		if !found {
			t.Errorf("%q: want a warning containing %q, got %v", src, want, r.Diagnostics)
		}
	}
}

func TestTranslationMustCountDownTheSameWay(t *testing.T) {
	a := Parse("[loop: timer 15]\nBack in {timer}\n[/loop]\nHi", Options{})
	b := Parse("[loop: timer 10]\nVuelva en {timer}\n[/loop]\nHola", Options{})
	if len(CompareStructure(a.Cues, b.Cues, "English")) == 0 {
		t.Error("different timers should be reported")
	}
	c := Parse("[loop: timer 15 min]\nVuelva en {timer}\n[/loop]\nHola", Options{})
	if issues := CompareStructure(a.Cues, c.Cues, "English"); len(issues) != 0 {
		t.Errorf("same timer written differently is fine: %v", issues)
	}
}
