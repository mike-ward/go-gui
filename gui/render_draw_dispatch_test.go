package gui

import "testing"

// applyTransformToTriangles returns a new slice of transformed
// triangle vertices. Each pair (x,y) is multiplied by the 2D
// affine matrix m [a,b,c,d,tx,ty].
func applyTransformToTriangles(tris []float32, m [6]float32) []float32 {
	out := make([]float32, 0, len(tris))
	return applyTransformToTrianglesInto(tris, m, out)
}

// applyTransformToTrianglesInto is the non-allocating version.
func applyTransformToTrianglesInto(tris []float32, m [6]float32, out []float32) []float32 {
	out = out[:0]
	if len(tris) == 0 {
		return out
	}
	if cap(out) < len(tris) {
		out = make([]float32, 0, len(tris))
	}
	for i := 0; i < len(tris)-1; i += 2 {
		x := tris[i]
		y := tris[i+1]
		out = append(out, m[0]*x+m[2]*y+m[4])
		out = append(out, m[1]*x+m[3]*y+m[5])
	}
	return out
}

func TestApplyTransformIdentity(t *testing.T) {
	tris := []float32{1, 2, 3, 4}
	m := [6]float32{1, 0, 0, 1, 0, 0}
	out := applyTransformToTriangles(tris, m)
	if len(out) != 4 || out[0] != 1 || out[1] != 2 || out[2] != 3 || out[3] != 4 {
		t.Errorf("identity: got %v", out)
	}
}

func TestApplyTransformScale(t *testing.T) {
	tris := []float32{1, 2, 3, 4}
	m := [6]float32{2, 0, 0, 3, 0, 0}
	out := applyTransformToTriangles(tris, m)
	if out[0] != 2 || out[1] != 6 || out[2] != 6 || out[3] != 12 {
		t.Errorf("scale: got %v", out)
	}
}

func TestApplyTransformTranslate(t *testing.T) {
	tris := []float32{0, 0}
	m := [6]float32{1, 0, 0, 1, 10, 20}
	out := applyTransformToTriangles(tris, m)
	if out[0] != 10 || out[1] != 20 {
		t.Errorf("translate: got %v", out)
	}
}

func TestApplyTransformEmpty(t *testing.T) {
	m := [6]float32{1, 0, 0, 1, 0, 0}
	out := applyTransformToTriangles(nil, m)
	if len(out) != 0 {
		t.Error("empty should return empty")
	}
}

func TestApplyTransformIntoReuse(t *testing.T) {
	tris := []float32{1, 0, 0, 1}
	m := [6]float32{2, 0, 0, 2, 0, 0}
	buf := make([]float32, 0, 10)
	out := applyTransformToTrianglesInto(tris, m, buf)
	if len(out) != 4 || out[0] != 2 || out[3] != 2 {
		t.Errorf("into: got %v", out)
	}
}

func TestApplyTransformRotation90(t *testing.T) {
	// 90-degree CCW rotation: [0, -1, 1, 0, 0, 0]
	tris := []float32{1, 0}
	m := [6]float32{0, -1, 1, 0, 0, 0}
	out := applyTransformToTriangles(tris, m)
	if out[0] != 0 || out[1] != -1 {
		t.Errorf("rotation: got %v", out)
	}
}
