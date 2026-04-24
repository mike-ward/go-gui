package svg

import (
	"cmp"
	"math"
	"slices"

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

	// Per-GroupID svgAnimState holds one base; sibling paths with
	// divergent transforms would collapse onto one representative
	// at render. Force-bake those into vertex coords.
	forceBake := buildForceBakeSet(paths, vg.Animations)

	clipGroupCounter := 0

	for i := range paths {
		path := &paths[i]

		clipGroup := 0
		if path.ClipPathID != "" {
			result, clipGroup = vg.appendClipMasks(result, path,
				tolerance, &clipGroupCounter)
		}

		seed, bake := seedFromTransform(path, forceBake)
		seed.ClipGroup = clipGroup
		seed.GroupID = path.GroupID
		seed.Animated = path.Animated
		seed.Primitive = path.Primitive
		polylines := flattenPathWithBake(path, tolerance, bake)

		emittedGeometry := false

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
						for vi := range nVerts {
							vx := fillTris[vi*2]
							vy := fillTris[vi*2+1]
							t := projectOntoGradient(vx, vy, grad)
							c := interpolateGradient(grad.Stops, t)
							if opacity < 1.0 {
								c = applyOpacity(c, opacity)
							}
							vcols[vi] = c
						}
						out := seed
						out.Triangles = fillTris
						out.Color = path.FillColor
						out.VertexColors = vcols
						result = append(result, out)
						emittedGeometry = true
					}
				} else {
					out := seed
					out.Triangles = rawTris
					out.Color = path.FillColor
					result = append(result, out)
					emittedGeometry = true
				}
			}
		}

		// Stroke tessellation. Stroke width stays in viewBox units
		// here — render-side vertex scaling applies once. Pre-scaling
		// here plus render scaling produces width × scale² on screen,
		// which collapses to sub-pixel for large viewBoxes (e.g.
		// spinner.svg at scale=0.03 rendered 200→0.18px).
		hasStrokeGrad := path.StrokeGradientID != ""
		if (path.StrokeColor.A > 0 || hasStrokeGrad) && path.StrokeWidth > 0 {
			strokeWidth := path.StrokeWidth
			strokePoly := polylines
			if len(path.StrokeDasharray) > 0 {
				strokePoly = applyDasharray(polylines,
					path.StrokeDasharray, path.StrokeDashOffset)
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
						for vi := range nVerts {
							vx := sTris[vi*2]
							vy := sTris[vi*2+1]
							t := projectOntoGradient(vx, vy, grad)
							c := interpolateGradient(grad.Stops, t)
							if opacity < 1.0 {
								c = applyOpacity(c, opacity)
							}
							vcols[vi] = c
						}
						out := seed
						out.Triangles = sTris
						out.Color = path.StrokeColor
						out.VertexColors = vcols
						out.IsStroke = true
						result = append(result, out)
						emittedGeometry = true
					}
				} else {
					out := seed
					out.Triangles = rawStroke
					out.Color = path.StrokeColor
					out.IsStroke = true
					result = append(result, out)
					emittedGeometry = true
				}
			}
		}

		if !emittedGeometry && path.Animated &&
			path.Primitive.Kind != gui.SvgPrimNone {
			result = appendDegeneratePlaceholders(result, path, seed)
		}
	}
	// Bake viewBox origin into vertex coords so every downstream
	// coord is in content-from-origin space. Skip non-finite to
	// avoid splatting NaN/Inf across the whole mesh.
	if (vg.ViewBoxX != 0 || vg.ViewBoxY != 0) &&
		finiteViewBox(vg.ViewBoxX, vg.ViewBoxY) {
		shiftTriangles(result, -vg.ViewBoxX, -vg.ViewBoxY)
	}
	return result
}

// finiteViewBox guards coord shifts that would splat NaN/Inf
// across every vertex in the tessellated output.
func finiteViewBox(x, y float32) bool {
	return finiteF32(x) && finiteF32(y)
}

