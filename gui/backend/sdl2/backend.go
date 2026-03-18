//go:build !js

// Package sdl2 provides an SDL2-based backend for go-gui.
package sdl2

import (
	"fmt"
	"log"
	"os"
	"runtime"

	"github.com/mike-ward/go-glyph"
	glyphsdl "github.com/mike-ward/go-glyph/backend/sdl2"
	"github.com/veandco/go-sdl2/sdl"

	"github.com/mike-ward/go-gui/gui"
	"github.com/mike-ward/go-gui/gui/backend/internal/imgpath"
	"github.com/mike-ward/go-gui/gui/backend/internal/tempfont"
	"github.com/mike-ward/go-gui/gui/backend/internal/texcache"
	"github.com/mike-ward/go-gui/gui/svg"
)

// Backend is the SDL2 backend for go-gui.
type Backend struct {
	window             *sdl.Window
	renderer           *sdl.Renderer
	textSys            *glyph.TextSystem
	dpiScale           float32
	cursors            [11]*sdl.Cursor
	filterTex          *sdl.Texture // temporary render target for filter groups
	filterPrevTarget   *sdl.Texture
	filterPool         *sdl.Texture // reusable filter render target
	filterPoolW        int32
	filterPoolH        int32
	filterBlur         float32      // blur radius in pixels
	filterLayers       int          // number of blur layers
	filterColorMatrix  *[16]float32 // color transform matrix
	filterPixels       []uint32     // reusable pixel buffer for color matrix
	svgVerts           []sdl.Vertex // reusable vertex buffer for SVG geometry
	textPathPlacements []glyph.GlyphPlacement
	texCache           texcache.Cache[string, *sdl.Texture]
	iconFontPath       string
	allowedImageRoots  []string
	imagePathCache     texcache.Cache[string, string]
	roundedClipStack   []roundedClipState
	maxImageBytes      int64
	maxImagePixels     int64
	normBuf            []gui.GradientStop // reusable buffer for gradient normalization
	sampledBuf         []gui.GradientStop // reusable buffer for downsampled stops
}

// New creates an SDL2 backend and initializes the window.
func New(w *gui.Window) (*Backend, error) {
	runtime.LockOSThread()

	if err := sdl.Init(sdl.INIT_VIDEO | sdl.INIT_EVENTS); err != nil {
		return nil, fmt.Errorf("sdl2: Init: %w", err)
	}

	cfg := w.Config
	title := cfg.Title
	if title == "" {
		title = "go-gui"
	}
	width := int32(cfg.Width)
	if width <= 0 {
		width = 640
	}
	height := int32(cfg.Height)
	if height <= 0 {
		height = 480
	}
	flags := uint32(sdl.WINDOW_SHOWN | sdl.WINDOW_ALLOW_HIGHDPI)
	if !cfg.FixedSize {
		flags |= sdl.WINDOW_RESIZABLE
	}

	win, err := sdl.CreateWindow(
		title,
		sdl.WINDOWPOS_CENTERED, sdl.WINDOWPOS_CENTERED,
		width, height,
		flags,
	)
	if err != nil {
		sdl.Quit()
		return nil, fmt.Errorf("sdl2: CreateWindow: %w", err)
	}

	ren, err := sdl.CreateRenderer(win, -1,
		sdl.RENDERER_ACCELERATED|sdl.RENDERER_PRESENTVSYNC)
	if err != nil {
		_ = win.Destroy()
		sdl.Quit()
		return nil, fmt.Errorf("sdl2: CreateRenderer: %w", err)
	}
	_ = ren.SetDrawBlendMode(sdl.BLENDMODE_BLEND)

	// Compute DPI scale.
	outW, _, _ := ren.GetOutputSize()
	winW, _ := win.GetSize()
	dpiScale := float32(1.0)
	if winW > 0 {
		dpiScale = float32(outW) / float32(winW)
	}

	// Initialize glyph text system.
	glyphBack := glyphsdl.New(ren, dpiScale)
	textSys, err := glyph.NewTextSystem(glyphBack)
	if err != nil {
		_ = ren.Destroy()
		_ = win.Destroy()
		sdl.Quit()
		return nil, fmt.Errorf("sdl2: NewTextSystem: %w", err)
	}

	// Load embedded icon font into glyph via temp file.
	var iconFontPath string
	if data := gui.IconFontData; len(data) > 0 {
		tmp, err := tempfont.Write("go_gui_feathericon", data)
		if err != nil {
			log.Printf("sdl2: write icon font: %v", err)
		} else if err := textSys.AddFontFile(tmp); err != nil {
			log.Printf("sdl2: load icon font: %v", err)
			_ = os.Remove(tmp)
		} else {
			iconFontPath = tmp
		}
	}

	b := &Backend{
		window:            win,
		renderer:          ren,
		textSys:           textSys,
		dpiScale:          dpiScale,
		texCache:          newSDLTexCache(128),
		iconFontPath:      iconFontPath,
		allowedImageRoots: imgpath.NormalizeRoots(cfg.AllowedImageRoots),
		imagePathCache:    texcache.New[string, string](1024, nil),
		maxImageBytes:     cfg.MaxImageBytes,
		maxImagePixels:    cfg.MaxImagePixels,
	}

	// Create system cursors.
	b.cursors[gui.CursorDefault] = sdl.CreateSystemCursor(sdl.SYSTEM_CURSOR_ARROW)
	b.cursors[gui.CursorArrow] = sdl.CreateSystemCursor(sdl.SYSTEM_CURSOR_ARROW)
	b.cursors[gui.CursorIBeam] = sdl.CreateSystemCursor(sdl.SYSTEM_CURSOR_IBEAM)
	b.cursors[gui.CursorCrosshair] = sdl.CreateSystemCursor(sdl.SYSTEM_CURSOR_CROSSHAIR)
	b.cursors[gui.CursorPointingHand] = sdl.CreateSystemCursor(sdl.SYSTEM_CURSOR_HAND)
	b.cursors[gui.CursorResizeEW] = sdl.CreateSystemCursor(sdl.SYSTEM_CURSOR_SIZEWE)
	b.cursors[gui.CursorResizeNS] = sdl.CreateSystemCursor(sdl.SYSTEM_CURSOR_SIZENS)
	b.cursors[gui.CursorResizeNWSE] = sdl.CreateSystemCursor(sdl.SYSTEM_CURSOR_SIZENWSE)
	b.cursors[gui.CursorResizeNESW] = sdl.CreateSystemCursor(sdl.SYSTEM_CURSOR_SIZENESW)
	b.cursors[gui.CursorResizeAll] = sdl.CreateSystemCursor(sdl.SYSTEM_CURSOR_SIZEALL)
	b.cursors[gui.CursorNotAllowed] = sdl.CreateSystemCursor(sdl.SYSTEM_CURSOR_NO)

	// Set text measurer on gui Window.
	w.SetTextMeasurer(&textMeasurer{textSys: textSys})

	// Set SVG parser on gui Window.
	w.SetSvgParser(svg.New())

	// Set clipboard functions.
	w.SetClipboardFn(func(text string) {
		if err := sdl.SetClipboardText(text); err != nil {
			log.Printf("sdl2: set clipboard: %v", err)
		}
	})
	w.SetClipboardGetFn(func() string {
		text, _ := sdl.GetClipboardText()
		return text
	})

	// Set native platform.
	w.SetNativePlatform(&nativePlatform{})

	return b, nil
}

