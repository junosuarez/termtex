package tex

// Node represents an AST node in the parsed LaTeX math expression.
type Node interface {
	node()
}

// CharNode represents a single character or symbol (letter, digit, operator).
type CharNode struct {
	Symbol     string
	FontFamily string // "math-italic", "math-upright", "math-bold", "symbol", "math-bb", "math-cal"
	IsOperator bool
}

func (c *CharNode) node() {}

// GroupNode represents a sequence of AST nodes grouped together.
type GroupNode struct {
	Children []Node
}

func (g *GroupNode) node() {}

// SubSupNode represents a base expression with subscript and/or superscript.
type SubSupNode struct {
	Base Node
	Sub  Node
	Sup  Node
}

func (s *SubSupNode) node() {}

// FracNode represents a fraction \frac{num}{den}.
type FracNode struct {
	Num Node
	Den Node
}

func (f *FracNode) node() {}

// BinomNode represents a binomial coefficient \binom{n}{k}.
type BinomNode struct {
	Top    Node
	Bottom Node
}

func (b *BinomNode) node() {}

// SqrtNode represents a square or n-th root \sqrt[index]{content}.
type SqrtNode struct {
	Index   Node
	Content Node
}

func (s *SqrtNode) node() {}

// BigOpNode represents a large operator (\sum, \prod, \int) with limits above/below.
type BigOpNode struct {
	Op    string
	Under Node
	Over  Node
}

func (b *BigOpNode) node() {}

// DelimNode represents delimiters surrounding an expression \left( ... \right).
type DelimNode struct {
	Left  string
	Right string
	Inner Node
}

func (d *DelimNode) node() {}

// MatrixNode represents a grid environment (\begin{matrix}, \begin{pmatrix}, \begin{cases}).
type MatrixNode struct {
	Kind string // "matrix", "pmatrix", "bmatrix", "cases"
	Rows [][]Node
}

func (m *MatrixNode) node() {}

// TextNode represents plain text inside math mode \text{...}.
type TextNode struct {
	Text  string
	Style string
}

func (t *TextNode) node() {}

// AccNode represents an accented expression (\vec{v}, \hat{x}, \bar{z}).
type AccNode struct {
	Accent string
	Target Node
}

func (a *AccNode) node() {}

// SpaceNode represents explicit spacing (\quad, \;, \,).
type SpaceNode struct {
	Width float64 // in em
}

func (s *SpaceNode) node() {}
