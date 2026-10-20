// Package logbuf keeps the most recent log lines in memory so the support
// bundle can include them (REQ-061).
package logbuf

import (
	"strings"
	"sync"
)

// Buffer is an io.Writer that retains the last Max lines.
type Buffer struct {
	mu      sync.Mutex
	max     int
	lines   []string
	partial string
}

// New returns a buffer retaining up to max lines.
func New(max int) *Buffer { return &Buffer{max: max} }

// Write implements io.Writer, splitting input into lines.
func (b *Buffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	data := b.partial + string(p)
	parts := strings.Split(data, "\n")
	b.partial = parts[len(parts)-1]
	b.lines = append(b.lines, parts[:len(parts)-1]...)
	if over := len(b.lines) - b.max; over > 0 {
		b.lines = append([]string(nil), b.lines[over:]...)
	}
	return len(p), nil
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
