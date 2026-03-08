package gui

// scratch_pools.go — reusable per-frame buffers. Zero-value valid.

// Retain/shrink thresholds matching V's scratch_pools.v.
const (
	scratchFilterRenderersRetainMax     = 131_072
	scratchFilterRenderersShrinkTo      = 8192
	scratchFloatingLayoutsRetainMax     = 4096
	scratchFloatingLayoutsShrinkTo      = 256
	scratchFloatingPoolRetainMax        = 512
	scratchFloatingPoolShrinkTo         = 64
	scratchFocusCandidatesRetainMax     = 4096
	scratchFocusCandidatesShrinkTo      = 512
	scratchGradientNormRetainMax        = 64
	scratchGradientNormShrinkTo         = 8
	scratchSvgAnimValsRetainMax         = 32
	scratchSvgAnimValsShrinkTo          = 8
	scratchSvgTrisRetainMax             = 1_048_576
	scratchSvgTrisShrinkTo              = 8192
	scratchSvgGroupMatricesRetainMax    = 4096
	scratchSvgGroupOpacitiesRetainMax   = 4096
	scratchSvgTransformBatchesRetainMax = 128
	scratchSvgTransformBatchesShrinkTo  = 16
	scratchWrapRowsRetainMax            = 4096
	scratchWrapRowsShrinkTo             = 256
	scratchLayerLayoutsRetainMax        = 4096
	scratchLayerLayoutsShrinkTo         = 256
	scratchFocusSeenRetainMax           = 4096
	scratchPlaceholderPoolRetainMax     = 4096
	scratchPlaceholderPoolShrinkTo      = 256
)

// scratchPools holds reusable per-frame buffers.
type scratchPools struct {
	filterRenderers         []RenderCmd
	floatingLayouts         []*Layout
	floatingLayoutPool      []*Layout
	floatingPoolUsed        int
	focusCandidates         []focusCandidate
	focusSeen               map[uint32]struct{}
	gradientNormStops       []GradientStop
	gradientSampleStops     []GradientStop
	svgAnimVals             []float32
	svgAnimStates           map[string]svgAnimState
	svgGroupMatrices        map[string][6]float32
	svgGroupOpacities       map[string]float32
	svgTransformBatches     [][]float32
	svgTransformBatchesUsed int
	wrapRows                []wrapRowRange
	layerLayouts            []Layout
	placeholderShapePool    []*Shape
	placeholderPoolUsed     int
}

func (p *scratchPools) takeFilterRenderers(requiredCap int) []RenderCmd {
	s := p.filterRenderers
	s = s[:0]
	if cap(s) < requiredCap {
		s = make([]RenderCmd, 0, requiredCap)
	}
	return s
}

func (p *scratchPools) putFilterRenderers(s []RenderCmd) {
	if cap(s) > scratchFilterRenderersRetainMax {
		s = make([]RenderCmd, 0, scratchFilterRenderersShrinkTo)
	}
	p.filterRenderers = s[:0]
}

func (p *scratchPools) takeFloatingLayouts(requiredCap int) []*Layout {
	s := p.floatingLayouts
	s = s[:0]
	if cap(s) < requiredCap {
		s = make([]*Layout, 0, requiredCap)
	}
	p.floatingPoolUsed = 0
	p.placeholderPoolUsed = 0
	return s
}

func (p *scratchPools) putFloatingLayouts(s []*Layout) {
	if cap(s) > scratchFloatingLayoutsRetainMax {
		s = make([]*Layout, 0, scratchFloatingLayoutsShrinkTo)
	}
	p.floatingLayouts = s[:0]
	if len(p.floatingLayoutPool) > scratchFloatingPoolRetainMax {
		p.floatingLayoutPool = make([]*Layout, 0, scratchFloatingPoolShrinkTo)
	}
	if len(p.placeholderShapePool) > scratchPlaceholderPoolRetainMax {
		p.placeholderShapePool = make([]*Shape, 0, scratchPlaceholderPoolShrinkTo)
	}
}

func (p *scratchPools) allocFloatingLayout(src Layout) *Layout {
	idx := p.floatingPoolUsed
	p.floatingPoolUsed++
	if idx < len(p.floatingLayoutPool) {
		reused := p.floatingLayoutPool[idx]
		*reused = src
		return reused
	}
	cp := src
	allocated := &cp
	p.floatingLayoutPool = append(p.floatingLayoutPool, allocated)
	return allocated
}

func (p *scratchPools) allocPlaceholderShape() *Shape {
	idx := p.placeholderPoolUsed
	p.placeholderPoolUsed++
	if idx < len(p.placeholderShapePool) {
		reused := p.placeholderShapePool[idx]
		*reused = Shape{ShapeType: ShapeNone}
		return reused
	}
	allocated := &Shape{ShapeType: ShapeNone}
	p.placeholderShapePool = append(p.placeholderShapePool, allocated)
	return allocated
}

func (p *scratchPools) takeFocusCandidates() []focusCandidate {
	s := p.focusCandidates
	s = s[:0]
	return s
}

func (p *scratchPools) putFocusCandidates(s []focusCandidate) {
	if cap(s) > scratchFocusCandidatesRetainMax {
		s = make([]focusCandidate, 0, scratchFocusCandidatesShrinkTo)
	}
	p.focusCandidates = s[:0]
}

