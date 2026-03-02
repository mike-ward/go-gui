package sdl2

import (
	"math"

	"github.com/mike-ward/go-glyph"
	"github.com/mike-ward/go-gui/gui"
	"github.com/veandco/go-sdl2/sdl"
)

// renderersDraw iterates render commands and draws them.
func (b *Backend) renderersDraw(w *gui.Window) {
	cmds := w.Renderers()
	for i := range cmds {
		r := &cmds[i]
		switch r.Kind {
		case gui.RenderClip:
			b.drawClip(r)
		case gui.RenderRect:
			b.drawRect(r)
		case gui.RenderStrokeRect:
			b.drawStrokeRect(r)
		case gui.RenderText:
			b.drawText(r)
		case gui.RenderCircle:
			b.drawCircle(r)
		case gui.RenderLine:
			b.drawLine(r)
		case gui.RenderShadow:
			b.drawShadow(r)
		case gui.RenderBlur:
			b.drawBlur(r)
		case gui.RenderGradient:
			b.drawGradient(r)
		case gui.RenderGradientBorder:
			b.drawGradientBorder(r)
		case gui.RenderImage:
			b.drawImagePlaceholder(r)
		case gui.RenderSvg:
			b.drawSvg(r)
		case gui.RenderFilterBegin, gui.RenderFilterEnd,
			gui.RenderFilterComposite,
			gui.RenderLayout, gui.RenderLayoutTransformed,
			gui.RenderLayoutPlaced,
			gui.RenderCustomShader:
			// Skip — requires GPU pipeline or unsupported in SDL2.
		default:
			// Unimplemented render kinds are silently skipped.
		}
	}
}

func (b *Backend) drawClip(r *gui.RenderCmd) {
	s := b.dpiScale
	rect := sdl.Rect{
		X: int32(r.X * s),
		Y: int32(r.Y * s),
		W: int32(r.W * s),
		H: int32(r.H * s),
	}
	b.renderer.SetClipRect(&rect)
}

func (b *Backend) drawRect(r *gui.RenderCmd) {
	if !r.Fill {
		return
	}
	s := b.dpiScale
	if r.Radius > 0 {
		b.fillRoundedRect(
			r.X*s, r.Y*s, r.W*s, r.H*s,
			r.Radius*s, r.Color)
		return
	}
	b.renderer.SetDrawColor(r.Color.R, r.Color.G, r.Color.B, r.Color.A)
	rect := sdl.FRect{
		X: r.X * s, Y: r.Y * s,
		W: r.W * s, H: r.H * s,
	}
	b.renderer.FillRectF(&rect)
}

func (b *Backend) drawStrokeRect(r *gui.RenderCmd) {
	s := b.dpiScale
	if r.Radius > 0 {
		b.strokeRoundedRect(
			r.X*s, r.Y*s, r.W*s, r.H*s,
			r.Radius*s, r.Color)
		return
	}
	b.renderer.SetDrawColor(r.Color.R, r.Color.G, r.Color.B, r.Color.A)
	rect := sdl.FRect{
		X: r.X * s, Y: r.Y * s,
		W: r.W * s, H: r.H * s,
	}
	b.renderer.DrawRectF(&rect)
}

func (b *Backend) drawText(r *gui.RenderCmd) {
	if b.textSys == nil || len(r.Text) == 0 {
		return
	}
	cfg := glyph.TextConfig{
		Style: glyph.TextStyle{
			FontName: r.FontName,
			Size:     r.FontSize,
			Color: glyph.Color{
				R: r.Color.R,
				G: r.Color.G,
				B: r.Color.B,
				A: r.Color.A,
			},
		},
		Block: glyph.DefaultBlockStyle(),
	}
	b.textSys.DrawText(r.X, r.Y, r.Text, cfg)
}

func (b *Backend) drawCircle(r *gui.RenderCmd) {
	if !r.Fill || r.Radius <= 0 {
		return
	}
	// Approximate circle as filled rect (MVP).
	s := b.dpiScale
	b.renderer.SetDrawColor(r.Color.R, r.Color.G, r.Color.B, r.Color.A)
	rect := sdl.FRect{
		X: (r.X - r.Radius) * s,
		Y: (r.Y - r.Radius) * s,
		W: r.Radius * 2 * s,
		H: r.Radius * 2 * s,
	}
	b.renderer.FillRectF(&rect)
}

