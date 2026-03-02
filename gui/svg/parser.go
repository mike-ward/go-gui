package svg

import (
	"sync"

	"github.com/mike-ward/go-gui/gui"
)

// Parser implements gui.SvgParser.
type Parser struct {
	parsed sync.Map // *gui.SvgParsed → *VectorGraphic
}

// New returns a new SVG parser.
func New() *Parser {
	return &Parser{}
}

// ParseSvg parses SVG string data.
func (p *Parser) ParseSvg(data string) (*gui.SvgParsed, error) {
	vg, err := parseSvg(data)
	if err != nil {
		return nil, err
	}
	return p.buildParsed(vg, 1), nil
}

// ParseSvgFile loads and parses an SVG file.
func (p *Parser) ParseSvgFile(path string) (*gui.SvgParsed, error) {
	vg, err := parseSvgFile(path)
	if err != nil {
		return nil, err
	}
	return p.buildParsed(vg, 1), nil
}

// ParseSvgDimensions extracts width/height without full parse.
func (p *Parser) ParseSvgDimensions(data string) (float32, float32, error) {
	w, h := parseSvgDimensions(data)
	return w, h, nil
}

// Tessellate re-tessellates at a new scale. Also updates
// parsed.FilteredGroups with re-tessellated filter group paths.
func (p *Parser) Tessellate(parsed *gui.SvgParsed, scale float32) []gui.TessellatedPath {
	val, ok := p.parsed.Load(parsed)
	if !ok {
		return nil
	}
	vg := val.(*VectorGraphic)
	parsed.FilteredGroups = tessellateFilteredGroups(vg, scale)
	return vg.getTriangles(scale)
}

func (p *Parser) buildParsed(vg *VectorGraphic, scale float32) *gui.SvgParsed {
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
	p.parsed.Store(result, vg)
	return result
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
