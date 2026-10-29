package automata

import (
	"slices"
	"unicode/utf8"
)

const (
	StartOfLine    rune = -1
	EndOfLine      rune = -2
	EmptyLine      rune = -3
	otherCharacter rune = -4
)

type RangeTransition struct {
	Lo, Hi rune
	To     []int
}

func alphabet(states []State) []rune {
	cuts := []rune{0, utf8.MaxRune + 1}
	for _, state := range states {
		for letter := range state.On {
			if letter >= 0 {
				cuts = append(cuts, letter, letter+1)
			}
		}
		for _, r := range state.Ranges {
			cuts = append(cuts, r.Lo, r.Hi+1)
		}
	}
	slices.Sort(cuts)
	return slices.Compact(cuts)
}

func (s *State) setRange(lo, hi rune, destination int) {
	if lo == hi {
		s.On[lo] = []int{destination}
		return
	}
	s.Ranges = append(s.Ranges, RangeTransition{Lo: lo, Hi: hi, To: []int{destination}})
}

func (d *DFA) Next(state int, letter rune) int {
	if state < 0 || state >= len(d.States) {
		return -1
	}
	s := d.States[state]
	next, explicit := s.On[letter]
	if !explicit {
		if letter < 0 {
			return state
		}
		next = s.Any
		for _, r := range s.Ranges {
			if r.Lo <= letter && letter <= r.Hi {
				next = r.To
				break
			}
		}
	}
	if len(next) == 0 {
		return -1
	}
	return next[0]
}

func (a *Automaton) boundaryClosure(seeds []int, boundary rune) []int {
	visited := make([]bool, len(a.States))
	pending := append([]int(nil), seeds...)
	var result []int
	for len(pending) > 0 {
		state := pending[len(pending)-1]
		pending = pending[:len(pending)-1]
		if visited[state] {
			continue
		}
		visited[state] = true
		result = append(result, state)
		pending = append(pending, a.States[state].Epsilon...)
		if boundary == EmptyLine {
			pending = append(pending, a.States[state].On[StartOfLine]...)
			pending = append(pending, a.States[state].On[EndOfLine]...)
		} else {
			pending = append(pending, a.States[state].On[boundary]...)
		}
	}
	slices.Sort(result)
	return result
}

func (a *Automaton) destinations(states []int, letter rune) []int {
	seen := make([]bool, len(a.States))
	for _, state := range states {
		next := append([]int(nil), a.States[state].Any...)
		if letter >= 0 {
			next = append(next, a.States[state].On[letter]...)
			for _, r := range a.States[state].Ranges {
				if r.Lo <= letter && letter <= r.Hi {
					next = append(next, r.To...)
				}
			}
		}
		for _, target := range next {
			for _, reachable := range a.epsilonClosure(target) {
				seen[reachable] = true
			}
		}
	}
	var result []int
	for state, exists := range seen {
		if exists {
			result = append(result, state)
		}
	}
	return result
}
