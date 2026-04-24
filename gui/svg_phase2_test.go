package gui

import "testing"

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
		applyAttrOverride(&o, c.attr, 7, false)
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
		Kind:          SvgAnimAttr,
		GroupID:       "g",
		TargetPathIDs: []uint32{1},
		Values:        []float32{12, 6, 12},
		DurSec:        0.6,
		BeginSec:      0,
		AttrName:      SvgAttrCY,
	}}
	states := make(map[uint32]svgAnimState, 1)
	// frac=0.5 → segment index = 1 exactly → returns vals[1] = 6.
	states = computeSvgAnimations(anims, 0.3, states)
	st, ok := states[1]
	if !ok {
		t.Fatal("no state for PathID 1")
	}
	if st.AttrOverride.Mask&SvgAnimMaskCY == 0 {
		t.Fatal("CY mask not set")
	}
	if st.AttrOverride.CY < 5.9 || st.AttrOverride.CY > 6.1 {
		t.Fatalf("want CY≈6, got %.3f", st.AttrOverride.CY)
	}
}
