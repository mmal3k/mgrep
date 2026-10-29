package search

import (
	"reflect"
	"regexp"
	"testing"
)

func TestExtendedSearchAndHighlighting(t *testing.T) {
	texts := []string{"", "abc123 hello!", "color colour colouur", "cat", "cat sleeps", "the cat", "cats", "a.b a*b [a]", "ééaa🙂", "ababa", "123456", "aaabaaa", "!", "-", "]", "Sargon and Sagon", "Sargon123!"}
	for _, pattern := range []string{
		"abc123", "hello!", "[abc]", "[a-z]+", "[A-Za-z][A-Za-z0-9]*", "[^abc]+", "[^a-z]", "[é🙂]", "[^é]", "[0-9]{2,4}",
		"colou?r", "a{0}", "a{2,4}", "a{2,}", "(ab){1,2}", "^cat", "cat$", "^cat$", "^$", "$^", "^a|b$", "a^", "$a", "a$|ab", ".*[^a]", "^S[a-z]+[0-9]*!$",
		`a\.b`, `a\*b`, `\[a\]`,
	} {
		t.Run(pattern, func(t *testing.T) {
			matcher, err := NewMatcher(pattern)
			if err != nil {
				t.Fatal(err)
			}
			highlighter, err := NewHighlighter(pattern, "dfa")
			if err != nil {
				t.Fatal(err)
			}
			reference := regexp.MustCompile(pattern)
			reference.Longest()
			for _, text := range texts {
				if got, want := matcher.ContainsMatch(text), reference.MatchString(text); got != want {
					t.Fatalf("text %q: match %v, want %v", text, got, want)
				}
				var want []MatchSpan
				for _, match := range reference.FindAllStringIndex(text, -1) {
					if match[0] != match[1] {
						want = append(want, MatchSpan{Start: match[0], End: match[1]})
					}
				}
				if got := highlighter.FindMatches(text); !reflect.DeepEqual(got, want) {
					t.Fatalf("text %q: spans %v, want %v", text, got, want)
				}
			}
		})
	}
}
