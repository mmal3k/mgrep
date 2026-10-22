package automata

import (
	"regexp"
	"testing"

	"github.com/mmal3k/mgrep/internal/parse"
)

func TestPlusAcceptance(t *testing.T) {
	texts := []string{""}
	level := []string{""}
	for length := 0; length < 5; length++ {
		var next []string
		for _, prefix := range level {
			for _, letter := range []string{"a", "b", "c"} {
				next = append(next, prefix+letter)
			}
		}
		texts = append(texts, next...)
		level = next
	}
	for _, pattern := range []string{"a+", "(ab)+", "(a|b)+", "a+b", "a|b+c", ".+", "(a*)+", "(a+)*", "(ab)+ab", "a+a+"} {
		t.Run(pattern, func(t *testing.T) {
			tree, err := parse.Parse(pattern)
			if err != nil {
				t.Fatal(err)
			}
			nfa := FromTree(tree)
			dfa := nfa.Determinize()
			minimized := dfa.Minimize()
			reference := regexp.MustCompile("^(" + pattern + ")$")
			for _, text := range texts {
				want := reference.MatchString(text)
				for _, stage := range []struct {
					name   string
					accept func(string) bool
				}{
					{"NFA", nfa.IsAccepted},
					{"DFA", dfa.IsAccepted},
					{"minimized DFA", minimized.IsAccepted},
				} {
					if got := stage.accept(text); got != want {
						t.Errorf("%s, text %q: got %v, want %v", stage.name, text, got, want)
					}
				}
			}
		})
	}
}
