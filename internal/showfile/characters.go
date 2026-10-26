// Copyright the open-caption authors
// SPDX-License-Identifier: Apache-2.0

package showfile

import (
	"sort"
	"strings"
)

// Character is one name that speaks or sings in a show file, for Setup's list of characters (the editor's way to
// spot a misspelled name). A group label such as "ALICE & BOB" counts for each of its names.
type Character struct {
	Name string `json:"name"`
	// Lines are the line numbers (ascending, once each) of the lines this name speaks on.
	Lines []int `json:"lines"`
	// Label is true when the name is also used as a speaker label on its own (so it has its own colour).
	Label bool `json:"label,omitempty"`
	// Similar is a name used on more lines that differs from this one by a single letter, which suggests this one is a
	// misspelling of it ("" when there is none).
	Similar string `json:"similar,omitempty"`
}

// namesOf splits a speaker label into the names of a group ("ALICE, BOB & CAROL" is three).
func namesOf(label string) []string {
	var out []string
	for _, n := range strings.FieldsFunc(label, func(r rune) bool { return r == ',' || r == '&' || r == '/' }) {
		if n = strings.TrimSpace(n); n != "" {
			out = append(out, n)
		}
	}
	if len(out) == 0 {
		return []string{label}
	}
	return out
}

// charactersOf builds the character list from the lines each speaker label was used on (labels in order of first
// appearance). The most used names come first; names with the same number of lines keep their order of appearance.
func charactersOf(order []string, lines map[string][]int) []Character {
	var out []Character
	index := map[string]int{}
	for _, label := range order {
		names := namesOf(label)
		for _, n := range names {
			i, ok := index[n]
			if !ok {
				i = len(out)
				index[n] = i
				out = append(out, Character{Name: n})
			}
			if len(names) == 1 {
				out[i].Label = true
			}
			out[i].Lines = append(out[i].Lines, lines[label]...)
		}
	}
	for i := range out {
		ls := out[i].Lines
		sort.Ints(ls)
		w := 0
		for j, l := range ls {
			if j == 0 || l != ls[j-1] {
				ls[w] = l
				w++
			}
		}
		out[i].Lines = ls[:w]
	}
	for i := range out {
		best := -1
		for j := range out {
			if j != i && len(out[j].Lines) > len(out[i].Lines) && oneLetterApart(out[i].Name, out[j].Name) &&
				(best < 0 || len(out[j].Lines) > len(out[best].Lines)) {
				best = j
			}
		}
		if best >= 0 {
			out[i].Similar = out[best].Name
		}
	}
	sort.SliceStable(out, func(a, b int) bool { return len(out[a].Lines) > len(out[b].Lines) })
	return out
}

// oneLetterApart reports whether two names differ by one letter added, left out, changed, or two neighbours swapped
// (capitals and lower case count as the same). Very short names are never similar: "ANN" and "AMY" are different people.
func oneLetterApart(a, b string) bool {
	x, y := []rune(strings.ToUpper(a)), []rune(strings.ToUpper(b))
	if len(x) < 4 || len(y) < 4 || string(x) == string(y) {
		return false
	}
	if len(x) < len(y) {
		x, y = y, x
	}
	switch len(x) - len(y) {
	case 0:
		var at []int
		for i := range x {
			if x[i] != y[i] {
				at = append(at, i)
			}
		}
		return len(at) == 1 || (len(at) == 2 && at[1] == at[0]+1 && x[at[0]] == y[at[1]] && x[at[1]] == y[at[0]])
	case 1:
		i := 0
		for i < len(y) && x[i] == y[i] {
			i++
		}
		return string(x[i+1:]) == string(y[i:])
	}
	return false
}
