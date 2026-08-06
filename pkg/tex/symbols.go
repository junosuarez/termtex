package tex

// SymbolInfo holds information about a math symbol or macro.
type SymbolInfo struct {
	Unicode    string
	IsOperator bool
	IsBigOp    bool
	SVGPath    string  // SVG path definition if available
	ViewBox    string  // SVG viewBox for path glyph
	WidthEm    float64 // Width in em units
}

// SymbolMap maps TeX macros to their SymbolInfo.
var SymbolMap = map[string]SymbolInfo{
	// Lowercase Greek
	"\\alpha":      {Unicode: "α", WidthEm: 0.6},
	"\\beta":       {Unicode: "β", WidthEm: 0.55},
	"\\gamma":      {Unicode: "γ", WidthEm: 0.5},
	"\\delta":      {Unicode: "δ", WidthEm: 0.5},
	"\\epsilon":    {Unicode: "ϵ", WidthEm: 0.5},
	"\\varepsilon": {Unicode: "ε", WidthEm: 0.5},
	"\\zeta":       {Unicode: "ζ", WidthEm: 0.45},
	"\\eta":        {Unicode: "η", WidthEm: 0.55},
	"\\theta":      {Unicode: "θ", WidthEm: 0.5},
	"\\iota":       {Unicode: "ι", WidthEm: 0.35},
	"\\kappa":      {Unicode: "κ", WidthEm: 0.55},
	"\\lambda":     {Unicode: "λ", WidthEm: 0.55},
	"\\mu":         {Unicode: "μ", WidthEm: 0.55},
	"\\nu":         {Unicode: "ν", WidthEm: 0.5},
	"\\xi":         {Unicode: "ξ", WidthEm: 0.5},
	"\\pi":         {Unicode: "π", WidthEm: 0.55},
	"\\rho":        {Unicode: "ρ", WidthEm: 0.55},
	"\\sigma":      {Unicode: "σ", WidthEm: 0.55},
	"\\tau":        {Unicode: "τ", WidthEm: 0.45},
	"\\upsilon":    {Unicode: "υ", WidthEm: 0.5},
	"\\phi":        {Unicode: "ϕ", WidthEm: 0.6},
	"\\varphi":     {Unicode: "φ", WidthEm: 0.6},
	"\\chi":        {Unicode: "χ", WidthEm: 0.55},
	"\\psi":        {Unicode: "ψ", WidthEm: 0.65},
	"\\omega":      {Unicode: "ω", WidthEm: 0.6},

	// Uppercase Greek
	"\\Gamma":   {Unicode: "Γ", WidthEm: 0.6},
	"\\Delta":   {Unicode: "Δ", WidthEm: 0.7},
	"\\Theta":   {Unicode: "Θ", WidthEm: 0.75},
	"\\Lambda":  {Unicode: "Λ", WidthEm: 0.7},
	"\\Xi":      {Unicode: "Ξ", WidthEm: 0.65},
	"\\Pi":      {Unicode: "Π", WidthEm: 0.75},
	"\\Sigma":   {Unicode: "Σ", WidthEm: 0.75},
	"\\Upsilon": {Unicode: "Υ", WidthEm: 0.7},
	"\\Phi":     {Unicode: "Φ", WidthEm: 0.75},
	"\\Psi":     {Unicode: "Ψ", WidthEm: 0.8},
	"\\Omega":   {Unicode: "Ω", WidthEm: 0.75},

	// Binary Operators
	"\\pm":     {Unicode: "±", IsOperator: true, WidthEm: 0.65},
	"\\mp":     {Unicode: "∓", IsOperator: true, WidthEm: 0.65},
	"\\times":  {Unicode: "×", IsOperator: true, WidthEm: 0.65},
	"\\div":    {Unicode: "÷", IsOperator: true, WidthEm: 0.65},
	"\\cdot":   {Unicode: "·", IsOperator: true, WidthEm: 0.4},
	"\\star":   {Unicode: "⋆", IsOperator: true, WidthEm: 0.5},
	"\\ast":    {Unicode: "*", IsOperator: true, WidthEm: 0.5},
	"\\cap":    {Unicode: "∩", IsOperator: true, WidthEm: 0.65},
	"\\cup":    {Unicode: "∪", IsOperator: true, WidthEm: 0.65},

	// Relations & Set Operators
	"\\neq":       {Unicode: "≠", IsOperator: true, WidthEm: 0.65},
	"\\approx":    {Unicode: "≈", IsOperator: true, WidthEm: 0.65},
	"\\le":        {Unicode: "≤", IsOperator: true, WidthEm: 0.65},
	"\\leq":       {Unicode: "≤", IsOperator: true, WidthEm: 0.65},
	"\\ge":        {Unicode: "≥", IsOperator: true, WidthEm: 0.65},
	"\\geq":       {Unicode: "≥", IsOperator: true, WidthEm: 0.65},
	"\\in":        {Unicode: "∈", IsOperator: true, WidthEm: 0.65},
	"\\notin":     {Unicode: "∉", IsOperator: true, WidthEm: 0.65},
	"\\subset":    {Unicode: "⊂", IsOperator: true, WidthEm: 0.65},
	"\\subseteq":  {Unicode: "⊆", IsOperator: true, WidthEm: 0.65},
	"\\supset":    {Unicode: "⊃", IsOperator: true, WidthEm: 0.65},
	"\\supseteq":  {Unicode: "⊇", IsOperator: true, WidthEm: 0.65},
	"\\equiv":     {Unicode: "≡", IsOperator: true, WidthEm: 0.65},
	"\\sim":       {Unicode: "∼", IsOperator: true, WidthEm: 0.65},
	"\\propto":    {Unicode: "∝", IsOperator: true, WidthEm: 0.65},

	// Arrows & Logic
	"\\to":                {Unicode: "→", IsOperator: true, WidthEm: 0.85},
	"\\rightarrow":        {Unicode: "→", IsOperator: true, WidthEm: 0.85},
	"\\leftarrow":         {Unicode: "←", IsOperator: true, WidthEm: 0.85},
	"\\leftrightarrow":     {Unicode: "↔", IsOperator: true, WidthEm: 0.95},
	"\\Rightarrow":        {Unicode: "⇒", IsOperator: true, WidthEm: 0.85},
	"\\Leftarrow":         {Unicode: "⇐", IsOperator: true, WidthEm: 0.85},
	"\\Longleftrightarrow": {Unicode: "⇔", IsOperator: true, WidthEm: 1.2},
	"\\mapsto":            {Unicode: "↦", IsOperator: true, WidthEm: 0.85},
	"\\therefore":         {Unicode: "∴", IsOperator: true, WidthEm: 0.65},
	"\\because":           {Unicode: "∵", IsOperator: true, WidthEm: 0.65},

	// Big Operators
	"\\sum":    {Unicode: "∑", IsOperator: true, IsBigOp: true, WidthEm: 1.0},
	"\\prod":   {Unicode: "∏", IsOperator: true, IsBigOp: true, WidthEm: 1.0},
	"\\coprod": {Unicode: "∐", IsOperator: true, IsBigOp: true, WidthEm: 1.0},
	"\\int":    {Unicode: "∫", IsOperator: true, IsBigOp: true, WidthEm: 0.8},
	"\\iint":   {Unicode: "∬", IsOperator: true, IsBigOp: true, WidthEm: 1.2},
	"\\iiint":  {Unicode: "∭", IsOperator: true, IsBigOp: true, WidthEm: 1.6},
	"\\oint":   {Unicode: "∮", IsOperator: true, IsBigOp: true, WidthEm: 0.8},
	"\\lim":    {Unicode: "lim", IsOperator: true, IsBigOp: true, WidthEm: 1.2},
	"\\max":    {Unicode: "max", IsOperator: true, IsBigOp: true, WidthEm: 1.4},
	"\\min":    {Unicode: "min", IsOperator: true, IsBigOp: true, WidthEm: 1.2},

	// Delimiters
	"\\langle": {Unicode: "⟨", WidthEm: 0.45},
	"\\rangle": {Unicode: "⟩", WidthEm: 0.45},
	"\\{":      {Unicode: "{", WidthEm: 0.45},
	"\\}":      {Unicode: "}", WidthEm: 0.45},

	// Miscellaneous Symbols
	"\\infty":    {Unicode: "∞", WidthEm: 0.75},
	"\\partial":  {Unicode: "∂", WidthEm: 0.55},
	"\\nabla":    {Unicode: "∇", WidthEm: 0.7},
	"\\hbar":     {Unicode: "ħ", WidthEm: 0.55},
	"\\forall":   {Unicode: "∀", WidthEm: 0.65},
	"\\exists":   {Unicode: "∃", WidthEm: 0.65},
	"\\emptyset": {Unicode: "∅", WidthEm: 0.6},
	"\\ldots":    {Unicode: "…", WidthEm: 0.8},
	"\\cdots":    {Unicode: "⋯", WidthEm: 0.8},

	// Standard Math Functions
	"\\sin": {Unicode: "sin", IsOperator: true, WidthEm: 1.1},
	"\\cos": {Unicode: "cos", IsOperator: true, WidthEm: 1.1},
	"\\tan": {Unicode: "tan", IsOperator: true, WidthEm: 1.1},
	"\\csc": {Unicode: "csc", IsOperator: true, WidthEm: 1.1},
	"\\sec": {Unicode: "sec", IsOperator: true, WidthEm: 1.1},
	"\\cot": {Unicode: "cot", IsOperator: true, WidthEm: 1.1},
	"\\log": {Unicode: "log", IsOperator: true, WidthEm: 1.1},
	"\\ln":  {Unicode: "ln", IsOperator: true, WidthEm: 0.75},
	"\\exp": {Unicode: "exp", IsOperator: true, WidthEm: 1.2},
	"\\det": {Unicode: "det", IsOperator: true, WidthEm: 1.1},
	"\\dim": {Unicode: "dim", IsOperator: true, WidthEm: 1.2},
	"\\ker": {Unicode: "ker", IsOperator: true, WidthEm: 1.1},
}

// GetSymbolInfo returns information for a given symbol or macro.
func GetSymbolInfo(sym string) SymbolInfo {
	if info, ok := SymbolMap[sym]; ok {
		return info
	}
	return SymbolInfo{
		Unicode: sym,
		WidthEm: 0.55,
	}
}
