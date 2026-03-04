package sdl2

import (
	"fmt"
	"image"
	"image/draw"
	_ "image/jpeg"
	_ "image/png"
	"log"
	"os"
	"path/filepath"
	"strings"
	"unsafe"

	"github.com/mike-ward/go-gui/gui"
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
		tex.Destroy()
		return texCacheEntry{}, err
	}
	tex.SetBlendMode(sdl.BLENDMODE_BLEND)

	return texCacheEntry{tex: tex}, nil
}

// drawImage renders an image from a file path, using a texture cache.
func (b *Backend) drawImage(r *gui.RenderCmd) {
	path := b.imagePathCache[r.Resource]
	if path == "" {
		var err error
		path, err = b.resolveValidatedImagePath(r.Resource)
		if path == "" {
			return
		}
		if err != nil {
			log.Printf("sdl2: drawImage: %v", err)
			return
		}
		b.imagePathCache[r.Resource] = path
	}
	if path == "" {
		return
	}

	entry, ok := b.texCache.get(path)
	if !ok {
		entry, err := b.loadTexture(path)
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
	defer f.Close()
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

func validatePathAllowed(path string, allowedRoots []string) error {
	for i := range allowedRoots {
		root := strings.TrimSpace(allowedRoots[i])
		if root == "" {
			continue
		}
		if pathWithinRoot(path, root) {
			return nil
		}
		rootAbs, err := filepath.Abs(root)
		if err != nil {
			continue
		}
		if pathWithinRoot(path, resolvePathWithParentFallback(rootAbs)) {
			return nil
		}
	}
	return fmt.Errorf("image path not allowed: %s", path)
}

func normalizeAllowedRoots(allowedRoots []string) []string {
	if len(allowedRoots) == 0 {
		return nil
	}
	roots := make([]string, 0, len(allowedRoots))
	for i := range allowedRoots {
		root := strings.TrimSpace(allowedRoots[i])
		if root == "" {
			continue
		}
		rootAbs, err := filepath.Abs(root)
		if err != nil {
			continue
		}
		roots = append(roots, resolvePathWithParentFallback(rootAbs))
	}
	return roots
}

func pathWithinRoot(path, root string) bool {
	rel, err := filepath.Rel(root, path)
	if err != nil {
		return false
	}
	return rel == "." || (rel != ".." &&
		!strings.HasPrefix(rel, ".."+string(filepath.Separator)))
}
