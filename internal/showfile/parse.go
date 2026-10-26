// Copyright the open-caption authors
// SPDX-License-Identifier: Apache-2.0

package showfile

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"
)

var (
	speakerRe = regexp.MustCompile(`^([A-Z][A-Z0-9 .'\-&,/()#]{0,59}):\s+(\S.*)$`)
	// a speaker's name and colon with nothing after them: probably the words were meant to follow
	// a label with lower-case letters, like "Barry: Hi": not a speaker, unless the name is a speaker elsewhere
	mixedLabelRe  = regexp.MustCompile(`^([A-Za-z][A-Za-z0-9 .'\-&,/()#]{0,59}):\s+\S`)
	speakerOnlyRe = regexp.MustCompile(`^([A-Z][A-Z0-9 .'\-&,/()#]{0,59}):\s*$`)
	durationRe    = regexp.MustCompile(`^(\d+(?:\.\d+)?)\s*s$`)
	// "timer 15", "timer: 7:30pm", "countdown 10 minutes" (the keyword, then a space or a colon)
	timerOptionRe = regexp.MustCompile(`(?i)^(?:timer|countdown)(?:\s+|\s*:\s*|$)`)
	// "name Intermission", "name: Act 1 break" (the keyword, then a space or a colon)
	nameOptionRe = regexp.MustCompile(`(?i)^name(?:\s+|\s*:\s*|$)`)
	imageMdRe    = regexp.MustCompile(`^!\[(.*?)\]\((.+?)\)$`)
	htmlTagRe    = regexp.MustCompile(`<[A-Za-z/!][^>]*>`)
	spacesRe     = regexp.MustCompile(`\s+`)
)

var directiveAlias = map[string]string{
	"clear screen":  "clear",
	"blank screen":  "blank",
	"full screen":   "fullscreen",
	"/full screen":  "/fullscreen",
	"end loop":      "/loop",
	"duet":          "together",
	"/duet":         "/together",
	"counterpoint":  "together",
	"/counterpoint": "/together",
}

type parser struct {
	lines []string
	i     int // index of the line being processed
	res   *Result
	opts  Options

	song     bool // inside [song] ... [/song]: every caption is a sung lyric
	songLine int

	loopTimer *Timer // the timer of the [loop] being read

	together *togetherBlock // inside [together] ... [/together]: lines are voices shown at once

	speakers map[string]bool // the speaker labels found
	notCaps  []labelAt       // lines that start like "Barry: text", checked against speakers when the file has been read
}

type labelAt struct {
	line  int
	label string
}

type togetherBlock struct {
	line  int
	parts []Part
}

// Parse parses a show file. It never fails: problems become Diagnostics and the
// affected construct is skipped.
func Parse(src string, opts Options) Result {
	src = strings.TrimPrefix(src, "\uFEFF")
	src = strings.ReplaceAll(src, "\r\n", "\n")
	src = strings.ReplaceAll(src, "\r", "\n")

	res := Result{Cues: []Cue{}, Diagnostics: []Diagnostic{}}
	src = stripComments(src, &res)
	p := &parser{lines: strings.Split(src, "\n"), res: &res, opts: opts}
	p.run()
	for _, c := range p.notCaps {
		if up := strings.ToUpper(c.label); p.speakers[up] {
			p.diag(c.line, SevWarning, "%q is not in capitals, so it is shown as text, not as the speaker %s. Write the name in capitals.", c.label+":", up+":")
		}
	}
	if p.song {
		p.diag(p.songLine, SevError, "[song] was opened but never closed with [/song]. The lyrics so far were kept.")
	}
	if p.together != nil {
		p.diag(p.together.line, SevError, "[together] was opened but never closed with [/together]. The lines so far were kept.")
		p.closeTogether()
	}

	for i := range res.Cues {
		res.Cues[i].Index = i
	}
	p.summary()
	res.Speakers = SpeakersOf(res.Cues)
	return res
}

func (p *parser) diag(line int, sev, format string, a ...any) {
	p.res.Diagnostics = append(p.res.Diagnostics, Diagnostic{Line: line, Severity: sev, Message: fmt.Sprintf(format, a...)})
}

