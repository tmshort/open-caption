// Copyright the open-caption authors
// SPDX-License-Identifier: Apache-2.0

// Package help loads the in-app help topics (Markdown files embedded in the
// binary) and renders them to safe HTML (docs/05-ui-specs.md, section 4).
package help

import (
	"fmt"
	"html"
	"io/fs"
	"path"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/tmshort/open-caption/internal/showfile"
)

// Topic is one help page.
type Topic struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	Role  string `json:"role"` // "admin" | "operator"
	// Guide marks the long guides and references (docs/) as against the short how-to topics (web/help/), so the topic
	// list can show them under their own heading.
	Guide    bool   `json:"guide,omitempty"`
	Markdown string `json:"-"`
}

// topicOrder is the order of the topic list: what to read first, then the rest, then the guides and references.
// A topic that is not listed (a new file) goes after the listed ones of its kind, by id.
var topicOrder = []string{
	"admin-getting-started", "admin-preparing-a-show-file", "admin-writing-a-show-file", "admin-images", "admin-look-and-feel",
	"admin-languages", "admin-loops", "admin-musicals", "admin-network", "admin-troubleshooting",
	"operator-running-the-show", "operator-keyboard", "operator-indicators", "operator-loops", "operator-if-something-goes-wrong",
	"admin-usage-guide", "admin-venue-network-guide", "admin-https-certificate-guide", "admin-configuration-reference",
}

// SortTopics puts topics in topicOrder.
func SortTopics(ts []Topic) {
	rank := func(id string) int {
		for i, o := range topicOrder {
			if o == id {
				return i
			}
		}
		return len(topicOrder)
	}
	sort.SliceStable(ts, func(i, j int) bool {
		if ri, rj := rank(ts[i].ID), rank(ts[j].ID); ri != rj {
			return ri < rj
		}
		return ts[i].ID < ts[j].ID
	})
}

// Languages are the languages the help is written in; English is the original and the others are translations
// (web/help/<code>/<topic>.md and docs/i18n/<code>/<guide>.md). A topic with no translation shows in English.
var Languages = []string{"en", "es", "fr", "de", "it", "pt"}

// IsLanguage reports whether code is one of Languages.
func IsLanguage(code string) bool {
	for _, l := range Languages {
		if l == code {
			return true
		}
	}
	return false
}

var unaccent = strings.NewReplacer("á", "a", "à", "a", "â", "a", "ä", "a", "ã", "a", "å", "a", "ç", "c", "é", "e", "è", "e", "ê", "e", "ë", "e",
	"í", "i", "ì", "i", "î", "i", "ï", "i", "ñ", "n", "ó", "o", "ò", "o", "ô", "o", "ö", "o", "õ", "o", "ú", "u", "ù", "u", "û", "u", "ü", "u",
	"ý", "y", "ÿ", "y", "ß", "ss", "œ", "oe", "æ", "ae")

// Slug turns a heading into the id used for links to it ("Déploiement et réseau" is "deploiement-et-reseau").
func Slug(heading string) string {
	return strings.Trim(slugRe.ReplaceAllString(unaccent.Replace(strings.ToLower(heading)), "-"), "-")
}

// LoadLang returns base with each topic replaced by its translation into lang, read from dir/<lang>/<id>.md,
// where one exists. A file for a topic that is not in base is an error (it would never be shown).
func LoadLang(fsys fs.FS, dir, lang string, base []Topic) ([]Topic, error) {
	ents, err := fs.ReadDir(fsys, path.Join(dir, lang))
	if err != nil {
		return base, nil // no translations yet: English throughout
	}
	byID := map[string]Topic{}
	for _, e := range ents {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
			continue
		}
		id := strings.TrimSuffix(e.Name(), ".md")
		b, err := fs.ReadFile(fsys, path.Join(dir, lang, e.Name()))
		if err != nil {
			return nil, err
		}
		role, _, _ := strings.Cut(id, "-")
		t, err := FromMarkdown(id, role, string(b))
		if err != nil {
			return nil, fmt.Errorf("%s/%s: %w", lang, e.Name(), err)
		}
		byID[id] = t
	}
	out := make([]Topic, len(base))
	for i, b := range base {
		if t, ok := byID[b.ID]; ok {
			out[i] = t
			delete(byID, b.ID)
		} else {
			out[i] = b
		}
	}
	for id := range byID {
		return nil, fmt.Errorf("%s/%s.md translates a topic that does not exist", lang, id)
	}
	return out, nil
}

// plain removes Markdown punctuation from search snippets.
var plain = strings.NewReplacer("`", "", "**", "", "|", " ", "#", "", "\\", "", "[[", "", "]]", "", ">", "")

