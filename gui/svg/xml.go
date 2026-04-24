package svg

import (
	"fmt"
	"os"
	"strings"

	"github.com/mike-ward/go-gui/gui"
)

// parseSvg parses an SVG string and returns a VectorGraphic.
func parseSvg(content string) (*VectorGraphic, error) {
	vg := &VectorGraphic{
		Width:  defaultIconSize,
		Height: defaultIconSize,
	}

	// Parse viewBox
	if vb, ok := findAttr(content, "viewBox"); ok {
		nums := parseNumberList(vb)
		if len(nums) >= 4 {
			vg.ViewBoxX = nums[0]
			vg.ViewBoxY = nums[1]
			vg.Width = clampViewBoxDim(nums[2])
			vg.Height = clampViewBoxDim(nums[3])
		}
	} else {
		if w, ok := findAttr(content, "width"); ok {
			vg.Width = clampViewBoxDim(parseLength(w))
		}
		if h, ok := findAttr(content, "height"); ok {
			vg.Height = clampViewBoxDim(parseLength(h))
		}
	}

	// Pre-pass: extract <defs>
	vg.ClipPaths = parseDefsClipPaths(content)
	vg.Gradients = parseDefsGradients(content)
	vg.Filters = parseDefsFilters(content)
	vg.DefsPaths = parseDefsPaths(content)

	// viewBox offset is applied at tessellation time by shifting
	// vertex coords (see tessellatePaths). Keeping it out of the
	// inherited path.Transform chain prevents SMIL <animateTransform>
	// replace semantics from clobbering the viewBox shift when
	// decomposed bases feed the per-group animation state.
	defStyle := defaultGroupStyle(identityTransform)
	// Merge presentation attributes from the root <svg> tag (fill,
	// stroke, stroke-width, stroke-linecap, stroke-linejoin, …) so
	// shapes that rely on inheriting e.g. fill="currentColor" from
	// the root pick it up.
	if svgTag, ok := findRootElementTag(content, "svg"); ok {
		defStyle = mergeGroupStyle(svgTag, defStyle)
	}
	state := &parseState{defsPaths: vg.DefsPaths}
	allPaths := parseSvgContent(content, defStyle, 0, state)

	// Separate filtered paths from main paths
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
	return parseSvg(string(data))
}

// parseSvgDimensions extracts only width/height without full parse.
func parseSvgDimensions(content string) (float32, float32) {
	if vb, ok := findAttr(content, "viewBox"); ok {
		nums := parseNumberList(vb)
		if len(nums) >= 4 {
			return clampViewBoxDim(nums[2]), clampViewBoxDim(nums[3])
		}
	}
	w := float32(defaultIconSize)
	h := float32(defaultIconSize)
	if ws, ok := findAttr(content, "width"); ok {
		w = clampViewBoxDim(parseLength(ws))
	}
	if hs, ok := findAttr(content, "height"); ok {
		h = clampViewBoxDim(parseLength(hs))
	}
	return w, h
}