func (p *scratchPools) takeFocusSeen(requiredCap int) map[uint32]struct{} {
	m := p.focusSeen
	if m == nil {
		if requiredCap < 8 {
			requiredCap = 8
		}
		m = make(map[uint32]struct{}, requiredCap)
	}
	clear(m)
	return m
}

func (p *scratchPools) putFocusSeen(m map[uint32]struct{}) {
	if len(m) > scratchFocusSeenRetainMax {
		p.focusSeen = nil
		return
	}
	p.focusSeen = m
}

func (p *scratchPools) takeGradientNormStops(requiredCap int) []GradientStop {
	s := p.gradientNormStops
	s = s[:0]
	if cap(s) < requiredCap {
		s = make([]GradientStop, 0, requiredCap)
	}
	return s
}

func (p *scratchPools) putGradientNormStops(s []GradientStop) {
	if cap(s) > scratchGradientNormRetainMax {
		s = make([]GradientStop, 0, scratchGradientNormShrinkTo)
	}
	p.gradientNormStops = s[:0]
}

func (p *scratchPools) takeGradientSampleStops(requiredCap int) []GradientStop {
	s := p.gradientSampleStops
	s = s[:0]
	if cap(s) < requiredCap {
		s = make([]GradientStop, 0, requiredCap)
	}
	return s
}

func (p *scratchPools) putGradientSampleStops(s []GradientStop) {
	if cap(s) > scratchGradientNormRetainMax {
		s = make([]GradientStop, 0, scratchGradientNormShrinkTo)
	}
	p.gradientSampleStops = s[:0]
}

func (p *scratchPools) takeSvgAnimVals(requiredCap int) []float32 {
	s := p.svgAnimVals
	s = s[:0]
	if cap(s) < requiredCap {
		s = make([]float32, 0, requiredCap)
	}
	return s
}

func (p *scratchPools) putSvgAnimVals(s []float32) {
	if cap(s) > scratchSvgAnimValsRetainMax {
		s = make([]float32, 0, scratchSvgAnimValsShrinkTo)
	}
	p.svgAnimVals = s[:0]
}

func (p *scratchPools) takeSvgAnimStates(requiredCap int) map[string]svgAnimState {
	m := p.svgAnimStates
	if m == nil {
		if requiredCap < 8 {
			requiredCap = 8
		}
		m = make(map[string]svgAnimState, requiredCap)
	}
	clear(m)
	return m
}

func (p *scratchPools) putSvgAnimStates(m map[string]svgAnimState) {
	if len(m) > scratchSvgGroupOpacitiesRetainMax {
		p.svgAnimStates = nil
		return
	}
	p.svgAnimStates = m
}

func (p *scratchPools) takeWrapRows(requiredCap int) []wrapRowRange {
	s := p.wrapRows
	s = s[:0]
	if cap(s) < requiredCap {
		s = make([]wrapRowRange, 0, requiredCap)
	}
	return s
}

func (p *scratchPools) putWrapRows(s []wrapRowRange) {
	if cap(s) > scratchWrapRowsRetainMax {
		s = make([]wrapRowRange, 0, scratchWrapRowsShrinkTo)
	}
	p.wrapRows = s[:0]
}

func (p *scratchPools) takeLayerLayouts(requiredCap int) []Layout {
	s := p.layerLayouts
	s = s[:0]
	if cap(s) < requiredCap {
		s = make([]Layout, 0, requiredCap)
	}
	return s
}

func (p *scratchPools) putLayerLayouts(s []Layout) {
	if cap(s) > scratchLayerLayoutsRetainMax {
		s = make([]Layout, 0, scratchLayerLayoutsShrinkTo)
	}
	p.layerLayouts = s[:0]
}

func (p *scratchPools) transformSvgTriangles(tris []float32, m [6]float32) []float32 {
	idx := p.svgTransformBatchesUsed
	p.svgTransformBatchesUsed++
	if idx >= len(p.svgTransformBatches) {
		p.svgTransformBatches = append(p.svgTransformBatches, make([]float32, 0, len(tris)))
	}
	out := p.svgTransformBatches[idx]
	if cap(out) < len(tris) {
		out = make([]float32, 0, len(tris))
	} else {
		out = out[:0]
	}
	out = applyTransformToTrianglesInto(tris, m, out)
	p.svgTransformBatches[idx] = out
	return out
}

func (p *scratchPools) trimSvgGroupMaps() {
	if len(p.svgGroupMatrices) > scratchSvgGroupMatricesRetainMax {
		p.svgGroupMatrices = nil
	}
	if len(p.svgGroupOpacities) > scratchSvgGroupOpacitiesRetainMax {
		p.svgGroupOpacities = nil
	}
}

func (p *scratchPools) trimSvgTransformBatches() {
	for i := range p.svgTransformBatches {
		if cap(p.svgTransformBatches[i]) > scratchSvgTrisRetainMax {
			p.svgTransformBatches[i] = make([]float32, 0, scratchSvgTrisShrinkTo)
		}
	}
	if len(p.svgTransformBatches) > scratchSvgTransformBatchesRetainMax {
		p.svgTransformBatches = p.svgTransformBatches[:scratchSvgTransformBatchesShrinkTo]
	}
}
