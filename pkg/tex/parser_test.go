package tex

import (
	"testing"
)

func TestParseFractionsAndRoots(t *testing.T) {
	expr := `\frac{1}{x^2+1}`
	node, err := Parse(expr)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	fracNode, ok := node.(*FracNode)
	if !ok {
		t.Fatalf("Expected FracNode, got %T", node)
	}

	if fracNode.Num == nil || fracNode.Den == nil {
		t.Fatalf("Expected numerator and denominator")
	}
}

func TestParseBigOperators(t *testing.T) {
	expr := `\sum_{i=1}^n x_i`
	node, err := Parse(expr)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	group, ok := node.(*GroupNode)
	if !ok || len(group.Children) == 0 {
		t.Fatalf("Expected GroupNode with children")
	}

	bigOp, ok := group.Children[0].(*BigOpNode)
	if !ok {
		t.Fatalf("Expected BigOpNode, got %T", group.Children[0])
	}

	if bigOp.Op != "\\sum" {
		t.Errorf("Expected \\sum op, got %s", bigOp.Op)
	}
	if bigOp.Under == nil || bigOp.Over == nil {
		t.Errorf("Expected limits under and over")
	}
}

func TestParseUnderOverBrace(t *testing.T) {
	expr := `\underbrace{a + b}_{n \text{ terms}}`
	node, err := Parse(expr)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	brace, ok := node.(*UnderOverBraceNode)
	if !ok {
		t.Fatalf("Expected UnderOverBraceNode, got %T", node)
	}

	if brace.Kind != "underbrace" {
		t.Errorf("Expected underbrace kind, got %s", brace.Kind)
	}
	if brace.Target == nil || brace.Annotation == nil {
		t.Errorf("Expected Target and Annotation nodes")
	}
}

func TestParseMatrix(t *testing.T) {
	expr := `\begin{pmatrix} a & b \\ c & d \end{pmatrix}`
	node, err := Parse(expr)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	matrix, ok := node.(*MatrixNode)
	if !ok {
		t.Fatalf("Expected MatrixNode, got %T", node)
	}

	if len(matrix.Rows) != 2 {
		t.Fatalf("Expected 2 rows, got %d", len(matrix.Rows))
	}
}

func TestRenderSVG(t *testing.T) {
	expr := `x = \frac{-b \pm \sqrt{b^2 - 4ac}}{2a}`
	node, err := Parse(expr)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	opts := DefaultRenderOptions()
	svg, err := RenderSVG(node, opts)
	if err != nil {
		t.Fatalf("RenderSVG error: %v", err)
	}

	if len(svg) == 0 {
		t.Fatalf("Expected non-empty SVG output")
	}
}
