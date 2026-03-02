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
// element targeting type="rotate". Returns animation and true
// if valid.
func parseAnimateTransformElement(
	elem string, inherited groupStyle,
) (gui.SvgAnimation, bool) {
	typ, ok := findAttr(elem, "type")
	if !ok || typ != "rotate" {
		return gui.SvgAnimation{}, false
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
	dur := parseDuration(elem)
	if dur <= 0 {
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
