package automata

import (
	"gutenberg-backend/internal/parse"
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