func shiftTriangles(paths []gui.TessellatedPath, dx, dy float32) {
	for i := range paths {
		tris := paths[i].Triangles
		for j := 0; j+1 < len(tris); j += 2 {
			tris[j] += dx
			tris[j+1] += dy
		}
	}
}

// seedFromTransform decides whether path.Transform is deferred to
// render-time Base* composition or baked into vertex coords. Deferred
// on TRS-decomposable matrices when the owning group is not in the
// force-bake set. Baked otherwise (shear, or conflicting siblings
// sharing one animation state). Returns the seed TessellatedPath
// carrying Base* (when deferred) and a bake flag for the flattener.
func seedFromTransform(
	path *VectorPath, forceBake map[string]bool,
) (gui.TessellatedPath, bool) {
	var seed gui.TessellatedPath
	if isIdentityTransform(path.Transform) {
		return seed, false
	}
	if forceBake[path.GroupID] {
		return seed, true
	}
	tx, ty, sx, sy, rot, ok := decomposeTRS(path.Transform)
	if !ok {
		return seed, true
	}
	seed.BaseTransX, seed.BaseTransY = tx, ty
	seed.BaseScaleX, seed.BaseScaleY = sx, sy
	seed.BaseRotAngle = rot
	seed.HasBaseXform = true
	return seed, false
}

// buildForceBakeSet returns GroupIDs where >=2 member paths carry
// non-identity Transforms AND the group owns a transform-kind SMIL
// animation. See pre-scan comment in tessellatePaths.
func buildForceBakeSet(
	paths []VectorPath, anims []gui.SvgAnimation,
) map[string]bool {
	if len(anims) == 0 {
		return nil
	}
	xformAnimGroups := make(map[string]struct{}, len(anims))
	for i := range anims {
		a := &anims[i]
		if a.GroupID == "" {
			continue
		}
		switch a.Kind {
		case gui.SvgAnimRotate, gui.SvgAnimTranslate,
			gui.SvgAnimScale, gui.SvgAnimMotion:
			xformAnimGroups[a.GroupID] = struct{}{}
		}
	}
	if len(xformAnimGroups) == 0 {
		return nil
	}
	counts := make(map[string]int)
	for i := range paths {
		p := &paths[i]
		if p.GroupID == "" || isIdentityTransform(p.Transform) {
			continue
		}
		if _, ok := xformAnimGroups[p.GroupID]; !ok {
			continue
		}
		counts[p.GroupID]++
	}
	var out map[string]bool
	for gid, n := range counts {
		if n > 1 {
			if out == nil {
				out = make(map[string]bool)
			}
			out[gid] = true
		}
	}
	return out
}

// appendClipMasks emits per-subpath clip-mask TessellatedPaths for
// the path's referenced clipPath. Bumps counter and returns the
// assigned clipGroup so subsequent fill/stroke entries inherit it.
func (vg *VectorGraphic) appendClipMasks(result []gui.TessellatedPath,
	path *VectorPath, tolerance float32, counter *int,
) ([]gui.TessellatedPath, int) {
	clipGeom, ok := vg.ClipPaths[path.ClipPathID]
	if !ok {
		return result, 0
	}
	*counter++
	clipGroup := *counter
	for j := range clipGeom {
		cpPoly := flattenPath(&clipGeom[j], tolerance)
		clipTris := tessellatePolylines(cpPoly)
		if len(clipTris) == 0 {
			continue
		}
		result = append(result, gui.TessellatedPath{
			Triangles:  clipTris,
			Color:      gui.SvgColor{R: 255, G: 255, B: 255, A: 255},
			IsClipMask: true,
			ClipGroup:  clipGroup,
			GroupID:    path.GroupID,
		})
	}
	return result, clipGroup
}

