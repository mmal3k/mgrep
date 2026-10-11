package automata

import (
	"fmt"
	"slices"
	"testing"

	"github.com/mmal3k/regex-cli/internal/parse"
)

func TestDeterminizePreservesLanguage(t *testing.T) {
	patterns := []string{"a", "a|bc*", "a*", "(a|b)*c", "(ab|ac)*", "a.b", "(a|.)b", "(a|.)*", "((a*)*)", ".*a", "ab|.c"}
	inputs := []string{""}
	level := []string{""}
	for length := 1; length <= 4; length++ {
		var next []string
		for _, prefix := range level {
			for _, letter := range []rune{'a', 'b', 'c', 'é', '\n'} {
				next = append(next, prefix+string(letter))
			}
		}
		inputs = append(inputs, next...)
		level = next
	}
	for _, pattern := range patterns {
		t.Run(pattern, func(t *testing.T) {
			tree, err := parse.Parse(pattern)
			if err != nil {
				t.Fatal(err)
			}
			n := FromTree(tree)
			before := n.String()
			d := n.Determinize()
			if n.String() != before {
				t.Fatal("conversion changed the original NFA")
			}
			if d.Start != 0 {
				t.Fatal("start state must be zero")
			}
			for _, state := range d.States {
				if len(state.Epsilon) != 0 {
					t.Fatal("epsilon transition remains")
				}
				if len(state.Any) > 1 {
					t.Fatal("multiple fallback destinations")
				}
				for _, destinations := range state.On {
					if len(destinations) != 1 {
						t.Fatal("letter transition must have one destination")
					}
				}
			}
			for _, input := range inputs {
				want := accepts(n, input)
				if got := acceptsDFA(d, input); got != want {
					t.Fatalf("DFA on %q: got %v, want %v", input, got, want)
				}
			}
			if d.String() != n.Determinize().String() {
				t.Fatal("DFA numbering is unstable")
			}
		})
	}
}

func TestDeterminizeBookExample(t *testing.T) {
	tree, err := parse.Parse("a|bc*")
	if err != nil {
		t.Fatal(err)
	}

	nfa := FromTree(tree)
	fmt.Println(nfa)

	dfa := nfa.Determinize()
	fmt.Println(dfa)

	want := "Start: 0\nAccept: [1 2 3]\n\n0 --a--> 1\n0 --b--> 2\n2 --c--> 3\n3 --c--> 3\n"
	if got := dfa.String(); got != want {
		t.Fatalf("got:\n%s\nwant:\n%s", got, want)
	}
}

func TestDFAStringFallback(t *testing.T) {
	tree, err := parse.Parse(".")
	if err != nil {
		t.Fatal(err)
	}

	nfa := FromTree(tree)
	fmt.Println(nfa)

	dfa := nfa.Determinize()
	fmt.Println(dfa)
	want := "Start: 0\nAccept: [1]\n\n0 --OTHER--> 1\n"
	if got := dfa.String(); got != want {
		t.Fatalf("got:\n%s\nwant:\n%s", got, want)
	}
}

func acceptsDFA(d *DFA, input string) bool {
	current := d.Start
	for _, letter := range input {
		next := d.States[current].On[letter]
		if len(next) == 0 {
			next = d.States[current].Any
		}
		if len(next) == 0 {
			return false
		}
		current = next[0]
	}
	return slices.Contains(d.Accept, current)
}
