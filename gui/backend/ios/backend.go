//go:build ios

// Package ios provides an iOS backend for go-gui using Metal
// rendering and UIKit for windowing/events.
package ios

/*
#cgo CFLAGS: -fobjc-arc
#cgo LDFLAGS: -framework Metal -framework QuartzCore -framework Foundation -framework UIKit
#include "metal_darwin.h"
#include "ios_app.h"
*/
import "C"

import (
	"fmt"
	"log"
	"os"
	"runtime"
	"unsafe"

	"github.com/mike-ward/go-glyph"

	"github.com/mike-ward/go-gui/gui"
	"github.com/mike-ward/go-gui/gui/backend/internal/imgpath"
	"github.com/mike-ward/go-gui/gui/backend/internal/tempfont"
	"github.com/mike-ward/go-gui/gui/backend/internal/texcache"
	"github.com/mike-ward/go-gui/gui/svg"
)

// Pipeline IDs matching the C enum.
const (
	pipeSolid       = C.PIPE_SOLID
	pipeShadow      = C.PIPE_SHADOW
	pipeBlur        = C.PIPE_BLUR
	pipeGradient    = C.PIPE_GRADIENT
	pipeImageClip   = C.PIPE_IMAGE_CLIP
	pipeGlyphTex    = C.PIPE_GLYPH_TEX
	pipeGlyphColor  = C.PIPE_GLYPH_COLOR
)

const maxCustomPipelines = 32

// Package-level singleton (iOS has exactly one window).
var (
	iosBackend *Backend
	iosWindow  *gui.Window
)

// Backend is the iOS Metal backend for go-gui.
type Backend struct {
	textSys  *glyph.TextSystem
	dpiScale float32
	physW    int32
	physH    int32
	mvp      [16]float32

	mvpStack [][16]float32

	// Reusable buffers.
	svgVerts           []vertex
	textPathPlacements []glyph.GlyphPlacement
	normBuf            []gui.GradientStop
	sampledBuf         []gui.GradientStop

	textures          texcache.Cache[string, metalTexture]
	glyphBack         *metalGlyphBackend
	filterBlur        float32
	filterLayer       int
	filterColorMatrix *[16]float32
	customCache       texcache.Cache[uint64, C.int]
	iconFontPath      string

	allowedImageRoots []string
	imagePathCache    texcache.Cache[string, string]
	maxImageBytes     int64
	maxImagePixels    int64
}

// --- Pattern A: Go-driven (backend.Run) ---

// Run creates the UIKit app from Go. Calls UIApplicationMain
// which never returns.
func Run(w *gui.Window) {
	runtime.LockOSThread()
	iosWindow = w
	C.iosStartApp()
}

// --- Pattern B: Swift-driven (c-archive) ---

// SetWindow sets the gui.Window for the c-archive pattern.
// Must be called from an init() function before GoGuiStart.
func SetWindow(w *gui.Window) { iosWindow = w }

// (Pattern B initialization is via the Start function below.)

// --- Shared initialization ---

func initBackend(layerPtr unsafe.Pointer,
	w, h int32, scale float32) {

	rc := C.metalInit(layerPtr)
	if rc != 0 {
		panic(fmt.Sprintf("ios: metalInit failed: %d", rc))
	}

	physW := int32(float32(w) * scale)
	physH := int32(float32(h) * scale)
	C.metalResize(C.int(physW), C.int(physH))

	cfg := iosWindow.Config
	b := &Backend{
		dpiScale: scale,
		physW:    physW,
		physH:    physH,
		textures: newMetalTexCacheLRU(128),
		customCache: texcache.New[uint64, C.int](
			maxCustomPipelines,
			func(idx C.int) { C.metalDeleteCustomPipeline(idx) },
		),
		imagePathCache: texcache.New[string, string](1024, nil),
		maxImageBytes:  cfg.MaxImageBytes,
		maxImagePixels: cfg.MaxImagePixels,
	}
	b.allowedImageRoots = imgpath.NormalizeRoots(
		cfg.AllowedImageRoots)
	b.updateProjection()

	// Initialize glyph text system with Metal backend.
	b.glyphBack = newMetalGlyphBackend(scale)
	textSys, err := glyph.NewTextSystem(b.glyphBack)
	if err != nil {
		panic(fmt.Sprintf("ios: NewTextSystem: %v", err))
	}
	b.textSys = textSys

	// Load embedded icon font.
	if data := gui.IconFontData; len(data) > 0 {
		tmp, err := tempfont.Write("go_gui_feathericon", data)
		if err != nil {
			log.Printf("ios: write icon font: %v", err)
		} else if err = textSys.AddFontFile(tmp); err != nil {
			log.Printf("ios: load icon font: %v", err)
			_ = os.Remove(tmp)
		} else {
			b.iconFontPath = tmp
		}
	}

	// Set injected interfaces on gui Window.
	iosWindow.SetTextMeasurer(&textMeasurer{textSys: textSys})
	iosWindow.SetSvgParser(svg.New())
	iosWindow.SetClipboardFn(func(_ string) {})
	iosWindow.SetClipboardGetFn(func() string { return "" })
	iosWindow.SetNativePlatform(&nativePlatform{})

	iosBackend = b

	// Fire initial resize so w.WindowSize() returns the
	// correct dimensions when the view function runs.
	evt := gui.Event{
		Type:         gui.EventResized,
		WindowWidth:  int(w),
		WindowHeight: int(h),
	}
	iosWindow.EventFn(&evt)

	if iosWindow.Config.OnInit != nil {
		iosWindow.Config.OnInit(iosWindow)
	}
}

