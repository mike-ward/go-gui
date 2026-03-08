package sdl2

import (
	"math"

	"github.com/mike-ward/go-gui/gui"
	"github.com/veandco/go-sdl2/sdl"
)

func (b *Backend) drawSvg(r *gui.RenderCmd) {
	if len(r.Triangles) == 0 || len(r.Triangles)%6 != 0 {
		return
	}
	s := b.dpiScale
	numVerts := len(r.Triangles) / 2
	if cap(b.svgVerts) < numVerts {
		b.svgVerts = make([]sdl.Vertex, numVerts)
	}
	verts := b.svgVerts[:numVerts]
	hasVCols := len(r.VertexColors) == numVerts
	vAlpha := float32(1)
	if r.HasVertexAlpha {
		vAlpha = r.VertexAlphaScale
		if vAlpha < 0 {
			vAlpha = 0
		}
		if vAlpha > 1 {
			vAlpha = 1
		}
	}

	// Precompute rotation if needed.
	hasRot := r.RotAngle != 0
	var sinA, cosA float32
	if hasRot {
		rad := float64(r.RotAngle) * math.Pi / 180
		sinA = float32(math.Sin(rad))
		cosA = float32(math.Cos(rad))
	}

	for i := range numVerts {
		vx := r.Triangles[i*2]
		vy := r.Triangles[i*2+1]
		if hasRot {
			dx := vx - r.RotCX
			dy := vy - r.RotCY
			vx = r.RotCX + dx*cosA - dy*sinA
			vy = r.RotCY + dx*sinA + dy*cosA
		}
		verts[i].Position = sdl.FPoint{
			X: (r.X + vx*r.Scale) * s,
			Y: (r.Y + vy*r.Scale) * s,
		}
		if hasVCols {
			vc := r.VertexColors[i]
			alpha := vc.A
			if r.HasVertexAlpha {
				alpha = uint8(float32(alpha) * vAlpha)
			}
			verts[i].Color = sdl.Color{R: vc.R, G: vc.G, B: vc.B, A: vc.A}
			verts[i].Color.A = alpha
		} else {
			verts[i].Color = sdl.Color{
				R: r.Color.R, G: r.Color.G, B: r.Color.B, A: r.Color.A,
			}
		}
	}
	_ = b.renderer.RenderGeometry(nil, verts, nil)
	b.svgVerts = verts[:0]
}

// beginFilter starts rendering to a temporary texture for filter
// effects. All subsequent draw commands render to this texture
// until endFilter is called.
func (b *Backend) beginFilter(r *gui.RenderCmd) {
	outW, outH, _ := b.renderer.GetOutputSize()
	if b.filterPool == nil || b.filterPoolW != outW || b.filterPoolH != outH {
		if b.filterPool != nil {
			_ = b.filterPool.Destroy()
			b.filterPool = nil
		}
		tex, err := b.renderer.CreateTexture(
			sdl.PIXELFORMAT_RGBA8888,
			sdl.TEXTUREACCESS_TARGET,
			outW, outH,
		)
		if err != nil {
			return
		}
		_ = tex.SetBlendMode(sdl.BLENDMODE_BLEND)
		b.filterPool = tex
		b.filterPoolW = outW
		b.filterPoolH = outH
	}
	b.filterTex = b.filterPool
	b.filterBlur = r.BlurRadius * b.dpiScale
	b.filterLayers = r.Layers
	_ = b.renderer.SetRenderTarget(b.filterTex)
	_ = b.renderer.SetDrawColor(0, 0, 0, 0)
	_ = b.renderer.Clear()
}

// endFilter composites the filter texture back with a glow
// approximation: render blurred copies at reduced alpha, then
// the sharp source on top.
func (b *Backend) endFilter() {
	tex := b.filterTex
	if tex == nil {
		return
	}
	b.filterTex = nil

	_ = b.renderer.SetRenderTarget(nil)

	blur := b.filterBlur
	if blur < 1 {
		blur = 1
	}
	layers := b.filterLayers
	if layers < 1 {
		layers = 1
	}

	outW, outH, _ := b.renderer.GetOutputSize()

	// Simulate glow: render texture at increasing offsets with
	// Gaussian-like alpha falloff. 8 directions per ring for a
	// circular glow.
	steps := int(blur + 0.5)
	if steps < 1 {
		steps = 1
	}
	if steps > 16 {
		steps = 16
	}
	baseAlpha := float32(60) * float32(layers)
	diag := func(off int32) int32 {
		return off * 7 / 10 // ~0.707 approximation
	}
	for s := steps; s >= 1; s-- {
		t := float32(s) / float32(steps)
		a := baseAlpha * (1 - t*t)
		if a < 1 {
			continue
		}
		if a > 255 {
			a = 255
		}
		_ = tex.SetAlphaMod(uint8(a))
		off := int32(s)
		d := diag(off)
		offsets := [8][2]int32{
			{-off, 0}, {off, 0}, {0, -off}, {0, off},
			{-d, -d}, {d, -d}, {-d, d}, {d, d},
		}
		for _, o := range offsets {
			_ = b.renderer.Copy(tex, nil,
				&sdl.Rect{X: o[0], Y: o[1], W: outW, H: outH})
		}
	}

	// Render source graphic on top at full alpha.
	_ = tex.SetAlphaMod(255)
	_ = b.renderer.Copy(tex, nil, nil)
}