// stripComments removes <!-- ... --> (possibly multi-line, possibly inline),
// preserving the line count so later diagnostics keep correct line numbers.
func stripComments(src string, res *Result) string {
	var out strings.Builder
	for {
		start := strings.Index(src, "<!--")
		if start < 0 {
			out.WriteString(src)
			return out.String()
		}
		out.WriteString(src[:start])
		line := 1 + strings.Count(out.String(), "\n")
		rest := src[start+4:]
		end := strings.Index(rest, "-->")
		if end < 0 {
			res.Diagnostics = append(res.Diagnostics, Diagnostic{Line: line, Severity: SevError,
				Message: "A comment <!-- was opened but never closed. Everything after it was ignored."})
			out.WriteString(strings.Repeat("\n", strings.Count(rest, "\n")))
			return out.String()
		}
		out.WriteString(strings.Repeat("\n", strings.Count(rest[:end], "\n")))
		src = rest[end+3:]
	}
}

func (p *parser) run() {
	for p.i = 0; p.i < len(p.lines); p.i++ {
		raw := strings.TrimSpace(p.lines[p.i])
		lineNo := p.i + 1
		switch {
		case raw == "":
		case strings.HasPrefix(raw, "#"):
			p.diag(lineNo, SevInfo, "Lines starting with # are reserved for headings and were ignored. Use <!-- comment --> for comments or [note: …] for operator notes.")
		case raw == "[[":
			p.multiLine(lineNo)
		case imageMdRe.MatchString(raw):
			m := imageMdRe.FindStringSubmatch(raw)
			p.addImage(lineNo, m[2], m[1], 0)
		case isDirective(raw):
			p.directive(raw, lineNo)
		default:
			p.caption(p.continuation(raw), lineNo)
		}
	}
}

func isDirective(s string) bool {
	return len(s) >= 3 && s[0] == '[' && s[len(s)-1] == ']' && s[1] != '['
}

func parseDirective(s string) (name, args string) {
	inner := s[1 : len(s)-1]
	if k := strings.Index(inner, ":"); k >= 0 {
		name, args = inner[:k], strings.TrimSpace(inner[k+1:])
	} else {
		name = inner
	}
	name = strings.ToLower(spacesRe.ReplaceAllString(strings.TrimSpace(name), " "))
	if a, ok := directiveAlias[name]; ok {
		name = a
	}
	return
}

// continuation joins lines ending in a backslash (Markdown hard break).
func (p *parser) continuation(first string) []string {
	lines := []string{first}
	for endsWithContinuation(lines[len(lines)-1]) && p.i+1 < len(p.lines) {
		last := lines[len(lines)-1]
		lines[len(lines)-1] = strings.TrimSpace(last[:len(last)-1])
		p.i++
		next := strings.TrimSpace(p.lines[p.i])
		if next == "" {
			return lines
		}
		lines = append(lines, next)
	}
	if n := len(lines); endsWithContinuation(lines[n-1]) {
		lines[n-1] = strings.TrimSpace(lines[n-1][:len(lines[n-1])-1])
	}
	return lines
}

func endsWithContinuation(s string) bool {
	n := 0
	for i := len(s) - 1; i >= 0 && s[i] == '\\'; i-- {
		n++
	}
	return n%2 == 1
}

func (p *parser) multiLine(start int) {
	var lines []string
	for p.i++; p.i < len(p.lines); p.i++ {
		t := strings.TrimSpace(p.lines[p.i])
		if t == "]]" {
			if len(lines) == 0 {
				p.diag(start, SevWarning, "This [[ ]] block is empty and was skipped.")
				return
			}
			p.caption(lines, start)
			return
		}
		if t != "" {
			lines = append(lines, t)
		}
	}
	p.diag(start, SevError, "[[ was opened but never closed with ]]. The block was skipped.")
}

// singMarks reports whether a line is written as a lyric with the ~ shorthand
// ("~ lyric ~" or "~ lyric") and returns it without the markers.
func singMarks(l string) (string, bool) {
	t := strings.TrimSpace(l)
	switch {
	case t == "~":
		return "", true
	case strings.HasPrefix(t, "~ "):
		t = strings.TrimSpace(t[2:])
		if t == "~" {
			return "", true
		}
		return strings.TrimSpace(strings.TrimSuffix(t, " ~")), true
	}
	return l, false
}

