package svg

// xml_text.go — SVG <text>, <tspan>, and <textPath> parsing.

import (
	"html"
	"strings"

	"github.com/mike-ward/go-gui/gui"
)

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
