package svg

import (
	"math"
	"strconv"
	"strings"

	"github.com/mike-ward/go-gui/gui"
)

// Named SVG colors lookup table.
var stringColors = map[string]gui.SvgColor{
	"blue":                 {R: 0, G: 0, B: 255, A: 255},
	"red":                  {R: 255, G: 0, B: 0, A: 255},
	"green":                {R: 0, G: 128, B: 0, A: 255},
	"yellow":               {R: 255, G: 255, B: 0, A: 255},
	"orange":               {R: 255, G: 165, B: 0, A: 255},
	"purple":               {R: 128, G: 0, B: 128, A: 255},
	"black":                {R: 0, G: 0, B: 0, A: 255},
	"gray":                 {R: 128, G: 128, B: 128, A: 255},
	"grey":                 {R: 128, G: 128, B: 128, A: 255},
	"indigo":               {R: 75, G: 0, B: 130, A: 255},
	"pink":                 {R: 255, G: 192, B: 203, A: 255},
	"violet":               {R: 238, G: 130, B: 238, A: 255},
	"white":                {R: 255, G: 255, B: 255, A: 255},
	"cyan":                 {R: 0, G: 255, B: 255, A: 255},
	"magenta":              {R: 255, G: 0, B: 255, A: 255},
	"aliceblue":            {R: 240, G: 248, B: 255, A: 255},
	"antiquewhite":         {R: 250, G: 235, B: 215, A: 255},
	"aqua":                 {R: 0, G: 255, B: 255, A: 255},
	"aquamarine":           {R: 127, G: 255, B: 212, A: 255},
	"azure":                {R: 240, G: 255, B: 255, A: 255},
	"beige":                {R: 245, G: 245, B: 220, A: 255},
	"bisque":               {R: 255, G: 228, B: 196, A: 255},
	"blanchedalmond":       {R: 255, G: 235, B: 205, A: 255},
	"blueviolet":           {R: 138, G: 43, B: 226, A: 255},
	"brown":                {R: 165, G: 42, B: 42, A: 255},
	"burlywood":            {R: 222, G: 184, B: 135, A: 255},
	"cadetblue":            {R: 95, G: 158, B: 160, A: 255},
	"chartreuse":           {R: 127, G: 255, B: 0, A: 255},
	"chocolate":            {R: 210, G: 105, B: 30, A: 255},
	"coral":                {R: 255, G: 127, B: 80, A: 255},
	"cornflowerblue":       {R: 100, G: 149, B: 237, A: 255},
	"cornsilk":             {R: 255, G: 248, B: 220, A: 255},
	"crimson":              {R: 220, G: 20, B: 60, A: 255},
	"darkblue":             {R: 0, G: 0, B: 139, A: 255},
	"darkcyan":             {R: 0, G: 139, B: 139, A: 255},
	"darkgoldenrod":        {R: 184, G: 134, B: 11, A: 255},
	"darkgray":             {R: 169, G: 169, B: 169, A: 255},
	"darkgreen":            {R: 0, G: 100, B: 0, A: 255},
	"darkgrey":             {R: 169, G: 169, B: 169, A: 255},
	"darkkhaki":            {R: 189, G: 183, B: 107, A: 255},
	"darkmagenta":          {R: 139, G: 0, B: 139, A: 255},
	"darkolivegreen":       {R: 85, G: 107, B: 47, A: 255},
	"darkorange":           {R: 255, G: 140, B: 0, A: 255},
	"darkorchid":           {R: 153, G: 50, B: 204, A: 255},
	"darkred":              {R: 139, G: 0, B: 0, A: 255},
	"darksalmon":           {R: 233, G: 150, B: 122, A: 255},
	"darkseagreen":         {R: 143, G: 188, B: 143, A: 255},
	"darkslateblue":        {R: 72, G: 61, B: 139, A: 255},
	"darkslategray":        {R: 47, G: 79, B: 79, A: 255},
	"darkslategrey":        {R: 47, G: 79, B: 79, A: 255},
	"darkturquoise":        {R: 0, G: 206, B: 209, A: 255},
	"darkviolet":           {R: 148, G: 0, B: 211, A: 255},
	"deeppink":             {R: 255, G: 20, B: 147, A: 255},
	"deepskyblue":          {R: 0, G: 191, B: 255, A: 255},
	"dimgray":              {R: 105, G: 105, B: 105, A: 255},
	"dimgrey":              {R: 105, G: 105, B: 105, A: 255},
	"dodgerblue":           {R: 30, G: 144, B: 255, A: 255},
	"firebrick":            {R: 178, G: 34, B: 34, A: 255},
	"floralwhite":          {R: 255, G: 250, B: 240, A: 255},
	"forestgreen":          {R: 34, G: 139, B: 34, A: 255},
	"fuchsia":              {R: 255, G: 0, B: 255, A: 255},
	"gainsboro":            {R: 220, G: 220, B: 220, A: 255},
	"ghostwhite":           {R: 248, G: 248, B: 255, A: 255},
	"gold":                 {R: 255, G: 215, B: 0, A: 255},
	"goldenrod":            {R: 218, G: 165, B: 32, A: 255},
	"greenyellow":          {R: 173, G: 255, B: 47, A: 255},
	"honeydew":             {R: 240, G: 255, B: 240, A: 255},
	"hotpink":              {R: 255, G: 105, B: 180, A: 255},
	"indianred":            {R: 205, G: 92, B: 92, A: 255},
	"ivory":                {R: 255, G: 255, B: 240, A: 255},
	"khaki":                {R: 240, G: 230, B: 140, A: 255},
	"lavender":             {R: 230, G: 230, B: 250, A: 255},
	"lavenderblush":        {R: 255, G: 240, B: 245, A: 255},
	"lawngreen":            {R: 124, G: 252, B: 0, A: 255},
	"lemonchiffon":         {R: 255, G: 250, B: 205, A: 255},
	"lightblue":            {R: 173, G: 216, B: 230, A: 255},
	"lightcoral":           {R: 240, G: 128, B: 128, A: 255},
	"lightcyan":            {R: 224, G: 255, B: 255, A: 255},
	"lightgoldenrodyellow": {R: 250, G: 250, B: 210, A: 255},
	"lightgray":            {R: 211, G: 211, B: 211, A: 255},
	"lightgreen":           {R: 144, G: 238, B: 144, A: 255},
	"lightgrey":            {R: 211, G: 211, B: 211, A: 255},
	"lightpink":            {R: 255, G: 182, B: 193, A: 255},
	"lightsalmon":          {R: 255, G: 160, B: 122, A: 255},
	"lightseagreen":        {R: 32, G: 178, B: 170, A: 255},
	"lightskyblue":         {R: 135, G: 206, B: 250, A: 255},
	"lightslategray":       {R: 119, G: 136, B: 153, A: 255},
	"lightslategrey":       {R: 119, G: 136, B: 153, A: 255},
	"lightsteelblue":       {R: 176, G: 196, B: 222, A: 255},
	"lightyellow":          {R: 255, G: 255, B: 224, A: 255},
	"lime":                 {R: 0, G: 255, B: 0, A: 255},
	"limegreen":            {R: 50, G: 205, B: 50, A: 255},
	"linen":                {R: 250, G: 240, B: 230, A: 255},
	"maroon":               {R: 128, G: 0, B: 0, A: 255},
	"mediumaquamarine":     {R: 102, G: 205, B: 170, A: 255},
	"mediumblue":           {R: 0, G: 0, B: 205, A: 255},
	"mediumorchid":         {R: 186, G: 85, B: 211, A: 255},
	"mediumpurple":         {R: 147, G: 111, B: 219, A: 255},
	"mediumseagreen":       {R: 60, G: 179, B: 113, A: 255},
	"mediumslateblue":      {R: 123, G: 104, B: 238, A: 255},
	"mediumspringgreen":    {R: 0, G: 250, B: 154, A: 255},
	"mediumturquoise":      {R: 72, G: 209, B: 204, A: 255},
	"mediumvioletred":      {R: 199, G: 21, B: 133, A: 255},
	"midnightblue":         {R: 25, G: 25, B: 112, A: 255},
	"mintcream":            {R: 245, G: 255, B: 250, A: 255},
	"mistyrose":            {R: 255, G: 228, B: 225, A: 255},
	"moccasin":             {R: 255, G: 228, B: 181, A: 255},
	"navajowhite":          {R: 255, G: 222, B: 173, A: 255},
	"navy":                 {R: 0, G: 0, B: 128, A: 255},
	"oldlace":              {R: 253, G: 245, B: 230, A: 255},
	"olive":                {R: 128, G: 128, B: 0, A: 255},
	"olivedrab":            {R: 107, G: 142, B: 35, A: 255},
	"orangered":            {R: 255, G: 69, B: 0, A: 255},
	"orchid":               {R: 218, G: 112, B: 214, A: 255},
	"palegoldenrod":        {R: 238, G: 232, B: 170, A: 255},
	"palegreen":            {R: 152, G: 251, B: 152, A: 255},
	"paleturquoise":        {R: 175, G: 238, B: 238, A: 255},
	"palevioletred":        {R: 219, G: 112, B: 147, A: 255},
	"papayawhip":           {R: 255, G: 239, B: 213, A: 255},
	"peachpuff":            {R: 255, G: 218, B: 185, A: 255},
	"peru":                 {R: 205, G: 133, B: 63, A: 255},
	"plum":                 {R: 221, G: 160, B: 221, A: 255},
	"powderblue":           {R: 176, G: 224, B: 230, A: 255},
	"rebeccapurple":        {R: 102, G: 51, B: 153, A: 255},
	"rosybrown":            {R: 188, G: 143, B: 143, A: 255},
	"royalblue":            {R: 65, G: 105, B: 225, A: 255},
	"saddlebrown":          {R: 139, G: 69, B: 19, A: 255},
	"salmon":               {R: 250, G: 128, B: 114, A: 255},
	"sandybrown":           {R: 244, G: 164, B: 96, A: 255},
	"seagreen":             {R: 46, G: 139, B: 87, A: 255},
	"seashell":             {R: 255, G: 245, B: 238, A: 255},
	"sienna":               {R: 160, G: 82, B: 45, A: 255},
	"silver":               {R: 192, G: 192, B: 192, A: 255},
	"skyblue":              {R: 135, G: 206, B: 235, A: 255},
	"slateblue":            {R: 106, G: 90, B: 205, A: 255},
	"slategray":            {R: 112, G: 128, B: 144, A: 255},
	"slategrey":            {R: 112, G: 128, B: 144, A: 255},
	"snow":                 {R: 255, G: 250, B: 250, A: 255},
	"springgreen":          {R: 0, G: 255, B: 127, A: 255},
	"steelblue":            {R: 70, G: 130, B: 180, A: 255},
	"tan":                  {R: 210, G: 180, B: 140, A: 255},
	"teal":                 {R: 0, G: 128, B: 128, A: 255},
	"thistle":              {R: 216, G: 191, B: 216, A: 255},
	"tomato":               {R: 255, G: 99, B: 71, A: 255},
	"turquoise":            {R: 64, G: 224, B: 208, A: 255},
	"wheat":                {R: 245, G: 222, B: 179, A: 255},
	"whitesmoke":           {R: 245, G: 245, B: 245, A: 255},
	"yellowgreen":          {R: 154, G: 205, B: 50, A: 255},
}

