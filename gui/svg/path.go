package svg

import (
	"math"
	"strconv"
	"strings"
)

// parsePathD parses the SVG path d attribute into segments.
func parsePathD(d string) []PathSegment {
	segments := make([]PathSegment, 0, 32)
	tokens := tokenizePath(d)
	i := 0

	var curX, curY float32
	var startX, startY float32
	var lastCtrlX, lastCtrlY float32
	var lastCmd byte

	for i < len(tokens) && len(segments) < maxPathSegments {
		token := tokens[i]
		if len(token) == 0 {
			i++
			continue
		}

		c := token[0]
		isCmd := (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z')

		cmd := lastCmd
		if isCmd {
			cmd = c
			i++
		}

		switch cmd {
		case 'M', 'm':
			relative := cmd == 'm'
			first := true
			for i < len(tokens) && isNumberToken(tokens[i]) {
				x := parseF32(tokens[i])
				y := float32(0)
				if i+1 < len(tokens) {
					y = parseF32(tokens[i+1])
				}
				i += 2
				if relative {
					curX += x
					curY += y
				} else {
					curX = x
					curY = y
				}
				if first {
					segments = append(segments, PathSegment{CmdMoveTo, []float32{curX, curY}})
					startX = curX
					startY = curY
					first = false
					if relative {
						cmd = 'l'
					} else {
						cmd = 'L'
					}
				} else {
					segments = append(segments, PathSegment{CmdLineTo, []float32{curX, curY}})
				}
			}

		case 'L', 'l':
			relative := cmd == 'l'
			for i < len(tokens) && isNumberToken(tokens[i]) {
				x := parseF32(tokens[i])
				y := float32(0)
				if i+1 < len(tokens) {
					y = parseF32(tokens[i+1])
				}
				i += 2
				if relative {
					curX += x
					curY += y
				} else {
					curX = x
					curY = y
				}
				segments = append(segments, PathSegment{CmdLineTo, []float32{curX, curY}})
			}

		case 'H', 'h':
			relative := cmd == 'h'
			for i < len(tokens) && isNumberToken(tokens[i]) {
				x := parseF32(tokens[i])
				i++
				if relative {
					curX += x
				} else {
					curX = x
				}
				segments = append(segments, PathSegment{CmdLineTo, []float32{curX, curY}})
			}

		case 'V', 'v':
			relative := cmd == 'v'
			for i < len(tokens) && isNumberToken(tokens[i]) {
				y := parseF32(tokens[i])
				i++
				if relative {
					curY += y
				} else {
					curY = y
				}
				segments = append(segments, PathSegment{CmdLineTo, []float32{curX, curY}})
			}

		case 'C', 'c':
			relative := cmd == 'c'
			for i+5 < len(tokens) && isNumberToken(tokens[i]) {
				c1x := parseF32(tokens[i])
				c1y := parseF32(tokens[i+1])
				c2x := parseF32(tokens[i+2])
				c2y := parseF32(tokens[i+3])
				x := parseF32(tokens[i+4])
				y := parseF32(tokens[i+5])
				i += 6
				if relative {
					segments = append(segments, PathSegment{CmdCubicTo, []float32{
						curX + c1x, curY + c1y,
						curX + c2x, curY + c2y,
						curX + x, curY + y,
					}})
					lastCtrlX = curX + c2x
					lastCtrlY = curY + c2y
					curX += x
					curY += y
				} else {
					segments = append(segments, PathSegment{CmdCubicTo, []float32{
						c1x, c1y, c2x, c2y, x, y,
					}})
					lastCtrlX = c2x
					lastCtrlY = c2y
					curX = x
					curY = y
				}
			}

		case 'S', 's':
			relative := cmd == 's'
			for i+3 < len(tokens) && isNumberToken(tokens[i]) {
				var c1x, c1y float32
				if lastCmd == 'C' || lastCmd == 'c' || lastCmd == 'S' || lastCmd == 's' {
					c1x = curX*2 - lastCtrlX
					c1y = curY*2 - lastCtrlY
				} else {
					c1x = curX
					c1y = curY
				}
				c2x := parseF32(tokens[i])
				c2y := parseF32(tokens[i+1])
				x := parseF32(tokens[i+2])
				y := parseF32(tokens[i+3])
				i += 4
				if relative {
					segments = append(segments, PathSegment{CmdCubicTo, []float32{
						c1x, c1y,
						curX + c2x, curY + c2y,
						curX + x, curY + y,
					}})
					lastCtrlX = curX + c2x
					lastCtrlY = curY + c2y
					curX += x
					curY += y
				} else {
					segments = append(segments, PathSegment{CmdCubicTo, []float32{
						c1x, c1y, c2x, c2y, x, y,
					}})
					lastCtrlX = c2x
					lastCtrlY = c2y
					curX = x
					curY = y
				}
				lastCmd = cmd
			}

		case 'Q', 'q':
			relative := cmd == 'q'
			for i+3 < len(tokens) && isNumberToken(tokens[i]) {
				cx := parseF32(tokens[i])
				cy := parseF32(tokens[i+1])
				x := parseF32(tokens[i+2])
				y := parseF32(tokens[i+3])
				i += 4
				if relative {
					segments = append(segments, PathSegment{CmdQuadTo, []float32{
						curX + cx, curY + cy,
						curX + x, curY + y,
					}})
					lastCtrlX = curX + cx
					lastCtrlY = curY + cy
					curX += x
					curY += y
				} else {
					segments = append(segments, PathSegment{CmdQuadTo, []float32{
						cx, cy, x, y,
					}})
					lastCtrlX = cx
					lastCtrlY = cy
					curX = x
					curY = y
				}
			}

		case 'T', 't':
			relative := cmd == 't'
			for i+1 < len(tokens) && isNumberToken(tokens[i]) {
				var cx, cy float32
				if lastCmd == 'Q' || lastCmd == 'q' || lastCmd == 'T' || lastCmd == 't' {
					cx = curX*2 - lastCtrlX
					cy = curY*2 - lastCtrlY
				} else {
					cx = curX
					cy = curY
				}
				x := parseF32(tokens[i])
				y := parseF32(tokens[i+1])
				i += 2
				if relative {
					segments = append(segments, PathSegment{CmdQuadTo, []float32{
						cx, cy, curX + x, curY + y,
					}})
					lastCtrlX = cx
					lastCtrlY = cy
					curX += x
					curY += y
				} else {
					segments = append(segments, PathSegment{CmdQuadTo, []float32{
						cx, cy, x, y,
					}})
					lastCtrlX = cx
					lastCtrlY = cy
					curX = x
					curY = y
				}
				lastCmd = cmd
			}

		case 'A', 'a':
			relative := cmd == 'a'
			for i+6 < len(tokens) && isNumberToken(tokens[i]) {
				rx := parseF32(tokens[i])
				ry := parseF32(tokens[i+1])
				phi := parseF32(tokens[i+2])
				largeArc := parseF32(tokens[i+3]) != 0
				sweep := parseF32(tokens[i+4]) != 0
				x := parseF32(tokens[i+5])
				y := parseF32(tokens[i+6])
				i += 7

				ex, ey := x, y
				if relative {
					ex += curX
					ey += curY
				}

				if rx <= 0 || ry <= 0 {
					segments = append(segments, PathSegment{CmdLineTo, []float32{ex, ey}})
				} else {
					arcSegs := arcToCubic(curX, curY, rx, ry, phi, largeArc, sweep, ex, ey)
					segments = append(segments, arcSegs...)
				}
				curX = ex
				curY = ey
			}

		case 'Z', 'z':
			segments = append(segments, PathSegment{CmdClose, nil})
			curX = startX
			curY = startY

		default:
			i++
		}
		lastCmd = cmd
	}
	return segments
}

// arcToCubic converts an SVG arc to cubic bezier curves.
func arcToCubic(x1, y1, rx, ry, phi float32, largeArc, sweep bool, x2, y2 float32) []PathSegment {
	if rx == 0 || ry == 0 {
		return []PathSegment{{CmdLineTo, []float32{x2, y2}}}
	}

	rxAbs := f32Abs(rx)
	ryAbs := f32Abs(ry)
	phiRad := float64(phi) * math.Pi / 180.0

	cosPhi := float32(math.Cos(phiRad))
	sinPhi := float32(math.Sin(phiRad))

	dx := (x1 - x2) / 2
	dy := (y1 - y2) / 2
	x1p := cosPhi*dx + sinPhi*dy
	y1p := -sinPhi*dx + cosPhi*dy

	// Correct radii
	lambda := (x1p*x1p)/(rxAbs*rxAbs) + (y1p*y1p)/(ryAbs*ryAbs)
	if lambda > 1 {
		sqrtLambda := float32(math.Sqrt(float64(lambda)))
		rxAbs *= sqrtLambda
		ryAbs *= sqrtLambda
	}

	rx2 := rxAbs * rxAbs
	ry2 := ryAbs * ryAbs
	x1p2 := x1p * x1p
	y1p2 := y1p * y1p

	sq := (rx2*ry2 - rx2*y1p2 - ry2*x1p2) / (rx2*y1p2 + ry2*x1p2)
	if sq < 0 {
		sq = 0
	}
	coef := float32(math.Sqrt(float64(sq)))
	if largeArc == sweep {
		coef = -coef
	}

	cxp := coef * rxAbs * y1p / ryAbs
	cyp := -coef * ryAbs * x1p / rxAbs

	cx := cosPhi*cxp - sinPhi*cyp + (x1+x2)/2
	cy := sinPhi*cxp + cosPhi*cyp + (y1+y2)/2

	theta1 := vectorAngle(1, 0, (x1p-cxp)/rxAbs, (y1p-cyp)/ryAbs)
	dtheta := vectorAngle((x1p-cxp)/rxAbs, (y1p-cyp)/ryAbs, (-x1p-cxp)/rxAbs, (-y1p-cyp)/ryAbs)

	if !sweep && dtheta > 0 {
		dtheta -= 2 * math.Pi
	} else if sweep && dtheta < 0 {
		dtheta += 2 * math.Pi
	}

	nSegs := int(math.Ceil(math.Abs(float64(dtheta)) / (math.Pi / 2)))
	dTheta := dtheta / float32(nSegs)

	segments := make([]PathSegment, 0, nSegs)
	theta := theta1
	for range nSegs {
		seg := arcSegmentToCubic(cx, cy, rxAbs, ryAbs, float32(phiRad), theta, dTheta)
		segments = append(segments, seg)
		theta += dTheta
	}
	return segments
}

func vectorAngle(ux, uy, vx, vy float32) float32 {
	n := float32(math.Sqrt(float64(ux*ux+uy*uy))) * float32(math.Sqrt(float64(vx*vx+vy*vy)))
	if n == 0 {
		return 0
	}
	c := (ux*vx + uy*vy) / n
	if c < -1 {
		c = -1
	}
	if c > 1 {
		c = 1
	}
	angle := float32(math.Acos(float64(c)))
	if ux*vy-uy*vx < 0 {
		return -angle
	}
	return angle
}

func arcSegmentToCubic(cx, cy, rx, ry, phi, theta, dtheta float32) PathSegment {
	t := float32(math.Tan(float64(dtheta/4))) * 4 / 3

	cosTheta := float32(math.Cos(float64(theta)))
	sinTheta := float32(math.Sin(float64(theta)))
	cosTheta2 := float32(math.Cos(float64(theta + dtheta)))
	sinTheta2 := float32(math.Sin(float64(theta + dtheta)))

	cosPhi := float32(math.Cos(float64(phi)))
	sinPhi := float32(math.Sin(float64(phi)))

	x1 := rx * cosTheta
	y1 := ry * sinTheta
	dx1 := -rx * sinTheta * t
	dy1 := ry * cosTheta * t

	x2 := rx * cosTheta2
	y2 := ry * sinTheta2
	dx2 := -rx * sinTheta2 * t
	dy2 := ry * cosTheta2 * t

	p1x := cosPhi*(x1+dx1) - sinPhi*(y1+dy1) + cx
	p1y := sinPhi*(x1+dx1) + cosPhi*(y1+dy1) + cy
	p2x := cosPhi*(x2-dx2) - sinPhi*(y2-dy2) + cx
	p2y := sinPhi*(x2-dx2) + cosPhi*(y2-dy2) + cy
	ex := cosPhi*x2 - sinPhi*y2 + cx
	ey := sinPhi*x2 + cosPhi*y2 + cy

	return PathSegment{CmdCubicTo, []float32{p1x, p1y, p2x, p2y, ex, ey}}
}

// tokenizePath splits path d string into tokens.
func tokenizePath(d string) []string {
	tokens := make([]string, 0, len(d)/4+1)
	var current strings.Builder
	current.Grow(16)
	hasDot := false

	for i := 0; i < len(d); i++ {
		if len(tokens) >= maxPathSegments {
			break
		}
		c := d[i]

		if c == ' ' || c == '\t' || c == '\n' || c == '\r' || c == ',' {
			if current.Len() > 0 {
				tokens = append(tokens, current.String())
				current.Reset()
				hasDot = false
			}
			continue
		}

		// Command letters
		if (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') {
			if (c == 'e' || c == 'E') && current.Len() > 0 {
				current.WriteByte(c)
				continue
			}
			if current.Len() > 0 {
				tokens = append(tokens, current.String())
				current.Reset()
				hasDot = false
			}
			tokens = append(tokens, string(c))
			continue
		}

		// Numbers
		if (c >= '0' && c <= '9') || c == '-' || c == '+' || c == '.' {
			if (c == '-' || c == '+') && current.Len() > 0 {
				last := current.String()
				lastByte := last[len(last)-1]
				if lastByte != 'e' && lastByte != 'E' {
					tokens = append(tokens, last)
					current.Reset()
					hasDot = false
				}
			}
			if c == '.' && hasDot {
				tokens = append(tokens, current.String())
				current.Reset()
				hasDot = false
			}
			current.WriteByte(c)
			if c == '.' {
				hasDot = true
			}
			continue
		}
	}

	if current.Len() > 0 {
		tokens = append(tokens, current.String())
	}
	return tokens
}

func isNumberToken(s string) bool {
	if len(s) == 0 {
		return false
	}
	c := s[0]
	return (c >= '0' && c <= '9') || c == '-' || c == '+' || c == '.'
}

func parseF32(s string) float32 {
	v, _ := strconv.ParseFloat(s, 32)
	if math.IsNaN(v) || math.IsInf(v, 0) {
		return 0
	}
	return float32(v)
}

// parseNumberList parses a space/comma separated list of numbers.
func parseNumberList(s string) []float32 {
	tokens := tokenizePath(s)
	numbers := make([]float32, 0, len(tokens))
	for _, t := range tokens {
		if isNumberToken(t) {
			numbers = append(numbers, parseF32(t))
		}
	}
	return numbers
}
