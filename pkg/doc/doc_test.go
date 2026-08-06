package doc

import (
	"testing"
)

func TestParseDocument(t *testing.T) {
	input := `Where $V(S_t)$ is the state and $$\sum_{i=1}^n x_i$$ is the sum.`

	segments := ParseDocument(input)
	if len(segments) < 4 {
		t.Fatalf("Expected at least 4 segments, got %d", len(segments))
	}

	if segments[0].Type != SegmentText || segments[0].Content != "Where " {
		t.Errorf("Unexpected segment 0: %+v", segments[0])
	}

	if segments[1].Type != SegmentInlineMath || segments[1].Content != "V(S_t)" {
		t.Errorf("Unexpected segment 1: %+v", segments[1])
	}

	if segments[2].Type != SegmentText || segments[2].Content != " is the state and " {
		t.Errorf("Unexpected segment 2: %+v", segments[2])
	}

	if segments[3].Type != SegmentBlockMath || segments[3].Content != `\sum_{i=1}^n x_i` {
		t.Errorf("Unexpected segment 3: %+v", segments[3])
	}
}
