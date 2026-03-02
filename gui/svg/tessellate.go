package svg

import (
	"math"
	"sort"

	"github.com/mike-ward/go-gui/gui"
)

// getTriangles tessellates all paths in the graphic into GPU-ready
// triangle geometry. Returns triangles in viewBox coordinate space.
func (vg *VectorGraphic) getTriangles(scale float32) []gui.TessellatedPath {
	return vg.tessellatePaths(vg.Paths, scale)
}

// tessellatePaths tessellates an arbitrary set of VectorPaths,
// using the VectorGraphic's clip paths and gradients.
func (vg *VectorGraphic) tessellatePaths(paths []VectorPath, scale float32) []gui.TessellatedPath {
	result := make([]gui.TessellatedPath, 0, len(paths)*2)

	baseTol := 0.5 / scale
	tolerance := baseTol
	if tolerance < 0.15 {
		tolerance = 0.15
	}

	clipGroupCounter := 0

	for i := range paths {
		path := &paths[i]

		// Clip group
		clipGroup := 0
		if path.ClipPathID != "" {
			if clipGeom, ok := vg.ClipPaths[path.ClipPathID]; ok {
				clipGroupCounter++
				clipGroup = clipGroupCounter
				for j := range clipGeom {
					cpPoly := flattenPath(&clipGeom[j], tolerance)
					clipTris := tessellatePolylines(cpPoly)
					if len(clipTris) > 0 {
						result = append(result, gui.TessellatedPath{
							Triangles:  clipTris,
							Color:      gui.SvgColor{R: 255, G: 255, B: 255, A: 255},
							IsClipMask: true,
							ClipGroup:  clipGroup,
							GroupID:    path.GroupID,
						})
					}
				}
			}
		}

		polylines := flattenPath(path, tolerance)

		// Fill tessellation
		hasGradient := path.FillGradientID != ""
		if path.FillColor.A > 0 || hasGradient {
			rawTris := tessellatePolylines(polylines)
			if len(rawTris) > 0 {
				if hasGradient {
					if g, ok := vg.Gradients[path.FillGradientID]; ok {
						grad := g
						if g.GradientUnits == "objectBoundingBox" || g.GradientUnits == "" {
							bx0, by0, bx1, by1 := bboxFromTriangles(rawTris)
							grad = resolveGradient(g, bx0, by0, bx1, by1)
						}
						fillTris := subdivideGradientTris(rawTris, grad)
						nVerts := len(fillTris) / 2
						vcols := make([]gui.SvgColor, nVerts)
						opacity := path.Opacity * path.FillOpacity
						for vi := 0; vi < nVerts; vi++ {
							vx := fillTris[vi*2]
							vy := fillTris[vi*2+1]
							t := projectOntoGradient(vx, vy, grad)
							c := interpolateGradient(grad.Stops, t)
							if opacity < 1.0 {
								c = applyOpacity(c, opacity)
							}
							vcols[vi] = c
						}
						result = append(result, gui.TessellatedPath{
							Triangles:    fillTris,
							Color:        path.FillColor,
							VertexColors: vcols,
							ClipGroup:    clipGroup,
							GroupID:      path.GroupID,
						})
					}
				} else {
					result = append(result, gui.TessellatedPath{
						Triangles: rawTris,
						Color:     path.FillColor,
						ClipGroup: clipGroup,
						GroupID:   path.GroupID,
					})
				}
			}
		}

		// Stroke tessellation
		hasStrokeGrad := path.StrokeGradientID != ""
		if (path.StrokeColor.A > 0 || hasStrokeGrad) && path.StrokeWidth > 0 {
			strokeWidth := path.StrokeWidth * scale
			strokePoly := polylines
			if len(path.StrokeDasharray) > 0 {
				strokePoly = applyDasharray(polylines, path.StrokeDasharray)
			}
			rawStroke := tessellateStroke(strokePoly, strokeWidth, path.StrokeCap, path.StrokeJoin)
			if len(rawStroke) > 0 {
				if hasStrokeGrad {
					if g, ok := vg.Gradients[path.StrokeGradientID]; ok {
						grad := g
						if g.GradientUnits == "objectBoundingBox" || g.GradientUnits == "" {
							bx0, by0, bx1, by1 := bboxFromTriangles(rawStroke)
							grad = resolveGradient(g, bx0, by0, bx1, by1)
						}
						sTris := subdivideGradientTris(rawStroke, grad)
						nVerts := len(sTris) / 2
						vcols := make([]gui.SvgColor, nVerts)
						opacity := path.Opacity * path.StrokeOpacity
						for vi := 0; vi < nVerts; vi++ {
							vx := sTris[vi*2]
							vy := sTris[vi*2+1]
							t := projectOntoGradient(vx, vy, grad)
							c := interpolateGradient(grad.Stops, t)
							if opacity < 1.0 {
								c = applyOpacity(c, opacity)
							}
							vcols[vi] = c
						}
						result = append(result, gui.TessellatedPath{
							Triangles:    sTris,
							Color:        path.StrokeColor,
							VertexColors: vcols,
							ClipGroup:    clipGroup,
							GroupID:      path.GroupID,
						})
					}
				} else {
					result = append(result, gui.TessellatedPath{
						Triangles: rawStroke,
						Color:     path.StrokeColor,
						ClipGroup: clipGroup,
						GroupID:   path.GroupID,
					})
				}
			}
		}
	}
	return result
}

