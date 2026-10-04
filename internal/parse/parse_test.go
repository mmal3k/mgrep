package parse

import (
	"testing"
)

func TestParseTrees(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"a", "a"},
		{"ab", ".(a,b)"},
		{"a|b", "|(a,b)"},
		{"a*", "*(a)"},
		{"(a|b)*", "*(|(a,b))"},
		{"a|bc*", "|(a,.(b,*(c)))"}, // Aho–Ullman ex. 10.25
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := Parse(tt.input)
			if err != nil {
				t.Fatalf("Parse: %v", err)
			}
			if got.String() != tt.want {
				t.Fatalf("got %s, want %s", got, tt.want)
			}
		})
	}
}

func TestParseSyntaxError(t *testing.T) {
	for _, input := range []string{"", "*", "|a", "a|", "("} {
		if _, err := Parse(input); err == nil {
			t.Fatalf("%q: expected syntax error", input)
		}
	}
}
