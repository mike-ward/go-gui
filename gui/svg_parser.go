package gui

// SvgParsed holds the result of parsing an SVG document.
type SvgParsed struct {
	Paths          []TessellatedPath
	Texts          []SvgText
	TextPaths      []SvgTextPath
	DefsPaths      map[string]string // id → raw SVG path d attribute
	Gradients      map[string]SvgGradientDef
	FilteredGroups []SvgParsedFilteredGroup
	Animations     []SvgAnimation
	Width          float32
	Height         float32
}

// SvgParser parses and tessellates SVG documents. Set by the
// backend; nil in tests (SVG views degrade to error placeholders).
type SvgParser interface {
	ParseSvg(data string) (*SvgParsed, error)
	ParseSvgFile(path string) (*SvgParsed, error)
	ParseSvgDimensions(data string) (float32, float32, error)
	Tessellate(parsed *SvgParsed, scale float32) []TessellatedPath
}

// SvgAnimAttrMask flags which fields of SvgAnimAttrOverride are set.
// Flat bits + flat float32 fields avoid per-frame heap allocations
// that *float32 fields would incur through the map value.
type SvgAnimAttrMask uint16

// SvgAnimAttrMask bits.
const (
	SvgAnimMaskCX SvgAnimAttrMask = 1 << iota
	SvgAnimMaskCY
	SvgAnimMaskR
	SvgAnimMaskRX
	SvgAnimMaskRY
	SvgAnimMaskX
	SvgAnimMaskY
	SvgAnimMaskWidth
	SvgAnimMaskHeight
)

// SvgAnimAttrOverride carries current animated attribute values to
// apply when re-tessellating a primitive. Fields are valid only when
// their Mask bit is set; unset fields retain the parsed static value.
// When AdditiveMask bit is also set, the override value is a delta
// to add to the parsed primitive value rather than a replacement —
// matches SMIL additive="sum" / by= shorthand semantics.
type SvgAnimAttrOverride struct {
	Mask         SvgAnimAttrMask
	AdditiveMask SvgAnimAttrMask
	CX, CY       float32
	R            float32
	RX, RY       float32
	X, Y         float32
	Width        float32
	Height       float32
}

// AnimatedSvgParser is an optional extension of SvgParser for backends
// that can re-tessellate animated primitive shapes with current
// attribute values. Detected via type assertion; parsers that do not
// implement it simply render animated paths from their static cache.
type AnimatedSvgParser interface {
	// TessellateAnimated returns fresh triangles for every path in
	// parsed whose Animated flag is set, at the given scale, with
	// optional attribute overrides keyed by GroupID. Returned slice
	// order matches the Animated-flagged paths' document order.
	// reuse, if non-nil, is a caller-supplied slice the parser may
	// append into to amortize per-frame slice allocations; returned
	// slice may alias reuse's backing array. An empty overrides map
	// or nil map yields nil (caller should use cached triangles).
	TessellateAnimated(parsed *SvgParsed, scale float32,
		overrides map[string]SvgAnimAttrOverride,
		reuse []TessellatedPath) []TessellatedPath
}

// SetSvgParser sets the SVG parser backend.
func (w *Window) SetSvgParser(p SvgParser) {
	w.svgParser = p
}

// SvgParser returns the current SVG parser, or nil.
func (w *Window) SvgParser() SvgParser {
	return w.svgParser
}
