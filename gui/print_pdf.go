package gui

import (
	"fmt"
	"os"
	"strings"
	"time"
	"unicode"

	"github.com/go-pdf/fpdf"
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

	// Margin offset in mm.
	ox := float64(m.Left * ptToMM)
	oy := float64(m.Top * ptToMM)

	// Header/footer rendering.
	if job.Header.Enabled {
		renderHeaderFooter(pdf, job.Header, job, pageW, m, true)
	}
	if job.Footer.Enabled {
		renderHeaderFooter(pdf, job.Footer, job, pageW, m, false)
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
			text := stripUnprintable(cmd.Text)
			if text == "" {
				continue
			}
			r, g, b := cmd.Color.R, cmd.Color.G, cmd.Color.B
			pdf.SetTextColor(int(r), int(g), int(b))
			setAlpha(pdf, cmd.Color)
			family := pdfFontName(cmd.FontName)
			size := float64(cmd.FontSize * scale / ptToMM)
			pdf.SetFont(family, "", size)
			// Y is top of text box; fpdf expects baseline.
			// Ascent ≈ 75% of em for standard PDF fonts.
			ascent := size * ptToMM64() * 0.75
			lineH := size * ptToMM64() * 1.2
			lines := strings.Split(text, "\n")
			for i, line := range lines {
				pdf.Text(px(cmd.X),
					py(cmd.Y)+ascent+float64(i)*lineH, line)
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
			for i := 0; i+5 < len(cmd.Triangles); i += 6 {
				x1, y1 := cmd.Triangles[i], cmd.Triangles[i+1]
				x2, y2 := cmd.Triangles[i+2], cmd.Triangles[i+3]
				x3, y3 := cmd.Triangles[i+4], cmd.Triangles[i+5]

				// Use vertex color if available, else cmd color.
				var cr, cg, cb uint8
				vi := i / 2 // vertex index
				if vi < len(cmd.VertexColors) {
					vc := cmd.VertexColors[vi]
					cr, cg, cb = vc.R, vc.G, vc.B
				} else {
					cr, cg, cb = cmd.Color.R, cmd.Color.G, cmd.Color.B
				}
				pdf.SetFillColor(int(cr), int(cg), int(cb))

				pts := []fpdf.PointType{
					{X: px(x1), Y: py(y1)},
					{X: px(x2), Y: py(y2)},
					{X: px(x3), Y: py(y3)},
				}
				pdf.Polygon(pts, "F")
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
				text := stripUnprintable(
					layoutText[item.StartIndex:end])
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

		// Unsupported kinds — skip silently.
		case RenderShadow, RenderBlur, RenderFilterBegin,
			RenderFilterEnd, RenderFilterComposite,
			RenderCustomShader, RenderTextPath,
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
func renderHeaderFooter(pdf *fpdf.Fpdf, cfg PrintHeaderFooterCfg,
	job PrintJob, pageW float32, m PrintMargins, isHeader bool) {

	fontSize := 8.0 // points
	pdf.SetFont("Helvetica", "", fontSize)
	pdf.SetTextColor(0, 0, 0)

	yPt := m.Top * 0.5 // center in top margin
	if !isHeader {
		yPt = pageW - m.Bottom*0.5 // approximate
		_, pageH := PrintPageSize(job.Paper, job.Orientation)
		yPt = pageH - m.Bottom*0.5
	}
	y := float64(yPt * ptToMM)

	leftX := float64(m.Left * ptToMM)
	rightX := float64((pageW - m.Right) * ptToMM)
	centerX := float64(pageW * ptToMM / 2)

	replacer := headerFooterReplacer(job)

	if cfg.Left != "" {
		text := replacer.Replace(cfg.Left)
		pdf.Text(leftX, y, text)
	}
	if cfg.Center != "" {
		text := replacer.Replace(cfg.Center)
		tw := pdf.GetStringWidth(text)
		pdf.Text(centerX-tw/2, y, text)
	}
	if cfg.Right != "" {
		text := replacer.Replace(cfg.Right)
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
