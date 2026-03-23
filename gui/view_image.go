package gui

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// ImageCfg configures an image view.
type ImageCfg struct {
	ID        string
	Src       string
	Width     float32
	Height    float32
	MinWidth  float32
	MaxWidth  float32
	MinHeight float32
	MaxHeight float32
	OnClick   func(*Layout, *Event, *Window)
	OnHover   func(*Layout, *Event, *Window)
	Invisible bool

	Opacity Opt[float32]
	BgColor Color // opaque fill drawn behind image (e.g. white for mermaid PNGs)

	// Accessibility
	A11YLabel       string
	A11YDescription string
}

// imageView implements View for image rendering.
type imageView struct {
	cfg ImageCfg
}

// Image creates a new image view. Supports local paths and remote
// http/https URLs. Remote images are cached locally.
func Image(cfg ImageCfg) View {
	if cfg.Invisible {
		return invisibleContainerView()
	}
	cfg.OnClick = leftClickOnly(cfg.OnClick)
	return &imageView{cfg: cfg}
}

func (iv *imageView) Content() []View { return nil }

func (iv *imageView) GenerateLayout(w *Window) Layout {
	c := &iv.cfg
	imagePath := c.Src
	isURL := strings.HasPrefix(c.Src, "http://") ||
		strings.HasPrefix(c.Src, "https://")
	isDataURL := strings.HasPrefix(c.Src, "data:")

	if isURL {
		hash := hashString(c.Src)
		cacheDir := filepath.Join(os.TempDir(), "gui_cache", "images")
		if err := os.MkdirAll(cacheDir, 0o755); err != nil {
			log.Printf("image: mkdir failed: %v", err)
		}
		basePath := filepath.Join(cacheDir, fmt.Sprintf("%x", hash))
		cachePath := findCachedImage(basePath)

		if cachePath != "" {
			if strings.HasSuffix(cachePath, ".svg") {
				sv := &svgView{cfg: SvgCfg{
					FileName: cachePath,
					Width:    c.Width,
					Height:   c.Height,
				}}
				return sv.GenerateLayout(w)
			}
			imagePath = cachePath
		} else {
			// Check if already downloading.
			downloads := StateMap[string, int64](
				w, nsActiveDownloads, capModerate)
			if !downloads.Contains(c.Src) {
				downloads.Set(c.Src, time.Now().Unix())
				wCtx := w.Ctx()
				maxBytes := w.Config.MaxImageBytes
				go downloadImage(wCtx, c.Src, basePath, maxBytes, w)
			}
			// Placeholder while downloading.
			width := c.Width
			if width <= 0 {
				width = 100
			}
			height := c.Height
			if height <= 0 {
				height = 100
			}
			layout := Layout{
				Shape: &Shape{
					ShapeType: ShapeRectangle,
					ID:        c.ID,
					Width:     width,
					Height:    height,
					Color:     guiTheme.ColorBackground,
					Opacity:   c.Opacity.Get(1.0),
				},
			}
			ApplyFixedSizingConstraints(layout.Shape)
			return layout
		}
	}

	// Data URLs are passed directly to the backend renderer
	// (used by WASM for embedded image assets).
	if !isDataURL {
		// Validate local path.
		if err := ValidateImagePath(imagePath); err != nil {
			log.Printf("image: %v", err)
			return errorTextLayout(c.Src, w)
		}
		// Check file exists.
		if _, err := os.Stat(imagePath); err != nil {
			log.Printf("image: %v", err)
			return errorTextLayout(c.Src, w)
		}
	}

	width := c.Width
	height := c.Height
	// Use defaults when no explicit size given.
	if width <= 0 {
		width = 100
	}
	if height <= 0 {
		height = 100
	}

	var events *EventHandlers
	if c.OnClick != nil || c.OnHover != nil {
		events = &EventHandlers{
			OnClick: c.OnClick,
			OnHover: c.OnHover,
		}
	}
	layout := Layout{
		Shape: &Shape{
			ShapeType: ShapeImage,
			ID:        c.ID,
			A11YRole:  AccessRoleImage,
			A11Y: makeA11YInfo(
				a11yLabel(c.A11YLabel, c.ID),
				c.A11YDescription,
			),
			Resource:  imagePath,
			Color:     c.BgColor,
			Opacity:   c.Opacity.Get(1.0),
			Width:     width,
			MinWidth:  c.MinWidth,
			MaxWidth:  c.MaxWidth,
			Height:    height,
			MinHeight: c.MinHeight,
			MaxHeight: c.MaxHeight,
			Events:    events,
		},
	}
	ApplyFixedSizingConstraints(layout.Shape)
	return layout
}

