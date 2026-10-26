// Copyright the open-caption authors
// SPDX-License-Identifier: Apache-2.0

package showfile

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"
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

func TestCharactersKeepPunctuationAndTellLengthsApart(t *testing.T) {
	src := "SOLDIER #1: a\nSOLDIER #2: b\nO'BRIEN: c\nMARY-JANE: d\nALL (EXCEPT BOB): e\n" +
		"ALICE: f\nALICE: g\nALICE: h\nALICES: i\nLICE: j\nALICE. : k\n"
	got := map[string]Character{}
	for _, c := range Parse(src, Options{}).Characters {
		got[c.Name] = c
	}
	for _, n := range []string{"SOLDIER #1", "SOLDIER #2", "O'BRIEN", "MARY-JANE", "ALL (EXCEPT BOB)"} {
		if _, ok := got[n]; !ok {
			t.Errorf("%q missing from %v", n, got)
		}
	}
	if got["SOLDIER #1"].Similar != "" || got["SOLDIER #2"].Similar != "" {
		t.Error("numbered parts are different people, not misspellings of each other")
	}
	if got["ALICES"].Similar != "ALICE" || got["LICE"].Similar != "ALICE" {
		t.Errorf("a letter added or left out was not flagged: %+v %+v", got["ALICES"], got["LICE"])
	}
}

func TestCharactersOfAShowWithManyNames(t *testing.T) {
	var b strings.Builder
	for i := 0; i < 1500; i++ {
		fmt.Fprintf(&b, "PERSON %d: hello\n", i)
	}
	start := time.Now()
	r := Parse(b.String(), Options{})
	if len(r.Characters) != 1500 {
		t.Fatalf("%d characters", len(r.Characters))
	}
	if d := time.Since(start); d > 5*time.Second {
		t.Errorf("1500 names took %v: the misspelling check is too slow for the editor's keystroke-by-keystroke analysis", d)
	}
}
