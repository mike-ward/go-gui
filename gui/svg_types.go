package gui

// SvgColor represents an RGBA color from SVG parsing.
type SvgColor struct {
	R, G, B, A uint8
}

// TessellatedPath holds triangulated SVG path geometry.
type TessellatedPath struct {
	Triangles    []float32
	Color        SvgColor
	VertexColors []SvgColor
	IsClipMask   bool
	ClipGroup    int
	GroupID      string
	// Animated marks the path as a re-tessellation target. Set when
	// an inline <animate> with an animatable attribute (cx, cy, r,
	// x, y, width, height, rx, ry) targets this shape.
	Animated bool
	// Primitive carries raw attributes for re-tessellation; zero
	// value when the source is not a primitive.
	Primitive SvgPrimitive
}

// SvgText holds a parsed SVG text element.
type SvgText struct {
	Text           string
	FontFamily     string
	FontSize       float32
	X, Y           float32
	Opacity        float32
	IsBold         bool
	IsItalic       bool
	FontWeight     int // CSS numeric weight (100-900); 0 = default (400)
	Color          SvgColor
	StrokeColor    SvgColor
	StrokeWidth    float32
	FillGradientID string
	FilterID       string
	Anchor         int // 0=start, 1=middle, 2=end
	Underline      bool
	Strikethrough  bool
	LetterSpacing  float32
}

// SvgTextPath holds a parsed SVG textPath element.
type SvgTextPath struct {
	Text          string
	PathID        string
	FontFamily    string
	FontSize      float32
	Opacity       float32
	StartOffset   float32
	LetterSpacing float32
	IsBold        bool
	IsItalic      bool
	FontWeight    int // CSS numeric weight (100-900); 0 = default (400)
	Color         SvgColor
	StrokeColor   SvgColor
	StrokeWidth   float32
	FilterID      string
	Anchor        int // 0=start, 1=middle, 2=end
	Method        int // 0=align, 1=stretch
	Side          int // 0=left, 1=right
	IsPercent     bool
}

// SvgFilter holds a parsed SVG filter definition.
type SvgFilter struct {
	ID         string
	StdDev     float32
	BlurLayers int
	KeepSource bool
}

// SvgParsedFilteredGroup holds parsed+tessellated geometry for a
// filter group (paths that share a common filter="url(#id)").
type SvgParsedFilteredGroup struct {
	Filter    SvgFilter
	Paths     []TessellatedPath
	Texts     []SvgText
	TextPaths []SvgTextPath
}

// SvgGradientStop defines a color stop in an SVG gradient.
type SvgGradientStop struct {
	Offset float32
	Color  SvgColor
}

// SvgGradientDef defines an SVG gradient (linear or radial).
type SvgGradientDef struct {
	Stops         []SvgGradientStop
	X1, Y1        float32
	X2, Y2        float32
	CX, CY, R     float32
	FX, FY        float32
	IsRadial      bool
	GradientUnits string
}

// SvgAnimKind identifies the type of SMIL animation.
type SvgAnimKind uint8

// SvgAnimKind constants.
const (
	SvgAnimOpacity SvgAnimKind = iota
	SvgAnimRotate
	// SvgAnimAttr is a generic per-attribute animation (cx, cy, r,
	// x, y, width, height, rx, ry). Phase-1 records the animation;
	// phase-2 evaluates overrides and re-tessellates.
	SvgAnimAttr
)

// SvgAttrName identifies an animatable primitive attribute.
type SvgAttrName uint8

// SvgAttrName constants.
const (
	SvgAttrNone SvgAttrName = iota
	SvgAttrCX
	SvgAttrCY
	SvgAttrR
	SvgAttrX
	SvgAttrY
	SvgAttrWidth
	SvgAttrHeight
	SvgAttrRX
	SvgAttrRY
)

// SvgAnimation holds parsed SMIL animation data.
type SvgAnimation struct {
	Kind     SvgAnimKind
	GroupID  string
	Values   []float32 // opacity: keyframes; rotate: [from, to]
	CenterX  float32   // rotation center (SVG coords)
	CenterY  float32
	DurSec   float32
	BeginSec float32
	AttrName SvgAttrName // valid when Kind == SvgAnimAttr
}

// SvgPrimitiveKind identifies the source primitive of a VectorPath.
type SvgPrimitiveKind uint8

// SvgPrimitiveKind constants.
const (
	SvgPrimNone SvgPrimitiveKind = iota
	SvgPrimCircle
	SvgPrimEllipse
	SvgPrimRect
	SvgPrimLine
)

// SvgPrimitive carries raw primitive attributes so animated paths can
// be re-tessellated each frame with current attribute values. Kind is
// SvgPrimNone for non-primitive paths (<path>, polygons, etc.). The
// composed transform lives on the owning VectorPath / GroupID; only
// the primitive-local attributes are captured here.
type SvgPrimitive struct {
	Kind   SvgPrimitiveKind
	CX, CY float32 // circle / ellipse center
	R      float32 // circle radius
	RX, RY float32 // ellipse / rect corner radii
	X, Y   float32 // rect upper-left, line endpoint
	W, H   float32 // rect size
	X2, Y2 float32 // line far endpoint
}

// StrokeCap defines SVG stroke line cap styles.
type StrokeCap uint8

// StrokeCap constants.
const (
	ButtCap StrokeCap = iota
	RoundCap
	SquareCap
)

// StrokeJoin defines SVG stroke line join styles.
type StrokeJoin uint8

// StrokeJoin constants.
const (
	MiterJoin StrokeJoin = iota
	RoundJoin
	BevelJoin
)

// svgToColor converts an SvgColor to a gui Color.
func svgToColor(c SvgColor) Color {
	return Color{c.R, c.G, c.B, c.A, true}
}
