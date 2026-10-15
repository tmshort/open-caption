package logx

import (
	"bytes"
	"log"
	"os"
	"path/filepath"
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

func TestLogFileRotates(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "oc.log")
	l, err := OpenFile(path, 100, 2)
	if err != nil {
		t.Fatal(err)
	}
	line := strings.Repeat("x", 39) + "\n" // 40 bytes
	for i := 0; i < 12; i++ {
		l.Write([]byte(line))
	}
	l.Close()
	size := func(p string) int64 {
		fi, err := os.Stat(p)
		if err != nil {
			return -1
		}
		return fi.Size()
	}
	if size(path) < 0 || size(path) > 100 || size(path+".1") < 0 || size(path+".2") < 0 || size(path+".3") != -1 {
		t.Errorf("sizes: %d %d %d %d", size(path), size(path+".1"), size(path+".2"), size(path+".3"))
	}
	for _, p := range []string{path, path + ".1", path + ".2"} {
		b, _ := os.ReadFile(p)
		if len(b)%40 != 0 {
			t.Errorf("%s holds a split line (%d bytes)", p, len(b))
		}
	}
	// Reopening continues the same file.
	l2, _ := OpenFile(path, 100, 2)
	l2.Write([]byte("more\n"))
	l2.Close()
	if b, _ := os.ReadFile(path); !strings.HasSuffix(string(b), "more\n") {
		t.Error("appended")
	}
}