// Run starts the event loop. Blocks until quit.
func (b *Backend) Run(w *gui.Window) {
	defer w.WindowCleanup()
	if w.Config.OnInit != nil {
		w.Config.OnInit(w)
	}

	// Register event watcher for live resize rendering on macOS.
	// During window drag-resize, macOS enters a modal loop that
	// blocks PollEvent. This callback fires from within that loop,
	// allowing re-layout and re-render at the new size.
	var watchHandle sdl.EventWatchHandle
	if runtime.GOOS == "darwin" {
		resizeEvent := &gui.Event{Type: gui.EventResized}
		watchHandle = sdl.AddEventWatchFunc(
			func(ev sdl.Event, _ interface{}) bool {
				we, ok := ev.(*sdl.WindowEvent)
				if !ok || we.Event != sdl.WINDOWEVENT_SIZE_CHANGED {
					return true
				}
				resizeEvent.WindowWidth = int(we.Data1)
				resizeEvent.WindowHeight = int(we.Data2)
				w.EventFn(resizeEvent)
				w.FrameFn()
				b.renderFrame(w)
				return true
			}, nil)
		defer sdl.DelEventWatch(watchHandle)
	}

	running := true
	evt := new(gui.Event)
	for running {
		for ev := sdl.PollEvent(); ev != nil; ev = sdl.PollEvent() {
			mapped, cont := mapEvent(ev, b)
			*evt = mapped
			if !cont {
				running = false
				break
			}
			if evt.Type != gui.EventInvalid {
				w.EventFn(evt)
			}
		}
		if !running {
			break
		}

		w.FrameFn()
		b.renderFrame(w)

		// Update cursor.
		mc := w.MouseCursorState()
		if int(mc) < len(b.cursors) && b.cursors[mc] != nil {
			sdl.SetCursor(b.cursors[mc])
		}
	}
}

// renderFrame clears the screen, draws the current layout, and presents.
func (b *Backend) renderFrame(w *gui.Window) {
	bg := w.Config.BgColor
	if bg == (gui.Color{}) {
		t := gui.CurrentTheme()
		bg = t.ColorBackground
	}
	_ = b.renderer.SetDrawColor(bg.R, bg.G, bg.B, bg.A)
	_ = b.renderer.Clear()
	_ = b.renderer.SetClipRect(nil)

	w.Lock()
	b.renderersDraw(w)
	w.Unlock()

	b.textSys.Commit()
	b.renderer.Present()
}

// Run initializes the SDL2 backend, runs the event loop,
// and cleans up on exit.
func Run(w *gui.Window) {
	b, err := New(w)
	if err != nil {
		panic(err)
	}
	defer b.Destroy()
	b.Run(w)
}

// Destroy releases all backend resources.
func (b *Backend) Destroy() {
	b.texCache.DestroyAll()
	if b.filterPool != nil {
		_ = b.filterPool.Destroy()
		b.filterPool = nil
	}
	if b.textSys != nil {
		b.textSys.Free()
	}
	if b.iconFontPath != "" {
		_ = os.Remove(b.iconFontPath)
		b.iconFontPath = ""
	}
	for i, c := range b.cursors {
		if c != nil {
			sdl.FreeCursor(c)
			b.cursors[i] = nil
		}
	}
	if b.renderer != nil {
		_ = b.renderer.Destroy()
	}
	if b.window != nil {
		_ = b.window.Destroy()
	}
	sdl.Quit()
}
