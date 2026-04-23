package svg

import (
	"path/filepath"
	"slices"
	"strings"
	"sync"

	"github.com/mike-ward/go-gui/gui"
)

// Parser implements gui.SvgParser.
type Parser struct {
	mu       sync.Mutex
	byHash   map[uint64]parserCacheEntry
	byParsed map[*gui.SvgParsed]uint64
	order    []uint64
}

const maxParsedRetained = 512

type parserCacheEntry struct {
	parsed *gui.SvgParsed
	vg     *VectorGraphic
}

// New returns a new SVG parser.
func New() *Parser {
	return &Parser{
		byHash:   make(map[uint64]parserCacheEntry),
		byParsed: make(map[*gui.SvgParsed]uint64),
	}
}

// ParseSvg parses SVG string data.
func (p *Parser) ParseSvg(data string) (*gui.SvgParsed, error) {
	hash := parserSourceHash(data, true)
	if parsed := p.cachedParsed(hash); parsed != nil {
		return parsed, nil
	}
	vg, err := parseSvg(data)
	if err != nil {
		return nil, err
	}
	return p.buildParsed(hash, vg, 1), nil
}

// ParseSvgFile loads and parses an SVG file.
func (p *Parser) ParseSvgFile(path string) (*gui.SvgParsed, error) {
	hash := parserSourceHash(path, false)
	if parsed := p.cachedParsed(hash); parsed != nil {
		return parsed, nil
	}
	vg, err := parseSvgFile(path)
	if err != nil {
		return nil, err
	}
	return p.buildParsed(hash, vg, 1), nil
}

// ParseSvgDimensions extracts width/height without full parse.
func (p *Parser) ParseSvgDimensions(data string) (float32, float32, error) {
	w, h := parseSvgDimensions(data)
	return w, h, nil
}

// Tessellate re-tessellates at a new scale. Also updates
// parsed.FilteredGroups with re-tessellated filter group paths.
func (p *Parser) Tessellate(parsed *gui.SvgParsed, scale float32) []gui.TessellatedPath {
	p.mu.Lock()
	hash, ok := p.byParsed[parsed]
	var entry parserCacheEntry
	if ok {
		entry, ok = p.byHash[hash]
	}
	p.mu.Unlock()
	if !ok {
		return nil
	}
	vg := entry.vg
	parsed.FilteredGroups = tessellateFilteredGroups(vg, scale)
	return vg.getTriangles(scale)
}

// TessellateAnimated implements gui.AnimatedSvgParser. Returns fresh
// triangles for every VectorPath flagged Animated at the given scale,
// applying attribute overrides keyed by GroupID. Result order follows
// the Animated-flagged paths' document order. Animated paths that
// carry a ClipPathID are skipped: the caller should fall back to
// cached triangles for them.
//
// Returns nil when overrides is empty/nil or no animated paths
// qualify. When reuse is non-nil its backing array is reused.
func (p *Parser) TessellateAnimated(
	parsed *gui.SvgParsed, scale float32,
	overrides map[string]gui.SvgAnimAttrOverride,
	reuse []gui.TessellatedPath,
) []gui.TessellatedPath {
	if len(overrides) == 0 {
		return nil
	}
	p.mu.Lock()
	hash, ok := p.byParsed[parsed]
	var entry parserCacheEntry
	if ok {
		entry, ok = p.byHash[hash]
	}
	p.mu.Unlock()
	if !ok {
		return nil
	}
	vg := entry.vg
	var animated []VectorPath
	for i := range vg.Paths {
		src := &vg.Paths[i]
		if !src.Animated {
			continue
		}
		// Skip clip-pathed animated paths; caller falls back to
		// cached triangles.
		if src.ClipPathID != "" {
			continue
		}
		ov, hasOv := overrides[src.GroupID]
		if !hasOv || ov.Mask == 0 {
			// Animated but no live overrides this frame — still
			// re-tessellate from static primitive to preserve
			// cross-frame consistency with the override branch.
			clone := *src
			animated = append(animated, clone)
			continue
		}
		clone := *src
		applyOverridesToPath(&clone, ov)
		animated = append(animated, clone)
	}
	if len(animated) == 0 {
		return nil
	}
	result := vg.tessellatePaths(animated, scale)
	if reuse != nil && cap(reuse) >= len(result) {
		reuse = reuse[:len(result)]
		copy(reuse, result)
		return reuse
	}
	return result
}

