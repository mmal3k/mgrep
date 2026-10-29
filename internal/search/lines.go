package search

import (
	"bufio"
	"fmt"
	"io"
	"os"
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
	file, err := os.Open(path)

	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	lineNumber := 0
	for scanner.Scan() {
		lineNumber++
		line := scanner.Text()
		if !matcher.ContainsMatch(line) {
			continue
		}
		if style.Highlighter != nil {
			line = style.Highlighter.Highlight(line)
		}
		if style.Pretty {
			if _, err := fmt.Fprintf(output, "Line %d | %s\n\n", lineNumber, line); err != nil {
				return err
			}
			continue
		}
		if _, err := fmt.Fprintln(output, line); err != nil {
			return err
		}
	}

	return scanner.Err()
}
