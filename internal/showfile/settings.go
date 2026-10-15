package showfile

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// Settings are the show-wide options a show file may set in a [settings] block. A nil field means
// "not set here": the value then comes from config.yaml's defaults, and those from the program's
// built-in defaults. So the layers are: built-in defaults, then config.yaml, then the show file.
type Settings struct {
	Title           *string  `json:"title,omitempty"`
	Background      *string  `json:"background,omitempty"`
	TextColor       *string  `json:"textColor,omitempty"`
	SpeakerColor    *string  `json:"speakerColor,omitempty"`
	Font            *string  `json:"font,omitempty"`
	TextSize        *int     `json:"textSize,omitempty"` // percent
	SpeakerNames    *string  `json:"speakerNames,omitempty"`
	IdleMessage     *string  `json:"idleMessage,omitempty"`
	LoopSeconds     *float64 `json:"loopSeconds,omitempty"`
	ScrollbackLines *int     `json:"scrollbackLines,omitempty"`
	Archived        *bool    `json:"archived,omitempty"` // set only when true

	// Extra holds lines of a [settings] block this version does not understand, kept so that
	// rewriting the block never throws away what someone typed.
	Extra []string `json:"-"`
}

// TitleOr returns the title, or fallback if the show file sets none.
func (s Settings) TitleOr(fallback string) string {
	if s.Title != nil {
		return *s.Title
	}
	return fallback
}

// LookSet reports whether any of the look settings (colors, font, size, speaker names) are set.
func (s Settings) LookSet() bool {
	return s.Background != nil || s.TextColor != nil || s.SpeakerColor != nil || s.Font != nil || s.TextSize != nil || s.SpeakerNames != nil
}

// setting describes one key of the [settings] block.
type setting struct {
	name    string   // the name written to files
	aliases []string // other spellings accepted
	help    string
	// translated: a translation's show file may set it (only the pre-show message).
	translated bool
}

var settingTable = []setting{
	{name: "title", aliases: []string{"name", "show title", "show name"}, help: "The show's name, as listed in Setup."},
	{name: "background", aliases: []string{"background color", "background colour", "bg"}, help: "Screen color behind the captions, like #000000."},
	{name: "text color", aliases: []string{"text colour", "foreground", "foreground color", "fg"}, help: "Color of the caption text, like #ffffff."},
	{name: "speaker name color", aliases: []string{"speaker color", "speaker colour", "accent", "accent color", "speaker name colour"}, help: "Color of speaker names (when not one color per character)."},
	{name: "font", aliases: []string{"typeface"}, help: "system, serif, mono, dyslexic or hyperlegible."},
	{name: "text size", aliases: []string{"size", "font size", "scale"}, help: "Caption size as a percentage, 50 to 300 (100 is normal)."},
	{name: "speaker names", aliases: []string{"speakers", "speaker style"}, help: "label (one color), color (a different color for each character) or hidden."},
	{name: "idle message", aliases: []string{"idle", "pre-show message", "preshow message", "message before the show starts"}, translated: true, help: "Text shown before the first line is sent."},
	{name: "loop seconds", aliases: []string{"loop time", "default loop seconds", "loop item seconds"}, help: "How long each item of a [loop] stays when it does not say."},
	{name: "scrollback lines", aliases: []string{"scrollback", "lines kept"}, help: "How many captions a viewer can scroll back through (20 to 2000)."},
	{name: "archived", aliases: []string{"archive"}, help: "yes if the show is archived: kept, but out of the way in Setup and not able to go on air until it is restored."},
}

// SettingNames lists the names written to show files, in the order they are written. Help is tested
// against this list so a setting cannot be added without being documented.
func SettingNames() []string {
	out := make([]string, len(settingTable))
	for i, s := range settingTable {
		out[i] = s.name
	}
	return out
}

// SettingHelp returns the one-line description of a setting.
func SettingHelp(name string) string {
	for _, s := range settingTable {
		if s.name == name {
			return s.help
		}
	}
	return ""
}

func normKey(k string) string {
	k = strings.ToLower(strings.TrimSpace(k))
	k = strings.NewReplacer("_", " ", "-", " ").Replace(k)
	return strings.Join(strings.Fields(k), " ")
}

func lookupSetting(key string) (setting, bool) {
	k := normKey(key)
	for _, s := range settingTable {
		if k == normKey(s.name) {
			return s, true
		}
		for _, a := range s.aliases {
			if k == normKey(a) {
				return s, true
			}
		}
	}
	return setting{}, false
}

// ---- value normalizers (also used to validate config.yaml defaults) ----

