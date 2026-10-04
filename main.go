package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"gutenberg-backend/internal/parse"
)

func main() {
	input, err := readRegEx()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	tree, err := parse.Parse(input)
	if err != nil {
		fmt.Fprintf(os.Stderr, "syntax error for %q: %v\n", input, err)
		os.Exit(1)
	}
	fmt.Println(tree)
}

func readRegEx() (string, error) {
	if len(os.Args) > 1 {
		return os.Args[1], nil
	}
	fmt.Print("regEx: ")
	line, err := bufio.NewReader(os.Stdin).ReadString('\n')
	if err != nil && err != io.EOF {
		return "", err
	}
	s := strings.TrimSpace(line)
	if s == "" {
		return "", fmt.Errorf("empty regEx")
	}
	return s, nil
}
