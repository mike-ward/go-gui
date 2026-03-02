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
	vg.DefsPaths = parseDefsPaths(content)

	// Parse with viewBox offset
	vbTransform := identityTransform
	if vg.ViewBoxX != 0 || vg.ViewBoxY != 0 {
		vbTransform = [6]float32{1, 0, 0, 1, -vg.ViewBoxX, -vg.ViewBoxY}
	}
	defStyle := defaultGroupStyle(vbTransform)
	state := &parseState{}
	allPaths := parseSvgContent(content, defStyle, 0, state)

	vg.Paths = allPaths
	vg.Texts = state.texts
	vg.TextPaths = state.textPaths
	vg.Animations = state.animations
	return vg, nil
}

// parseSvgFile loads and parses an SVG file.
func parseSvgFile(path string) (*VectorGraphic, error) {
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
				paths = append(paths, parseSvgContent(groupContent, gs, depth+1, state)...)
			}
			closeEnd := strings.IndexByte(content[groupEnd:], '>')
			if closeEnd < 0 {
				break
			}
			pos = groupEnd + closeEnd + 1

		case "path":
			state.elemCount++
			if p, ok := parsePathWithStyle(elem, inherited); ok {
				paths = append(paths, p)
			}
			pos = elemEnd + 1

		case "rect":
			state.elemCount++
			if p, ok := parseRectWithStyle(elem, inherited); ok {
				paths = append(paths, p)
			}
			pos = elemEnd + 1

		case "circle":
			state.elemCount++
			if p, ok := parseCircleWithStyle(elem, inherited); ok {
				paths = append(paths, p)
			}
			pos = elemEnd + 1

		case "ellipse":
			state.elemCount++
			if p, ok := parseEllipseWithStyle(elem, inherited); ok {
				paths = append(paths, p)
			}
			pos = elemEnd + 1

		case "polygon":
			state.elemCount++
			if p, ok := parsePolygonWithStyle(elem, inherited, true); ok {
				paths = append(paths, p)
			}
			pos = elemEnd + 1

		case "polyline":
			state.elemCount++
			if p, ok := parsePolygonWithStyle(elem, inherited, false); ok {
				paths = append(paths, p)
			}
			pos = elemEnd + 1

		case "line":
			state.elemCount++
			if p, ok := parseLineWithStyle(elem, inherited); ok {
				paths = append(paths, p)
			}
			pos = elemEnd + 1

		case "text":
			state.elemCount++
			if !isSelfClosing {
				textStart := elemEnd + 1
				textEnd := findClosingTag(content, "text", textStart)
				if textEnd > textStart {
					// Basic text extraction
					textBody := content[textStart:textEnd]
					parseTextElement(elem, textBody, inherited, state)
				}
				closeIdx := strings.IndexByte(content[textEnd:], '>')
				if closeIdx < 0 {
					break
				}
				pos = textEnd + closeIdx + 1
			} else {
				pos = elemEnd + 1
			}

		default:
			pos = elemEnd + 1
		}
	}
	return paths
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

		// Closing tag
		if next+len(closeTag) <= len(content) && content[next:next+len(closeTag)] == closeTag {
			depth--
			if depth == 0 {
				return next
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

// parseTextElement extracts text from a <text> element.
func parseTextElement(elem, body string, inherited groupStyle, state *parseState) {
	x := attrFloat(elem, "x", 0)
	y := attrFloat(elem, "y", 0)
	fontSize := float32(16)
	if inherited.FontSize != "" {
		fontSize = parseLength(inherited.FontSize)
	}
	if fs, ok := findAttrOrStyle(elem, "font-size"); ok {
		fontSize = parseLength(fs)
	}
	fontFamily := inherited.FontFamily
	if ff, ok := findAttrOrStyle(elem, "font-family"); ok {
		fontFamily = ff
	}
	fillStr, _ := findAttrOrStyle(elem, "fill")
	color := parseSvgColor(fillStr)
	if color == colorInherit {
		if inherited.Fill != "" {
			color = parseSvgColor(inherited.Fill)
		} else {
			color = colorBlack
		}
	}

	anchor := uint8(0)
	if anc := attrOrDefault(elem, "text-anchor", inherited.TextAnchor); anc != "" {
		switch anc {
		case "middle":
			anchor = 1
		case "end":
			anchor = 2
		}
	}

	bold := false
	if fw := attrOrDefault(elem, "font-weight", inherited.FontWeight); fw == "bold" || fw == "700" {
		bold = true
	}
	italic := false
	if fs := attrOrDefault(elem, "font-style", inherited.FontStyle); fs == "italic" || fs == "oblique" {
		italic = true
	}

	opacity := inherited.Opacity * parseOpacityAttr(elem, "opacity", 1.0)

	// Extract plain text (before first child element)
	text := extractPlainText(body)
	if text != "" {
		state.texts = append(state.texts, gui.SvgText{
			Text:       text,
			X:          x,
			Y:          y,
			FontFamily: fontFamily,
			FontSize:   fontSize,
			IsBold:     bold,
			IsItalic:   italic,
			Color:      color,
			Anchor:     int(anchor),
			Opacity:    opacity,
		})
	}
}

func extractPlainText(body string) string {
	lt := strings.IndexByte(body, '<')
	if lt < 0 {
		return strings.TrimSpace(body)
	}
	return strings.TrimSpace(body[:lt])
}

// --- Defs parsing ---

func parseDefsClipPaths(content string) map[string][]VectorPath {
	clipPaths := make(map[string][]VectorPath)
	pos := 0
	for pos < len(content) {
		cpStart := strings.Index(content[pos:], "<clipPath")
		if cpStart < 0 {
			break
		}
		cpStart += pos
		tagEndRel := strings.IndexByte(content[cpStart:], '>')
		if tagEndRel < 0 {
			break
		}
		tagEnd := cpStart + tagEndRel
		openingTag := content[cpStart : tagEnd+1]
		isSelfClosing := content[tagEnd-1] == '/'

		clipID, ok := findAttr(openingTag, "id")
		if !ok {
			pos = tagEnd + 1
			continue
		}
		if isSelfClosing {
			pos = tagEnd + 1
			continue
		}

		cpContentStart := tagEnd + 1
		cpEnd := findClosingTag(content, "clipPath", cpContentStart)
		if cpEnd <= cpContentStart {
			pos = tagEnd + 1
			continue
		}

		cpContent := content[cpContentStart:cpEnd]
		defStyle := defaultGroupStyle(identityTransform)
		st := &parseState{}
		paths := parseSvgContent(cpContent, defStyle, 0, st)
		if len(paths) > 0 {
			clipPaths[clipID] = paths
		}

		closeEndRel := strings.IndexByte(content[cpEnd:], '>')
		if closeEndRel < 0 {
			break
		}
		pos = cpEnd + closeEndRel + 1
	}
	return clipPaths
}

func parseDefsGradients(content string) map[string]gui.SvgGradientDef {
	gradients := make(map[string]gui.SvgGradientDef)
	pos := 0
	for pos < len(content) {
		lgStart := strings.Index(content[pos:], "<linearGradient")
		if lgStart < 0 {
			break
		}
		lgStart += pos
		tagEndRel := strings.IndexByte(content[lgStart:], '>')
		if tagEndRel < 0 {
			break
		}
		tagEnd := lgStart + tagEndRel
		openingTag := content[lgStart : tagEnd+1]
		isSelfClosing := content[tagEnd-1] == '/'

		gradID, ok := findAttr(openingTag, "id")
		if !ok {
			pos = tagEnd + 1
			continue
		}

		unitsStr, _ := findAttr(openingTag, "gradientUnits")
		if unitsStr == "" {
			unitsStr = "objectBoundingBox"
		}
		isOBB := unitsStr != "userSpaceOnUse"

		x1Str, _ := findAttr(openingTag, "x1")
		y1Str, _ := findAttr(openingTag, "y1")
		x2Str, _ := findAttr(openingTag, "x2")
		y2Str, _ := findAttr(openingTag, "y2")

		x1 := parseGradientCoord(x1Str, isOBB)
		y1 := parseGradientCoord(y1Str, isOBB)
		x2 := parseGradientCoord(x2Str, isOBB)
		y2 := parseGradientCoord(y2Str, isOBB)

		if isSelfClosing {
			gradients[gradID] = gui.SvgGradientDef{
				X1: x1, Y1: y1, X2: x2, Y2: y2,
				GradientUnits: unitsStr,
			}
			pos = tagEnd + 1
			continue
		}

		lgContentStart := tagEnd + 1
		lgEnd := findClosingTag(content, "linearGradient", lgContentStart)
		if lgEnd <= lgContentStart {
			pos = tagEnd + 1
			continue
		}

		lgContent := content[lgContentStart:lgEnd]
		stops := parseGradientStops(lgContent)

		gradients[gradID] = gui.SvgGradientDef{
			X1: x1, Y1: y1, X2: x2, Y2: y2,
			Stops:         stops,
			GradientUnits: unitsStr,
		}

		closeEndRel := strings.IndexByte(content[lgEnd:], '>')
		if closeEndRel < 0 {
			break
		}
		pos = lgEnd + closeEndRel + 1
	}
	return gradients
}

func parseGradientCoord(s string, isOBB bool) float32 {
	trimmed := strings.TrimSpace(s)
	if isOBB && strings.HasSuffix(trimmed, "%") {
		return parseF32(trimmed[:len(trimmed)-1]) / 100.0
	}
	return parseF32(trimmed)
}

func parseGradientStops(content string) []gui.SvgGradientStop {
	var stops []gui.SvgGradientStop
	pos := 0
	for pos < len(content) {
		stopStart := strings.Index(content[pos:], "<stop")
		if stopStart < 0 {
			break
		}
		stopStart += pos
		stopEndRel := strings.IndexByte(content[stopStart:], '>')
		if stopEndRel < 0 {
			break
		}
		stopEnd := stopStart + stopEndRel
		stopElem := content[stopStart : stopEnd+1]

		offsetStr := "0"
		if os, ok := findAttrOrStyle(stopElem, "offset"); ok {
			offsetStr = os
		}
		offset := float32(0)
		if strings.HasSuffix(offsetStr, "%") {
			offset = parseF32(offsetStr[:len(offsetStr)-1]) / 100.0
		} else {
			offset = parseF32(offsetStr)
		}
		if offset < 0 {
			offset = 0
		}
		if offset > 1 {
			offset = 1
		}

		colorStr := "#000000"
		if cs, ok := findAttrOrStyle(stopElem, "stop-color"); ok {
			colorStr = cs
		}
		color := parseSvgColor(colorStr)
		if color == colorInherit {
			color = colorBlack
		}

		stopOpacity := parseOpacityAttr(stopElem, "stop-opacity", 1.0)
		if stopOpacity < 1.0 {
			color = applyOpacity(color, stopOpacity)
		}

		stops = append(stops, gui.SvgGradientStop{Offset: offset, Color: color})
		pos = stopEnd + 1
	}
	return stops
}

func parseDefsPaths(content string) map[string]string {
	paths := make(map[string]string)
	pos := 0
	for pos < len(content) {
		defsStart := strings.Index(content[pos:], "<defs")
		if defsStart < 0 {
			break
		}
		defsStart += pos
		defsTagEndRel := strings.IndexByte(content[defsStart:], '>')
		if defsTagEndRel < 0 {
			break
		}
		defsTagEnd := defsStart + defsTagEndRel
		isSelfClosing := content[defsTagEnd-1] == '/'
		if isSelfClosing {
			pos = defsTagEnd + 1
			continue
		}
		defsContentStart := defsTagEnd + 1
		defsEnd := findClosingTag(content, "defs", defsContentStart)
		if defsEnd <= defsContentStart {
			pos = defsTagEnd + 1
			continue
		}
		defsBody := content[defsContentStart:defsEnd]

		ppos := 0
		for ppos < len(defsBody) {
			pStart := strings.Index(defsBody[ppos:], "<path")
			if pStart < 0 {
				break
			}
			pStart += ppos
			pEndRel := strings.IndexByte(defsBody[pStart:], '>')
			if pEndRel < 0 {
				break
			}
			pEnd := pStart + pEndRel
			pElem := defsBody[pStart : pEnd+1]
			pid, ok := findAttr(pElem, "id")
			if !ok {
				ppos = pEnd + 1
				continue
			}
			d, ok := findAttr(pElem, "d")
			if !ok {
				ppos = pEnd + 1
				continue
			}
			paths[pid] = d
			ppos = pEnd + 1
		}

		closeEndRel := strings.IndexByte(content[defsEnd:], '>')
		if closeEndRel < 0 {
			break
		}
		pos = defsEnd + closeEndRel + 1
	}
	return paths
}
