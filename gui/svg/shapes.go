package svg

// parsePathWithStyle parses a <path> element with inherited style.
func parsePathWithStyle(elem string, inherited groupStyle) (VectorPath, bool) {
	path, ok := parsePathElement(elem)
	if !ok {
		return VectorPath{}, false
	}
	if clipID, found := parseClipPathURL(elem); found {
		path.ClipPathID = clipID
	}
	applyInheritedStyle(&path, inherited)
	return path, true
}

func parseRectWithStyle(elem string, inherited groupStyle) (VectorPath, bool) {
	path, ok := parseRectElement(elem)
	if !ok {
		return VectorPath{}, false
	}
	if clipID, found := parseClipPathURL(elem); found {
		path.ClipPathID = clipID
	}
	applyInheritedStyle(&path, inherited)
	return path, true
}

func parseCircleWithStyle(elem string, inherited groupStyle) (VectorPath, bool) {
	path, ok := parseCircleElement(elem)
	if !ok {
		return VectorPath{}, false
	}
	if clipID, found := parseClipPathURL(elem); found {
		path.ClipPathID = clipID
	}
	applyInheritedStyle(&path, inherited)
	return path, true
}

func parseEllipseWithStyle(elem string, inherited groupStyle) (VectorPath, bool) {
	path, ok := parseEllipseElement(elem)
	if !ok {
		return VectorPath{}, false
	}
	if clipID, found := parseClipPathURL(elem); found {
		path.ClipPathID = clipID
	}
	applyInheritedStyle(&path, inherited)
	return path, true
}

func parsePolygonWithStyle(elem string, inherited groupStyle, close bool) (VectorPath, bool) {
	path, ok := parsePolygonElement(elem, close)
	if !ok {
		return VectorPath{}, false
	}
	if clipID, found := parseClipPathURL(elem); found {
		path.ClipPathID = clipID
	}
	applyInheritedStyle(&path, inherited)
	return path, true
}

func parseLineWithStyle(elem string, inherited groupStyle) (VectorPath, bool) {
	path, ok := parseLineElement(elem)
	if !ok {
		return VectorPath{}, false
	}
	if clipID, found := parseClipPathURL(elem); found {
		path.ClipPathID = clipID
	}
	applyInheritedStyle(&path, inherited)
	return path, true
}

// parsePathElement parses a <path> element.
func parsePathElement(elem string) (VectorPath, bool) {
	d, ok := findAttr(elem, "d")
	if !ok {
		return VectorPath{}, false
	}
	fill, _ := findAttrOrStyle(elem, "fill")
	s := parseElementStyle(elem)

	path := VectorPath{
		FillColor:        parseSvgColor(fill),
		Transform:        s.Transform,
		StrokeColor:      s.StrokeColor,
		StrokeWidth:      s.StrokeWidth,
		StrokeCap:        s.StrokeCap,
		StrokeJoin:       s.StrokeJoin,
		Opacity:          s.Opacity,
		FillOpacity:      s.FillOpacity,
		StrokeOpacity:    s.StrokeOpacity,
		StrokeGradientID: s.StrokeGradientID,
		StrokeDasharray:  s.StrokeDasharray,
	}
	if gid, found := parseFillURL(fill); found {
		path.FillGradientID = gid
	}
	path.Segments = parsePathD(d)
	if len(path.Segments) == 0 {
		return VectorPath{}, false
	}
	return path, true
}

