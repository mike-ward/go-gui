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
	CX, CY, R    float32
	FX, FY        float32
	IsRadial      bool
	GradientUnits string
}

// SvgAnimKind identifies the type of SMIL animation.
type SvgAnimKind uint8

const (
	SvgAnimOpacity SvgAnimKind = iota
	SvgAnimRotate
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
}

// StrokeCap defines SVG stroke line cap styles.
type StrokeCap uint8

const (
	ButtCap StrokeCap = iota
	RoundCap
	SquareCap
)

// StrokeJoin defines SVG stroke line join styles.
type StrokeJoin uint8

const (
	MiterJoin StrokeJoin = iota
	RoundJoin
	BevelJoin
)

// svgToColor converts an SvgColor to a gui Color.
func svgToColor(c SvgColor) Color {
	return Color{c.R, c.G, c.B, c.A}
}
