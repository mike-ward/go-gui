package sdl2

import (
	"github.com/mike-ward/go-glyph"
	"github.com/veandco/go-sdl2/sdl"
	"github.com/mike-ward/go-gui/gui"
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
	b.renderer.SetDrawColor(r.Color.R, r.Color.G, r.Color.B, r.Color.A)
	rect := sdl.FRect{
		X: r.X * s, Y: r.Y * s,
		W: r.W * s, H: r.H * s,
	}
	b.renderer.FillRectF(&rect)
}

func (b *Backend) drawStrokeRect(r *gui.RenderCmd) {
	s := b.dpiScale
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
