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
		{"abc123", ".(.(.(.(.(a,b),c),1),2),3)"},
		{"colou?r", ".(.(.(.(.(c,o),l),o),?(u)),r)"},
		{"a{2,4}", "{2,4}(a)"},
		{"[a-z]+", "+([a-z])"},
		{`a\.b`, ".(.(a,.),b)"},
		{"^cat$", ".(.(.(.(^,c),a),t),$)"},
		{"ab", ".(a,b)"},
		{"a|b", "|(a,b)"},
		{"a*", "*(a)"},
		{"a+", "+(a)"},
		{"a+b", ".(+(a),b)"},
		{"(ab)+", "+(.(a,b))"},
		{"a|bc+", "|(a,.(b,+(c)))"},
		{"(a|b)+", "+(|(a,b))"},
		{"a+*", "*(+(a))"},
		{"a*+", "+(*(a))"},
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
	for _, input := range []string{"", "*", "|a", "a|", "(", "a|*b", "+", "+a", "a|+b", "(+a)", "[", "[z-a]", "[]", "a{2,1}", "a{,3}", "a{1001}", "a{2", "a{no}", "a{1000}{1000}", "^*", "[[:digit:]]", `a\`, `\1`} {
		if _, err := Parse(input); err == nil {
			t.Fatalf("%q: expected syntax error", input)
		}
	}
}
