package search

import (
	"github.com/mmal3k/mgrep/internal/automata"
	"github.com/mmal3k/mgrep/internal/parse"
)

type Matcher struct {
	dfa       *automata.DFA
	accepting []bool
}

func NewMatcher(pattern string) (*Matcher, error) {
	tree, err := parse.Parse(".*(" + pattern + ")")
	if err != nil {
		return nil, err
	}
	dfa := automata.FromTree(tree).Determinize()
	accepting := make([]bool, len(dfa.States))
	for _, state := range dfa.Accept {
		accepting[state] = true
	}

	return &Matcher{dfa: dfa, accepting: accepting}, nil
}

func (m *Matcher) ContainsMatch(text string) bool {
	current := m.dfa.Start

	if m.accepting[current] == true {
		return true
	}

	for _, letter := range text {
		next, ok := m.dfa.States[current].On[letter]

		if !ok {
			next = m.dfa.States[current].Any
		}

		if len(next) == 0 {
			return false
		}

		current = next[0]

		if m.accepting[current] == true {
			return true
		}
	}
	return false
}
