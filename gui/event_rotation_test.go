package gui

import "testing"

func TestRotateMouseInverseNilShape(t *testing.T) {
	e := &Event{MouseX: 10, MouseY: 20}
	origX, origY := rotateMouseInverse(nil, e)
	if origX != 10 || origY != 20 {
		t.Errorf("nil shape: originals (%v,%v), want (10,20)", origX, origY)
	}
	if e.MouseX != 10 || e.MouseY != 20 {
		t.Error("nil shape should not modify event")
	}
}

func TestRotateMouseInverseZeroTurns(t *testing.T) {
	s := &Shape{X: 0, Y: 0, Width: 100, Height: 100}
	e := &Event{MouseX: 30, MouseY: 40}
	origX, origY := rotateMouseInverse(s, e)
	if origX != 30 || origY != 40 {
		t.Errorf("0 turns: originals (%v,%v), want (30,40)", origX, origY)
	}
	if e.MouseX != 30 || e.MouseY != 40 {
		t.Error("0 turns should not modify event")
	}
}

func TestRotateMouseInverseAllQuarters(t *testing.T) {
	// Shape 100x100 at origin, center at (50,50).
	// Input: (70,60) => dx=20, dy=10.
	tests := []struct {
		turns        uint8
		wantX, wantY float32
	}{
		// Q1 inverse(90°CW): outX = cx+dy, outY = cy-dx
		{1, 60, 30},
		// Q2 inverse(180°): outX = cx-dx, outY = cy-dy
		{2, 30, 40},
		// Q3 inverse(270°CW): outX = cx-dy, outY = cy+dx
		{3, 40, 70},
	}
	for _, tt := range tests {
		s := &Shape{X: 0, Y: 0, Width: 100, Height: 100, QuarterTurns: tt.turns}
		e := &Event{MouseX: 70, MouseY: 60}
		rotateMouseInverse(s, e)
		if e.MouseX != tt.wantX || e.MouseY != tt.wantY {
			t.Errorf("turns=%d: got (%v,%v), want (%v,%v)",
				tt.turns, e.MouseX, e.MouseY, tt.wantX, tt.wantY)
		}
	}
}

func TestRotateMouseInverseReturnsOriginals(t *testing.T) {
	s := &Shape{X: 0, Y: 0, Width: 100, Height: 100, QuarterTurns: 2}
	e := &Event{MouseX: 70, MouseY: 60}
	origX, origY := rotateMouseInverse(s, e)
	if origX != 70 || origY != 60 {
		t.Errorf("originals: got (%v,%v), want (70,60)", origX, origY)
	}
	// Event should be modified.
	if e.MouseX == 70 && e.MouseY == 60 {
		t.Error("event should be modified for non-zero turns")
	}
}

func TestRotateMouseInverseRoundTrip(t *testing.T) {
	s := &Shape{X: 10, Y: 20, Width: 200, Height: 150, QuarterTurns: 1}
	e := &Event{MouseX: 80, MouseY: 90}
	origX, origY := rotateMouseInverse(s, e)
	// Restore originals.
	e.MouseX = origX
	e.MouseY = origY
	if e.MouseX != 80 || e.MouseY != 90 {
		t.Errorf("round trip: got (%v,%v), want (80,90)", e.MouseX, e.MouseY)
	}
}
