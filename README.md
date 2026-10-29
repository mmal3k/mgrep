# mgrep

A Go command-line tool for searching text files with a small regular-expression engine implemented from scratch.

It parses a pattern into a syntax tree, builds a Thompson epsilon-NFA, determinizes it into a DFA, and minimizes the DFA with partition refinement. For substring searches, it compiles `.*(pattern)` once and scans each line until a match is found. Matching lines are printed once, in file order. Terminal output adds line numbers and match highlighting; redirected output stays plain by default.

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
./bin/mgrep -regex 'S(a|g|r)+on' -file book.txt
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

The default algorithm is `dfa`, which supports the regex operators below. Select `kmp` for literal patterns containing ASCII letters and digits:

```bash
./bin/mgrep -algorithm kmp -regex 'Sargon' -file internal/search/testdata/sample.txt
./bin/mgrep -algorithm kmp -regex 'Sargon' -text 'Hello Sargon'
```

The fallback table is built once per pattern and reused across file lines. KMP scans in linear time; its current educational table-construction code is not linear-time preprocessing yet. KMP does not interpret regex operators, so nonliteral patterns are rejected. Automaton debug output is available only with `-algorithm dfa`.

Boyer–Moore also supports nonempty patterns containing literal ASCII letters and digits:

```bash
./bin/mgrep -algorithm boyer-moore -regex 'Sargon' -file internal/search/testdata/sample.txt
./bin/mgrep -algorithm boyer-moore -regex 'Sargon' -text 'Hello Sargon'
```

For example, `-regex '123'` and `-regex 'a1B2'` work with either literal matcher.

It compares from right to left and combines bad-character and strong good-suffix shifts. Both tables are built once and reused across lines.

## Supported patterns

- Literal digits, punctuation, and Unicode characters, with case-sensitive matching
- Concatenation: `ab`
- Alternation: `a|b`
- Parentheses: `(a|b)c`
- Zero or more repetitions: `a*`
- One or more repetitions: `a+`, including groups such as `(ab)+`
- Optional expressions: `a?`
- Repetition bounds: `a{3}`, `a{2,4}`, `a{2,}`
- Bracket lists, ranges, and negation: `[abc]`, `[a-z]`, `[^abc]`
- Line anchors: `^` and `$`
- Escaped punctuation: `a\.b`
- Any one character: `.`

The DFA engine decodes input text as Unicode runes. KMP and Boyer–Moore compare bytes, which supports exact searches for their ASCII alphanumeric patterns in UTF-8 text. Search operates on individual file lines; matches do not cross line boundaries.

This is a restricted regex implementation, not a full replacement for `grep -E`. POSIX named/collating classes and backreferences are not supported.

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

- File reading has no fixed line-length limit, but memory must hold the current line and any highlighting data.
- Input is supplied with `-text` or `-file`; stdin and multiple-file search are not implemented.
- A completed search exits successfully even when no lines match; this differs from grep's no-match exit status.
- Determinization may produce many states for complex patterns; benchmarks and resource limits are future work.

## Readable terminal output

Highlighting uses the selected algorithm: DFA for regex patterns, KMP for
KMP searches, and Boyer–Moore for Boyer–Moore searches.

File searches automatically show line numbers, separate matching lines with a
blank line, and highlight matched text in bold red when writing to a terminal.
Regex highlighting selects the leftmost, longest nonoverlapping matches; empty
matches still select lines but have no visible highlight. Input lines remain
complete, including long paragraphs.

Redirected or piped output stays plain by default. Control formatting explicitly:

```bash
go build -o bin/mgrep .
./bin/mgrep -regex 'S(a|g|r)+on' -file book.txt
./bin/mgrep -regex 'Sargon' -file book.txt -format pretty -color always
./bin/mgrep -regex 'Sargon' -file book.txt -format plain -color never
```

`-format` accepts `auto`, `plain`,
or `pretty`; `-color` accepts `auto`, `always`, or `never`. `NO_COLOR` and
`TERM=dumb` disable automatic highlighting. Formatting and highlighting apply
to file-search output; `-text` keeps its acceptance summary.

## Extended DFA pattern syntax

The default DFA engine supports digits, punctuation, and Unicode literals in
addition to the assignment's ASCII-letter subset. These examples work with
`-regex`:

| Syntax | Example | Meaning |
|---|---|---|
| Literal text | `abc123`, `hello!`, `café` | Match the given characters |
| Character list | `[abc]` | One of the listed characters |
| Range | `[a-z]`, `[0-9]` | One character in the range |
| Negated class | `[^abc]` | One character outside the class |
| Optional | `colou?r` | `color` or `colour` |
| Exact count | `a{3}` | Exactly three repetitions |
| Bounded count | `a{2,4}` | Two through four repetitions |
| Minimum count | `a{2,}` | At least two repetitions |
| Start/end anchors | `^cat`, `cat$`, `^cat$` | Restrict a match to line boundaries |
| Escaped operator | `a\.b` | Match a literal dot |

Groups can also be repeated, for example `(ab){2,4}`. Quote patterns in the
shell, using single quotes to preserve backslashes:

```bash
./bin/mgrep -regex '^S[a-z]+[0-9]*!$' -text 'Sargon123!'
./bin/mgrep -regex 'a\.b' -text 'a.b'
```

KMP and Boyer–Moore retain their literal ASCII-letter-and-digit pattern syntax.
Ranges use Unicode code-point order rather than locale collation. POSIX named
classes such as `[[:digit:]]`, collating/equivalence classes, backreferences,
and shorthand escapes such as `\d` are not implemented. Inside bracket
expressions, backslash is a literal character; put `]` first and `-` first or
last when matching those characters literally.

Repetition bounds are limited to 1000 and estimated expanded pattern size to
10000 states. Unbounded repetition such as `a{2,}` has no maximum match length.
The NFA, DFA, minimized DFA, line matcher, and highlighter support these
extensions. Compatibility tests compare CLI output with `grep -E` in the C
locale; Unicode behavior is also checked against a reference in unit tests.
