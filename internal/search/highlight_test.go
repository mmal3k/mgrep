package search

import (
	"reflect"
	"regexp"
	"strings"
	"testing"
)

func TestHighlightMatches(t *testing.T) {
	for _, tt := range []struct {
		pattern   string
		text      string
		algorithm string
		want      string
	}{
		{"Sargon", "Sargon and Sargonids", "kmp", "\x1b[1;31mSargon\x1b[0m and \x1b[1;31mSargon\x1b[0mids"},
		{"S(a|g|r)+on", "é Sargon and Sagon", "dfa", "é \x1b[1;31mSargon\x1b[0m and \x1b[1;31mSagon\x1b[0m"},
		{"a.b", "x aéb y", "dfa", "x \x1b[1;31maéb\x1b[0m y"},
		{"a|aa", "aaa", "dfa", "\x1b[1;31maa\x1b[0m\x1b[1;31ma\x1b[0m"},
		{"a*", "bbb", "dfa", "bbb"},
		{"a*", "baaab", "dfa", "b\x1b[1;31maaa\x1b[0mb"},
		{"(a*)+", "", "dfa", ""},
		{"12", "1212", "kmp", "\x1b[1;31m12\x1b[0m\x1b[1;31m12\x1b[0m"},
	} {
		h, err := NewHighlighter(tt.pattern, tt.algorithm)
		if err != nil {
			t.Fatal(err)
		}
		if got := h.Highlight(tt.text); got != tt.want {
			t.Errorf("pattern %q, text %q: got %q, want %q", tt.pattern, tt.text, got, tt.want)
		}
	}
}

func TestHighlightRegexSpans(t *testing.T) {
	texts := []string{""}
	level := []string{""}
	for i := 0; i < 5; i++ {
		var next []string
		for _, prefix := range level {
			next = append(next, prefix+"a", prefix+"b")
		}
		texts = append(texts, next...)
		level = next
	}
	for _, pattern := range []string{"a", "a+", "ab|a", "(ab)*", "a.*b", ".*", "(a|b)+", "a*b*", "a.b", "(a*)+"} {
		h, err := NewHighlighter(pattern, "dfa")
		if err != nil {
			t.Fatal(err)
		}
		reference := regexp.MustCompile(pattern)
		reference.Longest()
		for _, text := range texts {
			var want []MatchSpan
			for _, index := range reference.FindAllStringIndex(text, -1) {
				if index[0] < index[1] {
					want = append(want, MatchSpan{Start: index[0], End: index[1]})
				}
			}
			if got := h.FindMatches(text); !reflect.DeepEqual(got, want) {
				t.Fatalf("pattern %q, text %q: got %v, want %v", pattern, text, got, want)
			}
		}
	}
}

func TestHighlightLongLine(t *testing.T) {
	h, err := NewHighlighter("a.*b", "dfa")
	if err != nil {
		t.Fatal(err)
	}
	text := "a" + strings.Repeat("x", 100000) + "b"
	if got := h.FindMatches(text); !reflect.DeepEqual(got, []MatchSpan{{Start: 0, End: len(text)}}) {
		t.Fatalf("got %v", got)
	}
}

func TestLiteralHighlightPositionsByAlgorithm(t *testing.T) {
	for _, algorithm := range []string{"kmp", "boyer-moore"} {
		for _, tt := range []struct {
			pattern string
			text    string
		}{
			{"Sargon", "Sargon met Sargon and Sargonids"},
			{"abab", "ababababab"},
			{"123", "é123 123123"},
			{"a", "aaa"},
			{"Sargon", ""},
			{"Sargon", "no matches"},
		} {
			h, err := NewHighlighter(tt.pattern, algorithm)
			if err != nil {
				t.Fatal(err)
			}
			var want []MatchSpan
			for offset := 0; offset < len(tt.text); {
				index := strings.Index(tt.text[offset:], tt.pattern)
				if index < 0 {
					break
				}
				start := offset + index
				offset = start + len(tt.pattern)
				want = append(want, MatchSpan{Start: start, End: offset})
			}
			if got := h.FindMatches(tt.text); !reflect.DeepEqual(got, want) {
				t.Errorf("%s, pattern %q, text %q: got %v, want %v", algorithm, tt.pattern, tt.text, got, want)
			}
		}
	}
}
