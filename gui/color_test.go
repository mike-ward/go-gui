package gui

import "testing"

func TestRGBAIsSet(t *testing.T) {
	c := RGBA(0, 0, 0, 0)
	if !c.IsSet() {
		t.Fatal("RGBA should produce a set color")
	}
}

func TestRGBIsSet(t *testing.T) {
	c := RGB(10, 20, 30)
	if !c.IsSet() {
		t.Fatal("RGB should produce a set color")
	}
	if c.A != 255 {
		t.Fatalf("RGB alpha: got %d, want 255", c.A)
	}
}

func TestHexIsSet(t *testing.T) {
	c := Hex(0xFF0000)
	if !c.IsSet() {
		t.Fatal("Hex should produce a set color")
	}
}

func TestZeroColorNotSet(t *testing.T) {
	var c Color
	if c.IsSet() {
		t.Fatal("zero Color should not be set")
	}
}

func TestPredefinedColorsAreSet(t *testing.T) {
	for _, c := range []Color{
		Black, White, Red, Green, Blue, ColorTransparent,
	} {
		if !c.IsSet() {
			t.Fatalf("predefined color %v should be set", c)
		}
	}
}

func TestColorTransparentIsSet(t *testing.T) {
	if !ColorTransparent.IsSet() {
		t.Fatal("ColorTransparent should be set")
	}
	if ColorTransparent.R != 0 || ColorTransparent.A != 0 {
		t.Fatal("ColorTransparent should be fully transparent")
	}
}

func TestWithOpacityPreservesSet(t *testing.T) {
	c := RGBA(255, 0, 0, 255).WithOpacity(0.5)
	if !c.IsSet() {
		t.Fatal("WithOpacity should preserve set")
	}
}

func TestAddProducesSet(t *testing.T) {
	c := Red.Add(Blue)
	if !c.IsSet() {
		t.Fatal("Add should produce set color")
	}
}

func TestSubProducesSet(t *testing.T) {
	c := White.Sub(Red)
	if !c.IsSet() {
		t.Fatal("Sub should produce set color")
	}
}

func TestOverProducesSet(t *testing.T) {
	c := Red.WithOpacity(0.5).Over(Blue)
	if !c.IsSet() {
		t.Fatal("Over should produce set color")
	}
}

func TestEqIgnoresSet(t *testing.T) {
	a := Color{R: 255, A: 255, set: true}
	b := Color{R: 255, A: 255}
	if !a.Eq(b) {
		t.Fatal("Eq should compare only RGBA channels")
	}
}

func TestString(t *testing.T) {
	c := RGB(10, 20, 30)
	want := "Color{10, 20, 30, 255}"
	if got := c.String(); got != want {
		t.Errorf("String() = %q, want %q", got, want)
	}
}

func TestRGBA8(t *testing.T) {
	c := RGBA(0xAA, 0xBB, 0xCC, 0xDD)
	got := c.RGBA8()
	want := int(0xAABBCCDD)
	if got != want {
		t.Errorf("RGBA8() = 0x%X, want 0x%X", got, want)
	}
}

func TestBGRA8(t *testing.T) {
	c := RGBA(0xAA, 0xBB, 0xCC, 0xDD)
	got := c.BGRA8()
	want := int(0xCCBBAADD)
	if got != want {
		t.Errorf("BGRA8() = 0x%X, want 0x%X", got, want)
	}
}

func TestABGR8(t *testing.T) {
	c := RGBA(0xAA, 0xBB, 0xCC, 0xDD)
	got := c.ABGR8()
	want := int(0xDDCCBBAA)
	if got != want {
		t.Errorf("ABGR8() = 0x%X, want 0x%X", got, want)
	}
}

func TestToCSSString(t *testing.T) {
	c := RGBA(10, 20, 30, 128)
	want := "rgba(10,20,30,128)"
	if got := c.ToCSSString(); got != want {
		t.Errorf("ToCSSString() = %q, want %q", got, want)
	}
}

func TestColorFromStringNamed(t *testing.T) {
	c := ColorFromString("red")
	if !c.Eq(Red) {
		t.Errorf("ColorFromString(red) = %v, want %v", c, Red)
	}
}

func TestColorFromStringHex(t *testing.T) {
	c := ColorFromString("#FF0000")
	if c.R != 255 || c.G != 0 || c.B != 0 {
		t.Errorf("ColorFromString(#FF0000) = %v, want red", c)
	}
}

func TestColorFromStringInvalidHex(t *testing.T) {
	c := ColorFromString("#ZZZZZZ")
	if c.R != 0 || c.G != 0 || c.B != 0 || c.A != 255 {
		t.Errorf("invalid hex should return black: %v", c)
	}
}

func TestColorFromStringUnknown(t *testing.T) {
	c := ColorFromString("chartreuse")
	if c.A != 255 || !c.IsSet() {
		t.Errorf("unknown name should return black: %v", c)
	}
}

func TestSubAlphaPreservesHigher(t *testing.T) {
	a := RGBA(200, 200, 200, 100)
	b := RGBA(50, 50, 50, 50)
	r := a.Sub(b)
	if r.A != 100 {
		t.Errorf("Sub alpha: got %d, want 100 (max of a,b)", r.A)
	}
}

func TestSubClampsToZero(t *testing.T) {
	r := RGB(10, 10, 10).Sub(RGB(20, 20, 20))
	if r.R != 0 || r.G != 0 || r.B != 0 {
		t.Errorf("Sub should clamp to 0: got %v", r)
	}
}

func TestSubUsesHigherAlpha(t *testing.T) {
	a := RGBA(200, 200, 200, 50)
	b := RGBA(10, 10, 10, 200)
	r := a.Sub(b)
	if r.A != 200 {
		t.Errorf("Sub alpha: got %d, want 200 (b.A > c.A)", r.A)
	}
}

func TestOverBothTransparent(t *testing.T) {
	a := RGBA(0, 0, 0, 0)
	b := RGBA(0, 0, 0, 0)
	r := a.Over(b)
	if r.IsSet() {
		t.Error("Over of two transparent colors should return zero")
	}
}

func TestAddClampsTo255(t *testing.T) {
	r := RGB(200, 200, 200).Add(RGB(200, 200, 200))
	if r.R != 255 || r.G != 255 || r.B != 255 {
		t.Errorf("Add should clamp to 255: got %v", r)
	}
}

func TestWithOpacityClampsRange(t *testing.T) {
	c := RGB(255, 0, 0)
	over := c.WithOpacity(2.0)
	if over.A != 255 {
		t.Errorf("WithOpacity(2.0) should clamp: got alpha %d", over.A)
	}
	under := c.WithOpacity(-1.0)
	if under.A != 0 {
		t.Errorf("WithOpacity(-1.0) should clamp: got alpha %d", under.A)
	}
}

func TestHexChannels(t *testing.T) {
	c := Hex(0x1A2B3C)
	if c.R != 0x1A || c.G != 0x2B || c.B != 0x3C || c.A != 255 {
		t.Errorf("Hex channels wrong: %v", c)
	}
}

func TestOverSemiTransparent(t *testing.T) {
	fg := RGBA(255, 0, 0, 128)
	bg := RGBA(0, 0, 255, 255)
	r := fg.Over(bg)
	if r.A == 0 {
		t.Error("Over result should not be fully transparent")
	}
	if r.R == 0 {
		t.Error("Over result should have some red")
	}
	if r.B == 0 {
		t.Error("Over result should have some blue")
	}
}
