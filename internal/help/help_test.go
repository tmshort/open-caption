package help

import (
	"io/fs"
	"regexp"
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
	if len(topics) != 2 {
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
}
