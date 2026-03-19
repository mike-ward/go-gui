package gui

// scratch_pools.go — reusable per-frame buffers. Zero-value valid.

// scratchSlice is a reusable slice pool with retain/shrink thresholds.
type scratchSlice[T any] struct {
	buf       []T
	retainMax int
	shrinkTo  int
}

func (s *scratchSlice[T]) take(requiredCap int) []T {
	b := s.buf
	b = b[:0]
	if cap(b) < requiredCap {
		b = make([]T, 0, requiredCap)
	}
	return b
}

func (s *scratchSlice[T]) put(b []T) {
	if cap(b) > s.retainMax {
		b = make([]T, 0, s.shrinkTo)
	}
	s.buf = b[:0]
}

// scratchMap is a reusable map pool with a retain threshold.
type scratchMap[K comparable, V any] struct {
	m         map[K]V
	retainMax int
}

func (s *scratchMap[K, V]) take(requiredCap int) map[K]V {
	m := s.m
	if m == nil {
		if requiredCap < 8 {
			requiredCap = 8
		}
		m = make(map[K]V, requiredCap)
	}
	clear(m)
	return m
}

func (s *scratchMap[K, V]) put(m map[K]V) {
	if len(m) > s.retainMax {
		s.m = nil
		return
	}
	s.m = m
}

// scratchPools holds reusable per-frame buffers.
type scratchPools struct {
	filterRenderers     scratchSlice[RenderCmd]
	focusCandidates     scratchSlice[focusCandidate]
	gradientNormStops   scratchSlice[GradientStop]
	gradientSampleStops scratchSlice[GradientStop]
	svgAnimVals         scratchSlice[float32]
	wrapRows            scratchSlice[wrapRowRange]
	layerLayouts        scratchSlice[Layout]

	focusSeen     scratchMap[uint32, struct{}]
	svgAnimStates scratchMap[string, svgAnimState]

	floatingLayouts      []*Layout
	floatingLayoutPool   []*Layout
	floatingPoolUsed     int
	placeholderShapePool []*Shape
	placeholderPoolUsed  int
}

const (
	scratchFloatingLayoutsRetainMax = 4096
	scratchFloatingLayoutsShrinkTo  = 256
	scratchFloatingPoolRetainMax    = 512
	scratchFloatingPoolShrinkTo     = 64
	scratchPlaceholderPoolRetainMax = 4096
	scratchPlaceholderPoolShrinkTo  = 256
)

func newScratchPools() scratchPools {
	return scratchPools{
		filterRenderers:     scratchSlice[RenderCmd]{retainMax: 131_072, shrinkTo: 8192},
		focusCandidates:     scratchSlice[focusCandidate]{retainMax: 4096, shrinkTo: 512},
		gradientNormStops:   scratchSlice[GradientStop]{retainMax: 64, shrinkTo: 8},
		gradientSampleStops: scratchSlice[GradientStop]{retainMax: 64, shrinkTo: 8},
		svgAnimVals:         scratchSlice[float32]{retainMax: 32, shrinkTo: 8},
		wrapRows:            scratchSlice[wrapRowRange]{retainMax: 4096, shrinkTo: 256},
		layerLayouts:        scratchSlice[Layout]{retainMax: 4096, shrinkTo: 256},
		focusSeen:           scratchMap[uint32, struct{}]{retainMax: 4096},
		svgAnimStates:       scratchMap[string, svgAnimState]{retainMax: 4096},
	}
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
