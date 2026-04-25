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
