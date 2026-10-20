// Copyright the open-caption authors
// SPDX-License-Identifier: Apache-2.0

package deploy

import (
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

// packageFiles are the files a release download holds besides the program, as .goreleaser.yaml lists them
// (globs expanded). The generated README is named by its place in the download.
func packageFiles(t *testing.T) (files map[string]string, cfg map[string]any) {
	t.Helper()
	b, err := os.ReadFile("../.goreleaser.yaml")
	if err != nil {
		t.Fatal(err)
	}
	var y struct {
		Archives []struct {
			Files []any `yaml:"files"`
		} `yaml:"archives"`
	}
	if err := yaml.Unmarshal(b, &y); err != nil || len(y.Archives) != 1 {
		t.Fatalf("cannot read the archive section: %v", err)
	}
	files = map[string]string{} // path in the download -> file on disk
	for _, f := range y.Archives[0].Files {
		src, dst := "", ""
		switch v := f.(type) {
		case string:
			src = v
		case map[string]any:
			src, _ = v["src"].(string)
			dst, _ = v["dst"].(string)
		}
		if strings.Contains(src, "**") {
			t.Fatalf("%s: use plain * patterns (GoReleaser's ** does not match files directly in the folder)", src)
		}
		if strings.ContainsAny(src, "*?[") {
			matches, _ := filepath.Glob("../" + src)
			if len(matches) == 0 {
				t.Errorf("%s matches no file", src)
			}
			for _, m := range matches {
				files[strings.TrimPrefix(filepath.ToSlash(m), "../")] = m
			}
			continue
		}
		if dst == "" {
			dst = src
		}
		disk := "../" + src
		if src == ".tools/README.release.md" {
			disk = ""
		}
		files[dst] = disk
	}
	return files, nil
}

// The download holds the guides for the people who run the program, never the design specifications (01 to 09).
func TestReleasePackageHoldsTheGuidesAndNotTheDesignDocuments(t *testing.T) {
	files, _ := packageFiles(t)
	for _, want := range []string{"README.md", "LICENSE", "NOTICE", "docs/open-caption.1", "docs/usage-guide.md", "docs/venue-network-guide.md",
		"docs/https-certificate-guide.md", "docs/configuration-reference.md", "docs/i18n/es/usage-guide.md", "web/help/admin-getting-started.md", "deploy/open-caption.service"} {
		if _, ok := files[want]; !ok {
			t.Errorf("the release download lacks %s", want)
		}
	}
	for p, disk := range files {
		if regexp.MustCompile(`^docs/(0\d|10)-`).MatchString(p) {
			t.Errorf("%s is a design document (01 to 10) and must not be in the download", p)
		}
		if disk != "" {
			if _, err := os.Stat(disk); err != nil {
				t.Errorf("%s is listed for the download but missing: %v", p, err)
			}
		}
	}
}

// Every relative link inside the download must lead to a file that is in the download, so nobody meets a dead link
// in a README or guide they were given.
func TestReleasePackageHasNoDeadLinks(t *testing.T) {
	if _, err := exec.LookPath("sh"); err != nil {
		t.Skip("needs a Unix shell for tools/package-readme.sh")
	}
	out := filepath.Join(t.TempDir(), "README.md")
	if b, err := exec.Command("sh", "../tools/package-readme.sh", out).CombinedOutput(); err != nil {
		t.Fatalf("package-readme.sh: %v\n%s", err, b)
	}
	files, _ := packageFiles(t)
	files["README.md"] = out
	link := regexp.MustCompile(`\]\(([^)\s]+)\)`)
	var bad []string
	for p, disk := range files {
		if !strings.HasSuffix(p, ".md") || strings.HasPrefix(p, "web/sample/") { // (the sample shows name images, they are not documents)
			continue
		}
		b, err := os.ReadFile(disk)
		if err != nil {
			t.Fatal(err)
		}
		text := regexp.MustCompile("(?s)```.*?```").ReplaceAllString(string(b), "") // examples are not links
		text = regexp.MustCompile("`[^`\n]*`").ReplaceAllString(text, "")
		for _, m := range link.FindAllStringSubmatch(text, -1) {
			target := m[1]
			if strings.Contains(target, "://") || strings.HasPrefix(target, "#") || strings.HasPrefix(target, "mailto:") ||
				strings.HasPrefix(target, "topic:") || strings.HasPrefix(target, "/") { // /help/... and /operator are the program's own addresses
				continue
			}
			target = strings.SplitN(target, "#", 2)[0]
			if target == "" {
				continue
			}
			resolved := path.Clean(path.Join(path.Dir(p), target))
			if _, ok := files[resolved]; !ok && !hasFolder(files, resolved) {
				bad = append(bad, p+" -> "+m[1])
			}
		}
	}
	sort.Strings(bad)
	for _, b := range bad {
		t.Error("dead link in the release download: " + b)
	}
}

// The release workflow and the Makefile use the same GoReleaser version.
func TestReleaseWorkflowUsesTheMakefile(t *testing.T) {
	wf, err := os.ReadFile("../.github/workflows/release.yml")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(wf), "make release") {
		t.Error("release.yml must run `make release`, so a local release and the workflow are the same")
	}
	if !strings.Contains(string(wf), "'v*'") && !strings.Contains(string(wf), `"v*"`) {
		t.Error("release.yml must start on a v* tag")
	}
}

// hasFolder reports whether the download holds anything inside the folder dir (a link to "deploy/").
func hasFolder(files map[string]string, dir string) bool {
	dir = strings.TrimSuffix(dir, "/") + "/"
	for p := range files {
		if strings.HasPrefix(p, dir) {
			return true
		}
	}
	return false
}
