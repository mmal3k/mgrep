package automata

import (
	"slices"
	"strings"
)

type DFA struct {
	Automaton
}

func (d *DFA) String() string {
	return strings.ReplaceAll(d.Automaton.String(), "--ANY-->", "--OTHER-->")
}

func (d *DFA) isAccepted(text string) bool {
	current := d.Start

	for _, letter := range text {
		next, ok := d.States[current].On[letter]
		if !ok {
			next = d.States[current].Any
		}
		if len(next) == 0 {
			return false
		}
		current = next[0]
	}

	return slices.Contains(d.Accept, current)
}
