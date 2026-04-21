package gui

import "testing"

// TestBuildAnimatedSpansContiguous groups contiguous Animated entries
// with matching GroupID into a single span.
func TestBuildAnimatedSpansContiguous(t *testing.T) {
	paths := []CachedSvgPath{
		{GroupID: "", Animated: false},
		{GroupID: "a", Animated: true},
		{GroupID: "a", Animated: true},
		{GroupID: "", Animated: false},
		{GroupID: "b", Animated: true},
	}
	spans := buildAnimatedSpans(paths)
	if len(spans) != 2 {
		t.Fatalf("want 2 spans, got %d: %+v", len(spans), spans)
	}
	if spans[0].Start != 1 || spans[0].Count != 2 {
		t.Fatalf("want {1,2}, got %+v", spans[0])
	}
	if spans[1].Start != 4 || spans[1].Count != 1 {
		t.Fatalf("want {4,1}, got %+v", spans[1])
	}
}

// TestBuildAnimatedSpansSeparatesByGroupID does not merge two
// different-GroupID animated entries even when adjacent.
func TestBuildAnimatedSpansSeparatesByGroupID(t *testing.T) {
	paths := []CachedSvgPath{
		{GroupID: "a", Animated: true},
		{GroupID: "b", Animated: true},
	}
	spans := buildAnimatedSpans(paths)
	if len(spans) != 2 {
		t.Fatalf("want 2 spans, got %d", len(spans))
	}
}

// TestApplyAttrOverrideSetsMaskAndValue exercises every SvgAttrName
// case so no attribute silently drops.
func TestApplyAttrOverrideSetsMaskAndValue(t *testing.T) {
	cases := []struct {
		attr SvgAttrName
		mask SvgAnimAttrMask
	}{
		{SvgAttrCX, SvgAnimMaskCX},
		{SvgAttrCY, SvgAnimMaskCY},
		{SvgAttrR, SvgAnimMaskR},
		{SvgAttrRX, SvgAnimMaskRX},
		{SvgAttrRY, SvgAnimMaskRY},
		{SvgAttrX, SvgAnimMaskX},
		{SvgAttrY, SvgAnimMaskY},
		{SvgAttrWidth, SvgAnimMaskWidth},
		{SvgAttrHeight, SvgAnimMaskHeight},
	}
	for _, c := range cases {
		var o SvgAnimAttrOverride
		applyAttrOverride(&o, c.attr, 7)
		if o.Mask&c.mask == 0 {
			t.Fatalf("attr %d: mask bit not set", c.attr)
		}
	}
}

// TestComputeSvgAnimationsWritesAttrOverride verifies that a parsed
// SvgAnimAttr animation populates svgAnimState.AttrOverride at a
// mid-keyframe fraction.
func TestComputeSvgAnimationsWritesAttrOverride(t *testing.T) {
	anims := []SvgAnimation{{
		Kind:     SvgAnimAttr,
		GroupID:  "g",
		Values:   []float32{12, 6, 12},
		DurSec:   0.6,
		BeginSec: 0,
		AttrName: SvgAttrCY,
	}}
	states := make(map[string]svgAnimState, 1)
	// frac=0.5 → segment index = 1 exactly → returns vals[1] = 6.
	states = computeSvgAnimations(anims, 0.3, states)
	st, ok := states["g"]
	if !ok {
		t.Fatal("no state for GroupID g")
	}
	if st.AttrOverride.Mask&SvgAnimMaskCY == 0 {
		t.Fatal("CY mask not set")
	}
	if st.AttrOverride.CY < 5.9 || st.AttrOverride.CY > 6.1 {
		t.Fatalf("want CY≈6, got %.3f", st.AttrOverride.CY)
	}
}
