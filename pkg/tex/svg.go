package tex

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

// RenderOptions configures SVG rendering parameters.
type RenderOptions struct {
	FgColor     string  // Hex color or "currentColor"
	BgColor     string  // Hex color or "transparent"
	FontSize    float64 // Base font size in pixels (default 28)
	Padding     float64 // Canvas padding in pixels (default 16)
	DisplayMode bool    // Display mode vs inline
}

// DefaultRenderOptions returns default rendering settings.
func DefaultRenderOptions() RenderOptions {
	return RenderOptions{
		FgColor:  "#cdd6f4", // Catppuccin Text / Terminal white
		BgColor:  "transparent",
		FontSize: 28,
		Padding:  16,
	}
}

// RenderTeXToSVG converts a LaTeX string into an SVG string, preferring MathJax if available.
func RenderTeXToSVG(texInput string, opts RenderOptions) (string, error) {
	// Try MathJax engine first for pixel-perfect TeX rendering
	svgStr, err := RenderTeXWithMathJax(texInput, opts)
	if err == nil && svgStr != "" {
		return svgStr, nil
	}

	// Fallback to native Go TeX layout engine
	astNode, parseErr := Parse(texInput)
	if parseErr != nil {
		return "", parseErr
	}
	return RenderSVG(astNode, opts)
}

// RenderTeXWithMathJax executes MathJax engine to generate vector SVG with Computer Modern TeX paths.
func RenderTeXWithMathJax(texInput string, opts RenderOptions) (string, error) {
	displayStr := "true"
	if !opts.DisplayMode {
		displayStr = "false"
	}

	cmd := exec.Command("node", "render_mathjax.js", texInput, displayStr, opts.FgColor)
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("MathJax execution failed: %w (%s)", err, stderr.String())
	}

	svgStr := strings.TrimSpace(out.String())
	if !strings.HasPrefix(svgStr, "<svg") {
		return "", fmt.Errorf("invalid SVG output from MathJax")
	}

	return svgStr, nil
}

// RenderSVG converts an AST Node into a fallback vector SVG string.
func RenderSVG(root Node, opts RenderOptions) (string, error) {
	if opts.FontSize <= 0 {
		opts.FontSize = 28
	}
	if opts.Padding < 0 {
		opts.Padding = 16
	}
	if opts.FgColor == "" {
		opts.FgColor = "#cdd6f4"
	}

	engine := NewLayoutEngine(opts.DisplayMode)
	layout := engine.BuildLayout(root, 1.0)

	emPx := opts.FontSize

	boxW := layout.Width * emPx
	boxH := layout.Height * emPx
	boxD := layout.Depth * emPx

	totalWidth := boxW + opts.Padding*2
	totalHeight := boxH + boxD + opts.Padding*2

	baselineY := opts.Padding + boxH

	var sb strings.Builder

	sb.WriteString(fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" width="%.2f" height="%.2f" viewBox="0 0 %.2f %.2f">`,
		totalWidth, totalHeight, totalWidth, totalHeight))
	sb.WriteString("\n")

	sb.WriteString(fmt.Sprintf(`  <style>
    .math-text { fill: %s; font-size: %.2fpx; }
    .math-italic { font-family: "Cambria Math", "Latin Modern Math", "STIX Two Math", "Times New Roman", serif; font-style: italic; }
    .math-upright { font-family: "Cambria Math", "Latin Modern Math", "STIX Two Math", "Times New Roman", serif; font-style: normal; }
    .math-bold { font-family: "Cambria Math", "Latin Modern Math", "STIX Two Math", serif; font-weight: bold; }
    .symbol { font-family: "Cambria Math", "STIX Two Math", "DejaVu Sans", sans-serif; }
    .rule { fill: %s; }
  </style>`, opts.FgColor, opts.FontSize, opts.FgColor))
	sb.WriteString("\n")

	if opts.BgColor != "" && opts.BgColor != "transparent" {
		sb.WriteString(fmt.Sprintf(`  <rect width="100%%" height="100%%" fill="%s" rx="8"/>`, opts.BgColor))
		sb.WriteString("\n")
	}

	renderBoxRecursive(&sb, layout, opts.Padding, baselineY, emPx, opts.FgColor)

	sb.WriteString("</svg>\n")

	return sb.String(), nil
}

func renderBoxRecursive(sb *strings.Builder, box *Box, currentX, baselineY, emPx float64, fgColor string) {
	if box == nil {
		return
	}

	absX := currentX + box.X*emPx
	absY := baselineY + box.Y*emPx

	switch box.Type {
	case "char", "text":
		class := "math-italic"
		if box.FontFamily == "math-upright" || box.FontFamily == "normal" {
			class = "math-upright"
		} else if box.FontFamily == "math-bold" || box.FontFamily == "bold" {
			class = "math-bold"
		} else if box.FontFamily == "symbol" {
			class = "symbol"
		}

		fontSize := emPx * box.Scale
		escapedText := escapeXML(box.Text)

		sb.WriteString(fmt.Sprintf(`  <text x="%.2f" y="%.2f" class="math-text %s" font-size="%.2fpx">%s</text>`,
			absX, absY, class, fontSize, escapedText))
		sb.WriteString("\n")

	case "rule":
		ruleW := box.Width * emPx
		ruleH := (box.Height + box.Depth) * emPx
		ruleY := absY - box.Height*emPx

		sb.WriteString(fmt.Sprintf(`  <rect x="%.2f" y="%.2f" width="%.2f" height="%.2f" class="rule" rx="1"/>`,
			absX, ruleY, ruleW, ruleH))
		sb.WriteString("\n")
	}

	for _, child := range box.Children {
		renderBoxRecursive(sb, child, absX, absY, emPx, fgColor)
	}
}

func escapeXML(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, `"`, "&quot;")
	return s
}
