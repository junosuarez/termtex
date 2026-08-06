package tex

import (
	"bytes"
	"fmt"
	"math"
	"os"
	"os/exec"
	"strings"
)

// RenderOptions configures SVG rendering parameters.
type RenderOptions struct {
	FgColor     string  // Hex color or "currentColor"
	BgColor     string  // Hex color or "transparent"
	FontSize    float64 // Base font size in pixels (default 32)
	Padding     float64 // Canvas padding in pixels (default 8)
	DisplayMode bool    // Display mode vs inline
}

// DefaultRenderOptions returns default rendering settings.
func DefaultRenderOptions() RenderOptions {
	return RenderOptions{
		FgColor:  "#cdd6f4", // Catppuccin Text / Terminal white
		BgColor:  "transparent",
		FontSize: 32,
		Padding:  8,
	}
}

// RenderTeXToSVG converts a LaTeX string into an SVG string, preferring MathJax if available.
func RenderTeXToSVG(texInput string, opts RenderOptions) (string, error) {
	svgStr, err := RenderTeXWithMathJax(texInput, opts)
	if err == nil && svgStr != "" {
		return svgStr, nil
	}

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

	scriptPath := "render_mathjax.js"
	if _, err := os.Stat(scriptPath); err != nil {
		if _, err2 := os.Stat("../../render_mathjax.js"); err2 == nil {
			scriptPath = "../../render_mathjax.js"
		}
	}

	cmd := exec.Command("node", scriptPath, texInput, displayStr, opts.FgColor)
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

// BBoxBounds tracks exact min/max bounds during SVG vector rendering.
type BBoxBounds struct {
	MinX float64
	MaxX float64
	MinY float64
	MaxY float64
	Init bool
}

func (b *BBoxBounds) AddPoint(x, y float64) {
	if !b.Init {
		b.MinX, b.MaxX = x, x
		b.MinY, b.MaxY = y, y
		b.Init = true
		return
	}
	if x < b.MinX {
		b.MinX = x
	}
	if x > b.MaxX {
		b.MaxX = x
	}
	if y < b.MinY {
		b.MinY = y
	}
	if y > b.MaxY {
		b.MaxY = y
	}
}

// RenderSVG converts an AST Node into a Computer Modern vector path SVG string.
func RenderSVG(root Node, opts RenderOptions) (string, error) {
	if opts.FontSize <= 0 {
		opts.FontSize = 32
	}
	if opts.Padding < 0 {
		opts.Padding = 8
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

	defsMap := make(map[string]GlyphInfo)
	bounds := &BBoxBounds{}

	var bodySb strings.Builder

	renderBoxVectorRecursive(&bodySb, layout, opts.Padding, baselineY, emPx, opts.FgColor, defsMap, bounds)

	// If bounds were collected, refine canvas size to tight bounding box
	if bounds.Init {
		w := math.Max(bounds.MaxX-bounds.MinX+opts.Padding*2, 10.0)
		h := math.Max(bounds.MaxY-bounds.MinY+opts.Padding*2, 10.0)
		if w > 0 && h > 0 {
			totalWidth = w
			totalHeight = h
		}
	}

	var sb strings.Builder

	sb.WriteString(fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink" width="%.2f" height="%.2f" viewBox="0 0 %.2f %.2f">`,
		totalWidth, totalHeight, totalWidth, totalHeight))
	sb.WriteString("\n")

	sb.WriteString("  <defs>\n")
	for _, glyph := range defsMap {
		sb.WriteString(fmt.Sprintf(`    <path id="%s" d="%s"/>`, glyph.ID, glyph.PathD))
		sb.WriteString("\n")
	}
	sb.WriteString("  </defs>\n")

	if opts.BgColor != "" && opts.BgColor != "transparent" {
		sb.WriteString(fmt.Sprintf(`  <rect width="100%%" height="100%%" fill="%s" rx="8"/>`, opts.BgColor))
		sb.WriteString("\n")
	}

	sb.WriteString(fmt.Sprintf(`  <g fill="%s" stroke="%s" stroke-width="0">`, opts.FgColor, opts.FgColor))
	sb.WriteString("\n")
	sb.WriteString(bodySb.String())
	sb.WriteString("  </g>\n")
	sb.WriteString("</svg>\n")

	return sb.String(), nil
}

func renderBoxVectorRecursive(sb *strings.Builder, box *Box, currentX, baselineY, emPx float64, fgColor string, defsMap map[string]GlyphInfo, bounds *BBoxBounds) {
	if box == nil {
		return
	}

	absX := currentX + box.X*emPx
	absY := baselineY + box.Y*emPx

	switch box.Type {
	case "char", "text":
		runes := []rune(box.Text)
		curX := absX

		for _, r := range runes {
			symStr := string(r)
			glyph, ok := GetGlyphInfo(symStr)

			if !ok {
				fontSize := emPx * box.Scale
				escapedText := escapeXML(symStr)
				sb.WriteString(fmt.Sprintf(`    <text x="%.2f" y="%.2f" fill="%s" font-size="%.2fpx" font-family="serif">%s</text>`,
					curX, absY, fgColor, fontSize, escapedText))
				sb.WriteString("\n")

				wPx := 0.55 * emPx * box.Scale
				bounds.AddPoint(curX, absY-fontSize)
				bounds.AddPoint(curX+wPx, absY)
				curX += wPx
				continue
			}

			defsMap[glyph.ID] = glyph
			scaleFactor := (emPx / 1000.0) * box.Scale

			sb.WriteString(fmt.Sprintf(`    <g transform="translate(%.2f, %.2f) scale(%.4f, -%.4f)">`,
				curX, absY, scaleFactor, scaleFactor))
			sb.WriteString(fmt.Sprintf(`      <use href="#%s"/>`, glyph.ID))
			sb.WriteString("    </g>\n")

			wPx := (glyph.Width / 1000.0) * emPx * box.Scale
			bounds.AddPoint(curX, absY-0.75*emPx*box.Scale)
			bounds.AddPoint(curX+wPx, absY+0.25*emPx*box.Scale)

			curX += wPx
		}

	case "rule":
		ruleW := box.Width * emPx
		ruleH := (box.Height + box.Depth) * emPx
		ruleY := absY - box.Height*emPx

		sb.WriteString(fmt.Sprintf(`    <rect x="%.2f" y="%.2f" width="%.2f" height="%.2f" rx="0.5"/>`,
			absX, ruleY, ruleW, ruleH))
		sb.WriteString("\n")

		bounds.AddPoint(absX, ruleY)
		bounds.AddPoint(absX+ruleW, ruleY+ruleH)
	}

	for _, child := range box.Children {
		renderBoxVectorRecursive(sb, child, absX, absY, emPx, fgColor, defsMap, bounds)
	}
}

func escapeXML(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, `"`, "&quot;")
	return s
}
