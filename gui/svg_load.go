package gui

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	glyph "github.com/mike-ward/go-glyph"
)

const maxSvgSourceBytes = int64(4 * 1024 * 1024)

var validSvgExtensions = []string{".svg"}

type svgParserCacheInvalidator interface {
	InvalidateSvgSource(svgSrc string)
	ClearSvgParserCache()
}

// CachedSvgPath holds tessellated geometry with vertex colors.
type CachedSvgPath struct {
	Triangles    []float32
	Color        Color
	VertexColors []Color
	IsClipMask   bool
	ClipGroup    int
	GroupID      string
}

// CachedSvgTextDraw holds cached text rendering data.
type CachedSvgTextDraw struct {
	Text      string
	TextStyle TextStyle
	X, Y      float32
	Gradient  *glyph.GradientConfig
}

// CachedFilteredGroup holds tessellated geometry for a filter group.
type CachedFilteredGroup struct {
	Filter      SvgFilter
	RenderPaths []CachedSvgPath
	Texts       []SvgText
	TextDraws   []CachedSvgTextDraw
	TextPaths   []SvgTextPath
	Gradients   map[string]SvgGradientDef
	BBox        [4]float32 // x, y, width, height
}

// CachedSvg holds pre-tessellated SVG data for efficient rendering.
type CachedSvg struct {
	RenderPaths    []CachedSvgPath
	Texts          []SvgText
	TextDraws      []CachedSvgTextDraw
	TextPaths      []SvgTextPath
	DefsPaths      map[string]string
	FilteredGroups []CachedFilteredGroup
	Gradients      map[string]SvgGradientDef
	Animations     []SvgAnimation
	HasAnimations  bool
	AnimStartNs    int64
	AnimHash       string
	Width          float32
	Height         float32
	Scale          float32
	defsPathData   map[string]cachedDefsPathData
}

// cachedSvgPaths converts TessellatedPath slices to CachedSvgPath.
func cachedSvgPaths(paths []TessellatedPath) []CachedSvgPath {
	out := make([]CachedSvgPath, len(paths))
	for i := range paths {
		p := &paths[i]
		var vcols []Color
		if len(p.VertexColors) != 0 {
			vcols = make([]Color, len(p.VertexColors))
			for j := range p.VertexColors {
				vc := p.VertexColors[j]
				vcols[j] = Color{vc.R, vc.G, vc.B, vc.A}
			}
		}
		out[i] = CachedSvgPath{
			Triangles:    p.Triangles,
			Color:        Color{p.Color.R, p.Color.G, p.Color.B, p.Color.A},
			VertexColors: vcols,
			IsClipMask:   p.IsClipMask,
			ClipGroup:    p.ClipGroup,
			GroupID:      p.GroupID,
		}
	}
	return out
}

// cachedSvgTextDraws converts SvgText elements to CachedSvgTextDraw.
func cachedSvgTextDraws(texts []SvgText, scale float32,
	gradients map[string]SvgGradientDef, w *Window) []CachedSvgTextDraw {
	draws := make([]CachedSvgTextDraw, 0, len(texts))
	for _, t := range texts {
		if len(t.Text) == 0 {
			continue
		}
		// Build Pango-style font name with weight/style.
		fontName := t.FontFamily
		if wn := pangoWeightName(t.FontWeight); wn != "" {
			fontName += " " + wn
		}
		typeface := glyph.TypefaceRegular
		if t.IsItalic {
			typeface = glyph.TypefaceItalic
		}
		ts := TextStyle{
			Family:        fontName,
			Size:          t.FontSize * scale,
			LetterSpacing: t.LetterSpacing * scale,
			Typeface:      typeface,
			Underline:     t.Underline,
			Strikethrough: t.Strikethrough,
			StrokeWidth:   t.StrokeWidth * scale,
			StrokeColor:   svgToColor(t.StrokeColor),
		}
		if t.Opacity < 1.0 {
			ts.Color = Color{t.Color.R, t.Color.G, t.Color.B,
				uint8(float32(t.Color.A) * t.Opacity)}
		} else {
			ts.Color = svgToColor(t.Color)
		}

		// Stroke-only text: fill=none + stroke set → transparent fill.
		if ts.StrokeWidth > 0 && ts.Color.A == 0 {
			ts.Color = Color{0, 0, 0, 0}
		}

		// Build gradient config from SVG gradient def.
		var grad *glyph.GradientConfig
		if t.FillGradientID != "" && gradients != nil {
			if gdef, ok := gradients[t.FillGradientID]; ok {
				grad = svgGradientToGlyph(gdef)
			}
		}

		// Measure text width for anchor adjustment.
		var tw float32
		var fh float32
		if w.textMeasurer != nil {
			tw = w.textMeasurer.TextWidth(t.Text, ts)
			fh = w.textMeasurer.FontHeight(ts)
		} else {
			fh = t.FontSize * scale
		}
		ascent := fh * 0.8
		x := t.X * scale
		y := t.Y*scale - ascent
		if t.Anchor == 1 {
			x -= tw / 2
		} else if t.Anchor == 2 {
			x -= tw
		}
		draws = append(draws, CachedSvgTextDraw{
			Text:      t.Text,
			TextStyle: ts,
			X:         x,
			Y:         y,
			Gradient:  grad,
		})
	}
	return draws
}

