package svg

import (
	"fmt"
	"log"
	"math"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"

	"github.com/mike-ward/go-gui/gui"
)

// svgTrace enables per-frame diagnostic logging of animated SVG
// primitive overrides and geometry bounds. Set GOGUI_SVG_TRACE=1 to
// activate. Off by default; evaluated once per process.
var svgTrace = os.Getenv("GOGUI_SVG_TRACE") == "1"

// Parser implements gui.SvgParser.
type Parser struct {
	mu              sync.Mutex
	byHash          map[uint64]parserCacheEntry
	byParsed        map[*gui.SvgParsed]uint64
	order           []uint64
	animatedScratch sync.Pool
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
	return p.ParseSvgWithOpts(data, gui.SvgParseOpts{})
}

// ParseSvgFile loads and parses an SVG file.
func (p *Parser) ParseSvgFile(path string) (*gui.SvgParsed, error) {
	return p.ParseSvgFileWithOpts(path, gui.SvgParseOpts{})
}

// ParseSvgWithOpts parses SVG string data, snapshotting environment
// flags (e.g. prefers-reduced-motion) into the cascade. Implements
// gui.SvgParserWithOpts.
func (p *Parser) ParseSvgWithOpts(
	data string, opts gui.SvgParseOpts,
) (*gui.SvgParsed, error) {
	hash := parserSourceHashWithOpts(data, true, opts)
	if parsed := p.cachedParsed(hash); parsed != nil {
		return parsed, nil
	}
	vg, err := parseSvgWith(data, optsToSvg(opts))
	if err != nil {
		return nil, err
	}
	return p.buildParsed(hash, vg, 1), nil
}

// ParseSvgFileWithOpts loads and parses an SVG file with options.
// Implements gui.SvgParserWithOpts.
func (p *Parser) ParseSvgFileWithOpts(
	path string, opts gui.SvgParseOpts,
) (*gui.SvgParsed, error) {
	fileData, err := loadSvgFile(path)
	if err != nil {
		return nil, err
	}
	hash := mixOptsHash(parserFileHash(path, fileData), opts)
	if parsed := p.cachedParsed(hash); parsed != nil {
		return parsed, nil
	}
	vg, err := parseSvgWith(string(fileData), optsToSvg(opts))
	if err != nil {
		return nil, err
	}
	return p.buildParsed(hash, vg, 1), nil
}

func optsToSvg(o gui.SvgParseOpts) ParseOptions {
	return ParseOptions{PrefersReducedMotion: o.PrefersReducedMotion}
}

func parserSourceHashWithOpts(
	src string, inline bool, opts gui.SvgParseOpts,
) uint64 {
	return mixOptsHash(parserSourceHash(src, inline), opts)
}

