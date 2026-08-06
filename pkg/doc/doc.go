package doc

import (
	"fmt"
	"io"

	"termtex/pkg/render"
	"termtex/pkg/tex"
)

// SegmentType represents the type of document segment.
type SegmentType int

const (
	SegmentText SegmentType = iota
	SegmentInlineMath
	SegmentBlockMath
)

// Segment represents a piece of text or math formula in a document.
type Segment struct {
	Type    SegmentType
	Content string
}

// ParseDocument tokenizes a Markdown/LaTeX document string into text and math segments.
func ParseDocument(input string) []Segment {
	var segments []Segment
	runes := []rune(input)
	n := len(runes)
	pos := 0

	textStart := 0

	for pos < n {
		// Check for Block Math: $$ ... $$ or \[ ... \]
		if pos+1 < n && runes[pos] == '$' && runes[pos+1] == '$' {
			if pos > textStart {
				segments = append(segments, Segment{Type: SegmentText, Content: string(runes[textStart:pos])})
			}
			pos += 2
			mathStart := pos
			for pos+1 < n && !(runes[pos] == '$' && runes[pos+1] == '$') {
				pos++
			}
			mathContent := string(runes[mathStart:pos])
			if pos+1 < n {
				pos += 2
			}
			segments = append(segments, Segment{Type: SegmentBlockMath, Content: mathContent})
			textStart = pos
			continue
		}

		if pos+1 < n && runes[pos] == '\\' && runes[pos+1] == '[' {
			if pos > textStart {
				segments = append(segments, Segment{Type: SegmentText, Content: string(runes[textStart:pos])})
			}
			pos += 2
			mathStart := pos
			for pos+1 < n && !(runes[pos] == '\\' && runes[pos+1] == ']') {
				pos++
			}
			mathContent := string(runes[mathStart:pos])
			if pos+1 < n {
				pos += 2
			}
			segments = append(segments, Segment{Type: SegmentBlockMath, Content: mathContent})
			textStart = pos
			continue
		}

		// Check for Inline Math: $ ... $ or \( ... \)
		if runes[pos] == '$' {
			if pos > textStart {
				segments = append(segments, Segment{Type: SegmentText, Content: string(runes[textStart:pos])})
			}
			pos++
			mathStart := pos
			for pos < n && runes[pos] != '$' && runes[pos] != '\n' {
				pos++
			}
			if pos < n && runes[pos] == '$' {
				mathContent := string(runes[mathStart:pos])
				pos++
				segments = append(segments, Segment{Type: SegmentInlineMath, Content: mathContent})
				textStart = pos
				continue
			} else {
				// Unmatched $, treat as regular text
				pos = mathStart
				continue
			}
		}

		if pos+1 < n && runes[pos] == '\\' && runes[pos+1] == '(' {
			if pos > textStart {
				segments = append(segments, Segment{Type: SegmentText, Content: string(runes[textStart:pos])})
			}
			pos += 2
			mathStart := pos
			for pos+1 < n && !(runes[pos] == '\\' && runes[pos+1] == ')') {
				pos++
			}
			mathContent := string(runes[mathStart:pos])
			if pos+1 < n {
				pos += 2
			}
			segments = append(segments, Segment{Type: SegmentInlineMath, Content: mathContent})
			textStart = pos
			continue
		}

		pos++
	}

	if textStart < n {
		segments = append(segments, Segment{Type: SegmentText, Content: string(runes[textStart:n])})
	}

	return segments
}

// RenderDocument renders a document containing mixed text, inline math, and block math.
func RenderDocument(w io.Writer, docInput string, baseOpts tex.RenderOptions) error {
	segments := ParseDocument(docInput)

	inlineOpts := baseOpts
	inlineOpts.FontSize = 20.0
	inlineOpts.Padding = 2.0
	inlineOpts.DisplayMode = false

	blockOpts := baseOpts
	blockOpts.FontSize = 34.0
	blockOpts.Padding = 12.0
	blockOpts.DisplayMode = true

	for _, seg := range segments {
		switch seg.Type {
		case SegmentText:
			fmt.Fprint(w, seg.Content)

		case SegmentInlineMath:
			astNode, err := tex.Parse(seg.Content)
			if err != nil {
				// Fallback to raw text if parsing fails
				fmt.Fprint(w, seg.Content)
				continue
			}
			svg, err := tex.RenderSVG(astNode, inlineOpts)
			if err != nil {
				fmt.Fprint(w, seg.Content)
				continue
			}
			err = render.RenderInlineToTerminal(svg, w)
			if err != nil {
				// Fallback to ASCII text
				fmt.Fprint(w, render.RenderASCII(astNode))
			}

		case SegmentBlockMath:
			astNode, err := tex.Parse(seg.Content)
			if err != nil {
				fmt.Fprintln(w, seg.Content)
				continue
			}
			svg, err := tex.RenderSVG(astNode, blockOpts)
			if err != nil {
				fmt.Fprintln(w, seg.Content)
				continue
			}
			fmt.Fprintln(w) // spacing before block equation
			err = render.RenderToTerminal(svg, w)
			if err != nil {
				fmt.Fprintln(w, render.RenderASCII(astNode))
			}
			fmt.Fprintln(w) // spacing after block equation
		}
	}

	return nil
}