// repeatedName is the first name that a group label such as "BARRY, DEE DEE & BARRY" gives twice, or "".
func repeatedName(label string) string {
	seen := map[string]bool{}
	for _, n := range strings.FieldsFunc(label, func(r rune) bool { return r == ',' || r == '&' || r == '/' }) {
		if n = strings.TrimSpace(n); n != "" {
			if seen[n] {
				return n
			}
			seen[n] = true
		}
	}
	return ""
}

// part renders one voice: speaker label, ~ lyric markers, Markdown.
func (p *parser) part(lines []string, lineNo int) Part {
	var pt Part
	lines = append([]string(nil), lines...)
	if m := speakerRe.FindStringSubmatch(lines[0]); m != nil {
		pt.Speaker = m[1]
		lines[0] = m[2]
		if p.speakers == nil {
			p.speakers = map[string]bool{}
		}
		p.speakers[m[1]] = true
		if dup := repeatedName(m[1]); dup != "" {
			p.diag(lineNo, SevWarning, "The speaker label %q names %s twice. Is one of them a different character?", m[1]+":", dup)
		}
	} else if m := mixedLabelRe.FindStringSubmatch(lines[0]); m != nil && m[1] != strings.ToUpper(m[1]) {
		p.notCaps = append(p.notCaps, labelAt{lineNo, m[1]})
	} else if m := speakerOnlyRe.FindStringSubmatch(lines[0]); m != nil {
		p.diag(lineNo, SevWarning, "%q is a speaker's name with nothing after the colon. Put the words on the same line, or write \\: if the colon is part of the text.", m[1]+":")
	}
	_, shorthand := singMarks(lines[0])
	pt.Sung = p.song || shorthand
	var hs, ps []string
	for _, l := range lines {
		if htmlTagRe.MatchString(l) {
			p.diag(lineNo, SevInfo, "HTML is not supported; %q will appear as plain text. Use **bold** or *italic* instead.", htmlTagRe.FindString(l))
		}
		if pt.Sung {
			l, _ = singMarks(l)
			l = strings.TrimSpace(strings.TrimSuffix(strings.TrimSpace(l), " ~")) // a closing marker on a later line
		}
		h, t := RenderInline(l)
		if pt.Sung {
			// Lyrics sit between musical notes; an empty lyric is just a note.
			if t == "" {
				h, t = "♪", "♪"
			} else {
				h, t = "♪ "+h+" ♪", "♪ "+t+" ♪"
			}
		}
		hs, ps = append(hs, h), append(ps, t)
	}
	pt.HTML = strings.Join(hs, "<br>")
	pt.Text = strings.Join(ps, "\n")
	if pt.Speaker == "" && !pt.Sung && strings.HasPrefix(pt.Text, "(") && strings.HasSuffix(pt.Text, ")") {
		pt.Stage = true
	}
	return pt
}

// voiceSep separates voices that are shown at the same time on one line.
const voiceSep = " || "

func (p *parser) caption(lines []string, lineNo int) {
	if p.together != nil { // inside [together]: every line is one more voice
		p.together.parts = append(p.together.parts, p.part(lines, lineNo))
		return
	}
	if len(lines) == 1 && strings.Contains(lines[0], voiceSep) {
		var parts []Part
		for _, seg := range strings.Split(lines[0], voiceSep) {
			if seg = strings.TrimSpace(seg); seg != "" {
				parts = append(parts, p.part([]string{seg}, lineNo))
			}
		}
		p.addVoices(parts, lineNo)
		return
	}
	p.addVoices([]Part{p.part(lines, lineNo)}, lineNo)
}

// LongCaption is the length, in characters, above which a caption line is reported as long: on a phone it wraps over
// several lines and is slow to read.
const LongCaption = 100

// checkLength reports the longest line of a caption when it is over LongCaption characters.
func (p *parser) checkLength(lineNo int, text string) {
	longest := 0
	for _, l := range strings.Split(text, "\n") {
		longest = max(longest, utf8.RuneCountInString(l))
	}
	if longest > LongCaption {
		p.diag(lineNo, SevInfo, "This caption line is %d characters long. On a phone it wraps over several lines; consider splitting it into shorter lines.", longest)
	}
}

