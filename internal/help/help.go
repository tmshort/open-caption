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
	"strings"

	"github.com/tmshort/open-caption/internal/script"
)

// Topic is one help page.
type Topic struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	Role     string `json:"role"` // "admin" | "operator"
	Markdown string `json:"-"`
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
	Info string // text after the opening fence, e.g. "script"
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
	ulRe   = regexp.MustCompile(`^[-*] +(.+)$`)
	olRe   = regexp.MustCompile(`^\d+\. +(.+)$`)
	linkRe = regexp.MustCompile(`\[([^\]]+)\]\(([^)\s]+)\)`)
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
			b.WriteString(fmt.Sprintf("<h%d id=\"%s\">%s</h%d>\n", n, strings.Trim(slugRe.ReplaceAllString(strings.ToLower(m[2]), "-"), "-"), inline(m[2]), n))
		case strings.TrimSpace(ln) == "":
			flush()
		case ulRe.MatchString(ln) || olRe.MatchString(ln):
			flush()
			tag, re := "ul", ulRe
			if olRe.MatchString(ln) {
				tag, re = "ol", olRe
			}
			b.WriteString("<" + tag + ">\n")
			for ; i < len(lines) && re.MatchString(lines[i]); i++ {
				item := re.FindStringSubmatch(lines[i])[1]
				for i+1 < len(lines) && strings.HasPrefix(lines[i+1], "  ") && !re.MatchString(lines[i+1]) && strings.TrimSpace(lines[i+1]) != "" {
					i++
					item += " " + strings.TrimSpace(lines[i])
				}
				b.WriteString("<li>" + inline(item) + "</li>\n")
			}
			i--
			b.WriteString("</" + tag + ">\n")
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

// inline renders code spans, links, bold and italic.
func inline(s string) string {
	var out strings.Builder
	for len(s) > 0 {
		i := strings.IndexByte(s, '`')
		if i < 0 {
			out.WriteString(links(s))
			break
		}
		j := strings.IndexByte(s[i+1:], '`')
		if j < 0 {
			out.WriteString(links(s))
			break
		}
		out.WriteString(links(s[:i]))
		out.WriteString("<code>" + html.EscapeString(s[i+1:i+1+j]) + "</code>")
		s = s[i+j+2:]
	}
	return out.String()
}

func links(s string) string {
	var out strings.Builder
	last := 0
	for _, m := range linkRe.FindAllStringSubmatchIndex(s, -1) {
		out.WriteString(emph(s[last:m[0]]))
		text, href := s[m[2]:m[3]], s[m[4]:m[5]]
		switch {
		case strings.HasPrefix(href, "topic:") && validID(href[6:]):
			out.WriteString(`<a href="#` + href[6:] + `" data-topic="` + href[6:] + `">` + emph(text) + "</a>")
		case strings.HasPrefix(href, "/") && !strings.HasPrefix(href, "//") && !strings.ContainsAny(href, `"'<>`):
			out.WriteString(`<a href="` + href + `">` + emph(text) + "</a>")
		default:
			out.WriteString(emph(s[m[0]:m[1]]))
		}
		last = m[1]
	}
	out.WriteString(emph(s[last:]))
	return out.String()
}

func validID(id string) bool {
	return regexp.MustCompile(`^(admin|operator)-[a-z0-9-]+$`).MatchString(id)
}

// emph applies **bold** and *italic* with the same renderer scripts use (everything is escaped).
func emph(s string) string {
	h, _ := script.RenderInline(s)
	return h
}
