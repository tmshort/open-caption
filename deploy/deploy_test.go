package deploy

import (
	"encoding/xml"
	"os"
	"regexp"
	"strings"
	"testing"
)

// The start-at-boot recipes are not run by CI, so this keeps them honest: they must be well formed and use
// only options the program really has.
func definedFlags(t *testing.T) map[string]bool {
	src, err := os.ReadFile("../cmd/open-caption/main.go")
	if err != nil {
		t.Fatal(err)
	}
	flags := map[string]bool{}
	for _, m := range regexp.MustCompile(`flag\.(?:String|Bool|Int)\("([a-z-]+)"`).FindAllStringSubmatch(string(src), -1) {
		flags[m[1]] = true
	}
	if !flags["data"] || !flags["no-browser"] || !flags["log-file"] {
		t.Fatalf("could not read the flags from main.go: %v", flags)
	}
	return flags
}

func TestRecipesUseRealFlags(t *testing.T) {
	flags := definedFlags(t)
	for _, f := range []string{"open-caption.service", "org.open-caption.server.plist", "windows-install.ps1"} {
		b, err := os.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		// Only the lines that start the program (the comments mention other commands' options).
		for _, line := range strings.Split(string(b), "\n") {
			if t := strings.TrimSpace(line); strings.HasPrefix(t, "#") || strings.HasPrefix(t, "<!--") {
				continue
			}
			for _, u := range regexp.MustCompile(`--([a-z][a-z-]+)`).FindAllStringSubmatch(line, -1) {
				if !flags[u[1]] {
					t.Errorf("%s uses --%s, which the program does not have", f, u[1])
				}
			}
		}
		for _, must := range []string{"--no-browser", "--data"} {
			if !strings.Contains(string(b), must) {
				t.Errorf("%s does not pass %s", f, must)
			}
		}
	}
}

func TestLaunchdPlistIsWellFormed(t *testing.T) {
	b, err := os.ReadFile("org.open-caption.server.plist")
	if err != nil {
		t.Fatal(err)
	}
	dec := xml.NewDecoder(strings.NewReader(string(b)))
	dec.Strict = false
	var keys []string
	for {
		tok, err := dec.Token()
		if err != nil {
			break
		}
		if se, ok := tok.(xml.StartElement); ok && se.Name.Local == "key" {
			var k string
			dec.DecodeElement(&k, &se)
			keys = append(keys, k)
		}
	}
	got := strings.Join(keys, ",")
	for _, want := range []string{"Label", "ProgramArguments", "RunAtLoad", "KeepAlive"} {
		if !strings.Contains(got, want) {
			t.Errorf("plist lacks %s (found %s)", want, got)
		}
	}
}

func TestSystemdUnitHasItsSections(t *testing.T) {
	b, _ := os.ReadFile("open-caption.service")
	s := string(b)
	for _, want := range []string{"\n[Unit]\n", "\n[Service]\n", "\n[Install]\n", "ExecStart=/usr/local/bin/open-caption ", "Restart=", "WantedBy=multi-user.target"} {
		if !strings.Contains(s, want) {
			t.Errorf("unit lacks %q", want)
		}
	}
	if exec := regexp.MustCompile(`ExecStart=.*`).FindString(s); strings.Contains(exec, "--log-file") && !strings.Contains(s, "ReadWritePaths=/var/lib/open-caption") {
		t.Error("the log file is written outside the paths the sandbox allows")
	}
}

// build.ps1 is what Windows users run instead of `make build`; it must keep building the same thing.
func TestWindowsBuildScriptMatchesTheMakefile(t *testing.T) {
	mk, _ := os.ReadFile("../Makefile")
	ps, err := os.ReadFile("../build.ps1")
	if err != nil {
		t.Fatal(err)
	}
	bat, err := os.ReadFile("../build.bat")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(bat), "build.ps1") {
		t.Error("build.bat must run build.ps1")
	}
	// the default providers, the version flags and the package are the Makefile's
	m := regexp.MustCompile(`(?m)^ACME_PROVIDERS \?= (.+)$`).FindStringSubmatch(string(mk))
	if m == nil || !strings.Contains(string(ps), "[string]$AcmeProviders = '"+m[1]+"'") {
		t.Errorf("build.ps1's default -AcmeProviders differs from the Makefile's ACME_PROVIDERS (%v)", m)
	}
	l := regexp.MustCompile(`(?m)^LDFLAGS := (.+)$`).FindStringSubmatch(string(mk))
	if l == nil || !strings.Contains(string(ps), strings.Replace(l[1], "$(VERSION)", "$version", 1)) {
		t.Errorf("build.ps1's -ldflags differ from the Makefile's LDFLAGS (%v)", l)
	}
	if !strings.Contains(string(ps), `.\cmd\open-caption`) || !strings.Contains(string(ps), "tools/acme-go.sh") {
		t.Error("build.ps1 must build ./cmd/open-caption and use tools/acme-go.sh for other providers")
	}
}