// Hit is a search result.
type Hit struct {
	ID      string `json:"id"`
	Title   string `json:"title"`
	Snippet string `json:"snippet"`
}

// Load reads every *.md file in dir. Files are named <role>-<topic>.md, where
// role is "admin" or "operator"; the first line is "# Title". Topics are
// ordered by file name.
func Load(fsys fs.FS, dir string) ([]Topic, error) {
	ents, err := fs.ReadDir(fsys, dir)
	if err != nil {
		return nil, err
	}
	var out []Topic
	for _, e := range ents {
		name := e.Name()
		if e.IsDir() || !strings.HasSuffix(name, ".md") {
			continue
		}
		id := strings.TrimSuffix(name, ".md")
		role, _, ok := strings.Cut(id, "-")
		if !ok || (role != "admin" && role != "operator") {
			return nil, fmt.Errorf("help file %q must start with admin- or operator-", name)
		}
		b, err := fs.ReadFile(fsys, path.Join(dir, name))
		if err != nil {
			return nil, err
		}
		md := string(b)
		first, _, _ := strings.Cut(md, "\n")
		if !strings.HasPrefix(first, "# ") {
			return nil, fmt.Errorf("help file %q must start with a '# Title' line", name)
		}
		out = append(out, Topic{ID: id, Title: strings.TrimSpace(first[2:]), Role: role, Markdown: md})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out, nil
}

// Search finds topics containing every word of q (case-insensitive).
func Search(topics []Topic, q string) []Hit {
	words := strings.Fields(strings.ToLower(q))
	if len(words) == 0 {
		return []Hit{}
	}
	hits := []Hit{}
	for _, t := range topics {
		low := strings.ToLower(t.Markdown)
		all := true
		for _, w := range words {
			if !strings.Contains(low, w) {
				all = false
				break
			}
		}
		if !all {
			continue
		}
		i := strings.Index(low, words[0])
		start, end := max(0, i-50), min(len(t.Markdown), i+90)
		snip := strings.Join(strings.Fields(plain.Replace(t.Markdown[start:end])), " ")
		hits = append(hits, Hit{ID: t.ID, Title: t.Title, Snippet: snip})
	}
	return hits
}

// CodeBlock is a fenced block found in a topic.
type CodeBlock struct {
	Info string // text after the opening fence, e.g. "showfile"
	Code string
}

// CodeBlocks returns the fenced blocks of a Markdown document.
func CodeBlocks(md string) []CodeBlock {
	var out []CodeBlock
	var cur *CodeBlock
	for _, ln := range strings.Split(md, "\n") {
		if strings.HasPrefix(ln, "```") {
			if cur == nil {
				cur = &CodeBlock{Info: strings.TrimSpace(ln[3:])}
			} else {
				cur.Code = strings.TrimSuffix(cur.Code, "\n")
				out = append(out, *cur)
				cur = nil
			}
			continue
		}
		if cur != nil {
			cur.Code += ln + "\n"
		}
	}
	return out
}

// ---- rendering ----

var (
	headRe = regexp.MustCompile(`^(#{1,4}) +(.+)$`)
	listRe = regexp.MustCompile(`^( *)([-*]|\d+\.) +(.+)$`)
	linkAt = regexp.MustCompile(`^\[((?:[^\]\\]|\\.)+)\]\(([^)\s]+)\)`)
	sepRe  = regexp.MustCompile(`^\|? *:?-{2,}:? *(\| *:?-{2,}:? *)*\|?$`)
	slugRe = regexp.MustCompile(`[^a-z0-9]+`)
	noteRe = regexp.MustCompile(`^> ?(.*)$`)
)

// Render converts a help topic's Markdown to HTML. All text is escaped; only
// tags generated here appear in the output. Links may point to another topic
// (topic:<id>) or to a path on this server (/...); anything else is plain text.
func Render(md string) string {
	var b strings.Builder
	lines := strings.Split(strings.ReplaceAll(md, "\r\n", "\n"), "\n")
	var para []string
	flush := func() {
		if len(para) > 0 {
			b.WriteString("<p>" + inline(strings.Join(para, " ")) + "</p>\n")
			para = nil
		}
	}
	for i := 0; i < len(lines); i++ {
		ln := lines[i]
		switch {
		case strings.HasPrefix(ln, "```"):
			flush()
			info := strings.TrimSpace(ln[3:])
			var code []string
			for i++; i < len(lines) && !strings.HasPrefix(lines[i], "```"); i++ {
				code = append(code, lines[i])
			}
			cls := ""
			if info != "" {
				cls = ` class="lang-` + html.EscapeString(slugRe.ReplaceAllString(strings.ToLower(info), "-")) + `"`
			}
			b.WriteString("<pre><code" + cls + ">" + html.EscapeString(strings.Join(code, "\n")) + "</code></pre>\n")
		case headRe.MatchString(ln):
			flush()
			m := headRe.FindStringSubmatch(ln)
			n := len(m[1])
			b.WriteString(fmt.Sprintf("<h%d id=\"%s\">%s</h%d>\n", n, Slug(m[2]), inline(m[2]), n))
		case strings.TrimSpace(ln) == "":
			flush()
		case listRe.MatchString(ln):
			flush()
			var h string
			h, i = renderList(lines, i)
			b.WriteString(h)
		case strings.HasPrefix(ln, "|"):
			flush()
			var rows [][]string
			for ; i < len(lines) && strings.HasPrefix(lines[i], "|"); i++ {
				if sepRe.MatchString(strings.TrimSpace(lines[i])) {
					continue
				}
				rows = append(rows, cells(lines[i]))
			}
			i--
			b.WriteString("<table>\n")
			for r, row := range rows {
				tag := "td"
				if r == 0 {
					tag = "th"
				}
				b.WriteString("<tr>")
				for _, c := range row {
					b.WriteString("<" + tag + ">" + inline(c) + "</" + tag + ">")
				}
				b.WriteString("</tr>\n")
			}
			b.WriteString("</table>\n")
		case noteRe.MatchString(ln) && strings.HasPrefix(ln, ">"):
			flush()
			var q []string
			for ; i < len(lines) && strings.HasPrefix(lines[i], ">"); i++ {
				q = append(q, noteRe.FindStringSubmatch(lines[i])[1])
			}
			i--
			b.WriteString("<blockquote><p>" + inline(strings.Join(q, " ")) + "</p></blockquote>\n")
		default:
			para = append(para, strings.TrimSpace(ln))
		}
	}
	flush()
	return b.String()
}

func cells(row string) []string {
	row = strings.TrimSpace(row)
	row = strings.TrimPrefix(row, "|")
	row = strings.TrimSuffix(row, "|")
	var out []string
	var cur strings.Builder
	for i := 0; i < len(row); i++ {
		switch {
		case row[i] == '\\' && i+1 < len(row) && row[i+1] == '|':
			cur.WriteByte('|') // \| is a literal pipe inside a cell
			i++
		case row[i] == '|':
			out = append(out, strings.TrimSpace(cur.String()))
			cur.Reset()
		default:
			cur.WriteByte(row[i])
		}
	}
	return append(out, strings.TrimSpace(cur.String()))
}

// renderList renders a (possibly nested) list starting at lines[i] and returns the HTML and the
// index of its last line. Nesting follows indentation; wrapped lines continue the previous item.
func renderList(lines []string, i int) (string, int) {
	type level struct {
		indent int
		tag    string
	}
	var b strings.Builder
	var stack []level
	closeTo := func(indent int) {
		for len(stack) > 0 && stack[len(stack)-1].indent >= indent {
			b.WriteString("</li>\n</" + stack[len(stack)-1].tag + ">\n")
			stack = stack[:len(stack)-1]
		}
	}
	var item strings.Builder
	flushItem := func() {
		if item.Len() > 0 {
			b.WriteString(inline(item.String()))
			item.Reset()
		}
	}
	last := i
	for ; i < len(lines); i++ {
		m := listRe.FindStringSubmatch(lines[i])
		if m == nil {
			if strings.HasPrefix(lines[i], " ") && strings.TrimSpace(lines[i]) != "" && len(stack) > 0 { // wrapped text
				item.WriteString(" " + strings.TrimSpace(lines[i]))
				last = i
				continue
			}
			break
		}
		flushItem()
		indent, tag := len(m[1]), "ul"
		if m[2] != "-" && m[2] != "*" {
			tag = "ol"
		}
		switch {
		case len(stack) == 0 || indent > stack[len(stack)-1].indent:
			if len(stack) > 0 {
				b.WriteString("\n")
			}
			b.WriteString("<" + tag + ">\n<li>")
			stack = append(stack, level{indent, tag})
		default:
			if indent < stack[len(stack)-1].indent {
				closeTo(indent + 1)
			}
			if len(stack) > 0 && stack[len(stack)-1].indent == indent {
				b.WriteString("</li>\n<li>")
			} else { // dedented to a level that was never opened
				b.WriteString("<" + tag + ">\n<li>")
				stack = append(stack, level{indent, tag})
			}
		}
		item.WriteString(m[3])
		last = i
	}
	flushItem()
	closeTo(0)
	return b.String(), last
}

var topicHref = regexp.MustCompile(`^(admin|operator)-[a-z0-9-]+$`)

// Aliases maps file names used in Markdown links (for example "usage-guide.md") to topic ids,
// so documents written for GitHub keep working when shown in the app.
var Aliases = map[string]string{
	"usage-guide.md":             "admin-usage-guide",
	"venue-network-guide.md":     "admin-venue-network-guide",
	"https-certificate-guide.md": "admin-https-certificate-guide",
	"configuration-reference.md": "admin-configuration-reference",
}

// pathLinks are help topics as GitHub shows them (web/help/<id>.md), so a guide can link to them from either place.
var pathLinks = map[string]string{
	"web/help/admin-preparing-a-show-file.md": "admin-preparing-a-show-file",
	"web/help/admin-getting-started.md":       "admin-getting-started",
	"web/help/admin-writing-a-show-file.md":   "admin-writing-a-show-file",
}

// relativeTarget drops the leading "../" of a relative link, so a guide links the same way from docs/ and from
// docs/i18n/<code>/ (where the right number of "../" differs but the target is the same file).
func relativeTarget(base string) string {
	for strings.HasPrefix(base, "../") {
		base = base[3:]
	}
	return base
}

// link resolves an href to a safe anchor, or returns just the text when the target is not available.
func link(text, href string) string {
	label := inline(text)
	frag := ""
	base := href
	if k := strings.IndexByte(href, '#'); k >= 0 {
		base, frag = href[:k], href[k+1:]
	}
	fragAttr := ""
	if frag != "" && !strings.ContainsAny(frag, `"'<>& `) {
		fragAttr = ` data-anchor="` + frag + `"`
	}
	rel := relativeTarget(base)
	switch {
	case strings.HasPrefix(base, "topic:") && topicHref.MatchString(base[6:]):
		return `<a href="#` + base[6:] + `" data-topic="` + base[6:] + `"` + fragAttr + `>` + label + "</a>"
	case rel == "sample/sample-show.md": // the starter file, as GitHub shows it; in the app it is the download
		return `<a href="/help/sample-show.md">` + label + "</a>"
	case pathLinks[rel] != "":
		return `<a href="#` + pathLinks[rel] + `" data-topic="` + pathLinks[rel] + `"` + fragAttr + `>` + label + "</a>"
	case Aliases[rel] != "":
		return `<a href="#` + Aliases[rel] + `" data-topic="` + Aliases[rel] + `"` + fragAttr + `>` + label + "</a>"
	case base == "" && fragAttr != "":
		return `<a href="#` + frag + `"` + fragAttr + `>` + label + "</a>"
	case strings.HasPrefix(base, "/") && !strings.HasPrefix(base, "//") && !strings.ContainsAny(href, `"'<>`):
		return `<a href="` + href + `">` + label + "</a>"
	}
	return label
}

// inline renders code spans, links, bold and italic. Code spans and links are set aside first
// (as placeholders) so that emphasis can wrap them, as in **a `code` word**. Link syntax inside a
// code span is left as typed.
func inline(s string) string {
	s = strings.ReplaceAll(s, "\x00", "")
	var tokens []string
	var plain strings.Builder
	hold := func(h string) {
		plain.WriteString("\x00" + strconv.Itoa(len(tokens)) + "\x00")
		tokens = append(tokens, h)
	}
	for i := 0; i < len(s); {
		switch s[i] {
		case '`':
			if j := strings.IndexByte(s[i+1:], '`'); j >= 0 {
				hold("<code>" + html.EscapeString(s[i+1:i+1+j]) + "</code>")
				i += j + 2
				continue
			}
		case '[':
			if m := linkAt.FindStringSubmatch(s[i:]); m != nil {
				hold(link(m[1], m[2]))
				i += len(m[0])
				continue
			}
		case '\\':
			if i+1 < len(s) {
				plain.WriteString(s[i : i+2])
				i += 2
				continue
			}
		}
		plain.WriteByte(s[i])
		i++
	}
	out := emph(plain.String())
	return placeholderRe.ReplaceAllStringFunc(out, func(m string) string {
		n, _ := strconv.Atoi(strings.Trim(m, "\x00"))
		return tokens[n]
	})
}

var placeholderRe = regexp.MustCompile("\x00[0-9]+\x00")

// emph applies **bold** and *italic* with the same renderer show files use (everything is escaped).
func emph(s string) string {
	h, _ := showfile.RenderInline(s)
	return h
}

// FromMarkdown builds a topic from a Markdown document whose first line is "# Title".
func FromMarkdown(id, role, md string) (Topic, error) {
	first, _, _ := strings.Cut(md, "\n")
	if !strings.HasPrefix(first, "# ") {
		return Topic{}, fmt.Errorf("topic %q must start with a '# Title' line", id)
	}
	// the title is shown as plain text in the topic list, so the backticks of a title such as "`config.yaml` reference" go
	return Topic{ID: id, Title: strings.ReplaceAll(strings.TrimSpace(first[2:]), "`", ""), Role: role, Markdown: md}, nil
}
