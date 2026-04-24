package svg

// xml_defs.go — SVG <defs> parsing: clip paths, gradients,
// filters, and named path definitions. Walks the decoded xmlNode
// tree; each helper finds the <defs> subtree and its children by
// element name.

import (
	"strings"

	"github.com/mike-ward/go-gui/gui"
)

// walkDefs invokes fn on every <defs> node in the subtree rooted at
// root, in document order.
func walkDefs(root *xmlNode, fn func(defs *xmlNode)) {
	for i := range root.Children {
		c := &root.Children[i]
		if c.Name == "defs" {
			fn(c)
		}
		// Nested defs under groups are rare but spec-legal.
		if len(c.Children) > 0 {
			walkDefs(c, fn)
		}
	}
}

// findAllByName walks root's descendants and returns pointers to
// every direct-or-nested node matching name. Order is pre-order.
func findAllByName(root *xmlNode, name string, out *[]*xmlNode) {
	for i := range root.Children {
		c := &root.Children[i]
		if c.Name == name {
			*out = append(*out, c)
		}
		if len(c.Children) > 0 {
			findAllByName(c, name, out)
		}
	}
}

func parseDefsClipPaths(root *xmlNode) map[string][]VectorPath {
	clipPaths := make(map[string][]VectorPath)
	var nodes []*xmlNode
	findAllByName(root, "clipPath", &nodes)
	for _, cp := range nodes {
		clipID := cp.AttrMap["id"]
		if clipID == "" {
			continue
		}
		if cp.SelfClose {
			continue
		}
		defStyle := defaultGroupStyle(identityTransform)
		st := &parseState{}
		paths := parseSvgContent(cp, defStyle, 0, st)
		if len(paths) > 0 {
			clipPaths[clipID] = paths
		}
	}
	return clipPaths
}

func parseDefsGradients(root *xmlNode) map[string]gui.SvgGradientDef {
	gradients := make(map[string]gui.SvgGradientDef)
	var nodes []*xmlNode
	findAllByName(root, "linearGradient", &nodes)
	for _, lg := range nodes {
		gradID := lg.AttrMap["id"]
		if gradID == "" {
			continue
		}

		unitsStr := lg.AttrMap["gradientUnits"]
		if unitsStr == "" {
			unitsStr = "objectBoundingBox"
		}
		isOBB := unitsStr != "userSpaceOnUse"

		x1 := parseGradientCoord(lg.AttrMap["x1"], isOBB)
		y1 := parseGradientCoord(lg.AttrMap["y1"], isOBB)
		x2 := parseGradientCoord(lg.AttrMap["x2"], isOBB)
		y2 := parseGradientCoord(lg.AttrMap["y2"], isOBB)

		stops := parseGradientStops(lg)

		gradients[gradID] = gui.SvgGradientDef{
			X1: x1, Y1: y1, X2: x2, Y2: y2,
			Stops:         stops,
			GradientUnits: unitsStr,
		}
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

func parseGradientStops(gradient *xmlNode) []gui.SvgGradientStop {
	var stops []gui.SvgGradientStop
	for i := range gradient.Children {
		c := &gradient.Children[i]
		if c.Name != "stop" {
			continue
		}
		stopElem := c.OpenTag

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
			// Bare "inherit" at gradient-stop scope has no parent to
			// resolve against; fall back to black.
			color = colorBlack
		} else if isSentinelColor(color) {
			// Explicit currentColor: preserve sentinel RGB so the
			// render-time tint can substitute, but lift A to opaque
			// before stop-opacity bakes so small opacities survive.
			color.A = 255
		}

		stopOpacity := parseOpacityAttr(stopElem, "stop-opacity", 1.0)
		if stopOpacity < 1.0 {
			color = applyOpacity(color, stopOpacity)
		}

		stops = append(stops, gui.SvgGradientStop{Offset: offset, Color: color})
	}
	return stops
}

func parseDefsPaths(root *xmlNode) map[string]string {
	paths := make(map[string]string)
	walkDefs(root, func(defs *xmlNode) {
		for i := range defs.Children {
			c := &defs.Children[i]
			if c.Name != "path" {
				continue
			}
			pid := c.AttrMap["id"]
			d := c.AttrMap["d"]
			if pid == "" || d == "" {
				continue
			}
			paths[pid] = d
		}
	})
	return paths
}

// parseDefsFilters extracts <filter> definitions from the tree.
func parseDefsFilters(root *xmlNode) map[string]gui.SvgFilter {
	filters := make(map[string]gui.SvgFilter)
	var nodes []*xmlNode
	findAllByName(root, "filter", &nodes)
	for _, f := range nodes {
		filterID := f.AttrMap["id"]
		if filterID == "" {
			continue
		}
		if f.SelfClose {
			continue
		}

		// Extract stdDeviation from feGaussianBlur child.
		stdDev := float32(0)
		if gb := f.findChild("feGaussianBlur"); gb != nil {
			if sd, ok := gb.AttrMap["stdDeviation"]; ok {
				stdDev = parseF32(sd)
			}
		}
		if stdDev <= 0 {
			continue
		}

		// Walk feMerge children for feMergeNode entries.
		blurLayers := 0
		keepSource := false
		if fm := f.findChild("feMerge"); fm != nil {
			for i := range fm.Children {
				c := &fm.Children[i]
				if c.Name != "feMergeNode" {
					continue
				}
				if c.AttrMap["in"] == "SourceGraphic" {
					keepSource = true
				} else {
					blurLayers++
				}
			}
		}
		// Top-level feMergeNode (no wrapping feMerge) still counts.
		for i := range f.Children {
			c := &f.Children[i]
			if c.Name != "feMergeNode" {
				continue
			}
			if c.AttrMap["in"] == "SourceGraphic" {
				keepSource = true
			} else {
				blurLayers++
			}
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
	}
	return filters
}
