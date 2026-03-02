package gui

// render_draw_dispatch.go — draw dispatch helpers ported from
// V's render_svg.v. Pure geometry/batching — no GPU calls.

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
