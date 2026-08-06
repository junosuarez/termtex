package tex

import (
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

	if brace, ok := base.(*UnderOverBraceNode); ok {
		if brace.Kind == "underbrace" && sub != nil {
			brace.Annotation = sub
			return brace, nil
		}
		if brace.Kind == "overbrace" && sup != nil {
			brace.Annotation = sup
			return brace, nil
		}
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

	case "\\binom":
		top, err := p.parseGroupOrAtom()
		if err != nil {
			return nil, err
		}
		bottom, err := p.parseGroupOrAtom()
		if err != nil {
			return nil, err
		}
		return &BinomNode{Top: top, Bottom: bottom}, nil

	case "\\underbrace":
		target, err := p.parseGroupOrAtom()
		if err != nil {
			return nil, err
		}
		return &UnderOverBraceNode{Kind: "underbrace", Target: target}, nil

	case "\\overbrace":
		target, err := p.parseGroupOrAtom()
		if err != nil {
			return nil, err
		}
		return &UnderOverBraceNode{Kind: "overbrace", Target: target}, nil

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
				p.pos += 6
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

		return &DelimNode{Left: leftDelim, Right: rightDelim, Inner: inner}, nil

	case "\\begin":
		p.skipWhitespace()
		if p.pos < len(p.input) && p.input[p.pos] == '{' {
			p.pos++
			envStart := p.pos
			for p.pos < len(p.input) && p.input[p.pos] != '}' {
				p.pos++
			}
			envKind := string(p.input[envStart:p.pos])
			if p.pos < len(p.input) && p.input[p.pos] == '}' {
				p.pos++
			}

			var rows [][]Node
			var currentRow []Node

			for p.pos < len(p.input) {
				p.skipWhitespace()
				if p.peekCommand() == "\\end" {
					if len(currentRow) > 0 {
						rows = append(rows, currentRow)
						currentRow = nil
					}
					p.pos += 4
					p.skipWhitespace()
					if p.pos < len(p.input) && p.input[p.pos] == '{' {
						p.pos++
						for p.pos < len(p.input) && p.input[p.pos] != '}' {
							p.pos++
						}
						if p.pos < len(p.input) && p.input[p.pos] == '}' {
							p.pos++
						}
					}
					break
				}

				cell, err := p.ParseExpr()
				if err != nil {
					return nil, err
				}
				if cell != nil {
					currentRow = append(currentRow, cell)
				}

				p.skipWhitespace()
				if p.pos < len(p.input) {
					if p.input[p.pos] == '&' {
						p.pos++
					} else if p.input[p.pos] == '\\' && p.pos+1 < len(p.input) && p.input[p.pos+1] == '\\' {
						p.pos += 2
						rows = append(rows, currentRow)
						currentRow = nil
					}
				}
			}

			return &MatrixNode{Kind: envKind, Rows: rows}, nil
		}

	case "\\text":
		p.skipWhitespace()
		if p.pos < len(p.input) && p.input[p.pos] == '{' {
			p.pos++
			txtStart := p.pos
			for p.pos < len(p.input) && p.input[p.pos] != '}' {
				p.pos++
			}
			txt := string(p.input[txtStart:p.pos])
			if p.pos < len(p.input) && p.input[p.pos] == '}' {
				p.pos++
			}
			return &TextNode{Text: txt, Style: "upright"}, nil
		}

	case "\\vec", "\\hat", "\\bar", "\\tilde", "\\dot", "\\ddot":
		target, err := p.parseGroupOrAtom()
		if err != nil {
			return nil, err
		}
		return &AccNode{Accent: cmd, Target: target}, nil

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

func (p *Parser) skipWhitespace() {
	for p.pos < len(p.input) && (p.input[p.pos] == ' ' || p.input[p.pos] == '\t' || p.input[p.pos] == '\n' || p.input[p.pos] == '\r') {
		p.pos++
	}
}

func (p *Parser) peekCommand() string {
	if p.pos >= len(p.input) || p.input[p.pos] != '\\' {
		return ""
	}
	start := p.pos + 1
	end := start
	for end < len(p.input) && unicode.IsLetter(p.input[end]) {
		end++
	}
	if end == start {
		return ""
	}
	return "\\" + string(p.input[start:end])
}
