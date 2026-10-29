package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestExtendedPatternsAgreeWithGrep(t *testing.T) {
	grep, err := exec.LookPath("grep")
	if err != nil {
		t.Skip("grep is not installed")
	}
	path := filepath.Join(t.TempDir(), "input.txt")
	input := strings.Join([]string{"", "abc123", "hello!", "a", "b", "c", "z", "12345", "color", "colour", "colouur", "cat", "cat sleeps", "the cat", "cats", "a.b", "axb", "a*b", "[a]", "aa", "aaa", "aaaaa", "abab", "ababa", "!", "-", "]", "Sargon", "Sargon123!", "a\\b", "last line"}, "\n")
	if err := os.WriteFile(path, []byte(input), 0600); err != nil {
		t.Fatal(err)
	}
	for _, pattern := range []string{
		"abc123", "hello!", "[abc]", "[a-z]+", "[A-Za-z0-9]+", "[^abc]+", "[^a-z]", "[-a]", "[a-]", "[]a]", "[0-9]{2,4}", "colou?r", "a{0}", "a{3}", "a{2,4}", "a{2,}", "(ab){1,2}", "^cat", "cat$", "^cat$", "^$", "$^", "^^a$$", "^a|b$", "a^", "$a", "a$|ab", "^S[a-z]+[0-9]*!$", `a\.b`, `a\*b`, `\[a\]`, `a\\b`,
	} {
		t.Run(pattern, func(t *testing.T) {
			cmd := exec.Command(grep, "-E", "--", pattern, path)
			cmd.Env = append(os.Environ(), "LC_ALL=C")
			want, err := cmd.Output()
			if err != nil {
				exit, ok := err.(*exec.ExitError)
				if !ok || exit.ExitCode() != 1 {
					t.Fatalf("grep failed: %v", err)
				}
			}
			var output, diagnostics bytes.Buffer
			if err := run([]string{"-regex", pattern, "-file", path, "-format", "plain", "-color", "never"}, &output, &diagnostics); err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(output.Bytes(), want) {
				t.Fatalf("output %q, grep %q", output.String(), want)
			}
		})
	}
}