// parseSvgColor converts SVG color strings to SvgColor.
func parseSvgColor(s string) gui.SvgColor {
	str := strings.TrimSpace(s)
	if len(str) == 0 {
		return colorInherit
	}
	if str == "none" {
		return colorTransparent
	}
	if str == "currentColor" || str == "inherit" {
		return colorInherit
	}
	if strings.HasPrefix(str, "url(") {
		return colorTransparent
	}
	if str[0] == '#' {
		return parseHexColor(str)
	}
	if strings.HasPrefix(str, "rgb") {
		return parseRGBColor(str)
	}
	if c, ok := stringColors[str]; ok {
		return c
	}
	return gui.SvgColor{}
}

// parseHexColor parses #RGB, #RRGGBB, #RGBA, #RRGGBBAA.
func parseHexColor(s string) gui.SvgColor {
	hex := s[1:]
	switch len(hex) {
	case 3:
		r := hexDigit(hex[0]) * 17
		g := hexDigit(hex[1]) * 17
		b := hexDigit(hex[2]) * 17
		return gui.SvgColor{R: uint8(r), G: uint8(g), B: uint8(b), A: 255}
	case 4:
		r := hexDigit(hex[0]) * 17
		g := hexDigit(hex[1]) * 17
		b := hexDigit(hex[2]) * 17
		a := hexDigit(hex[3]) * 17
		return gui.SvgColor{R: uint8(r), G: uint8(g), B: uint8(b), A: uint8(a)}
	case 6:
		r := hexDigit(hex[0])*16 + hexDigit(hex[1])
		g := hexDigit(hex[2])*16 + hexDigit(hex[3])
		b := hexDigit(hex[4])*16 + hexDigit(hex[5])
		return gui.SvgColor{R: uint8(r), G: uint8(g), B: uint8(b), A: 255}
	case 8:
		r := hexDigit(hex[0])*16 + hexDigit(hex[1])
		g := hexDigit(hex[2])*16 + hexDigit(hex[3])
		b := hexDigit(hex[4])*16 + hexDigit(hex[5])
		a := hexDigit(hex[6])*16 + hexDigit(hex[7])
		return gui.SvgColor{R: uint8(r), G: uint8(g), B: uint8(b), A: uint8(a)}
	}
	return colorBlack
}

