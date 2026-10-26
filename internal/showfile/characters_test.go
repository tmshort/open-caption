// Copyright the open-caption authors
// SPDX-License-Identifier: Apache-2.0

package showfile

import (
	"reflect"
	"testing"
)

func TestCharactersListEveryNameWithItsLines(t *testing.T) {
	src := "ALICE: Hello there\nALICE & BOB: Hi\nALICE: Again\n[together]\nCAROL: one\nBOB: two\n[/together]\nALICE: Bye || CAROL: Bye\n"
	got := Parse(src, Options{}).Characters
	names := []string{}
	for _, c := range got {
		names = append(names, c.Name)
	}
	if want := []string{"ALICE", "BOB", "CAROL"}; !reflect.DeepEqual(names, want) {
		t.Fatalf("names = %v, want %v (most lines first)", names, want)
	}
	if want := []int{1, 2, 3, 8}; !reflect.DeepEqual(got[0].Lines, want) {
		t.Errorf("ALICE lines = %v, want %v (a group label counts for each of its names)", got[0].Lines, want)
	}
	if !got[0].Label || !got[1].Label {
		t.Errorf("ALICE and BOB are also labels on their own: %+v", got)
	}
}

func TestACharacterNameOneLetterOffIsFlagged(t *testing.T) {
	src := "ALICE: a\nALICE: b\nALICE: c\nALCE: d\nCAROL: e\nCAROL: f\nCARLO: g\nANN: h\nANN: i\nAMY: j\n"
	similar := map[string]string{}
	for _, c := range Parse(src, Options{}).Characters {
		similar[c.Name] = c.Similar
	}
	if similar["ALCE"] != "ALICE" || similar["CARLO"] != "CAROL" {
		t.Errorf("misspellings not flagged: %v", similar)
	}
	if similar["ALICE"] != "" || similar["AMY"] != "" || similar["ANN"] != "" {
		t.Errorf("correct or short names flagged: %v", similar)
	}
}
