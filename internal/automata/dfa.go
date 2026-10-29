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

func (d *DFA) IsAccepted(text string) bool {
	boundary := StartOfLine
	if text == "" {
		boundary = EmptyLine
	}
	current := d.Next(d.Start, boundary)
	for _, letter := range text {
		current = d.Next(current, letter)
		if current < 0 {
			return false
		}
	}
	if text != "" {
		current = d.Next(current, EndOfLine)
	}
	return slices.Contains(d.Accept, current)
}

func (d *DFA) Minimize() *DFA {
	var groups [][]int
	non_accepting := []int{len(d.States)}

	for s := range d.States {
		if !slices.Contains(d.Accept, s) {
			non_accepting = append(non_accepting, s)
		}
	}

	groups = append(groups, non_accepting)
	if len(d.Accept) > 0 {
		groups = append(groups, d.Accept)
	}

	for {
		groupOf := d.groupOfStates(groups)
		var newGroups [][]int

		for _, group := range groups {
			var subgroups [][]int
		stateLoop:
			for _, state := range group {
				for id, subgroup := range subgroups {
					if d.equivalent(state, subgroup[0], groupOf) {
						subgroups[id] = append(subgroups[id], state)
						continue stateLoop
					}
				}
				subgroups = append(subgroups, []int{state})
			}
			newGroups = append(newGroups, subgroups...)
		}

		if len(newGroups) == len(groups) {
			groups = newGroups
			break
		}
		groups = newGroups
	}

	groupOf := d.groupOfStates(groups)
	minimized := &DFA{}
	stateForGroup := make(map[int]int)
	var pending []int

	addGroup := func(groupID int) int {
		if id, ok := stateForGroup[groupID]; ok {
			return id
		}
		id := minimized.newState()
		stateForGroup[groupID] = id
		pending = append(pending, groupID)
		if slices.Contains(d.Accept, groups[groupID][0]) {
			minimized.Accept = append(minimized.Accept, id)
		}
		return id
	}

	destinationGroup := func(state int, letter rune) int {
		next := d.Next(state, letter)
		if next < 0 {
			return groupOf[len(d.States)]
		}
		return groupOf[next]
	}

	cuts := alphabet(d.States)
	minimized.Start = addGroup(groupOf[d.Start])
	for current := 0; current < len(pending); current++ {
		representative := groups[pending[current]][0]
		fallback := destinationGroup(representative, 0)
		next := addGroup(fallback)
		minimized.States[current].Any = []int{next}
		for i := 0; i+1 < len(cuts); i++ {
			group := destinationGroup(representative, cuts[i])
			if group == fallback {
				continue
			}
			next := addGroup(group)
			minimized.States[current].setRange(cuts[i], cuts[i+1]-1, next)
		}
		for _, marker := range []rune{StartOfLine, EndOfLine, EmptyLine} {
			group := destinationGroup(representative, marker)
			if group == pending[current] {
				continue
			}
			next := addGroup(group)
			minimized.States[current].On[marker] = []int{next}
		}
	}

	return minimized
}

func (d *DFA) groupOfStates(groups [][]int) []int {
	groupOf := make([]int, len(d.States)+1)
	for groupID, group := range groups {
		for _, state := range group {
			groupOf[state] = groupID
		}
	}
	return groupOf
}

func (d *DFA) equivalent(s1, s2 int, groupOf []int) bool {
	if groupOf[s1] != groupOf[s2] {
		return false
	}
	var states []State
	if s1 < len(d.States) {
		states = append(states, d.States[s1])
	}
	if s2 < len(d.States) {
		states = append(states, d.States[s2])
	}
	group := func(state int, letter rune) int {
		next := d.Next(state, letter)
		if next < 0 {
			return groupOf[len(d.States)]
		}
		return groupOf[next]
	}
	cuts := alphabet(states)
	for _, letter := range append(cuts[:len(cuts)-1], StartOfLine, EndOfLine, EmptyLine) {
		if group(s1, letter) != group(s2, letter) {
			return false
		}
	}
	return true
}
