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
