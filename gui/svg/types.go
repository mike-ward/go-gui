package svg

import "github.com/mike-ward/go-gui/gui"

// Security limits for SVG parsing to prevent DoS attacks.
const (
	defaultIconSize = 24
	maxGroupDepth   = 32
	maxElements     = 100000
	maxPathSegments = 100000
	maxViewBoxDim   = float32(10000)
	maxAttrLen      = 1048576 // 1MB
	maxCoordinate   = float32(1000000)
	maxAnimations   = 100
	maxFlattenDepth = 16
	// maxKeyframes caps the number of keyframes (values) and
	// syncbase begin entries parsed from a single <animate> /
	// <animateTransform>. Real assets use <20; this guards
	// against a pathological input that lists thousands of
	// semicolon-separated entries.
	maxKeyframes = 256
)

// Tessellation and stroke constants.
const (
	strokeCrossTolerance = float32(0.001)
	strokeMiterLimit     = float32(4.0)
	strokeRoundCapSegs   = 8
	curveDegenThreshold  = float32(0.0001)
	closedPathEpsilon    = float32(0.0001)
	maxSplitTriDepth     = 8
)

// PathCmd identifies a path segment type.
type PathCmd uint8

// PathCmd constants.
const (
	CmdMoveTo PathCmd = iota
	CmdLineTo
	CmdQuadTo
	CmdCubicTo
	CmdClose
)

// FillRule selects how overlapping subpaths are filled. Matches
// the SVG fill-rule presentation attribute.
type FillRule uint8

// FillRule constants. Zero value is nonzero (SVG default).
const (
	FillRuleNonzero FillRule = iota
	FillRuleEvenOdd
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
	StrokeDashOffset float32
	Opacity          float32
	FillOpacity      float32
	StrokeOpacity    float32
	GroupID          string
	Animated         bool
	Primitive        gui.SvgPrimitive
	FillRule         FillRule
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
// colorInherit marks a fill/stroke that was not explicitly set —
// resolve from the parent chain during applyInheritedStyle.
// colorCurrent marks an explicit "currentColor" or "inherit"
// keyword — preserve it so the render-time tint (SvgCfg.Color)
// can substitute the actual color. Both are recognized by their
// magenta RGB so opacity baking can still scale alpha safely.
var (
	colorInherit     = gui.SvgColor{R: 255, G: 0, B: 255, A: 1}
	colorCurrent     = gui.SvgColor{R: 255, G: 0, B: 255, A: 2}
	colorTransparent = gui.SvgColor{}
	colorBlack       = gui.SvgColor{R: 0, G: 0, B: 0, A: 255}
)

// groupStyle holds inherited style properties for groups.
type groupStyle struct {
	Transform     [6]float32
	Fill          string
	Stroke        string
	StrokeWidth   string
	StrokeCap     string
	StrokeJoin    string
	ClipPathID    string
	FilterID      string
	FontFamily    string
	FontSize      string
	FontWeight    string
	FontStyle     string
	TextAnchor    string
	Opacity       float32
	FillOpacity   float32
	StrokeOpacity float32
	GroupID       string
	// SkipOpacity / SkipFillOpacity / SkipStrokeOpacity tell the
	// shape-style baker to leave the corresponding alpha channel
	// at full strength because an inline opacity animation will
	// supply the value at render time. Without this, a static
	// fill-opacity="0" bakes the fill alpha to zero and tessellate
	// drops the geometry, leaving the animation nothing to scale.
	SkipOpacity       bool
	SkipFillOpacity   bool
	SkipStrokeOpacity bool
	// FillRule is the resolved fill-rule inherited from ancestor
	// groups. Zero value (FillRuleNonzero) matches the SVG default
	// when no ancestor sets the attribute.
	FillRule FillRule
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
	synthID    int
	texts      []gui.SvgText
	textPaths  []gui.SvgTextPath
	animations []gui.SvgAnimation
	// animIDIndex maps an <animate id="..."> to its index in
	// animations. Populated only when an animation carries an id.
	animIDIndex map[string]int
	// animBeginSpecs holds unresolved begin specs per animation
	// index. Populated only when begin= references another
	// animation via syncbase (.begin/.end).
	animBeginSpecs map[int][]beginSpec
	// defsPaths maps defs <path id="..."> to its d-attribute. Used
	// by animateMotion's <mpath xlink:href="#id"/>.
	defsPaths map[string]string
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
