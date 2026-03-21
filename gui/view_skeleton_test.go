package gui

import "testing"

func TestSkeletonDefaultLayout(t *testing.T) {
	v := Skeleton(SkeletonCfg{ID: "s1"})
	layout := GenerateViewLayout(v, &Window{})
	if layout.Shape.Axis != AxisLeftToRight {
		t.Error("default should be horizontal (row)")
	}
}

func TestSkeletonCircleVariant(t *testing.T) {
	v := Skeleton(SkeletonCfg{
		ID:      "s2",
		Variant: SkeletonCircle,
	})
	layout := GenerateViewLayout(v, &Window{})
	if layout.Shape.ShapeType != ShapeCircle {
		t.Errorf("ShapeType = %d, want ShapeCircle (%d)",
			layout.Shape.ShapeType, ShapeCircle)
	}
}

func TestSkeletonA11YRole(t *testing.T) {
	v := Skeleton(SkeletonCfg{ID: "s3"})
	layout := GenerateViewLayout(v, &Window{})
	if layout.Shape.A11YRole != AccessRoleProgressBar {
		t.Errorf("role = %d, want ProgressBar",
			layout.Shape.A11YRole)
	}
}

func TestSkeletonA11YState(t *testing.T) {
	v := Skeleton(SkeletonCfg{ID: "s4"})
	layout := GenerateViewLayout(v, &Window{})
	want := AccessStateBusy | AccessStateLive
	if layout.Shape.A11YState != want {
		t.Errorf("state = %d, want %d",
			layout.Shape.A11YState, want)
	}
}

func TestSkeletonA11YLabel(t *testing.T) {
	v := Skeleton(SkeletonCfg{
		ID:        "s5",
		A11YLabel: "avatar",
	})
	layout := GenerateViewLayout(v, &Window{})
	if layout.Shape.A11Y == nil {
		t.Fatal("a11y nil")
	}
	if layout.Shape.A11Y.Label != "avatar" {
		t.Errorf("label = %q, want %q",
			layout.Shape.A11Y.Label, "avatar")
	}
}

func TestSkeletonA11YLabelDefault(t *testing.T) {
	v := Skeleton(SkeletonCfg{ID: "s6"})
	layout := GenerateViewLayout(v, &Window{})
	if layout.Shape.A11Y == nil {
		t.Fatal("a11y nil")
	}
	if layout.Shape.A11Y.Label != "Loading" {
		t.Errorf("label = %q, want %q",
			layout.Shape.A11Y.Label, "Loading")
	}
}

func TestSkeletonThemeColor(t *testing.T) {
	v := Skeleton(SkeletonCfg{ID: "s7"})
	layout := GenerateViewLayout(v, &Window{})
	want := guiTheme.SkeletonStyle.Color
	if layout.Shape.Color != want {
		t.Errorf("color = %v, want %v",
			layout.Shape.Color, want)
	}
}

func TestSkeletonCustomColor(t *testing.T) {
	c := Color{R: 200, G: 50, B: 50, A: 255, set: true}
	v := Skeleton(SkeletonCfg{ID: "s8", Color: c})
	layout := GenerateViewLayout(v, &Window{})
	if layout.Shape.Color != c {
		t.Errorf("color = %v, want %v",
			layout.Shape.Color, c)
	}
}

func TestSkeletonRadiusZeroOverride(t *testing.T) {
	v := Skeleton(SkeletonCfg{
		ID:     "s9",
		Radius: NoRadius,
	})
	layout := GenerateViewLayout(v, &Window{})
	if layout.Shape.Radius != 0 {
		t.Errorf("radius = %f, want 0", layout.Shape.Radius)
	}
}

func TestSkeletonSizeBorderNone(t *testing.T) {
	v := Skeleton(SkeletonCfg{ID: "s10"})
	layout := GenerateViewLayout(v, &Window{})
	if layout.Shape.SizeBorder != 0 {
		t.Errorf("SizeBorder = %f, want 0",
			layout.Shape.SizeBorder)
	}
}

func TestSkeletonInvisible(t *testing.T) {
	v := Skeleton(SkeletonCfg{
		ID:        "s11",
		Invisible: true,
	})
	layout := GenerateViewLayout(v, &Window{})
	// Invisible containers return a disabled, zero-size view.
	if !layout.Shape.Disabled {
		t.Error("invisible skeleton should be disabled")
	}
}

func TestSkeletonDisabled(t *testing.T) {
	v := Skeleton(SkeletonCfg{
		ID:       "s12",
		Disabled: true,
	})
	layout := GenerateViewLayout(v, &Window{})
	if !layout.Shape.Disabled {
		t.Error("Disabled not propagated")
	}
}

func TestSkeletonAmendLayoutSetsGradient(t *testing.T) {
	v := Skeleton(SkeletonCfg{ID: "s13"})
	w := &Window{}
	layout := GenerateViewLayout(v, w)
	if layout.Shape.Events == nil ||
		layout.Shape.Events.AmendLayout == nil {
		t.Fatal("AmendLayout not set")
	}
	layout.Shape.Events.AmendLayout(&layout, w)
	if layout.Shape.FX == nil {
		t.Fatal("FX nil after AmendLayout")
	}
	if layout.Shape.FX.Gradient == nil {
		t.Fatal("Gradient nil after AmendLayout")
	}
	if len(layout.Shape.FX.Gradient.Stops) != 5 {
		t.Errorf("gradient stops = %d, want 5",
			len(layout.Shape.FX.Gradient.Stops))
	}
	if layout.Shape.FX.Gradient.Direction != GradientToRight {
		t.Error("gradient direction should be ToRight")
	}
}
