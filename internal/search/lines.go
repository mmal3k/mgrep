package search

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
)

type TextMatcher interface {
	ContainsMatch(text string) bool
}

type LineStyle struct {
	Pretty      bool
	Highlighter *Highlighter
}

func Lines(path string, output io.Writer, matcher TextMatcher) error {
	return LinesWithStyle(path, output, matcher, LineStyle{})
}

func LinesWithStyle(path string, output io.Writer, matcher TextMatcher, style LineStyle) error {
	_, err := SearchLines(path, output, matcher, style)
	return err
}

func SearchLines(path string, output io.Writer, matcher TextMatcher, style LineStyle) (int, error) {
	matches := 0
	file, err := os.Open(path)

	if err != nil {
		return matches, err
	}
	defer file.Close()

	reader := bufio.NewReader(file)
	for lineNumber := 1; ; lineNumber++ {
		line, err := reader.ReadString('\n')
		if err != nil && err != io.EOF {
			return matches, err
		}
		if len(line) == 0 && err == io.EOF {
			return matches, nil
		}
		line = strings.TrimSuffix(strings.TrimSuffix(line, "\n"), "\r")
		if matcher.ContainsMatch(line) {
			if writeErr := writeMatchingLine(output, style, lineNumber, line); writeErr != nil {
				return matches, writeErr
			}
			matches++
		}
		if err == io.EOF {
			return matches, nil
		}
	}
}

func writeMatchingLine(output io.Writer, style LineStyle, lineNumber int, line string) error {
	if style.Highlighter != nil {
		line = style.Highlighter.Highlight(line)
	}
	if style.Pretty {
		_, err := fmt.Fprintf(output, "Line %d | %s\n\n", lineNumber, line)
		return err
	}
	_, err := fmt.Fprintln(output, line)
	return err
}
