package check

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/tmshort/open-caption/internal/show"
	"github.com/tmshort/open-caption/internal/showfile"
)

// Options for Run.
type Options struct {
	// Files are the show files: the first is the main one, the rest are translations of it.
	Files []string
	// AssetsDir is the folder of images. Empty means the "assets" folder beside the main file, if there is one;
	// without a folder the images are not checked.
	AssetsDir string
	// Strict makes warnings fail the check too (the exit status is 1).
	Strict bool
	// Stdin is read when a file is "-".
	Stdin io.Reader
}

// Exit statuses of Run.
const (
	OK       = 0 // no errors (and, with Strict, no warnings)
	Problems = 1 // the show file has errors
	Failed   = 2 // a file could not be read, or the command was used wrongly
)

var langFromName = regexp.MustCompile(`(?i)\.([a-z]{2,3}(?:-[a-z0-9]{2,8})?)\.(?:md|txt)$`)

// Run checks the show files and prints a report to w, in the form editors understand (file:line: severity: message).
// It returns the exit status.
func Run(w io.Writer, o Options) int {
	if len(o.Files) == 0 {
		fmt.Fprintln(w, "check-show: no show file given")
		return Failed
	}
	var assets []show.Asset
	assetsNote := ""
	dir := o.AssetsDir
	if dir == "" && o.Files[0] != "-" {
		if d := filepath.Join(filepath.Dir(o.Files[0]), "assets"); isDir(d) {
			dir = d
		}
	}
	have := map[string]bool{}
	var names []string
	checkImages := dir != ""
	if checkImages {
		var err error
		if assets, err = show.ListAssetsDir(dir); err != nil {
			fmt.Fprintf(w, "check-show: cannot read the images folder %s: %v\n", dir, err)
			return Failed
		}
		for _, a := range assets {
			have[a.Name] = true
			names = append(names, a.Name)
		}
		assetsNote = fmt.Sprintf("images checked against %s (%d found)", dir, len(assets))
	} else {
		assetsNote = "images not checked (no assets folder beside the show file; give one with --assets FOLDER)"
	}

	errs, warns, notes := 0, 0, 0
	var missingImages []string
	var mainRes *showfile.Result
	mainName := ""
	seenLang := map[string]string{}
	for i, f := range o.Files {
		text, err := readFile(f, o.Stdin)
		if err != nil {
			fmt.Fprintf(w, "check-show: %v\n", err)
			return Failed
		}
		translation := i > 0
		opts := showfile.Options{Translation: translation}
		if checkImages {
			opts.Assets = names
			if names == nil {
				opts.Assets = []string{}
			}
		}
		res := showfile.Parse(text, opts)
		diags := append([]showfile.Diagnostic(nil), res.Diagnostics...)
		lang := res.Language
		if lang == "" {
			if m := langFromName.FindStringSubmatch(filepath.Base(f)); m != nil && translation {
				lang = strings.ToLower(m[1])
			}
		}
		label := f
		if f == "-" {
			label = "<stdin>"
		}
		if translation {
			if lang == "" {
				diags = append(diags, showfile.Diagnostic{Line: 1, Severity: showfile.SevWarning,
					Message: "cannot tell which language this translation is: name the file like show.es.md or add a [language: es] line"})
			} else if other, dup := seenLang[lang]; dup {
				diags = append(diags, showfile.Diagnostic{Line: 1, Severity: showfile.SevError,
					Message: fmt.Sprintf("the language %q is already used by %s", lang, other)})
			}
			seenLang[lang] = label
			if issues := showfile.CompareStructure(mainRes.Cues, res.Cues, mainName); len(issues) > 0 {
				diags = append(issues, diags...)
			}
		} else {
			mainRes = &res
			mainName = res.LanguageName
			if mainName == "" && res.Language != "" {
				mainName = showfile.LanguageName(res.Language)
			}
			if mainName == "" {
				mainName = "the main show file"
			}
			if lang != "" {
				seenLang[lang] = label
			}
			if !HasTitle(text) {
				diags = append(diags, showfile.Diagnostic{Line: 1, Severity: showfile.SevInfo,
					Message: `this show has no title, so it is listed as "Untitled show": add a [settings] block with a title: line`})
			}
		}
		if checkImages {
			for _, m := range MissingImages(ImageUse(res.Cues), have) {
				missingImages = append(missingImages, fmt.Sprintf("%s (line%s %s of %s)", m.Name, plural(len(m.Lines)), joinInts(m.Lines), label))
			}
		}
		sort.SliceStable(diags, func(a, b int) bool { return diags[a].Line < diags[b].Line })
		for _, d := range diags {
			if d.Line > 0 {
				fmt.Fprintf(w, "%s:%d: %s: %s\n", label, d.Line, d.Severity, d.Message)
			} else {
				fmt.Fprintf(w, "%s: %s: %s\n", label, d.Severity, d.Message) // about the whole file
			}
			switch d.Severity {
			case showfile.SevError:
				errs++
			case showfile.SevWarning:
				warns++
			default:
				notes++
			}
		}
		fmt.Fprintf(w, "%s: %s\n\n", label, summary(res, translation, lang))
	}
	fmt.Fprintln(w, assetsNote)
	if len(missingImages) > 0 {
		fmt.Fprintf(w, "missing images: %s\n", strings.Join(missingImages, "; "))
	}
	fmt.Fprintf(w, "%s, %s, %s\n", count(errs, "error"), count(warns, "warning"), count(notes, "note"))
	switch {
	case errs > 0:
		return Problems
	case warns > 0 && o.Strict:
		return Problems
	}
	return OK
}

func readFile(f string, stdin io.Reader) (string, error) {
	var b []byte
	var err error
	if f == "-" {
		if stdin == nil {
			stdin = os.Stdin
		}
		b, err = io.ReadAll(stdin)
	} else {
		b, err = os.ReadFile(f)
	}
	if err != nil {
		return "", fmt.Errorf("cannot read %s: %v", f, err)
	}
	return string(b), nil
}

func isDir(p string) bool { st, err := os.Stat(p); return err == nil && st.IsDir() }

func plural(n int) string {
	if n == 1 {
		return ""
	}
	return "s"
}

func count(n int, word string) string { return fmt.Sprintf("%d %s%s", n, word, plural(n)) }

func joinInts(a []int) string {
	s := make([]string, len(a))
	for i, n := range a {
		s[i] = fmt.Sprint(n)
	}
	return strings.Join(s, ", ")
}

// summary is the one-line description of what the show file contains (the parser's own note, printed above, counts the
// kinds of cue; this adds the cue count, the speakers and the language).
func summary(res showfile.Result, translation bool, lang string) string {
	s := fmt.Sprintf("%s", count(len(res.Cues), "cue"))
	var more []string
	if len(res.Speakers) > 0 {
		more = append(more, count(len(res.Speakers), "speaker")+": "+strings.Join(res.Speakers, ", "))
	}
	if translation && lang != "" {
		more = append(more, "language "+lang)
	} else if !translation && res.Language != "" {
		more = append(more, "language "+res.Language)
	}
	if len(more) > 0 {
		s += "; " + strings.Join(more, "; ")
	}
	return s
}
