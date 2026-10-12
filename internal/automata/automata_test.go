package automata

import (
	"testing"

	"github.com/mmal3k/mgrep/internal/parse"
)

func TestString(t *testing.T) {
	n := &NFA{
		Automaton: Automaton{
			States: []State{
				{On: map[rune][]int{'b': {2}, 'a': {1}}, Any: []int{2}, Epsilon: []int{1}},
				{On: map[rune][]int{'a': {1}}},
				{},
			},
			Start:  0,
			Accept: []int{2},
		},
	}
	want := "Start: 0\nAccept: [2]\n\n0 --a--> 1\n0 --b--> 2\n0 --ANY--> 2\n0 --ε--> 1\n1 --a--> 1\n"
	if got := n.String(); got != want {
		t.Fatalf("got:\n%s\nwant:\n%s", got, want)
	}
}

func TestFromTree(t *testing.T) {
	tests := []struct {
		pattern string
		accept  []string
		reject  []string
	}{
		{"a.b", []string{"acb", "aZb", "aab"}, []string{"ab", "acbb"}},
		{"a|b", []string{"a", "b"}, []string{"", "ab"}},
		{"(a|b)*c", []string{"c", "abbac"}, []string{"", "ab"}},
		{"a|bc*", []string{"a", "b", "bc", "bccc"}, []string{"", "ab", "c"}},
		{"a*", []string{"", "a", "aaa"}, []string{"b", "ab"}},
	}
	for _, tt := range tests {
		t.Run(tt.pattern, func(t *testing.T) {
			tree, err := parse.Parse(tt.pattern)
			if err != nil {
				t.Fatal(err)
			}
			n := FromTree(tree)
			if n.Start != 0 || len(n.Accept) != 1 || n.Accept[0] != len(n.States)-1 {
				t.Fatalf("got start %d and accept %v; want 0 and %d", n.Start, n.Accept, len(n.States)-1)
			}
			for _, input := range tt.accept {
				if !accepts(n, input) {
					t.Errorf("should accept %q", input)
				}
			}
			for _, input := range tt.reject {
				if accepts(n, input) {
					t.Errorf("should reject %q", input)
				}
			}
		})
	}
}

func accepts(n *NFA, input string) bool {
	closure := func(seeds []int) map[int]bool {
		visited := make(map[int]bool)
		pending := append([]int(nil), seeds...)
		for len(pending) > 0 {
			id := pending[len(pending)-1]
			pending = pending[:len(pending)-1]
			if visited[id] {
				continue
			}
			visited[id] = true
			pending = append(pending, n.States[id].Epsilon...)
		}
		return visited
	}
	current := closure([]int{n.Start})
	for _, character := range input {
		var next []int
		for id := range current {
			next = append(next, n.States[id].On[character]...)
			next = append(next, n.States[id].Any...)
		}
		current = closure(next)
	}
	for _, state := range n.Accept {
		if current[state] {
			return true
		}
	}
	return false
}