var hexColorRe = regexp.MustCompile(`^#([0-9a-fA-F]{3}|[0-9a-fA-F]{6})$`)

var namedColors = map[string]string{
	"black": "#000000", "white": "#ffffff", "yellow": "#ffee00", "red": "#ff3b30", "green": "#34c759", "blue": "#0a84ff",
	"cyan": "#00e5ff", "magenta": "#ff40ff", "orange": "#ff9500", "gray": "#808080", "grey": "#808080",
}

// NormalizeColor accepts #rgb, #rrggbb or a basic color name and returns #rrggbb.
func NormalizeColor(v string) (string, error) {
	v = strings.ToLower(strings.TrimSpace(v))
	if c, ok := namedColors[v]; ok {
		return c, nil
	}
	if !hexColorRe.MatchString(v) {
		return "", fmt.Errorf("%q is not a color: write it like #ffffff or use a name such as white, black or yellow", v)
	}
	if len(v) == 4 {
		v = "#" + string([]byte{v[1], v[1], v[2], v[2], v[3], v[3]})
	}
	return v, nil
}

var fontAliases = map[string]string{
	"system": "system", "default": "system", "clean": "system", "sans": "system", "sans serif": "system",
	"serif": "serif", "mono": "mono", "monospace": "mono",
	"dyslexic": "dyslexic", "dyslexia": "dyslexic", "dyslexia friendly": "dyslexic", "opendyslexic": "dyslexic", "open dyslexic": "dyslexic",
	"hyperlegible": "hyperlegible", "high legibility": "hyperlegible", "legible": "hyperlegible", "atkinson": "hyperlegible", "atkinson hyperlegible": "hyperlegible",
}

// NormalizeFont returns system, serif, mono, dyslexic or hyperlegible.
func NormalizeFont(v string) (string, error) {
	if f, ok := fontAliases[normKey(v)]; ok {
		return f, nil
	}
	return "", fmt.Errorf("%q is not a font: use system, serif, mono, dyslexic or hyperlegible", v)
}

var speakerAliases = map[string]string{
	"label": "label", "one color": "label", "one colour": "label", "show": "label", "above": "label", "same color": "label",
	"color": "color", "colour": "color", "colors": "color", "colours": "color", "different colors": "color", "per character": "color", "each character": "color",
	"hidden": "hidden", "hide": "hidden", "off": "hidden", "none": "hidden", "no": "hidden",
}

// NormalizeSpeakerNames returns label, color or hidden.
func NormalizeSpeakerNames(v string) (string, error) {
	if f, ok := speakerAliases[normKey(v)]; ok {
		return f, nil
	}
	return "", fmt.Errorf("%q is not a speaker-name style: use label (one color), color (a different color for each character) or hidden", v)
}

// NormalizeTextSize accepts "120", "120%" and returns the percentage (50 to 300).
func NormalizeTextSize(v string) (int, error) {
	n, err := strconv.Atoi(strings.TrimSpace(strings.TrimSuffix(strings.TrimSpace(v), "%")))
	if err != nil || n < 50 || n > 300 {
		return 0, fmt.Errorf("%q is not a text size: use a percentage from 50 to 300, such as 120%%", v)
	}
	return n, nil
}

// NormalizeLoopSeconds accepts "10" or "10s" (more than 0, at most 3600).
func NormalizeLoopSeconds(v string) (float64, error) {
	f, err := strconv.ParseFloat(strings.TrimSpace(strings.TrimSuffix(strings.TrimSpace(strings.ToLower(v)), "s")), 64)
	if err != nil || f <= 0 || f > 3600 {
		return 0, fmt.Errorf("%q is not a time: use a number of seconds such as 10", v)
	}
	return f, nil
}

// NormalizeScrollback accepts a whole number from 20 to 2000.
func NormalizeScrollback(v string) (int, error) {
	n, err := strconv.Atoi(strings.TrimSpace(v))
	if err != nil || n < 20 || n > 2000 {
		return 0, fmt.Errorf("%q is not a line count: use a whole number from 20 to 2000", v)
	}
	return n, nil
}

// MaxIdleMessage is the longest pre-show message.
const MaxIdleMessage = 200