// applyOverridesToPath mutates p's primitive fields and segments to
// reflect the live animation overrides. Only primitive paths react
// to CX/CY/R/...; dash overrides apply regardless of kind since
// stroke-dasharray/offset work on any path. AdditiveMask bits add
// the override to the parsed base value; non-additive bits replace.
func applyOverridesToPath(p *VectorPath, ov gui.SvgAnimAttrOverride) {
	if ov.Mask&gui.SvgAnimMaskStrokeDashArray != 0 {
		n := min(int(ov.StrokeDashArrayLen), gui.SvgAnimDashArrayCap)
		// Fresh alloc required: clone shares backing with cached
		// src; in-place mutation would corrupt the cache.
		p.StrokeDasharray = slices.Clone(ov.StrokeDashArray[:n])
	}
	if ov.Mask&gui.SvgAnimMaskStrokeDashOffset != 0 {
		if ov.AdditiveMask&gui.SvgAnimMaskStrokeDashOffset != 0 {
			p.StrokeDashOffset += ov.StrokeDashOffset
		} else {
			p.StrokeDashOffset = ov.StrokeDashOffset
		}
	}
	prim := p.Primitive
	switch prim.Kind {
	case gui.SvgPrimCircle:
		prim.CX = overrideScalar(prim.CX, ov.CX, &ov, gui.SvgAnimMaskCX)
		prim.CY = overrideScalar(prim.CY, ov.CY, &ov, gui.SvgAnimMaskCY)
		prim.R = overrideScalar(prim.R, ov.R, &ov, gui.SvgAnimMaskR)
		p.Segments = segmentsForEllipse(prim.CX, prim.CY, prim.R, prim.R)
	case gui.SvgPrimEllipse:
		prim.CX = overrideScalar(prim.CX, ov.CX, &ov, gui.SvgAnimMaskCX)
		prim.CY = overrideScalar(prim.CY, ov.CY, &ov, gui.SvgAnimMaskCY)
		prim.RX = overrideScalar(prim.RX, ov.RX, &ov, gui.SvgAnimMaskRX)
		prim.RY = overrideScalar(prim.RY, ov.RY, &ov, gui.SvgAnimMaskRY)
		p.Segments = segmentsForEllipse(prim.CX, prim.CY, prim.RX, prim.RY)
	case gui.SvgPrimRect:
		prim.X = overrideScalar(prim.X, ov.X, &ov, gui.SvgAnimMaskX)
		prim.Y = overrideScalar(prim.Y, ov.Y, &ov, gui.SvgAnimMaskY)
		prim.W = overrideScalar(prim.W, ov.Width, &ov,
			gui.SvgAnimMaskWidth)
		prim.H = overrideScalar(prim.H, ov.Height, &ov,
			gui.SvgAnimMaskHeight)
		prim.RX = overrideScalar(prim.RX, ov.RX, &ov, gui.SvgAnimMaskRX)
		prim.RY = overrideScalar(prim.RY, ov.RY, &ov, gui.SvgAnimMaskRY)
		p.Segments = segmentsForRect(prim.X, prim.Y, prim.W, prim.H,
			prim.RX, prim.RY)
	case gui.SvgPrimLine:
		prim.X = overrideScalar(prim.X, ov.X, &ov, gui.SvgAnimMaskX)
		prim.Y = overrideScalar(prim.Y, ov.Y, &ov, gui.SvgAnimMaskY)
		p.Segments = segmentsForLine(prim.X, prim.Y, prim.X2, prim.Y2)
	}
	p.Primitive = prim
}

// overrideScalar returns base when the mask bit is unset, base+v
// for additive overrides, or v for replace overrides.
func overrideScalar(base, v float32, ov *gui.SvgAnimAttrOverride,
	bit gui.SvgAnimAttrMask) float32 {
	if ov.Mask&bit == 0 {
		return base
	}
	if ov.AdditiveMask&bit != 0 {
		return base + v
	}
	return v
}

