package help

import (
	"io/fs"
	"regexp"
	"strings"
	"testing"

	"github.com/tmshort/open-caption/internal/script"
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
	out := Render("# Title <script>\n\nText <b>x</b> & `code <i>` **bold**\n\n```script\n<script>alert(1)</script>\n```\n")
	for _, bad := range []string{"<script", "<b>", "<i>"} {
		if strings.Contains(out, bad) {
			t.Errorf("unescaped %q in %s", bad, out)
		}
	}
	for _, want := range []string{"<h1", "&lt;script&gt;", "<code>code &lt;i&gt;</code>", "<strong>bold</strong>", `class="lang-script"`} {
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

// Help must document every script convention the parser knows (REQ-037).
func TestScriptHelpCoversEveryConvention(t *testing.T) {
	topics := loadAll(t)
	text := topics["admin-writing-a-script"].Markdown + topics["admin-loops"].Markdown
	for _, name := range script.DirectiveNames {
		if !strings.Contains(text, "["+name) {
			t.Errorf("help never shows the [%s] command", name)
		}
	}
	for _, closer := range []string{"[/loop]", "[/fullscreen]", "[/together]", " || ", "[[", "]]", "<!--", "-->", "![", "**", "\\\n"} {
		if !strings.Contains(text, closer) {
			t.Errorf("help never shows %q", closer)
		}
	}
	for _, ch := range script.Escapable {
		if !strings.Contains(text, `\`+string(ch)) {
			t.Errorf("help never mentions the escape \\%c", ch)
		}
	}
	// Every diagnostic the parser can raise is explained in plain words.
	for _, phrase := range []string{"Unknown command", "not found", "never closed", "HTML is not supported", "can't go inside a loop"} {
		if !strings.Contains(text, phrase) {
			t.Errorf("help does not explain the warning %q", phrase)
		}
	}
}

// Every example marked ```script must be valid: no errors and no warnings.
func TestScriptExamplesParseClean(t *testing.T) {
	n := 0
	for id, tp := range loadAll(t) {
		for _, b := range CodeBlocks(tp.Markdown) {
			if b.Info != "script" {
				continue
			}
			n++
			for _, d := range script.Parse(b.Code, script.Options{}).Diagnostics {
				if d.Severity != script.SevInfo {
					t.Errorf("%s: example has a %s on line %d: %s\n%s", id, d.Severity, d.Line, d.Message, b.Code)
				}
			}
		}
	}
	if n < 5 {
		t.Errorf("expected several script examples, found %d", n)
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