func hexDigit(c byte) int {
	switch {
	case c >= '0' && c <= '9':
		return int(c - '0')
	case c >= 'a' && c <= 'f':
		return int(c-'a') + 10
	case c >= 'A' && c <= 'F':
		return int(c-'A') + 10
	}
	return 0
}

// parseRGBColor parses rgb(r,g,b) or rgba(r,g,b,a).
func parseRGBColor(s string) gui.SvgColor {
	start := strings.IndexByte(s, '(')
	end := strings.IndexByte(s, ')')
	if start < 0 || end < 0 || end <= start+1 {
		return colorBlack
	}
	parts := strings.Split(s[start+1:end], ",")
	if len(parts) < 3 {
		return colorBlack
	}
	r := clampByte(parseIntTrimmed(parts[0]))
	g := clampByte(parseIntTrimmed(parts[1]))
	b := clampByte(parseIntTrimmed(parts[2]))
	a := 255
	if len(parts) >= 4 {
		alpha := parseFloatTrimmed(parts[3])
		if alpha <= 1.0 {
			a = clampByte(int(alpha * 255))
		} else {
			a = clampByte(int(alpha))
		}
	}
	return gui.SvgColor{R: uint8(r), G: uint8(g), B: uint8(b), A: uint8(a)}
}

func parseIntTrimmed(s string) int {
	v, _ := strconv.Atoi(strings.TrimSpace(s))
	return v
}

