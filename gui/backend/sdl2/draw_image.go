package sdl2

import (
	"fmt"
	"image"
	"image/draw"
	_ "image/jpeg" // Register JPEG decoder.
	_ "image/png"  // Register PNG decoder.
	"log"
	"math"
	"os"
	"path/filepath"
	"strings"
	"unsafe"

	"github.com/mike-ward/go-gui/gui"
	"github.com/mike-ward/go-gui/gui/backend/internal/imgpath"
	"github.com/veandco/go-sdl2/sdl"
)

const (
	defaultMaxImageBytes  = int64(16 * 1024 * 1024)
	defaultMaxImagePixels = int64(40_000_000)
)

// texCacheEntry holds a cached SDL texture. A nil tex indicates a
// previously failed load (negative cache).
type texCacheEntry struct {
	tex *sdl.Texture
}

// texCache is a bounded FIFO texture cache. On eviction the SDL
// texture is destroyed to free GPU memory.
type texCache struct {
	data    map[string]texCacheEntry
	order   []string
	maxSize int
}

func newTexCache(maxSize int) texCache {
	return texCache{
		data:    make(map[string]texCacheEntry, maxSize),
		order:   make([]string, 0, maxSize),
		maxSize: maxSize,
	}
}

func (c *texCache) get(path string) (texCacheEntry, bool) {
	e, ok := c.data[path]
	if ok {
		c.promote(path)
	}
	return e, ok
}

// promote moves path to the end of the order slice (most
// recently used).
func (c *texCache) promote(path string) {
	for i, k := range c.order {
		if k == path {
			c.order = append(
				c.order[:i], c.order[i+1:]...)
			c.order = append(c.order, path)
			return
		}
	}
}

func (c *texCache) set(path string, entry texCacheEntry) {
	if _, exists := c.data[path]; exists {
		c.data[path] = entry
		return
	}
	if len(c.order) >= c.maxSize {
		evict := c.order[0]
		c.order = c.order[1:]
		if len(c.order) < cap(c.order)/2 {
			c.order = append([]string(nil), c.order...)
		}
		if old, ok := c.data[evict]; ok {
			if old.tex != nil {
				_ = old.tex.Destroy()
			}
			delete(c.data, evict)
		}
	}
	c.order = append(c.order, path)
	c.data[path] = entry
}

func (c *texCache) destroyAll() {
	for _, e := range c.data {
		if e.tex != nil {
			_ = e.tex.Destroy()
		}
	}
	c.data = nil
	c.order = nil
}

// loadTextureFromFile validates image metadata, decodes pixels, and uploads to
// an SDL texture.
func (b *Backend) loadTextureFromFile(path string, f *os.File) (texCacheEntry, error) {
	if err := b.validateImageFile(path, f); err != nil {
		return texCacheEntry{}, err
	}
	if _, err := f.Seek(0, 0); err != nil {
		return texCacheEntry{}, err
	}

	src, _, err := image.Decode(f)
	if err != nil {
		return texCacheEntry{}, err
	}

	// Fast path for already-compatible pixel layout.
	var nrgba *image.NRGBA
	if existing, ok := src.(*image.NRGBA); ok {
		nrgba = existing
	} else {
		// Convert to NRGBA for consistent pixel layout.
		bounds := src.Bounds()
		nrgba = image.NewNRGBA(bounds)
		draw.Draw(nrgba, bounds, src, bounds.Min, draw.Src)
	}

	bounds := nrgba.Bounds()

	w := int32(bounds.Dx())
	h := int32(bounds.Dy())

	// Go NRGBA stores [R,G,B,A] per pixel. On little-endian
	// this maps to SDL's PIXELFORMAT_ABGR8888.
	tex, err := b.renderer.CreateTexture(
		sdl.PIXELFORMAT_ABGR8888,
		sdl.TEXTUREACCESS_STATIC,
		w, h,
	)
	if err != nil {
		return texCacheEntry{}, err
	}

	if err := tex.Update(nil, unsafe.Pointer(&nrgba.Pix[0]), nrgba.Stride); err != nil {
		_ = tex.Destroy()
		return texCacheEntry{}, err
	}
	_ = tex.SetBlendMode(sdl.BLENDMODE_BLEND)

	return texCacheEntry{tex: tex}, nil
}

