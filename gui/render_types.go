package gui

import "github.com/mike-ward/go-glyph"

// RenderKind identifies the type of drawing command stored in a
// RenderCmd. The renderer pipeline emits a flat []RenderCmd slice;
// the dispatch loop switches on Kind to draw each command.
type RenderKind uint8

const (
	RenderNone RenderKind = iota
	RenderClip
	RenderRect
	RenderStrokeRect
	RenderCircle
	RenderImage
	RenderText
	RenderLine
	RenderShadow
	RenderBlur
	RenderGradient
	RenderGradientBorder
	RenderSvg
	RenderLayout
	RenderLayoutTransformed
	RenderLayoutPlaced
	RenderFilterBegin
	RenderFilterEnd
	RenderFilterComposite
	RenderCustomShader
	RenderTextPath
	RenderRTF
)

// RenderCmd is a flat discriminated struct holding all draw
// command variants. Kind selects which fields are meaningful.
// Stored in a pre-allocated slice reused via renderers[:0] each
// frame to minimize heap allocations.
type RenderCmd struct {
	Kind RenderKind

	// Position/size — used by most kinds.
	X, Y float32
	W, H float32

	// Visual properties.
	Color  Color
	Radius float32

	// Type-specific numerics.
	Thickness  float32 // StrokeRect, GradientBorder
	BlurRadius float32 // Shadow, Blur
	Scale      float32 // Svg, FilterBegin
	OffsetX    float32 // Shadow; Line X1
	OffsetY    float32 // Shadow; Line Y1
	ClipRadius float32 // Image

	// Flags.
	Fill       bool // Rect fill, Circle fill
	IsClipMask bool // Svg stencil mask
	ClipGroup  int  // Svg clip group id
	Layers     int  // FilterComposite
	GroupIdx   int  // FilterBegin

	// String data.
	Text       string  // Text
	FontName   string  // Text font family
	FontSize   float32 // Text font size (points)
	FontAscent float32 // Text font ascent (pixels)
	TextWidth  float32 // Text source width (pixels)
	Resource   string  // Image file path

	// SVG animation rotation (degrees, center in SVG space).
	RotAngle float32
	RotCX    float32
	RotCY    float32

	// Slice data (Svg).
	Triangles    []float32
	VertexColors []Color
	// Optional multiplier for SVG vertex alpha (0..1) to avoid
	// per-frame vertex color copies when animating opacity.
	VertexAlphaScale float32
	HasVertexAlpha   bool

	// Pointer fields.
	Shader       *Shader
	Gradient     *GradientDef
	TextStylePtr *TextStyle            // full text style (typeface, etc.)
	TextGradient *glyph.GradientConfig // SVG gradient text
	TextPath     *TextPathData         // SVG textPath placement data
	LayoutPtr    *glyph.Layout         // RTF pre-shaped layout
}

// TextPathData holds pre-computed path data for RenderTextPath.
type TextPathData struct {
	Polyline []float32 // flattened path [x0,y0, x1,y1, ...]
	Table    []float32 // cumulative arc-length table
	TotalLen float32
	Offset   float32 // resolved start offset (screen coords)
	Anchor   int     // 0=start, 1=middle, 2=end
	Method   int     // 0=align, 1=stretch
}

// renderCmdKindName returns a debug name for the given RenderKind.
func renderCmdKindName(k RenderKind) string {
	switch k {
	case RenderNone:
		return "RenderNone"
	case RenderClip:
		return "RenderClip"
	case RenderRect:
		return "RenderRect"
	case RenderStrokeRect:
		return "RenderStrokeRect"
	case RenderCircle:
		return "RenderCircle"
	case RenderImage:
		return "RenderImage"
	case RenderText:
		return "RenderText"
	case RenderLine:
		return "RenderLine"
	case RenderShadow:
		return "RenderShadow"
	case RenderBlur:
		return "RenderBlur"
	case RenderGradient:
		return "RenderGradient"
	case RenderGradientBorder:
		return "RenderGradientBorder"
	case RenderSvg:
		return "RenderSvg"
	case RenderLayout:
		return "RenderLayout"
	case RenderLayoutTransformed:
		return "RenderLayoutTransformed"
	case RenderLayoutPlaced:
		return "RenderLayoutPlaced"
	case RenderFilterBegin:
		return "RenderFilterBegin"
	case RenderFilterEnd:
		return "RenderFilterEnd"
	case RenderFilterComposite:
		return "RenderFilterComposite"
	case RenderCustomShader:
		return "RenderCustomShader"
	case RenderTextPath:
		return "RenderTextPath"
	case RenderRTF:
		return "RenderRTF"
	default:
		return "Unknown"
	}
}

// FilterBracketRange describes a matched DrawFilterBegin..DrawFilterEnd
// range within the renderers slice.
type FilterBracketRange struct {
	StartIdx int
	EndIdx   int
	NextIdx  int
	FoundEnd bool
}

// findFilterBracketRange scans renderers from startIdx looking for
// a DrawFilterBegin..DrawFilterEnd pair.
func findFilterBracketRange(renderers []RenderCmd, startIdx int) FilterBracketRange {
	for i := startIdx; i < len(renderers); i++ {
		if renderers[i].Kind == RenderFilterEnd {
			return FilterBracketRange{
				StartIdx: startIdx,
				EndIdx:   i,
				NextIdx:  i + 1,
				FoundEnd: true,
			}
		}
	}
	return FilterBracketRange{
		StartIdx: startIdx,
		EndIdx:   len(renderers),
		NextIdx:  len(renderers),
		FoundEnd: false,
	}
}
