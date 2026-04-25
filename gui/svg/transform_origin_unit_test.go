package svg

import "testing"

func toBox(minX, minY, w, h float32) bbox {
	return bbox{
		MinX: minX, MinY: minY,
		MaxX: minX + w, MaxY: minY + h,
		Set: true,
	}
}

func TestResolveTransformOrigin_EmptyDefaultsCenter(t *testing.T) {
	b := toBox(10, 20, 40, 60) // center (30, 50)
	x, y := resolveTransformOrigin("", b)
	if x != 30 || y != 50 {
		t.Errorf("got (%v,%v) want (30,50)", x, y)
	}
}

func TestResolveTransformOrigin_UnsetBboxYieldsZero(t *testing.T) {
	x, y := resolveTransformOrigin("50% 50%", bbox{})
	if x != 0 || y != 0 {
		t.Errorf("got (%v,%v) want (0,0)", x, y)
	}
}

func TestResolveTransformOrigin_SwapWhenYKeywordFirst(t *testing.T) {
	// "top left": Y-only keyword first → axes swap.
	b := toBox(0, 0, 100, 200)
	x, y := resolveTransformOrigin("top left", b)
	if x != 0 || y != 0 {
		t.Errorf("got (%v,%v) want (0,0) — top-left", x, y)
	}
}

func TestResolveTransformOrigin_SwapWhenXKeywordSecond(t *testing.T) {
	b := toBox(0, 0, 100, 200)
	// "50% right" — X-only keyword in second slot → swap;
	// "right" goes to X = MaxX = 100, "50%" → Y = 100.
	x, y := resolveTransformOrigin("50% right", b)
	if x != 100 || y != 100 {
		t.Errorf("got (%v,%v) want (100,100)", x, y)
	}
}

func TestResolveTransformOrigin_ThreeTokenIgnoresZ(t *testing.T) {
	// Per CSS Transforms 1, the Z component is ignored.
	b := toBox(0, 0, 100, 100)
	x, y := resolveTransformOrigin("25% 75% 999", b)
	if x != 25 || y != 75 {
		t.Errorf("got (%v,%v) want (25,75); Z must be ignored", x, y)
	}
}

func TestResolveTransformOrigin_NegativePercent(t *testing.T) {
	b := toBox(0, 0, 200, 100)
	x, y := resolveTransformOrigin("-50% 0%", b)
	if x != -100 || y != 0 {
		t.Errorf("got (%v,%v) want (-100,0)", x, y)
	}
}

func TestResolveTransformOrigin_SingleKeywordOnlyX(t *testing.T) {
	b := toBox(0, 0, 100, 80)
	// "right" alone → X=MaxX, Y defaults to center.
	x, y := resolveTransformOrigin("right", b)
	if x != 100 || y != 40 {
		t.Errorf("got (%v,%v) want (100,40)", x, y)
	}
}

func TestResolveTransformOrigin_SingleKeywordOnlyY(t *testing.T) {
	b := toBox(0, 0, 100, 80)
	// "bottom" alone → swap to Y axis; X defaults to center.
	x, y := resolveTransformOrigin("bottom", b)
	if x != 50 || y != 80 {
		t.Errorf("got (%v,%v) want (50,80)", x, y)
	}
}

func TestResolveTransformOrigin_PxSuffixActsAsAuthorUnits(t *testing.T) {
	b := toBox(0, 0, 100, 100)
	x, y := resolveTransformOrigin("12px 34px", b)
	if x != 12 || y != 34 {
		t.Errorf("got (%v,%v) want (12,34)", x, y)
	}
}
