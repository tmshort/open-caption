package logx

import (
	"bytes"
	"log"
	"strings"
	"testing"
)

func capture(t *testing.T) *bytes.Buffer {
	var b bytes.Buffer
	log.SetOutput(&b)
	log.SetFlags(0)
	t.Cleanup(func() { log.SetOutput(nil); SetLevel(Info) })
	return &b
}

func TestLevelsFilterAndLabel(t *testing.T) {
	b := capture(t)
	SetLevel(Warn)
	Debugf("d %d", 1)
	Infof("i")
	Warnf("w %s", "x")
	Errorf("e")
	if got := b.String(); got != "WARN w x\nERROR e\n" {
		t.Errorf("%q", got)
	}
	b.Reset()
	SetLevel(Debug)
	Debugf("d")
	if !strings.HasPrefix(b.String(), "DEBUG d") || !Enabled(Debug) {
		t.Errorf("%q", b.String())
	}
}

func TestParse(t *testing.T) {
	for in, want := range map[string]Level{"": Info, "INFO": Info, "debug": Debug, " Warning ": Warn, "warn": Warn, "error": Error} {
		if got, err := Parse(in); err != nil || got != want {
			t.Errorf("%q: %v %v", in, got, err)
		}
	}
	if _, err := Parse("verbose"); err == nil || !strings.Contains(err.Error(), "debug, info, warn or error") {
		t.Errorf("%v", err)
	}
}

func TestFatalAlwaysWritesAndExits(t *testing.T) {
	b := capture(t)
	SetLevel(Error)
	code := 0
	exit = func(c int) { code = c }
	orig := exit
	t.Cleanup(func() { exit = orig })
	Fatalf("boom")
	if code != 1 || b.String() != "ERROR boom\n" {
		t.Errorf("%d %q", code, b.String())
	}
}
