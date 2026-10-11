package automata

import (
	"fmt"
	"slices"
	"sort"
	"strings"
)

type State struct {
	On      map[rune][]int
	Any     []int
	Epsilon []int
}

type Fragment struct {
	Start  int
	Accept int
}

type Automaton struct {
	States []State
	Start  int
	Accept []int
}

func (n *Automaton) String() string {
	var output strings.Builder
	fmt.Fprintf(&output, "Start: %d\nAccept: %v\n\n", n.Start, n.Accept)
	for id, state := range n.States {
		letters := make([]rune, 0, len(state.On))
		for letter := range state.On {
			letters = append(letters, letter)
		}
		sort.Slice(letters, func(i, j int) bool { return letters[i] < letters[j] })
		for _, letter := range letters {
			for _, destination := range state.On[letter] {
				fmt.Fprintf(&output, "%d --%c--> %d\n", id, letter, destination)
			}
		}
		for _, destination := range state.Any {
			fmt.Fprintf(&output, "%d --ANY--> %d\n", id, destination)
		}
		for _, destination := range state.Epsilon {
			fmt.Fprintf(&output, "%d --ε--> %d\n", id, destination)
		}
	}
	return output.String()
}

func (a *Automaton) newState() int {
	id := len(a.States)
	a.States = append(a.States, State{
		On: make(map[rune][]int),
	})
	return id
}

func (a *Automaton) epsilonClosure(state int) []int {
	visited := make(map[int]bool)
	closure := []int{}

	var visit func(int)
	visit = func(state int) {
		if visited[state] {
			return
		}

		visited[state] = true
		closure = append(closure, state)

		for _, next := range a.States[state].Epsilon {
			visit(next)
		}
	}

	visit(state)

	return closure
}

func (a *Automaton) IsAccepted(text string) bool {
	current := make(map[int]bool)

	for _, state := range a.epsilonClosure(a.Start) {
		current[state] = true
	}

	for _, c := range text {
		next := make(map[int]bool)

		for state := range current {
			destinations := append([]int{}, a.States[state].On[c]...)
			destinations = append(destinations, a.States[state].Any...)

			for _, destination := range destinations {
				for _, reachable := range a.epsilonClosure(destination) {
					next[reachable] = true
				}
			}
		}

		if len(next) == 0 {
			return false
		}

		current = next
	}

	for state := range current {
		if slices.Contains(a.Accept, state) {
			return true
		}
	}
	return false
}