// ManyVoices is the number of voices shown together from which a caption is reported: a phone stacks them all, and a
// shared screen has room for about four.
const ManyVoices = 5

// addVoices appends a caption made of one voice, or of several shown together.
func (p *parser) addVoices(parts []Part, lineNo int) {
	switch len(parts) {
	case 0:
		return
	case 1:
		pt := parts[0]
		p.checkLength(lineNo, pt.Text)
		p.res.Cues = append(p.res.Cues, Cue{Kind: KindCaption, Line: lineNo, Speaker: pt.Speaker, HTML: pt.HTML, Text: pt.Text, Stage: pt.Stage, Sung: pt.Sung})
		return
	}
	if len(parts) >= ManyVoices {
		p.diag(lineNo, SevWarning, "This caption shows %d voices at once. A phone stacks them and a shared screen may not have room for them all; consider splitting it into two captions.", len(parts))
	}
	c := Cue{Kind: KindCaption, Line: lineNo, Parts: parts, Sung: true}
	for _, pt := range parts {
		p.checkLength(lineNo, pt.Text)
	}
	var hs, ts []string
	for _, pt := range parts {
		c.Sung = c.Sung && pt.Sung
		hs = append(hs, pt.HTML)
		if pt.Speaker != "" {
			ts = append(ts, pt.Speaker+": "+pt.Text)
		} else {
			ts = append(ts, pt.Text)
		}
	}
	c.HTML, c.Text = strings.Join(hs, "<br>"), strings.Join(ts, "\n")
	p.res.Cues = append(p.res.Cues, c)
}

func (p *parser) closeTogether() {
	t := p.together
	p.together = nil
	switch len(t.parts) {
	case 0:
		p.diag(t.line, SevWarning, "This [together] block is empty and was skipped.")
	case 1:
		p.diag(t.line, SevWarning, "[together] needs at least two lines to show at the same time; this one was shown as an ordinary caption.")
		p.addVoices(t.parts, t.line)
	default:
		p.addVoices(t.parts, t.line)
	}
}

func (p *parser) directive(raw string, lineNo int) {
	name, args := parseDirective(raw)
	if p.together != nil && name != "/together" && name != "together" {
		p.diag(lineNo, SevWarning, "[%s] can't go inside [together]; only lines of text can. This line was skipped.", name)
		return
	}
	switch name {
	case "clear", "blank":
		if args != "" {
			p.diag(lineNo, SevWarning, "[%s] doesn't take arguments; %q was ignored.", name, args)
		}
		p.res.Cues = append(p.res.Cues, Cue{Kind: CueKind(name), Line: lineNo})
	case "image":
		fields, secs := p.splitDuration(lineNo, args, false)
		if len(fields) == 0 || fields[0] == "" {
			p.diag(lineNo, SevWarning, "[image] needs a file name, e.g. [image: curtain.jpg]. The line was skipped.")
			return
		}
		p.addImage(lineNo, fields[0], strings.Join(fields[1:], " | "), secs)
	case "fullscreen":
		if args == "" {
			p.fullscreenBlock(lineNo)
			return
		}
		fields, _ := p.splitDuration(lineNo, args, false)
		h, t := RenderInline(strings.Join(fields, " | "))
		p.res.Cues = append(p.res.Cues, Cue{Kind: KindFullscreen, Line: lineNo, Text: t, HTML: h,
			Item: &Item{Kind: "text", HTML: h, Text: t}})
	case "loop":
		p.loop(lineNo, args)
	case "song":
		if p.song {
			p.diag(lineNo, SevWarning, "[song] can't go inside another song. Close the first one with [/song].")
			return
		}
		p.song, p.songLine = true, lineNo
		if args != "" { // the title is a bookmark for the operator
			h, t := RenderInline(args)
			p.res.Cues = append(p.res.Cues, Cue{Kind: KindSection, Line: lineNo, Text: "♪ " + t, HTML: "♪ " + h})
		}
	case "together":
		if p.together != nil {
			p.diag(lineNo, SevWarning, "[together] can't go inside another [together]. Close the first one with [/together].")
			return
		}
		p.together = &togetherBlock{line: lineNo}
	case "/together":
		if p.together == nil {
			p.diag(lineNo, SevWarning, "[/together] has no matching [together]. The line was skipped.")
			return
		}
		p.closeTogether()
	case "/song":
		if !p.song {
			p.diag(lineNo, SevWarning, "[/song] has no matching [song]. The line was skipped.")
			return
		}
		p.song = false
	case "settings":
		p.settingsBlock(lineNo)
	case "/settings":
		p.diag(lineNo, SevWarning, "[/settings] has no matching [settings]. The line was skipped.")
	case "section", "note":
		if args == "" {
			p.diag(lineNo, SevWarning, "[%s] needs some text, e.g. [%s: Act 2]. The line was skipped.", name, name)
			return
		}
		h, t := RenderInline(args)
		p.res.Cues = append(p.res.Cues, Cue{Kind: CueKind(name), Line: lineNo, Text: t, HTML: h})
	case "language":
		f := splitFields(args)
		if len(f) == 0 || f[0] == "" {
			p.diag(lineNo, SevWarning, "[language] needs a code, e.g. [language: es]. The line was skipped.")
			return
		}
		p.res.Language = strings.ToLower(f[0])
		if len(f) > 1 {
			p.res.LanguageName = f[1]
		}
	case "/loop", "/fullscreen":
		p.diag(lineNo, SevWarning, "[%s] has no matching opening command. The line was skipped.", name)
	default:
		p.diag(lineNo, SevWarning, "Unknown command [%s]. This line was skipped.", name)
	}
}

