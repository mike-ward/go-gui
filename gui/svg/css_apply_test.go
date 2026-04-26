package svg

import "testing"

func TestApplyCSSProp_OpacityPercent(t *testing.T) {
	out := ComputedStyle{}
	applyCSSProp("opacity", "50%", &out)
	if out.Opacity != 0.5 {
		t.Errorf("opacity 50%%: got %v want 0.5", out.Opacity)
	}
}

func TestApplyCSSProp_FillOpacityPercent(t *testing.T) {
	out := ComputedStyle{}
	applyCSSProp("fill-opacity", "25%", &out)
	if out.FillOpacity != 0.25 {
		t.Errorf("fill-opacity 25%%: got %v", out.FillOpacity)
	}
}

func TestApplyCSSProp_StrokeOpacityPercent(t *testing.T) {
	out := ComputedStyle{}
	applyCSSProp("stroke-opacity", "100%", &out)
	if out.StrokeOpacity != 1 {
		t.Errorf("stroke-opacity 100%%: got %v", out.StrokeOpacity)
	}
}

func TestApplyCSSProp_OpacityUnitlessUnchanged(t *testing.T) {
	out := ComputedStyle{}
	applyCSSProp("opacity", "0.4", &out)
	if out.Opacity != 0.4 {
		t.Errorf("opacity 0.4: got %v", out.Opacity)
	}
}

func TestApplyCSSProp_OpacityPercentClamped(t *testing.T) {
	out := ComputedStyle{}
	applyCSSProp("opacity", "150%", &out)
	if out.Opacity != 1 {
		t.Errorf("150%% should clamp to 1: got %v", out.Opacity)
	}
	applyCSSProp("opacity", "-10%", &out)
	if out.Opacity != 0 {
		t.Errorf("-10%% should clamp to 0: got %v", out.Opacity)
	}
}

// Malformed % → parseFloatTrimmed returns NaN → clampOpacity01 maps
// to 0. Without the clamp, the uint8 alpha cast at bake time is
// implementation-defined for NaN.
func TestApplyCSSProp_OpacityMalformedPercent(t *testing.T) {
	out := ComputedStyle{Opacity: 0.7}
	applyCSSProp("opacity", "abc%", &out)
	if out.Opacity != 0 {
		t.Errorf("abc%%: got %v want 0 (clamped NaN)", out.Opacity)
	}
}

func TestSplitVarArgs_EmptyAndNoComma(t *testing.T) {
	cases := []struct {
		in               string
		wantName, wantFB string
		wantHas          bool
	}{
		{"", "", "", false},
		{"--brand", "--brand", "", false},
		{"--brand,red", "--brand", "red", true},
		{"  --brand  ,  red  ", "  --brand  ", "  red  ", true},
	}
	for _, tc := range cases {
		name, fb, has := splitVarArgs(tc.in)
		if name != tc.wantName || fb != tc.wantFB || has != tc.wantHas {
			t.Errorf("splitVarArgs(%q) = (%q,%q,%v), want (%q,%q,%v)",
				tc.in, name, fb, has, tc.wantName, tc.wantFB, tc.wantHas)
		}
	}
}

func TestSplitVarArgs_NestedParens(t *testing.T) {
	// Top-level comma split must not fire inside nested parens.
	in := "--a, var(--b, red)"
	name, fb, has := splitVarArgs(in)
	if !has || name != "--a" || fb != " var(--b, red)" {
		t.Errorf("nested parens split wrong: name=%q fb=%q has=%v",
			name, fb, has)
	}
	// calc() inside fallback.
	in = "--w, calc(10px + 4px)"
	name, fb, has = splitVarArgs(in)
	if !has || name != "--w" || fb != " calc(10px + 4px)" {
		t.Errorf("calc fallback split: name=%q fb=%q has=%v",
			name, fb, has)
	}
}