// applyDasharray splits polylines into dash segments.
func applyDasharray(polylines [][]float32, dasharray []float32) [][]float32 {
	if len(dasharray) == 0 {
		return polylines
	}
	var result [][]float32
	for _, poly := range polylines {
		if len(poly) < 4 {
			continue
		}
		dashIdx := 0
		drawing := true
		remaining := dasharray[0]
		current := make([]float32, 0, len(poly))
		px, py := poly[0], poly[1]
		if drawing {
			current = append(current, px, py)
		}
		for i := 2; i < len(poly); i += 2 {
			nx, ny := poly[i], poly[i+1]
			dx, dy := nx-px, ny-py
			segLen := float32(math.Sqrt(float64(dx*dx + dy*dy)))
			if segLen < 1e-6 {
				continue
			}
			consumed := float32(0)
			for consumed < segLen-1e-6 {
				avail := segLen - consumed
				if remaining <= avail {
					t := (consumed + remaining) / segLen
					ix := px + t*dx
					iy := py + t*dy
					if drawing {
						current = append(current, ix, iy)
						if len(current) >= 4 {
							result = append(result, current)
						}
						current = make([]float32, 0, len(poly))
					} else {
						current = append(current, ix, iy)
					}
					consumed += remaining
					drawing = !drawing
					dashIdx = (dashIdx + 1) % len(dasharray)
					remaining = dasharray[dashIdx]
				} else {
					remaining -= avail
					if drawing {
						current = append(current, nx, ny)
					}
					break
				}
			}
			px, py = nx, ny
		}
		if drawing && len(current) >= 4 {
			result = append(result, current)
		}
	}
	return result
}

// --- Curve flattening ---