func (p *parser) fullscreenBlock(start int) {
	var hs, ps []string
	for p.i++; p.i < len(p.lines); p.i++ {
		t := strings.TrimSpace(p.lines[p.i])
		if isDirective(t) {
			if n, _ := parseDirective(t); n == "/fullscreen" {
				if len(ps) == 0 {
					p.diag(start, SevWarning, "This [fullscreen] block is empty and was skipped.")
					return
				}
				h, pl := strings.Join(hs, "<br>"), strings.Join(ps, "\n")
				p.res.Cues = append(p.res.Cues, Cue{Kind: KindFullscreen, Line: start, Text: pl, HTML: h,
					Item: &Item{Kind: "text", HTML: h, Text: pl}})
				return
			}
		}
		if t != "" {
			h, pl := RenderInline(t)
			hs, ps = append(hs, h), append(ps, pl)
		}
	}
	p.diag(start, SevError, "[fullscreen] was opened but never closed with [/fullscreen]. The block was skipped.")
}

func (p *parser) loop(start int, args string) {
	cue := Cue{Kind: KindLoop, Line: start}
	// Options, separated by commas: an item time (10s), a name (name Intermission) and/or a timer (timer 15, timer 7:30pm).
	for _, part := range strings.Split(args, ",") {
		part = strings.TrimSpace(part)
		low := strings.ToLower(part)
		switch {
		case part == "":
		case durationRe.MatchString(part):
			cue.Period = p.seconds(start, durationRe.FindStringSubmatch(part)[1])
		case nameOptionRe.MatchString(low):
			name := strings.Join(strings.Fields(nameOptionRe.ReplaceAllString(part, "")), " ")
			switch r := []rune(name); {
			case name == "":
				p.diag(start, SevWarning, "The loop name is empty. Write it after the word name, like [loop: name Intermission]. The loop has no name.")
			case len(r) > MaxLoopNameLength:
				p.diag(start, SevWarning, "The loop name is longer than %d characters and was cut.", MaxLoopNameLength)
				cue.Name = strings.TrimSpace(string(r[:MaxLoopNameLength]))
			default:
				cue.Name = name
			}
		case timerOptionRe.MatchString(low):
			val := strings.TrimSpace(timerOptionRe.ReplaceAllString(part, ""))
			t, err := ParseTimer(val)
			switch {
			case errors.Is(err, ErrTimerRange):
				p.diag(start, SevWarning, "The timer length must be more than 0 and at most %d minutes. The loop has no timer.", MaxTimerMinutes)
			case err != nil:
				p.diag(start, SevWarning, "Timer %q is not a number of minutes (like 15) or a time of day (like 7:30pm). The loop has no timer.", val)
			default:
				cue.Timer = t
			}
		default:
			p.diag(start, SevWarning, "[loop: %s] — expected a time like 10s. The default time was used.", part)
		}
	}
	p.loopTimer = cue.Timer
	defer func() { p.loopTimer = nil }()
	closed := false
	for p.i++; p.i < len(p.lines); p.i++ {
		t := strings.TrimSpace(p.lines[p.i])
		lineNo := p.i + 1
		if t == "" {
			continue
		}
		if isDirective(t) {
			if n, _ := parseDirective(t); n == "/loop" {
				closed = true
				break
			}
		}
		if it, ok := p.loopItem(t, lineNo); ok {
			cue.Items = append(cue.Items, it)
		}
	}
	switch {
	case !closed:
		p.diag(start, SevError, "[loop] was opened but never closed with [/loop]. The loop was skipped.")
	case len(cue.Items) == 0:
		p.diag(start, SevWarning, "This loop has no items and was skipped.")
	default:
		if cue.Timer != nil {
			shown := false
			for _, it := range cue.Items {
				shown = shown || it.Timer
			}
			if !shown {
				p.diag(start, SevWarning, "This loop has a timer, but no text in it contains {timer}, so the countdown is never shown.")
			}
		}
		p.res.Cues = append(p.res.Cues, cue)
	}
}