// parseRectElement converts <rect> to path.
func parseRectElement(elem string) (VectorPath, bool) {
	x := attrFloat(elem, "x", 0)
	y := attrFloat(elem, "y", 0)
	w, wok := findAttr(elem, "width")
	h, hok := findAttr(elem, "height")
	if !wok || !hok {
		return VectorPath{}, false
	}
	rw := parseF32(w)
	rh := parseF32(h)

	rx := attrFloat(elem, "rx", 0)
	ry := attrFloat(elem, "ry", 0)
	fill, _ := findAttrOrStyle(elem, "fill")
	s := parseElementStyle(elem)

	if rx == 0 && ry > 0 {
		rx = ry
	}
	if ry == 0 && rx > 0 {
		ry = rx
	}

	var segments []PathSegment

	if rx == 0 && ry == 0 {
		segments = []PathSegment{
			{CmdMoveTo, []float32{x, y}},
			{CmdLineTo, []float32{x + rw, y}},
			{CmdLineTo, []float32{x + rw, y + rh}},
			{CmdLineTo, []float32{x, y + rh}},
			{CmdClose, nil},
		}
	} else {
		if rx > rw/2 {
			rx = rw / 2
		}
		if ry > rh/2 {
			ry = rh / 2
		}
		segments = make([]PathSegment, 0, 16)
		segments = append(segments, PathSegment{CmdMoveTo, []float32{x + rx, y}})
		segments = append(segments, PathSegment{CmdLineTo, []float32{x + rw - rx, y}})
		segments = append(segments, arcToCubic(x+rw-rx, y, rx, ry, 0, false, true, x+rw, y+ry)...)
		segments = append(segments, PathSegment{CmdLineTo, []float32{x + rw, y + rh - ry}})
		segments = append(segments, arcToCubic(x+rw, y+rh-ry, rx, ry, 0, false, true, x+rw-rx, y+rh)...)
		segments = append(segments, PathSegment{CmdLineTo, []float32{x + rx, y + rh}})
		segments = append(segments, arcToCubic(x+rx, y+rh, rx, ry, 0, false, true, x, y+rh-ry)...)
		segments = append(segments, PathSegment{CmdLineTo, []float32{x, y + ry}})
		segments = append(segments, arcToCubic(x, y+ry, rx, ry, 0, false, true, x+rx, y)...)
		segments = append(segments, PathSegment{CmdClose, nil})
	}

	vp := VectorPath{
		Segments:         segments,
		FillColor:        parseSvgColor(fill),
		Transform:        s.Transform,
		StrokeColor:      s.StrokeColor,
		StrokeWidth:      s.StrokeWidth,
		StrokeCap:        s.StrokeCap,
		StrokeJoin:       s.StrokeJoin,
		Opacity:          s.Opacity,
		FillOpacity:      s.FillOpacity,
		StrokeOpacity:    s.StrokeOpacity,
		StrokeGradientID: s.StrokeGradientID,
		StrokeDasharray:  s.StrokeDasharray,
	}
	if gid, found := parseFillURL(fill); found {
		vp.FillGradientID = gid
	}
	return vp, true
}

// parseCircleElement converts <circle> to path.
func parseCircleElement(elem string) (VectorPath, bool) {
	cx := attrFloat(elem, "cx", 0)
	cy := attrFloat(elem, "cy", 0)
	_, rok := findAttr(elem, "r")
	if !rok {
		return VectorPath{}, false
	}
	r := attrFloat(elem, "r", 0)
	fill, _ := findAttrOrStyle(elem, "fill")
	return ellipseToPath(cx, cy, r, r, elem, fill, parseElementStyle(elem)), true
}

// parseEllipseElement converts <ellipse> to path.
func parseEllipseElement(elem string) (VectorPath, bool) {
	_, rxok := findAttr(elem, "rx")
	_, ryok := findAttr(elem, "ry")
	if !rxok || !ryok {
		return VectorPath{}, false
	}
	cx := attrFloat(elem, "cx", 0)
	cy := attrFloat(elem, "cy", 0)
	rx := attrFloat(elem, "rx", 0)
	ry := attrFloat(elem, "ry", 0)
	fill, _ := findAttrOrStyle(elem, "fill")
	return ellipseToPath(cx, cy, rx, ry, elem, fill, parseElementStyle(elem)), true
}

func ellipseToPath(cx, cy, rx, ry float32, elem, fill string, s elementStyle) VectorPath {
	const k = float32(0.5522847498)
	kx := rx * k
	ky := ry * k

	segments := []PathSegment{
		{CmdMoveTo, []float32{cx, cy - ry}},
		{CmdCubicTo, []float32{cx + kx, cy - ry, cx + rx, cy - ky, cx + rx, cy}},
		{CmdCubicTo, []float32{cx + rx, cy + ky, cx + kx, cy + ry, cx, cy + ry}},
		{CmdCubicTo, []float32{cx - kx, cy + ry, cx - rx, cy + ky, cx - rx, cy}},
		{CmdCubicTo, []float32{cx - rx, cy - ky, cx - kx, cy - ry, cx, cy - ry}},
		{CmdClose, nil},
	}

	vp := VectorPath{
		Segments:         segments,
		FillColor:        parseSvgColor(fill),
		Transform:        s.Transform,
		StrokeColor:      s.StrokeColor,
		StrokeWidth:      s.StrokeWidth,
		StrokeCap:        s.StrokeCap,
		StrokeJoin:       s.StrokeJoin,
		Opacity:          s.Opacity,
		FillOpacity:      s.FillOpacity,
		StrokeOpacity:    s.StrokeOpacity,
		StrokeGradientID: s.StrokeGradientID,
		StrokeDasharray:  s.StrokeDasharray,
	}
	if gid, found := parseFillURL(fill); found {
		vp.FillGradientID = gid
	}
	return vp
}

