package search

import (
	"strings"
	"testing"
)

func TestLiteralMatchersWithDigits(t *testing.T) {
	factories := []struct {
		name string
		new  func(string) (TextMatcher, error)
	}{
		{"kmp", func(p string) (TextMatcher, error) { return NewKMPMatcher(p) }},
		{"boyer-moore", func(p string) (TextMatcher, error) { return NewBoyerMooreMatcher(p) }},
	}
	for _, factory := range factories {
		t.Run(factory.name, func(t *testing.T) {
			for _, pattern := range []string{"0", "123", "a1B2", "12121", "0001", "0123456789"} {
				matcher, err := factory.new(pattern)
				if err != nil {
					t.Fatal(err)
				}
				for _, text := range []string{"", pattern, "xx" + pattern + "yy", "é" + pattern, "0000001", "1212121", "a1b2", "9876543210"} {
					if got, want := matcher.ContainsMatch(text), strings.Contains(text, pattern); got != want {
						t.Errorf("pattern %q, text %q: got %v, want %v", pattern, text, got, want)
					}
				}
			}
		})
	}
}