// errorTextLayout returns a magenta "[missing: src]" text layout.
func errorTextLayout(src string, w *Window) Layout {
	ts := guiTheme.TextStyleDef
	ts.Color = Magenta
	tv := Text(TextCfg{
		Text:      fmt.Sprintf("[missing: %s]", src),
		TextStyle: ts,
	})
	return tv.GenerateLayout(w)
}

// defaultMaxDownloadBytes is the download size limit when
// WindowCfg.MaxImageBytes is not set.
const defaultMaxDownloadBytes = int64(16 * 1024 * 1024)

// downloadImage fetches a remote image to a local cache path.
// wCtx is the window's context — cancellation stops the download.
// maxBytes is the size limit (0 uses defaultMaxDownloadBytes).
func downloadImage(
	wCtx context.Context, url, basePath string,
	maxBytes int64, w *Window,
) {
	maxSize := maxBytes
	if maxSize <= 0 {
		maxSize = defaultMaxDownloadBytes
	}

	ctx, cancel := context.WithTimeout(wCtx, 30*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodHead, url, nil)
	if err != nil {
		log.Printf("image download: %v", err)
		removeDownload(url, w)
		return
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("image download HEAD failed: %v", err)
		removeDownload(url, w)
		return
	}
	_ = resp.Body.Close()

	// Validate content length.
	if resp.ContentLength > maxSize {
		log.Printf("image too large (%d bytes): %s",
			resp.ContentLength, url)
		removeDownload(url, w)
		return
	}

	// Validate content type.
	ct := resp.Header.Get("Content-Type")
	if !strings.HasPrefix(ct, "image/") {
		log.Printf("invalid content type for image: %s", url)
		removeDownload(url, w)
		return
	}

	ext := contentTypeToExt(ct)
	path := basePath + ext

	// Download file.
	req2, err := http.NewRequestWithContext(
		ctx, http.MethodGet, url, nil)
	if err != nil {
		log.Printf("image download: %v", err)
		removeDownload(url, w)
		return
	}
	resp2, err := http.DefaultClient.Do(req2)
	if err != nil {
		log.Printf("image download GET failed: %v", err)
		removeDownload(url, w)
		return
	}
	defer func() { _ = resp2.Body.Close() }()

	f, err := os.Create(path)
	if err != nil {
		log.Printf("image download create: %v", err)
		removeDownload(url, w)
		return
	}
	written, err := io.Copy(f, io.LimitReader(resp2.Body, maxSize))
	_ = f.Close()
	if err != nil {
		_ = os.Remove(path)
		log.Printf("image download write: %v", err)
		removeDownload(url, w)
		return
	}
	if written >= maxSize {
		_ = os.Remove(path)
		log.Printf("image download body exceeds limit: %s", url)
		removeDownload(url, w)
		return
	}

	// Success: remove from active downloads and refresh.
	w.QueueCommand(func(w *Window) {
		dl := StateMap[string, int64](
			w, nsActiveDownloads, capModerate)
		dl.Delete(url)
		w.UpdateWindow()
	})
}

// removeDownload removes a URL from the active downloads map.
func removeDownload(url string, w *Window) {
	w.QueueCommand(func(w *Window) {
		dl := StateMap[string, int64](
			w, nsActiveDownloads, capModerate)
		dl.Delete(url)
	})
}

// contentTypeToExt maps Content-Type to file extension.
func contentTypeToExt(ct string) string {
	switch {
	case strings.HasPrefix(ct, "image/svg+xml"):
		return ".svg"
	case strings.HasPrefix(ct, "image/png"):
		return ".png"
	case strings.HasPrefix(ct, "image/jpeg"):
		return ".jpg"
	default:
		return ".png"
	}
}

// findCachedImage searches for a cached image file with any valid
// extension matching the given base path.
func findCachedImage(basePath string) string {
	for _, ext := range []string{
		".png", ".jpg", ".jpeg", ".svg",
	} {
		candidate := basePath + ext
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
	}
	return ""
}
