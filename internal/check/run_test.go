package check

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func write(t *testing.T, dir, name, text string) string {
	t.Helper()
	p := filepath.Join(dir, name)
	os.MkdirAll(filepath.Dir(p), 0o755)
	if err := os.WriteFile(p, []byte(text), 0o644); err != nil {
		t.Fatal(err)
	}
	return p
}

func run(t *testing.T, o Options) (int, string) {
	t.Helper()
	var b bytes.Buffer
	code := Run(&b, o)
	return code, b.String()
}

const clean = "[settings]\ntitle: Demo\n[/settings]\n\n[clear]\nALICE: Hello.\nBOB: Hi!\n"

func TestACleanShowFileHasNoProblems(t *testing.T) {
	p := write(t, t.TempDir(), "show.md", clean)
	code, out := run(t, Options{Files: []string{p}})
	if code != OK || !strings.Contains(out, "0 errors, 0 warnings") || !strings.Contains(out, "3 cues") || !strings.Contains(out, "2 speakers: ALICE, BOB") {
		t.Errorf("exit %d\n%s", code, out)
	}
	if !strings.Contains(out, "images not checked") {
		t.Errorf("without an assets folder the images are not checked, and it says so:\n%s", out)
	}
}

func TestProblemsAreListedWithLinesAndSetTheExitStatus(t *testing.T) {
	p := write(t, t.TempDir(), "show.md", clean+"[song: Opening]\nla la\n[bogus: x]\n")
	code, out := run(t, Options{Files: []string{p}})
	if code != Problems {
		t.Errorf("an unclosed song is an error: exit %d", code)
	}
	for _, want := range []string{"show.md:8: error: [song] was opened but never closed", "warning: Unknown command [bogus]"} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in:\n%s", want, out)
		}
	}
	// the line is always "file:line: severity: message", so editors can jump to it
	for _, l := range strings.Split(out, "\n") {
		if strings.Contains(l, ": error: ") && !strings.HasPrefix(l, p+":") {
			t.Errorf("not in file:line form: %q", l)
		}
	}
}

func TestWarningsOnlyPassUnlessStrict(t *testing.T) {
	p := write(t, t.TempDir(), "show.md", clean+"[bogus: x]\n")
	if code, out := run(t, Options{Files: []string{p}}); code != OK || !strings.Contains(out, "1 warning") {
		t.Errorf("exit %d\n%s", code, out)
	}
	if code, _ := run(t, Options{Files: []string{p}, Strict: true}); code != Problems {
		t.Errorf("--strict: exit %d", code)
	}
}

func TestMissingImagesAreFoundInTheAssetsFolderBesideTheFile(t *testing.T) {
	dir := t.TempDir()
	p := write(t, dir, "show.md", clean+"[image: curtain.jpg]\n[image: logo.png]\n")
	write(t, dir, "assets/logo.png", "x")
	code, out := run(t, Options{Files: []string{p}})
	if !strings.Contains(out, `Image "curtain.jpg" not found in `+filepath.Join(dir, "assets")+".") || strings.Contains(out, "Upload it in the Images step") || strings.Contains(out, `"logo.png"`) || !strings.Contains(out, "missing images: curtain.jpg (line 8 of "+p+")") || !strings.Contains(out, "1 found") {
		t.Errorf("exit %d\n%s", code, out)
	}
	// an explicit folder wins over the one beside the file
	other := t.TempDir()
	write(t, other, "curtain.jpg", "x")
	write(t, other, "logo.png", "x")
	if _, out := run(t, Options{Files: []string{p}, AssetsDir: other}); strings.Contains(out, "missing images") {
		t.Errorf("both images are in %s:\n%s", other, out)
	}
}

func TestTranslationsAreComparedWithTheMainShowFile(t *testing.T) {
	dir := t.TempDir()
	main := write(t, dir, "show.md", clean)
	good := write(t, dir, "show.es.md", "[clear]\nALICE: Hola.\nBOB: ¡Hola!\n")
	bad := write(t, dir, "show.fr.md", "[clear]\nALICE: Bonjour.\n")
	code, out := run(t, Options{Files: []string{main, good}})
	if code != OK || !strings.Contains(out, "language es") {
		t.Errorf("a matching translation passes: exit %d\n%s", code, out)
	}
	code, out = run(t, Options{Files: []string{main, bad}})
	if code != Problems || !strings.Contains(out, "show.fr.md:") || !strings.Contains(out, "error") {
		t.Errorf("a translation with a different shape is an error: exit %d\n%s", code, out)
	}
	if code, out := run(t, Options{Files: []string{main, good, write(t, dir, "again.es.md", "[clear]\nA: x\nB: y\n")}}); code != Problems || !strings.Contains(out, `already used`) {
		t.Errorf("two files for one language: exit %d\n%s", code, out)
	}
}

func TestStdinAndUnreadableFiles(t *testing.T) {
	code, out := run(t, Options{Files: []string{"-"}, Stdin: strings.NewReader(clean)})
	if code != OK || !strings.Contains(out, "3 cues") {
		t.Errorf("stdin: exit %d\n%s", code, out)
	}
	if code, out := run(t, Options{Files: []string{filepath.Join(t.TempDir(), "nope.md")}}); code != Failed || !strings.Contains(out, "cannot read") {
		t.Errorf("missing file: exit %d\n%s", code, out)
	}
	if code, _ := run(t, Options{}); code != Failed {
		t.Errorf("no file: exit %d", code)
	}
}

func TestImageUseAndMissingImagesAreShared(t *testing.T) {
	// the editor (internal/httpapi) and the command line use these two functions
	if got := MissingImages(map[string][]int{"b.png": {3}, "a.png": {1, 2}}, map[string]bool{}); len(got) != 2 || got[0].Name != "a.png" || got[1].Name != "b.png" {
		t.Errorf("%+v", got)
	}
	if got := MissingImages(nil, nil); got == nil {
		t.Error("never nil: the editor sends [] in JSON")
	}
}
