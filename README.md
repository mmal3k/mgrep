# mgrep

A Go command-line tool for searching text files with a small regular-expression engine implemented from scratch.

It parses a pattern into a syntax tree, builds a Thompson epsilon-NFA, determinizes it into a DFA, and minimizes the DFA with partition refinement. For substring searches, it compiles `.*(pattern)` once and scans each line until a match is found. Matching lines are printed once, in file order.

## Requirements

Go 1.26.5 or newer. There are no third-party Go dependencies.

## Build

```bash
go build -o bin/mgrep .
```

## Usage

Search a file:

```bash
./bin/mgrep -regex 'ab' -file internal/search/testdata/sample.txt
```

A line such as `xxabyy` matches because it contains `ab`.

Test a string:

```bash
./bin/mgrep -regex 'ab' -text 'xxabyy'
```

```text
Text: "xxabyy"
Accepted: false
Contains match: true
```

`Accepted` checks the entire string. `Contains match` checks for a matching substring.

Test an empty string:

```bash
./bin/mgrep -regex 'a*' -text ''
```

Inspect the syntax tree and automata:

```bash
./bin/mgrep -regex 'a|bc*' -debug
```

Debug output includes the syntax tree, epsilon-NFA, DFA, and minimized DFA. It goes to stderr, so it can also be enabled during a file search without mixing it into stdout. Use `-h` for the flag descriptions. All examples can also be run with `go run .` instead of `./bin/mgrep`.

## Algorithms

The default algorithm is `dfa`, which supports the regex operators below. Select `kmp` for literal patterns containing ASCII letters:

```bash
./bin/mgrep -algorithm kmp -regex 'Sargon' -file internal/search/testdata/sample.txt
./bin/mgrep -algorithm kmp -regex 'Sargon' -text 'Hello Sargon'
```

The fallback table is built once per pattern and reused across file lines. KMP scans in linear time; its current educational table-construction code is not linear-time preprocessing yet. KMP does not interpret regex operators, so nonliteral patterns are rejected. Automaton debug output is available only with `-algorithm dfa`.

Boyer–Moore also supports nonempty literal ASCII-letter patterns:

```bash
./bin/mgrep -algorithm boyer-moore -regex 'Sargon' -file internal/search/testdata/sample.txt
./bin/mgrep -algorithm boyer-moore -regex 'Sargon' -text 'Hello Sargon'
```

It compares from right to left and combines bad-character and strong good-suffix shifts. Both tables are built once and reused across lines.

## Supported patterns

- ASCII letters, with case-sensitive matching
- Concatenation: `ab`
- Alternation: `a|b`
- Parentheses: `(a|b)c`
- Zero or more repetitions: `a*`
- Any one character: `.`

The DFA engine decodes input text as Unicode runes. KMP and Boyer–Moore compare bytes, which supports exact searches for their ASCII-letter patterns in UTF-8 text. Search operates on individual file lines; matches do not cross line boundaries.

This is a restricted regex implementation, not a full replacement for `grep -E`. Character classes, `+`, `?`, anchors, escaping, and backreferences are not supported.

## Tests

```bash
go test ./...
go vet ./...
```

Tests cover parsing, automaton construction, determinization, minimization (including language equivalence and state distinguishability), whole-string acceptance, substring matching, file output, and CLI arguments.

Compare prepared matchers using synthetic late-match, absent-pattern, and repeated-suffix inputs:

```bash
go test ./internal/search -run '^$' -bench '^BenchmarkLiteralSearch$' -benchmem
```

These benchmarks exclude pattern compilation and file I/O; they do not measure a Gutenberg corpus.

## Layout

```text
main.go             CLI options and command helpers
internal/parse/     Regex parsing and syntax trees
internal/automata/  NFA/DFA construction, minimization, and acceptance
internal/search/    Compiled search matcher and line processing
```

## Current limitations

- File scanning currently uses Go Scanner's default limit of roughly 64 KiB per line; longer lines return an error.
- Input is supplied with `-text` or `-file`; stdin and multiple-file search are not implemented.
- A completed search exits successfully even when no lines match; this differs from grep's no-match exit status.
- Determinization may produce many states for complex patterns; benchmarks and resource limits are future work.