// appendDegeneratePlaceholders emits zero-triangle TessellatedPath
// entries for an Animated primitive that produced no static geometry
// (e.g. <circle r="0"> animating r). One placeholder per configured
// paint (fill / stroke) keeps span counts matching the live result
// from TessellateAnimated; without these the spinner is invisible.
// seed carries clip/group/primitive/base-transform state shared with
// the concrete fill/stroke emissions.
func appendDegeneratePlaceholders(result []gui.TessellatedPath,
	path *VectorPath, seed gui.TessellatedPath,
) []gui.TessellatedPath {
	wantsFill := path.FillColor.A > 0 || path.FillGradientID != ""
	wantsStroke := (path.StrokeColor.A > 0 ||
		path.StrokeGradientID != "") && path.StrokeWidth > 0
	if !wantsFill && !wantsStroke {
		wantsFill = true // ensure at least one placeholder
	}
	seed.Animated = true
	if wantsFill {
		out := seed
		out.Color = path.FillColor
		result = append(result, out)
	}
	if wantsStroke {
		out := seed
		out.Color = path.StrokeColor
		out.IsStroke = true
		result = append(result, out)
	}
	return result
}

// minDashCycleLen bounds the smallest accepted dasharray cycle.
// Sub-threshold cycles would force the inner consume loop to iterate
// segLen / cycleLen times — hostile or buggy authors with cycles
// near float32 epsilon could DoS the tessellator. ~thousandth of a
// pixel is finer than any real renderer needs.
const minDashCycleLen = float32(1e-3)

// maxDashIterPerPoly caps the inner dash-consume loop per polyline
// so a finite-but-pathological dasharray (extremely small relative
// to segment length) cannot stall tessellation.
const maxDashIterPerPoly = 1 << 20

// applyDasharray splits polylines into dash segments. offset is the
// SVG stroke-dashoffset in viewBox units: positive advances the dash
// phase forward (first dash starts later); negative wraps backward.
func applyDasharray(polylines [][]float32, dasharray []float32,
	offset float32) [][]float32 {
	if len(dasharray) == 0 {
		return polylines
	}
	// All-zero / non-finite dasharray: per SVG spec, treat as solid.
	// Also guards the inner loop, where remaining=0 never advances.
	cycleLen := float32(0)
	for _, v := range dasharray {
		cycleLen += v
	}
	if cycleLen < minDashCycleLen ||
		math.IsInf(float64(cycleLen), 0) ||
		math.IsNaN(float64(cycleLen)) {
		return polylines
	}
	startIdx, startDrawing, startRemaining := dashPhase(dasharray, offset, cycleLen)
	result := make([][]float32, 0, len(polylines)*2)
	for _, poly := range polylines {
		if len(poly) < 4 {
			continue
		}
		dashIdx := startIdx
		drawing := startDrawing
		remaining := startRemaining
		// Emitted sub-slices use cap=len to block future appends
		// from stomping retained segments.
		arena := make([]float32, 0, len(poly)*2)
		segStart := 0
		px, py := poly[0], poly[1]
		if drawing {
			arena = append(arena, px, py)
		}
		iter := 0
	walkPoly:
		for i := 2; i < len(poly); i += 2 {
			nx, ny := poly[i], poly[i+1]
			dx, dy := nx-px, ny-py
			segLen := float32(math.Sqrt(float64(dx*dx + dy*dy)))
			if segLen < 1e-6 {
				continue
			}
			consumed := float32(0)
			for consumed < segLen-1e-6 {
				if iter++; iter > maxDashIterPerPoly {
					break walkPoly
				}
				avail := segLen - consumed
				if remaining <= avail {
					t := (consumed + remaining) / segLen
					ix := px + t*dx
					iy := py + t*dy
					if drawing {
						arena = append(arena, ix, iy)
						if len(arena)-segStart >= 4 {
							end := len(arena)
							result = append(result,
								arena[segStart:end:end])
						}
						segStart = len(arena)
					} else {
						segStart = len(arena)
						arena = append(arena, ix, iy)
					}
					consumed += remaining
					drawing = !drawing
					dashIdx = (dashIdx + 1) % len(dasharray)
					remaining = dasharray[dashIdx]
				} else {
					remaining -= avail
					if drawing {
						arena = append(arena, nx, ny)
					}
					break
				}
			}
			px, py = nx, ny
		}
		if drawing && len(arena)-segStart >= 4 {
			end := len(arena)
			result = append(result, arena[segStart:end:end])
		}
	}
	return result
}

