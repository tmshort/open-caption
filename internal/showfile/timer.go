// Copyright the open-caption authors
// SPDX-License-Identifier: Apache-2.0

package showfile

import (
	"errors"
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Timer is the countdown a loop can show with {timer}: a number of minutes counted from the moment the loop
// starts (an intermission), or a time of day (a pre-show: "7:30pm").
type Timer struct {
	Minutes float64 `json:"minutes,omitempty"` // length of the countdown; 0 when Clock is set
	Clock   string  `json:"clock,omitempty"`   // "19:30": the time of day it counts down to (the computer's local time)
	Text    string  `json:"text"`              // as written, for messages and the operator's list
}

// TimerPlaceholder is what an author writes in a loop's text where the time left should appear.
const TimerPlaceholder = "{timer}"

// timerSpanHTML is what the placeholder becomes: an empty element the page fills in every second.
const timerSpanHTML = `<span class="oc-timer" data-timer role="timer" aria-live="off"></span>`

// MaxTimerMinutes bounds a countdown given in minutes.
const MaxTimerMinutes = 600

var (
	timerMinutesRe = regexp.MustCompile(`^(\d+(?:\.\d+)?)\s*(?:m|min|mins|minute|minutes)?$`)
	timerClockRe   = regexp.MustCompile(`^(\d{1,2})(?::(\d{2}))?\s*([ap])\.?m\.?$|^(\d{1,2}):(\d{2})$`)
)

// Why a timer value was refused.
var (
	ErrTimerFormat = errors.New("not a number of minutes or a time of day")
	ErrTimerRange  = errors.New("the timer length is out of range")
)

// ParseTimer reads a timer value: "15", "15 min" or "15 minutes" (minutes), or a time of day such as "7:30pm",
// "7pm", "19:30" (a time with a colon and no am or pm is on the 24-hour clock).
func ParseTimer(v string) (*Timer, error) {
	s := strings.ToLower(strings.TrimSpace(v))
	if s == "" {
		return nil, ErrTimerFormat
	}
	if m := timerMinutesRe.FindStringSubmatch(s); m != nil {
		n, _ := strconv.ParseFloat(m[1], 64)
		if n <= 0 || n > MaxTimerMinutes || math.IsNaN(n) {
			return nil, ErrTimerRange
		}
		return &Timer{Minutes: n, Text: strings.TrimSpace(v)}, nil
	}
	if m := timerClockRe.FindStringSubmatch(s); m != nil {
		var h, mi int
		if m[4] != "" { // 24-hour
			h, _ = strconv.Atoi(m[4])
			mi, _ = strconv.Atoi(m[5])
			if h > 23 || mi > 59 {
				return nil, ErrTimerFormat
			}
		} else {
			h, _ = strconv.Atoi(m[1])
			if m[2] != "" {
				mi, _ = strconv.Atoi(m[2])
			}
			if h < 1 || h > 12 || mi > 59 {
				return nil, ErrTimerFormat
			}
			h %= 12
			if m[3] == "p" {
				h += 12
			}
		}
		return &Timer{Clock: fmt.Sprintf("%02d:%02d", h, mi), Text: strings.TrimSpace(v)}, nil
	}
	return nil, ErrTimerFormat
}

// Deadline is the moment the countdown reaches zero, in unix milliseconds, for a loop that starts at now: a
// length counts from now; a time of day is that time today (a time already past stays in the past, so the
// timer reads 0:00 rather than counting towards tomorrow).
func (t *Timer) Deadline(now time.Time) int64 {
	if t == nil {
		return 0
	}
	if t.Clock == "" {
		return now.Add(time.Duration(t.Minutes * float64(time.Minute))).UnixMilli()
	}
	h, _ := strconv.Atoi(t.Clock[:2])
	m, _ := strconv.Atoi(t.Clock[3:])
	return time.Date(now.Year(), now.Month(), now.Day(), h, m, 0, 0, now.Location()).UnixMilli()
}

// Equal reports whether two loops count the same way (for checking a translation).
func (t *Timer) Equal(o *Timer) bool {
	if t == nil || o == nil {
		return t == o
	}
	return t.Minutes == o.Minutes && t.Clock == o.Clock
}

// usesTimer reports whether text asks for the countdown.
func usesTimer(text string) bool { return strings.Contains(text, TimerPlaceholder) }