// fillRoundedRect draws a filled rectangle with rounded corners
// using three rectangles and four filled quarter-circle fans.
func (b *Backend) fillRoundedRect(x, y, w, h, rad float32, c gui.Color) {
	rad = min(rad, w/2, h/2)
	b.renderer.SetDrawColor(c.R, c.G, c.B, c.A)

	// Center strip.
	b.renderer.FillRectF(&sdl.FRect{
		X: x + rad, Y: y, W: w - 2*rad, H: h,
	})
	// Left strip.
	b.renderer.FillRectF(&sdl.FRect{
		X: x, Y: y + rad, W: rad, H: h - 2*rad,
	})
	// Right strip.
	b.renderer.FillRectF(&sdl.FRect{
		X: x + w - rad, Y: y + rad, W: rad, H: h - 2*rad,
	})

	// Four corner quarter-circles (scanline fill).
	ri := int(rad)
	for dy := 0; dy <= ri; dy++ {
		dx := int(math.Sqrt(float64(rad*rad) - float64(dy*dy)))
		fy := float32(dy)
		fx := float32(dx)
		// Top-left.
		b.renderer.DrawLineF(x+rad-fx, y+rad-fy, x+rad, y+rad-fy)
		// Top-right.
		b.renderer.DrawLineF(x+w-rad, y+rad-fy, x+w-rad+fx, y+rad-fy)
		// Bottom-left.
		b.renderer.DrawLineF(x+rad-fx, y+h-rad+fy, x+rad, y+h-rad+fy)
		// Bottom-right.
		b.renderer.DrawLineF(x+w-rad, y+h-rad+fy, x+w-rad+fx, y+h-rad+fy)
	}
}

func (b *Backend) drawLine(r *gui.RenderCmd) {
	s := b.dpiScale
	b.renderer.SetDrawColor(r.Color.R, r.Color.G, r.Color.B, r.Color.A)
	b.renderer.DrawLineF(r.X*s, r.Y*s, r.OffsetX*s, r.OffsetY*s)
}

func (b *Backend) drawShadow(r *gui.RenderCmd) {
	s := b.dpiScale
	// 3 concentric offset rects at decreasing alpha.
	for i := range 3 {
		off := float32(i+1) * r.BlurRadius * 0.5 * s
		a := r.Color.A / uint8(i+2)
		x := (r.X+r.OffsetX)*s - off
		y := (r.Y+r.OffsetY)*s - off
		w := r.W*s + 2*off
		h := r.H*s + 2*off
		if r.Radius > 0 {
			c := gui.Color{R: r.Color.R, G: r.Color.G, B: r.Color.B, A: a}
			b.fillRoundedRect(x, y, w, h, r.Radius*s+off, c)
		} else {
			b.renderer.SetDrawColor(r.Color.R, r.Color.G, r.Color.B, a)
			b.renderer.FillRectF(&sdl.FRect{X: x, Y: y, W: w, H: h})
		}
	}
}

func (b *Backend) drawBlur(r *gui.RenderCmd) {
	// Placeholder: single rect at reduced alpha.
	s := b.dpiScale
	a := r.Color.A / 2
	b.renderer.SetDrawColor(r.Color.R, r.Color.G, r.Color.B, a)
	rect := sdl.FRect{
		X: r.X * s, Y: r.Y * s,
		W: r.W * s, H: r.H * s,
	}
	b.renderer.FillRectF(&rect)
}