// dashPhase advances the dash cycle by offset and returns the
// starting (dashIdx, drawing, remaining) triple. Positive offset
// advances forward; negative wraps via cycleLen. NaN/Inf → 0.
func dashPhase(dasharray []float32, offset, cycleLen float32) (int, bool, float32) {
	if math.IsNaN(float64(offset)) || math.IsInf(float64(offset), 0) {
		return 0, true, dasharray[0]
	}
	// Normalize into [0, cycleLen).
	skip := float32(math.Mod(float64(offset), float64(cycleLen)))
	if skip < 0 {
		skip += cycleLen
	}
	dashIdx := 0
	drawing := true
	remaining := dasharray[0]
	for skip > remaining {
		skip -= remaining
		dashIdx = (dashIdx + 1) % len(dasharray)
		remaining = dasharray[dashIdx]
		drawing = !drawing
	}
	remaining -= skip
	// Skip past zero-length dash/gap entries so the main loop does
	// not emit degenerate zero-point segments or mis-start in the
	// wrong phase ([0,150] wants to begin in the gap).
	for remaining == 0 && len(dasharray) > 1 {
		dashIdx = (dashIdx + 1) % len(dasharray)
		remaining = dasharray[dashIdx]
		drawing = !drawing
	}
	return dashIdx, drawing, remaining
}

// --- Curve flattening ---

func flattenPath(path *VectorPath, tolerance float32) [][]float32 {
	return flattenPathWithBake(path, tolerance, !isIdentityTransform(path.Transform))
}