func flattenPath(path *VectorPath, tolerance float32) [][]float32 {
	var polylines [][]float32
	estimatedCap := len(path.Segments) * 16
	current := make([]float32, 0, estimatedCap)
	var x, y, startX, startY float32
	hasTx := !isIdentityTransform(path.Transform)

	for _, seg := range path.Segments {
		switch seg.Cmd {
		case CmdMoveTo:
			if len(current) >= 4 {
				polylines = append(polylines, current)
			}
			current = make([]float32, 0, estimatedCap)
			x = seg.Points[0]
			y = seg.Points[1]
			startX = x
			startY = y
			if hasTx {
				tx, ty := applyTransformPt(x, y, path.Transform)
				current = append(current, tx, ty)
			} else {
				current = append(current, x, y)
			}

		case CmdLineTo:
			x = seg.Points[0]
			y = seg.Points[1]
			if hasTx {
				tx, ty := applyTransformPt(x, y, path.Transform)
				if len(current) >= 2 && tx == current[len(current)-2] && ty == current[len(current)-1] {
					continue
				}
				current = append(current, tx, ty)
			} else {
				if len(current) >= 2 && x == current[len(current)-2] && y == current[len(current)-1] {
					continue
				}
				current = append(current, x, y)
			}

		case CmdQuadTo:
			cx := seg.Points[0]
			cy := seg.Points[1]
			ex := seg.Points[2]
			ey := seg.Points[3]
			if hasTx {
				tx, ty := applyTransformPt(x, y, path.Transform)
				tcx, tcy := applyTransformPt(cx, cy, path.Transform)
				tex, tey := applyTransformPt(ex, ey, path.Transform)
				flattenQuad(tx, ty, tcx, tcy, tex, tey, tolerance, &current)
			} else {
				flattenQuad(x, y, cx, cy, ex, ey, tolerance, &current)
			}
			x = ex
			y = ey

		case CmdCubicTo:
			c1x := seg.Points[0]
			c1y := seg.Points[1]
			c2x := seg.Points[2]
			c2y := seg.Points[3]
			ex := seg.Points[4]
			ey := seg.Points[5]
			if hasTx {
				tx, ty := applyTransformPt(x, y, path.Transform)
				tc1x, tc1y := applyTransformPt(c1x, c1y, path.Transform)
				tc2x, tc2y := applyTransformPt(c2x, c2y, path.Transform)
				tex, tey := applyTransformPt(ex, ey, path.Transform)
				flattenCubic(tx, ty, tc1x, tc1y, tc2x, tc2y, tex, tey, tolerance, &current)
			} else {
				flattenCubic(x, y, c1x, c1y, c2x, c2y, ex, ey, tolerance, &current)
			}
			x = ex
			y = ey

		case CmdClose:
			if len(current) >= 2 {
				if x != startX || y != startY {
					if hasTx {
						tx, ty := applyTransformPt(startX, startY, path.Transform)
						current = append(current, tx, ty)
					} else {
						current = append(current, startX, startY)
					}
				}
			}
			if len(current) >= 6 {
				polylines = append(polylines, current)
			}
			current = make([]float32, 0, estimatedCap)
			x = startX
			y = startY
		}
	}

	if len(current) >= 6 {
		polylines = append(polylines, current)
	}
	return polylines
}

func flattenQuad(x0, y0, cx, cy, x1, y1, tolerance float32, points *[]float32) {
	flattenQuadRec(x0, y0, cx, cy, x1, y1, tolerance, 0, points)
}

func flattenQuadRec(x0, y0, cx, cy, x1, y1, tolerance float32, depth int, points *[]float32) {
	mx := (x0 + x1) / 2
	my := (y0 + y1) / 2
	dx := cx - mx
	dy := cy - my
	d := float32(math.Sqrt(float64(dx*dx + dy*dy)))

	if d <= tolerance || depth >= maxFlattenDepth {
		*points = append(*points, x1, y1)
	} else {
		ax := (x0 + cx) / 2
		ay := (y0 + cy) / 2
		bx := (cx + x1) / 2
		by := (cy + y1) / 2
		abx := (ax + bx) / 2
		aby := (ay + by) / 2
		flattenQuadRec(x0, y0, ax, ay, abx, aby, tolerance, depth+1, points)
		flattenQuadRec(abx, aby, bx, by, x1, y1, tolerance, depth+1, points)
	}
}

func flattenCubic(x0, y0, c1x, c1y, c2x, c2y, x1, y1, tolerance float32, points *[]float32) {
	flattenCubicRec(x0, y0, c1x, c1y, c2x, c2y, x1, y1, tolerance, 0, points)
}

