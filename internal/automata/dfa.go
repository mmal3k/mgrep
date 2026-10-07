package automata

import (
	"strings"
)

type DFA struct {
	Automaton
}

func (d *DFA) String() string {
	return strings.ReplaceAll(d.Automaton.String(), "--ANY-->", "--OTHER-->")
}