// pangoWeightName maps CSS font-weight (100-900) to a Pango
// weight descriptor. Returns "" for regular (400) or unset (0).
func pangoWeightName(w int) string {
	switch w {
	case 100:
		return "Thin"
	case 200:
		return "Ultra-Light"
	case 300:
		return "Light"
	case 500:
		return "Medium"
	case 600:
		return "Semi-Bold"
	case 700:
		return "Bold"
	case 800:
		return "Ultra-Bold"
	case 900:
		return "Heavy"
	default:
		return ""
	}
}

// svgGradientToGlyph converts an SvgGradientDef to a glyph
// GradientConfig.
func svgGradientToGlyph(g SvgGradientDef) *glyph.GradientConfig {
	if len(g.Stops) == 0 {
		return nil
	}
	stops := make([]glyph.GradientStop, len(g.Stops))
	for i, s := range g.Stops {
		stops[i] = glyph.GradientStop{
			Color:    glyph.Color{R: s.Color.R, G: s.Color.G, B: s.Color.B, A: s.Color.A},
			Position: s.Offset,
		}
	}
	dir := glyph.GradientHorizontal
	// Determine direction from gradient vector.
	dx := g.X2 - g.X1
	dy := g.Y2 - g.Y1
	if dy*dy > dx*dx {
		dir = glyph.GradientVertical
	}
	return &glyph.GradientConfig{
		Stops:     stops,
		Direction: dir,
	}
}

// validateSvgSource rejects file paths containing '..'.
func validateSvgSource(svgSrc string) error {
	return validateSvgSourceWithRoots(svgSrc, nil)
}

func validateSvgSourceWithRoots(svgSrc string, allowedRoots []string) error {
	if strings.HasPrefix(svgSrc, "<") {
		return nil
	}
	if strings.ContainsRune(svgSrc, 0) {
		return fmt.Errorf("invalid svg path: contains NUL")
	}
	cleanPath := filepath.Clean(svgSrc)
	if cleanPath == "." {
		return fmt.Errorf("invalid svg path")
	}
	for _, part := range strings.Split(filepath.ToSlash(cleanPath), "/") {
		if part == ".." {
			return fmt.Errorf("invalid svg path: contains ..")
		}
	}
	ext := strings.ToLower(filepath.Ext(cleanPath))
	valid := false
	for _, e := range validSvgExtensions {
		if ext == e {
			valid = true
			break
		}
	}
	if !valid {
		return fmt.Errorf("unsupported svg format: %s", ext)
	}
	if len(allowedRoots) > 0 {
		if err := validateSvgPathAllowed(cleanPath, allowedRoots); err != nil {
			return err
		}
	}
	return nil
}

