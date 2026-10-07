package automata

import (
	"gutenberg-backend/internal/parse"
	"maps"
	"slices"
)

type NFA struct {
	Automaton
}

func (n *NFA) build(tree *parse.RegExTree) Fragment {
	switch tree.Root {
	case parse.CONCAT:
		R1 := n.build(tree.SubTrees[0])
		R2 := n.build(tree.SubTrees[1])
		n.States[R1.Accept].Epsilon = append(n.States[R1.Accept].Epsilon, R2.Start)
		return Fragment{Start: R1.Start, Accept: R2.Accept}
	case parse.ALTERNATION:
		newStart := n.newState()
		R1 := n.build(tree.SubTrees[0])
		R2 := n.build(tree.SubTrees[1])
		newAccept := n.newState()
		n.States[newStart].Epsilon = append(n.States[newStart].Epsilon, R1.Start)
		n.States[newStart].Epsilon = append(n.States[newStart].Epsilon, R2.Start)
		n.States[R1.Accept].Epsilon = append(n.States[R1.Accept].Epsilon, newAccept)
		n.States[R2.Accept].Epsilon = append(n.States[R2.Accept].Epsilon, newAccept)
		return Fragment{Start: newStart, Accept: newAccept}
	case parse.STAR:
		newStart := n.newState()
		R1 := n.build(tree.SubTrees[0])
		newAccept := n.newState()
		n.States[newStart].Epsilon = append(n.States[newStart].Epsilon, newAccept)
		n.States[newStart].Epsilon = append(n.States[newStart].Epsilon, R1.Start)
		n.States[R1.Accept].Epsilon = append(n.States[R1.Accept].Epsilon, R1.Start)
		n.States[R1.Accept].Epsilon = append(n.States[R1.Accept].Epsilon, newAccept)
		return Fragment{Start: newStart, Accept: newAccept}
	case parse.DOT:
		newStart := n.newState()
		newAccept := n.newState()
		n.States[newStart].Any = []int{newAccept}
		return Fragment{Start: newStart, Accept: newAccept}
	default:
		newStart := n.newState()
		newAccept := n.newState()
		letter := rune(tree.Root)
		n.States[newStart].On[letter] = []int{newAccept}
		return Fragment{Start: newStart, Accept: newAccept}
	}
}

func FromTree(tree *parse.RegExTree) *NFA {
	n := &NFA{}
	fragment := n.build(tree)

	n.Start = fragment.Start
	n.Accept = append(n.Accept, fragment.Accept)

	return n
}

func (n *NFA) Determinize() *DFA {
	d := &DFA{}

	// subset is to map each state of the DFA to its list on the NFA
	closure := n.epsilonClosure(n.Start)
	slices.Sort(closure)
	d.Start = d.newState()
	subset := [][]int{closure}

	addSubset := func(currentId int, anyDestinations map[int]bool) {
		destinations := make(map[rune]map[int]bool)
		for _, state := range subset[currentId] {
			for letter, nextStates := range n.States[state].On {
				if destinations[letter] == nil {
					destinations[letter] = make(map[int]bool)
				}
				for _, nextId := range nextStates {
					for _, reachable := range n.epsilonClosure(nextId) {
						destinations[letter][reachable] = true
					}
				}
			}
		}

		for _, letter := range slices.Sorted(maps.Keys(destinations)) {
			states := destinations[letter]
			for reachable := range anyDestinations {
				states[reachable] = true
			}
			nextList := slices.Sorted(maps.Keys(states))
			nextId := slices.IndexFunc(subset, func(existing []int) bool {
				return slices.Equal(existing, nextList)
			})
			if nextId == -1 {
				nextId = d.newState()
				subset = append(subset, nextList)
			}
			d.States[currentId].On[letter] = []int{nextId}
		}
	}

	addAccept := func(currentId int) {
		for _, state := range subset[currentId] {
			if slices.Contains(n.Accept, state) {
				d.Accept = append(d.Accept, currentId)
				break
			}
		}
	}

	addAny := func(currentId int) map[int]bool {
		destinations := make(map[int]bool)
		for _, state := range subset[currentId] {
			for _, nextId := range n.States[state].Any {
				for _, reachable := range n.epsilonClosure(nextId) {
					destinations[reachable] = true
				}
			}
		}

		if len(destinations) == 0 {
			return destinations
		}

		nextList := slices.Sorted(maps.Keys(destinations))
		nextId := slices.IndexFunc(subset, func(existing []int) bool {
			return slices.Equal(existing, nextList)
		})
		if nextId == -1 {
			nextId = d.newState()
			subset = append(subset, nextList)
		}
		d.States[currentId].Any = []int{nextId}
		return destinations
	}

	for currentId := 0; currentId < len(subset); currentId++ {
		anyDestinations := addAny(currentId)
		addSubset(currentId, anyDestinations)
		addAccept(currentId)
	}

	return d
}
