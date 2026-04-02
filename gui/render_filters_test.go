package gui

import (
	"math"
	"testing"
)

func TestFilterTextureDimsValid(t *testing.T) {
	d := filterTextureDimsFromBBox(100.5, 200.3, 4096)
	if !d.valid || d.width != 101 || d.height != 201 {
		t.Errorf("got %+v", d)
	}
}

func TestFilterTextureDimsInvalidZero(t *testing.T) {
	d := filterTextureDimsFromBBox(0, 100, 4096)
	if d.valid {
		t.Error("zero width should be invalid")
	}
}

func TestFilterTextureDimsInvalidNaN(t *testing.T) {
	d := filterTextureDimsFromBBox(float32(math.NaN()), 100, 4096)
	if d.valid {
		t.Error("NaN should be invalid")
	}
}

func TestFilterTextureDimsInvalidOversize(t *testing.T) {
	d := filterTextureDimsFromBBox(5000, 100, 4096)
	if d.valid {
		t.Error("oversize should be invalid")
	}
}

func TestFilterTextureDimsMaxTexZero(t *testing.T) {
	d := filterTextureDimsFromBBox(100, 100, 0)
	if d.valid {
		t.Error("maxTexSize 0 should be invalid")
	}
}

// indexFilterBracketEnds builds a map from FilterBegin index to
// FilterEnd index using a stack-based scan.
func indexFilterBracketEnds(renderers []RenderCmd) map[int]int {
	endByBegin := make(map[int]int)
	var stack []int
	for i := range renderers {
		switch renderers[i].Kind {
		case RenderFilterBegin:
			stack = append(stack, i)
		case RenderFilterEnd:
			if len(stack) > 0 {
				beginIdx := stack[len(stack)-1]
				stack = stack[:len(stack)-1]
				endByBegin[beginIdx] = i
			}
		}
	}
	return endByBegin
}

// appendRendererRange appends renderers[start:end] to dst.
func appendRendererRange(dst, src []RenderCmd, start, end int) []RenderCmd {
	if len(src) == 0 || start < 0 || start >= len(src) || end <= start {
		return dst
	}
	if end > len(src) {
		end = len(src)
	}
	if end <= start {
		return dst
	}
	return append(dst, src[start:end]...)
}

func TestIndexFilterBracketEndsEmpty(t *testing.T) {
	m := indexFilterBracketEnds(nil)
	if len(m) != 0 {
		t.Error("empty should return empty map")
	}
}

func TestIndexFilterBracketEndsSingle(t *testing.T) {
	cmds := []RenderCmd{
		{Kind: RenderFilterBegin},
		{Kind: RenderRect},
		{Kind: RenderFilterEnd},
	}
	m := indexFilterBracketEnds(cmds)
	if end, ok := m[0]; !ok || end != 2 {
		t.Errorf("got %v", m)
	}
}

func TestIndexFilterBracketEndsNested(t *testing.T) {
	cmds := []RenderCmd{
		{Kind: RenderFilterBegin}, // 0
		{Kind: RenderFilterBegin}, // 1
		{Kind: RenderFilterEnd},   // 2 → closes 1
		{Kind: RenderFilterEnd},   // 3 → closes 0
	}
	m := indexFilterBracketEnds(cmds)
	if m[1] != 2 {
		t.Errorf("inner: got %d, want 2", m[1])
	}
	if m[0] != 3 {
		t.Errorf("outer: got %d, want 3", m[0])
	}
}

func TestAppendRendererRangeBasic(t *testing.T) {
	src := []RenderCmd{{Kind: RenderRect}, {Kind: RenderText}, {Kind: RenderCircle}}
	dst := appendRendererRange(nil, src, 1, 3)
	if len(dst) != 2 || dst[0].Kind != RenderText || dst[1].Kind != RenderCircle {
		t.Errorf("got %v", dst)
	}
}

func TestAppendRendererRangeEmpty(t *testing.T) {
	dst := appendRendererRange(nil, nil, 0, 1)
	if len(dst) != 0 {
		t.Error("empty src should return nil")
	}
}

func TestAppendRendererRangeClampEnd(t *testing.T) {
	src := []RenderCmd{{Kind: RenderRect}, {Kind: RenderText}}
	dst := appendRendererRange(nil, src, 0, 100)
	if len(dst) != 2 {
		t.Errorf("should clamp end: got %d", len(dst))
	}
}
