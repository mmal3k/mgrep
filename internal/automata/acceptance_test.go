package automata

import (
	"testing"

	"gutenberg-backend/internal/parse"
)

func TestIsAccepted(t *testing.T) {
	tests := []struct {
		pattern string
		text    string
		want    bool
	}{
		{"a|b", "a", true},
		{"ab", "ab", true},
		{"ab", "a", false},
		{"a", "ab", false},
		{"a*", "", true},
		{"a", "", false},
		{"((a*)*)", "aaa", true},
		{"((a*)*)", "b", false},
		{"(a|b)*c", "abbac", true},
		{"(a|b)*c", "abba", false},
		{".", "", false},
		{".", "é", true},
		{".", "\n", true},
		{"ab|.c", "ac", true},
		{"ab|.c", "ab", true},
		{"ab|.c", "éb", false},
		{"ab|.c", "éc", true},
	}

	for _, tt := range tests {
		t.Run(tt.pattern+"/"+tt.text, func(t *testing.T) {
			tree, err := parse.Parse(tt.pattern)
			if err != nil {
				t.Fatal(err)
			}
			n := FromTree(tree)
			if got := n.IsAccepted(tt.text); got != tt.want {
				t.Errorf("NFA: got %v, want %v", got, tt.want)
			}
			if got := n.Determinize().IsAccepted(tt.text); got != tt.want {
				t.Errorf("DFA: got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsAcceptedWildcardSemantics(t *testing.T) {
	a := Automaton{
		States: []State{
			{On: map[rune][]int{'a': {1}}, Any: []int{2}},
			{},
			{},
		},
		Start:  0,
		Accept: []int{2},
	}
	n := &NFA{Automaton: a}
	d := &DFA{Automaton: a}
	if !n.IsAccepted("a") {
		t.Fatal("NFA must also follow the wildcard on an explicit letter")
	}
	if d.IsAccepted("a") {
		t.Fatal("DFA must use the explicit transition instead of the fallback")
	}
	if !d.IsAccepted("b") {
		t.Fatal("DFA must use the fallback when the letter has no transition")
	}
}