// parseSvgContent parses SVG content recursively, handling groups.
//
//nolint:gocyclo // SVG element switch
func parseSvgContent(content string, inherited groupStyle, depth int, state *parseState) []VectorPath {
	var paths []VectorPath
	pos := 0

	if depth > maxGroupDepth {
		return paths
	}

	for pos < len(content) {
		if state.elemCount >= maxElements {
			break
		}
		// Find next element
		start := strings.Index(content[pos:], "<")
		if start < 0 {
			break
		}
		start += pos

		// Skip comments and declarations
		if start+3 < len(content) {
			if content[start:start+4] == "<!--" {
				end := strings.Index(content[start:], "-->")
				if end < 0 {
					break
				}
				pos = start + end + 3
				continue
			}
			if content[start+1] == '!' || content[start+1] == '?' {
				end := strings.IndexByte(content[start:], '>')
				if end < 0 {
					break
				}
				pos = start + end + 1
				continue
			}
		}

		// Closing tag
		if start+1 < len(content) && content[start+1] == '/' {
			end := strings.IndexByte(content[start:], '>')
			if end < 0 {
				break
			}
			pos = start + end + 1
			continue
		}

		// Tag name
		tagEnd := findTagNameEnd(content, start+1)
		if tagEnd <= start+1 {
			pos = start + 1
			continue
		}
		tagName := content[start+1 : tagEnd]

		// Element end
		elemEndRel := strings.IndexByte(content[start:], '>')
		if elemEndRel < 0 {
			break
		}
		elemEnd := start + elemEndRel
		elem := content[start : elemEnd+1]
		isSelfClosing := elemEnd > 0 && content[elemEnd-1] == '/'

		// Skip <defs> (already parsed)
		if tagName == "defs" {
			if isSelfClosing {
				pos = elemEnd + 1
				continue
			}
			defsEnd := findClosingTag(content, "defs", elemEnd+1)
			closeEnd := strings.IndexByte(content[defsEnd:], '>')
			if closeEnd < 0 {
				break
			}
			pos = defsEnd + closeEnd + 1
			continue
		}

		switch tagName {
		case "g", "a":
			gs := mergeGroupStyle(elem, inherited)
			state.elemCount++
			if isSelfClosing {
				pos = elemEnd + 1
				continue
			}
			groupStart := elemEnd + 1
			groupEnd := findClosingTag(content, tagName, groupStart)
			if groupEnd > groupStart {
				groupContent := content[groupStart:groupEnd]
				// When a group lacks an explicit id but contains
				// inline <animate>/<animateTransform> children,
				// synthesize a GroupID so descendants can bind to
				// the animation. Mirrors parseShapeElement's
				// synth-id logic. Substring check also matches
				// nested animations; inner groups with their own
				// id or synth assignment override.
				if gs.GroupID == "" &&
					shapeHasInlineAnimation(groupContent) {
					state.synthID++
					gs.GroupID = fmt.Sprintf("__anim_%d",
						state.synthID)
				}
				paths = append(paths,
					parseSvgContent(groupContent, gs, depth+1, state)...)
			}
			closeEnd := strings.IndexByte(content[groupEnd:], '>')
			if closeEnd < 0 {
				pos = len(content)
				continue
			}
			pos = groupEnd + closeEnd + 1

		case "path":
			pos = parseShapeElement(content, elem, tagName, elemEnd,
				isSelfClosing, inherited, state, &paths,
				func(gs groupStyle) (VectorPath, bool) {
					return parsePathWithStyle(elem, gs)
				})

		case "rect":
			pos = parseShapeElement(content, elem, tagName, elemEnd,
				isSelfClosing, inherited, state, &paths,
				func(gs groupStyle) (VectorPath, bool) {
					return parseRectWithStyle(elem, gs)
				})

		case "circle":
			pos = parseShapeElement(content, elem, tagName, elemEnd,
				isSelfClosing, inherited, state, &paths,
				func(gs groupStyle) (VectorPath, bool) {
					return parseCircleWithStyle(elem, gs)
				})

		case "ellipse":
			pos = parseShapeElement(content, elem, tagName, elemEnd,
				isSelfClosing, inherited, state, &paths,
				func(gs groupStyle) (VectorPath, bool) {
					return parseEllipseWithStyle(elem, gs)
				})

		case "polygon":
			pos = parseShapeElement(content, elem, tagName, elemEnd,
				isSelfClosing, inherited, state, &paths,
				func(gs groupStyle) (VectorPath, bool) {
					return parsePolygonWithStyle(elem, gs, true)
				})

		case "polyline":
			pos = parseShapeElement(content, elem, tagName, elemEnd,
				isSelfClosing, inherited, state, &paths,
				func(gs groupStyle) (VectorPath, bool) {
					return parsePolygonWithStyle(elem, gs, false)
				})

		case "line":
			pos = parseShapeElement(content, elem, tagName, elemEnd,
				isSelfClosing, inherited, state, &paths,
				func(gs groupStyle) (VectorPath, bool) {
					return parseLineWithStyle(elem, gs)
				})

		case "text":
			state.elemCount++
			if !isSelfClosing {
				textStart := elemEnd + 1
				textEnd := findClosingTag(content, "text", textStart)
				if textEnd > textStart {
					textBody := content[textStart:textEnd]
					parseTextElement(elem, textBody, inherited, state)
				}
				closeIdx := strings.IndexByte(content[textEnd:], '>')
				if closeIdx < 0 {
					pos = len(content)
					continue
				}
				pos = textEnd + closeIdx + 1
			} else {
				pos = elemEnd + 1
			}

		case "animate":
			state.elemCount++
			if len(state.animations) < maxAnimations {
				if a, ok := parseAnimateElement(elem, inherited); ok {
					state.animations = append(state.animations, a)
					registerAnimation(state, elem,
						len(state.animations)-1)
				}
			}
			pos = elemEnd + 1

		case "animateTransform":
			state.elemCount++
			if len(state.animations) < maxAnimations {
				if a, ok := parseAnimateTransformElement(elem, inherited); ok {
					state.animations = append(state.animations, a)
					registerAnimation(state, elem,
						len(state.animations)-1)
				}
			}
			pos = elemEnd + 1

		case "set":
			state.elemCount++
			if len(state.animations) < maxAnimations {
				if a, ok := parseSetElement(elem, inherited); ok {
					state.animations = append(state.animations, a)
					registerAnimation(state, elem,
						len(state.animations)-1)
				}
			}
			pos = elemEnd + 1

		case "animateMotion":
			state.elemCount++
			motionBody, next := readElementBody(
				content, elem, elemEnd, "animateMotion")
			if len(state.animations) < maxAnimations {
				if a, ok := parseAnimateMotionElement(
					elem, motionBody, inherited, state); ok {
					state.animations = append(state.animations, a)
					registerAnimation(state, elem,
						len(state.animations)-1)
				}
			}
			pos = next

		default:
			pos = elemEnd + 1
		}
	}
	return paths
}