func (b *Backend) drawGradient(r *gui.RenderCmd) {
	if r.Gradient == nil || len(r.Gradient.Stops) == 0 || r.W <= 0 || r.H <= 0 {
		return
	}
	s := b.dpiScale
	dx, dy := gui.GradientDir(r.Gradient, r.W, r.H)

	const bands = 32
	// Determine if gradient is more horizontal or vertical.
	horizontal := math.Abs(float64(dx)) >= math.Abs(float64(dy))

	for i := range bands {
		t := float32(i) / float32(bands)
		t2 := float32(i+1) / float32(bands)
		c := gui.SampleGradientStopColor(r.Gradient.Stops, (t+t2)/2)
		b.renderer.SetDrawColor(c.R, c.G, c.B, c.A)
		var rect sdl.FRect
		if horizontal {
			rect = sdl.FRect{
				X: (r.X + t*r.W) * s,
				Y: r.Y * s,
				W: (t2 - t) * r.W * s,
				H: r.H * s,
			}
		} else {
			rect = sdl.FRect{
				X: r.X * s,
				Y: (r.Y + t*r.H) * s,
				W: r.W * s,
				H: (t2 - t) * r.H * s,
			}
		}
		b.renderer.FillRectF(&rect)
	}
}

func (b *Backend) drawGradientBorder(r *gui.RenderCmd) {
	if r.Gradient == nil || len(r.Gradient.Stops) == 0 {
		return
	}
	s := b.dpiScale
	th := r.Thickness * s
	// 4 border rects with sampled colors at 0.0, 0.25, 0.5, 0.75.
	positions := [4]float32{0.0, 0.25, 0.5, 0.75}
	rects := [4]sdl.FRect{
		{X: r.X * s, Y: r.Y * s, W: r.W * s, H: th},                         // top
		{X: r.X * s, Y: (r.Y+r.H)*s - th, W: r.W * s, H: th},               // bottom
		{X: r.X * s, Y: r.Y * s, W: th, H: r.H * s},                         // left
		{X: (r.X+r.W)*s - th, Y: r.Y * s, W: th, H: r.H * s},               // right
	}
	for i := range 4 {
		c := gui.SampleGradientStopColor(r.Gradient.Stops, positions[i])
		b.renderer.SetDrawColor(c.R, c.G, c.B, c.A)
		b.renderer.FillRectF(&rects[i])
	}
}

func (b *Backend) drawImagePlaceholder(r *gui.RenderCmd) {
	// Placeholder: colored rect until texture cache is implemented.
	s := b.dpiScale
	b.renderer.SetDrawColor(200, 200, 200, 255)
	rect := sdl.FRect{
		X: r.X * s, Y: r.Y * s,
		W: r.W * s, H: r.H * s,
	}
	b.renderer.FillRectF(&rect)
}

// strokeRoundedRect draws a stroked rectangle with rounded corners
// using four lines and four quarter-circle arcs (midpoint algorithm).
func (b *Backend) strokeRoundedRect(x, y, w, h, rad float32, c gui.Color) {
	rad = min(rad, w/2, h/2)
	b.renderer.SetDrawColor(c.R, c.G, c.B, c.A)

	// Straight edges.
	b.renderer.DrawLineF(x+rad, y, x+w-rad, y)             // top
	b.renderer.DrawLineF(x+rad, y+h, x+w-rad, y+h)         // bottom
	b.renderer.DrawLineF(x, y+rad, x, y+h-rad)             // left
	b.renderer.DrawLineF(x+w, y+rad, x+w, y+h-rad)         // right

	// Quarter-circle arcs at each corner (midpoint algorithm).
	cx := [4]float32{x + rad, x + w - rad, x + w - rad, x + rad}
	cy := [4]float32{y + rad, y + rad, y + h - rad, y + h - rad}
	sx := [4]float32{-1, 1, 1, -1}
	sy := [4]float32{-1, -1, 1, 1}

	ri := int(rad)
	px, py := ri, 0
	d := 1 - ri
	for px >= py {
		fx, fy := float32(px), float32(py)
		for i := range 4 {
			b.renderer.DrawPointF(cx[i]+fx*sx[i], cy[i]+fy*sy[i])
			b.renderer.DrawPointF(cx[i]+fy*sx[i], cy[i]+fx*sy[i])
		}
		py++
		if d < 0 {
			d += 2*py + 1
		} else {
			px--
			d += 2*(py-px) + 1
		}
	}
}