// set stores one key's value in s, returning a problem description if the value is unusable.
func (s *Settings) set(name, value string) error {
	value = strings.TrimSpace(value)
	switch name {
	case "title":
		if value == "" || len(value) > 100 {
			return fmt.Errorf("the title must be 1 to 100 characters")
		}
		s.Title = &value
	case "background":
		c, err := NormalizeColor(value)
		if err != nil {
			return err
		}
		s.Background = &c
	case "text color":
		c, err := NormalizeColor(value)
		if err != nil {
			return err
		}
		s.TextColor = &c
	case "speaker name color":
		c, err := NormalizeColor(value)
		if err != nil {
			return err
		}
		s.SpeakerColor = &c
	case "font":
		f, err := NormalizeFont(value)
		if err != nil {
			return err
		}
		s.Font = &f
	case "text size":
		n, err := NormalizeTextSize(value)
		if err != nil {
			return err
		}
		s.TextSize = &n
	case "speaker names":
		f, err := NormalizeSpeakerNames(value)
		if err != nil {
			return err
		}
		s.SpeakerNames = &f
	case "idle message":
		if len(value) > MaxIdleMessage {
			return fmt.Errorf("the idle message is too long (%d characters; the limit is %d)", len(value), MaxIdleMessage)
		}
		s.IdleMessage = &value
	case "loop seconds":
		f, err := NormalizeLoopSeconds(value)
		if err != nil {
			return err
		}
		s.LoopSeconds = &f
	case "scrollback lines":
		n, err := NormalizeScrollback(value)
		if err != nil {
			return err
		}
		s.ScrollbackLines = &n
	case "archived":
		switch strings.ToLower(value) {
		case "yes", "true", "on", "1":
			t := true
			s.Archived = &t
		case "no", "false", "off", "0", "":
			s.Archived = nil
		default:
			return fmt.Errorf("%q is not yes or no", value)
		}
	}
	return nil
}

// ---- parsing a [settings] block (called from the main parser) ----

// settingsBlock reads lines after "[settings]" up to "[/settings]". It never fails: each problem is a
// diagnostic and the other lines still count.
func (p *parser) settingsBlock(start int) {
	seen := map[string]int{}
	for p.i++; p.i < len(p.lines); p.i++ {
		line := strings.TrimSpace(p.lines[p.i])
		lineNo := p.i + 1
		if line == "" {
			continue
		}
		if isDirective(line) {
			if n, _ := parseDirective(line); n == "/settings" {
				return
			}
			p.diag(lineNo, SevWarning, "Only settings (name: value) can go inside [settings]. This line was skipped.")
			continue
		}
		key, val, ok := strings.Cut(line, ":")
		if !ok {
			p.diag(lineNo, SevWarning, "Expected \"name: value\" here, for example \"font: dyslexic\". This line was skipped.")
			continue
		}
		def, known := lookupSetting(key)
		if !known {
			p.res.Settings.Extra = append(p.res.Settings.Extra, line)
			p.diag(lineNo, SevWarning, "Unknown setting %q.%s It was ignored.", strings.TrimSpace(key), didYouMean(key))
			continue
		}
		if p.opts.Translation && !def.translated {
			if def.name != "title" {
				p.diag(lineNo, SevWarning, "%q is set once for the whole show, in the main show file. This line was ignored.", def.name)
			}
			continue
		}
		if prev, dup := seen[def.name]; dup {
			p.diag(lineNo, SevInfo, "%q was already set on line %d; this later value is used.", def.name, prev)
		}
		seen[def.name] = lineNo
		if err := p.res.Settings.set(def.name, val); err != nil {
			p.diag(lineNo, SevWarning, "%s: %v. The default is used instead.", def.name, err)
		}
	}
	p.diag(start, SevError, "[settings] was opened but never closed with [/settings]. The settings so far were kept.")
}

func didYouMean(key string) string {
	k := normKey(key)
	best, bestD := "", 4
	for _, s := range settingTable {
		for _, cand := range append([]string{s.name}, s.aliases...) {
			if d := levenshtein(k, normKey(cand)); d < bestD {
				best, bestD = s.name, d
			}
		}
	}
	if best == "" {
		return " Known settings: " + strings.Join(SettingNames(), ", ") + "."
	}
	return fmt.Sprintf(" Did you mean %q?", best)
}

// ---- writing settings back into a show file ----

// blockSpans finds the [settings]...[/settings] blocks of a show file (ignoring any that sit inside a
// <!-- comment -->), as inclusive line index ranges.
func blockSpans(lines []string) [][2]int {
	var spans [][2]int
	inComment := false
	open := -1
	for i, ln := range lines {
		startsInComment := inComment
		// update the comment state across this line
		rest := ln
		for {
			if inComment {
				j := strings.Index(rest, "-->")
				if j < 0 {
					break
				}
				inComment, rest = false, rest[j+3:]
			} else {
				j := strings.Index(rest, "<!--")
				if j < 0 {
					break
				}
				inComment, rest = true, rest[j+4:]
			}
		}
		if startsInComment {
			continue
		}
		t := strings.ToLower(strings.TrimSpace(ln))
		switch {
		case open < 0 && t == "[settings]":
			open = i
		case open >= 0 && t == "[/settings]":
			spans = append(spans, [2]int{open, i})
			open = -1
		}
	}
	if open >= 0 { // unclosed: treat the rest of the file as the block so it is not duplicated
		spans = append(spans, [2]int{open, len(lines) - 1})
	}
	return spans
}

