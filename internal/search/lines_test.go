package search

import (
	"bytes"
	"testing"
)

const sampleText = "ab\n" +
	"abcd\n" +
	"xxabyy\n" +
	"xyz\n" +
	"\n" +
	"The cats sleep.\n" +
	"A dog runs.\n" +
	"Sargon ruled Babylon.\n" +
	"Sagon appears in this line.\n" +
	"Sron appears in this line.\n" +
	"Babylon was an ancient city.\n" +
	"acb\n" +
	"a b\n" +
	"a.b\n" +
	"ab ab ab\n"

func TestReadFile(t *testing.T) {
	var output bytes.Buffer
	want := sampleText
	matcher, err := NewMatcher(".*")
	if err != nil {
		t.Fatal(err)
	}
	err = Lines("testdata/sample.txt", &output, matcher)
	if err != nil {
		t.Fatal(err)
	}

	if got := output.String(); got != want {
		t.Fatalf("got:\n%s\nwant:\n%s", got, want)
	}
}

func TestLinesContainingMatch(t *testing.T) {
	tests := []struct {
		pattern string
		want    string
	}{
		{pattern: "ab", want: "ab\nabcd\nxxabyy\nSargon ruled Babylon.\nBabylon was an ancient city.\nab ab ab\n"},
		{pattern: "a.b", want: "acb\na b\na.b\n"},
		{pattern: "a*", want: sampleText},
		{pattern: "z", want: "xyz\n"},
		{pattern: "zz", want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.pattern, func(t *testing.T) {
			matcher, err := NewMatcher(tt.pattern)
			if err != nil {
				t.Fatal(err)
			}
			var output bytes.Buffer

			if err := Lines("testdata/sample.txt", &output, matcher); err != nil {
				t.Fatal(err)
			}

			if got := output.String(); got != tt.want {
				t.Fatalf("got:\n%s\nwant:\n%s", got, tt.want)
			}
		})
	}
}

func TestPrettyLinesWithoutColor(t *testing.T) {
	matcher, err := NewMatcher("S(a|g|r)+on")
	if err != nil {
		t.Fatal(err)
	}
	var output bytes.Buffer
	if err := LinesWithStyle("testdata/sample.txt", &output, matcher, LineStyle{Pretty: true}); err != nil {
		t.Fatal(err)
	}
	want := "Line 8 | Sargon ruled Babylon.\n\nLine 9 | Sagon appears in this line.\n\nLine 10 | Sron appears in this line.\n\n"
	if got := output.String(); got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}