func resolveValidatedSvgPath(svgSrc string, allowedRoots []string) (string, error) {
	if strings.HasPrefix(svgSrc, "<") {
		return svgSrc, nil
	}
	if err := validateSvgSourceWithRoots(svgSrc, allowedRoots); err != nil {
		return "", err
	}
	cleanPath := filepath.Clean(svgSrc)
	pathAbs, err := filepath.Abs(cleanPath)
	if err != nil {
		return "", fmt.Errorf("invalid svg path: %w", err)
	}
	resolvedPath := resolvePathWithParentFallback(pathAbs)
	if len(allowedRoots) > 0 {
		if err := validateSvgPathAllowed(resolvedPath, allowedRoots); err != nil {
			return "", err
		}
	}
	return resolvedPath, nil
}

func validateSvgPathAllowed(cleanPath string, allowedRoots []string) error {
	pathAbs, err := filepath.Abs(cleanPath)
	if err != nil {
		return fmt.Errorf("invalid svg path: %w", err)
	}
	resolvedPath := resolvePathWithParentFallback(pathAbs)
	for i := range allowedRoots {
		root := strings.TrimSpace(allowedRoots[i])
		if root == "" {
			continue
		}
		rootAbs, err := filepath.Abs(root)
		if err != nil {
			continue
		}
		resolvedRoot := resolvePathWithParentFallback(rootAbs)
		if pathWithinRoot(resolvedPath, resolvedRoot) {
			return nil
		}
	}
	return fmt.Errorf("svg path not allowed: %s", cleanPath)
}

func resolvePathWithParentFallback(path string) string {
	if p, err := filepath.EvalSymlinks(path); err == nil {
		return p
	}
	dir := filepath.Dir(path)
	if d, err := filepath.EvalSymlinks(dir); err == nil {
		return filepath.Join(d, filepath.Base(path))
	}
	return path
}

func pathWithinRoot(path, root string) bool {
	rel, err := filepath.Rel(root, path)
	if err != nil {
		return false
	}
	return rel == "." || (rel != ".." &&
		!strings.HasPrefix(rel, ".."+string(filepath.Separator)))
}

// checkSvgSourceSize validates SVG source size.
func checkSvgSourceSize(svgSrc string) error {
	if strings.HasPrefix(svgSrc, "<") {
		if int64(len(svgSrc)) > maxSvgSourceBytes {
			return fmt.Errorf("SVG source too large")
		}
		return nil
	}
	info, err := os.Stat(svgSrc)
	if err != nil {
		return fmt.Errorf("SVG not found: %s", svgSrc)
	}
	if info.Size() > maxSvgSourceBytes {
		return fmt.Errorf("SVG file too large")
	}
	return nil
}

func buildDefsPathDataCache(
	textPaths []SvgTextPath,
	filtered []SvgParsedFilteredGroup,
	defsPaths map[string]string,
	scale float32,
) map[string]cachedDefsPathData {
	if (len(textPaths) == 0 && len(filtered) == 0) || len(defsPaths) == 0 {
		return nil
	}
	pathIDs := make(map[string]struct{}, len(textPaths))
	for i := range textPaths {
		id := textPaths[i].PathID
		if id != "" {
			pathIDs[id] = struct{}{}
		}
	}
	for i := range filtered {
		for j := range filtered[i].TextPaths {
			id := filtered[i].TextPaths[j].PathID
			if id != "" {
				pathIDs[id] = struct{}{}
			}
		}
	}
	if len(pathIDs) == 0 {
		return nil
	}
	cached := make(map[string]cachedDefsPathData, len(pathIDs))
	for pathID := range pathIDs {
		d, ok := defsPaths[pathID]
		if !ok {
			continue
		}
		polyline := flattenDefsPath(d, scale)
		if len(polyline) < 4 {
			continue
		}
		table, totalLen := buildArcLengthTable(polyline)
		if totalLen <= 0 {
			continue
		}
		cached[pathID] = cachedDefsPathData{
			polyline: polyline,
			table:    table,
			totalLen: totalLen,
		}
	}
	if len(cached) == 0 {
		return nil
	}
	return cached
}

func svgHashHex(h uint64) string {
	var buf [24]byte
	b := strconv.AppendUint(buf[:0], h, 16)
	return string(b)
}

func buildSvgCacheKey(srcHash uint64, width, height float32) string {
	var buf [64]byte
	b := buf[:0]
	b = strconv.AppendUint(b, srcHash, 16)
	b = append(b, ':')
	b = strconv.AppendInt(b, int64(width*10), 10)
	b = append(b, 'x')
	b = strconv.AppendInt(b, int64(height*10), 10)
	return string(b)
}

