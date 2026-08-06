package tex

import (
	"fmt"
	"strings"
	"unicode"
)

// Parser parses a TeX math string into an AST.
type Parser struct {
	input []rune
	pos   int
}

// NewParser creates a new Parser instance.
func NewParser(input string) *Parser {
	return &Parser{
		input: []rune(input),
		pos:   0,
	}
}

// Parse parses the full input and returns the root Node.
func Parse(input string) (Node, error) {
	p := NewParser(input)
	return p.ParseExpr()
}

// ParseExpr parses a sequence of math expressions until end of string or closing delimiter.
func (p *Parser) ParseExpr() (Node, error) {
	var nodes []Node

	for p.pos < len(p.input) {
		p.skipWhitespace()
		if p.pos >= len(p.input) {
			break
		}

		ch := p.input[p.pos]
		if ch == '}' || ch == '&' {
			break
		}

		if ch == '\\' {
			cmd := p.peekCommand()
			if cmd == "\\end" || cmd == "\\right" {
				break
			}
			// Check for row separator "\\"
			if p.pos+1 < len(p.input) && p.input[p.pos+1] == '\\' {
				break
			}
		}

		node, err := p.parseSingleNode()
		if err != nil {
			return nil, err
		}
		if node != nil {
			nodes = append(nodes, node)
		}
	}

	if len(nodes) == 0 {
		return &GroupNode{Children: nil}, nil
	}
	if len(nodes) == 1 {
		return nodes[0], nil
	}
	return &GroupNode{Children: nodes}, nil
}

// parseSingleNode parses a single item, including subscripts and superscripts.
func (p *Parser) parseSingleNode() (Node, error) {
	p.skipWhitespace()
	if p.pos >= len(p.input) {
		return nil, nil
	}

	base, err := p.parseAtom()
	if err != nil {
		return nil, err
	}
	if base == nil {
		return nil, nil
	}

	// Check for attached subscripts '_' or superscripts '^'
	var sub, sup Node
	for p.pos < len(p.input) {
		p.skipWhitespace()
		if p.pos >= len(p.input) {
			break
		}
		ch := p.input[p.pos]
		if ch == '_' && sub == nil {
			p.pos++
			s, err := p.parseGroupOrAtom()
			if err != nil {
				return nil, err
			}
			sub = s
		} else if ch == '^' && sup == nil {
			p.pos++
			s, err := p.parseGroupOrAtom()
			if err != nil {
				return nil, err
			}
			sup = s
		} else {
			break
		}
	}

	if bigOp, ok := base.(*BigOpNode); ok && (sub != nil || sup != nil) {
		bigOp.Under = sub
		bigOp.Over = sup
		return bigOp, nil
	}

	if sub != nil || sup != nil {
		return &SubSupNode{
			Base: base,
			Sub:  sub,
			Sup:  sup,
		}, nil
	}

	return base, nil
}

// parseAtom parses an individual TeX token or macro construct.
func (p *Parser) parseAtom() (Node, error) {
	p.skipWhitespace()
	if p.pos >= len(p.input) {
		return nil, nil
	}

	ch := p.input[p.pos]

	if ch == '{' {
		p.pos++
		group, err := p.ParseExpr()
		if err != nil {
			return nil, err
		}
		p.skipWhitespace()
		if p.pos < len(p.input) && p.input[p.pos] == '}' {
			p.pos++
		}
		return group, nil
	}

	if ch == '\\' {
		return p.parseMacro()
	}

	p.pos++
	sym := string(ch)
	fontFamily := "math-italic"
	isOp := false

	if unicode.IsDigit(ch) || strings.ContainsRune("+-=()[]!?,.:;|", ch) {
		fontFamily = "math-upright"
		if strings.ContainsRune("+-=,", ch) {
			isOp = true
		}
	}

	return &CharNode{
		Symbol:     sym,
		FontFamily: fontFamily,
		IsOperator: isOp,
	}, nil
}

func (p *Parser) parseGroupOrAtom() (Node, error) {
	p.skipWhitespace()
	if p.pos >= len(p.input) {
		return &GroupNode{}, nil
	}
	if p.input[p.pos] == '{' {
		p.pos++
		node, err := p.ParseExpr()
		if err != nil {
			return nil, err
		}
		p.skipWhitespace()
		if p.pos < len(p.input) && p.input[p.pos] == '}' {
			p.pos++
		}
		return node, nil
	}
	if p.input[p.pos] == '\\' {
		return p.parseMacro()
	}
	ch := p.input[p.pos]
	p.pos++
	fontFamily := "math-italic"
	if unicode.IsDigit(ch) {
		fontFamily = "math-upright"
	}
	return &CharNode{
		Symbol:     string(ch),
		FontFamily: fontFamily,
	}, nil
}

