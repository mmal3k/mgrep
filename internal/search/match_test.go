package search

import "testing"

func TestContainsMatch(t *testing.T) {
	tests := []struct {
		pattern string
		text    string
		want    bool
	}{
		{"ab", "ab", true},
		{"ab", "abcd", true},
		{"ab", "xxabyy", true},
		{"ab", "aab", true},
		{"aab", "aaab", true},
		{"ab", "xyz", false},
		{"ab", "a", false},
		{"ab", "", false},
		{"a*", "", true},
		{"a*", "xyz", true},
		{"a|bc*", "xxbyy", true},
		{"a|bc*", "xxxx", false},
		{"ab|.c", "xxacyy", true},
		{"ab|.c", "xxécyy", true},
		{"ab|.c", "xxébyy", false},
		{"a.b", "xxa byy", true},
		{"a.b", "xxaébyy", true},
		{"(ab)*c", "xxababczz", true},
		{"(ab)*c", "xxababzz", false},
	}
	for _, tt := range tests {
		t.Run(tt.pattern+"/"+tt.text, func(t *testing.T) {
			matcher, err := NewMatcher(tt.pattern)
			if err != nil {
				t.Fatal(err)
			}
			if got := matcher.ContainsMatch(tt.text); got != tt.want {
				t.Fatalf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewMatcherRejectsInvalidPatterns(t *testing.T) {
	for _, pattern := range []string{"", "(", "a|", "a+", "[a-z]"} {
		if _, err := NewMatcher(pattern); err == nil {
			t.Errorf("should reject %q", pattern)
		}
	}
}

func TestContainsMatchReusesMatcher(t *testing.T) {
	matcher, err := NewMatcher("ab")
	if err != nil {
		t.Fatal(err)
	}
	for _, text := range []string{"xxabyy", "ab", "abcd"} {
		if !matcher.ContainsMatch(text) {
			t.Errorf("should match %q", text)
		}
		if matcher.ContainsMatch("xyz") {
			t.Fatal("state from the previous line must not carry over")
		}
	}
}

func TestNewMatcherMinimizesDFA(t *testing.T) {
	matcher, err := NewMatcher(".*")
	if err != nil {
		t.Fatal(err)
	}
	if len(matcher.dfa.States) != 1 {
		t.Fatalf("got %d states, want one for a pattern accepting every string", len(matcher.dfa.States))
	}
	if !matcher.ContainsMatch("") || !matcher.ContainsMatch("anything") {
		t.Fatal("minimized matcher must preserve acceptance")
	}
}