func parseFloatTrimmed(s string) float32 {
	v, _ := strconv.ParseFloat(strings.TrimSpace(s), 32)
	return float32(v)
}

func clampByte(v int) int {
	if v < 0 {
		return 0
	}
	if v > 255 {
		return 255
	}
	return v
}

// applyOpacity multiplies opacity into color alpha channel.
func applyOpacity(c gui.SvgColor, opacity float32) gui.SvgColor {
	if opacity >= 1.0 {
		return c
	}
	return gui.SvgColor{R: c.R, G: c.G, B: c.B, A: uint8(float32(c.A) * opacity)}
}

// parseOpacityAttr extracts an opacity value from element attrs.
func parseOpacityAttr(elem, name string, fallback float32) float32 {
	val, ok := findAttrOrStyle(elem, name)
	if !ok {
		return fallback
	}
	o := parseFloatTrimmed(val)
	if o < 0 {
		return 0
	}
	if o > 1.0 {
		return 1.0
	}
	return o
}

// parseLength parses a CSS length value (ignores units).
func parseLength(s string) float32 {
	end := 0
	for end < len(s) {
		c := s[end]
		if (c >= '0' && c <= '9') || c == '.' || c == '-' || c == '+' {
			end++
		} else if end > 0 && (c == 'e' || c == 'E') {
			// Scientific notation: accept 'e'/'E' optionally
			// followed by '+'/'-'.
			end++
			if end < len(s) && (s[end] == '+' || s[end] == '-') {
				end++
			}
		} else {
			break
		}
	}
	if end == 0 {
		return 0
	}
	v, _ := strconv.ParseFloat(s[:end], 32)
	value := float32(v)
	if value > maxCoordinate {
		return maxCoordinate
	}
	if value < -maxCoordinate {
		return -maxCoordinate
	}
	return value
}

func clampViewBoxDim(v float32) float32 {
	if v < 0 {
		return 0
	}
	if v > maxViewBoxDim {
		return maxViewBoxDim
	}
	return v
}

// --- Transform parsing ---

