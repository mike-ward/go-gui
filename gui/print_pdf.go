package gui

import (
	"fmt"
	"math"
	"os"
	"strings"
	"time"
	"unicode"

	"github.com/go-pdf/fpdf"
	"github.com/mike-ward/go-glyph"
)

// ptToMM converts PostScript points to millimeters.
const ptToMM = 25.4 / 72.0

// renderToPDF renders a slice of RenderCmd to a PDF file at
// job.OutputPath. sourceW and sourceH are the source viewport
// dimensions in pixels (points).
func renderToPDF(renderers []RenderCmd, job PrintJob,
	sourceW, sourceH float32) error {

	pageW, pageH := PrintPageSize(job.Paper, job.Orientation)
	m := job.Margins

	printableW := (pageW - m.Left - m.Right) * ptToMM
	printableH := (pageH - m.Top - m.Bottom) * ptToMM

	if printableW <= 0 || printableH <= 0 {
		return fmt.Errorf("printable area is zero or negative")
	}

	// Scale factor: fit source viewport into printable area.
	var scale float32
	switch job.ScaleMode {
	case PrintScaleActualSize:
		// 1pt source = 1pt print (assume 72 DPI screen)
		scale = ptToMM
	default: // FitToPage
		sx := printableW / sourceW
		sy := printableH / sourceH
		scale = min(sx, sy)
	}

	orientation := "P"
	if job.Orientation == PrintLandscape {
		orientation = "L"
	}

	pdf := fpdf.NewCustom(&fpdf.InitType{
		OrientationStr: orientation,
		UnitStr:        "mm",
		Size: fpdf.SizeType{
			Wd: float64(pageW * ptToMM),
			Ht: float64(pageH * ptToMM),
		},
	})
	pdf.SetAutoPageBreak(false, 0)
	pdf.AddPage()

	// Translate UTF-8 to cp1252 for built-in PDF fonts.
	tr := pdf.UnicodeTranslatorFromDescriptor("")

	// Margin offset in mm.
	ox := float64(m.Left * ptToMM)
	oy := float64(m.Top * ptToMM)

	// Header/footer rendering.
	if job.Header.Enabled {
		renderHeaderFooter(pdf, tr, job.Header, job, pageW, m, true)
	}
	if job.Footer.Enabled {
		renderHeaderFooter(pdf, tr, job.Footer, job, pageW, m, false)
	}

	// Coordinate helpers.
	px := func(x float32) float64 { return float64(x*scale) + ox }
	py := func(y float32) float64 { return float64(y*scale) + oy }
	pw := func(w float32) float64 { return float64(w * scale) }
	ph := func(h float32) float64 { return float64(h * scale) }

	type clipEntry struct{}
	var clipStack []clipEntry

	for _, cmd := range renderers {
		switch cmd.Kind {
		case RenderRect:
			r, g, b := cmd.Color.R, cmd.Color.G, cmd.Color.B
			pdf.SetFillColor(int(r), int(g), int(b))
			setAlpha(pdf, cmd.Color)
			if cmd.Radius > 0 {
				pdf.RoundedRect(px(cmd.X), py(cmd.Y),
					pw(cmd.W), ph(cmd.H),
					float64(cmd.Radius*scale), "1234", "F")
			} else {
				pdf.Rect(px(cmd.X), py(cmd.Y),
					pw(cmd.W), ph(cmd.H), "F")
			}
			resetAlpha(pdf)

		case RenderStrokeRect:
			r, g, b := cmd.Color.R, cmd.Color.G, cmd.Color.B
			pdf.SetDrawColor(int(r), int(g), int(b))
			setAlpha(pdf, cmd.Color)
			lw := max(cmd.Thickness, 1)
			pdf.SetLineWidth(float64(lw * scale))
			if cmd.Radius > 0 {
				pdf.RoundedRect(px(cmd.X), py(cmd.Y),
					pw(cmd.W), ph(cmd.H),
					float64(cmd.Radius*scale), "1234", "D")
			} else {
				pdf.Rect(px(cmd.X), py(cmd.Y),
					pw(cmd.W), ph(cmd.H), "D")
			}
			resetAlpha(pdf)

		case RenderText:
			text := tr(stripUnprintable(cmd.Text))
			if text == "" {
				continue
			}
			tc := cmd.Color
			// Gradient text: use first stop color as fallback
			// (true gradient fill not supported in built-in PDF
			// fonts).
			if cmd.TextGradient != nil &&
				len(cmd.TextGradient.Stops) > 0 {
				gc := cmd.TextGradient.Stops[0].Color
				tc = Color{gc.R, gc.G, gc.B, gc.A, true}
			}
			r, g, b := tc.R, tc.G, tc.B
			pdf.SetTextColor(int(r), int(g), int(b))
			setAlpha(pdf, tc)
			family := pdfFontName(cmd.FontName)
			style := pdfFontStyle(cmd.TextStylePtr)
			size := float64(cmd.FontSize * scale / ptToMM)
			pdf.SetFont(family, style, size)

			// Scale font so PDF text width matches source.
			// Built-in PDF fonts have different metrics than
			// the system font used by glyph.
			if cmd.TextWidth > 0 && !strings.Contains(text, "\n") {
				wantW := float64(cmd.TextWidth) * float64(scale)
				pdfW := pdf.GetStringWidth(text)
				if pdfW > 0 && wantW > 0 {
					size *= wantW / pdfW
					pdf.SetFont(family, style, size)
				}
			}

			// Y is top of text box; fpdf expects baseline.
			// Use actual font ascent when available; fall back
			// to 75% of em for SVG text and other paths that
			// don't populate FontAscent.
			fa := cmd.FontAscent
			if fa == 0 {
				fa = cmd.FontSize * 0.75
			}
			ascent := float64(fa * scale)
			lineH := size * ptToMM64() * 1.2
			lines := strings.Split(text, "\n")
			for i, line := range lines {
				pdf.Text(px(cmd.X),
					py(cmd.Y)+ascent+float64(i)*lineH, line)
			}

			// Strikethrough — fpdf has no built-in support.
			if cmd.TextStylePtr != nil && cmd.TextStylePtr.Strikethrough {
				pdf.SetDrawColor(int(r), int(g), int(b))
				lw := max(float64(scale)*0.5, 0.1)
				pdf.SetLineWidth(lw)
				textW := pdf.GetStringWidth(text)
				for i := range lines {
					sy := py(cmd.Y) + ascent*0.65 + float64(i)*lineH
					pdf.Line(px(cmd.X), sy, px(cmd.X)+textW, sy)
				}
			}
			resetAlpha(pdf)

		case RenderLine:
			r, g, b := cmd.Color.R, cmd.Color.G, cmd.Color.B
			pdf.SetDrawColor(int(r), int(g), int(b))
			setAlpha(pdf, cmd.Color)
			lw := max(cmd.Thickness, 1)
			pdf.SetLineWidth(float64(lw * scale))
			pdf.Line(px(cmd.X), py(cmd.Y),
				px(cmd.OffsetX), py(cmd.OffsetY))
			resetAlpha(pdf)

		case RenderCircle:
			r, g, b := cmd.Color.R, cmd.Color.G, cmd.Color.B
			radius := cmd.W / 2
			cx := cmd.X + radius
			cy := cmd.Y + radius
			style := "F"
			if cmd.Fill {
				pdf.SetFillColor(int(r), int(g), int(b))
			} else {
				pdf.SetDrawColor(int(r), int(g), int(b))
				style = "D"
			}
			setAlpha(pdf, cmd.Color)
			pdf.Circle(px(cx), py(cy), pw(radius), style)
			resetAlpha(pdf)

		case RenderImage:
			path := cmd.Resource
			if path == "" {
				continue
			}
			if _, err := os.Stat(path); err != nil {
				continue
			}
			opts := fpdf.ImageOptions{ReadDpi: true}
			pdf.ImageOptions(path, px(cmd.X), py(cmd.Y),
				pw(cmd.W), ph(cmd.H), false, opts, 0, "")

		case RenderClip:
			if cmd.W == 0 && cmd.H == 0 {
				// End clip.
				if len(clipStack) > 0 {
					clipStack = clipStack[:len(clipStack)-1]
					pdf.ClipEnd()
				}
			} else {
				// Begin clip.
				pdf.ClipRect(px(cmd.X), py(cmd.Y),
					pw(cmd.W), ph(cmd.H), false)
				clipStack = append(clipStack, clipEntry{})
			}

		case RenderGradient:
			// fpdf only supports two-color linear gradients;
			// use first and last stop colors as approximation.
			if cmd.Gradient == nil || len(cmd.Gradient.Stops) == 0 {
				continue
			}
			stops := cmd.Gradient.Stops
			c1 := stops[0].Color
			c2 := stops[len(stops)-1].Color
			x := px(cmd.X)
			y := py(cmd.Y)
			w := pw(cmd.W)
			h := ph(cmd.H)
			gx1, gy1, gx2, gy2 := gradientCoords(cmd.Gradient.Direction)
			pdf.LinearGradient(x, y, w, h,
				int(c1.R), int(c1.G), int(c1.B),
				int(c2.R), int(c2.G), int(c2.B),
				gx1, gy1, gx2, gy2)

		case RenderGradientBorder:
			// fpdf has no gradient stroke; use first stop color
			// as a solid border approximation.
			if cmd.Gradient == nil || len(cmd.Gradient.Stops) == 0 {
				continue
			}
			c := cmd.Gradient.Stops[0].Color
			pdf.SetDrawColor(int(c.R), int(c.G), int(c.B))
			lw := max(cmd.Thickness, 1)
			pdf.SetLineWidth(float64(lw * scale))
			if cmd.Radius > 0 {
				pdf.RoundedRect(px(cmd.X), py(cmd.Y),
					pw(cmd.W), ph(cmd.H),
					float64(cmd.Radius*scale), "1234", "D")
			} else {
				pdf.Rect(px(cmd.X), py(cmd.Y),
					pw(cmd.W), ph(cmd.H), "D")
			}

		case RenderSvg:
			if len(cmd.Triangles) < 6 {
				continue
			}
			// Render SVG triangles as filled polygons.
			// Vertices are in local SVG space; transform with
			// cmd.X/Y offset and cmd.Scale (matching GPU backends).
			svgScale := cmd.Scale
			if svgScale == 0 {
				svgScale = 1
			}
			for i := 0; i+5 < len(cmd.Triangles); i += 6 {
				x1 := cmd.X + cmd.Triangles[i]*svgScale
				y1 := cmd.Y + cmd.Triangles[i+1]*svgScale
				x2 := cmd.X + cmd.Triangles[i+2]*svgScale
				y2 := cmd.Y + cmd.Triangles[i+3]*svgScale
				x3 := cmd.X + cmd.Triangles[i+4]*svgScale
				y3 := cmd.Y + cmd.Triangles[i+5]*svgScale

				// Use vertex color if available, else cmd color.
				var vc Color
				vi := i / 2 // vertex index
				if vi < len(cmd.VertexColors) {
					vc = cmd.VertexColors[vi]
				} else {
					vc = cmd.Color
				}
				pdf.SetFillColor(int(vc.R), int(vc.G), int(vc.B))
				pdf.SetDrawColor(int(vc.R), int(vc.G), int(vc.B))
				pdf.SetLineWidth(0.1)
				setAlpha(pdf, vc)

				pts := []fpdf.PointType{
					{X: px(x1), Y: py(y1)},
					{X: px(x2), Y: py(y2)},
					{X: px(x3), Y: py(y3)},
				}
				pdf.Polygon(pts, "FD")
				resetAlpha(pdf)
			}

		case RenderRTF:
			if cmd.LayoutPtr == nil {
				continue
			}
			layoutText := cmd.LayoutPtr.Text
			for i := range cmd.LayoutPtr.Items {
				item := &cmd.LayoutPtr.Items[i]
				if item.IsObject {
					continue
				}
				// Text lives in Layout.Text; Item references
				// a substring via StartIndex/Length.
				end := item.StartIndex + item.Length
				if end > len(layoutText) {
					continue
				}
				text := tr(stripUnprintable(
					layoutText[item.StartIndex:end]))
				if text == "" {
					continue
				}
				r, g, b := item.Color.R, item.Color.G, item.Color.B
				pdf.SetTextColor(int(r), int(g), int(b))
				if item.Color.A < 255 {
					pdf.SetAlpha(float64(item.Color.A)/255.0, "Normal")
				}
				// Derive font size from ascent (≈75% of em for
				// standard PDF fonts).
				srcSize := item.Ascent / 0.75
				size := srcSize * float64(scale) / float64(ptToMM)
				pdf.SetFont("Helvetica", "", size)

				// Scale font so PDF text width matches the
				// glyph-computed item width. Built-in PDF
				// fonts have different metrics than the
				// system font used by glyph.
				wantW := float64(item.Width) * float64(scale)
				pdfW := pdf.GetStringWidth(text)
				if pdfW > 0 && wantW > 0 {
					size *= wantW / pdfW
					pdf.SetFont("Helvetica", "", size)
				}

				ix := px(cmd.X + float32(item.X))
				iy := py(cmd.Y + float32(item.Y))
				pdf.Text(ix, iy, text)

				if item.HasUnderline {
					pdf.SetDrawColor(int(r), int(g), int(b))
					lw := max(item.UnderlineThickness*float64(scale), 0.1)
					pdf.SetLineWidth(lw)
					uy := py(cmd.Y + float32(item.Y+item.UnderlineOffset))
					pdf.Line(ix, uy,
						ix+pw(float32(item.Width)), uy)
				}
				if item.HasStrikethrough {
					pdf.SetDrawColor(int(r), int(g), int(b))
					lw := max(item.StrikethroughThickness*float64(scale), 0.1)
					pdf.SetLineWidth(lw)
					sy := py(cmd.Y + float32(item.Y+item.StrikethroughOffset))
					pdf.Line(ix, sy,
						ix+pw(float32(item.Width)), sy)
				}
				resetAlpha(pdf)
			}

		case RenderTextPath:
			tp := cmd.TextPath
			if tp == nil || cmd.TextStylePtr == nil || cmd.Text == "" {
				continue
			}
			text := tr(stripUnprintable(cmd.Text))
			if text == "" {
				continue
			}
			ts := cmd.TextStylePtr
			r, g, b := ts.Color.R, ts.Color.G, ts.Color.B
			pdf.SetTextColor(int(r), int(g), int(b))
			setAlpha(pdf, ts.Color)
			family := pdfFontName(ts.Family)
			style := pdfFontStyle(ts)
			size := float64(ts.Size * scale / ptToMM)
			pdf.SetFont(family, style, size)

			// Measure per-character advances with fpdf.
			runes := []rune(text)
			advances := make([]float64, len(runes))
			var totalAdv float64
			for i, ch := range runes {
				advances[i] = pdf.GetStringWidth(string(ch))
				totalAdv += advances[i]
			}

			// Apply text-anchor offset.
			offset := float64(tp.Offset * scale)
			if tp.Anchor == 1 {
				offset -= totalAdv / 2
			} else if tp.Anchor == 2 {
				offset -= totalAdv
			}

			// Method=stretch: scale advances to fill path.
			advScale := 1.0
			if tp.Method == 1 && totalAdv > 0 {
				remaining := float64(tp.TotalLen*scale) - offset
				if remaining > 0 {
					advScale = remaining / totalAdv
				}
			}

			// Place each character along the path.
			cumAdv := 0.0
			for i, ch := range runes {
				adv := advances[i] * advScale
				centerDist := float32((offset + cumAdv + adv/2) / float64(scale))
				pathX, pathY, angle := SamplePathAt(
					tp.Polyline, tp.Table, centerDist)
				halfAdv := float32(adv / 2 / float64(scale))
				cosA := float32(math.Cos(float64(angle)))
				sinA := float32(math.Sin(float64(angle)))
				gx := pathX + cmd.X - halfAdv*cosA
				gy := pathY + cmd.Y - halfAdv*sinA

				angleDeg := float64(angle) * 180 / math.Pi
				mx := px(gx)
				my := py(gy)
				pdf.TransformBegin()
				pdf.TransformRotate(-angleDeg, mx, my)
				pdf.Text(mx, my, string(ch))
				pdf.TransformEnd()
				cumAdv += adv
			}
			resetAlpha(pdf)

		// Unsupported kinds — skip silently.
		case RenderShadow, RenderBlur, RenderFilterBegin,
			RenderFilterEnd, RenderFilterComposite,
			RenderCustomShader,
			RenderLayout, RenderLayoutTransformed,
			RenderLayoutPlaced, RenderNone:
			continue
		}
	}

	// Close any remaining clip regions.
	for range clipStack {
		pdf.ClipEnd()
	}

	if pdf.Err() {
		return fmt.Errorf("pdf generation: %w", pdf.Error())
	}
	return pdf.OutputFileAndClose(job.OutputPath)
}

