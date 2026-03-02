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

// Tessellate re-tessellates at a new scale.
func (p *Parser) Tessellate(parsed *gui.SvgParsed, scale float32) []gui.TessellatedPath {
	val, ok := p.parsed.Load(parsed)
	if !ok {
		return nil
	}
	vg := val.(*VectorGraphic)
	return vg.getTriangles(scale)
}

func (p *Parser) buildParsed(vg *VectorGraphic, scale float32) *gui.SvgParsed {
	tpaths := vg.getTriangles(scale)
	result := &gui.SvgParsed{
		Paths:     tpaths,
		Texts:     vg.Texts,
		TextPaths: vg.TextPaths,
		DefsPaths: vg.DefsPaths,
		Gradients: vg.Gradients,
		Width:     vg.Width,
		Height:    vg.Height,
	}
	p.parsed.Store(result, vg)
	return result
}