// matrixMultiply composes two affine transforms: result = m1 * m2.
func matrixMultiply(m1, m2 [6]float32) [6]float32 {
	return [6]float32{
		m1[0]*m2[0] + m1[2]*m2[1],
		m1[1]*m2[0] + m1[3]*m2[1],
		m1[0]*m2[2] + m1[2]*m2[3],
		m1[1]*m2[2] + m1[3]*m2[3],
		m1[0]*m2[4] + m1[2]*m2[5] + m1[4],
		m1[1]*m2[4] + m1[3]*m2[5] + m1[5],
	}
}

// parseTransform parses SVG transform attribute.
func parseTransform(s string) [6]float32 {
	result := identityTransform
	str := strings.TrimSpace(s)
	pos := 0
	count := 0

	for pos < len(str) {
		count++
		if count > 100 {
			break
		}
		// Skip whitespace and commas
		for pos < len(str) && (str[pos] == ' ' || str[pos] == ',' || str[pos] == '\t') {
			pos++
		}
		if pos >= len(str) {
			break
		}
		// Find transform name
		nameEnd := pos
		for nameEnd < len(str) && str[nameEnd] != '(' && str[nameEnd] != ' ' {
			nameEnd++
		}
		name := str[pos:nameEnd]

		parenStart := strings.IndexByte(str[nameEnd:], '(')
		if parenStart < 0 {
			break
		}
		parenStart += nameEnd
		parenEnd := strings.IndexByte(str[parenStart:], ')')
		if parenEnd < 0 {
			break
		}
		parenEnd += parenStart

		args := parseNumberList(str[parenStart+1 : parenEnd])
		m := parseSingleTransform(name, args)
		result = matrixMultiply(result, m)
		pos = parenEnd + 1
	}
	return result
}

func parseSingleTransform(name string, args []float32) [6]float32 {
	switch name {
	case "matrix":
		if len(args) >= 6 {
			return [6]float32{args[0], args[1], args[2], args[3], args[4], args[5]}
		}
	case "translate":
		tx := float32(0)
		ty := float32(0)
		if len(args) >= 1 {
			tx = args[0]
		}
		if len(args) >= 2 {
			ty = args[1]
		}
		return [6]float32{1, 0, 0, 1, tx, ty}
	case "scale":
		sx := float32(1)
		sy := sx
		if len(args) >= 1 {
			sx = args[0]
			sy = sx
		}
		if len(args) >= 2 {
			sy = args[1]
		}
		return [6]float32{sx, 0, 0, sy, 0, 0}
	case "rotate":
		return parseRotateTransform(args)
	case "skewX":
		if len(args) >= 1 {
			angle := args[0] * math.Pi / 180.0
			return [6]float32{1, 0, float32(math.Tan(float64(angle))), 1, 0, 0}
		}
	case "skewY":
		if len(args) >= 1 {
			angle := args[0] * math.Pi / 180.0
			return [6]float32{1, float32(math.Tan(float64(angle))), 0, 1, 0, 0}
		}
	}
	return identityTransform
}

func parseRotateTransform(args []float32) [6]float32 {
	if len(args) < 1 {
		return identityTransform
	}
	angle := float64(args[0]) * math.Pi / 180.0
	cosA := float32(math.Cos(angle))
	sinA := float32(math.Sin(angle))
	if len(args) >= 3 {
		cx := args[1]
		cy := args[2]
		return [6]float32{
			cosA, sinA, -sinA, cosA,
			cx - cosA*cx + sinA*cy,
			cy - sinA*cx - cosA*cy,
		}
	}
	return [6]float32{cosA, sinA, -sinA, cosA, 0, 0}
}

// applyTransformPt transforms a point by affine matrix.
func applyTransformPt(x, y float32, m [6]float32) (float32, float32) {
	return m[0]*x + m[2]*y + m[4], m[1]*x + m[3]*y + m[5]
}

func isIdentityTransform(m [6]float32) bool {
	return m[0] == 1 && m[1] == 0 && m[2] == 0 && m[3] == 1 && m[4] == 0 && m[5] == 0
}

// --- Attribute extraction ---

