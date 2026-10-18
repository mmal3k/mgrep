package automata

import (
	"maps"
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

	destinationGroup := func(next []int) int {
		if len(next) == 0 {
			return groupOf[len(d.States)]
		}
		return groupOf[next[0]]
	}

	minimized.Start = addGroup(groupOf[d.Start])
	for current := 0; current < len(pending); current++ {
		representative := groups[pending[current]][0]
		var state State
		if representative < len(d.States) {
			state = d.States[representative]
		}
		for _, letter := range slices.Sorted(maps.Keys(state.On)) {
			next := addGroup(destinationGroup(state.On[letter]))
			minimized.States[current].On[letter] = []int{next}
		}
		next := addGroup(destinationGroup(state.Any))
		minimized.States[current].Any = []int{next}
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

	var first, second State
	if s1 < len(d.States) {
		first = d.States[s1]
	}
	if s2 < len(d.States) {
		second = d.States[s2]
	}

	destinationGroup := func(next []int) int {
		if len(next) == 0 {
			return groupOf[len(d.States)]
		}
		return groupOf[next[0]]
	}

	if destinationGroup(first.Any) != destinationGroup(second.Any) {
		return false
	}
	for letter, next1 := range first.On {
		next2, ok := second.On[letter]
		if !ok {
			next2 = second.Any
		}
		if destinationGroup(next1) != destinationGroup(next2) {
			return false
		}
	}
	for letter, next2 := range second.On {
		if _, ok := first.On[letter]; ok {
			continue
		}
		if destinationGroup(first.Any) != destinationGroup(next2) {
			return false
		}
	}
	return true
}
