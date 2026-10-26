// Copyright the open-caption authors
// SPDX-License-Identifier: Apache-2.0

package showfile

import (
	"reflect"
	"testing"
)

func TestCharactersListEveryNameWithItsLines(t *testing.T) {
	src := "BARRY: Hello there\nDEE DEE & BARRY: Hi\nBARRY: Again\n[together]\nTRENT: one\nDEE DEE: two\n[/together]\nBARRY: Bye || TRENT: Bye\n"
	got := Parse(src, Options{}).Characters
	names := []string{}
	for _, c := range got {
		names = append(names, c.Name)
	}
	if want := []string{"BARRY", "DEE DEE", "TRENT"}; !reflect.DeepEqual(names, want) {
		t.Fatalf("names = %v, want %v (most lines first)", names, want)
	}
	if want := []int{1, 2, 3, 8}; !reflect.DeepEqual(got[0].Lines, want) {
		t.Errorf("BARRY lines = %v, want %v (a group label counts for each of its names)", got[0].Lines, want)
	}
	if !got[0].Label || !got[1].Label {
		t.Errorf("BARRY and DEE DEE are also labels on their own: %+v", got)
	}
}

func TestACharacterNameOneLetterOffIsFlagged(t *testing.T) {
	src := "BARRY: a\nBARRY: b\nBARRY: c\nBARY: d\nTRENT: e\nTRENT: f\nTRNET: g\nANN: h\nANN: i\nAMY: j\n"
	similar := map[string]string{}
	for _, c := range Parse(src, Options{}).Characters {
		similar[c.Name] = c.Similar
	}
	if similar["BARY"] != "BARRY" || similar["TRNET"] != "TRENT" {
		t.Errorf("misspellings not flagged: %v", similar)
	}
	if similar["BARRY"] != "" || similar["AMY"] != "" || similar["ANN"] != "" {
		t.Errorf("correct or short names flagged: %v", similar)
	}
}