// findAttr extracts an attribute value from raw element text.
func findAttr(elem, name string) (string, bool) {
	pos := 0
	for pos < len(elem) {
		idx := strings.Index(elem[pos:], name)
		if idx < 0 {
			return "", false
		}
		idx += pos
		// Verify preceded by whitespace
		if idx == 0 || (elem[idx-1] != ' ' && elem[idx-1] != '\t' &&
			elem[idx-1] != '\n' && elem[idx-1] != '\r') {
			pos = idx + len(name)
			continue
		}
		// Check for '=' after name
		eq := idx + len(name)
		if eq >= len(elem) || elem[eq] != '=' {
			pos = eq
			continue
		}
		q := eq + 1
		if q >= len(elem) {
			return "", false
		}
		quote := elem[q]
		if quote != '"' && quote != '\'' {
			pos = q
			continue
		}
		start := q + 1
		endIdx := strings.IndexByte(elem[start:], quote)
		if endIdx < 0 {
			return "", false
		}
		attrLen := endIdx
		if attrLen > maxAttrLen {
			return "", false
		}
		if attrLen > 0 {
			return elem[start : start+endIdx], true
		}
		return "", false
	}
	return "", false
}

// findStyleProperty extracts a CSS property from a style string.
func findStyleProperty(style, name string) (string, bool) {
	pos := 0
	for pos < len(style) {
		idx := strings.Index(style[pos:], name)
		if idx < 0 {
			return "", false
		}
		idx += pos
		validStart := idx == 0 || style[idx-1] == ';' ||
			style[idx-1] == ' ' || style[idx-1] == '\t'
		if !validStart {
			pos = idx + len(name)
			continue
		}
		colon := idx + len(name)
		for colon < len(style) && (style[colon] == ' ' || style[colon] == '\t') {
			colon++
		}
		if colon >= len(style) || style[colon] != ':' {
			pos = colon
			continue
		}
		valStart := colon + 1
		valEnd := strings.IndexByte(style[valStart:], ';')
		if valEnd < 0 {
			valEnd = len(style) - valStart
		}
		if valEnd > 0 {
			return strings.TrimSpace(style[valStart : valStart+valEnd]), true
		}
		return "", false
	}
	return "", false
}

// findAttrOrStyle checks inline style first, then presentation attribute.
func findAttrOrStyle(elem, name string) (string, bool) {
	if style, ok := findAttr(elem, "style"); ok {
		if val, ok2 := findStyleProperty(style, name); ok2 {
			return val, true
		}
	}
	return findAttr(elem, name)
}

// parseFillURL extracts gradient ID from fill="url(#id)".
func parseFillURL(fill string) (string, bool) {
	str := strings.TrimSpace(fill)
	if !strings.HasPrefix(str, "url(") {
		return "", false
	}
	hashPos := strings.IndexByte(str, '#')
	if hashPos < 0 {
		return "", false
	}
	endPos := strings.IndexByte(str[hashPos:], ')')
	if endPos < 0 {
		return "", false
	}
	endPos += hashPos
	if endPos > hashPos+1 {
		return str[hashPos+1 : endPos], true
	}
	return "", false
}

// parseClipPathURL extracts clip path ID from clip-path="url(#id)".
func parseClipPathURL(elem string) (string, bool) {
	val, ok := findAttr(elem, "clip-path")
	if !ok {
		return "", false
	}
	hashPos := strings.IndexByte(val, '#')
	if hashPos < 0 {
		return "", false
	}
	endPos := strings.IndexByte(val[hashPos:], ')')
	if endPos < 0 {
		return "", false
	}
	endPos += hashPos
	if endPos > hashPos+1 {
		return val[hashPos+1 : endPos], true
	}
	return "", false
}

// parseFilterURL extracts filter ID from filter="url(#id)".
func parseFilterURL(elem string) (string, bool) {
	val, ok := findAttr(elem, "filter")
	if !ok {
		return "", false
	}
	hashPos := strings.IndexByte(val, '#')
	if hashPos < 0 {
		return "", false
	}
	endPos := strings.IndexByte(val[hashPos:], ')')
	if endPos < 0 {
		return "", false
	}
	endPos += hashPos
	if endPos > hashPos+1 {
		return val[hashPos+1 : endPos], true
	}
	return "", false
}

// --- Stroke attribute extraction ---

func getTransform(elem string) [6]float32 {
	if t, ok := findAttrOrStyle(elem, "transform"); ok {
		return parseTransform(t)
	}
	return identityTransform
}