// pdfFontStyle returns an fpdf style string ("B", "I", "BI", "U",
// etc.) derived from a TextStyle. Returns "" when ts is nil.
func pdfFontStyle(ts *TextStyle) string {
	if ts == nil {
		return ""
	}
	bold := ts.Typeface == glyph.TypefaceBold ||
		ts.Typeface == glyph.TypefaceBoldItalic
	italic := ts.Typeface == glyph.TypefaceItalic ||
		ts.Typeface == glyph.TypefaceBoldItalic

	// SVG text encodes weight/style in the Pango font name
	// (e.g. "Sans Bold") rather than in Typeface.
	if !bold || !italic {
		lower := strings.ToLower(ts.Family)
		if !bold && strings.Contains(lower, "bold") {
			bold = true
		}
		if !italic && strings.Contains(lower, "italic") {
			italic = true
		}
	}

	var s string
	if bold {
		s += "B"
	}
	if italic {
		s += "I"
	}
	if ts.Underline {
		s += "U"
	}
	return s
}

// pdfFontName maps a gui font family name to a built-in PDF font.
func pdfFontName(name string) string {
	lower := strings.ToLower(name)
	switch {
	case strings.Contains(lower, "mono"),
		strings.Contains(lower, "courier"),
		strings.Contains(lower, "consol"):
		return "Courier"
	case strings.Contains(lower, "serif") &&
		!strings.Contains(lower, "sans"):
		return "Times"
	case strings.Contains(lower, "georgia"),
		strings.Contains(lower, "times"),
		strings.Contains(lower, "palatino"),
		strings.Contains(lower, "garamond"):
		return "Times"
	default:
		return "Helvetica"
	}
}

