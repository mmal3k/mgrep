package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/mmal3k/mgrep/internal/automata"
	"github.com/mmal3k/mgrep/internal/parse"
	"github.com/mmal3k/mgrep/internal/search"
)

type options struct {
	algorithm string
	pattern   string
	text      *string
	path      string
	debug     bool
	format    string
	color     string
}

var errNoMatch = errors.New("no matching lines")

func main() {
	os.Exit(execute(os.Args[1:], os.Stdout, os.Stderr))
}

func execute(args []string, output io.Writer, diagnostics io.Writer) int {
	err := run(args, output, diagnostics)
	code := exitCode(err)
	if code == 2 {
		fmt.Fprintln(diagnostics, err)
	}
	return code
}

func exitCode(err error) int {
	if err == nil {
		return 0
	}
	if errors.Is(err, errNoMatch) {
		return 1
	}
	return 2
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
		return searchFile(opts, output)
	}
	if opts.text != nil {
		return testText(opts.pattern, *opts.text, opts.algorithm, output)
	}
	return nil
}

func readOptions(args []string, output io.Writer) (options, error) {
	var opts options
	flags := flag.NewFlagSet("mgrep", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	flags.StringVar(&opts.algorithm, "algorithm", "dfa", "Search algorithm: dfa, kmp, or boyer-moore (kmp and boyer-moore require literal ASCII letters and digits)")
	flags.StringVar(&opts.format, "format", "auto", "File output format: auto (pretty in a terminal), plain, or pretty")
	flags.StringVar(&opts.color, "color", "auto", "Highlight file matches: auto (terminal only), always, or never")
	flags.StringVar(&opts.pattern, "regex", "", "Regex to compile")
	flags.StringVar(&opts.path, "file", "", "Text file to search")
	flags.BoolVar(&opts.debug, "debug", false, "Print the syntax tree, NFA, DFA, and minimized DFA")
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
	if opts.format != "auto" && opts.format != "plain" && opts.format != "pretty" {
		return opts, fmt.Errorf("unknown format %q; choose auto, plain, or pretty", opts.format)
	}
	if opts.color != "auto" && opts.color != "always" && opts.color != "never" {
		return opts, fmt.Errorf("unknown color mode %q; choose auto, always, or never", opts.color)
	}
	if opts.pattern == "" {
		return opts, fmt.Errorf("provide a regex with -regex")
	}
	if opts.algorithm != "dfa" && opts.algorithm != "kmp" && opts.algorithm != "boyer-moore" {
		return opts, fmt.Errorf("unknown algorithm %q; choose dfa, kmp, or boyer-moore", opts.algorithm)
	}
	if opts.debug && opts.algorithm != "dfa" {
		return opts, fmt.Errorf("-debug displays automata and requires -algorithm dfa")
	}
	if opts.text != nil && opts.path != "" {
		return opts, fmt.Errorf("choose either -text or -file")
	}
	if opts.text == nil && opts.path == "" && !opts.debug {
		return opts, fmt.Errorf("choose -text, -file, or -debug")
	}
	return opts, nil
}

func newMatcher(pattern string, algorithm string) (search.TextMatcher, error) {
	if algorithm == "kmp" {
		return search.NewKMPMatcher(pattern)
	}
	if algorithm == "boyer-moore" {
		return search.NewBoyerMooreMatcher(pattern)
	}
	return search.NewMatcher(pattern)
}

func searchFile(opts options, output io.Writer) error {
	matcher, err := newMatcher(opts.pattern, opts.algorithm)
	if err != nil {
		return fmt.Errorf("invalid pattern %q: %w", opts.pattern, err)
	}
	terminal := isTerminal(output)
	style := search.LineStyle{Pretty: opts.format == "pretty" || opts.format == "auto" && terminal}
	_, noColor := os.LookupEnv("NO_COLOR")
	if opts.color == "always" || opts.color == "auto" && terminal && !noColor && os.Getenv("TERM") != "dumb" {
		style.Highlighter, err = search.NewHighlighter(opts.pattern, opts.algorithm)
		if err != nil {
			return err
		}
	}
	matches, err := search.SearchLines(opts.path, output, matcher, style)
	if err != nil {
		return err
	}
	if matches == 0 {
		return errNoMatch
	}
	return nil
}

func testText(pattern string, text string, algorithm string, output io.Writer) error {
	matcher, err := newMatcher(pattern, algorithm)
	if err != nil {
		return err
	}
	accepted := text == pattern
	if algorithm == "dfa" {
		tree, err := parse.Parse(pattern)
		if err != nil {
			return err
		}
		dfa := automata.FromTree(tree).Determinize().Minimize()
		accepted = dfa.IsAccepted(text)
	}
	matched := matcher.ContainsMatch(text)
	if _, err := fmt.Fprintf(output, "Text: %q\nAccepted: %t\nContains match: %t\n", text, accepted, matched); err != nil {
		return err
	}
	if !matched {
		return errNoMatch
	}
	return nil
}

func printAutomata(pattern string, output io.Writer) error {
	tree, err := parse.Parse(pattern)
	if err != nil {
		return fmt.Errorf("invalid regex %q: %w", pattern, err)
	}
	nfa := automata.FromTree(tree)
	dfa := nfa.Determinize()
	minimized := dfa.Minimize()
	_, err = fmt.Fprintf(output, "Syntax tree:\n%s\n\nNFA:\n%s\nDFA:\n%s\nMinimized DFA:\n%s", tree, nfa, dfa, minimized)
	return err
}