// renderFrame clears the screen, draws the current layout, and
// presents the Metal drawable.
func (b *Backend) renderFrame(w *gui.Window) {
	bg := w.Config.BgColor
	if bg == (gui.Color{}) {
		t := gui.CurrentTheme()
		bg = t.ColorBackground
	}

	rc := C.metalBeginFrame(
		C.float(float32(bg.R)/255.0),
		C.float(float32(bg.G)/255.0),
		C.float(float32(bg.B)/255.0),
		C.float(float32(bg.A)/255.0),
	)
	if rc != 0 {
		return
	}

	C.metalSetPipeline(C.int(pipeSolid))
	C.metalSetMVP((*C.float)(&b.mvp[0]))

	w.Lock()
	b.renderersDraw(w)
	w.Unlock()

	// Render queued text.
	b.useGlyphPipeline()
	b.textSys.Commit()

	C.metalEndFrame()
}

func (b *Backend) handleResize(w, h int32, scale float32) {
	b.dpiScale = scale
	b.physW = int32(float32(w) * scale)
	b.physH = int32(float32(h) * scale)
	C.metalResize(C.int(b.physW), C.int(b.physH))
	b.updateProjection()
}

func (b *Backend) updateProjection() {
	ortho(&b.mvp,
		0, float32(b.physW),
		float32(b.physH), 0,
		-1, 1)
}

// Destroy releases all backend resources.
func (b *Backend) Destroy() {
	b.textures.DestroyAll()
	b.customCache.DestroyAll()
	if b.glyphBack != nil {
		b.glyphBack.destroy()
	}
	if b.textSys != nil {
		b.textSys.Free()
	}
	if b.iconFontPath != "" {
		_ = os.Remove(b.iconFontPath)
		b.iconFontPath = ""
	}
	C.metalDestroy()
}

// useGlyphPipeline sets up Metal state for glyph text rendering.
func (b *Backend) useGlyphPipeline() {
	C.metalSetPipeline(C.int(pipeGlyphTex))
	C.metalSetMVP((*C.float)(&b.mvp[0]))
}

// --- Exported callbacks for ios_app.m (Pattern A) ---

//export goIOSInit
func goIOSInit(layerPtr unsafe.Pointer,
	w, h C.int, scale C.float) {
	initBackend(layerPtr, int32(w), int32(h), float32(scale))
}

//export goIOSRender
func goIOSRender() {
	if iosBackend == nil || iosWindow == nil {
		return
	}
	iosWindow.FrameFn()
	iosBackend.renderFrame(iosWindow)
}

//export goIOSResize
func goIOSResize(w, h C.int, scale C.float) {
	if iosBackend == nil {
		return
	}
	iosBackend.handleResize(int32(w), int32(h), float32(scale))
	if iosWindow != nil {
		evt := gui.Event{
			Type:         gui.EventResized,
			WindowWidth:  int(w),
			WindowHeight: int(h),
		}
		iosWindow.EventFn(&evt)
	}
}

//export goIOSTouchBegan
func goIOSTouchBegan(x, y C.float) {
	touchDown(float32(x), float32(y))
}

//export goIOSTouchMoved
func goIOSTouchMoved(x, y C.float) {
	touchMoved(float32(x), float32(y))
}

//export goIOSTouchEnded
func goIOSTouchEnded(x, y C.float) {
	touchUp(float32(x), float32(y))
}

// --- Public API for Swift host (Pattern B) ---
// These are regular Go functions that c-archive apps can call
// or re-export under their own names.

// Start initializes the backend with a pre-existing Metal
// layer. SetWindow should be called before this.
func Start(layerPtr unsafe.Pointer, w, h int, scale float32) {
	if iosWindow == nil {
		iosWindow = gui.NewWindow(gui.WindowCfg{})
	}
	initBackend(layerPtr, int32(w), int32(h), scale)
}

// Render runs one frame: layout + draw + present.
func Render() {
	if iosBackend == nil || iosWindow == nil {
		return
	}
	iosWindow.FrameFn()
	iosBackend.renderFrame(iosWindow)
}

// TouchBegan maps a touch-down to EventMouseDown.
func TouchBegan(x, y float32) { touchDown(x, y) }

// TouchMoved maps a touch-move to EventMouseMove.
func TouchMoved(x, y float32) { touchMoved(x, y) }

// TouchEnded maps a touch-up to EventMouseUp.
func TouchEnded(x, y float32) { touchUp(x, y) }

// Resize updates the viewport after a layout change.
func Resize(w, h int, scale float32) {
	if iosBackend == nil {
		return
	}
	iosBackend.handleResize(int32(w), int32(h), scale)
	if iosWindow != nil {
		evt := gui.Event{
			Type:         gui.EventResized,
			WindowWidth:  w,
			WindowHeight: h,
		}
		iosWindow.EventFn(&evt)
	}
}

// CleanUp releases all backend resources.
func CleanUp() {
	if iosBackend != nil {
		iosBackend.Destroy()
		iosBackend = nil
	}
	if iosWindow != nil {
		iosWindow.WindowCleanup()
		iosWindow = nil
	}
}
