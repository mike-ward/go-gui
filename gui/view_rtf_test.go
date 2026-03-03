package gui

import (
	"math"
	"testing"

	"github.com/mike-ward/go-glyph"
)

func TestRtfHitTestLogic(t *testing.T) {
	item := glyph.Item{
		X: 10, Y: 20, Width: 50,
		Ascent: 12, Descent: 4,
	}
	r := rtfRunRect(item)
	if r.X != 10 || r.Width != 50 {
		t.Fatalf("rect X/W: got %v/%v", r.X, r.Width)
	}
	// Y = run.Y - Ascent = 20 - 12 = 8.
	if r.Y != 8 {
		t.Fatalf("rect Y: expected 8, got %v", r.Y)
	}
	// Height = Ascent + Descent = 16.
	if r.Height != 16 {
		t.Fatalf("rect Height: expected 16, got %v", r.Height)
	}

	// Point inside.
	if !rtfHitTest(item, 30, 15, nil) {
		t.Error("expected hit at (30,15)")
	}
	// Point outside.
	if rtfHitTest(item, 5, 5, nil) {
		t.Error("expected miss at (5,5)")
	}
}

func TestRtfAffineInverse(t *testing.T) {
	// Identity matrix.
	id := glyph.AffineTransform{
		XX: 1, XY: 0, YX: 0, YY: 1, X0: 0, Y0: 0,
	}
	inv, ok := rtfAffineInverse(id)
	if !ok {
		t.Fatal("identity should be invertible")
	}
	if math.Abs(float64(inv.XX-1)) > 0.001 ||
		math.Abs(float64(inv.YY-1)) > 0.001 {
		t.Fatalf("identity inverse: XX=%v YY=%v",
			inv.XX, inv.YY)
	}

	// Singular matrix.
	singular := glyph.AffineTransform{
		XX: 0, XY: 0, YX: 0, YY: 0,
	}
	_, ok = rtfAffineInverse(singular)
	if ok {
		t.Fatal("singular matrix should not be invertible")
	}
}