// stripUnprintable removes Private Use Area and other non-printable
// runes that built-in PDF fonts cannot render.
func stripUnprintable(s string) string {
	return strings.Map(func(r rune) rune {
		if unicode.In(r, unicode.Co) { // Co = Private Use
			return -1
		}
		return r
	}, s)
}

// ptToMM64 returns ptToMM as float64 (used for font baseline).
func ptToMM64() float64 { return float64(ptToMM) }

// setAlpha applies alpha transparency to the PDF state. Unset colors
// (Color.IsSet() == false) are treated as fully opaque.
func setAlpha(pdf *fpdf.Fpdf, c Color) {
	if !c.IsSet() {
		return
	}
	if c.A < 255 {
		pdf.SetAlpha(float64(c.A)/255.0, "Normal")
	}
}

// resetAlpha restores full opacity.
func resetAlpha(pdf *fpdf.Fpdf) {
	pdf.SetAlpha(1.0, "Normal")
}

// gradientCoords returns (x1,y1,x2,y2) for fpdf.LinearGradient
// based on gradient direction.
func gradientCoords(dir GradientDirection) (float64, float64, float64, float64) {
	switch dir {
	case GradientToRight:
		return 0, 0, 1, 0
	case GradientToBottom:
		return 0, 0, 0, 1
	case GradientToLeft:
		return 1, 0, 0, 0
	case GradientToTop:
		return 0, 1, 0, 0
	case GradientToBottomRight:
		return 0, 0, 1, 1
	case GradientToBottomLeft:
		return 1, 0, 0, 1
	case GradientToTopRight:
		return 0, 1, 1, 0
	case GradientToTopLeft:
		return 1, 1, 0, 0
	default:
		return 0, 0, 0, 1
	}
}