func getStrokeColor(elem string) gui.SvgColor {
	stroke, ok := findAttrOrStyle(elem, "stroke")
	if !ok {
		return colorInherit
	}
	return parseSvgColor(stroke)
}

func getStrokeGradientID(elem string) string {
	stroke, ok := findAttrOrStyle(elem, "stroke")
	if !ok {
		return ""
	}
	id, _ := parseFillURL(stroke)
	return id
}

func getStrokeWidth(elem string) float32 {
	ws, ok := findAttrOrStyle(elem, "stroke-width")
	if !ok {
		return -1.0
	}
	return parseLength(ws)
}

func getStrokeLinecap(elem string) gui.StrokeCap {
	cap, ok := findAttrOrStyle(elem, "stroke-linecap")
	if !ok {
		return gui.StrokeCap(3) // inherit sentinel
	}
	switch cap {
	case "round":
		return gui.RoundCap
	case "square":
		return gui.SquareCap
	default:
		return gui.ButtCap
	}
}

func getStrokeLinejoin(elem string) gui.StrokeJoin {
	join, ok := findAttrOrStyle(elem, "stroke-linejoin")
	if !ok {
		return gui.StrokeJoin(3) // inherit sentinel
	}
	switch join {
	case "round":
		return gui.RoundJoin
	case "bevel":
		return gui.BevelJoin
	default:
		return gui.MiterJoin
	}
}

func getStrokeDasharray(elem string) []float32 {
	val, ok := findAttrOrStyle(elem, "stroke-dasharray")
	if !ok {
		return nil
	}
	if strings.TrimSpace(val) == "none" {
		return nil
	}
	parts := strings.Fields(strings.ReplaceAll(val, ",", " "))
	result := make([]float32, 0, len(parts))
	for _, p := range parts {
		n := parseFloatTrimmed(p)
		if n < 0 {
			return nil
		}
		result = append(result, n)
	}
	if len(result) > 0 && len(result)%2 != 0 {
		result = append(result, result...)
	}
	// Zero-sum dasharray = solid line (SVG spec).
	var sum float32
	for _, v := range result {
		sum += v
	}
	if sum <= 0 {
		return nil
	}
	return result
}

func parseElementStyle(elem string) elementStyle {
	return elementStyle{
		Transform:        getTransform(elem),
		StrokeColor:      getStrokeColor(elem),
		StrokeWidth:      getStrokeWidth(elem),
		StrokeCap:        getStrokeLinecap(elem),
		StrokeJoin:       getStrokeLinejoin(elem),
		Opacity:          parseOpacityAttr(elem, "opacity", 1.0),
		FillOpacity:      parseOpacityAttr(elem, "fill-opacity", 1.0),
		StrokeOpacity:    parseOpacityAttr(elem, "stroke-opacity", 1.0),
		StrokeGradientID: getStrokeGradientID(elem),
		StrokeDasharray:  getStrokeDasharray(elem),
	}
}

// --- Group style inheritance ---

const strokeCapInherit = gui.StrokeCap(3)
const strokeJoinInherit = gui.StrokeJoin(3)

func mergeGroupStyle(elem string, inherited groupStyle) groupStyle {
	elemTransform := getTransform(elem)
	combined := matrixMultiply(inherited.Transform, elemTransform)

	fill := attrOrDefault(elem, "fill", inherited.Fill)
	stroke := attrOrDefault(elem, "stroke", inherited.Stroke)
	strokeWidth := attrOrDefault(elem, "stroke-width", inherited.StrokeWidth)
	strokeCap := attrOrDefault(elem, "stroke-linecap", inherited.StrokeCap)
	strokeJoin := attrOrDefault(elem, "stroke-linejoin", inherited.StrokeJoin)
	clipID, _ := parseClipPathURL(elem)
	if clipID == "" {
		clipID = inherited.ClipPathID
	}
	filterID, _ := parseFilterURL(elem)
	if filterID == "" {
		filterID = inherited.FilterID
	}
	fontFamily := attrOrDefault(elem, "font-family", inherited.FontFamily)
	fontSize := attrOrDefault(elem, "font-size", inherited.FontSize)
	fontWeight := attrOrDefault(elem, "font-weight", inherited.FontWeight)
	fontStyle := attrOrDefault(elem, "font-style", inherited.FontStyle)
	textAnchor := attrOrDefault(elem, "text-anchor", inherited.TextAnchor)

	elemOpacity := parseOpacityAttr(elem, "opacity", 1.0)
	groupOpacity := inherited.Opacity * elemOpacity
	fillOpacity := parseOpacityAttr(elem, "fill-opacity", inherited.FillOpacity)
	strokeOpacity := parseOpacityAttr(elem, "stroke-opacity", inherited.StrokeOpacity)

	gid, ok := findAttr(elem, "id")
	if !ok {
		gid = inherited.GroupID
	}

	return groupStyle{
		Transform:     combined,
		Fill:          fill,
		Stroke:        stroke,
		StrokeWidth:   strokeWidth,
		StrokeCap:     strokeCap,
		StrokeJoin:    strokeJoin,
		ClipPathID:    clipID,
		FilterID:      filterID,
		FontFamily:    fontFamily,
		FontSize:      fontSize,
		FontWeight:    fontWeight,
		FontStyle:     fontStyle,
		TextAnchor:    textAnchor,
		Opacity:       groupOpacity,
		FillOpacity:   fillOpacity,
		StrokeOpacity: strokeOpacity,
		GroupID:       gid,
	}
}

