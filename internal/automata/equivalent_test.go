package automata

import "testing"

func TestEquivalent(t *testing.T) {
	tests := []struct {
		name   string
		states []State
		groups [][]int
		s1     int
		s2     int
		want   bool
	}{
		{
			name: "different destination IDs in the same group",
			states: []State{
				{On: map[rune][]int{'a': {2}}},
				{On: map[rune][]int{'a': {3}}},
				{}, {},
			},
			groups: [][]int{{0, 1, 4}, {2, 3}},
			s1:     0, s2: 1, want: true,
		},
		{
			name: "destinations in different groups",
			states: []State{
				{On: map[rune][]int{'a': {2}}},
				{On: map[rune][]int{'a': {3}}},
				{}, {},
			},
			groups: [][]int{{0, 1, 4}, {2}, {3}},
			s1:     0, s2: 1, want: false,
		},
		{
			name:   "current states in different groups",
			states: []State{{}, {}},
			groups: [][]int{{0, 2}, {1}},
			s1:     0, s2: 1, want: false,
		},
		{
			name: "explicit transition agrees with fallback",
			states: []State{
				{On: map[rune][]int{'a': {2}}, Any: []int{2}},
				{Any: []int{2}},
				{},
			},
			groups: [][]int{{0, 1, 3}, {2}},
			s1:     0, s2: 1, want: true,
		},
		{
			name: "letter found only on the second state",
			states: []State{
				{Any: []int{2}},
				{On: map[rune][]int{'a': {3}}, Any: []int{2}},
				{}, {},
			},
			groups: [][]int{{0, 1, 4}, {2}, {3}},
			s1:     0, s2: 1, want: false,
		},
		{
			name: "different fallbacks without explicit letters",
			states: []State{
				{Any: []int{2}},
				{}, {},
			},
			groups: [][]int{{0, 1, 3}, {2}},
			s1:     0, s2: 1, want: false,
		},
		{
			name: "missing transition agrees with a rejecting state",
			states: []State{
				{On: map[rune][]int{'a': {1}}},
				{},
			},
			groups: [][]int{{0, 1, 2}},
			s1:     0, s2: 1, want: true,
		},
		{
			name: "virtual sink behaves like an explicit rejecting loop",
			states: []State{
				{Any: []int{0}},
			},
			groups: [][]int{{0, 1}},
			s1:     0, s2: 1, want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &DFA{Automaton: Automaton{States: tt.states}}
			groupOf := d.groupOfStates(tt.groups)
			if got := d.equivalent(tt.s1, tt.s2, groupOf); got != tt.want {
				t.Fatalf("got %v, want %v", got, tt.want)
			}
			if got := d.equivalent(tt.s2, tt.s1, groupOf); got != tt.want {
				t.Fatalf("reverse comparison: got %v, want %v", got, tt.want)
			}
		})
	}
}
