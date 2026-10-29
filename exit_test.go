package main

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestSearchExitCodes(t *testing.T) {
	empty := filepath.Join(t.TempDir(), "empty.txt")
	if err := os.WriteFile(empty, nil, 0600); err != nil {
		t.Fatal(err)
	}
	for _, algorithm := range []string{"dfa", "kmp", "boyer-moore"} {
		for _, tt := range []struct {
			name string
			args []string
			code int
		}{
			{"match", []string{"-regex", "Sargon", "-file", "internal/search/testdata/sample.txt"}, 0},
			{"no match", []string{"-regex", "ZZZZZ", "-file", "internal/search/testdata/sample.txt"}, 1},
			{"empty file", []string{"-regex", "Sargon", "-file", empty}, 1},
			{"missing file", []string{"-regex", "Sargon", "-file", filepath.Join(t.TempDir(), "missing")}, 2},
			{"invalid pattern", []string{"-regex", "[", "-text", "abc"}, 2},
			{"text match", []string{"-regex", "Sargon", "-text", "Hello Sargon"}, 0},
			{"text no match", []string{"-regex", "Sargon", "-text", "Hello"}, 1},
		} {
			t.Run(algorithm+"/"+tt.name, func(t *testing.T) {
				var output, diagnostics bytes.Buffer
				args := append([]string{"-algorithm", algorithm}, tt.args...)
				if code := execute(args, &output, &diagnostics); code != tt.code {
					t.Fatalf("exit code %d, want %d", code, tt.code)
				}
				if tt.code == 2 && diagnostics.Len() == 0 {
					t.Fatal("errors must be reported on stderr")
				}
				if tt.code != 2 && diagnostics.Len() != 0 {
					t.Fatalf("unexpected diagnostics: %s", &diagnostics)
				}
			})
		}
	}
}

func TestNonSearchExitCodes(t *testing.T) {
	for _, tt := range []struct {
		args []string
		code int
	}{
		{[]string{"-h"}, 0},
		{[]string{"-regex", "a", "-debug"}, 0},
		{[]string{"-unknown"}, 2},
		{nil, 2},
	} {
		var output, diagnostics bytes.Buffer
		if code := execute(tt.args, &output, &diagnostics); code != tt.code {
			t.Errorf("args %v: code %d, want %d", tt.args, code, tt.code)
		}
	}
}

type failingWriter struct{}

func (failingWriter) Write(p []byte) (int, error) {
	return 0, errors.New("output failed")
}

func TestOutputFailureExitCode(t *testing.T) {
	for _, args := range [][]string{
		{"-regex", "Sargon", "-file", "internal/search/testdata/sample.txt"},
		{"-regex", "Sargon", "-text", "Hello"},
	} {
		var diagnostics bytes.Buffer
		if code := execute(args, failingWriter{}, &diagnostics); code != 2 {
			t.Fatalf("output error must exit 2, got %d", code)
		}
	}
}
