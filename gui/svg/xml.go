package svg

import (
	"fmt"
	"os"
	"strings"

	"github.com/mike-ward/go-gui/gui"
)

// parseSvg parses an SVG string and returns a VectorGraphic.
func parseSvg(content string) (*VectorGraphic, error) {
	root, err := decodeSvgTree(content)
	if err != nil {
		return nil, err
	}

	vg := &VectorGraphic{
		Width:  defaultIconSize,
		Height: defaultIconSize,
	}

	// viewBox on root.
	if vb, ok := root.AttrMap["viewBox"]; ok {
		nums := parseNumberList(vb)
		if len(nums) >= 4 {
			vg.ViewBoxX = nums[0]
			vg.ViewBoxY = nums[1]
			vg.Width = clampViewBoxDim(nums[2])
			vg.Height = clampViewBoxDim(nums[3])
		}
	} else {
		if w, ok := root.AttrMap["width"]; ok {
			vg.Width = clampViewBoxDim(parseLength(w))
		}
		if h, ok := root.AttrMap["height"]; ok {
			vg.Height = clampViewBoxDim(parseLength(h))
		}
	}

	// Pre-pass: extract <defs>.
	vg.ClipPaths = parseDefsClipPaths(root)
	vg.Gradients = parseDefsGradients(root)
	vg.Filters = parseDefsFilters(root)
	vg.DefsPaths = parseDefsPaths(root)

	// viewBox offset is applied at render time; triangles, animation
	// centers, and motion paths all stay in raw viewBox space.
	defStyle := defaultGroupStyle(identityTransform)
	// Merge presentation attributes from the root <svg> tag (fill,
	// stroke, stroke-width, stroke-linecap, stroke-linejoin, …) so
	// shapes that inherit e.g. fill="currentColor" pick it up.
	defStyle = mergeGroupStyle(root.OpenTag, defStyle)

	state := &parseState{defsPaths: vg.DefsPaths}
	allPaths := parseSvgContent(root, defStyle, 0, state)

	// Separate filtered paths from main paths.
	if len(vg.Filters) > 0 {
		filtered := map[string][]VectorPath{}
		filteredTexts := map[string][]gui.SvgText{}
		filteredTextPaths := map[string][]gui.SvgTextPath{}

		for _, p := range allPaths {
			if p.FilterID != "" {
				if _, ok := vg.Filters[p.FilterID]; ok {
					filtered[p.FilterID] = append(filtered[p.FilterID], p)
					continue
				}
			}
			vg.Paths = append(vg.Paths, p)
		}
		for _, t := range state.texts {
			if t.FilterID != "" {
				if _, ok := vg.Filters[t.FilterID]; ok {
					filteredTexts[t.FilterID] = append(filteredTexts[t.FilterID], t)
					continue
				}
			}
			vg.Texts = append(vg.Texts, t)
		}
		for _, tp := range state.textPaths {
			if tp.FilterID != "" {
				if _, ok := vg.Filters[tp.FilterID]; ok {
					filteredTextPaths[tp.FilterID] = append(filteredTextPaths[tp.FilterID], tp)
					continue
				}
			}
			vg.TextPaths = append(vg.TextPaths, tp)
		}
		for fid, fpaths := range filtered {
			vg.FilteredGroups = append(vg.FilteredGroups, svgFilteredGroup{
				FilterID:  fid,
				Paths:     fpaths,
				Texts:     filteredTexts[fid],
				TextPaths: filteredTextPaths[fid],
			})
		}
	} else {
		vg.Paths = allPaths
		vg.Texts = state.texts
		vg.TextPaths = state.textPaths
	}

	vg.Animations = state.animations
	resolveBegins(vg.Animations, state.animBeginSpecs, state.animIDIndex)
	return vg, nil
}

const maxSvgFileSize = 4 << 20 // 4 MB

// parseSvgFile loads and parses an SVG file.
func parseSvgFile(path string) (*VectorGraphic, error) {
	data, err := loadSvgFile(path)
	if err != nil {
		return nil, err
	}
	return parseSvg(string(data))
}

func loadSvgFile(path string) ([]byte, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("read SVG file: %w", err)
	}
	if info.Size() > maxSvgFileSize {
		return nil, fmt.Errorf("SVG file too large: %d bytes", info.Size())
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read SVG file: %w", err)
	}
	return data, nil
}

// parseSvgDimensions extracts only width/height without a full
// parse. Operates on the raw string so callers can probe dimensions
// on incomplete or fragment-only SVG content (no closing tag
// required).
func parseSvgDimensions(content string) (float32, float32) {
	openTag := extractRootSVGOpenTag(content)
	if openTag == "" {
		openTag = content
	}
	if vb, ok := findAttr(openTag, "viewBox"); ok {
		nums := parseNumberList(vb)
		if len(nums) >= 4 {
			return clampViewBoxDim(nums[2]), clampViewBoxDim(nums[3])
		}
	}
	w := float32(defaultIconSize)
	h := float32(defaultIconSize)
	if ws, ok := findAttr(openTag, "width"); ok {
		w = clampViewBoxDim(parseLength(ws))
	}
	if hs, ok := findAttr(openTag, "height"); ok {
		h = clampViewBoxDim(parseLength(hs))
	}
	return w, h
}