// LoadSvg loads and tessellates an SVG, caching the result.
// svgSrc can be a file path or inline SVG data (starting with '<').
func (w *Window) LoadSvg(svgSrc string, width, height float32) (*CachedSvg, error) {
	srcHash := hashString(svgSrc)
	cacheKey := buildSvgCacheKey(srcHash, width, height)

	sm := StateMapRead[string, *CachedSvg](w, nsSvgCache)
	if sm != nil {
		if cached, ok := sm.Get(cacheKey); ok {
			return cached, nil
		}
	}

	resolvedSrc, err := resolveValidatedSvgPath(svgSrc, w.Config.AllowedSvgRoots)
	if err != nil {
		return nil, err
	}
	sizeSrc := svgSrc
	if !strings.HasPrefix(svgSrc, "<") {
		sizeSrc = resolvedSrc
	}
	if err := checkSvgSourceSize(sizeSrc); err != nil {
		return nil, err
	}

	if w.svgParser == nil {
		return nil, fmt.Errorf("no SVG parser configured")
	}

	var parsed *SvgParsed
	if strings.HasPrefix(svgSrc, "<") {
		parsed, err = w.svgParser.ParseSvg(svgSrc)
	} else {
		parsed, err = w.svgParser.ParseSvgFile(resolvedSrc)
	}
	if err != nil {
		return nil, err
	}

	// Cache dimensions.
	dimCache := StateMap[uint64, [2]float32](w, nsSvgDimCache, capModerate)
	dimCache.Set(srcHash, [2]float32{parsed.Width, parsed.Height})

	// Compute scale.
	scale := float32(1)
	if width > 0 && height > 0 {
		scaleX := float32(1)
		if parsed.Width > 0 {
			scaleX = width / parsed.Width
		}
		scaleY := float32(1)
		if parsed.Height > 0 {
			scaleY = height / parsed.Height
		}
		if scaleX < scaleY {
			scale = scaleX
		} else {
			scale = scaleY
		}
	}

	triangles := w.svgParser.Tessellate(parsed, scale)
	renderPaths := cachedSvgPaths(triangles)
	textDraws := cachedSvgTextDraws(parsed.Texts, scale, parsed.Gradients, w)
	defsPathData := buildDefsPathDataCache(parsed.TextPaths, parsed.FilteredGroups, parsed.DefsPaths, scale)

	// Build filtered groups.
	var filteredGroups []CachedFilteredGroup
	for _, fg := range parsed.FilteredGroups {
		fgPaths := cachedSvgPaths(fg.Paths)
		fgTextDraws := cachedSvgTextDraws(fg.Texts, scale, parsed.Gradients, w)
		filteredGroups = append(filteredGroups, CachedFilteredGroup{
			Filter:      fg.Filter,
			RenderPaths: fgPaths,
			Texts:       fg.Texts,
			TextDraws:   fgTextDraws,
			TextPaths:   fg.TextPaths,
			Gradients:   parsed.Gradients,
			BBox:        computeTriangleBBox(fg.Paths),
		})
	}

	cached := &CachedSvg{
		RenderPaths:    renderPaths,
		Texts:          parsed.Texts,
		TextDraws:      textDraws,
		TextPaths:      parsed.TextPaths,
		DefsPaths:      parsed.DefsPaths,
		FilteredGroups: filteredGroups,
		Gradients:      parsed.Gradients,
		Animations:     parsed.Animations,
		HasAnimations:  len(parsed.Animations) > 0,
		AnimStartNs:    time.Now().UnixNano(),
		AnimHash:       svgHashHex(srcHash),
		Width:          parsed.Width,
		Height:         parsed.Height,
		Scale:          scale,
		defsPathData:   defsPathData,
	}

	// Cache if vertex count is reasonable.
	totalVerts := 0
	for _, p := range renderPaths {
		totalVerts += len(p.Triangles)
	}
	const maxCachedVerts = 1_250_000
	if totalVerts <= maxCachedVerts {
		svgCache := StateMap[string, *CachedSvg](w, nsSvgCache, capModerate)
		svgCache.Set(cacheKey, cached)
	}
	return cached, nil
}

