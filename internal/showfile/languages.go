package showfile

import (
	"fmt"
	"regexp"
	"strings"
)

// LangCodeRe is the shape of a language code: "es", "fil", or "pt-br".
var LangCodeRe = regexp.MustCompile(`^[a-z]{2,3}(-[a-z0-9]{2,8})?$`)

// KnownLanguages are names offered in Setup for common languages (written in the language itself).
var KnownLanguages = []struct{ Code, Name string }{
	{"en", "English"}, {"es", "Español"}, {"fr", "Français"}, {"de", "Deutsch"}, {"it", "Italiano"},
	{"pt", "Português"}, {"nl", "Nederlands"}, {"sv", "Svenska"}, {"da", "Dansk"}, {"nb", "Norsk"},
	{"fi", "Suomi"}, {"pl", "Polski"}, {"cs", "Čeština"}, {"hu", "Magyar"}, {"ro", "Română"},
	{"el", "Ελληνικά"}, {"ru", "Русский"}, {"uk", "Українська"}, {"tr", "Türkçe"}, {"ar", "العربية"},
	{"he", "עברית"}, {"hi", "हिन्दी"}, {"zh", "中文"}, {"ja", "日本語"}, {"ko", "한국어"},
	{"vi", "Tiếng Việt"}, {"th", "ไทย"}, {"id", "Bahasa Indonesia"}, {"tl", "Filipino"},
}

// LanguageName returns a display name for a language code, or the code itself if unknown.
func LanguageName(code string) string {
	base := code
	if i := strings.IndexByte(code, '-'); i >= 0 {
		base = code[:i]
	}
	for _, l := range KnownLanguages {
		if l.Code == code {
			return l.Name
		}
	}
	for _, l := range KnownLanguages {
		if l.Code == base {
			return l.Name + " (" + strings.ToUpper(code[len(base)+1:]) + ")"
		}
	}
	return code
}

const maxStructureIssues = 12

// CompareStructure checks that a translation has exactly the same shape as the primary show file:
// the same number of cues, in the same order, of the same kinds (a caption stays a caption, a loop
// has the same number and kinds of items, voices shown together stay the same number). The words may
// differ. The returned diagnostics use the translation's line numbers.
//
// One operator keypress sends the same cue number in every language, so a mismatch would show the
// wrong line (or run off the end) in that language.
func CompareStructure(primary, other []Cue, primaryName string) []Diagnostic {
	var out []Diagnostic
	add := func(line int, format string, a ...any) {
		if len(out) < maxStructureIssues {
			out = append(out, Diagnostic{Line: line, Severity: SevError, Message: fmt.Sprintf(format, a...)})
		}
	}
	describe := func(c Cue) string {
		switch c.Kind {
		case KindCaption:
			if len(c.Parts) > 1 {
				return fmt.Sprintf("%d voices shown together", len(c.Parts))
			}
			return "a caption"
		case KindClear:
			return "[clear]"
		case KindBlank:
			return "[blank]"
		case KindImage:
			return "an [image]"
		case KindFullscreen:
			return "a [fullscreen] message"
		case KindLoop:
			return fmt.Sprintf("a [loop] of %d items", len(c.Items))
		case KindSection:
			return "a [section]"
		case KindNote:
			return "a [note]"
		}
		return string(c.Kind)
	}
	n := min(len(primary), len(other))
	for i := 0; i < n; i++ {
		p, o := primary[i], other[i]
		same := p.Kind == o.Kind
		if same && p.Kind == KindCaption {
			same = max(1, len(p.Parts)) == max(1, len(o.Parts))
		}
		if same && p.Kind == KindLoop {
			same = len(p.Items) == len(o.Items)
			for j := 0; same && j < len(p.Items); j++ {
				same = p.Items[j].Kind == o.Items[j].Kind
			}
		}
		if !same {
			add(o.Line, "Cue %d does not match the %s show file: that show file has %s here, and this one has %s. "+
				"Every language must have the same cues in the same order; only the words may differ.", i+1, primaryName, describe(p), describe(o))
			if len(out) >= 1 { // after the first difference everything after it is shifted, so stop there
				break
			}
		}
	}
	if len(out) == 0 && len(primary) != len(other) {
		line := 0
		if len(other) > 0 {
			line = other[len(other)-1].Line
		}
		switch {
		case len(other) < len(primary):
			add(line, "This show file has %d cues but the %s show file has %d. Missing: %s after cue %d.",
				len(other), primaryName, len(primary), describe(primary[len(other)]), len(other))
		default:
			add(other[len(primary)].Line, "This show file has %d cues but the %s show file has %d. This line has no match there.",
				len(other), primaryName, len(primary))
		}
	}
	return out
}