// flattenPathWithBake is flattenPath with explicit control over whether
// path.Transform is baked into vertex coordinates. When bakeXform is
// false, vertices emit in local (pre-transform) space — caller applies
// the transform at render time via TessellatedPath.Base* fields.
func flattenPathWithBake(path *VectorPath, tolerance float32, bakeXform bool) [][]float32 {
	var polylines [][]float32
	estimatedCap := len(path.Segments) * 16
	current := make([]float32, 0, estimatedCap)
	var x, y, startX, startY float32
	hasTx := bakeXform

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

	if len(current) >= 4 {
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

	// Multiple contours — handle holes vs. separate regions.
	// Subpaths with the same winding as the top-level contour are
	// independent filled regions (e.g. circles.svg has 4 or 8
	// separate circles in one path). Subpaths with opposite winding
	// are holes to be bridged into their enclosing region.
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

	slices.SortFunc(contours, func(a, b contour) int {
		return cmp.Compare(f32Abs(b.area), f32Abs(a.area))
	})
	outerSign := contours[0].area

	// Group contours into regions: each region is one outer plus
	// any opposite-winding holes it encloses. Same-winding contours
	// become their own regions.
	type region struct {
		outer []float32
		holes [][]float32
	}
	regions := make([]region, 0, len(contours))
	for _, c := range contours {
		if sameSignArea(c.area, outerSign) {
			poly := make([]float32, len(c.points))
			copy(poly, c.points)
			if c.area < 0 {
				poly = reversePolygon(poly)
			}
			regions = append(regions, region{outer: poly})
			continue
		}
		// Opposite winding — hole. Attach to the first region whose
		// outer contains this hole's representative (centroid) point.
		// Centroid is more robust than the first vertex when the hole
		// shares vertices with its enclosing outer. A hole that lies
		// inside no region is dropped: rather than force-merge it
		// into regions[0] and risk a malformed bridge, treat it as
		// degenerate geometry.
		hole := make([]float32, len(c.points))
		copy(hole, c.points)
		if polygonArea(hole) > 0 {
			hole = reversePolygon(hole)
		}
		hx, hy, ok := polygonRepresentativePoint(hole)
		if !ok {
			continue
		}
		for i := range regions {
			if pointInPolygon(regions[i].outer, hx, hy) {
				regions[i].holes = append(regions[i].holes, hole)
				break
			}
		}
	}

	var result []float32
	for _, r := range regions {
		outer := r.outer
		for _, hole := range r.holes {
			outer = mergeHole(outer, hole)
		}
		result = append(result, earClip(outer)...)
	}
	return result
}

// sameSignArea reports whether two signed areas share a sign
// (both > 0 or both < 0). Zero on either side is treated as
// matching to avoid spurious hole promotion on degenerate shapes.
func sameSignArea(a, b float32) bool {
	if a == 0 || b == 0 {
		return true
	}
	return (a > 0) == (b > 0)
}

// polygonRepresentativePoint returns the centroid of poly's
// vertices. Using the centroid (vs. the first vertex) is robust
// when the contour shares vertices with an enclosing outer.
// Returns ok=false when poly has fewer than 3 vertices.
func polygonRepresentativePoint(poly []float32) (float32, float32, bool) {
	n := len(poly) / 2
	if n < 3 {
		return 0, 0, false
	}
	var sx, sy float32
	for i := range n {
		sx += poly[i*2]
		sy += poly[i*2+1]
	}
	inv := 1 / float32(n)
	return sx * inv, sy * inv, true
}

// pointInPolygon reports whether (x, y) lies inside poly using a
// ray-cast odd-even test. poly is a flat [x0,y0, x1,y1, ...] slice.
func pointInPolygon(poly []float32, x, y float32) bool {
	inside := false
	n := len(poly) / 2
	j := n - 1
	for i := range n {
		xi, yi := poly[i*2], poly[i*2+1]
		xj, yj := poly[j*2], poly[j*2+1]
		if (yi > y) != (yj > y) {
			xIntersect := (xj-xi)*(y-yi)/(yj-yi) + xi
			if x < xIntersect {
				inside = !inside
			}
		}
		j = i
	}
	return inside
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

	for i := range nOuter {
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

	for i := range nOuter {
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
	for i := range nHole {
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
	for i := range n {
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
		for i := range n {
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
	for i := range n {
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

	for i := range len(indices) {
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
		minX = min(minX, x)
		maxX = max(maxX, x)
		minY = min(minY, y)
		maxY = max(maxY, y)
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
			tris[i+4], tris[i+5], grad, stopTs, 0, &result)
	}
	return result
}

func splitTriAtStops(ax, ay, bx, by, cx, cy float32, grad gui.SvgGradientDef, stopTs []float32, depth int, result *[]float32) {
	if depth >= maxSplitTriDepth {
		*result = append(*result, ax, ay, bx, by, cx, cy)
		return
	}
	ta := projectOntoGradient(ax, ay, grad)
	tb := projectOntoGradient(bx, by, grad)
	tc := projectOntoGradient(cx, cy, grad)

	tMin := ta
	tMin = min(tMin, tb, tc)
	tMax := ta
	tMax = max(tMax, tb, tc)

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
				splitTriAtStops(p0x, p0y, i2x, i2y, i1x, i1y, grad, stopTs, depth+1, result)
				splitTriAtStops(i2x, i2y, p1x, p1y, i1x, i1y, grad, stopTs, depth+1, result)
				splitTriAtStops(p1x, p1y, p2x, p2y, i1x, i1y, grad, stopTs, depth+1, result)
			} else if tS > t1+1e-4 {
				f12 := float32(0.5)
				if t2-t1 > 1e-6 {
					f12 = (tS - t1) / (t2 - t1)
				}
				i2x := p1x + f12*(p2x-p1x)
				i2y := p1y + f12*(p2y-p1y)
				splitTriAtStops(p0x, p0y, p1x, p1y, i1x, i1y, grad, stopTs, depth+1, result)
				splitTriAtStops(p1x, p1y, i2x, i2y, i1x, i1y, grad, stopTs, depth+1, result)
				splitTriAtStops(i1x, i1y, i2x, i2y, p2x, p2y, grad, stopTs, depth+1, result)
			} else {
				splitTriAtStops(p0x, p0y, p1x, p1y, i1x, i1y, grad, stopTs, depth+1, result)
				splitTriAtStops(p1x, p1y, p2x, p2y, i1x, i1y, grad, stopTs, depth+1, result)
			}
			return
		}
	}
	*result = append(*result, ax, ay, bx, by, cx, cy)
}
