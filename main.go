package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/mmal3k/regex-cli/internal/automata"
	"github.com/mmal3k/regex-cli/internal/parse"
	"github.com/mmal3k/regex-cli/internal/search"
)

type options struct {
	pattern string
	text    *string
	path    string
	debug   bool
}

func main() {
	if err := run(os.Args[1:], os.Stdout, os.Stderr); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(args []string, output io.Writer, diagnostics io.Writer) error {
	opts, err := readOptions(args, diagnostics)
	if errors.Is(err, flag.ErrHelp) {
		return nil
	}
	if err != nil {
		return err
	}

	if opts.debug {
		if err := printAutomata(opts.pattern, diagnostics); err != nil {
			return err
		}
	}
	if opts.path != "" {
		return searchFile(opts.pattern, opts.path, output)
	}
	if opts.text != nil {
		return testText(opts.pattern, *opts.text, output)
	}
	return nil
}

func readOptions(args []string, output io.Writer) (options, error) {
	var opts options
	flags := flag.NewFlagSet("regex-cli", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	flags.StringVar(&opts.pattern, "regex", "", "Regex to compile")
	flags.StringVar(&opts.path, "file", "", "Text file to search")
	flags.BoolVar(&opts.debug, "debug", false, "Print the syntax tree, NFA, and DFA")
	flags.Func("text", "Text to test", func(text string) error {
		opts.text = &text
		return nil
	})

	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			flags.SetOutput(output)
			flags.Usage()
		}
		return opts, err
	}
	if flags.NArg() != 0 {
		return opts, fmt.Errorf("use -regex, -text, and -file instead of positional arguments")
	}
	if opts.pattern == "" {
		return opts, fmt.Errorf("provide a regex with -regex")
	}
	if opts.text != nil && opts.path != "" {
		return opts, fmt.Errorf("choose either -text or -file")
	}
	if opts.text == nil && opts.path == "" && !opts.debug {
		return opts, fmt.Errorf("choose -text, -file, or -debug")
	}
	return opts, nil
}

func searchFile(pattern string, path string, output io.Writer) error {
	matcher, err := search.NewMatcher(pattern)
	if err != nil {
		return fmt.Errorf("invalid regex %q: %w", pattern, err)
	}
	return search.Lines(path, output, matcher)
}

func testText(pattern string, text string, output io.Writer) error {
	tree, err := parse.Parse(pattern)
	if err != nil {
		return fmt.Errorf("invalid regex %q: %w", pattern, err)
	}
	dfa := automata.FromTree(tree).Determinize()
	matcher, err := search.NewMatcher(pattern)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(output, "Text: %q\nAccepted: %t\nContains match: %t\n", text, dfa.IsAccepted(text), matcher.ContainsMatch(text))
	return err
}

func printAutomata(pattern string, output io.Writer) error {
	tree, err := parse.Parse(pattern)
	if err != nil {
		return fmt.Errorf("invalid regex %q: %w", pattern, err)
	}
	nfa := automata.FromTree(tree)
	dfa := nfa.Determinize()
	_, err = fmt.Fprintf(output, "Syntax tree:\n%s\n\nNFA:\n%s\nDFA:\n%s", tree, nfa, dfa)
	return err
}
