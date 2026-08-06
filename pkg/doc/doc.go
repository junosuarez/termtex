package doc

import (
	"fmt"
	"io"

	"termtex/pkg/render"
	"termtex/pkg/tex"
)

type SegmentType int

const (
	SegmentText SegmentType = iota
	SegmentInlineMath
	SegmentBlockMath
)

type Segment struct {
	Type    SegmentType
	Content string
}

func ParseDocument(input string) []Segment {
	var segments []Segment
	runes := []rune(input)
	n := len(runes)
	pos := 0
	textStart := 0

	for pos < n {
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
			svg, err := tex.RenderTeXToSVG(seg.Content, inlineOpts)
			if err != nil {
				fmt.Fprint(w, seg.Content)
				continue
			}
			err = render.RenderInlineToTerminal(svg, w)
			if err != nil {
				fmt.Fprint(w, seg.Content)
			}

		case SegmentBlockMath:
			svg, err := tex.RenderTeXToSVG(seg.Content, blockOpts)
			if err != nil {
				fmt.Fprintln(w, seg.Content)
				continue
			}
			fmt.Fprintln(w)
			err = render.RenderToTerminal(svg, w)
			if err != nil {
				fmt.Fprintln(w, seg.Content)
			}
			fmt.Fprintln(w)
		}
	}

	return nil
}
