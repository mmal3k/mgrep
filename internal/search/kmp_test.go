package search

import (
	"reflect"
	"strings"
	"testing"
)

func TestFirstStep(t *testing.T) {
	want := []int{-1, 0, 0, 1}
	if got := first_step("abab"); !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestKMP(t *testing.T) {
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
	for _, pattern := range words(5) {
		for _, text := range words(7) {
			if got, want := kmp(text, pattern), strings.Index(text, pattern); got != want {
				t.Fatalf("text %q, pattern %q: got %d, want %d", text, pattern, got, want)
			}
		}
	}
}

func TestKMPMatcherReusesTable(t *testing.T) {
	matcher, err := NewKMPMatcher("abab")
	if err != nil {
		t.Fatal(err)
	}
	before := append([]int(nil), matcher.co...)
	for _, text := range []string{"xxababyy", "", "aaaabab", "éabab", "abc"} {
		if got, want := matcher.ContainsMatch(text), strings.Contains(text, "abab"); got != want {
			t.Errorf("text %q: got %v, want %v", text, got, want)
		}
	}
	if !reflect.DeepEqual(matcher.co, before) {
		t.Fatal("search changed the fallback table")
	}
}

func TestKMPMatcherRejectsNonliteralPatterns(t *testing.T) {
	for _, pattern := range []string{"", "a|b", "a*", "a.b", "(ab)", "a+", "[a-z]"} {
		if _, err := NewKMPMatcher(pattern); err == nil {
			t.Errorf("should reject %q", pattern)
		}
	}
}