func mixOptsHash(h uint64, opts gui.SvgParseOpts) uint64 {
	const fnvPrime = uint64(0x100000001b3)
	var b byte
	if opts.PrefersReducedMotion {
		b = 1
	}
	h ^= uint64(b)
	h *= fnvPrime
	return h
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
// applying attribute overrides keyed by PathID. Result order follows
// the Animated-flagged paths' document order. Animated paths that
// carry a ClipPathID are skipped: the caller should fall back to
// cached triangles for them.
//
// Returns nil when overrides is empty/nil or no animated paths
// qualify. When reuse is non-nil its backing array is reused.
func (p *Parser) TessellateAnimated(
	parsed *gui.SvgParsed, scale float32,
	overrides map[uint32]gui.SvgAnimAttrOverride,
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
	totalCap := len(vg.Paths)
	for i := range vg.FilteredGroups {
		totalCap += len(vg.FilteredGroups[i].Paths)
	}
	animated := p.getAnimatedScratch(totalCap)
	animated = collectAnimatedPaths(animated, vg.Paths, overrides)
	for gi := range vg.FilteredGroups {
		animated = collectAnimatedPaths(
			animated, vg.FilteredGroups[gi].Paths, overrides)
	}
	if len(animated) == 0 {
		p.putAnimatedScratch(animated)
		return nil
	}
	result := vg.tessellatePaths(animated, scale)
	if svgTrace {
		traceAnimatedTriangles(vg, result, animated, overrides)
	}
	if reuse != nil && cap(reuse) >= len(result) {
		reuse = reuse[:len(result)]
		copy(reuse, result)
		p.putAnimatedScratch(animated)
		return reuse
	}
	p.putAnimatedScratch(animated)
	return result
}

// collectAnimatedPaths appends clones of every Animated, non-clip-
// pathed entry in src into dst, applying any matching attribute
// override. Inlined (rather than a closure) so the hot path does not
// allocate a closure capturing the override map per call.
func collectAnimatedPaths(dst []VectorPath, src []VectorPath,
	overrides map[uint32]gui.SvgAnimAttrOverride) []VectorPath {
	for i := range src {
		s := &src[i]
		if !s.Animated || s.ClipPathID != "" {
			continue
		}
		clone := *s
		if ov, ok := overrides[s.PathID]; ok && ov.Mask != 0 {
			applyOverridesToPath(&clone, ov)
		}
		dst = append(dst, clone)
	}
	return dst
}

func (p *Parser) getAnimatedScratch(minCap int) []VectorPath {
	if v := p.animatedScratch.Get(); v != nil {
		if buf, ok := v.(*[]VectorPath); ok && cap(*buf) >= minCap {
			return (*buf)[:0]
		}
	}
	return make([]VectorPath, 0, minCap)
}

func (p *Parser) putAnimatedScratch(buf []VectorPath) {
	if cap(buf) == 0 {
		return
	}
	for i := range buf {
		buf[i] = VectorPath{}
	}
	buf = buf[:0]
	p.animatedScratch.Put(&buf)
}

func traceOverride(p *VectorPath, ov gui.SvgAnimAttrOverride) {
	check := func(name string, bit gui.SvgAnimAttrMask, v float32) {
		if ov.Mask&bit == 0 {
			return
		}
		if !finiteF32(v) || v < -1e4 || v > 1e4 {
			log.Printf("svg trace: gid=%q attr=%s val=%v "+
				"mask=%b additive=%b",
				p.GroupID, name, v, ov.Mask, ov.AdditiveMask)
		}
	}
	check("cx", gui.SvgAnimMaskCX, ov.CX)
	check("cy", gui.SvgAnimMaskCY, ov.CY)
	check("r", gui.SvgAnimMaskR, ov.R)
	check("rx", gui.SvgAnimMaskRX, ov.RX)
	check("ry", gui.SvgAnimMaskRY, ov.RY)
	check("x", gui.SvgAnimMaskX, ov.X)
	check("y", gui.SvgAnimMaskY, ov.Y)
	check("width", gui.SvgAnimMaskWidth, ov.Width)
	check("height", gui.SvgAnimMaskHeight, ov.Height)
}

// traceAnimatedTriangles logs animated paths whose bbox escapes 2x
// viewBox — diagnostic for spurious full-cell fills.
func traceAnimatedTriangles(vg *VectorGraphic,
	paths []gui.TessellatedPath, animated []VectorPath,
	overrides map[uint32]gui.SvgAnimAttrOverride,
) {
	xLim := vg.Width * 2
	yLim := vg.Height * 2
	for i := range paths {
		tris := paths[i].Triangles
		if len(tris) == 0 {
			continue
		}
		minX, minY := tris[0], tris[1]
		maxX, maxY := minX, minY
		for j := 2; j+1 < len(tris); j += 2 {
			if tris[j] < minX {
				minX = tris[j]
			}
			if tris[j] > maxX {
				maxX = tris[j]
			}
			if tris[j+1] < minY {
				minY = tris[j+1]
			}
			if tris[j+1] > maxY {
				maxY = tris[j+1]
			}
		}
		if finiteF32(minX) && finiteF32(minY) &&
			finiteF32(maxX) && finiteF32(maxY) &&
			maxX-minX <= xLim && maxY-minY <= yLim {
			continue
		}
		var primStr, ovStr string
		if i < len(animated) {
			p := &animated[i]
			primStr = fmt.Sprintf("prim={Kind:%d CX:%.3f CY:%.3f "+
				"R:%.3f RX:%.3f RY:%.3f X:%.3f Y:%.3f W:%.3f H:%.3f}",
				p.Primitive.Kind, p.Primitive.CX, p.Primitive.CY,
				p.Primitive.R, p.Primitive.RX, p.Primitive.RY,
				p.Primitive.X, p.Primitive.Y,
				p.Primitive.W, p.Primitive.H)
			if ov, ok := overrides[p.PathID]; ok {
				ovStr = fmt.Sprintf("ov={Mask:%b Add:%b CX:%.3f CY:%.3f "+
					"R:%.3f RX:%.3f RY:%.3f X:%.3f Y:%.3f W:%.3f H:%.3f}",
					ov.Mask, ov.AdditiveMask, ov.CX, ov.CY,
					ov.R, ov.RX, ov.RY, ov.X, ov.Y, ov.Width, ov.Height)
			}
		}
		log.Printf("svg trace: pid=%d oversized tris "+
			"bbox=(%.2f,%.2f)-(%.2f,%.2f) vb=%.0fx%.0f nTris=%d %s %s",
			paths[i].PathID, minX, minY, maxX, maxY,
			vg.Width, vg.Height, len(tris)/6, primStr, ovStr)
	}
}

// applyOverridesToPath mutates p's primitive fields and segments to
// reflect the live animation overrides. Only primitive paths react
// to CX/CY/R/...; dash overrides apply regardless of kind since
// stroke-dasharray/offset work on any path. AdditiveMask bits add
// the override to the parsed base value; non-additive bits replace.
func applyOverridesToPath(p *VectorPath, ov gui.SvgAnimAttrOverride) {
	if svgTrace {
		traceOverride(p, ov)
	}
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
		prim.R = nonNegF32(overrideScalar(prim.R, ov.R, &ov,
			gui.SvgAnimMaskR))
		p.Segments = segmentsForEllipse(prim.CX, prim.CY, prim.R, prim.R)
	case gui.SvgPrimEllipse:
		prim.CX = overrideScalar(prim.CX, ov.CX, &ov, gui.SvgAnimMaskCX)
		prim.CY = overrideScalar(prim.CY, ov.CY, &ov, gui.SvgAnimMaskCY)
		prim.RX = nonNegF32(overrideScalar(prim.RX, ov.RX, &ov,
			gui.SvgAnimMaskRX))
		prim.RY = nonNegF32(overrideScalar(prim.RY, ov.RY, &ov,
			gui.SvgAnimMaskRY))
		p.Segments = segmentsForEllipse(prim.CX, prim.CY, prim.RX, prim.RY)
	case gui.SvgPrimRect:
		prim.X = overrideScalar(prim.X, ov.X, &ov, gui.SvgAnimMaskX)
		prim.Y = overrideScalar(prim.Y, ov.Y, &ov, gui.SvgAnimMaskY)
		prim.W = nonNegF32(overrideScalar(prim.W, ov.Width, &ov,
			gui.SvgAnimMaskWidth))
		prim.H = nonNegF32(overrideScalar(prim.H, ov.Height, &ov,
			gui.SvgAnimMaskHeight))
		prim.RX = nonNegF32(overrideScalar(prim.RX, ov.RX, &ov,
			gui.SvgAnimMaskRX))
		prim.RY = nonNegF32(overrideScalar(prim.RY, ov.RY, &ov,
			gui.SvgAnimMaskRY))
		p.Segments = segmentsForRect(prim.X, prim.Y, prim.W, prim.H,
			prim.RX, prim.RY)
	case gui.SvgPrimLine:
		prim.X = overrideScalar(prim.X, ov.X, &ov, gui.SvgAnimMaskX)
		prim.Y = overrideScalar(prim.Y, ov.Y, &ov, gui.SvgAnimMaskY)
		p.Segments = segmentsForLine(prim.X, prim.Y, prim.X2, prim.Y2)
	}
	p.Primitive = prim
}