func flattenCubicRec(x0, y0, c1x, c1y, c2x, c2y, x1, y1, tolerance float32, depth int, points *[]float32) {
	dx := x1 - x0
	dy := y1 - y0
	d := float32(math.Sqrt(float64(dx*dx + dy*dy)))

	if d < curveDegenThreshold {
		*points = append(*points, x1, y1)
		return
	}

	d1 := f32Abs((c1x-x0)*dy-(c1y-y0)*dx) / d
	d2 := f32Abs((c2x-x0)*dy-(c2y-y0)*dx) / d

	if d1+d2 <= tolerance || depth >= maxFlattenDepth {
		*points = append(*points, x1, y1)
	} else {
		ax := (x0 + c1x) / 2
		ay := (y0 + c1y) / 2
		bx := (c1x + c2x) / 2
		by := (c1y + c2y) / 2
		cx := (c2x + x1) / 2
		cy := (c2y + y1) / 2
		abx := (ax + bx) / 2
		aby := (ay + by) / 2
		bcx := (bx + cx) / 2
		bcy := (by + cy) / 2
		mx := (abx + bcx) / 2
		my := (aby + bcy) / 2
		flattenCubicRec(x0, y0, ax, ay, abx, aby, mx, my, tolerance, depth+1, points)
		flattenCubicRec(mx, my, bcx, bcy, cx, cy, x1, y1, tolerance, depth+1, points)
	}
}

// --- Ear clipping triangulation ---

func tessellatePolylines(polylines [][]float32) []float32 {
	if len(polylines) == 0 {
		return nil
	}
	if len(polylines) == 1 {
		if len(polylines[0]) >= 6 {
			return earClip(polylines[0])
		}
		return nil
	}

	// Multiple contours — handle holes
	type contour struct {
		points []float32
		area   float32
	}
	contours := make([]contour, 0, len(polylines))
	for _, poly := range polylines {
		if len(poly) >= 6 {
			contours = append(contours, contour{points: poly, area: polygonArea(poly)})
		}
	}
	if len(contours) == 0 {
		return nil
	}

	sort.Slice(contours, func(i, j int) bool {
		return f32Abs(contours[i].area) > f32Abs(contours[j].area)
	})

	outer := make([]float32, len(contours[0].points))
	copy(outer, contours[0].points)
	if contours[0].area < 0 {
		outer = reversePolygon(outer)
	}

	for i := 1; i < len(contours); i++ {
		hole := make([]float32, len(contours[i].points))
		copy(hole, contours[i].points)
		if polygonArea(hole) > 0 {
			hole = reversePolygon(hole)
		}
		outer = mergeHole(outer, hole)
	}

	return earClip(outer)
}

func reversePolygon(poly []float32) []float32 {
	n := len(poly) / 2
	result := make([]float32, len(poly))
	for i := n - 1; i >= 0; i-- {
		result[(n-1-i)*2] = poly[i*2]
		result[(n-1-i)*2+1] = poly[i*2+1]
	}
	return result
}

func mergeHole(outer, hole []float32) []float32 {
	nHole := len(hole) / 2
	holeIdx := 0
	maxX := hole[0]
	for i := 1; i < nHole; i++ {
		if hole[i*2] > maxX {
			maxX = hole[i*2]
			holeIdx = i
		}
	}
	holeX := hole[holeIdx*2]
	holeY := hole[holeIdx*2+1]

	nOuter := len(outer) / 2
	bestIdx := 0
	bestDist := float32(1e30)

	for i := 0; i < nOuter; i++ {
		x1 := outer[i*2]
		y1 := outer[i*2+1]
		j := (i + 1) % nOuter
		x2 := outer[j*2]
		y2 := outer[j*2+1]

		if (y1 <= holeY && y2 > holeY) || (y2 <= holeY && y1 > holeY) {
			t := (holeY - y1) / (y2 - y1)
			ix := x1 + t*(x2-x1)
			if ix >= holeX {
				dist := ix - holeX
				if dist < bestDist {
					bestDist = dist
					if f32Abs(y1-holeY) < f32Abs(y2-holeY) {
						bestIdx = i
					} else {
						bestIdx = j
					}
				}
			}
		}
	}

	for i := 0; i < nOuter; i++ {
		x := outer[i*2]
		y := outer[i*2+1]
		if x >= holeX {
			dx := x - holeX
			dy := y - holeY
			dist := dx*dx + dy*dy
			if dist < bestDist*bestDist {
				if isVertexVisible(outer, i, holeX, holeY) {
					bestDist = float32(math.Sqrt(float64(dist)))
					bestIdx = i
				}
			}
		}
	}

	estCap := len(outer) + len(hole) + 4
	if estCap/2 > 1000000 {
		result := make([]float32, len(outer))
		copy(result, outer)
		return result
	}
	result := make([]float32, 0, estCap)

	for i := 0; i <= bestIdx; i++ {
		result = append(result, outer[i*2], outer[i*2+1])
	}
	for i := 0; i < nHole; i++ {
		idx := (holeIdx + i) % nHole
		result = append(result, hole[idx*2], hole[idx*2+1])
	}
	result = append(result, hole[holeIdx*2], hole[holeIdx*2+1])
	for i := bestIdx; i < nOuter; i++ {
		result = append(result, outer[i*2], outer[i*2+1])
	}
	return result
}