func (p *Parser) parseMacro() (Node, error) {
	p.pos++ // consume '\'
	if p.pos >= len(p.input) {
		return &CharNode{Symbol: "\\"}, nil
	}

	ch := p.input[p.pos]
	if !unicode.IsLetter(ch) {
		p.pos++
		switch ch {
		case '{', '}', '_', '^', '%', '$', '#', '&':
			return &CharNode{Symbol: string(ch), FontFamily: "math-upright"}, nil
		case ',', ' ', ';':
			return &SpaceNode{Width: 0.2}, nil
		case '!':
			return &SpaceNode{Width: -0.1}, nil
		case '\\':
			return &SpaceNode{Width: 0.5}, nil
		}
		return &CharNode{Symbol: string(ch), FontFamily: "math-upright"}, nil
	}

	start := p.pos
	for p.pos < len(p.input) && unicode.IsLetter(p.input[p.pos]) {
		p.pos++
	}
	cmd := "\\" + string(p.input[start:p.pos])

	switch cmd {
	case "\\frac":
		num, err := p.parseGroupOrAtom()
		if err != nil {
			return nil, err
		}
		den, err := p.parseGroupOrAtom()
		if err != nil {
			return nil, err
		}
		return &FracNode{Num: num, Den: den}, nil

	case "\\sqrt":
		var index Node
		p.skipWhitespace()
		if p.pos < len(p.input) && p.input[p.pos] == '[' {
			p.pos++
			idxStart := p.pos
			for p.pos < len(p.input) && p.input[p.pos] != ']' {
				p.pos++
			}
			idxStr := string(p.input[idxStart:p.pos])
			if p.pos < len(p.input) && p.input[p.pos] == ']' {
				p.pos++
			}
			index, _ = Parse(idxStr)
		}
		content, err := p.parseGroupOrAtom()
		if err != nil {
			return nil, err
		}
		return &SqrtNode{Index: index, Content: content}, nil

	case "\\left":
		p.skipWhitespace()
		leftDelim := ""
		if p.pos < len(p.input) {
			if p.input[p.pos] == '\\' {
				p.pos++
				if p.pos < len(p.input) {
					leftDelim = "\\" + string(p.input[p.pos])
					p.pos++
				}
			} else {
				leftDelim = string(p.input[p.pos])
				p.pos++
			}
		}

		inner, err := p.ParseExpr()
		if err != nil {
			return nil, err
		}

		p.skipWhitespace()
		rightDelim := ""
		if p.pos < len(p.input) && p.input[p.pos] == '\\' {
			if p.peekCommand() == "\\right" {
				p.pos += 6 // consume \right
				p.skipWhitespace()
				if p.pos < len(p.input) {
					if p.input[p.pos] == '\\' {
						p.pos++
						if p.pos < len(p.input) {
							rightDelim = "\\" + string(p.input[p.pos])
							p.pos++
						}
					} else {
						rightDelim = string(p.input[p.pos])
						p.pos++
					}
				}
			}
		}

		return &DelimNode{
			Left:  leftDelim,
			Right: rightDelim,
			Inner: inner,
		}, nil

	case "\\begin":
		envName, err := p.parseGroupString()
		if err != nil {
			return nil, err
		}
		return p.parseEnvironment(envName)

	case "\\hat", "\\vec", "\\bar", "\\tilde", "\\dot", "\\ddot":
		target, err := p.parseGroupOrAtom()
		if err != nil {
			return nil, err
		}
		return &AccNode{Accent: cmd, Target: target}, nil

	case "\\mathbf", "\\mathrm", "\\text":
		text, err := p.parseGroupString()
		if err != nil {
			return nil, err
		}
		style := "bold"
		if cmd == "\\mathrm" || cmd == "\\text" {
			style = "normal"
		}
		return &TextNode{Text: text, Style: style}, nil

	case "\\quad":
		return &SpaceNode{Width: 1.0}, nil
	case "\\qquad":
		return &SpaceNode{Width: 2.0}, nil
	}

	info := GetSymbolInfo(cmd)
	if info.IsBigOp {
		return &BigOpNode{Op: cmd}, nil
	}

	return &CharNode{
		Symbol:     cmd,
		FontFamily: "symbol",
		IsOperator: info.IsOperator,
	}, nil
}

func (p *Parser) parseGroupString() (string, error) {
	p.skipWhitespace()
	if p.pos >= len(p.input) || p.input[p.pos] != '{' {
		return "", fmt.Errorf("expected '{' at pos %d", p.pos)
	}
	p.pos++
	start := p.pos
	depth := 1
	for p.pos < len(p.input) {
		if p.input[p.pos] == '{' {
			depth++
		} else if p.input[p.pos] == '}' {
			depth--
			if depth == 0 {
				res := string(p.input[start:p.pos])
				p.pos++
				return res, nil
			}
		}
		p.pos++
	}
	return string(p.input[start:]), nil
}

func (p *Parser) parseEnvironment(envName string) (Node, error) {
	var rows [][]Node
	var currentRow []Node

	for p.pos < len(p.input) {
		p.skipWhitespace()
		if p.peekCommand() == "\\end" {
			p.pos += 4
			p.parseGroupString() // consume {envName}
			break
		}

		cell, err := p.ParseExpr()
		if err != nil {
			return nil, err
		}
		currentRow = append(currentRow, cell)

		p.skipWhitespace()
		if p.pos < len(p.input) {
			if p.input[p.pos] == '&' {
				p.pos++
			} else if p.pos+1 < len(p.input) && p.input[p.pos] == '\\' && p.input[p.pos+1] == '\\' {
				p.pos += 2
				rows = append(rows, currentRow)
				currentRow = nil
			}
		}
	}

	if len(currentRow) > 0 {
		rows = append(rows, currentRow)
	}

	return &MatrixNode{
		Kind: envName,
		Rows: rows,
	}, nil
}

func (p *Parser) peekCommand() string {
	if p.pos >= len(p.input) || p.input[p.pos] != '\\' {
		return ""
	}
	i := p.pos + 1
	for i < len(p.input) && unicode.IsLetter(p.input[i]) {
		i++
	}
	return "\\" + string(p.input[p.pos+1:i])
}

func (p *Parser) skipWhitespace() {
	for p.pos < len(p.input) && (p.input[p.pos] == ' ' || p.input[p.pos] == '\t' || p.input[p.pos] == '\n' || p.input[p.pos] == '\r') {
		p.pos++
	}
}
