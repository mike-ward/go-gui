package gui

import "testing"

func TestFloatAutoFlipVertical(t *testing.T) {
	// Parent near bottom edge; tooltip below overflows → flip above.
	parent := Layout{
		Shape: &Shape{X: 100, Y: 550, Width: 100, Height: 30},
		Children: []Layout{{Shape: &Shape{
			Float: true, FloatAutoFlip: true,
			Width: 80, Height: 100,
			FloatAnchor: FloatBottomCenter,
			FloatTieOff: FloatTopCenter,
		}}},
	}
	parent.Children[0].Parent = &parent
	win := DrawClip{Width: 800, Height: 600}

	x, y := floatAttachLayout(&parent.Children[0], win)
	if !f32AreClose(x, 110) {
		t.Errorf("X: got %f, want 110", x)
	}
	if !f32AreClose(y, 450) {
		t.Errorf("Y: got %f, want 450", y)
	}
}

func TestFloatAutoFlipHorizontal(t *testing.T) {
	// Parent near right edge; tooltip right overflows → flip left.
	parent := Layout{
		Shape: &Shape{X: 750, Y: 200, Width: 40, Height: 30},
		Children: []Layout{{Shape: &Shape{
			Float: true, FloatAutoFlip: true,
			Width: 100, Height: 50,
			FloatAnchor: FloatMiddleRight,
			FloatTieOff: FloatMiddleLeft,
		}}},
	}
	parent.Children[0].Parent = &parent
	win := DrawClip{Width: 800, Height: 600}

	x, y := floatAttachLayout(&parent.Children[0], win)
	if !f32AreClose(x, 650) {
		t.Errorf("X: got %f, want 650", x)
	}
	if !f32AreClose(y, 190) {
		t.Errorf("Y: got %f, want 190", y)
	}
}

func TestFloatAutoFlipNoChange(t *testing.T) {
	// Parent centered; tooltip fits → no flip.
	parent := Layout{
		Shape: &Shape{X: 350, Y: 250, Width: 100, Height: 30},
		Children: []Layout{{Shape: &Shape{
			Float: true, FloatAutoFlip: true,
			Width: 80, Height: 40,
			FloatAnchor: FloatBottomCenter,
			FloatTieOff: FloatTopCenter,
		}}},
	}
	parent.Children[0].Parent = &parent
	win := DrawClip{Width: 800, Height: 600}

	x, y := floatAttachLayout(&parent.Children[0], win)
	if !f32AreClose(x, 360) {
		t.Errorf("X: got %f, want 360", x)
	}
	if !f32AreClose(y, 280) {
		t.Errorf("Y: got %f, want 280", y)
	}
}

func TestFloatAutoFlipDisabled(t *testing.T) {
	// Same overflow but FloatAutoFlip=false → no flip.
	parent := Layout{
		Shape: &Shape{X: 100, Y: 550, Width: 100, Height: 30},
		Children: []Layout{{Shape: &Shape{
			Float:       true,
			Width:       80,
			Height:      100,
			FloatAnchor: FloatBottomCenter,
			FloatTieOff: FloatTopCenter,
		}}},
	}
	parent.Children[0].Parent = &parent
	win := DrawClip{Width: 800, Height: 600}

	x, y := floatAttachLayout(&parent.Children[0], win)
	if !f32AreClose(x, 110) {
		t.Errorf("X: got %f, want 110", x)
	}
	if !f32AreClose(y, 580) {
		t.Errorf("Y: got %f, want 580", y)
	}
}

func TestFloatAutoFlipClamp(t *testing.T) {
	// Tooltip taller than window → flip reduces overflow, clamp
	// pins to top edge.
	parent := Layout{
		Shape: &Shape{X: 50, Y: 100, Width: 40, Height: 30},
		Children: []Layout{{Shape: &Shape{
			Float: true, FloatAutoFlip: true,
			Width: 80, Height: 200,
			FloatAnchor: FloatBottomCenter,
			FloatTieOff: FloatTopCenter,
		}}},
	}
	parent.Children[0].Parent = &parent
	win := DrawClip{Width: 200, Height: 150}

	x, y := floatAttachLayout(&parent.Children[0], win)
	if !f32AreClose(x, 30) {
		t.Errorf("X: got %f, want 30", x)
	}
	if !f32AreClose(y, 0) {
		t.Errorf("Y: got %f, want 0", y)
	}
}
