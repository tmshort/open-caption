// Copyright the open-caption authors
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestHashPasswordPrintsADigestOfTheFirstLine(t *testing.T) {
	var out, errOut bytes.Buffer
	if code := hashPassword(strings.NewReader("secret1\r\nignored\n"), &out, &errOut, false); code != 0 {
		t.Fatalf("exit %d: %s", code, errOut.String())
	}
	if got, want := out.String(), "sha256:5b11618c2e44027877d0cd0921ed166b9f176f50587fc91e7534dd2946db77d6\n"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
	if errOut.Len() != 0 {
		t.Errorf("nothing but the digest on a pipe: %q", errOut.String())
	}
	out.Reset()
	if code := hashPassword(strings.NewReader("no newline at the end"), &out, &errOut, false); code != 0 || !strings.HasPrefix(out.String(), "sha256:") {
		t.Errorf("a last line without a newline still counts: %d %q", code, out.String())
	}
	out.Reset()
	if code := hashPassword(strings.NewReader("\n"), &out, &errOut, true); code != 1 || out.Len() != 0 || !strings.Contains(errOut.String(), "empty") || !strings.Contains(errOut.String(), "Password to hash") {
		t.Errorf("an empty password is refused, and a terminal is prompted: %d %q %q", code, out.String(), errOut.String())
	}
}
