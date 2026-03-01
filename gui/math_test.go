package gui

import "testing"

func TestIntClampBasic(t *testing.T) {
	if intClamp(-10, 0, 5) != 0 {
		t.Error("below min")
	}
	if intClamp(10, 0, 5) != 5 {
		t.Error("above max")
	}
	if intClamp(3, 0, 5) != 3 {
		t.Error("within range")
	}
}

func TestIntClampBoundaries(t *testing.T) {
	if intClamp(0, 0, 5) != 0 {
		t.Error("on min")
	}
	if intClamp(5, 0, 5) != 5 {
		t.Error("on max")
	}
	if intClamp(-3, -5, -1) != -3 {
		t.Error("negative within")
	}
	if intClamp(-10, -5, -1) != -5 {
		t.Error("negative below")
	}
	if intClamp(0, -5, -1) != -1 {
		t.Error("negative above")
	}
}

func TestF32ClampBasic(t *testing.T) {
	if f32Clamp(-1.5, 0.0, 2.5) != 0.0 {
		t.Error("below min")
	}
	if f32Clamp(3.14, 0.0, 2.5) != 2.5 {
		t.Error("above max")
	}
	if f32Clamp(1.25, 0.0, 2.5) != 1.25 {
		t.Error("within range")
	}
}

func TestF32ClampBoundaries(t *testing.T) {
	if f32Clamp(0.0, 0.0, 2.0) != 0.0 {
		t.Error("on min")
	}
	if f32Clamp(2.0, 0.0, 2.0) != 2.0 {
		t.Error("on max")
	}
	if f32Clamp(-3.0, -5.0, -1.0) != -3.0 {
		t.Error("negative within")
	}
	if f32Clamp(-10.0, -5.0, -1.0) != -5.0 {
		t.Error("negative below")
	}
	if f32Clamp(0.0, -5.0, -1.0) != -1.0 {
		t.Error("negative above")
	}
}

func TestF32AreCloseWithinTolerance(t *testing.T) {
	if !f32AreClose(1.00, 1.005) {
		t.Error("should be close")
	}
	if !f32AreClose(-2.50, -2.507) {
		t.Error("should be close negative")
	}
}

func TestF32AreCloseAtBoundary(t *testing.T) {
	if !f32AreClose(10.00, 10.009) {
		t.Error("should be close at boundary")
	}
	if !f32AreClose(-3.33, -3.339) {
		t.Error("should be close at boundary negative")
	}
}

func TestF32AreCloseOutsideTolerance(t *testing.T) {
	if f32AreClose(0.0, 0.02) {
		t.Error("should not be close")
	}
	if f32AreClose(-1.0, -1.02) {
		t.Error("should not be close negative")
	}
}
