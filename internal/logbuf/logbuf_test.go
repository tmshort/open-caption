// Copyright the open-caption authors
// SPDX-License-Identifier: Apache-2.0

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

func TestProblemsCountsWarningsAndErrorsEvenAfterTheyScrollOut(t *testing.T) {
	b := New(2)
	b.Write([]byte("2026/10/26 10:19:19 INFO fine\n2026/10/26 10:19:20 WARN odd\n2026/10/26 10:19:21 ERROR bad\n"))
	b.Write([]byte("2026/10/26 10:19:22 DEBUG detail\n2026/10/26 10:19:23 INFO fine again\nWARN"))
	if got := b.Problems(); got != 2 {
		t.Errorf("problems = %d, want 2 (the unfinished last line is not counted yet)", got)
	}
	if len(b.Lines()) > 3 {
		t.Errorf("%v", b.Lines())
	}
}
