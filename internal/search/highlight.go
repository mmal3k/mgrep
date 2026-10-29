package search

import (
	"fmt"
	"strings"

	"github.com/mmal3k/mgrep/internal/automata"
	"github.com/mmal3k/mgrep/internal/parse"
)

type MatchSpan struct {
	Start int
	End   int
}

type Highlighter struct {
	dfa       *automata.DFA
	accepting []bool
	pattern   string
	findIndex func(string) int
}

func NewHighlighter(pattern string, algorithm string) (*Highlighter, error) {
	switch algorithm {
	case "boyer-moore":
		matcher, err := NewBoyerMooreMatcher(pattern)
		if err != nil {
			return nil, err
		}
		return &Highlighter{pattern: pattern, findIndex: matcher.Index}, nil
	case "kmp":
		matcher, err := NewKMPMatcher(pattern)
		if err != nil {
			return nil, err
		}
		return &Highlighter{pattern: pattern, findIndex: func(text string) int {
			return kmpWithTable(text, matcher.pattern, matcher.co)
		}}, nil
	case "dfa":
	default:
		return nil, fmt.Errorf("unknown highlighting algorithm %q", algorithm)
	}
	tree, err := parse.Parse(pattern)
	if err != nil {
		return nil, err
	}
	dfa := automata.FromTree(tree).Determinize().Minimize()
	accepting := make([]bool, len(dfa.States))
	for _, state := range dfa.Accept {
		accepting[state] = true
	}
	return &Highlighter{dfa: dfa, accepting: accepting}, nil
}

func (h *Highlighter) FindMatches(text string) []MatchSpan {
	if h.findIndex != nil {
		var spans []MatchSpan
		for offset := 0; offset < len(text); {
			index := h.findIndex(text[offset:])
			if index < 0 {
				break
			}
			start := offset + index
			offset = start + len(h.pattern)
			spans = append(spans, MatchSpan{Start: start, End: offset})
		}
		return spans
	}
	letters := []rune(text)
	positions := make([]int, 0, len(letters)+1)
	for position := range text {
		positions = append(positions, position)
	}
	positions = append(positions, len(text))
	ends := make([]int, len(letters))
	next := make([]int, len(h.dfa.States))
	current := make([]int, len(next))
	for state := range next {
		next[state] = -1
		endState := h.dfa.Next(state, automata.EndOfLine)
		if endState >= 0 && h.accepting[endState] {
			next[state] = len(letters)
		}
	}
	for i := len(letters) - 1; i >= 0; i-- {
		for state := range h.dfa.States {
			current[state] = -1
			if h.accepting[state] {
				current[state] = i
			}
			destination := h.dfa.Next(state, letters[i])
			if destination >= 0 {
				current[state] = max(current[state], next[destination])
			}
		}
		startState := h.dfa.Start
		if i == 0 {
			startState = h.dfa.Next(startState, automata.StartOfLine)
		}
		ends[i] = -1
		if startState >= 0 {
			ends[i] = current[startState]
		}
		next, current = current, next
	}
	var spans []MatchSpan
	for i := 0; i < len(letters); {
		if ends[i] <= i {
			i++
			continue
		}
		spans = append(spans, MatchSpan{Start: positions[i], End: positions[ends[i]]})
		i = ends[i]
	}
	return spans
}

func (h *Highlighter) Highlight(text string) string {
	var output strings.Builder
	offset := 0
	for _, span := range h.FindMatches(text) {
		output.WriteString(text[offset:span.Start])
		output.WriteString("\x1b[1;31m")
		output.WriteString(text[span.Start:span.End])
		output.WriteString("\x1b[0m")
		offset = span.End
	}
	output.WriteString(text[offset:])
	return output.String()
}