// parseShapeElement parses a shape (path/rect/circle/…) via the
// supplied parser, handling nested <animate>/<animateTransform>
// children when the element is not self-closing. When inline
// animations are present the shape's id (or a synthesized one)
// is propagated to the path's GroupID so animations key onto
// the owning shape. Returns the new pos after parsing.
func parseShapeElement(
	content, elem, tagName string, elemEnd int,
	isSelfClosing bool,
	inherited groupStyle,
	state *parseState,
	paths *[]VectorPath,
	parser func(gs groupStyle) (VectorPath, bool),
) int {
	state.elemCount++

	if isSelfClosing {
		if p, ok := parser(inherited); ok {
			*paths = append(*paths, p)
		}
		return elemEnd + 1
	}

	bodyStart := elemEnd + 1
	bodyEnd := findClosingTag(content, tagName, bodyStart)
	if bodyEnd <= bodyStart {
		if p, ok := parser(inherited); ok {
			*paths = append(*paths, p)
		}
		return elemEnd + 1
	}
	body := content[bodyStart:bodyEnd]

	shapeGS := inherited
	if shapeHasInlineAnimation(body) {
		gid, ok := findAttr(elem, "id")
		if !ok || gid == "" {
			state.synthID++
			gid = fmt.Sprintf("__anim_%d", state.synthID)
		}
		shapeGS.GroupID = gid
		all, fill, stroke := scanOpacityAnimTargets(body)
		shapeGS.SkipOpacity = all
		shapeGS.SkipFillOpacity = fill
		shapeGS.SkipStrokeOpacity = stroke
	}

	pathIdx := -1
	if p, ok := parser(shapeGS); ok {
		if shapeGS.GroupID != inherited.GroupID {
			p.GroupID = shapeGS.GroupID
		}
		*paths = append(*paths, p)
		pathIdx = len(*paths) - 1
	}

	animStart := len(state.animations)
	parseShapeInlineChildren(body, shapeGS, state)
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

	closeEnd := strings.IndexByte(content[bodyEnd:], '>')
	if closeEnd < 0 {
		return len(content)
	}
	return bodyEnd + closeEnd + 1
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

// shapeHasInlineAnimation cheaply detects whether a shape body
// contains animation children worth parsing.
func shapeHasInlineAnimation(body string) bool {
	return strings.Contains(body, "<animate") ||
		strings.Contains(body, "<animateTransform") ||
		strings.Contains(body, "<animateMotion") ||
		strings.Contains(body, "<set")
}

// scanOpacityAnimTargets reports which opacity sub-attributes are
// animated by inline <animate attributeName="..."> children of a
// shape body. Used to suppress static opacity baking for channels
// the animation will overwrite at render time.
func scanOpacityAnimTargets(body string) (all, fill, stroke bool) {
	pos := 0
	for pos < len(body) {
		idx := strings.Index(body[pos:], "<animate")
		if idx < 0 {
			return
		}
		start := pos + idx + len("<animate")
		end := strings.IndexByte(body[start:], '>')
		if end < 0 {
			return
		}
		elem := body[pos+idx : start+end+1]
		pos = start + end + 1
		// Skip <animateTransform> — it never targets opacity.
		if strings.HasPrefix(elem, "<animateTransform") {
			continue
		}
		attr, ok := findAttr(elem, "attributeName")
		if !ok {
			continue
		}
		switch attr {
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

// parseShapeInlineChildren scans a shape body for
// <animate>/<animateTransform> elements and appends them to
// state.animations keyed by shapeGS.GroupID.
func parseShapeInlineChildren(
	body string, shapeGS groupStyle, state *parseState,
) {
	pos := 0
	for pos < len(body) {
		lt := strings.IndexByte(body[pos:], '<')
		if lt < 0 {
			return
		}
		start := pos + lt + 1
		if start >= len(body) {
			return
		}
		tagEnd := findTagNameEnd(body, start)
		if tagEnd <= start {
			pos = start
			continue
		}
		tag := body[start:tagEnd]
		elemEndRel := strings.IndexByte(body[start:], '>')
		if elemEndRel < 0 {
			return
		}
		elemEnd := start + elemEndRel
		elem := body[start-1 : elemEnd+1]

		switch tag {
		case "animate":
			if len(state.animations) < maxAnimations {
				if a, ok := parseAnimateForDispatch(elem, shapeGS); ok {
					state.animations = append(state.animations, a)
					registerAnimation(state, elem,
						len(state.animations)-1)
				}
			}
		case "animateTransform":
			if len(state.animations) < maxAnimations {
				if a, ok := parseAnimateTransformElement(elem, shapeGS); ok {
					state.animations = append(state.animations, a)
					registerAnimation(state, elem,
						len(state.animations)-1)
				}
			}
		case "set":
			if len(state.animations) < maxAnimations {
				if a, ok := parseSetElement(elem, shapeGS); ok {
					state.animations = append(state.animations, a)
					registerAnimation(state, elem,
						len(state.animations)-1)
				}
			}
		case "animateMotion":
			motionBody, next := readElementBody(
				body, elem, elemEnd, "animateMotion")
			if len(state.animations) < maxAnimations {
				if a, ok := parseAnimateMotionElement(
					elem, motionBody, shapeGS, state); ok {
					state.animations = append(state.animations, a)
					registerAnimation(state, elem,
						len(state.animations)-1)
				}
			}
			pos = next
			continue
		}
		pos = elemEnd + 1
	}
}

// readElementBody returns the inner content of an open element and
// the cursor position past its closing tag. For self-closing or
// missing-close elements returns empty body and pos just past elem.
func readElementBody(
	body, elem string, elemEnd int, tag string,
) (string, int) {
	if strings.HasSuffix(strings.TrimSpace(elem), "/>") {
		return "", elemEnd + 1
	}
	closeTok := "</" + tag
	closeIdx := strings.Index(body[elemEnd+1:], closeTok)
	if closeIdx < 0 {
		return "", elemEnd + 1
	}
	bodyStart := elemEnd + 1
	bodyEnd := bodyStart + closeIdx
	gt := strings.IndexByte(body[bodyEnd:], '>')
	if gt < 0 {
		return body[bodyStart:bodyEnd], bodyEnd
	}
	return body[bodyStart:bodyEnd], bodyEnd + gt + 1
}

// findRootElementTag locates the first `<tag` opening element in
// content and returns its full tag text up to and including the
// closing `>`. Used to read presentation attributes off the root
// <svg> element without recursing into it.
func findRootElementTag(content, tag string) (string, bool) {
	needle := "<" + tag
	idx := strings.Index(content, needle)
	if idx < 0 {
		return "", false
	}
	after := idx + len(needle)
	if after >= len(content) {
		return "", false
	}
	c := content[after]
	if c != ' ' && c != '\t' && c != '\n' && c != '\r' && c != '>' && c != '/' {
		return "", false
	}
	end := strings.IndexByte(content[idx:], '>')
	if end < 0 {
		return "", false
	}
	return content[idx : idx+end+1], true
}

func findTagNameEnd(s string, start int) int {
	i := start
	for i < len(s) {
		c := s[i]
		if c == ' ' || c == '\t' || c == '\n' || c == '\r' || c == '>' || c == '/' {
			break
		}
		i++
	}
	return i
}

func findClosingTag(content, tag string, start int) int {
	closeTag := "</" + tag
	openTag := "<" + tag
	depth := 1
	pos := start
	iterations := 0

	for pos < len(content) && depth > 0 {
		iterations++
		if iterations > maxElements {
			break
		}
		next := strings.IndexByte(content[pos:], '<')
		if next < 0 {
			break
		}
		next += pos

		// Closing tag — verify full tag name boundary so
		// e.g. "</textPath>" doesn't match "</text".
		if next+len(closeTag) <= len(content) && content[next:next+len(closeTag)] == closeTag {
			endPos := next + len(closeTag)
			if endPos >= len(content) || content[endPos] == '>' ||
				content[endPos] == ' ' || content[endPos] == '\t' ||
				content[endPos] == '\n' || content[endPos] == '\r' {
				depth--
				if depth == 0 {
					return next
				}
			}
			pos = next + len(closeTag)
			continue
		}

		// Opening tag
		if next+len(openTag) <= len(content) && content[next:next+len(openTag)] == openTag {
			endPos := next + len(openTag)
			if endPos < len(content) {
				c := content[endPos]
				if c == ' ' || c == '\t' || c == '\n' || c == '>' || c == '/' {
					depth++
				}
			}
		}
		pos = next + 1
	}
	return len(content)
}
