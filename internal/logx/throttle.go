// Copyright the open-caption authors
// SPDX-License-Identifier: Apache-2.0

package logx

import (
	"fmt"
	"log"
	"strings"
	"sync"
	"time"
)

// Throttle writes a message that can repeat many times (every refused viewer, every failed save) at most once in a
// while: the first one is written at once, the rest are counted, and the next one written says how many were left out.
// The zero value is ready to use and allows one line a minute.
type Throttle struct {
	Every time.Duration // the shortest time between two written lines (default one minute)

	mu      sync.Mutex
	last    time.Time
	skipped int
	now     func() time.Time // tests set this
}

// Warnf is logx.Warnf, limited.
func (t *Throttle) Warnf(format string, args ...any) { t.write(Warn, format, args) }

func (t *Throttle) write(l Level, format string, args []any) {
	if !Enabled(l) {
		return
	}
	every, now := t.Every, time.Now
	if every <= 0 {
		every = time.Minute
	}
	t.mu.Lock()
	if t.now != nil {
		now = t.now
	}
	at := now()
	if !t.last.IsZero() && at.Sub(t.last) < every {
		t.skipped++
		t.mu.Unlock()
		return
	}
	left := t.skipped
	t.skipped, t.last = 0, at
	t.mu.Unlock()
	msg := fmt.Sprintf(format, args...)
	if left > 0 {
		msg += fmt.Sprintf(" (and %d more like it since the last line about this)", left)
	}
	log.Output(3, strings.ToUpper(l.String())+" "+msg)
}