func isVertexVisible(outer []float32, idx int, px, py float32) bool {
	vx := outer[idx*2]
	vy := outer[idx*2+1]
	n := len(outer) / 2
	for i := 0; i < n; i++ {
		j := (i + 1) % n
		if i == idx || j == idx {
			continue
		}
		if segmentsIntersect(px, py, vx, vy, outer[i*2], outer[i*2+1], outer[j*2], outer[j*2+1]) {
			return false
		}
	}
	return true
}

func segmentsIntersect(ax, ay, bx, by, cx, cy, dx, dy float32) bool {
	d1 := crossProductSign(cx, cy, dx, dy, ax, ay)
	d2 := crossProductSign(cx, cy, dx, dy, bx, by)
	d3 := crossProductSign(ax, ay, bx, by, cx, cy)
	d4 := crossProductSign(ax, ay, bx, by, dx, dy)
	return ((d1 > 0 && d2 < 0) || (d1 < 0 && d2 > 0)) &&
		((d3 > 0 && d4 < 0) || (d3 < 0 && d4 > 0))
}

func crossProductSign(ax, ay, bx, by, cx, cy float32) float32 {
	return (bx-ax)*(cy-ay) - (by-ay)*(cx-ax)
}

func earClip(polygon []float32) []float32 {
	n := len(polygon) / 2
	if n < 3 {
		return nil
	}
	// Strip trailing duplicate
	if n > 3 {
		lx := polygon[(n-1)*2]
		ly := polygon[(n-1)*2+1]
		if f32Abs(lx-polygon[0]) < closedPathEpsilon && f32Abs(ly-polygon[1]) < closedPathEpsilon {
			n--
		}
	}
	if n < 3 {
		return nil
	}
	poly := polygon[:n*2]
	if n == 3 {
		result := make([]float32, 6)
		copy(result, poly)
		return result
	}

	indices := make([]int, n)
	if polygonArea(poly) > 0 {
		for i := n - 1; i >= 0; i-- {
			indices[n-1-i] = i
		}
	} else {
		for i := 0; i < n; i++ {
			indices[i] = i
		}
	}

	triangles := make([]float32, 0, (n-2)*6)
	count := 2 * n
	v := n - 1

	for len(indices) > 2 {
		if count <= 0 {
			break
		}
		count--

		u := v
		if u >= len(indices) {
			u = 0
		}
		v = u + 1
		if v >= len(indices) {
			v = 0
		}
		w := v + 1
		if w >= len(indices) {
			w = 0
		}

		if isEar(poly, indices, u, v, w) {
			a := indices[u]
			b := indices[v]
			c := indices[w]
			triangles = append(triangles,
				poly[a*2], poly[a*2+1],
				poly[b*2], poly[b*2+1],
				poly[c*2], poly[c*2+1],
			)
			indices = append(indices[:v], indices[v+1:]...)
			count = 2 * len(indices)
		}
	}
	return triangles
}

func polygonArea(polygon []float32) float32 {
	n := len(polygon) / 2
	area := float32(0)
	j := n - 1
	for i := 0; i < n; i++ {
		area += (polygon[j*2] + polygon[i*2]) * (polygon[j*2+1] - polygon[i*2+1])
		j = i
	}
	return area / 2
}

