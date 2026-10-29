package automata

import (
	"regexp"
	"testing"

	"github.com/mmal3k/mgrep/internal/parse"
)

func TestExtendedRegexPipeline(t *testing.T) {
	texts := []string{"", "hello!", "abc123", "color", "colour", "cat", "cats", "a.b", "a*b", "a\\b", "123", "é", "中", "🙂", "!", "-", "]", "aaab", "\n", "a\nb"}
	level := []string{""}
	for length := 0; length < 3; length++ {
		var next []string
		for _, prefix := range level {
			for _, letter := range []string{"a", "b", "c", "0", "é"} {
				next = append(next, prefix+letter)
			}
		}
		texts = append(texts, next...)
		level = next
	}
	for _, pattern := range []string{
		"abc123", "hello!", "café", "中", "🙂", "[abc]", "[a-c]+", "[a-cb-d]", "[^abc]", "[^a-c]+", "[^é]", "[é-ê]", "[-a]", "[a-]", "[]a]",
		"colou?r", "a?b?", "a{0}", "a{3}", "a{2,4}", "a{2,}", "(ab){0,2}", "(a|b){1,3}", "(a?){2}",
		"^cat", "cat$", "^cat$", "^a|b$", "(^a|b$)", "^$", "$^", "^^a$$", "(^$)+", "a^", "$a", "a$|ab", "[a-c]|.b", "[^a]a|aa", ".*[^a]", "[a-c]*",
		`a\.b`, `a\*b`, `a\\b`, `\[a\]`, `\^a\$`, `\{2\}`, "[^a-c]?a{0,2}$",
	} {
		t.Run(pattern, func(t *testing.T) {
			tree, err := parse.Parse(pattern)
			if err != nil {
				t.Fatal(err)
			}
			nfa := FromTree(tree)
			dfa := nfa.Determinize()
			minimized := dfa.Minimize()
			reference := regexp.MustCompile(`(?s)\A(?:` + pattern + `)\z`)
			for _, text := range texts {
				want := reference.MatchString(text)
				for _, stage := range []struct {
					name   string
					accept func(string) bool
				}{
					{"NFA", nfa.IsAccepted}, {"DFA", dfa.IsAccepted}, {"minimized DFA", minimized.IsAccepted},
				} {
					if got := stage.accept(text); got != want {
						t.Fatalf("%s on %q: got %v, want %v", stage.name, text, got, want)
					}
				}
			}
		})
	}
}
