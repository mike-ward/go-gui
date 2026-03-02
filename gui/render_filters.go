package gui

import "math"

// render_filters.go — filter pipeline infrastructure ported from
// V's render_filters.v. findFilterBracketRange already exists in
// render_types.go.

// filterTextureDims describes offscreen texture dimensions for
// a filter bracket.
type filterTextureDims struct {
	width  int
	height int
	valid  bool
}

// filterTextureDimsFromBBox computes texture dims from a bounding
// box, clamped to maxTexSize.
func filterTextureDimsFromBBox(bboxW, bboxH float32, maxTexSize int) filterTextureDims {
	if maxTexSize < 1 {
		return filterTextureDims{}
	}
	if math.IsInf(float64(bboxW), 0) || math.IsNaN(float64(bboxW)) ||
		math.IsInf(float64(bboxH), 0) || math.IsNaN(float64(bboxH)) ||
		bboxW <= 0 || bboxH <= 0 {
		return filterTextureDims{}
	}
	if bboxW > float32(maxTexSize) || bboxH > float32(maxTexSize) {
		return filterTextureDims{}
	}
	texW := int(math.Ceil(float64(bboxW)))
	texH := int(math.Ceil(float64(bboxH)))
	if texW <= 0 || texH <= 0 || texW > maxTexSize || texH > maxTexSize {
		return filterTextureDims{}
	}
	return filterTextureDims{width: texW, height: texH, valid: true}
}

// indexFilterBracketEnds builds a map from FilterBegin index to
// FilterEnd index using a stack-based scan.
func indexFilterBracketEnds(renderers []RenderCmd) map[int]int {
	endByBegin := make(map[int]int)
	var stack []int
	for i := range renderers {
		switch renderers[i].Kind {
		case RenderFilterBegin:
			stack = append(stack, i)
		case RenderFilterEnd:
			if len(stack) > 0 {
				beginIdx := stack[len(stack)-1]
				stack = stack[:len(stack)-1]
				endByBegin[beginIdx] = i
			}
		}
	}
	return endByBegin
}

// appendRendererRange appends renderers[start:end] to dst.
func appendRendererRange(dst, src []RenderCmd, start, end int) []RenderCmd {
	if len(src) == 0 || start < 0 || start >= len(src) || end <= start {
		return dst
	}
	if end > len(src) {
		end = len(src)
	}
	if end <= start {
		return dst
	}
	return append(dst, src[start:end]...)
}
