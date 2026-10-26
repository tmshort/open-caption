// Copyright the open-caption authors
// SPDX-License-Identifier: Apache-2.0

package help

import (
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/tmshort/open-caption/docs"
	"github.com/tmshort/open-caption/internal/showfile"
	"github.com/tmshort/open-caption/web"
)

func loadAll(t *testing.T) map[string]Topic {
	t.Helper()
	ts, err := Load(web.Help, "help")
	if err != nil {
		t.Fatal(err)
	}
	m := map[string]Topic{}
	for _, tp := range ts {
		m[tp.ID] = tp
	}
	return m
}

func TestRenderEscapesEverything(t *testing.T) {
	out := Render("# Title <script>\n\nText <b>x</b> & `code <i>` **bold**\n\n```showfile\n<script>alert(1)</script>\n```\n")
	for _, bad := range []string{"<script", "<b>", "<i>"} {
		if strings.Contains(out, bad) {
			t.Errorf("unescaped %q in %s", bad, out)
		}
	}
	for _, want := range []string{"<h1", "&lt;script&gt;", "<code>code &lt;i&gt;</code>", "<strong>bold</strong>", `class="lang-showfile"`} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in %s", want, out)
		}
	}
}

func TestLinksAreRestricted(t *testing.T) {
	out := Render("[ok](topic:admin-images) [path](/help/x) [bad](https://evil.example) [js](javascript:alert(1)) [proto](//evil.example/x)")
	if !strings.Contains(out, `data-topic="admin-images"`) || !strings.Contains(out, `href="/help/x"`) {
		t.Errorf("good links lost: %s", out)
	}
	for _, bad := range []string{`href="https`, `href="javascript`, `href="//`} {
		if strings.Contains(out, bad) {
			t.Errorf("unsafe link kept (%s): %s", bad, out)
		}
	}
}

func TestListsTablesQuotes(t *testing.T) {
	out := Render("- one\n- two\n\n1. first\n2. second\n\n| A | B |\n|---|---|\n| `x \\| y` | z |\n\n> careful\n")
	for _, want := range []string{"<ul>", "<li>two</li>", "<ol>", "<th>A</th>", "<td><code>x | y</code></td>", "<blockquote>"} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in\n%s", want, out)
		}
	}
}

func TestSearch(t *testing.T) {
	topics := []Topic{{ID: "admin-a", Title: "A", Markdown: "# A\nSend a message to the audience."}, {ID: "admin-b", Title: "B", Markdown: "# B\nUnrelated."}}
	hits := Search(topics, "AUDIENCE message")
	if len(hits) != 1 || hits[0].ID != "admin-a" || !strings.Contains(hits[0].Snippet, "message") {
		t.Errorf("%+v", hits)
	}
	if len(Search(topics, "  ")) != 0 {
		t.Error("blank query should match nothing")
	}
}

func TestEmbeddedTopicsLoadAndLinksResolve(t *testing.T) {
	topics := loadAll(t)
	for id, tp := range guideTopics(t) { // the guides from docs/ are topics too
		topics[id] = tp
	}
	if len(topics) < 10 {
		t.Fatalf("only %d topics", len(topics))
	}
	link := regexp.MustCompile(`\(topic:([a-z0-9-]+)\)`)
	for id, tp := range topics {
		for _, m := range link.FindAllStringSubmatch(tp.Markdown, -1) {
			if _, ok := topics[m[1]]; !ok {
				t.Errorf("%s links to missing topic %s", id, m[1])
			}
		}
		if out := Render(tp.Markdown); strings.Contains(out, "<script") {
			t.Errorf("%s renders a script tag", id)
		}
	}
}

