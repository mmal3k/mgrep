package automata

import (
	"github.com/mmal3k/mgrep/internal/parse"
	"slices"
	"unicode/utf8"
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
	case parse.STAR, parse.PLUS:
		newStart := n.newState()
		R1 := n.build(tree.SubTrees[0])
		newAccept := n.newState()
		if tree.Root == parse.STAR {
			n.States[newStart].Epsilon = append(n.States[newStart].Epsilon, newAccept)
		}
		n.States[newStart].Epsilon = append(n.States[newStart].Epsilon, R1.Start)
		n.States[R1.Accept].Epsilon = append(n.States[R1.Accept].Epsilon, R1.Start)
		n.States[R1.Accept].Epsilon = append(n.States[R1.Accept].Epsilon, newAccept)
		return Fragment{Start: newStart, Accept: newAccept}
	case parse.OPTIONAL:
		return n.repeat(tree.SubTrees[0], 0, 1)
	case parse.REPEAT:
		return n.repeat(tree.SubTrees[0], tree.Min, tree.Max)
	case parse.CLASS:
		start := n.newState()
		end := n.newState()
		ranges := append([]parse.RuneRange(nil), tree.Ranges...)
		if tree.Negated {
			slices.SortFunc(ranges, func(a, b parse.RuneRange) int { return int(a.Lo - b.Lo) })
			var complement []parse.RuneRange
			next := rune(0)
			for _, r := range ranges {
				if r.Lo > next {
					complement = append(complement, parse.RuneRange{Lo: next, Hi: r.Lo - 1})
				}
				next = max(next, r.Hi+1)
			}
			if next <= utf8.MaxRune {
				complement = append(complement, parse.RuneRange{Lo: next, Hi: utf8.MaxRune})
			}
			ranges = complement
		}
		for _, r := range ranges {
			n.States[start].setRange(r.Lo, r.Hi, end)
		}
		return Fragment{Start: start, Accept: end}
	case parse.BEGIN, parse.END:
		start := n.newState()
		end := n.newState()
		boundary := StartOfLine
		if tree.Root == parse.END {
			boundary = EndOfLine
		}
		n.States[start].On[boundary] = []int{end}
		return Fragment{Start: start, Accept: end}
	case parse.DOT:
		newStart := n.newState()
		newAccept := n.newState()
		n.States[newStart].Any = []int{newAccept}
		return Fragment{Start: newStart, Accept: newAccept}
	default:
		newStart := n.newState()
		newAccept := n.newState()
		letter := rune(tree.Root)
		if tree.Root == parse.LITERAL {
			letter = tree.Value
		}
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

func (n *NFA) repeat(tree *parse.RegExTree, min, max int) Fragment {
	start := n.newState()
	current := start
	for i := 0; i < min; i++ {
		part := n.build(tree)
		n.States[current].Epsilon = append(n.States[current].Epsilon, part.Start)
		current = part.Accept
	}
	if max < 0 {
		part := n.build(&parse.RegExTree{Root: parse.STAR, SubTrees: []*parse.RegExTree{tree}})
		n.States[current].Epsilon = append(n.States[current].Epsilon, part.Start)
		return Fragment{Start: start, Accept: part.Accept}
	}
	exits := []int{current}
	for i := min; i < max; i++ {
		part := n.build(tree)
		n.States[current].Epsilon = append(n.States[current].Epsilon, part.Start)
		current = part.Accept
		exits = append(exits, current)
	}
	end := n.newState()
	for _, state := range exits {
		n.States[state].Epsilon = append(n.States[state].Epsilon, end)
	}
	return Fragment{Start: start, Accept: end}
}

func (n *NFA) Determinize() *DFA {
	d := &DFA{}
	closure := n.epsilonClosure(n.Start)
	slices.Sort(closure)
	d.Start = d.newState()
	subsets := [][]int{closure}
	intern := func(states []int) int {
		for id, existing := range subsets {
			if slices.Equal(existing, states) {
				return id
			}
		}
		id := d.newState()
		subsets = append(subsets, states)
		return id
	}
	cuts := alphabet(n.States)
	boundaries := false
	for _, state := range n.States {
		if len(state.On[StartOfLine]) > 0 || len(state.On[EndOfLine]) > 0 {
			boundaries = true
			break
		}
	}
	for current := 0; current < len(subsets); current++ {
		states := subsets[current]
		fallback := n.destinations(states, otherCharacter)
		if len(fallback) > 0 {
			target := intern(fallback)
			d.States[current].Any = []int{target}
		}
		for i := 0; i+1 < len(cuts); i++ {
			next := n.destinations(states, cuts[i])
			if slices.Equal(next, fallback) {
				continue
			}
			target := intern(next)
			d.States[current].setRange(cuts[i], cuts[i+1]-1, target)
		}
		if boundaries {
			for _, marker := range []rune{StartOfLine, EndOfLine, EmptyLine} {
				next := n.boundaryClosure(states, marker)
				if slices.Equal(next, states) {
					continue
				}
				target := intern(next)
				d.States[current].On[marker] = []int{target}
			}
		}
		for _, state := range states {
			if slices.Contains(n.Accept, state) {
				d.Accept = append(d.Accept, current)
				break
			}
		}
	}
	return d
}