// drawImage renders an image from a file path, using a texture cache.
func (b *Backend) drawImage(r *gui.RenderCmd) {
	path := b.imagePathCache[r.Resource]
	if path == "" {
		var err error
		path, err = b.resolveValidatedImagePath(r.Resource)
		if len(b.imagePathCache) >= 1024 {
			clear(b.imagePathCache)
		}
		if err != nil {
			log.Printf("sdl2: drawImage: %s: %v",
				r.Resource, err)
			b.imagePathCache[r.Resource] = "-"
			return
		}
		b.imagePathCache[r.Resource] = path
	}
	if path == "-" {
		return
	}

	entry, ok := b.texCache.get(path)
	if !ok {
		var err error
		entry, err = b.loadTexture(path)
		if err != nil {
			log.Printf("sdl2: drawImage: %v", err)
			// Cache nil sentinel to avoid repeated load attempts.
			entry = texCacheEntry{}
		}
		b.texCache.set(path, entry)
	}
	if entry.tex == nil {
		return
	}

	s := b.dpiScale
	dst := sdl.Rect{
		X: int32(r.X * s),
		Y: int32(r.Y * s),
		W: int32(r.W * s),
		H: int32(r.H * s),
	}

	// Fill background when BgColor has alpha (e.g. mermaid/math PNGs).
	if r.Color.A > 0 {
		_ = b.renderer.SetDrawColor(
			r.Color.R, r.Color.G, r.Color.B, r.Color.A)
		_ = b.renderer.FillRect(&dst)
	}

	if r.ClipRadius > 0 {
		b.drawImageRounded(entry.tex, r)
		return
	}
	_ = b.renderer.Copy(entry.tex, nil, &dst)
}

// loadTexture resolves races between path validation and open by checking path
// stability before and after open, then decoding from the same file handle.
func (b *Backend) loadTexture(path string) (texCacheEntry, error) {
	if len(b.allowedImageRoots) > 0 {
		if err := validatePathAllowed(path, b.allowedImageRoots); err != nil {
			return texCacheEntry{}, err
		}
	}
	pre, err := os.Stat(path)
	if err != nil {
		return texCacheEntry{}, err
	}
	f, err := os.Open(path)
	if err != nil {
		return texCacheEntry{}, err
	}
	defer func() { _ = f.Close() }()
	post, err := f.Stat()
	if err != nil {
		return texCacheEntry{}, err
	}
	if !os.SameFile(pre, post) {
		return texCacheEntry{}, fmt.Errorf("image path changed during open: %s", path)
	}
	return b.loadTextureFromFile(path, f)
}

func (b *Backend) resolveValidatedImagePath(src string) (string, error) {
	if strings.ContainsRune(src, 0) {
		return "", fmt.Errorf("invalid image path: contains NUL")
	}
	cleanPath := filepath.Clean(src)
	if cleanPath == "." || cleanPath == "" {
		return "", fmt.Errorf("invalid image path")
	}
	pathAbs, err := filepath.Abs(cleanPath)
	if err != nil {
		return "", fmt.Errorf("invalid image path: %w", err)
	}
	resolvedPath := resolvePathWithParentFallback(pathAbs)
	if len(b.allowedImageRoots) > 0 {
		if err := validatePathAllowed(resolvedPath, b.allowedImageRoots); err != nil {
			return "", err
		}
	}
	return resolvedPath, nil
}