func extractRootSVGOpenTag(content string) string {
	start := strings.Index(content, "<svg")
	if start < 0 {
		return ""
	}
	nameEnd := start + len("<svg")
	if nameEnd < len(content) {
		switch content[nameEnd] {
		case '>', '/', ' ', '\t', '\n', '\r':
		default:
			return ""
		}
	}
	inQuote := byte(0)
	for i := nameEnd; i < len(content); i++ {
		switch c := content[i]; c {
		case '"', '\'':
			switch inQuote {
			case 0:
				inQuote = c
			case c:
				inQuote = 0
			}
		case '>':
			if inQuote == 0 {
				return content[start : i+1]
			}
		}
	}
	return content[start:]
}

// parseSvgContent walks n's children, emitting VectorPaths for
// shape/group elements and pushing animations onto state. Recurses
// into <g>/<a> groups with merged styles; defs children are skipped
// (defs pre-pass already ran). Returns the accumulated path list.
//
//nolint:gocyclo // SVG element switch
func parseSvgContent(n *xmlNode, inherited groupStyle, depth int,
	state *parseState) []VectorPath {
	var paths []VectorPath
	if depth > maxGroupDepth {
		return paths
	}
	for i := range n.Children {
		if state.elemCount >= maxElements {
			break
		}
		c := &n.Children[i]
		switch c.Name {
		case "defs":
			// Already handled by defs pre-pass; skip.
			continue

		case "g", "a":
			gs := mergeGroupStyle(c.OpenTag, inherited)
			state.elemCount++
			// Synthesize a GroupID when the group has no id but
			// contains inline animations, so descendants can bind.
			if gs.GroupID == "" && nodeHasInlineAnimation(c) {
				state.synthID++
				gs.GroupID = fmt.Sprintf("__anim_%d", state.synthID)
			}
			paths = append(paths, parseSvgContent(c, gs, depth+1, state)...)

		case "path":
			appendShape(c, inherited, state, &paths,
				func(gs groupStyle) (VectorPath, bool) {
					return parsePathWithStyle(c.OpenTag, gs)
				})

		case "rect":
			appendShape(c, inherited, state, &paths,
				func(gs groupStyle) (VectorPath, bool) {
					return parseRectWithStyle(c.OpenTag, gs)
				})

		case "circle":
			appendShape(c, inherited, state, &paths,
				func(gs groupStyle) (VectorPath, bool) {
					return parseCircleWithStyle(c.OpenTag, gs)
				})

		case "ellipse":
			appendShape(c, inherited, state, &paths,
				func(gs groupStyle) (VectorPath, bool) {
					return parseEllipseWithStyle(c.OpenTag, gs)
				})

		case "polygon":
			appendShape(c, inherited, state, &paths,
				func(gs groupStyle) (VectorPath, bool) {
					return parsePolygonWithStyle(c.OpenTag, gs, true)
				})

		case "polyline":
			appendShape(c, inherited, state, &paths,
				func(gs groupStyle) (VectorPath, bool) {
					return parsePolygonWithStyle(c.OpenTag, gs, false)
				})

		case "line":
			appendShape(c, inherited, state, &paths,
				func(gs groupStyle) (VectorPath, bool) {
					return parseLineWithStyle(c.OpenTag, gs)
				})

		case "text":
			state.elemCount++
			parseTextElement(c, inherited, state)

		case "animate":
			state.elemCount++
			if len(state.animations) < maxAnimations {
				if a, ok := parseAnimateElement(c.OpenTag, inherited); ok {
					state.animations = append(state.animations, a)
					registerAnimation(state, c.OpenTag,
						len(state.animations)-1)
				}
			}

		case "animateTransform":
			state.elemCount++
			if len(state.animations) < maxAnimations {
				if a, ok := parseAnimateTransformElement(
					c.OpenTag, inherited); ok {
					state.animations = append(state.animations, a)
					registerAnimation(state, c.OpenTag,
						len(state.animations)-1)
				}
			}

		case "set":
			state.elemCount++
			if len(state.animations) < maxAnimations {
				if a, ok := parseSetElement(c.OpenTag, inherited); ok {
					state.animations = append(state.animations, a)
					registerAnimation(state, c.OpenTag,
						len(state.animations)-1)
				}
			}

		case "animateMotion":
			state.elemCount++
			if len(state.animations) < maxAnimations {
				if a, ok := parseAnimateMotionElement(
					c, inherited, state); ok {
					state.animations = append(state.animations, a)
					registerAnimation(state, c.OpenTag,
						len(state.animations)-1)
				}
			}

		default:
			// Unknown element: ignore. (Descendants also ignored —
			// would need explicit handling.)
		}
	}
	return paths
}

