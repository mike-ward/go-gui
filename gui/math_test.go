package gui

import "testing"

func TestIntClamp(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		v, lo, hi int
		want      int
	}{
		{"below min", -10, 0, 5, 0},
		{"above max", 10, 0, 5, 5},
		{"within range", 3, 0, 5, 3},
		{"on min", 0, 0, 5, 0},
		{"on max", 5, 0, 5, 5},
		{"negative within", -3, -5, -1, -3},
		{"negative below", -10, -5, -1, -5},
		{"negative above", 0, -5, -1, -1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := intClamp(tt.v, tt.lo, tt.hi); got != tt.want {
				t.Errorf("intClamp(%d, %d, %d) = %d, want %d",
					tt.v, tt.lo, tt.hi, got, tt.want)
			}
		})
	}
}

func TestF32Clamp(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		v, lo, hi float32
		want      float32
	}{
		{"below min", -1.5, 0.0, 2.5, 0.0},
		{"above max", 3.14, 0.0, 2.5, 2.5},
		{"within range", 1.25, 0.0, 2.5, 1.25},
		{"on min", 0.0, 0.0, 2.0, 0.0},
		{"on max", 2.0, 0.0, 2.0, 2.0},
		{"negative within", -3.0, -5.0, -1.0, -3.0},
		{"negative below", -10.0, -5.0, -1.0, -5.0},
		{"negative above", 0.0, -5.0, -1.0, -1.0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := f32Clamp(tt.v, tt.lo, tt.hi); got != tt.want {
				t.Errorf("f32Clamp(%f, %f, %f) = %f, want %f",
					tt.v, tt.lo, tt.hi, got, tt.want)
			}
		})
	}
}

func TestF32AreClose(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		a, b float32
		want bool
	}{
		{"within tolerance", 1.00, 1.005, true},
		{"within negative", -2.50, -2.507, true},
		{"at boundary", 10.00, 10.009, true},
		{"at boundary negative", -3.33, -3.339, true},
		{"outside tolerance", 0.0, 0.02, false},
		{"outside negative", -1.0, -1.02, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := f32AreClose(tt.a, tt.b); got != tt.want {
				t.Errorf("f32AreClose(%f, %f) = %v, want %v",
					tt.a, tt.b, got, tt.want)
			}
		})
	}
}