// renderHeaderFooter draws a header or footer line on the page.
func renderHeaderFooter(pdf *fpdf.Fpdf, tr func(string) string,
	cfg PrintHeaderFooterCfg,
	job PrintJob, pageW float32, m PrintMargins, isHeader bool) {

	fontSize := 8.0 // points
	pdf.SetFont("Helvetica", "", fontSize)
	pdf.SetTextColor(0, 0, 0)

	var yPt float32
	if isHeader {
		yPt = m.Top * 0.5 // center in top margin
	} else {
		_, pageH := PrintPageSize(job.Paper, job.Orientation)
		yPt = pageH - m.Bottom*0.5
	}
	y := float64(yPt * ptToMM)

	leftX := float64(m.Left * ptToMM)
	rightX := float64((pageW - m.Right) * ptToMM)
	centerX := float64(pageW * ptToMM / 2)

	replacer := headerFooterReplacer(job)

	if cfg.Left != "" {
		text := tr(replacer.Replace(cfg.Left))
		pdf.Text(leftX, y, text)
	}
	if cfg.Center != "" {
		text := tr(replacer.Replace(cfg.Center))
		tw := pdf.GetStringWidth(text)
		pdf.Text(centerX-tw/2, y, text)
	}
	if cfg.Right != "" {
		text := tr(replacer.Replace(cfg.Right))
		tw := pdf.GetStringWidth(text)
		pdf.Text(rightX-tw, y, text)
	}
}

// headerFooterReplacer builds a strings.Replacer for print tokens.
func headerFooterReplacer(job PrintJob) *strings.Replacer {
	return strings.NewReplacer(
		"{page}", "1",
		"{pages}", "1",
		"{date}", time.Now().Format("2006-01-02"),
		"{title}", job.Title,
		"{job}", job.JobName,
	)
}
