package sdl2

import (
	"github.com/mike-ward/go-gui/gui"
	"github.com/veandco/go-sdl2/sdl"
)

func (b *Backend) drawSvg(r *gui.RenderCmd) {
	if len(r.Triangles) == 0 || len(r.Triangles)%6 != 0 {
		return
	}
	s := b.dpiScale
	numVerts := len(r.Triangles) / 2
	verts := make([]sdl.Vertex, numVerts)
	hasVCols := len(r.VertexColors) == numVerts

	for i := range numVerts {
		verts[i].Position = sdl.FPoint{
			X: (r.X + r.Triangles[i*2]*r.Scale) * s,
			Y: (r.Y + r.Triangles[i*2+1]*r.Scale) * s,
		}
		if hasVCols {
			vc := r.VertexColors[i]
			verts[i].Color = sdl.Color{R: vc.R, G: vc.G, B: vc.B, A: vc.A}
		} else {
			verts[i].Color = sdl.Color{
				R: r.Color.R, G: r.Color.G, B: r.Color.B, A: r.Color.A,
			}
		}
	}
	b.renderer.RenderGeometry(nil, verts, nil)
}
