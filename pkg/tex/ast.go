package tex

// Node represents a TeX math layout node in the AST.
type Node interface {
	isNode()
}

// GroupNode represents a sequence of nodes (e.g. {a + b}).
type GroupNode struct {
	Children []Node
}

func (GroupNode) isNode() {}

// CharNode represents a single math character or symbol.
type CharNode struct {
	Symbol     string // The character or TeX macro name (e.g. "a", "1", "\alpha", "\int")
	FontFamily string // "math-italic", "math-upright", "math-bold", "symbol"
	IsOperator bool
}

func (CharNode) isNode() {}

// SubSupNode represents a base node with optional subscript and superscript.
type SubSupNode struct {
	Base Node
	Sub  Node // Subscript (or nil)
	Sup  Node // Superscript (or nil)
}

func (SubSupNode) isNode() {}

// FracNode represents a fraction \frac{num}{den}.
type FracNode struct {
	Num Node
	Den Node
}

func (FracNode) isNode() {}

// SqrtNode represents a square root or N-th root \sqrt[opt]{content}.
type SqrtNode struct {
	Index   Node // Optional index for n-th root
	Content Node
}

func (SqrtNode) isNode() {}

// BigOpNode represents a large operator (\sum, \int, \prod, \lim) with limits.
type BigOpNode struct {
	Op    string // "\sum", "\int", "\prod", "\lim", etc.
	Under Node   // Lower limit / subscript
	Over  Node   // Upper limit / superscript
}

func (BigOpNode) isNode() {}

// DelimNode represents auto-sized delimiters \left( content \right).
type DelimNode struct {
	Left  string // "(", "[", "{", "|", "\langle", etc.
	Right string // ")", "]", "}", "|", "\rangle", etc.
	Inner Node
}

func (DelimNode) isNode() {}

// MatrixNode represents a matrix environment (\begin{matrix}...\end{matrix}, pmatrix, bmatrix, cases, etc.).
type MatrixNode struct {
	Kind string     // "matrix", "pmatrix", "bmatrix", "vmatrix", "cases"
	Rows [][]Node   // Grid of cell nodes
}

func (MatrixNode) isNode() {}

// AccNode represents an accent over a node (\hat{x}, \vec{v}, \bar{z}, \tilde{a}).
type AccNode struct {
	Accent string // "\hat", "\vec", "\bar", "\tilde", "\dot", "\ddot"
	Target Node
}

func (AccNode) isNode() {}

// SpaceNode represents explicit spacing (\,, \;, \quad, \qquad, \!).
type SpaceNode struct {
	Width float64 // Width in em units
}

func (SpaceNode) isNode() {}

// TextNode represents \text{...} or \mathrm{...} literal text.
type TextNode struct {
	Text  string
	Style string // "normal", "bold", "italic", "monospace"
}

func (TextNode) isNode() {}