// overrideScalar returns base, base+v (additive), or v (replace).
// Non-finite v falls back to base — would otherwise tessellate into
// huge/fullscreen triangles.
func overrideScalar(base, v float32, ov *gui.SvgAnimAttrOverride,
	bit gui.SvgAnimAttrMask) float32 {
	if ov.Mask&bit == 0 {
		return base
	}
	if !finiteF32(v) {
		return base
	}
	if ov.AdditiveMask&bit != 0 {
		return base + v
	}
	return v
}

func finiteF32(f float32) bool {
	return !math.IsNaN(float64(f)) && !math.IsInf(float64(f), 0)
}

// nonNegF32 maps NaN / negative → 0. Negative R/W/H tessellate with
// reversed winding under non-zero fill — i.e. the whole cell.
func nonNegF32(v float32) float32 {
	if v != v || v < 0 {
		return 0
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
// All option-variants of the source (e.g. with/without
// prefers-reduced-motion) are dropped together since the source
// string is the user-visible identity.
func (p *Parser) InvalidateSvgSource(svgSrc string) {
	dropVariants := func(base uint64) {
		p.removeHash(mixOptsHash(base, gui.SvgParseOpts{}))
		p.removeHash(mixOptsHash(base, gui.SvgParseOpts{
			PrefersReducedMotion: true,
		}))
	}
	inline := strings.HasPrefix(svgSrc, "<")
	if inline {
		dropVariants(parserSourceHash(svgSrc, true))
		return
	}
	dropVariants(parserSourceHash(svgSrc, false))
	clean := filepath.Clean(svgSrc)
	if abs, err := filepath.Abs(clean); err == nil {
		dropVariants(parserSourceHash(abs, false))
		if resolved, err := filepath.EvalSymlinks(abs); err == nil {
			dropVariants(parserSourceHash(resolved, false))
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
	resolveAnimationTargets(vg)
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
		ViewBoxX:       vg.ViewBoxX,
		ViewBoxY:       vg.ViewBoxY,
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

func parserFileHash(path string, data []byte) uint64 {
	h := parserSourceHash(string(data), true)
	clean := filepath.Clean(path)
	if abs, err := filepath.Abs(clean); err == nil {
		return mixHash(h, abs)
	}
	return mixHash(h, clean)
}

func mixHash(h uint64, extra string) uint64 {
	const fnvPrime = uint64(0x100000001b3)
	for i := range len(extra) {
		h ^= uint64(extra[i])
		h *= fnvPrime
	}
	return h
}

// maxGroupParentDepth caps GroupParent ancestor walks. Author-id
// collisions could in theory form a cycle; combined with the per-walk
// visited set this guards against pathological inputs.
const maxGroupParentDepth = 64

// resolveAnimationTargets populates each SvgAnimation.TargetPathIDs
// by scanning vg.Paths (and filtered-group paths) for VectorPaths
// whose GroupID matches the animation's binding. A single shape-id
// match yields one PathID; a group-id match yields every descendant
// primitive's PathID, enabling per-path animation routing without
// sibling-collision collapses.
func resolveAnimationTargets(vg *VectorGraphic) {
	if len(vg.Animations) == 0 {
		return
	}
	byGroup := make(map[string][]uint32, len(vg.GroupParent)+8)
	visited := make(map[string]struct{}, 8)
	collect := func(paths []VectorPath) {
		for i := range paths {
			p := &paths[i]
			if p.GroupID == "" || p.PathID == 0 {
				continue
			}
			gid := p.GroupID
			clear(visited)
			for depth := 0; gid != "" && depth < maxGroupParentDepth; depth++ {
				if _, seen := visited[gid]; seen {
					break
				}
				visited[gid] = struct{}{}
				byGroup[gid] = append(byGroup[gid], p.PathID)
				parent, ok := vg.GroupParent[gid]
				if !ok || parent == gid {
					break
				}
				gid = parent
			}
		}
	}
	collect(vg.Paths)
	for i := range vg.FilteredGroups {
		collect(vg.FilteredGroups[i].Paths)
	}
	for i := range vg.Animations {
		a := &vg.Animations[i]
		if a.GroupID == "" {
			continue
		}
		a.TargetPathIDs = byGroup[a.GroupID]
	}
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