// ReleaseParsed drops parser-side references for a parsed SVG once
// callers are done tessellating it.
func (p *Parser) ReleaseParsed(parsed *gui.SvgParsed) {
	if parsed == nil {
		return
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	hash, ok := p.byParsed[parsed]
	if ok {
		delete(p.byParsed, parsed)
		delete(p.byHash, hash)
		p.removeHashFromOrder(hash)
	}
}

// InvalidateSvgSource invalidates parser cache for one SVG source.
func (p *Parser) InvalidateSvgSource(svgSrc string) {
	inline := strings.HasPrefix(svgSrc, "<")
	if inline {
		p.removeHash(parserSourceHash(svgSrc, true))
		return
	}
	p.removeHash(parserSourceHash(svgSrc, false))
	clean := filepath.Clean(svgSrc)
	if abs, err := filepath.Abs(clean); err == nil {
		p.removeHash(parserSourceHash(abs, false))
		if resolved, err := filepath.EvalSymlinks(abs); err == nil {
			p.removeHash(parserSourceHash(resolved, false))
		}
	}
}

// ClearSvgParserCache removes all retained parsed entries.
func (p *Parser) ClearSvgParserCache() {
	p.mu.Lock()
	defer p.mu.Unlock()
	clear(p.byHash)
	clear(p.byParsed)
	p.order = nil
}

func (p *Parser) buildParsed(hash uint64, vg *VectorGraphic, scale float32) *gui.SvgParsed {
	tpaths := vg.getTriangles(scale)
	result := &gui.SvgParsed{
		Paths:          tpaths,
		Texts:          vg.Texts,
		TextPaths:      vg.TextPaths,
		DefsPaths:      vg.DefsPaths,
		Gradients:      vg.Gradients,
		FilteredGroups: tessellateFilteredGroups(vg, scale),
		Animations:     vg.Animations,
		Width:          vg.Width,
		Height:         vg.Height,
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	p.byHash[hash] = parserCacheEntry{parsed: result, vg: vg}
	p.byParsed[result] = hash
	p.order = append(p.order, hash)
	if len(p.order) > maxParsedRetained {
		evictHash := p.order[0]
		p.order = p.order[1:]
		if len(p.order) < cap(p.order)/2 {
			p.order = append([]uint64(nil), p.order...)
		}
		if entry, ok := p.byHash[evictHash]; ok {
			delete(p.byParsed, entry.parsed)
			delete(p.byHash, evictHash)
		}
	}
	return result
}

func (p *Parser) cachedParsed(hash uint64) *gui.SvgParsed {
	p.mu.Lock()
	defer p.mu.Unlock()
	entry, ok := p.byHash[hash]
	if !ok {
		return nil
	}
	p.touchHash(hash)
	return entry.parsed
}

func (p *Parser) touchHash(hash uint64) {
	for i := len(p.order) - 1; i >= 0; i-- {
		if p.order[i] == hash {
			copy(p.order[i:], p.order[i+1:])
			p.order[len(p.order)-1] = hash
			return
		}
	}
	p.order = append(p.order, hash)
}

func (p *Parser) removeHashFromOrder(hash uint64) {
	for i := range p.order {
		if p.order[i] == hash {
			p.order = append(p.order[:i], p.order[i+1:]...)
			return
		}
	}
}

func (p *Parser) removeHash(hash uint64) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if entry, ok := p.byHash[hash]; ok {
		delete(p.byParsed, entry.parsed)
		delete(p.byHash, hash)
		p.removeHashFromOrder(hash)
	}
}

func parserSourceHash(src string, inline bool) uint64 {
	const (
		fnvOffset = uint64(0xcbf29ce484222325)
		fnvPrime  = uint64(0x100000001b3)
	)
	h := fnvOffset
	prefix := "file:"
	if inline {
		prefix = "inline:"
	}
	for i := range len(prefix) {
		h ^= uint64(prefix[i])
		h *= fnvPrime
	}
	for i := range len(src) {
		h ^= uint64(src[i])
		h *= fnvPrime
	}
	return h
}

func tessellateFilteredGroups(vg *VectorGraphic, scale float32) []gui.SvgParsedFilteredGroup {
	if len(vg.FilteredGroups) == 0 {
		return nil
	}
	groups := make([]gui.SvgParsedFilteredGroup, 0, len(vg.FilteredGroups))
	for _, fg := range vg.FilteredGroups {
		filter := vg.Filters[fg.FilterID]
		tpaths := vg.tessellatePaths(fg.Paths, scale)
		groups = append(groups, gui.SvgParsedFilteredGroup{
			Filter:    filter,
			Paths:     tpaths,
			Texts:     fg.Texts,
			TextPaths: fg.TextPaths,
		})
	}
	return groups
}