// appendShape wraps parseShapeElement's original bookkeeping: the
// shape parser runs with an optionally-synthesized GroupID if the
// node carries inline animation children, then inline anims are
// folded onto the path's state.
func appendShape(
	c *xmlNode,
	inherited groupStyle,
	state *parseState,
	paths *[]VectorPath,
	parser func(gs groupStyle) (VectorPath, bool),
) {
	state.elemCount++

	shapeGS := inherited
	if nodeHasInlineAnimation(c) {
		gid := c.AttrMap["id"]
		if gid == "" {
			state.synthID++
			gid = fmt.Sprintf("__anim_%d", state.synthID)
		}
		shapeGS.GroupID = gid
		all, fill, stroke := scanOpacityAnimTargets(c)
		shapeGS.SkipOpacity = all
		shapeGS.SkipFillOpacity = fill
		shapeGS.SkipStrokeOpacity = stroke
	}

	pathIdx := -1
	if p, ok := parser(shapeGS); ok {
		if shapeGS.GroupID != inherited.GroupID {
			p.GroupID = shapeGS.GroupID
		}
		state.pathIDSeq++
		p.PathID = state.pathIDSeq
		*paths = append(*paths, p)
		pathIdx = len(*paths) - 1
	}

	animStart := len(state.animations)
	parseShapeInlineChildren(c, shapeGS, state)
	// Clip-pathed shapes skip re-tessellation.
	if pathIdx >= 0 && (*paths)[pathIdx].ClipPathID == "" {
		for i := animStart; i < len(state.animations); i++ {
			k := state.animations[i].Kind
			if k == gui.SvgAnimAttr ||
				k == gui.SvgAnimDashArray ||
				k == gui.SvgAnimDashOffset {
				(*paths)[pathIdx].Animated = true
				break
			}
		}
	}
}

// parseAnimateForDispatch picks the right <animate> parser based on
// attributeName: opacity/fill-opacity/stroke-opacity → opacity;
// stroke-dasharray → dash array; stroke-dashoffset → dash offset;
// primitive attrs (cx/cy/r/...) → attribute animation. Unknown names
// reject (ok=false).
func parseAnimateForDispatch(
	elem string, inherited groupStyle,
) (gui.SvgAnimation, bool) {
	attr, ok := findAttr(elem, "attributeName")
	if !ok {
		return gui.SvgAnimation{}, false
	}
	switch attr {
	case "opacity", "fill-opacity", "stroke-opacity":
		return parseAnimateElement(elem, inherited)
	case "stroke-dasharray":
		return parseAnimateDashArrayElement(elem, inherited)
	case "stroke-dashoffset":
		return parseAnimateDashOffsetElement(elem, inherited)
	}
	return parseAnimateAttributeElement(elem, inherited)
}

// nodeHasInlineAnimation reports whether any direct child of n is a
// SMIL animation element.
func nodeHasInlineAnimation(n *xmlNode) bool {
	for i := range n.Children {
		switch n.Children[i].Name {
		case "animate", "animateTransform", "animateMotion", "set":
			return true
		}
	}
	return false
}

// scanOpacityAnimTargets reports which opacity sub-attributes are
// animated by inline <animate> children of a shape. Used to suppress
// static opacity baking for channels the animation will overwrite at
// render time.
func scanOpacityAnimTargets(n *xmlNode) (all, fill, stroke bool) {
	for i := range n.Children {
		c := &n.Children[i]
		if c.Name != "animate" {
			continue
		}
		switch c.AttrMap["attributeName"] {
		case "opacity":
			all = true
		case "fill-opacity":
			fill = true
		case "stroke-opacity":
			stroke = true
		}
	}
	return
}

// parseShapeInlineChildren walks a shape's children for
// <animate>/<animateTransform>/<animateMotion>/<set> and appends
// them to state.animations keyed by shapeGS.GroupID.
func parseShapeInlineChildren(
	n *xmlNode, shapeGS groupStyle, state *parseState,
) {
	for i := range n.Children {
		c := &n.Children[i]
		switch c.Name {
		case "animate":
			if len(state.animations) < maxAnimations {
				if a, ok := parseAnimateForDispatch(c.OpenTag, shapeGS); ok {
					state.animations = append(state.animations, a)
					registerAnimation(state, c.OpenTag,
						len(state.animations)-1)
				}
			}
		case "animateTransform":
			if len(state.animations) < maxAnimations {
				if a, ok := parseAnimateTransformElement(
					c.OpenTag, shapeGS); ok {
					state.animations = append(state.animations, a)
					registerAnimation(state, c.OpenTag,
						len(state.animations)-1)
				}
			}
		case "set":
			if len(state.animations) < maxAnimations {
				if a, ok := parseSetElement(c.OpenTag, shapeGS); ok {
					state.animations = append(state.animations, a)
					registerAnimation(state, c.OpenTag,
						len(state.animations)-1)
				}
			}
		case "animateMotion":
			if len(state.animations) < maxAnimations {
				if a, ok := parseAnimateMotionElement(
					c, shapeGS, state); ok {
					state.animations = append(state.animations, a)
					registerAnimation(state, c.OpenTag,
						len(state.animations)-1)
				}
			}
		}
	}
}
