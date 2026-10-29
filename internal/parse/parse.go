package parse

import "fmt"

// Operator / special-node codes (same idea as RegEx.java).
const (
	CONCAT            = 0xC04CA7
	STAR              = 0xE7011E
	PLUS              = 0xA11CE
	ALTERNATION       = 0xA17E54
	PROTECTION        = 0xBADDAD
	OPEN_PARENTHESIS  = 0x16641664
	CLOSE_PARENTHESIS = 0x51515151
	DOT               = 0xD07
	LITERAL           = -10
	CLASS             = -11
	OPTIONAL          = -12
	REPEAT            = -13
	BEGIN             = -14
	END               = -15
)

type RegExTree struct {
	Root     int
	SubTrees []*RegExTree
	Value    rune
	Ranges   []RuneRange
	Negated  bool
	Min      int
	Max      int
}

func newRegExTree(root int, subTrees []*RegExTree) *RegExTree {
	if subTrees == nil {
		subTrees = []*RegExTree{}
	}
	return &RegExTree{Root: root, SubTrees: subTrees}
}

func (t *RegExTree) String() string {
	if len(t.SubTrees) == 0 {
		return t.rootToString()
	}
	result := t.rootToString() + "(" + t.SubTrees[0].String()
	for i := 1; i < len(t.SubTrees); i++ {
		result += "," + t.SubTrees[i].String()
	}
	return result + ")"
}

func (t *RegExTree) rootToString() string {
	switch t.Root {
	case LITERAL:
		return string(t.Value)
	case CLASS:
		return t.classString()
	case OPTIONAL:
		return "?"
	case REPEAT:
		if t.Max < 0 {
			return fmt.Sprintf("{%d,}", t.Min)
		}
		if t.Min == t.Max {
			return fmt.Sprintf("{%d}", t.Min)
		}
		return fmt.Sprintf("{%d,%d}", t.Min, t.Max)
	case BEGIN:
		return "^"
	case END:
		return "$"
	case CONCAT:
		return "."
	case STAR:
		return "*"
	case PLUS:
		return "+"
	case ALTERNATION:
		return "|"
	case DOT:
		return "."
	default:
		return string(rune(t.Root))
	}
}

func Parse(regEx string) (*RegExTree, error) {
	if len(regEx) < 1 {
		return nil, fmt.Errorf("empty regEx")
	}
	result, err := tokenize(regEx)
	if err != nil {
		return nil, err
	}
	tree, err := parseTrees(result)
	if err != nil {
		return nil, err
	}
	if treeSize(tree) > 10000 {
		return nil, fmt.Errorf("pattern expansion exceeds 10000 states")
	}
	return tree, nil
}

func charToRoot(c byte) int {
	switch c {
	case '.':
		return DOT
	case '*':
		return STAR
	case '+':
		return PLUS
	case '?':
		return OPTIONAL
	case '^':
		return BEGIN
	case '$':
		return END
	case '|':
		return ALTERNATION
	case '(':
		return OPEN_PARENTHESIS
	case ')':
		return CLOSE_PARENTHESIS
	default:
		return int(c)
	}
}

func parseTrees(result []*RegExTree) (*RegExTree, error) {
	var err error
	for containParentheses(result) {
		result, err = processParentheses(result)
		if err != nil {
			return nil, err
		}
	}
	for containRepetition(result) {
		result, err = processRepetition(result)
		if err != nil {
			return nil, err
		}
	}
	for containConcat(result) {
		result, err = processConcat(result)
		if err != nil {
			return nil, err
		}
	}
	for containAlternation(result) {
		result, err = processAlternation(result)
		if err != nil {
			return nil, err
		}
	}

	if len(result) != 1 {
		return nil, fmt.Errorf("syntax error")
	}
	return removeProtection(result[0])
}

func containParentheses(trees []*RegExTree) bool {
	for _, t := range trees {
		if t.Root == CLOSE_PARENTHESIS || t.Root == OPEN_PARENTHESIS {
			return true
		}
	}
	return false
}