// timerItem turns {timer} in a loop's text item into the placeholder the page fills with the time left.
func (p *parser) timerItem(it Item, lineNo int) Item {
	if !usesTimer(it.Text) {
		return it
	}
	if p.loopTimer == nil {
		p.diag(lineNo, SevWarning, "{timer} needs a timer on its loop, for example [loop: timer 15]. It is shown as typed.")
		return it
	}
	it.HTML = strings.ReplaceAll(it.HTML, TimerPlaceholder, timerSpanHTML)
	it.Timer = true
	return it
}

func (p *parser) loopItem(t string, lineNo int) (Item, bool) {
	switch {
	case strings.HasPrefix(t, "#"):
		p.diag(lineNo, SevInfo, "Lines starting with # are reserved for headings and were ignored.")
		return Item{}, false
	case imageMdRe.MatchString(t):
		m := imageMdRe.FindStringSubmatch(t)
		return p.imageItem(lineNo, m[2], m[1], 0)
	case t == "[[":
		p.diag(lineNo, SevWarning, "[[ ]] blocks can't go inside a loop. This line was skipped.")
		return Item{}, false
	case isDirective(t):
		name, args := parseDirective(t)
		switch name {
		case "image":
			fields, secs := p.splitDuration(lineNo, args, false)
			if len(fields) == 0 || fields[0] == "" {
				p.diag(lineNo, SevWarning, "[image] needs a file name. The line was skipped.")
				return Item{}, false
			}
			return p.imageItem(lineNo, fields[0], strings.Join(fields[1:], " | "), secs)
		case "fullscreen":
			fields, secs := p.splitDuration(lineNo, args, false)
			if len(fields) == 0 || fields[0] == "" {
				p.diag(lineNo, SevWarning, "Inside a loop, [fullscreen] needs text on the same line. The line was skipped.")
				return Item{}, false
			}
			h, pl := RenderInline(strings.Join(fields, " | "))
			return p.timerItem(Item{Kind: "text", HTML: h, Text: pl, Seconds: secs}, lineNo), true
		case "blank":
			_, secs := p.splitDuration(lineNo, args, true)
			return Item{Kind: "blank", Seconds: secs}, true
		case "loop":
			p.diag(lineNo, SevWarning, "Loops can't be nested. This line was skipped.")
		default:
			p.diag(lineNo, SevWarning, "[%s] can't go inside a loop. This line was skipped.", name)
		}
		return Item{}, false
	}
	fields, secs := p.splitDuration(lineNo, t, false)
	h, pl := RenderInline(strings.Join(fields, " | "))
	return p.timerItem(Item{Kind: "text", HTML: h, Text: pl, Seconds: secs}, lineNo), true
}