// GetSvgDimensions returns natural SVG dimensions without full
// parse+tessellate. Uses cached dimensions when available.
func (w *Window) GetSvgDimensions(svgSrc string) (float32, float32, error) {
	srcHash := hashString(svgSrc)
	dimCache := StateMapRead[uint64, [2]float32](w, nsSvgDimCache)
	if dimCache != nil {
		if dims, ok := dimCache.Get(srcHash); ok {
			return dims[0], dims[1], nil
		}
	}

	resolvedSrc, err := resolveValidatedSvgPath(svgSrc, w.Config.AllowedSvgRoots)
	if err != nil {
		return 0, 0, err
	}
	sizeSrc := svgSrc
	if !strings.HasPrefix(svgSrc, "<") {
		sizeSrc = resolvedSrc
	}
	if err := checkSvgSourceSize(sizeSrc); err != nil {
		return 0, 0, err
	}

	if w.svgParser == nil {
		return 0, 0, fmt.Errorf("no SVG parser configured")
	}

	var content string
	if strings.HasPrefix(svgSrc, "<") {
		content = svgSrc
	} else {
		data, err := os.ReadFile(resolvedSrc)
		if err != nil {
			return 0, 0, fmt.Errorf("SVG not found: %s", resolvedSrc)
		}
		content = string(data)
	}

	svgW, svgH, err := w.svgParser.ParseSvgDimensions(content)
	if err != nil {
		return 0, 0, err
	}

	dc := StateMap[uint64, [2]float32](w, nsSvgDimCache, capModerate)
	dc.Set(srcHash, [2]float32{svgW, svgH})
	return svgW, svgH, nil
}

// RemoveSvgFromCache removes all cached variants of an SVG.
func (w *Window) RemoveSvgFromCache(svgSrc string) {
	srcHash := hashString(svgSrc)
	prefix := svgHashHex(srcHash) + ":"

	svgCache := StateMapRead[string, *CachedSvg](w, nsSvgCache)
	if svgCache != nil {
		var keysToDelete []string
		for _, key := range svgCache.Keys() {
			if strings.HasPrefix(key, prefix) {
				keysToDelete = append(keysToDelete, key)
			}
		}
		for _, key := range keysToDelete {
			svgCache.Delete(key)
		}
	}

	dimCache := StateMapRead[uint64, [2]float32](w, nsSvgDimCache)
	if dimCache != nil {
		dimCache.Delete(srcHash)
	}
	if inv, ok := w.svgParser.(svgParserCacheInvalidator); ok {
		inv.InvalidateSvgSource(svgSrc)
	}
}

// ClearSvgCache removes all cached SVGs.
func (w *Window) ClearSvgCache() {
	svgCache := StateMapRead[string, *CachedSvg](w, nsSvgCache)
	if svgCache != nil {
		svgCache.Clear()
	}
	dimCache := StateMapRead[uint64, [2]float32](w, nsSvgDimCache)
	if dimCache != nil {
		dimCache.Clear()
	}
	if inv, ok := w.svgParser.(svgParserCacheInvalidator); ok {
		inv.ClearSvgParserCache()
	}
}

// computeTriangleBBox computes bounding box from tessellated paths.
// Returns [x, y, width, height].
func computeTriangleBBox(tpaths []TessellatedPath) [4]float32 {
	minX := float32(1e30)
	minY := float32(1e30)
	maxX := float32(-1e30)
	maxY := float32(-1e30)
	hasData := false

	for _, tp := range tpaths {
		for i := 0; i+1 < len(tp.Triangles); i += 2 {
			x := tp.Triangles[i]
			y := tp.Triangles[i+1]
			if x < minX {
				minX = x
			}
			if x > maxX {
				maxX = x
			}
			if y < minY {
				minY = y
			}
			if y > maxY {
				maxY = y
			}
			hasData = true
		}
	}

	if !hasData {
		return [4]float32{0, 0, 0, 0}
	}
	return [4]float32{minX, minY, maxX - minX, maxY - minY}
}
