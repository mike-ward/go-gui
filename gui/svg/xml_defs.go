package svg

// xml_defs.go — SVG <defs> parsing: clip paths, gradients,
// filters, and named path definitions.

import (
	"strings"

	"github.com/mike-ward/go-gui/gui"
)

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
