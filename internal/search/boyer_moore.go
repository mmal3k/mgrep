package search

import "fmt"

type BoyerMooreMatcher struct {
	pattern    string
	last       [256]int
	goodSuffix []int
}

func NewBoyerMooreMatcher(pattern string) (*BoyerMooreMatcher, error) {
	if len(pattern) == 0 {
		return nil, fmt.Errorf("Boyer-Moore requires a nonempty literal pattern")
	}
	for _, letter := range pattern {
		if !('a' <= letter && letter <= 'z' || 'A' <= letter && letter <= 'Z') {
			return nil, fmt.Errorf("Boyer-Moore supports only literal ASCII-letter patterns")
		}
	}
	return &BoyerMooreMatcher{
		pattern:    pattern,
		last:       badCharacterTable(pattern),
		goodSuffix: goodSuffixTable(pattern),
	}, nil
}

func badCharacterTable(pattern string) [256]int {
	var last [256]int
	for i := range last {
		last[i] = -1
	}
	for i := range len(pattern) {
		last[pattern[i]] = i
	}
	return last
}

func goodSuffixTable(pattern string) []int {
	n := len(pattern)
	shift := make([]int, n+1)
	border := make([]int, n+1)
	i, j := n, n+1
	border[i] = j
	for i > 0 {
		for j <= n && pattern[i-1] != pattern[j-1] {
			if shift[j] == 0 {
				shift[j] = j - i
			}
			j = border[j]
		}
		i--
		j--
		border[i] = j
	}
	j = border[0]
	for i := range shift {
		if shift[i] == 0 {
			shift[i] = j
		}
		if i == j {
			j = border[j]
		}
	}
	return shift
}

func (m *BoyerMooreMatcher) ContainsMatch(text string) bool {
	n := len(m.pattern)
	for offset := 0; offset <= len(text)-n; {
		j := n - 1
		for j >= 0 && m.pattern[j] == text[offset+j] {
			j--
		}
		if j < 0 {
			return true
		}
		offset += max(1, j-m.last[text[offset+j]], m.goodSuffix[j+1])
	}
	return false
}