func processParentheses(trees []*RegExTree) ([]*RegExTree, error) {
	result := make([]*RegExTree, 0, len(trees))
	for i, tree := range trees {
		if tree.Root != CLOSE_PARENTHESIS {
			result = append(result, tree)
			continue
		}

		content := make([]*RegExTree, 0)
		for {
			if len(result) == 0 {
				return nil, fmt.Errorf("unmatched ')'")
			}

			last := result[len(result)-1]
			result = result[:len(result)-1]

			if last.Root == OPEN_PARENTHESIS {
				break
			}

			content = append([]*RegExTree{last}, content...)
		}

		sub, err := parseTrees(content)
		if err != nil {
			return nil, err
		}

		result = append(result, newRegExTree(PROTECTION, []*RegExTree{sub}))
		result = append(result, trees[i+1:]...)
		return result, nil
	}
	return nil, fmt.Errorf("unmatched '('")
}

func containRepetition(trees []*RegExTree) bool {
	for _, t := range trees {
		if isRepetition(t.Root) && len(t.SubTrees) == 0 {
			return true
		}
	}
	return false
}

func processRepetition(trees []*RegExTree) ([]*RegExTree, error) {
	result := make([]*RegExTree, 0, len(trees))
	for i, tree := range trees {
		if isRepetition(tree.Root) && len(tree.SubTrees) == 0 {
			if len(result) == 0 || result[len(result)-1].Root == ALTERNATION ||
				result[len(result)-1].Root == OPEN_PARENTHESIS ||
				result[len(result)-1].Root == BEGIN || result[len(result)-1].Root == END {
				return nil, fmt.Errorf("lonely %q", tree.rootToString())
			}
			last := result[len(result)-1]
			result = result[:len(result)-1]
			repetition := *tree
			repetition.SubTrees = []*RegExTree{last}
			result = append(result, &repetition)
			result = append(result, trees[i+1:]...)
			return result, nil
		}
		result = append(result, tree)
	}
	return result, nil
}

func containConcat(trees []*RegExTree) bool {
	for i := 1; i < len(trees); i++ {
		if trees[i-1].Root != ALTERNATION && trees[i].Root != ALTERNATION {
			return true
		}
	}
	return false
}

func processConcat(trees []*RegExTree) ([]*RegExTree, error) {
	for i := 1; i < len(trees); i++ {
		if trees[i-1].Root == ALTERNATION || trees[i].Root == ALTERNATION {
			continue
		}

		result := make([]*RegExTree, 0, len(trees)-1)
		result = append(result, trees[:i-1]...)
		result = append(result, newRegExTree(CONCAT, []*RegExTree{trees[i-1], trees[i]}))
		result = append(result, trees[i+1:]...)
		return result, nil
	}
	return trees, nil
}

func containAlternation(trees []*RegExTree) bool {
	for _, t := range trees {
		if t.Root == ALTERNATION && len(t.SubTrees) == 0 {
			return true
		}
	}
	return false
}

func processAlternation(trees []*RegExTree) ([]*RegExTree, error) {
	for i, tree := range trees {
		if tree.Root == ALTERNATION && len(tree.SubTrees) == 0 {
			if i == 0 {
				return nil, fmt.Errorf("lonely '|'")
			}
			if i == len(trees)-1 {
				return nil, fmt.Errorf("lonely '|'")
			}
			result := make([]*RegExTree, 0, len(trees)-2)
			result = append(result, trees[:i-1]...)
			result = append(result, newRegExTree(ALTERNATION, []*RegExTree{trees[i-1], trees[i+1]}))
			result = append(result, trees[i+2:]...)
			return result, nil
		}
	}
	return trees, nil
}

func removeProtection(tree *RegExTree) (*RegExTree, error) {
	if tree.Root == PROTECTION && len(tree.SubTrees) != 1 {
		return nil, fmt.Errorf("bad protection")
	}
	if len(tree.SubTrees) == 0 {
		return tree, nil
	}
	if tree.Root == PROTECTION {
		return removeProtection(tree.SubTrees[0])
	}

	subTrees := make([]*RegExTree, 0, len(tree.SubTrees))
	for _, t := range tree.SubTrees {
		cleaned, err := removeProtection(t)
		if err != nil {
			return nil, err
		}
		subTrees = append(subTrees, cleaned)
	}
	cleaned := *tree
	cleaned.SubTrees = subTrees
	return &cleaned, nil
}
