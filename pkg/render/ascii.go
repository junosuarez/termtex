package render

import (
	"fmt"
	"strings"
	"termtex/pkg/tex"
)

// RenderASCII returns a Unicode / ASCII text representation of a math AST node.
func RenderASCII(node tex.Node) string {
	if node == nil {
		return ""
	}

	switch n := node.(type) {
	case *tex.CharNode:
		info := tex.GetSymbolInfo(n.Symbol)
		if info.Unicode != "" {
			return info.Unicode
		}
		return n.Symbol

	case *tex.GroupNode:
		var sb strings.Builder
		for _, child := range n.Children {
			sb.WriteString(RenderASCII(child))
		}
		return sb.String()

	case *tex.SubSupNode:
		base := RenderASCII(n.Base)
		sub := ""
		sup := ""
		if n.Sub != nil {
			sub = toSubscript(RenderASCII(n.Sub))
		}
		if n.Sup != nil {
			sup = toSuperscript(RenderASCII(n.Sup))
		}
		return base + sub + sup

	case *tex.FracNode:
		num := RenderASCII(n.Num)
		den := RenderASCII(n.Den)
		return fmt.Sprintf("(%s)/(%s)", num, den)

	case *tex.SqrtNode:
		content := RenderASCII(n.Content)
		if n.Index != nil {
			idx := RenderASCII(n.Index)
			return fmt.Sprintf("%s√( %s )", toSuperscript(idx), content)
		}
		return fmt.Sprintf("√( %s )", content)

	case *tex.BigOpNode:
		info := tex.GetSymbolInfo(n.Op)
		op := info.Unicode
		if op == "" {
			op = n.Op
		}
		under := ""
		over := ""
		if n.Under != nil {
			under = "_" + RenderASCII(n.Under)
		}
		if n.Over != nil {
			over = "^" + RenderASCII(n.Over)
		}
		return op + under + over

	case *tex.DelimNode:
		inner := RenderASCII(n.Inner)
		return n.Left + inner + n.Right

	case *tex.MatrixNode:
		var sb strings.Builder
		left := ""
		right := ""
		if n.Kind == "pmatrix" {
			left, right = "(", ")"
		} else if n.Kind == "bmatrix" {
			left, right = "[", "]"
		} else if n.Kind == "cases" {
			left = "{"
		}

		sb.WriteString(left)
		for r, row := range n.Rows {
			sb.WriteString("[ ")
			for c, cell := range row {
				sb.WriteString(RenderASCII(cell))
				if c < len(row)-1 {
					sb.WriteString("  ")
				}
			}
			sb.WriteString(" ]")
			if r < len(n.Rows)-1 {
				sb.WriteString(" ; ")
			}
		}
		sb.WriteString(right)
		return sb.String()

	case *tex.TextNode:
		return n.Text

	case *tex.SpaceNode:
		return " "
	}

	return ""
}

func toSubscript(s string) string {
	subMap := map[rune]rune{
		'0': '₀', '1': '₁', '2': '₂', '3': '₃', '4': '₄',
		'5': '₅', '6': '₆', '7': '₇', '8': '₈', '9': '₉',
		'+': '₊', '-': '₋', '=': '₌', '(': '₍', ')': '₎',
		'a': 'ₐ', 'e': 'ₑ', 'h': 'ₕ', 'i': 'ᵢ', 'j': 'ⱼ',
		'k': 'ₖ', 'l': 'ₗ', 'm': 'ₘ', 'n': 'ₙ', 'o': 'ₒ',
		'p': 'ₚ', 'r': 'ᵣ', 's': 'ₛ', 't': 'ₜ', 'u': 'ᵤ',
		'v': 'ᵥ', 'x': 'ₓ',
	}
	var res strings.Builder
	for _, r := range s {
		if sub, ok := subMap[r]; ok {
			res.WriteRune(sub)
		} else {
			res.WriteRune(r)
		}
	}
	return res.String()
}

func toSuperscript(s string) string {
	supMap := map[rune]rune{
		'0': '⁰', '1': '¹', '2': '²', '3': '³', '4': '⁴',
		'5': '⁵', '6': '⁶', '7': '⁷', '8': '⁸', '9': '⁹',
		'+': '⁺', '-': '⁻', '=': '⁼', '(': '⁽', ')': '⁾',
		'n': 'ⁿ', 'i': 'ⁱ', 'x': 'ˣ',
	}
	var res strings.Builder
	for _, r := range s {
		if sup, ok := supMap[r]; ok {
			res.WriteRune(sup)
		} else {
			res.WriteRune(r)
		}
	}
	return res.String()
}
