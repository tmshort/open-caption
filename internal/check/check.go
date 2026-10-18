// Package check is the analysis of a show file that Setup's editor and the command line (open-caption --check-show)
// share: which images a show file uses and which are missing, and the report the command line prints. The parsing
// itself is internal/showfile; this package adds what is around it. It reads files but never starts a server.
package check

import (
	"sort"

	"github.com/tmshort/open-caption/internal/showfile"
)

// Missing names an image the show file uses but that was not uploaded.
type Missing struct {
	Name  string `json:"name"`
	Lines []int  `json:"lines"`
}

// ImageUse lists, for every image the cues show, the source lines that use it (a loop's images count too).
func ImageUse(cues []showfile.Cue) map[string][]int {
	used := map[string][]int{}
	note := func(it showfile.Item, line int) {
		if it.Kind == "image" {
			used[it.Image] = append(used[it.Image], line)
		}
	}
	for _, c := range cues {
		if c.Item != nil {
			note(*c.Item, c.Line)
		}
		for _, it := range c.Items {
			note(it, c.Line)
		}
	}
	return used
}

// MissingImages are the images in used that are not in have, sorted by name (never nil, so JSON gets []).
func MissingImages(used map[string][]int, have map[string]bool) []Missing {
	missing := []Missing{}
	for name, lines := range used {
		if !have[name] {
			missing = append(missing, Missing{Name: name, Lines: lines})
		}
	}
	sort.Slice(missing, func(i, j int) bool { return missing[i].Name < missing[j].Name })
	return missing
}

// HasTitle says whether the show file's [settings] block sets a title (without one the show is listed as "Untitled show").
func HasTitle(text string) bool { return showfile.ReadSettings(text, false).Title != nil }
