package logbuf

import (
	"strings"
	"testing"
)

func TestKeepsLastLines(t *testing.T) {
	b := New(3)
	for _, s := range []string{"a\n", "b\nc", "\nd\ne\n"} {
		b.Write([]byte(s))
	}
	if got := b.String(); got != "c\nd\ne\n" {
		t.Errorf("%q", got)
	}
	b.Write([]byte("partial"))
	if !strings.HasSuffix(b.String(), "partial\n") {
		t.Errorf("%q", b.String())
	}
}

func TestLinesReturnsWholeLinesOldestFirst(t *testing.T) {
	b := New(3)
	b.Write([]byte("a\nb\nc\nd\npart"))
	if got := b.Lines(); len(got) != 4 || got[0] != "b" || got[3] != "part" {
		t.Errorf("%q", got)
	}
}
