package svg

import (
	"fmt"
	"html"
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

	// Parse with viewBox offset
	vbTransform := identityTransform
	if vg.ViewBoxX != 0 || vg.ViewBoxY != 0 {
		vbTransform = [6]float32{1, 0, 0, 1, -vg.ViewBoxX, -vg.ViewBoxY}
	}
	defStyle := defaultGroupStyle(vbTransform)
	state := &parseState{}
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
				paths = append(paths, parseSvgContent(groupContent, gs, depth+1, state)...)
			}
			closeEnd := strings.IndexByte(content[groupEnd:], '>')
			if closeEnd < 0 {
				pos = len(content)
				continue
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
				}
			}
			pos = elemEnd + 1

		case "animateTransform":
			state.elemCount++
			if len(state.animations) < maxAnimations {
				if a, ok := parseAnimateTransformElement(elem, inherited); ok {
					state.animations = append(state.animations, a)
				}
			}
			pos = elemEnd + 1

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

// parseTextElement extracts text from a <text> element,
// including <tspan> children and <textPath> references.
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
	fontFamily := cleanFontFamily(inherited.FontFamily)
	if ff, ok := findAttrOrStyle(elem, "font-family"); ok {
		fontFamily = cleanFontFamily(ff)
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

	// Fill gradient.
	var fillGradientID string
	if gid, ok := parseFillURL(fillStr); ok {
		fillGradientID = gid
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

	fontWeight := parseFontWeight(inherited.FontWeight)
	if fw := attrOrDefault(elem, "font-weight", ""); fw != "" {
		fontWeight = parseFontWeight(fw)
	}
	bold := fontWeight >= 600
	italic := false
	if fs := attrOrDefault(elem, "font-style", inherited.FontStyle); fs == "italic" || fs == "oblique" {
		italic = true
	}

	// Text decoration.
	underline := false
	strikethrough := false
	if td, ok := findAttrOrStyle(elem, "text-decoration"); ok {
		if strings.Contains(td, "underline") {
			underline = true
		}
		if strings.Contains(td, "line-through") {
			strikethrough = true
		}
	}

	// Letter spacing.
	var letterSpacing float32
	if ls, ok := findAttrOrStyle(elem, "letter-spacing"); ok {
		letterSpacing = parseLength(ls)
	}

	// Stroke.
	var strokeColor gui.SvgColor
	var strokeWidth float32
	if sw, ok := findAttrOrStyle(elem, "stroke-width"); ok {
		strokeWidth = parseLength(sw)
	}
	if sc, ok := findAttrOrStyle(elem, "stroke"); ok {
		if sc != "none" {
			strokeColor = parseSvgColor(sc)
			if strokeColor == colorInherit {
				strokeColor = colorBlack
			}
			if strokeWidth == 0 {
				strokeWidth = 1
			}
		}
	}

	opacity := inherited.Opacity * parseOpacityAttr(elem, "opacity", 1.0)

	// Parse body: direct text, <tspan>, and <textPath>.
	parseTextBody(elem, body, textParentAttrs{
		x: x, y: y,
		fontSize: fontSize, fontFamily: fontFamily,
		color: color, fillGradientID: fillGradientID,
		anchor: anchor, bold: bold, italic: italic,
		fontWeight: fontWeight,
		underline:  underline, strikethrough: strikethrough,
		letterSpacing: letterSpacing,
		strokeColor:   strokeColor, strokeWidth: strokeWidth,
		opacity: opacity, filterID: inherited.FilterID,
	}, state)
}

// textParentAttrs holds inherited attributes from a <text> element.
type textParentAttrs struct {
	x, y                     float32
	fontSize                 float32
	fontFamily               string
	color                    gui.SvgColor
	fillGradientID           string
	anchor                   uint8
	bold, italic             bool
	fontWeight               int
	underline, strikethrough bool
	letterSpacing            float32
	strokeColor              gui.SvgColor
	strokeWidth              float32
	opacity                  float32
	filterID                 string
}

// parseTextBody parses direct text, <tspan>, and <textPath>
// children within a <text> element body.
func parseTextBody(_, body string, p textParentAttrs, state *parseState) {
	pos := 0
	curY := p.y

	// Extract direct text before first child element.
	text := extractPlainText(body)
	if text != "" {
		state.texts = append(state.texts, makeTextFromParent(
			text, p.x, curY, p))
	}

	// Iterate child elements.
	for pos < len(body) {
		start := strings.Index(body[pos:], "<")
		if start < 0 {
			break
		}
		start += pos

		// Skip closing tags.
		if start+1 < len(body) && body[start+1] == '/' {
			end := strings.IndexByte(body[start:], '>')
			if end < 0 {
				break
			}
			pos = start + end + 1
			continue
		}

		tagEnd := findTagNameEnd(body, start+1)
		if tagEnd <= start+1 {
			pos = start + 1
			continue
		}
		tagName := body[start+1 : tagEnd]

		elemEndRel := strings.IndexByte(body[start:], '>')
		if elemEndRel < 0 {
			break
		}
		elemEnd := start + elemEndRel
		childElem := body[start : elemEnd+1]
		isSelfClosing := elemEnd > 0 && body[elemEnd-1] == '/'

		switch tagName {
		case "tspan":
			if isSelfClosing {
				pos = elemEnd + 1
				continue
			}
			tspanStart := elemEnd + 1
			tspanEnd := findClosingTag(body, "tspan", tspanStart)
			if tspanEnd <= tspanStart {
				pos = elemEnd + 1
				continue
			}
			tspanBody := body[tspanStart:tspanEnd]
			parseTspan(childElem, tspanBody, p, &curY, state)

			closeIdx := strings.IndexByte(body[tspanEnd:], '>')
			if closeIdx < 0 {
				pos = tspanEnd
				continue
			}
			pos = tspanEnd + closeIdx + 1

		case "textPath":
			if isSelfClosing {
				pos = elemEnd + 1
				continue
			}
			tpStart := elemEnd + 1
			tpEnd := findClosingTag(body, "textPath", tpStart)
			if tpEnd <= tpStart {
				pos = elemEnd + 1
				continue
			}
			tpBody := body[tpStart:tpEnd]
			parseTextPathChild(childElem, tpBody, p, state)

			closeIdx := strings.IndexByte(body[tpEnd:], '>')
			if closeIdx < 0 {
				pos = tpEnd
				continue
			}
			pos = tpEnd + closeIdx + 1

		default:
			pos = elemEnd + 1
		}
	}
}

// makeTextFromParent creates an SvgText inheriting parent attrs.
func makeTextFromParent(text string, x, y float32, p textParentAttrs) gui.SvgText {
	return gui.SvgText{
		Text:           text,
		X:              x,
		Y:              y,
		FontFamily:     p.fontFamily,
		FontSize:       p.fontSize,
		IsBold:         p.bold,
		IsItalic:       p.italic,
		FontWeight:     p.fontWeight,
		Color:          p.color,
		FillGradientID: p.fillGradientID,
		FilterID:       p.filterID,
		Anchor:         int(p.anchor),
		Opacity:        p.opacity,
		Underline:      p.underline,
		Strikethrough:  p.strikethrough,
		LetterSpacing:  p.letterSpacing,
		StrokeColor:    p.strokeColor,
		StrokeWidth:    p.strokeWidth,
	}
}

// parseTspan parses a <tspan> element, inheriting parent <text>
// attrs and applying overrides.
func parseTspan(elem, body string, p textParentAttrs, curY *float32, state *parseState) {
	text := html.UnescapeString(strings.TrimSpace(body))
	if text == "" {
		return
	}

	// Position: absolute x/y or relative dy.
	tx := p.x
	if xv, ok := findAttr(elem, "x"); ok {
		tx = parseF32(xv)
	}
	ty := *curY
	if yv, ok := findAttr(elem, "y"); ok {
		ty = parseF32(yv)
	}
	if dy, ok := findAttr(elem, "dy"); ok {
		ty += parseLength(dy)
	}
	*curY = ty

	// Override attrs from tspan.
	fontSize := p.fontSize
	if fs, ok := findAttrOrStyle(elem, "font-size"); ok {
		fontSize = parseLength(fs)
	}
	fontFamily := p.fontFamily
	if ff, ok := findAttrOrStyle(elem, "font-family"); ok {
		fontFamily = cleanFontFamily(ff)
	}
	fontWeight := p.fontWeight
	bold := p.bold
	if fw, ok := findAttrOrStyle(elem, "font-weight"); ok {
		fontWeight = parseFontWeight(fw)
		bold = fontWeight >= 600
	}
	italic := p.italic
	if fs, ok := findAttrOrStyle(elem, "font-style"); ok {
		italic = fs == "italic" || fs == "oblique"
	}
	color := p.color
	fillGradientID := p.fillGradientID
	if f, ok := findAttrOrStyle(elem, "fill"); ok {
		if gid, gok := parseFillURL(f); gok {
			fillGradientID = gid
		} else {
			c := parseSvgColor(f)
			if c != colorInherit {
				color = c
				fillGradientID = ""
			}
		}
	}

	state.texts = append(state.texts, gui.SvgText{
		Text:           text,
		X:              tx,
		Y:              ty,
		FontFamily:     fontFamily,
		FontSize:       fontSize,
		IsBold:         bold,
		IsItalic:       italic,
		FontWeight:     fontWeight,
		Color:          color,
		FillGradientID: fillGradientID,
		FilterID:       p.filterID,
		Anchor:         int(p.anchor),
		Opacity:        p.opacity,
		Underline:      p.underline,
		Strikethrough:  p.strikethrough,
		LetterSpacing:  p.letterSpacing,
		StrokeColor:    p.strokeColor,
		StrokeWidth:    p.strokeWidth,
	})
}

// parseTextPathChild parses a <textPath> child element.
func parseTextPathChild(elem, body string, p textParentAttrs, state *parseState) {
	text := html.UnescapeString(strings.TrimSpace(body))
	if text == "" {
		return
	}

	// Extract href (try href first, then xlink:href).
	pathRef, ok := findAttr(elem, "href")
	if !ok {
		pathRef, ok = findAttr(elem, "xlink:href")
	}
	if !ok || pathRef == "" {
		return
	}
	pathID := strings.TrimPrefix(pathRef, "#")

	// startOffset.
	var startOffset float32
	isPercent := false
	if so, ok := findAttr(elem, "startOffset"); ok {
		trimmed := strings.TrimSpace(so)
		if strings.HasSuffix(trimmed, "%") {
			startOffset = parseF32(trimmed[:len(trimmed)-1])
			isPercent = true
		} else {
			startOffset = parseLength(trimmed)
		}
	}

	// text-anchor on <textPath> overrides parent <text>.
	anchor := p.anchor
	if anc, ok := findAttr(elem, "text-anchor"); ok {
		switch anc {
		case "middle":
			anchor = 1
		case "end":
			anchor = 2
		}
	}

	state.textPaths = append(state.textPaths, gui.SvgTextPath{
		Text:          text,
		PathID:        pathID,
		FontFamily:    p.fontFamily,
		FontSize:      p.fontSize,
		IsBold:        p.bold,
		IsItalic:      p.italic,
		FontWeight:    p.fontWeight,
		Color:         p.color,
		StrokeColor:   p.strokeColor,
		StrokeWidth:   p.strokeWidth,
		FilterID:      p.filterID,
		Anchor:        int(anchor),
		Opacity:       p.opacity,
		LetterSpacing: p.letterSpacing,
		StartOffset:   startOffset,
		IsPercent:     isPercent,
	})
}

// parseFontWeight converts a CSS font-weight string to a
// numeric value (100-900). Returns 0 for unset/inherit.
func parseFontWeight(fw string) int {
	switch fw {
	case "bold", "bolder":
		return 700
	case "normal", "lighter":
		return 400
	case "":
		return 0
	}
	w := int(parseF32(fw))
	if w >= 100 && w <= 900 {
		return w
	}
	return 0
}

// cleanFontFamily extracts the first font name from a CSS
// font-family list (e.g. "Courier New, monospace" → "Courier New").
func cleanFontFamily(ff string) string {
	if before, _, found := strings.Cut(ff, ","); found {
		return strings.TrimSpace(before)
	}
	return ff
}

func extractPlainText(body string) string {
	before, _, found := strings.Cut(body, "<")
	if !found {
		return html.UnescapeString(strings.TrimSpace(body))
	}
	return html.UnescapeString(strings.TrimSpace(before))
}

// --- Defs parsing ---

func parseDefsClipPaths(content string) map[string][]VectorPath {
	clipPaths := make(map[string][]VectorPath)
	pos := 0
	iterations := 0
	for pos < len(content) {
		iterations++
		if iterations > maxElements {
			break
		}
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
	iterations := 0
	for pos < len(content) {
		iterations++
		if iterations > maxElements {
			break
		}
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
	iterations := 0
	for pos < len(content) {
		iterations++
		if iterations > maxElements {
			break
		}
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
		offset = min(max(offset, 0), 1)

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
	iterations := 0
	for pos < len(content) {
		iterations++
		if iterations > maxElements {
			break
		}
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
		pIterations := 0
		for ppos < len(defsBody) {
			pIterations++
			if pIterations > maxElements {
				break
			}
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

// parseDefsFilters extracts <filter> definitions from SVG content.
func parseDefsFilters(content string) map[string]gui.SvgFilter {
	filters := make(map[string]gui.SvgFilter)
	pos := 0
	iterations := 0
	for pos < len(content) {
		iterations++
		if iterations > maxElements {
			break
		}
		fStart := strings.Index(content[pos:], "<filter")
		if fStart < 0 {
			break
		}
		fStart += pos
		tagEndRel := strings.IndexByte(content[fStart:], '>')
		if tagEndRel < 0 {
			break
		}
		tagEnd := fStart + tagEndRel
		openingTag := content[fStart : tagEnd+1]
		isSelfClosing := content[tagEnd-1] == '/'

		filterID, ok := findAttr(openingTag, "id")
		if !ok {
			pos = tagEnd + 1
			continue
		}
		if isSelfClosing {
			pos = tagEnd + 1
			continue
		}

		fContentStart := tagEnd + 1
		fEnd := findClosingTag(content, "filter", fContentStart)
		if fEnd <= fContentStart {
			pos = tagEnd + 1
			continue
		}
		fContent := content[fContentStart:fEnd]

		// Extract stdDeviation from feGaussianBlur
		stdDev := float32(0)
		gbStart := strings.Index(fContent, "<feGaussianBlur")
		if gbStart >= 0 {
			gbEndRel := strings.IndexByte(fContent[gbStart:], '>')
			if gbEndRel > 0 {
				gbElem := fContent[gbStart : gbStart+gbEndRel+1]
				if sd, found := findAttr(gbElem, "stdDeviation"); found {
					stdDev = parseF32(sd)
				}
			}
		}

		if stdDev <= 0 {
			closeEndRel := strings.IndexByte(content[fEnd:], '>')
			if closeEndRel < 0 {
				break
			}
			pos = fEnd + closeEndRel + 1
			continue
		}

		// Count feMergeNode entries
		blurLayers := 0
		keepSource := false
		mPos := 0
		mIterations := 0
		for mPos < len(fContent) {
			mIterations++
			if mIterations > maxElements {
				break
			}
			mnStart := strings.Index(fContent[mPos:], "<feMergeNode")
			if mnStart < 0 {
				break
			}
			mnStart += mPos
			mnEndRel := strings.IndexByte(fContent[mnStart:], '>')
			if mnEndRel < 0 {
				break
			}
			mnEnd := mnStart + mnEndRel
			mnElem := fContent[mnStart : mnEnd+1]
			inVal, _ := findAttr(mnElem, "in")
			if inVal == "SourceGraphic" {
				keepSource = true
			} else {
				blurLayers++
			}
			mPos = mnEnd + 1
		}
		if blurLayers == 0 {
			blurLayers = 1
		}

		filters[filterID] = gui.SvgFilter{
			ID:         filterID,
			StdDev:     stdDev,
			BlurLayers: blurLayers,
			KeepSource: keepSource,
		}

		closeEndRel := strings.IndexByte(content[fEnd:], '>')
		if closeEndRel < 0 {
			break
		}
		pos = fEnd + closeEndRel + 1
	}
	return filters
}