func isEar(polygon []float32, indices []int, u, v, w int) bool {
	ax := polygon[indices[u]*2]
	ay := polygon[indices[u]*2+1]
	bx := polygon[indices[v]*2]
	by := polygon[indices[v]*2+1]
	cx := polygon[indices[w]*2]
	cy := polygon[indices[w]*2+1]

	cross := (bx-ax)*(cy-ay) - (by-ay)*(cx-ax)
	if cross <= 0 {
		return false
	}

	for i := 0; i < len(indices); i++ {
		if i == u || i == v || i == w {
			continue
		}
		px := polygon[indices[i]*2]
		py := polygon[indices[i]*2+1]
		if (px == ax && py == ay) || (px == bx && py == by) || (px == cx && py == cy) {
			continue
		}
		if pointInTriangle(px, py, ax, ay, bx, by, cx, cy) {
			return false
		}
	}
	return true
}

func pointInTriangle(px, py, ax, ay, bx, by, cx, cy float32) bool {
	v0x := cx - ax
	v0y := cy - ay
	v1x := bx - ax
	v1y := by - ay
	v2x := px - ax
	v2y := py - ay

	dot00 := v0x*v0x + v0y*v0y
	dot01 := v0x*v1x + v0y*v1y
	dot02 := v0x*v2x + v0y*v2y
	dot11 := v1x*v1x + v1y*v1y
	dot12 := v1x*v2x + v1y*v2y

	denom := dot00*dot11 - dot01*dot01
	if f32Abs(denom) < 1e-10 {
		return false
	}
	invDenom := 1.0 / denom
	uu := (dot11*dot02 - dot01*dot12) * invDenom
	vv := (dot00*dot12 - dot01*dot02) * invDenom

	return uu >= 0 && vv >= 0 && (uu+vv) < 1
}

// --- Gradient support ---

func resolveGradient(g gui.SvgGradientDef, minX, minY, maxX, maxY float32) gui.SvgGradientDef {
	w := maxX - minX
	h := maxY - minY
	return gui.SvgGradientDef{
		Stops:         g.Stops,
		X1:            minX + g.X1*w,
		Y1:            minY + g.Y1*h,
		X2:            minX + g.X2*w,
		Y2:            minY + g.Y2*h,
		GradientUnits: "userSpaceOnUse",
	}
}

func bboxFromTriangles(tris []float32) (float32, float32, float32, float32) {
	if len(tris) < 2 {
		return 0, 0, 0, 0
	}
	minX, minY := tris[0], tris[1]
	maxX, maxY := minX, minY
	for i := 2; i < len(tris); i += 2 {
		x, y := tris[i], tris[i+1]
		if x < minX {
			minX = x
		}
		if x > maxX {
			maxX = x
		}
		if y < minY {
			minY = y
		}
		if y > maxY {
			maxY = y
		}
	}
	return minX, minY, maxX, maxY
}

func projectOntoGradient(vx, vy float32, g gui.SvgGradientDef) float32 {
	dx := g.X2 - g.X1
	dy := g.Y2 - g.Y1
	lenSq := dx*dx + dy*dy
	if lenSq == 0 {
		return 0
	}
	t := ((vx-g.X1)*dx + (vy-g.Y1)*dy) / lenSq
	if t < 0 {
		return 0
	}
	if t > 1 {
		return 1
	}
	return t
}

func interpolateGradient(stops []gui.SvgGradientStop, t float32) gui.SvgColor {
	if len(stops) == 0 {
		return gui.SvgColor{A: 255}
	}
	if t <= stops[0].Offset || len(stops) == 1 {
		return stops[0].Color
	}
	last := stops[len(stops)-1]
	if t >= last.Offset {
		return last.Color
	}
	for i := 0; i < len(stops)-1; i++ {
		s0 := stops[i]
		s1 := stops[i+1]
		if t >= s0.Offset && t <= s1.Offset {
			r := s1.Offset - s0.Offset
			if r <= 0 {
				return s0.Color
			}
			f := (t - s0.Offset) / r
			return gui.SvgColor{
				R: uint8(float32(s0.Color.R) + (float32(s1.Color.R)-float32(s0.Color.R))*f),
				G: uint8(float32(s0.Color.G) + (float32(s1.Color.G)-float32(s0.Color.G))*f),
				B: uint8(float32(s0.Color.B) + (float32(s1.Color.B)-float32(s0.Color.B))*f),
				A: uint8(float32(s0.Color.A) + (float32(s1.Color.A)-float32(s0.Color.A))*f),
			}
		}
	}
	return last.Color
}

