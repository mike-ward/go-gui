package sdl2

import (
	"image"
	"image/draw"
	_ "image/jpeg"
	_ "image/png"
	"log"
	"os"
	"unsafe"

	"github.com/mike-ward/go-gui/gui"
	"github.com/veandco/go-sdl2/sdl"
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
	return e, ok
}

func (c *texCache) set(path string, entry texCacheEntry) {
	if _, exists := c.data[path]; exists {
		c.data[path] = entry
		return
	}
	if len(c.order) >= c.maxSize {
		evict := c.order[0]
		c.order = c.order[1:]
		if old, ok := c.data[evict]; ok {
			if old.tex != nil {
				old.tex.Destroy()
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
			e.tex.Destroy()
		}
	}
	c.data = nil
	c.order = nil
}

// loadTexture decodes an image file and uploads it as an SDL texture.
func (b *Backend) loadTexture(path string) (texCacheEntry, error) {
	f, err := os.Open(path)
	if err != nil {
		return texCacheEntry{}, err
	}
	defer f.Close()

	src, _, err := image.Decode(f)
	if err != nil {
		return texCacheEntry{}, err
	}

	// Convert to NRGBA for consistent pixel layout.
	bounds := src.Bounds()
	nrgba := image.NewNRGBA(bounds)
	draw.Draw(nrgba, bounds, src, bounds.Min, draw.Src)

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
		tex.Destroy()
		return texCacheEntry{}, err
	}
	tex.SetBlendMode(sdl.BLENDMODE_BLEND)

	return texCacheEntry{tex: tex}, nil
}

// drawImage renders an image from a file path, using a texture cache.
func (b *Backend) drawImage(r *gui.RenderCmd) {
	path := r.Resource
	if path == "" {
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

	// TODO: ClipRadius (rounded corners) requires
	// render-to-texture stencil work; not yet implemented.

	s := b.dpiScale
	dst := sdl.Rect{
		X: int32(r.X * s),
		Y: int32(r.Y * s),
		W: int32(r.W * s),
		H: int32(r.H * s),
	}

	// Fill background when BgColor has alpha (e.g. mermaid/math PNGs).
	if r.Color.A > 0 {
		b.renderer.SetDrawColor(
			r.Color.R, r.Color.G, r.Color.B, r.Color.A)
		b.renderer.FillRect(&dst)
	}

	b.renderer.Copy(entry.tex, nil, &dst)
}
