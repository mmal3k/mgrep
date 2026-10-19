package search

import "fmt"

type KMPMatcher struct {
	pattern string
	co      []int
}

func NewKMPMatcher(pattern string) (*KMPMatcher, error) {
	if len(pattern) == 0 {
		return nil, fmt.Errorf("KMP requires a nonempty literal pattern")
	}
	for _, letter := range pattern {
		if !('a' <= letter && letter <= 'z' || 'A' <= letter && letter <= 'Z' || '0' <= letter && letter <= '9') {
			return nil, fmt.Errorf("KMP supports only literal ASCII letters and digits")
		}
	}
	return &KMPMatcher{pattern: pattern, co: construct_co(pattern)}, nil
}

func (m *KMPMatcher) ContainsMatch(text string) bool {
	return kmpWithTable(text, m.pattern, m.co) >= 0
}

func first_step(f string) []int {
	if len(f) == 0 {
		return []int{}
	}

	co := make([]int, len(f))

	co[0] = -1

	for i := 1; i < len(f); i++ {
		for k := i - 1; k >= 0; k-- {
			if f[:k] == f[i-k:i] {
				co[i] = k
				break
			}
			co[i] = 0
		}
	}

	return co
}

func second_step(f string, co []int) []int {
	for i := range len(f) {
		if co[i] == -1 {
			continue
		}
		if f[i] == f[co[i]] && co[co[i]] == -1 {
			co[i] = -1
		}
	}

	return co
}

func third_step(f string, co []int) []int {
	for i := range len(f) {
		if co[i] == -1 {
			continue
		}
		if f[i] == f[co[i]] {
			co[i] = co[co[i]]
		}
	}
	return co
}

func construct_co(f string) []int {
	co := first_step(f)
	co = second_step(f, co)
	co = third_step(f, co)
	return co
}

func kmp(t, f string) int {
	return kmpWithTable(t, f, construct_co(f))
}

func kmpWithTable(t, f string, co []int) int {
	if len(f) == 0 {
		return 0
	}
	i, j := 0, 0
	n, m := len(t), len(f)

	for i < n {
		if j == -1 || t[i] == f[j] {
			i++
			j++

			if j == m {
				return i - j
			}
		} else {
			j = co[j]
		}
	}
	return -1
}
