package tex

import (
	"math"
)

// Box represents a laid-out box in TeX math layout coordinates.
type Box struct {
	Type       string   // "char", "group", "frac", "rule", "sqrt", "op", "delim", "matrix", "text"
	X, Y       float64  // Relative positioning
	Width      float64  // Width in em
	Height     float64  // Height above baseline in em
	Depth      float64  // Depth below baseline in em
	Text       string   // Text or character content
	FontFamily string   // Font style
	Scale      float64  // Scale factor (1.0 for normal, 0.7 for sub/sup)
	Children   []*Box   // Nested laid out boxes
}

// TotalHeight returns Height + Depth.
func (b *Box) TotalHeight() float64 {
	return b.Height + b.Depth
}

// LayoutEngine handles converting AST nodes into laid out Boxes.
type LayoutEngine struct {
	DisplayMode bool
}

// NewLayoutEngine creates a layout engine instance.
func NewLayoutEngine(displayMode bool) *LayoutEngine {
	return &LayoutEngine{DisplayMode: displayMode}
}

// BuildLayout computes the Box tree from an AST Node.
func (e *LayoutEngine) BuildLayout(node Node, scale float64) *Box {
	if node == nil {
		return &Box{Width: 0, Height: 0, Depth: 0, Scale: scale}
	}

	switch n := node.(type) {
	case *CharNode:
		info := GetSymbolInfo(n.Symbol)
		w := info.WidthEm * scale
		h := 0.75 * scale
		d := 0.2 * scale

		text := info.Unicode
		if text == "" {
			text = n.Symbol
		}

		box := &Box{
			Type:       "char",
			Width:      w,
			Height:     h,
			Depth:      d,
			Text:       text,
			FontFamily: n.FontFamily,
			Scale:      scale,
		}

		if n.IsOperator {
			// Add extra side padding for operators (TeX spacing rule)
			pad := 0.25 * scale
			gBox := &Box{
				Type:     "group",
				Width:    w + 2*pad,
				Height:   h,
				Depth:    d,
				Scale:    scale,
				Children: []*Box{box},
			}
			box.X = pad
			return gBox
		}
		return box

	case *GroupNode:
		gBox := &Box{Type: "group", Scale: scale}
		curX := 0.0
		maxH := 0.0
		maxD := 0.0

		for _, child := range n.Children {
			cBox := e.BuildLayout(child, scale)
			cBox.X = curX
			curX += cBox.Width
			if cBox.Height > maxH {
				maxH = cBox.Height
			}
			if cBox.Depth > maxD {
				maxD = cBox.Depth
			}
			gBox.Children = append(gBox.Children, cBox)
		}

		gBox.Width = curX
		gBox.Height = maxH
		gBox.Depth = maxD
		return gBox

	case *SubSupNode:
		baseBox := e.BuildLayout(n.Base, scale)
		subScale := scale * 0.7

		var subBox, supBox *Box
		if n.Sub != nil {
			subBox = e.BuildLayout(n.Sub, subScale)
		}
		if n.Sup != nil {
			supBox = e.BuildLayout(n.Sup, subScale)
		}

		maxSubSupW := 0.0
		if subBox != nil && subBox.Width > maxSubSupW {
			maxSubSupW = subBox.Width
		}
		if supBox != nil && supBox.Width > maxSubSupW {
			maxSubSupW = supBox.Width
		}

		gBox := &Box{
			Type:     "group",
			Width:    baseBox.Width + maxSubSupW,
			Height:   baseBox.Height,
			Depth:    baseBox.Depth,
			Scale:    scale,
			Children: []*Box{baseBox},
		}

		if supBox != nil {
			supBox.X = baseBox.Width
			supBox.Y = -(baseBox.Height * 0.6) // shift up
			gBox.Children = append(gBox.Children, supBox)
			if supH := -supBox.Y + supBox.Height; supH > gBox.Height {
				gBox.Height = supH
			}
		}

		if subBox != nil {
			subBox.X = baseBox.Width
			subBox.Y = baseBox.Depth * 0.6 // shift down
			gBox.Children = append(gBox.Children, subBox)
			if subD := subBox.Y + subBox.Depth; subD > gBox.Depth {
				gBox.Depth = subD
			}
		}

		return gBox

	case *FracNode:
		numBox := e.BuildLayout(n.Num, scale*0.85)
		denBox := e.BuildLayout(n.Den, scale*0.85)

		ruleThickness := 0.06 * scale
		padding := 0.2 * scale

		maxW := math.Max(numBox.Width, denBox.Width) + 2*padding

		numBox.X = (maxW - numBox.Width) / 2
		denBox.X = (maxW - denBox.Width) / 2

		axisY := 0.35 * scale // Math axis height above baseline

		numBox.Y = -(axisY + padding + numBox.Depth)
		denBox.Y = axisY + padding + denBox.Height

		ruleBox := &Box{
			Type:   "rule",
			X:      0,
			Y:      -axisY,
			Width:  maxW,
			Height: ruleThickness / 2,
			Depth:  ruleThickness / 2,
			Scale:  scale,
		}

		h := -numBox.Y + numBox.Height
		d := denBox.Y + denBox.Depth

		return &Box{
			Type:     "frac",
			Width:    maxW,
			Height:   h,
			Depth:    d,
			Scale:    scale,
			Children: []*Box{numBox, denBox, ruleBox},
		}

	case *SqrtNode:
		contentBox := e.BuildLayout(n.Content, scale)
		padding := 0.15 * scale
		ruleThickness := 0.06 * scale

		w := contentBox.Width + 0.6*scale + padding
		h := contentBox.Height + 0.25*scale
		d := contentBox.Depth

		contentBox.X = 0.5 * scale
		contentBox.Y = 0

		sqrtSymBox := &Box{
			Type:       "char",
			X:          0,
			Y:          0,
			Width:      0.5 * scale,
			Height:     h,
			Depth:      d,
			Text:       "√",
			FontFamily: "symbol",
			Scale:      scale,
		}

		overBarBox := &Box{
			Type:   "rule",
			X:      0.45 * scale,
			Y:      -h,
			Width:  contentBox.Width + padding,
			Height: ruleThickness,
			Depth:  0,
			Scale:  scale,
		}

		return &Box{
			Type:     "sqrt",
			Width:    w,
			Height:   h,
			Depth:    d,
			Scale:    scale,
			Children: []*Box{sqrtSymBox, contentBox, overBarBox},
		}

	case *BigOpNode:
		info := GetSymbolInfo(n.Op)
		opUnicode := info.Unicode
		if opUnicode == "" {
			opUnicode = n.Op
		}

		opScale := scale * 1.5
		opW := 1.0 * scale
		opH := 0.9 * scale
		opD := 0.3 * scale

		opBox := &Box{
			Type:       "char",
			X:          0,
			Y:          0,
			Width:      opW,
			Height:     opH,
			Depth:      opD,
			Text:       opUnicode,
			FontFamily: "symbol",
			Scale:      opScale,
		}

		limScale := scale * 0.7
		var underBox, overBox *Box
		if n.Under != nil {
			underBox = e.BuildLayout(n.Under, limScale)
		}
		if n.Over != nil {
			overBox = e.BuildLayout(n.Over, limScale)
		}

		maxW := opW
		if underBox != nil && underBox.Width > maxW {
			maxW = underBox.Width
		}
		if overBox != nil && overBox.Width > maxW {
			maxW = overBox.Width
		}

		opBox.X = (maxW - opW) / 2
		totalH := opH
		totalD := opD

		children := []*Box{opBox}

		if overBox != nil {
			overBox.X = (maxW - overBox.Width) / 2
			overBox.Y = -(opH + 0.15*scale + overBox.Depth)
			children = append(children, overBox)
			totalH = -overBox.Y + overBox.Height
		}

		if underBox != nil {
			underBox.X = (maxW - underBox.Width) / 2
			underBox.Y = opD + 0.15*scale + underBox.Height
			children = append(children, underBox)
			totalD = underBox.Y + underBox.Depth
		}

		return &Box{
			Type:     "op",
			Width:    maxW,
			Height:   totalH,
			Depth:    totalD,
			Scale:    scale,
			Children: children,
		}

	case *DelimNode:
		innerBox := e.BuildLayout(n.Inner, scale)
		h := innerBox.Height
		d := innerBox.Depth

		delimW := 0.45 * scale
		children := []*Box{}
		curX := 0.0

		if n.Left != "" && n.Left != "." {
			leftSym := GetSymbolInfo(n.Left).Unicode
			if leftSym == "" {
				leftSym = n.Left
			}
			leftBox := &Box{
				Type:       "char",
				X:          0,
				Y:          0,
				Width:      delimW,
				Height:     h,
				Depth:      d,
				Text:       leftSym,
				FontFamily: "symbol",
				Scale:      scale,
			}
			children = append(children, leftBox)
			curX += delimW
		}

		innerBox.X = curX
		children = append(children, innerBox)
		curX += innerBox.Width

		if n.Right != "" && n.Right != "." {
			rightSym := GetSymbolInfo(n.Right).Unicode
			if rightSym == "" {
				rightSym = n.Right
			}
			rightBox := &Box{
				Type:       "char",
				X:          curX,
				Y:          0,
				Width:      delimW,
				Height:     h,
				Depth:      d,
				Text:       rightSym,
				FontFamily: "symbol",
				Scale:      scale,
			}
			children = append(children, rightBox)
			curX += delimW
		}

		return &Box{
			Type:     "delim",
			Width:    curX,
			Height:   h,
			Depth:    d,
			Scale:    scale,
			Children: children,
		}

	case *MatrixNode:
		if n.Kind == "pmatrix" {
			plainMatrix := *n
			plainMatrix.Kind = "matrix"
			return e.BuildLayout(&DelimNode{Left: "(", Right: ")", Inner: &plainMatrix}, scale)
		} else if n.Kind == "bmatrix" {
			plainMatrix := *n
			plainMatrix.Kind = "matrix"
			return e.BuildLayout(&DelimNode{Left: "[", Right: "]", Inner: &plainMatrix}, scale)
		} else if n.Kind == "cases" {
			plainMatrix := *n
			plainMatrix.Kind = "matrix"
			return e.BuildLayout(&DelimNode{Left: "{", Right: "", Inner: &plainMatrix}, scale)
		}

		cellPaddingX := 0.4 * scale
		cellPaddingY := 0.3 * scale

		numRows := len(n.Rows)
		if numRows == 0 {
			return &Box{Width: 0, Height: 0, Depth: 0, Scale: scale}
		}

		numCols := 0
		for _, r := range n.Rows {
			if len(r) > numCols {
				numCols = len(r)
			}
		}

		colWidths := make([]float64, numCols)
		rowHeights := make([]float64, numRows)
		rowDepths := make([]float64, numRows)

		grid := make([][]*Box, numRows)

		for r, row := range n.Rows {
			grid[r] = make([]*Box, numCols)
			for c, cellNode := range row {
				cellBox := e.BuildLayout(cellNode, scale)
				grid[r][c] = cellBox
				if cellBox.Width > colWidths[c] {
					colWidths[c] = cellBox.Width
				}
				if cellBox.Height > rowHeights[r] {
					rowHeights[r] = cellBox.Height
				}
				if cellBox.Depth > rowDepths[r] {
					rowDepths[r] = cellBox.Depth
				}
			}
		}

		totalW := 0.0
		for _, w := range colWidths {
			totalW += w + cellPaddingX*2
		}

		var children []*Box
		curY := 0.0

		for r := 0; r < numRows; r++ {
			curX := cellPaddingX
			for c := 0; c < len(grid[r]); c++ {
				cellBox := grid[r][c]
				if cellBox != nil {
					cellBox.X = curX + (colWidths[c]-cellBox.Width)/2
					cellBox.Y = curY
					children = append(children, cellBox)
				}
				curX += colWidths[c] + cellPaddingX*2
			}
			if r < numRows-1 {
				curY += rowHeights[r] + rowDepths[r] + cellPaddingY*2
			}
		}

		totalH := rowHeights[0] + cellPaddingY
		totalD := curY + rowDepths[numRows-1] + cellPaddingY

		return &Box{
			Type:     "matrix",
			Width:    totalW,
			Height:   totalH,
			Depth:    totalD,
			Scale:    scale,
			Children: children,
		}

	case *AccNode:
		targetBox := e.BuildLayout(n.Target, scale)
		accText := "^"
		switch n.Accent {
		case "\\vec":
			accText = "→"
		case "\\bar":
			accText = "¯"
		case "\\tilde":
			accText = "~"
		case "\\dot":
			accText = "."
		case "\\ddot":
			accText = ".."
		}

		accBox := &Box{
			Type:       "char",
			X:          (targetBox.Width - 0.3*scale) / 2,
			Y:          -(targetBox.Height + 0.1*scale),
			Width:      0.3 * scale,
			Height:     0.2 * scale,
			Depth:      0,
			Text:       accText,
			FontFamily: "symbol",
			Scale:      scale * 0.8,
		}

		return &Box{
			Type:     "group",
			Width:    targetBox.Width,
			Height:   targetBox.Height + 0.3*scale,
			Depth:    targetBox.Depth,
			Scale:    scale,
			Children: []*Box{targetBox, accBox},
		}

	case *SpaceNode:
		return &Box{
			Type:   "space",
			Width:  n.Width * scale,
			Height: 0,
			Depth:  0,
			Scale:  scale,
		}

	case *TextNode:
		w := float64(len(n.Text)) * 0.55 * scale
		return &Box{
			Type:       "text",
			Width:      w,
			Height:     0.75 * scale,
			Depth:      0.2 * scale,
			Text:       n.Text,
			FontFamily: n.Style,
			Scale:      scale,
		}
	}

	return &Box{Width: 0, Height: 0, Depth: 0, Scale: scale}
}
