package parse

import (
	"fmt"
	"strconv"
	"strings"
	"unicode/utf8"
)

type RuneRange struct{ Lo, Hi rune }

func isRepetition(root int) bool {
	return root == STAR || root == PLUS || root == OPTIONAL || root == REPEAT
}

func tokenize(pattern string) ([]*RegExTree, error) {
	if !utf8.ValidString(pattern) {
		return nil, fmt.Errorf("pattern is not valid UTF-8")
	}
	letters := []rune(pattern)
	var tokens []*RegExTree
	for i := 0; i < len(letters); i++ {
		c := letters[i]
		switch c {
		case '\\':
			i++
			if i == len(letters) {
				return nil, fmt.Errorf("trailing escape")
			}
			if letters[i] >= '0' && letters[i] <= '9' || letters[i] >= 'a' && letters[i] <= 'z' || letters[i] >= 'A' && letters[i] <= 'Z' {
				return nil, fmt.Errorf("unsupported escape \\%c; escape punctuation literally", letters[i])
			}
			tokens = append(tokens, &RegExTree{Root: LITERAL, Value: letters[i]})
		case '[':
			tree, end, err := parseClass(letters, i+1)
			if err != nil {
				return nil, err
			}
			tokens = append(tokens, tree)
			i = end
		case '{':
			end := i + 1
			for end < len(letters) && letters[end] != '}' {
				end++
			}
			if end == len(letters) {
				return nil, fmt.Errorf("unclosed repetition bounds")
			}
			parts := strings.Split(string(letters[i+1:end]), ",")
			if len(parts) > 2 || parts[0] == "" {
				return nil, fmt.Errorf("invalid repetition bounds")
			}
			count := func(s string) (int, error) {
				for _, c := range s {
					if c < '0' || c > '9' {
						return 0, fmt.Errorf("invalid repetition count %q", s)
					}
				}
				n, err := strconv.Atoi(s)
				if err != nil || n > 1000 {
					return 0, fmt.Errorf("repetition count must be between 0 and 1000")
				}
				return n, nil
			}
			min, err := count(parts[0])
			if err != nil {
				return nil, err
			}
			max := min
			if len(parts) == 2 {
				max = -1
				if parts[1] != "" {
					max, err = count(parts[1])
					if err != nil {
						return nil, err
					}
				}
			}
			if max >= 0 && max < min {
				return nil, fmt.Errorf("repetition maximum is smaller than minimum")
			}
			tokens = append(tokens, &RegExTree{Root: REPEAT, Min: min, Max: max})
			i = end
		case '.', '*', '+', '?', '|', '(', ')', '^', '$':
			tokens = append(tokens, newRegExTree(charToRoot(byte(c)), nil))
		default:
			tokens = append(tokens, &RegExTree{Root: LITERAL, Value: c})
		}
	}
	return tokens, nil
}

func parseClass(letters []rune, start int) (*RegExTree, int, error) {
	tree := &RegExTree{Root: CLASS}
	i := start
	if i < len(letters) && letters[i] == '^' {
		tree.Negated = true
		i++
	}
	var items []rune
	for ; i < len(letters); i++ {
		c := letters[i]
		if c == ']' && len(items) > 0 {
			break
		}
		if c == '[' && i+1 < len(letters) && strings.ContainsRune(":.=", letters[i+1]) {
			return nil, 0, fmt.Errorf("POSIX named classes and collation constructs are not supported")
		}
		items = append(items, c)
	}
	if i == len(letters) {
		return nil, 0, fmt.Errorf("unclosed character class")
	}
	for j := 0; j < len(items); j++ {
		lo, hi := items[j], items[j]
		if j+2 < len(items) && items[j+1] == '-' {
			hi = items[j+2]
			if hi < lo {
				return nil, 0, fmt.Errorf("reversed character range %c-%c", lo, hi)
			}
			j += 2
		}
		tree.Ranges = append(tree.Ranges, RuneRange{Lo: lo, Hi: hi})
	}
	return tree, i, nil
}

func (t *RegExTree) classString() string {
	var out strings.Builder
	out.WriteByte('[')
	if t.Negated {
		out.WriteByte('^')
	}
	for _, r := range t.Ranges {
		out.WriteRune(r.Lo)
		if r.Lo != r.Hi {
			out.WriteByte('-')
			out.WriteRune(r.Hi)
		}
	}
	out.WriteByte(']')
	return out.String()
}

func treeSize(t *RegExTree) int {
	size := 2
	for _, sub := range t.SubTrees {
		size += treeSize(sub)
		if size > 10000 {
			return 10001
		}
	}
	if t.Root == REPEAT {
		copies := t.Max
		if copies < 0 {
			copies = t.Min + 1
		}
		if copies > 0 && size > 10000/copies {
			return 10001
		}
		size *= copies
	}
	return size
}