func subdivideGradientTris(tris []float32, grad gui.SvgGradientDef) []float32 {
	if len(grad.Stops) <= 2 {
		return tris
	}
	stopTs := make([]float32, 0, len(grad.Stops))
	for _, s := range grad.Stops {
		if s.Offset > 0.001 && s.Offset < 0.999 {
			stopTs = append(stopTs, s.Offset)
		}
	}
	if len(stopTs) == 0 {
		return tris
	}
	result := make([]float32, 0, len(tris)*2)
	for i := 0; i < len(tris)-5; i += 6 {
		splitTriAtStops(tris[i], tris[i+1], tris[i+2], tris[i+3],
			tris[i+4], tris[i+5], grad, stopTs, &result)
	}
	return result
}

func splitTriAtStops(ax, ay, bx, by, cx, cy float32, grad gui.SvgGradientDef, stopTs []float32, result *[]float32) {
	ta := projectOntoGradient(ax, ay, grad)
	tb := projectOntoGradient(bx, by, grad)
	tc := projectOntoGradient(cx, cy, grad)

	tMin := ta
	if tb < tMin {
		tMin = tb
	}
	if tc < tMin {
		tMin = tc
	}
	tMax := ta
	if tb > tMax {
		tMax = tb
	}
	if tc > tMax {
		tMax = tc
	}

	for _, tS := range stopTs {
		if tS > tMin+1e-4 && tS < tMax-1e-4 {
			// Sort vertices by t
			p0x, p0y, t0 := ax, ay, ta
			p1x, p1y, t1 := bx, by, tb
			p2x, p2y, t2 := cx, cy, tc
			if t0 > t1 {
				p0x, p0y, t0, p1x, p1y, t1 = p1x, p1y, t1, p0x, p0y, t0
			}
			if t1 > t2 {
				p1x, p1y, t1, p2x, p2y, t2 = p2x, p2y, t2, p1x, p1y, t1
			}
			if t0 > t1 {
				p0x, p0y, t0, p1x, p1y, t1 = p1x, p1y, t1, p0x, p0y, t0
			}

			f02 := float32(0.5)
			if t2-t0 > 1e-6 {
				f02 = (tS - t0) / (t2 - t0)
			}
			i1x := p0x + f02*(p2x-p0x)
			i1y := p0y + f02*(p2y-p0y)

			if tS < t1-1e-4 {
				f01 := float32(0.5)
				if t1-t0 > 1e-6 {
					f01 = (tS - t0) / (t1 - t0)
				}
				i2x := p0x + f01*(p1x-p0x)
				i2y := p0y + f01*(p1y-p0y)
				splitTriAtStops(p0x, p0y, i2x, i2y, i1x, i1y, grad, stopTs, result)
				splitTriAtStops(i2x, i2y, p1x, p1y, i1x, i1y, grad, stopTs, result)
				splitTriAtStops(p1x, p1y, p2x, p2y, i1x, i1y, grad, stopTs, result)
			} else if tS > t1+1e-4 {
				f12 := float32(0.5)
				if t2-t1 > 1e-6 {
					f12 = (tS - t1) / (t2 - t1)
				}
				i2x := p1x + f12*(p2x-p1x)
				i2y := p1y + f12*(p2y-p1y)
				splitTriAtStops(p0x, p0y, p1x, p1y, i1x, i1y, grad, stopTs, result)
				splitTriAtStops(p1x, p1y, i2x, i2y, i1x, i1y, grad, stopTs, result)
				splitTriAtStops(i1x, i1y, i2x, i2y, p2x, p2y, grad, stopTs, result)
			} else {
				splitTriAtStops(p0x, p0y, p1x, p1y, i1x, i1y, grad, stopTs, result)
				splitTriAtStops(p1x, p1y, p2x, p2y, i1x, i1y, grad, stopTs, result)
			}
			return
		}
	}
	*result = append(*result, ax, ay, bx, by, cx, cy)
}
