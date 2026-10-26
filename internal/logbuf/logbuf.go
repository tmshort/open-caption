// Copyright the open-caption authors
// SPDX-License-Identifier: Apache-2.0

// Package logbuf keeps the most recent log lines in memory so the support
// bundle can include them (REQ-061).
package logbuf

import (
	"regexp"
	"strings"
	"sync"
)

// Buffer is an io.Writer that retains the last Max lines.
type Buffer struct {
	mu       sync.Mutex
	max      int
	lines    []string
	partial  string
	problems int
}

// problemRe matches a log line written at warn or error level ("2026/10/26 10:19:19 WARN …").
var problemRe = regexp.MustCompile(`^[0-9/: .]*(WARN|ERROR) `)

// New returns a buffer retaining up to max lines.
func New(max int) *Buffer { return &Buffer{max: max} }

// Write implements io.Writer, splitting input into lines.
func (b *Buffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	data := b.partial + string(p)
	parts := strings.Split(data, "\n")
	b.partial = parts[len(parts)-1]
	for _, ln := range parts[:len(parts)-1] {
		if problemRe.MatchString(ln) {
			b.problems++
		}
	}
	b.lines = append(b.lines, parts[:len(parts)-1]...)
	if over := len(b.lines) - b.max; over > 0 {
		b.lines = append([]string(nil), b.lines[over:]...)
	}
	return len(p), nil
}

// Problems is how many warning and error lines have been written since the program started, including lines that
// have since scrolled out of the buffer, so Setup can tell how many are new since it last looked.
func (b *Buffer) Problems() int {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.problems
}

// Lines returns the retained lines, oldest first.
func (b *Buffer) Lines() []string {
	b.mu.Lock()
	defer b.mu.Unlock()
	out := append([]string(nil), b.lines...)
	if b.partial != "" {
		out = append(out, b.partial)
	}
	return out
}

// String returns the retained lines, oldest first.
func (b *Buffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	s := strings.Join(b.lines, "\n")
	if b.partial != "" {
		s += "\n" + b.partial
	}
	return s + "\n"
}
