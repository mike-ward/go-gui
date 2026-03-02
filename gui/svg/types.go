package svg

import "github.com/mike-ward/go-gui/gui"

// Security limits for SVG parsing to prevent DoS attacks.
const (
	defaultIconSize  = 24
	maxGroupDepth    = 32
	maxElements      = 100000
	maxPathSegments  = 100000
	maxViewBoxDim    = float32(10000)
	maxAttrLen       = 1048576 // 1MB
	maxCoordinate    = float32(1000000)
	maxAnimations    = 100
	maxFlattenDepth  = 16
)

// Tessellation and stroke constants.
const (
	strokeCrossTolerance  = float32(0.001)
	strokeMiterLimit      = float32(4.0)
	strokeRoundCapSegs    = 8
	curveDegenThreshold   = float32(0.0001)
	closedPathEpsilon     = float32(0.0001)
)

// PathCmd identifies a path segment type.
type PathCmd uint8

const (
	CmdMoveTo  PathCmd = iota
	CmdLineTo
	CmdQuadTo
	CmdCubicTo
	CmdClose
)

// PathSegment is one segment of an SVG path.
type PathSegment struct {
	Cmd    PathCmd
	Points []float32 // moveTo/lineTo: [x,y]; quad: [cx,cy,x,y]; cubic: [6]
}

// VectorPath represents a single filled/stroked path.
type VectorPath struct {
	Segments         []PathSegment
	FillColor        gui.SvgColor
	StrokeColor      gui.SvgColor
	StrokeWidth      float32
	StrokeCap        gui.StrokeCap
	StrokeJoin       gui.StrokeJoin
	Transform        [6]float32 // affine [a,b,c,d,e,f]
	ClipPathID       string
	FillGradientID   string
	StrokeGradientID string
	FilterID         string
	StrokeDasharray  []float32
	Opacity          float32
	FillOpacity      float32
	StrokeOpacity    float32
	GroupID          string
}

// svgFilteredGroup holds paths/texts belonging to a filtered <g>.
type svgFilteredGroup struct {
	FilterID  string
	Paths     []VectorPath
	Texts     []gui.SvgText
	TextPaths []gui.SvgTextPath
}

// VectorGraphic holds the full parsed SVG before tessellation.
type VectorGraphic struct {
	Width, Height      float32
	ViewBoxX, ViewBoxY float32
	Paths              []VectorPath
	Texts              []gui.SvgText
	TextPaths          []gui.SvgTextPath
	DefsPaths          map[string]string
	ClipPaths          map[string][]VectorPath
	Gradients          map[string]gui.SvgGradientDef
	Filters            map[string]gui.SvgFilter
	FilteredGroups     []svgFilteredGroup
	Animations         []gui.SvgAnimation
}

// identityTransform is the identity affine matrix.
var identityTransform = [6]float32{1, 0, 0, 1, 0, 0}

// Sentinel colors for style inheritance.
var (
	colorInherit     = gui.SvgColor{R: 255, G: 0, B: 255, A: 1}
	colorTransparent = gui.SvgColor{}
	colorBlack       = gui.SvgColor{R: 0, G: 0, B: 0, A: 255}
)

// groupStyle holds inherited style properties for groups.
type groupStyle struct {
	Transform    [6]float32
	Fill         string
	Stroke       string
	StrokeWidth  string
	StrokeCap    string
	StrokeJoin   string
	ClipPathID   string
	FilterID     string
	FontFamily   string
	FontSize     string
	FontWeight   string
	FontStyle    string
	TextAnchor   string
	Opacity      float32
	FillOpacity  float32
	StrokeOpacity float32
	GroupID      string
}

// defaultGroupStyle returns the root group style.
func defaultGroupStyle(transform [6]float32) groupStyle {
	return groupStyle{
		Transform:     transform,
		Opacity:       1.0,
		FillOpacity:   1.0,
		StrokeOpacity: 1.0,
	}
}

// parseState tracks mutable state during SVG parsing.
type parseState struct {
	elemCount  int
	texts      []gui.SvgText
	textPaths  []gui.SvgTextPath
	animations []gui.SvgAnimation
}

// elementStyle holds common style properties extracted from an SVG element.
type elementStyle struct {
	Transform        [6]float32
	StrokeColor      gui.SvgColor
	StrokeWidth      float32
	StrokeCap        gui.StrokeCap
	StrokeJoin       gui.StrokeJoin
	Opacity          float32
	FillOpacity      float32
	StrokeOpacity    float32
	StrokeGradientID string
	StrokeDasharray  []float32
}
