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
	tree, err := parse.Parse(pattern)
	if err != nil {
		return nil, err
	}
	tree = &parse.RegExTree{Root: parse.CONCAT, SubTrees: []*parse.RegExTree{
		{Root: parse.STAR, SubTrees: []*parse.RegExTree{{Root: parse.DOT}}}, tree,
	}}
	dfa := automata.FromTree(tree).Determinize().Minimize()
	accepting := make([]bool, len(dfa.States))
	for _, state := range dfa.Accept {
		accepting[state] = true
	}

	return &Matcher{dfa: dfa, accepting: accepting}, nil
}

func (m *Matcher) ContainsMatch(text string) bool {
	boundary := automata.StartOfLine
	if text == "" {
		boundary = automata.EmptyLine
	}
	current := m.dfa.Next(m.dfa.Start, boundary)
	if current >= 0 && m.accepting[current] {
		return true
	}
	for _, letter := range text {
		current = m.dfa.Next(current, letter)
		if current < 0 {
			return false
		}
		if m.accepting[current] {
			return true
		}
	}
	current = m.dfa.Next(current, automata.EndOfLine)
	return current >= 0 && m.accepting[current]
}
