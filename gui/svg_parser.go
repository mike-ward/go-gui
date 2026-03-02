package gui

// SvgParsed holds the result of parsing an SVG document.
type SvgParsed struct {
	Paths      []TessellatedPath
	Texts      []SvgText
	TextPaths  []SvgTextPath
	DefsPaths  map[string]string // id → raw SVG path d attribute
	Gradients  map[string]SvgGradientDef
	Animations []SvgAnimation
	Width      float32
	Height     float32
}

// SvgParser parses and tessellates SVG documents. Set by the
// backend; nil in tests (SVG views degrade to error placeholders).
type SvgParser interface {
	ParseSvg(data string) (*SvgParsed, error)
	ParseSvgFile(path string) (*SvgParsed, error)
	ParseSvgDimensions(data string) (float32, float32, error)
	Tessellate(parsed *SvgParsed, scale float32) []TessellatedPath
}

// SetSvgParser sets the SVG parser backend.
func (w *Window) SetSvgParser(p SvgParser) {
	w.svgParser = p
}

// SvgParser returns the current SVG parser, or nil.
func (w *Window) SvgParser() SvgParser {
	return w.svgParser
}