func (b *Backend) validateImageFile(path string, f *os.File) error {
	maxBytes := b.maxImageBytes
	if maxBytes <= 0 {
		maxBytes = defaultMaxImageBytes
	}
	info, err := f.Stat()
	if err != nil {
		return err
	}
	if info.Size() > maxBytes {
		return fmt.Errorf("image file too large: %s", path)
	}

	maxPixels := b.maxImagePixels
	if maxPixels <= 0 {
		maxPixels = defaultMaxImagePixels
	}
	cfg, _, err := image.DecodeConfig(f)
	if err != nil {
		return err
	}
	if cfg.Width <= 0 || cfg.Height <= 0 {
		return fmt.Errorf("invalid image dimensions: %s", path)
	}
	if int64(cfg.Width)*int64(cfg.Height) > maxPixels {
		return fmt.Errorf("image dimensions too large: %s", path)
	}
	return nil
}

// Delegating wrappers — shared implementation in imgpath.

func resolvePathWithParentFallback(path string) string {
	return imgpath.ResolveWithParentFallback(path)
}

func validatePathAllowed(path string, allowedRoots []string) error {
	return imgpath.ValidateAllowed(path, allowedRoots)
}

func normalizeAllowedRoots(allowedRoots []string) []string {
	return imgpath.NormalizeRoots(allowedRoots)
}

// drawImageRounded renders a textured rounded rectangle using
// RenderGeometry with UV-mapped vertices. Mesh: 3 rectangles
// (center, left, right strips) + 4 quarter-circle corner fans.
func (b *Backend) drawImageRounded(tex *sdl.Texture, r *gui.RenderCmd) {
	s := b.dpiScale
	x := r.X * s
	y := r.Y * s
	w := r.W * s
	h := r.H * s
	rad := r.ClipRadius * s
	rad = min(rad, w/2, h/2)

	white := sdl.Color{R: 255, G: 255, B: 255, A: 255}
	vert := func(px, py float32) sdl.Vertex {
		return sdl.Vertex{
			Position: sdl.FPoint{X: px, Y: py},
			Color:    white,
			TexCoord: sdl.FPoint{
				X: (px - x) / w,
				Y: (py - y) / h,
			},
		}
	}

	const segments = 8
	// 3 quads (6 verts each) + 4 corners (segments*3 verts each).
	numVerts := 18 + 4*segments*3
	if cap(b.svgVerts) < numVerts {
		b.svgVerts = make([]sdl.Vertex, 0, numVerts)
	}
	verts := b.svgVerts[:0]

	quad := func(x0, y0, x1, y1 float32) {
		verts = append(verts,
			vert(x0, y0), vert(x1, y0), vert(x1, y1),
			vert(x0, y0), vert(x1, y1), vert(x0, y1),
		)
	}

	// Center strip.
	quad(x+rad, y, x+w-rad, y+h)
	// Left strip.
	quad(x, y+rad, x+rad, y+h-rad)
	// Right strip.
	quad(x+w-rad, y+rad, x+w, y+h-rad)

	// Corner fans. Start angles (screen coords, Y down):
	// top-left=π, top-right=3π/2, bottom-right=0, bottom-left=π/2.
	type corner struct {
		cx, cy float32
		start  float64
	}
	corners := [4]corner{
		{x + rad, y + rad, math.Pi},
		{x + w - rad, y + rad, 3 * math.Pi / 2},
		{x + w - rad, y + h - rad, 0},
		{x + rad, y + h - rad, math.Pi / 2},
	}
	step := (math.Pi / 2) / float64(segments)
	for _, c := range corners {
		for j := range segments {
			a0 := c.start + float64(j)*step
			a1 := c.start + float64(j+1)*step
			verts = append(verts,
				vert(c.cx, c.cy),
				vert(c.cx+rad*float32(math.Cos(a0)),
					c.cy+rad*float32(math.Sin(a0))),
				vert(c.cx+rad*float32(math.Cos(a1)),
					c.cy+rad*float32(math.Sin(a1))),
			)
		}
	}

	_ = b.renderer.RenderGeometry(tex, verts, nil)
	b.svgVerts = verts[:0]
}
