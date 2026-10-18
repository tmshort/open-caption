package httpapi

import (
	"os"
	"path/filepath"
	"testing"
)

// A script on this computer that keeps no cookies is signed in again on every call; that must not pile up
// sessions or write the sessions file each time.
func TestAutomaticSessionsAreCappedAndNotSaved(t *testing.T) {
	path := filepath.Join(t.TempDir(), "sessions.json")
	a := newAuth("a", "o", path)
	for i := 0; i < 5*maxAutoSessions; i++ {
		if _, ok := a.issue(RoleAdmin, true); !ok {
			t.Fatal("issue failed")
		}
	}
	if n := len(a.sessions); n != maxAutoSessions {
		t.Errorf("%d automatic sessions kept, want %d", n, maxAutoSessions)
	}
	if _, err := os.Stat(path); err == nil {
		t.Error("automatic sessions must not be written to disk")
	}
	tok, _ := a.issue(RoleOperator, false) // a real sign-in is kept and saved
	if _, ok := a.lookup(tok); !ok {
		t.Error("password session lost")
	}
	if _, err := os.Stat(path); err != nil {
		t.Errorf("password session not saved: %v", err)
	}
}