// insertionPoint is the line index where a new settings block goes: after leading blank lines,
// comments and the [language: ...] line.
func insertionPoint(lines []string) int {
	inComment := false
	for i, ln := range lines {
		t := strings.TrimSpace(ln)
		before := inComment
		rest := ln
		for {
			if inComment {
				j := strings.Index(rest, "-->")
				if j < 0 {
					break
				}
				inComment, rest = false, rest[j+3:]
			} else {
				j := strings.Index(rest, "<!--")
				if j < 0 {
					break
				}
				inComment, rest = true, rest[j+4:]
			}
		}
		switch {
		case before || inComment || t == "" || strings.HasPrefix(t, "<!--"):
			continue
		case strings.HasPrefix(strings.ToLower(t), "[language"):
			continue
		}
		return i
	}
	return len(lines)
}

// render writes the block for s: the known settings in a fixed order, then any lines this program
// did not understand.
func (s Settings) render() []string {
	out := []string{"[settings]"}
	add := func(name, v string) { out = append(out, name+": "+v) }
	if s.Title != nil {
		add("title", *s.Title)
	}
	if s.Background != nil {
		add("background", *s.Background)
	}
	if s.TextColor != nil {
		add("text color", *s.TextColor)
	}
	if s.SpeakerColor != nil {
		add("speaker name color", *s.SpeakerColor)
	}
	if s.Font != nil {
		add("font", *s.Font)
	}
	if s.TextSize != nil {
		add("text size", strconv.Itoa(*s.TextSize)+"%")
	}
	if s.SpeakerNames != nil {
		add("speaker names", *s.SpeakerNames)
	}
	if s.IdleMessage != nil {
		add("idle message", *s.IdleMessage)
	}
	if s.LoopSeconds != nil {
		add("loop seconds", strconv.FormatFloat(*s.LoopSeconds, 'f', -1, 64))
	}
	if s.ScrollbackLines != nil {
		add("scrollback lines", strconv.Itoa(*s.ScrollbackLines))
	}
	if s.Archived != nil && *s.Archived {
		add("archived", "yes")
	}
	out = append(out, s.Extra...)
	return append(out, "[/settings]")
}

// ReadSettings returns the settings a show file sets (translation: whether it is a translation).
func ReadSettings(text string, translation bool) Settings {
	return Parse(text, Options{Translation: translation}).Settings
}

// UpdateSettings rewrites the settings of a show file: it reads the current settings, lets change
// edit them, and writes one [settings] block near the top, replacing any existing blocks. Everything
// else in the file (including comments outside the block) is left exactly as it was.
func UpdateSettings(text string, translation bool, change func(*Settings)) string {
	cur := ReadSettings(text, translation)
	change(&cur)
	return ReplaceSettings(text, cur)
}

// ReplaceSettings writes s as the file's single [settings] block (or removes the block if s is empty).
func ReplaceSettings(text string, s Settings) string {
	text = strings.ReplaceAll(strings.ReplaceAll(text, "\r\n", "\n"), "\r", "\n")
	lines := strings.Split(text, "\n")
	// remove existing blocks, last first so indexes stay valid
	spans := blockSpans(lines)
	for i := len(spans) - 1; i >= 0; i-- {
		a, b := spans[i][0], spans[i][1]
		if b+1 < len(lines) && strings.TrimSpace(lines[b+1]) == "" {
			b++ // also drop the blank line that followed the block
		}
		lines = append(lines[:a:a], lines[b+1:]...)
	}
	empty := s.Title == nil && !s.LookSet() && s.IdleMessage == nil && s.LoopSeconds == nil && s.ScrollbackLines == nil && (s.Archived == nil || !*s.Archived) && len(s.Extra) == 0
	if empty {
		return strings.Join(lines, "\n")
	}
	if strings.TrimSpace(strings.Join(lines, "\n")) == "" {
		lines = nil // nothing but blank lines: start clean
	}
	at := insertionPoint(lines)
	block := append(s.render(), "")
	// keep one blank line between what precedes and the block
	if at > 0 && strings.TrimSpace(lines[at-1]) != "" {
		block = append([]string{""}, block...)
	}
	out := append(append(append([]string{}, lines[:at]...), block...), lines[at:]...)
	return strings.Join(out, "\n")
}
