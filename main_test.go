package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestRunText(t *testing.T) {
	tests := []struct {
		pattern string
		text    string
		want    string
	}{
		{"ab", "xxabyy", "Text: \"xxabyy\"\nAccepted: false\nContains match: true\n"},
		{"ab", "xyz", "Text: \"xyz\"\nAccepted: false\nContains match: false\n"},
		{"a*", "", "Text: \"\"\nAccepted: true\nContains match: true\n"},
	}
	for _, tt := range tests {
		t.Run(tt.pattern+"/"+tt.text, func(t *testing.T) {
			var output, diagnostics bytes.Buffer
			if err := run([]string{"-regex", tt.pattern, "-text", tt.text}, &output, &diagnostics); err != nil {
				t.Fatal(err)
			}
			if got := output.String(); got != tt.want {
				t.Fatalf("got %q, want %q", got, tt.want)
			}
			if diagnostics.Len() != 0 {
				t.Fatalf("unexpected diagnostics: %s", &diagnostics)
			}
		})
	}
}

func TestRunFileWithDebug(t *testing.T) {
	var output, diagnostics bytes.Buffer
	args := []string{"-regex", "ab", "-file", "internal/search/testdata/sample.txt", "-debug"}
	if err := run(args, &output, &diagnostics); err != nil {
		t.Fatal(err)
	}
	want := "ab\nabcd\nxxabyy\nSargon ruled Babylon.\nBabylon was an ancient city.\nab ab ab\n"
	if got := output.String(); got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
	for _, label := range []string{"Syntax tree:", "NFA:", "DFA:", "Minimized DFA:"} {
		if !strings.Contains(diagnostics.String(), label) {
			t.Errorf("missing diagnostic %q", label)
		}
	}
}

func TestRunRejectsInvalidArguments(t *testing.T) {
	tests := [][]string{
		{},
		{"-text", "ab"},
		{"-regex", "ab"},
		{"-regex", "ab", "-text", "ab", "-file", "book.txt"},
		{"-regex", "a+", "-text", "a"},
		{"-regex", "ab", "-file", "testdata/missing.txt"},
		{"-unknown"},
		{"-algorithm", "unknown", "-regex", "ab", "-text", "ab"},
		{"-algorithm", "kmp", "-regex", "a|b", "-text", "a"},
		{"-algorithm", "kmp", "-regex", "ab", "-debug"},
		{"ab", "xxabyy"},
	}
	for _, args := range tests {
		var output, diagnostics bytes.Buffer
		if err := run(args, &output, &diagnostics); err == nil {
			t.Errorf("expected error for %v", args)
		}
		if output.Len() != 0 {
			t.Errorf("unexpected output for %v: %s", args, &output)
		}
	}
}

func TestRunHelp(t *testing.T) {
	var output, diagnostics bytes.Buffer
	if err := run([]string{"-h"}, &output, &diagnostics); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(diagnostics.String(), "-regex") {
		t.Fatal("help must describe the regex option")
	}
}

func TestRunKMP(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{
		{
			name: "file",
			args: []string{"-algorithm", "kmp", "-regex", "Sargon", "-file", "internal/search/testdata/sample.txt"},
			want: "Sargon ruled Babylon.\n",
		},
		{
			name: "text",
			args: []string{"-algorithm", "kmp", "-regex", "Sargon", "-text", "Hello Sargon"},
			want: "Text: \"Hello Sargon\"\nAccepted: false\nContains match: true\n",
		},
		{
			name: "whole string",
			args: []string{"-algorithm", "kmp", "-regex", "Sargon", "-text", "Sargon"},
			want: "Text: \"Sargon\"\nAccepted: true\nContains match: true\n",
		},
		{
			name: "empty text",
			args: []string{"-algorithm", "kmp", "-regex", "Sargon", "-text", ""},
			want: "Text: \"\"\nAccepted: false\nContains match: false\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var output, diagnostics bytes.Buffer
			if err := run(tt.args, &output, &diagnostics); err != nil {
				t.Fatal(err)
			}
			if got := output.String(); got != tt.want {
				t.Fatalf("got %q, want %q", got, tt.want)
			}
			if diagnostics.Len() != 0 {
				t.Fatalf("unexpected diagnostics: %s", &diagnostics)
			}
		})
	}
}
