package search

import (
	"reflect"
	"strings"
	"testing"
)

func TestBoyerMooreExhaustive(t *testing.T) {
	words := func(maxLength int) []string {
		result := []string{""}
		level := []string{""}
		for length := 0; length < maxLength; length++ {
			var next []string
			for _, prefix := range level {
				next = append(next, prefix+"a", prefix+"b")
			}
			result = append(result, next...)
			level = next
		}
		return result
	}
	texts := words(9)
	for _, pattern := range words(6)[1:] {
		matcher, err := NewBoyerMooreMatcher(pattern)
		if err != nil {
			t.Fatal(err)
		}
		for _, text := range texts {
			if got, want := matcher.Index(text), strings.Index(text, pattern); got != want {
				t.Fatalf("pattern %q, text %q: got index %d, want %d", pattern, text, got, want)
			}
			if got, want := matcher.ContainsMatch(text), strings.Contains(text, pattern); got != want {
				t.Fatalf("pattern %q, text %q: got %v, want %v", pattern, text, got, want)
			}
		}
	}
}

func TestBoyerMooreMatcherReusesTables(t *testing.T) {
	for _, pattern := range []string{"a", "abab", "ABCDABD", "ANPANMAN", "aaaaab", "Sargon"} {
		matcher, err := NewBoyerMooreMatcher(pattern)
		if err != nil {
			t.Fatal(err)
		}
		last := matcher.last
		suffix := append([]int(nil), matcher.goodSuffix...)
		for _, text := range []string{"", pattern, "xx" + pattern, pattern + "yy", "é🙂" + pattern, "\x00\xff" + pattern, "aaaaaaaaaaaa", "ABCDABCDABD", "sargon", "ANPANPANMAN"} {
			if got, want := matcher.ContainsMatch(text), strings.Contains(text, pattern); got != want {
				t.Errorf("pattern %q, text %q: got %v, want %v", pattern, text, got, want)
			}
		}
		if matcher.last != last || !reflect.DeepEqual(matcher.goodSuffix, suffix) {
			t.Fatal("search changed the shift tables")
		}
	}
}

func TestBoyerMooreMatcherRejectsNonliteralPatterns(t *testing.T) {
	for _, pattern := range []string{"", "a|b", "a*", "a.b", "(ab)", "a+", "[a-z]", "é", "a b", "١", "１"} {
		if _, err := NewBoyerMooreMatcher(pattern); err == nil {
			t.Errorf("should reject %q", pattern)
		}
	}
}

func BenchmarkLiteralSearch(b *testing.B) {
	factories := []struct {
		name string
		new  func(string) (TextMatcher, error)
	}{
		{"dfa", func(p string) (TextMatcher, error) { return NewMatcher(p) }},
		{"kmp", func(p string) (TextMatcher, error) { return NewKMPMatcher(p) }},
		{"boyer-moore", func(p string) (TextMatcher, error) { return NewBoyerMooreMatcher(p) }},
	}
	for _, sample := range []struct {
		name    string
		pattern string
		text    string
		want    bool
	}{
		{"late-match", "Sargon", strings.Repeat("Babylon was an ancient city. ", 256) + "Sargon", true},
		{"absent", "Sargon", strings.Repeat("Babylon was an ancient city. ", 256), false},
		{"repeated-suffix", "baaaaaa", strings.Repeat("a", 8192), false},
	} {
		for _, factory := range factories {
			b.Run(sample.name+"/"+factory.name, func(b *testing.B) {
				matcher, err := factory.new(sample.pattern)
				if err != nil {
					b.Fatal(err)
				}
				b.SetBytes(int64(len(sample.text)))
				b.ReportAllocs()
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					if matcher.ContainsMatch(sample.text) != sample.want {
						b.Fatal("incorrect match result")
					}
				}
			})
		}
	}
}
