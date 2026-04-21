// Package svg parses and tessellates SVG content.
package svg

import (
	"strings"

	"github.com/mike-ward/go-gui/gui"
)

// parseAnimateElement parses an <animate> element targeting
// opacity. Returns the animation and true if valid.
func parseAnimateElement(
	elem string, inherited groupStyle,
) (gui.SvgAnimation, bool) {
	attr, ok := findAttr(elem, "attributeName")
	if !ok || attr != "opacity" {
		return gui.SvgAnimation{}, false
	}
	valStr, ok := findAttr(elem, "values")
	if !ok || valStr == "" {
		return gui.SvgAnimation{}, false
	}
	vals := parseSemicolonFloats(valStr)
	if len(vals) < 2 {
		return gui.SvgAnimation{}, false
	}
	dur := parseDuration(elem)
	if dur <= 0 {
		return gui.SvgAnimation{}, false
	}
	return gui.SvgAnimation{
		Kind:     gui.SvgAnimOpacity,
		GroupID:  inherited.GroupID,
		Values:   vals,
		DurSec:   dur,
		BeginSec: parseBeginOffset(elem),
	}, true
}

// parseAnimateTransformElement parses an <animateTransform>
// element targeting type="rotate". Accepts either from/to form
// or values="a cx cy;b cx cy;..." form. Returns animation and
// true if valid.
func parseAnimateTransformElement(
	elem string, inherited groupStyle,
) (gui.SvgAnimation, bool) {
	typ, ok := findAttr(elem, "type")
	if !ok || typ != "rotate" {
		return gui.SvgAnimation{}, false
	}
	dur := parseDuration(elem)
	if dur <= 0 {
		return gui.SvgAnimation{}, false
	}

	if valStr, ok := findAttr(elem, "values"); ok && valStr != "" {
		angles, cx, cy, ok := parseRotateValues(valStr)
		if !ok {
			return gui.SvgAnimation{}, false
		}
		return gui.SvgAnimation{
			Kind:     gui.SvgAnimRotate,
			GroupID:  inherited.GroupID,
			Values:   angles,
			CenterX:  cx,
			CenterY:  cy,
			DurSec:   dur,
			BeginSec: parseBeginOffset(elem),
		}, true
	}

	fromStr, _ := findAttr(elem, "from")
	toStr, _ := findAttr(elem, "to")
	if fromStr == "" || toStr == "" {
		return gui.SvgAnimation{}, false
	}
	fromParts := parseSpaceFloats(fromStr)
	toParts := parseSpaceFloats(toStr)
	if len(fromParts) < 3 || len(toParts) < 1 {
		return gui.SvgAnimation{}, false
	}
	return gui.SvgAnimation{
		Kind:     gui.SvgAnimRotate,
		GroupID:  inherited.GroupID,
		Values:   []float32{fromParts[0], toParts[0]},
		CenterX:  fromParts[1],
		CenterY:  fromParts[2],
		DurSec:   dur,
		BeginSec: parseBeginOffset(elem),
	}, true
}

// parseRotateValues parses a semicolon-separated list of rotate
// keyframes like "0 12 12;360 12 12". Each keyframe is "angle
// [cx cy]". Returns angle slice + center from the first keyframe.
// Center must stay constant across keyframes; mismatches are
// accepted but only the first is honored (rare in practice).
func parseRotateValues(s string) ([]float32, float32, float32, bool) {
	parts := strings.Split(s, ";")
	angles := make([]float32, 0, len(parts))
	var cx, cy float32
	first := true
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		trip := parseSpaceFloats(p)
		if len(trip) == 0 {
			return nil, 0, 0, false
		}
		angles = append(angles, trip[0])
		if first && len(trip) >= 3 {
			cx = trip[1]
			cy = trip[2]
		}
		first = false
	}
	if len(angles) < 2 {
		return nil, 0, 0, false
	}
	return angles, cx, cy, true
}

// parseDuration extracts the "dur" attribute as seconds.
func parseDuration(elem string) float32 {
	s, ok := findAttr(elem, "dur")
	if !ok || s == "" {
		return 0
	}
	return parseTimeValue(s)
}

// parseBeginOffset extracts the "begin" attribute as seconds.
func parseBeginOffset(elem string) float32 {
	s, ok := findAttr(elem, "begin")
	if !ok || s == "" {
		return 0
	}
	return parseTimeValue(s)
}

// parseTimeValue converts a time string like "1.5s" or "200ms"
// to seconds.
func parseTimeValue(s string) float32 {
	s = strings.TrimSpace(s)
	if strings.HasSuffix(s, "ms") {
		return parseF32(s[:len(s)-2]) / 1000
	}
	if strings.HasSuffix(s, "s") {
		return parseF32(s[:len(s)-1])
	}
	// Bare number defaults to seconds per SVG spec.
	return parseF32(s)
}

// parseSemicolonFloats splits a semicolon-separated string into
// float32 values.
func parseSemicolonFloats(s string) []float32 {
	parts := strings.Split(s, ";")
	out := make([]float32, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		out = append(out, parseF32(p))
	}
	return out
}

// parseSpaceFloats splits a space-separated string into float32
// values.
func parseSpaceFloats(s string) []float32 {
	fields := strings.Fields(s)
	out := make([]float32, 0, len(fields))
	for _, f := range fields {
		out = append(out, parseF32(f))
	}
	return out
}
