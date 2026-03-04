package svg

import (
	"path/filepath"
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

// ReleaseParsed drops parser-side references for a parsed SVG once
// callers are done tessellating it.
func (p *Parser) ReleaseParsed(parsed *gui.SvgParsed) {
	if parsed == nil {
		return
	}
	p.mu.Lock()
	hash, ok := p.byParsed[parsed]
	if ok {
		delete(p.byParsed, parsed)
		delete(p.byHash, hash)
		p.removeHashFromOrder(hash)
	}
	p.mu.Unlock()
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
	clear(p.byHash)
	clear(p.byParsed)
	p.order = nil
	p.mu.Unlock()
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
	p.byHash[hash] = parserCacheEntry{parsed: result, vg: vg}
	p.byParsed[result] = hash
	p.order = append(p.order, hash)
	if len(p.order) > maxParsedRetained {
		evictHash := p.order[0]
		p.order = p.order[1:]
		if entry, ok := p.byHash[evictHash]; ok {
			delete(p.byParsed, entry.parsed)
			delete(p.byHash, evictHash)
		}
	}
	p.mu.Unlock()
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
	if entry, ok := p.byHash[hash]; ok {
		delete(p.byParsed, entry.parsed)
		delete(p.byHash, hash)
		p.removeHashFromOrder(hash)
	}
	p.mu.Unlock()
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
	for i := 0; i < len(prefix); i++ {
		h ^= uint64(prefix[i])
		h *= fnvPrime
	}
	for i := 0; i < len(src); i++ {
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