func attrOrDefault(elem, name, fallback string) string {
	if val, ok := findAttrOrStyle(elem, name); ok {
		return val
	}
	return fallback
}

func applyInheritedStyle(path *VectorPath, inherited groupStyle) {
	path.Transform = matrixMultiply(inherited.Transform, path.Transform)

	if path.ClipPathID == "" && inherited.ClipPathID != "" {
		path.ClipPathID = inherited.ClipPathID
	}
	if path.FilterID == "" && inherited.FilterID != "" {
		path.FilterID = inherited.FilterID
	}

	// Inherit fill
	if path.FillGradientID != "" {
		// Gradient fill takes precedence
	} else if path.FillColor == colorInherit {
		if inherited.Fill != "" {
			if gid, ok := parseFillURL(inherited.Fill); ok {
				path.FillGradientID = gid
			} else {
				path.FillColor = parseSvgColor(inherited.Fill)
			}
		} else {
			path.FillColor = colorBlack
		}
	}

	// Inherit stroke
	if path.StrokeColor == colorInherit {
		if inherited.Stroke != "" {
			path.StrokeColor = parseSvgColor(inherited.Stroke)
		} else {
			path.StrokeColor = colorTransparent
		}
	}
	if inherited.StrokeWidth != "" && path.StrokeWidth < 0 {
		path.StrokeWidth = parseLength(inherited.StrokeWidth)
	}
	if path.StrokeWidth < 0 {
		path.StrokeWidth = 1.0
	}
	if inherited.StrokeCap != "" && path.StrokeCap == strokeCapInherit {
		switch inherited.StrokeCap {
		case "round":
			path.StrokeCap = gui.RoundCap
		case "square":
			path.StrokeCap = gui.SquareCap
		default:
			path.StrokeCap = gui.ButtCap
		}
	}
	if path.StrokeCap == strokeCapInherit {
		path.StrokeCap = gui.ButtCap
	}
	if inherited.StrokeJoin != "" && path.StrokeJoin == strokeJoinInherit {
		switch inherited.StrokeJoin {
		case "round":
			path.StrokeJoin = gui.RoundJoin
		case "bevel":
			path.StrokeJoin = gui.BevelJoin
		default:
			path.StrokeJoin = gui.MiterJoin
		}
	}
	if path.StrokeJoin == strokeJoinInherit {
		path.StrokeJoin = gui.MiterJoin
	}

	if path.GroupID == "" && inherited.GroupID != "" {
		path.GroupID = inherited.GroupID
	}

	// Opacity
	combinedOpacity := inherited.Opacity * path.Opacity
	fillOpacity := path.FillOpacity
	if fillOpacity >= 1.0 {
		fillOpacity = inherited.FillOpacity
	}
	strokeOpacity := path.StrokeOpacity
	if strokeOpacity >= 1.0 {
		strokeOpacity = inherited.StrokeOpacity
	}
	path.FillColor = applyOpacity(path.FillColor, combinedOpacity*fillOpacity)
	path.StrokeColor = applyOpacity(path.StrokeColor, combinedOpacity*strokeOpacity)
}

func f32Abs(x float32) float32 {
	if x < 0 {
		return -x
	}
	return x
}
