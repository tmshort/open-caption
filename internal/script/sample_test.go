package script

import (
	"os"
	"path/filepath"
	"testing"
)

// The shipped sample script must parse with no warnings or errors, given its images.
func TestSampleScriptParsesClean(t *testing.T) {
	dir := filepath.Join("..", "..", "web", "sample")
	b, err := os.ReadFile(filepath.Join(dir, "sample-script.md"))
	if err != nil {
		t.Fatal(err)
	}
	ents, _ := os.ReadDir(filepath.Join(dir, "assets"))
	var assets []string
	for _, e := range ents {
		assets = append(assets, e.Name())
	}
	r := Parse(string(b), Options{Assets: assets})
	for _, d := range r.Diagnostics {
		if d.Severity != SevInfo {
			t.Errorf("line %d: %s: %s", d.Line, d.Severity, d.Message)
		}
	}
	counts := map[CueKind]int{}
	for _, c := range r.Cues {
		counts[c.Kind]++
	}
	for _, k := range []CueKind{KindCaption, KindClear, KindBlank, KindImage, KindFullscreen, KindLoop, KindSection, KindNote} {
		if counts[k] == 0 {
			t.Errorf("sample script should demonstrate %s", k)
		}
	}
	if r.Cues[0].Kind != KindLoop {
		t.Error("sample should open with a pre-show loop")
	}
}
