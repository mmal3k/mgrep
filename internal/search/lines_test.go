package search

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
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

func TestLongLines(t *testing.T) {
	long := strings.Repeat("x", 1024*1024) + "Sargon"
	factories := []struct {
		name string
		new  func(string) (TextMatcher, error)
	}{
		{"dfa", func(pattern string) (TextMatcher, error) { return NewMatcher(pattern) }},
		{"kmp", func(pattern string) (TextMatcher, error) { return NewKMPMatcher(pattern) }},
		{"boyer-moore", func(pattern string) (TextMatcher, error) { return NewBoyerMooreMatcher(pattern) }},
	}
	for _, factory := range factories {
		for _, ending := range []string{"\n", "\r\n", ""} {
			t.Run(factory.name+"/"+fmtEnding(ending), func(t *testing.T) {
				path := filepath.Join(t.TempDir(), "long.txt")
				input := strings.Repeat("x", 1024*1024) + "\nSargon\n" + long + ending
				if err := os.WriteFile(path, []byte(input), 0600); err != nil {
					t.Fatal(err)
				}
				matcher, err := factory.new("Sargon")
				if err != nil {
					t.Fatal(err)
				}
				var output bytes.Buffer
				if err := Lines(path, &output, matcher); err != nil {
					t.Fatal(err)
				}
				if got, want := output.String(), "Sargon\n"+long+"\n"; got != want {
					t.Fatalf("incorrect long-line output: got %d bytes, want %d", len(got), len(want))
				}
			})
		}
	}
}

func fmtEnding(ending string) string {
	switch ending {
	case "\n":
		return "LF"
	case "\r\n":
		return "CRLF"
	default:
		return "no-newline"
	}
}

func TestLineEndings(t *testing.T) {
	matcher, err := NewMatcher(".*")
	if err != nil {
		t.Fatal(err)
	}
	for _, tt := range []struct{ input, want string }{
		{"", ""},
		{"\n", "\n"},
		{"ab\n", "ab\n"},
		{"ab", "ab\n"},
		{"ab\r\n\r\ncd\r", "ab\n\ncd\n"},
	} {
		path := filepath.Join(t.TempDir(), "lines.txt")
		if err := os.WriteFile(path, []byte(tt.input), 0600); err != nil {
			t.Fatal(err)
		}
		var output bytes.Buffer
		if err := Lines(path, &output, matcher); err != nil {
			t.Fatal(err)
		}
		if got := output.String(); got != tt.want {
			t.Errorf("input %q: got %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestPrettyLongLineWithHighlighting(t *testing.T) {
	prefix := strings.Repeat("x", 128*1024)
	path := filepath.Join(t.TempDir(), "long.txt")
	if err := os.WriteFile(path, []byte("ignored\n"+prefix+"Sargon"), 0600); err != nil {
		t.Fatal(err)
	}
	matcher, err := NewMatcher("Sargon")
	if err != nil {
		t.Fatal(err)
	}
	highlighter, err := NewHighlighter("Sargon", "dfa")
	if err != nil {
		t.Fatal(err)
	}
	var output bytes.Buffer
	if err := LinesWithStyle(path, &output, matcher, LineStyle{Pretty: true, Highlighter: highlighter}); err != nil {
		t.Fatal(err)
	}
	want := "Line 2 | " + prefix + "\x1b[1;31mSargon\x1b[0m\n\n"
	if output.String() != want {
		t.Fatal("incorrect formatting or highlighting of long line")
	}
}
