package svg

import "strings"

// resolveTransformOrigin parses a CSS `transform-origin` value
// against a shape's bbox and returns the absolute (x, y) pivot in
// author coordinates.
//
// Recognized:
//   - Keywords: left/center/right (X), top/center/bottom (Y).
//   - Percentages: `50%` → bbox.MinX + 0.5*Width.
//   - Lengths: bare numbers and "px" treated identically (author units).
//
// Out of scope: em/rem/vw/vh/calc(). Such tokens fall through
// parseFloatTrimmed, yield 0, and pivot at the origin — accepted
// lossy behavior for spinners.
//
// Empty input falls back to bbox center — matches CSS default of
// `50% 50%`. The third (Z) component, if present, is ignored.
// Bbox.Set==false yields (0,0) because no geometry is known.
//
// Token order: with two tokens, the first is X and the second Y,
// unless an exclusively-Y keyword sits in the first slot (or an
// exclusively-X keyword sits in the second) — then swap. This
// honours forms like `top left` per CSS Transforms 1.
func resolveTransformOrigin(v string, b bbox) (float32, float32) {
	if !b.Set {
		return 0, 0
	}
	cx := b.MinX + b.Width()*0.5
	cy := b.MinY + b.Height()*0.5
	v = strings.TrimSpace(v)
	if v == "" {
		return cx, cy
	}
	parts := strings.Fields(v)
	var xTok, yTok string
	switch len(parts) {
	case 0:
		return cx, cy
	case 1:
		xTok, yTok = parts[0], "center"
		if isYKeywordOnly(parts[0]) {
			xTok, yTok = "center", parts[0]
		}
	default:
		xTok, yTok = parts[0], parts[1]
		if isYKeywordOnly(xTok) || isXKeywordOnly(yTok) {
			xTok, yTok = yTok, xTok
		}
	}
	return resolveAxisX(xTok, b, cx), resolveAxisY(yTok, b, cy)
}

func resolveAxisX(tok string, b bbox, fallback float32) float32 {
	if pct, ok := strings.CutSuffix(tok, "%"); ok {
		return b.MinX + parseFloatTrimmed(pct)/100*b.Width()
	}
	switch strings.ToLower(tok) {
	case "left":
		return b.MinX
	case "center":
		return fallback
	case "right":
		return b.MaxX
	}
	return parseFloatTrimmed(strings.TrimSuffix(tok, "px"))
}

func resolveAxisY(tok string, b bbox, fallback float32) float32 {
	if pct, ok := strings.CutSuffix(tok, "%"); ok {
		return b.MinY + parseFloatTrimmed(pct)/100*b.Height()
	}
	switch strings.ToLower(tok) {
	case "top":
		return b.MinY
	case "center":
		return fallback
	case "bottom":
		return b.MaxY
	}
	return parseFloatTrimmed(strings.TrimSuffix(tok, "px"))
}

func isXKeywordOnly(s string) bool {
	switch strings.ToLower(s) {
	case "left", "right":
		return true
	}
	return false
}

func isYKeywordOnly(s string) bool {
	switch strings.ToLower(s) {
	case "top", "bottom":
		return true
	}
	return false
}