// parsePolygonElement converts <polygon> or <polyline> to path.
func parsePolygonElement(elem string, close bool) (VectorPath, bool) {
	pointsStr, ok := findAttr(elem, "points")
	if !ok {
		return VectorPath{}, false
	}
	fill, _ := findAttrOrStyle(elem, "fill")
	s := parseElementStyle(elem)

	numbers := parseNumberList(pointsStr)
	if len(numbers) < 4 || len(numbers)%2 != 0 {
		return VectorPath{}, false
	}

	segments := make([]PathSegment, 0, len(numbers)/2+2)
	segments = append(segments, PathSegment{CmdMoveTo, []float32{numbers[0], numbers[1]}})
	for i := 2; i < len(numbers)-1; i += 2 {
		segments = append(segments, PathSegment{CmdLineTo, []float32{numbers[i], numbers[i+1]}})
	}
	if close {
		segments = append(segments, PathSegment{CmdClose, nil})
	}

	vp := VectorPath{
		Segments:         segments,
		FillColor:        parseSvgColor(fill),
		Transform:        s.Transform,
		StrokeColor:      s.StrokeColor,
		StrokeWidth:      s.StrokeWidth,
		StrokeCap:        s.StrokeCap,
		StrokeJoin:       s.StrokeJoin,
		Opacity:          s.Opacity,
		FillOpacity:      s.FillOpacity,
		StrokeOpacity:    s.StrokeOpacity,
		StrokeGradientID: s.StrokeGradientID,
		StrokeDasharray:  s.StrokeDasharray,
	}
	if gid, found := parseFillURL(fill); found {
		vp.FillGradientID = gid
	}
	return vp, true
}

// parseLineElement converts <line> to path.
func parseLineElement(elem string) (VectorPath, bool) {
	x1 := attrFloat(elem, "x1", 0)
	y1 := attrFloat(elem, "y1", 0)
	x2 := attrFloat(elem, "x2", 0)
	y2 := attrFloat(elem, "y2", 0)

	if x1 == x2 && y1 == y2 {
		return VectorPath{}, false
	}

	s := parseElementStyle(elem)
	return VectorPath{
		Segments: []PathSegment{
			{CmdMoveTo, []float32{x1, y1}},
			{CmdLineTo, []float32{x2, y2}},
		},
		FillColor:        colorTransparent,
		Transform:        s.Transform,
		StrokeColor:      s.StrokeColor,
		StrokeWidth:      s.StrokeWidth,
		StrokeCap:        s.StrokeCap,
		StrokeJoin:       s.StrokeJoin,
		Opacity:          s.Opacity,
		FillOpacity:      s.FillOpacity,
		StrokeOpacity:    s.StrokeOpacity,
		StrokeGradientID: s.StrokeGradientID,
		StrokeDasharray:  s.StrokeDasharray,
	}, true
}

func attrFloat(elem, name string, fallback float32) float32 {
	v, ok := findAttr(elem, name)
	if !ok {
		return fallback
	}
	return parseF32(v)
}

func newPathFromStyle(segments []PathSegment, fill string, s elementStyle) VectorPath {
	vp := VectorPath{
		Segments:         segments,
		FillColor:        parseSvgColor(fill),
		Transform:        s.Transform,
		StrokeColor:      s.StrokeColor,
		StrokeWidth:      s.StrokeWidth,
		StrokeCap:        s.StrokeCap,
		StrokeJoin:       s.StrokeJoin,
		Opacity:          s.Opacity,
		FillOpacity:      s.FillOpacity,
		StrokeOpacity:    s.StrokeOpacity,
		StrokeGradientID: s.StrokeGradientID,
		StrokeDasharray:  s.StrokeDasharray,
	}
	if gid, found := parseFillURL(fill); found {
		vp.FillGradientID = gid
	}
	return vp
}