func (p *parser) addImage(lineNo int, file, caption string, secs float64) {
	it, ok := p.imageItem(lineNo, file, caption, secs)
	if !ok {
		return
	}
	cp := it
	p.res.Cues = append(p.res.Cues, Cue{Kind: KindImage, Line: lineNo, Text: it.Image, Item: &cp})
}

func (p *parser) imageItem(lineNo int, file, caption string, secs float64) (Item, bool) {
	file = strings.TrimSpace(file)
	if p.opts.Assets != nil && !contains(p.opts.Assets, file) {
		if s := suggest(file, p.opts.Assets); s != "" {
			p.diag(lineNo, SevWarning, "Image %q not found. Did you mean %q?", file, s)
		} else {
			if p.opts.ImageFolder != "" { // not Setup: the command line checks a folder
				p.diag(lineNo, SevWarning, "Image %q not found in %s.", file, p.opts.ImageFolder)
			} else {
				p.diag(lineNo, SevWarning, "Image %q not found. Upload it in the Images step.", file)
			}
		}
	}
	return Item{Kind: "image", Image: file, Caption: strings.TrimSpace(caption), Seconds: secs}, true
}

func (p *parser) summary() {
	counts := map[CueKind]int{}
	for _, c := range p.res.Cues {
		counts[c.Kind]++
	}
	if len(p.res.Cues) == 0 {
		return
	}
	p.diag(0, SevInfo, "%d captions, %d images, %d fullscreen, %d loops, %d sections, %d notes.",
		counts[KindCaption], counts[KindImage], counts[KindFullscreen], counts[KindLoop], counts[KindSection], counts[KindNote])
}

func splitFields(args string) []string {
	var out []string
	for _, f := range strings.Split(args, "|") {
		out = append(out, strings.TrimSpace(f))
	}
	return out
}

// MaxItemSeconds is the longest time one loop item or a loop's period may be given (as for the loop time setting);
// a longer time is cut to it with a warning.
const MaxItemSeconds = 3600

// seconds reads a time given in seconds, cutting it to MaxItemSeconds (a huge number would overflow a duration).
func (p *parser) seconds(lineNo int, digits string) float64 {
	secs, _ := strconv.ParseFloat(digits, 64)
	if secs > MaxItemSeconds {
		p.diag(lineNo, SevWarning, "A time of %ss was cut to the most allowed, %d seconds.", digits, MaxItemSeconds)
		return MaxItemSeconds
	}
	return secs
}

// splitDuration splits "a | b | 15s" into fields and a trailing duration.
// A lone duration counts as one only when allowSole is set (e.g. [blank: 3s]).
func (p *parser) splitDuration(lineNo int, args string, allowSole bool) ([]string, float64) {
	f := splitFields(args)
	if len(f) == 0 {
		return nil, 0
	}
	last := f[len(f)-1]
	if m := durationRe.FindStringSubmatch(last); m != nil && (len(f) > 1 || allowSole) {
		return f[:len(f)-1], p.seconds(lineNo, m[1])
	}
	return f, 0
}

func contains(list []string, s string) bool {
	for _, x := range list {
		if x == s {
			return true
		}
	}
	return false
}

// suggest returns the closest asset name (case-insensitive match, or a small edit distance).
func suggest(name string, assets []string) string {
	best, bestD := "", 4
	for _, a := range assets {
		d := levenshtein(strings.ToLower(name), strings.ToLower(a))
		if d < bestD {
			best, bestD = a, d
		}
	}
	return best
}

func levenshtein(a, b string) int {
	ra, rb := []rune(a), []rune(b)
	prev := make([]int, len(rb)+1)
	for j := range prev {
		prev[j] = j
	}
	for i := 1; i <= len(ra); i++ {
		cur := make([]int, len(rb)+1)
		cur[0] = i
		for j := 1; j <= len(rb); j++ {
			cost := 1
			if ra[i-1] == rb[j-1] {
				cost = 0
			}
			cur[j] = min(prev[j]+1, cur[j-1]+1, prev[j-1]+cost)
		}
		prev = cur
	}
	return prev[len(rb)]
}
