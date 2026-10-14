package script

import (
	"html"
	"strings"
	"unicode"
)

// RenderInline converts the supported Markdown subset (**bold**, *italic*,
// _italic_, backslash escapes) to escaped HTML and a plain-text rendition.
// All input text is HTML-escaped; only <strong> and <em> are generated.
func RenderInline(s string) (htmlOut, plain string) {
	var h, p strings.Builder
	renderInto(&h, &p, []rune(s))
	return h.String(), p.String()
}

// Escapable lists the characters a backslash can escape in script text.
const Escapable = `\*_#[]:()!`

const escapable = Escapable

func renderInto(h, p *strings.Builder, r []rune) {
	for i := 0; i < len(r); {
		c := r[i]
		switch {
		case c == '\\' && i+1 < len(r) && strings.ContainsRune(escapable, r[i+1]):
			writeLit(h, p, r[i+1])
			i += 2
		case c == '*' && i+1 < len(r) && r[i+1] == '*':
			if end := findClose(r, i+2, "**"); end > i+2 {
				h.WriteString("<strong>")
				renderInto(h, p, r[i+2:end])
				h.WriteString("</strong>")
				i = end + 2
				continue
			}
			writeLit(h, p, c)
			i++
		case c == '*' || c == '_':
			if canOpen(r, i) {
				if end := findClose(r, i+1, string(c)); end > i+1 && canClose(r, end) {
					h.WriteString("<em>")
					renderInto(h, p, r[i+1:end])
					h.WriteString("</em>")
					i = end + 1
					continue
				}
			}
			writeLit(h, p, c)
			i++
		default:
			writeLit(h, p, c)
			i++
		}
	}
}

func writeLit(h, p *strings.Builder, c rune) {
	h.WriteString(html.EscapeString(string(c)))
	p.WriteRune(c)
}

// canOpen: an emphasis marker opens only before a non-space and, for '_',
// not in the middle of a word (snake_case stays literal).
func canOpen(r []rune, i int) bool {
	if i+1 >= len(r) || unicode.IsSpace(r[i+1]) {
		return false
	}
	if r[i] == '_' && i > 0 && isWord(r[i-1]) {
		return false
	}
	return true
}

func canClose(r []rune, i int) bool {
	if i == 0 || unicode.IsSpace(r[i-1]) {
		return false
	}
	if r[i] == '_' && i+1 < len(r) && isWord(r[i+1]) {
		return false
	}
	return true
}

func isWord(c rune) bool { return unicode.IsLetter(c) || unicode.IsDigit(c) }

// findClose returns the index of the next unescaped occurrence of marker at or
// after from, or -1. A single '*' marker does not match half of a '**'.
func findClose(r []rune, from int, marker string) int {
	m := []rune(marker)
	for i := from; i+len(m) <= len(r); i++ {
		if r[i] == '\\' {
			i++
			continue
		}
		if r[i] != m[0] {
			continue
		}
		if len(m) == 1 && m[0] == '*' && i+1 < len(r) && r[i+1] == '*' {
			i++ // skip a ** pair inside an *italic* span
			continue
		}
		if len(m) == 2 && (i+1 >= len(r) || r[i+1] != m[1]) {
			continue
		}
		return i
	}
	return -1
}