// Help must document every show file convention the parser knows (REQ-037).
func TestScriptHelpCoversEveryConvention(t *testing.T) {
	topics := loadAll(t)
	text := topics["admin-writing-a-show-file"].Markdown + topics["admin-loops"].Markdown
	for _, name := range showfile.DirectiveNames {
		if !strings.Contains(text, "["+name) {
			t.Errorf("help never shows the [%s] command", name)
		}
	}
	for _, closer := range []string{"[/loop]", "[/fullscreen]", "[/together]", "[/settings]", " || ", "[[", "]]", "<!--", "-->", "![", "**", "\\\n"} {
		if !strings.Contains(text, closer) {
			t.Errorf("help never shows %q", closer)
		}
	}
	for _, name := range showfile.SettingNames() {
		if !strings.Contains(text, "`"+name+"`") {
			t.Errorf("help never documents the setting %q", name)
		}
	}
	for _, ch := range showfile.Escapable {
		if !strings.Contains(text, `\`+string(ch)) {
			t.Errorf("help never mentions the escape \\%c", ch)
		}
	}
	for _, want := range []string{"***", "~~", "H₂O"} {
		if !strings.Contains(topics["admin-writing-a-show-file"].Markdown, want) {
			t.Errorf("the show file topic never mentions %q (bold italic, strikethrough, Unicode sub/superscripts)", want)
		}
	}
	// The loop timer is explained: the option, both kinds of value and the placeholder.
	loops := topics["admin-loops"].Markdown
	for _, want := range []string{"timer 15", "timer 7:30pm", "{timer}", "[loop: timer", "name Intermission"} {
		if !strings.Contains(loops, want) {
			t.Errorf("the Loops topic never shows %q", want)
		}
	}
	// Every diagnostic the parser can raise is explained in plain words.
	for _, phrase := range []string{"Unknown setting", "never closed", "Unknown command", "not found", "never closed", "HTML is not supported", "can't go inside a loop"} {
		if !strings.Contains(text, phrase) {
			t.Errorf("help does not explain the warning %q", phrase)
		}
	}
}

// Every example marked ```showfile must be valid: no errors and no warnings.
func TestScriptExamplesParseClean(t *testing.T) {
	n := 0
	for id, tp := range loadAll(t) {
		for _, b := range CodeBlocks(tp.Markdown) {
			if b.Info != "showfile" {
				continue
			}
			n++
			for _, d := range showfile.Parse(b.Code, showfile.Options{}).Diagnostics {
				if d.Severity != showfile.SevInfo {
					t.Errorf("%s: example has a %s on line %d: %s\n%s", id, d.Severity, d.Line, d.Message, b.Code)
				}
			}
		}
	}
	if n < 5 {
		t.Errorf("expected several show file examples, found %d", n)
	}
}

// The keys documented for the operator must exist in the operator page's key handler.
func TestOperatorKeysDocumentedAndImplemented(t *testing.T) {
	doc := loadAll(t)["operator-keyboard"].Markdown
	js, err := fs.ReadFile(web.Operator, "operator/operator.js")
	if err != nil {
		t.Fatal(err)
	}
	pairs := []struct{ doc, code string }{
		{"Space", "case ' '"}, {"Page Down", "'PageDown'"}, {"Right arrow", "'ArrowRight'"}, {"Enter", "'Enter'"},
		{"Backspace", "'Backspace'"}, {"Page Up", "'PageUp'"}, {"Left arrow", "'ArrowLeft'"},
		{"Shift+Q", "case 'Q'"}, {"`B`", "case 'b'"}, {"`C`", "case 'c'"}, {"`/`", "case '/'"}, {"Home", "case 'Home'"}, {"Esc", "case 'Escape'"},
	}
	for _, p := range pairs {
		if !strings.Contains(doc, p.doc) {
			t.Errorf("keyboard help does not mention %s", p.doc)
		}
		if !strings.Contains(string(js), p.code) {
			t.Errorf("operator.js has no handler for documented key %s (%s)", p.doc, p.code)
		}
	}
}

func TestLoadRejectsBadFiles(t *testing.T) {
	good := fstest_("admin-x.md", "# T\nbody")
	if _, err := Load(good, "."); err != nil {
		t.Fatal(err)
	}
	for name, body := range map[string]string{"x-y.md": "# T", "admin-y.md": "no title"} {
		if _, err := Load(fstest_(name, body), "."); err == nil {
			t.Errorf("%s should be rejected", name)
		}
	}
}

func guideTopics(t *testing.T) map[string]Topic {
	t.Helper()
	out := map[string]Topic{}
	for file, id := range Aliases {
		b, err := fs.ReadFile(docs.Guides, file)
		if err != nil {
			t.Fatal(err)
		}
		tp, err := FromMarkdown(id, "admin", string(b))
		if err != nil {
			t.Fatal(err)
		}
		out[id] = tp
	}
	return out
}

func TestNestedLists(t *testing.T) {
	out := Render("- one\n  - nested a\n  - nested b\n- two wrapped\n  onto the next line\n\n1. first\n   - sub\n2. second")
	for _, want := range []string{"<li>one\n<ul>", "<li>nested a</li>", "<li>nested b</li>", "<li>two wrapped onto the next line</li>", "<ol>", "<li>sub</li>", "<li>second</li>"} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in\n%s", want, out)
		}
	}
	if strings.Count(out, "<ul>") != strings.Count(out, "</ul>") || strings.Count(out, "<li>") != strings.Count(out, "</li>") {
		t.Errorf("unbalanced list tags:\n%s", out)
	}
}

func TestEmphasisCanWrapCodeAndLinks(t *testing.T) {
	out := Render("**Option 114 and `/captive-api` (advanced)** and *see [images](topic:admin-images)*")
	if !strings.Contains(out, "<strong>Option 114 and <code>/captive-api</code> (advanced)</strong>") ||
		!strings.Contains(out, `<em>see <a href="#admin-images"`) || strings.Contains(out, "**") {
		t.Errorf("%s", out)
	}
	if out := Render("odd \x00 0 \x00 text `x`"); strings.Contains(out, "\x00") {
		t.Errorf("placeholder characters leaked: %q", out)
	}
}

func TestLinksInsideCodeAreLeftAlone(t *testing.T) {
	out := Render("Write `![caption](file.png)` or [a link](topic:admin-images) or [`code` text](/help/x).")
	if !strings.Contains(out, "<code>![caption](file.png)</code>") || !strings.Contains(out, `data-topic="admin-images"`) || !strings.Contains(out, `href="/help/x"><code>code</code> text`) {
		t.Errorf("%s", out)
	}
	// A link we cannot honour shows its words, not the raw Markdown.
	out = Render("See [`deploy/`](../deploy/) and [the web](https://example.com).")
	if strings.Contains(out, "](") || strings.Contains(out, "href") || !strings.Contains(out, "<code>deploy/</code>") || !strings.Contains(out, "the web") {
		t.Errorf("%s", out)
	}
}

// The guides in docs/ ship inside the program, so they must display properly there.
func TestGuidesRenderCleanlyInTheApp(t *testing.T) {
	topics := guideTopics(t)
	if len(topics) != 4 {
		t.Fatalf("guides: %d", len(topics))
	}
	all := loadAll(t)
	known := func(id string) bool { _, a := all[id]; _, b := topics[id]; return a || b }
	for id, tp := range topics {
		out := Render(tp.Markdown)
		// no Markdown left over in the HTML (outside code)
		noCode := regexp.MustCompile(`(?s)<code[^>]*>.*?</code>`).ReplaceAllString(out, "")
		for _, bad := range []string{"](", "**", "<script"} {
			if strings.Contains(noCode, bad) {
				i := strings.Index(noCode, bad)
				t.Errorf("%s: leftover %q near %q", id, bad, noCode[max(0, i-60):min(len(noCode), i+60)])
			}
		}
		ids := regexp.MustCompile(`<h[1-4] id="([^"]*)"`).FindAllStringSubmatch(out, -1)
		have := map[string]bool{}
		for _, m := range ids {
			have[m[1]] = true
		}
		for _, m := range regexp.MustCompile(`data-topic="([^"]+)"`).FindAllStringSubmatch(out, -1) {
			if !known(m[1]) {
				t.Errorf("%s links to unknown topic %s", id, m[1])
			}
		}
		for _, m := range regexp.MustCompile(`data-anchor="([^"]+)"`).FindAllStringSubmatch(out, -1) {
			if !have[m[1]] && !strings.Contains(m[0], "topic") {
				// anchors in these guides point at headings of the same guide
				t.Errorf("%s: link to #%s has no matching heading", id, m[1])
			}
		}
		if !strings.Contains(out, "<table>") || !strings.Contains(out, "<ul>") {
			t.Errorf("%s: expected tables and lists", id)
		}
	}
	// The usage guide links to the venue guide; both resolve to topics.
	if out := Render(topics["admin-usage-guide"].Markdown); !strings.Contains(out, `data-topic="admin-venue-network-guide"`) {
		t.Error("usage guide does not link to the venue network guide")
	}
	// The usage guide and the venue guide both lead to the certificate guide (a link that works in the app).
	for _, id := range []string{"admin-usage-guide", "admin-venue-network-guide"} {
		if out := Render(topics[id].Markdown); !strings.Contains(out, `data-topic="admin-https-certificate-guide"`) {
			t.Errorf("%s does not link to the certificate guide", id)
		}
	}
	// ...and to the preparer's topics, by the path GitHub shows.
	for _, id := range []string{"admin-preparing-a-show-file", "admin-getting-started"} {
		if out := Render(topics["admin-usage-guide"].Markdown); !strings.Contains(out, `data-topic="`+id+`"`) {
			t.Errorf("usage guide does not link to %s", id)
		}
	}
}

// ---- translations ----

var linkTargetRe = regexp.MustCompile(`\]\(([^)\s]+)\)`)
var inlineCodeRe = regexp.MustCompile("`[^`]*`")

type docShape struct {
	heads   []string // "#"-count for each heading, in order
	slugs   map[string]bool
	tables  int
	code    []CodeBlock
	targets []string // link targets other than in-page anchors, sorted
	anchors []string // in-page anchors, in order
}

func shapeOf(md string) docShape {
	s := docShape{slugs: map[string]bool{}, code: CodeBlocks(md)}
	inFence := false
	for _, ln := range strings.Split(md, "\n") {
		if strings.HasPrefix(ln, "```") {
			inFence = !inFence
			continue
		}
		if inFence {
			continue
		}
		if m := headRe.FindStringSubmatch(ln); m != nil {
			s.heads = append(s.heads, m[1])
			s.slugs[Slug(m[2])] = true
		}
		if strings.HasPrefix(ln, "|") {
			s.tables++
		}
		for _, m := range linkTargetRe.FindAllStringSubmatch(inlineCodeRe.ReplaceAllString(ln, ""), -1) {
			if strings.HasPrefix(m[1], "#") {
				s.anchors = append(s.anchors, m[1][1:])
			} else {
				s.targets = append(s.targets, relativeTarget(m[1])) // docs/ and docs/i18n/<code>/ differ in the number of ../
			}
		}
	}
	sort.Strings(s.targets)
	return s
}

// translatedDocs returns every English document that has to be translated, by id: the help topics and the four guides.
func englishDocs(t *testing.T) map[string]string {
	t.Helper()
	out := map[string]string{}
	for id, tp := range loadAll(t) {
		out[id] = tp.Markdown
	}
	for file, id := range Aliases {
		b, err := fs.ReadFile(docs.Guides, file)
		if err != nil {
			t.Fatal(err)
		}
		out[id] = string(b)
	}
	return out
}

func translatedDoc(t *testing.T, lang, id string) (string, bool) {
	t.Helper()
	for file, gid := range Aliases {
		if gid == id {
			b, err := fs.ReadFile(docs.Guides, "i18n/"+lang+"/"+file)
			return string(b), err == nil
		}
	}
	b, err := fs.ReadFile(web.Help, "help/"+lang+"/"+id+".md")
	return string(b), err == nil
}

func TestAnchorsInEveryDocumentPointAtAHeading(t *testing.T) {
	for id, md := range englishDocs(t) {
		s := shapeOf(md)
		for _, a := range s.anchors {
			if !s.slugs[a] {
				t.Errorf("%s: the link #%s points at no heading", id, a)
			}
		}
	}
}

func TestEveryTopicAndGuideIsTranslatedIntoEveryLanguageWithTheSameShape(t *testing.T) {
	for id, en := range englishDocs(t) {
		es := shapeOf(en)
		for _, lang := range Languages[1:] {
			md, ok := translatedDoc(t, lang, id)
			if !ok {
				t.Errorf("%s: no %s translation", id, lang)
				continue
			}
			first, _, _ := strings.Cut(md, "\n")
			if !strings.HasPrefix(first, "# ") || strings.TrimSpace(first[2:]) == "" {
				t.Errorf("%s/%s: must start with a '# Title' line", lang, id)
			}
			if md == en {
				t.Errorf("%s/%s: identical to the English", lang, id)
			}
			s := shapeOf(md)
			if strings.Join(s.heads, "") != strings.Join(es.heads, "") {
				t.Errorf("%s/%s: headings differ: %v, English %v", lang, id, s.heads, es.heads)
			}
			if s.tables != es.tables {
				t.Errorf("%s/%s: %d table lines, English has %d", lang, id, s.tables, es.tables)
			}
			if len(s.code) != len(es.code) {
				t.Errorf("%s/%s: %d code blocks, English has %d", lang, id, len(s.code), len(es.code))
			} else {
				for i := range s.code {
					if s.code[i] != es.code[i] {
						t.Errorf("%s/%s: code block %d was changed (show file examples must be copied exactly):\n%s\nEnglish:\n%s", lang, id, i+1, s.code[i].Code, es.code[i].Code)
					}
				}
			}
			if strings.Join(s.targets, " ") != strings.Join(es.targets, " ") {
				t.Errorf("%s/%s: links differ: %v, English %v", lang, id, s.targets, es.targets)
			}
			if len(s.anchors) != len(es.anchors) {
				t.Errorf("%s/%s: %d in-page links, English has %d", lang, id, len(s.anchors), len(es.anchors))
			}
			for _, a := range s.anchors {
				if !s.slugs[a] {
					t.Errorf("%s/%s: the link #%s points at no heading (the id is the translated heading without accents, lower case, with - between words)", lang, id, a)
				}
			}
		}
	}
}

func TestTranslatedTopicsRender(t *testing.T) {
	for id, en := range englishDocs(t) {
		for _, lang := range Languages[1:] {
			md, ok := translatedDoc(t, lang, id)
			if !ok {
				continue
			}
			out := Render(md)
			if strings.Contains(out, "<script") || strings.Count(out, "<h") == 0 || len(out) < len(Render(en))/4 {
				t.Errorf("%s/%s renders badly (%d bytes)", lang, id, len(out))
			}
		}
	}
}

func TestSlugIgnoresAccents(t *testing.T) {
	for in, want := range map[string]string{"Déploiement et réseau": "deploiement-et-reseau", "Über uns": "uber-uns", "Guía de instalación": "guia-de-instalacion", "Straße": "strasse", "Configuração": "configuracao"} {
		if got := Slug(in); got != want {
			t.Errorf("%q: %q", in, got)
		}
	}
}

// A translated guide lives one folder deeper (docs/i18n/<code>/), and people also read it on GitHub: every relative
// link in it must reach an existing file from where it is (a sibling translated guide, or a file up in docs/ or the repository).
func TestLinksInTranslatedGuidesWorkWhereTheyAre(t *testing.T) {
	files, _ := filepath.Glob("../../docs/i18n/*/*.md")
	if len(files) < 10 {
		t.Fatalf("found only %d translated guides", len(files))
	}
	for _, f := range files {
		b, err := os.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		text := regexp.MustCompile("(?s)```.*?```").ReplaceAllString(string(b), "")
		text = inlineCodeRe.ReplaceAllString(text, "")
		for _, m := range linkTargetRe.FindAllStringSubmatch(text, -1) {
			target := m[1]
			if strings.HasPrefix(target, "#") || strings.Contains(target, "://") || strings.HasPrefix(target, "mailto:") || strings.HasPrefix(target, "topic:") {
				continue
			}
			target, _, _ = strings.Cut(target, "#")
			if _, err := os.Stat(filepath.Join(filepath.Dir(f), target)); err != nil {
				t.Errorf("%s: the link %q does not reach a file from here (links in translations are relative to docs/i18n/<code>/)", f, m[1])
			}
		}
	}
}

func TestGuideLinksResolveWhateverTheNumberOfDotDotIs(t *testing.T) {
	for _, href := range []string{"../web/help/admin-getting-started.md", "../../../web/help/admin-getting-started.md"} {
		if out := link("Getting started", href); !strings.Contains(out, `data-topic="admin-getting-started"`) {
			t.Errorf("%s: %s", href, out)
		}
	}
	for _, href := range []string{"venue-network-guide.md", "../venue-network-guide.md"} {
		if out := link("venue", href); !strings.Contains(out, `data-topic="admin-venue-network-guide"`) {
			t.Errorf("%s: %s", href, out)
		}
	}
	if out := link("x", "../../../etc/passwd"); strings.Contains(out, "<a ") {
		t.Errorf("an unknown path must not become a link: %s", out)
	}
}

func TestTopicListOrderPutsGettingStartedFirstAndTheGuidesLast(t *testing.T) {
	all := loadAll(t)
	listed := map[string]bool{}
	for _, id := range topicOrder {
		listed[id] = true
	}
	var ts []Topic
	for id, tp := range all {
		if !listed[id] {
			t.Errorf("the topic %s is not in topicOrder: add it where it belongs in the list", id)
		}
		ts = append(ts, tp)
	}
	for id := range guideTopics(t) {
		if !listed[id] {
			t.Errorf("the guide %s is not in topicOrder", id)
		}
	}
	SortTopics(ts)
	if len(ts) == 0 || ts[0].ID != "admin-getting-started" {
		t.Errorf("the first topic should be Getting started, not %v", ts[0].ID)
	}
	if tp, err := FromMarkdown("admin-x", "admin", "# `config.yaml` reference\n\nText"); err != nil || tp.Title != "config.yaml reference" {
		t.Errorf("a title is plain text in the list: %q %v", tp.Title, err)
	}
	// unknown topics follow the listed ones
	more := []Topic{{ID: "admin-zzz-new"}, {ID: "admin-getting-started"}}
	SortTopics(more)
	if more[0].ID != "admin-getting-started" || more[1].ID != "admin-zzz-new" {
		t.Errorf("%v", more)
	}
}

// The help topics link to the starter show file by its place in the repository, so the link works on GitHub;
// in the app it is the download.
func TestSampleShowLinkWorksOnGitHubAndInTheApp(t *testing.T) {
	for _, in := range []string{"../sample/sample-show.md", "../../sample/sample-show.md"} {
		if got := link("starter", in); got != `<a href="/help/sample-show.md">starter</a>` {
			t.Errorf("%s -> %s", in, got)
		}
	}
	if _, err := os.Stat("../../web/sample/sample-show.md"); err != nil {
		t.Errorf("the file the topics link to is missing: %v", err)
	}
}

// A topic can link to another as a file beside it, which GitHub can follow; in the program it is the same link as topic:.
func TestTopicFileLinksWorkLikeTopicLinks(t *testing.T) {
	want := `<a href="#admin-loops" data-topic="admin-loops">Loops</a>`
	for _, in := range []string{"topic:admin-loops", "admin-loops.md"} {
		if got := link("Loops", in); got != want {
			t.Errorf("%s: %s", in, got)
		}
	}
	if got := link("x", "admin-nonsense/../evil.md"); strings.Contains(got, "<a") {
		t.Errorf("a path that is not a topic file became a link: %s", got)
	}
}
