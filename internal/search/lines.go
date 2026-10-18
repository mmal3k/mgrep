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

func Lines(path string, output io.Writer, matcher TextMatcher) error {
	file, err := os.Open(path)

	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()
		if !matcher.ContainsMatch(line) {
			continue
		}
		if _, err := fmt.Fprintln(output, line); err != nil {
			return err
		}
	}

	return scanner.Err()
}
