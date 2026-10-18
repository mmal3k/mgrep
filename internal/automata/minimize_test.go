package automata

import (
	"maps"
	"slices"
	"testing"

	"github.com/mmal3k/mgrep/internal/parse"
)

func TestMinimize(t *testing.T) {
	tests := []struct {
		pattern string
		states  int
	}{
		{"a", 3},
		{"a|b", 3},
		{"a|bc*", 4},
		{"(a|b)*", 2},
		{"(ab|ac)*", 3},
		{".*", 1},
		{"((a*)*)", 2},
		{"a.b", 5},
		{"ab|.c", 0},
		{".*(ab|.c)", 0},
	}
	for _, tt := range tests {
		t.Run(tt.pattern, func(t *testing.T) {
			tree, err := parse.Parse(tt.pattern)
			if err != nil {
				t.Fatal(err)
			}
			d := FromTree(tree).Determinize()
			before := d.String()
			minimized := d.Minimize()
			if d.String() != before {
				t.Fatal("minimization changed the original DFA")
			}
			if minimized.Start != 0 {
				t.Fatal("start must be zero")
			}
			if tt.states != 0 && len(minimized.States) != tt.states {
				t.Fatalf("got %d states, want %d including the rejecting sink", len(minimized.States), tt.states)
			}
			if word, differs := distinguishingWord(d, minimized, d.Start, minimized.Start); differs {
				t.Fatalf("language changed on %q", word)
			}
			for s1 := range minimized.States {
				for s2 := s1 + 1; s2 < len(minimized.States); s2++ {
					if _, differs := distinguishingWord(minimized, minimized, s1, s2); !differs {
						t.Fatalf("states %d and %d are still equivalent", s1, s2)
					}
				}
			}
		})
	}
}

func TestMinimizeUnreachableStates(t *testing.T) {
	d := &DFA{Automaton: Automaton{
		States: []State{
			{},
			{On: map[rune][]int{'a': {2}}},
			{},
			{Any: []int{3}},
		},
		Start: 1, Accept: []int{0, 2},
	}}
	minimized := d.Minimize()
	if len(minimized.States) != 3 || minimized.Start != 0 {
		t.Fatalf("unexpected minimized automaton:\n%s", minimized)
	}
	if word, differs := distinguishingWord(d, minimized, d.Start, minimized.Start); differs {
		t.Fatalf("language changed on %q", word)
	}
}

func TestMinimizeEmptyLanguage(t *testing.T) {
	d := &DFA{Automaton: Automaton{States: []State{
		{On: map[rune][]int{'a': {0}}},
	}}}
	minimized := d.Minimize()
	if len(minimized.States) != 1 || len(minimized.Accept) != 0 {
		t.Fatalf("unexpected minimized automaton:\n%s", minimized)
	}
}

func TestMinimizeRejectingOverride(t *testing.T) {
	d := &DFA{Automaton: Automaton{
		States: []State{
			{On: map[rune][]int{'a': {1}}, Any: []int{2}},
			{Any: []int{1}},
			{},
		},
		Accept: []int{2},
	}}
	minimized := d.Minimize()
	if minimized.IsAccepted("a") || !minimized.IsAccepted("b") {
		t.Fatal("explicit rejection must override an accepting fallback")
	}
	if word, differs := distinguishingWord(d, minimized, d.Start, minimized.Start); differs {
		t.Fatalf("language changed on %q", word)
	}
}

func distinguishingWord(a, b *DFA, startA, startB int) (string, bool) {
	type pair struct {
		states [2]int
		word   string
	}
	next := func(d *DFA, state int, letter rune) int {
		if state < 0 {
			return -1
		}
		destinations, ok := d.States[state].On[letter]
		if !ok {
			destinations = d.States[state].Any
		}
		if len(destinations) == 0 {
			return -1
		}
		return destinations[0]
	}
	pending := []pair{{states: [2]int{startA, startB}}}
	visited := map[[2]int]bool{pending[0].states: true}
	for current := 0; current < len(pending); current++ {
		p := pending[current]
		if slices.Contains(a.Accept, p.states[0]) != slices.Contains(b.Accept, p.states[1]) {
			return p.word, true
		}
		letters := make(map[rune]bool)
		for i, d := range []*DFA{a, b} {
			if p.states[i] >= 0 {
				for letter := range d.States[p.states[i]].On {
					letters[letter] = true
				}
			}
		}
		var other rune
		for letters[other] {
			other++
		}
		letters[other] = true
		for _, letter := range slices.Sorted(maps.Keys(letters)) {
			states := [2]int{next(a, p.states[0], letter), next(b, p.states[1], letter)}
			if visited[states] {
				continue
			}
			visited[states] = true
			pending = append(pending, pair{states: states, word: p.word + string(letter)})
		}
	}
	return "", false
}
