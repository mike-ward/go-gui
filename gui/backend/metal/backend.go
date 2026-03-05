//go:build darwin

// Package metal provides a Metal backend for go-gui on macOS.
// Uses SDL2 for window creation and event handling, and Metal
// for GPU rendering. Eliminates the macOS compositor content
// shift during window resize via CAMetalLayer's
// presentsWithTransaction.
package metal

/*
#cgo CFLAGS: -fobjc-arc
#cgo pkg-config: sdl2
#cgo LDFLAGS: -framework Metal -framework QuartzCore -framework Foundation
#include "metal_darwin.h"
#include <SDL.h>
#include <SDL_metal.h>
*/
import "C"

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"unsafe"

	"github.com/mike-ward/go-glyph"
	"github.com/veandco/go-sdl2/sdl"

	"github.com/mike-ward/go-gui/gui"
	"github.com/mike-ward/go-gui/gui/svg"
)

// Pipeline IDs matching the C enum.
const (
	pipeSolid       = C.PIPE_SOLID
	pipeShadow      = C.PIPE_SHADOW
	pipeBlur        = C.PIPE_BLUR
	pipeGradient    = C.PIPE_GRADIENT
	pipeImageClip   = C.PIPE_IMAGE_CLIP
	pipeFilterBlurH = C.PIPE_FILTER_BLUR_H
	pipeFilterBlurV = C.PIPE_FILTER_BLUR_V
	pipeFilterTex   = C.PIPE_FILTER_TEX
	pipeFilterColor = C.PIPE_FILTER_COLOR
	pipeGlyphTex    = C.PIPE_GLYPH_TEX
	pipeGlyphColor  = C.PIPE_GLYPH_COLOR
)

// Backend is the Metal backend for go-gui.
type Backend struct {
	window    *sdl.Window
	metalView unsafe.Pointer
	textSys   *glyph.TextSystem
	dpiScale  float32
	physW     int32
	physH     int32
	cursors   [11]*sdl.Cursor
	mvp       [16]float32

	// Reusable buffers.
	svgVerts           []vertex
	textPathPlacements []glyph.GlyphPlacement
	normBuf            []gui.GradientStop
	sampledBuf         []gui.GradientStop

	textures    metalTexCache
	glyphBack    *metalGlyphBackend
	filterBlur   float32
	filterLayer  int
	customCache  map[uint64]C.int

	allowedImageRoots []string
	imagePathCache    map[string]string
	maxImageBytes     int64
	maxImagePixels    int64
}

// New creates a Metal backend and initializes the window.
func New(w *gui.Window) (*Backend, error) {
	runtime.LockOSThread()

	if err := sdl.Init(sdl.INIT_VIDEO | sdl.INIT_EVENTS); err != nil {
		return nil, fmt.Errorf("metal: Init: %w", err)
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

	const windowMetal = 0x20000000 // SDL_WINDOW_METAL
	win, err := sdl.CreateWindow(
		title,
		sdl.WINDOWPOS_CENTERED, sdl.WINDOWPOS_CENTERED,
		width, height,
		sdl.WINDOW_SHOWN|sdl.WINDOW_RESIZABLE|
			sdl.WINDOW_ALLOW_HIGHDPI|windowMetal,
	)
	if err != nil {
		sdl.Quit()
		return nil, fmt.Errorf("metal: CreateWindow: %w", err)
	}

	// Get the CAMetalLayer from SDL via C calls.
	cWin := (*C.SDL_Window)(unsafe.Pointer(win))
	metalView := unsafe.Pointer(
		C.SDL_Metal_CreateView(cWin))
	if metalView == nil {
		_ = win.Destroy()
		sdl.Quit()
		return nil, fmt.Errorf("metal: MetalCreateView failed")
	}
	layer := C.SDL_Metal_GetLayer(
		C.SDL_MetalView(metalView))
	if layer == nil {
		C.SDL_Metal_DestroyView(
			C.SDL_MetalView(metalView))
		_ = win.Destroy()
		sdl.Quit()
		return nil, fmt.Errorf("metal: MetalGetLayer failed")
	}

	// Initialize Metal device, pipelines, etc.
	rc := C.metalInit(layer)
	if rc != 0 {
		C.SDL_Metal_DestroyView(
			C.SDL_MetalView(metalView))
		_ = win.Destroy()
		sdl.Quit()
		return nil, fmt.Errorf("metal: init failed: %d", rc)
	}

	// Compute DPI scale.
	var dw, dh C.int
	C.SDL_Metal_GetDrawableSize(cWin, &dw, &dh)
	winW, _ := win.GetSize()
	dpiScale := float32(1.0)
	if winW > 0 {
		dpiScale = float32(dw) / float32(winW)
	}

	C.metalResize(dw, dh)

	b := &Backend{
		window:         win,
		metalView:      metalView,
		dpiScale:       dpiScale,
		physW:          int32(dw),
		physH:          int32(dh),
		textures:       newMetalTexCache(128),
		customCache:    make(map[uint64]C.int),
		imagePathCache: make(map[string]string, 64),
		maxImageBytes:  cfg.MaxImageBytes,
		maxImagePixels: cfg.MaxImagePixels,
	}
	b.allowedImageRoots = normalizeAllowedRoots(
		cfg.AllowedImageRoots)
	b.updateProjection()

	// Initialize glyph text system with Metal backend.
	b.glyphBack = newMetalGlyphBackend(dpiScale)
	textSys, err := glyph.NewTextSystem(b.glyphBack)
	if err != nil {
		b.Destroy()
		return nil, fmt.Errorf("metal: NewTextSystem: %w", err)
	}
	b.textSys = textSys

	// Load embedded icon font. File must persist because
	// FontConfig registers the path; FreeType reads it lazily.
	if data := gui.IconFontData; len(data) > 0 {
		tmp := filepath.Join(os.TempDir(),
			"go_gui_feathericon.ttf")
		if err := os.WriteFile(tmp, data, 0o644); err != nil {
			log.Printf("metal: write icon font: %v", err)
		} else if err := textSys.AddFontFile(tmp); err != nil {
			log.Printf("metal: load icon font: %v", err)
		}
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

	// Set injected interfaces on gui Window.
	w.SetTextMeasurer(&textMeasurer{textSys: textSys})
	w.SetSvgParser(svg.New())
	w.SetClipboardFn(func(text string) {
		if err := sdl.SetClipboardText(text); err != nil {
			log.Printf("metal: set clipboard: %v", err)
		}
	})
	w.SetClipboardGetFn(func() string {
		text, _ := sdl.GetClipboardText()
		return text
	})
	w.SetNativePlatform(&nativePlatform{})

	return b, nil
}

// Run starts the event loop. Blocks until quit.
func (b *Backend) Run(w *gui.Window) {
	defer w.WindowCleanup()
	if w.Config.OnInit != nil {
		w.Config.OnInit(w)
	}

	// Register event watcher for live resize on macOS.
	// During window drag-resize, macOS enters a modal loop that
	// blocks PollEvent. This callback fires from within that
	// loop, allowing re-layout and re-render at the new size.
	resizeEvent := &gui.Event{Type: gui.EventResized}
	watchHandle := sdl.AddEventWatchFunc(
		func(ev sdl.Event, _ interface{}) bool {
			we, ok := ev.(*sdl.WindowEvent)
			if !ok ||
				we.Event != sdl.WINDOWEVENT_SIZE_CHANGED {
				return true
			}
			b.handleResize()
			resizeEvent.WindowWidth = int(we.Data1)
			resizeEvent.WindowHeight = int(we.Data2)
			w.EventFn(resizeEvent)
			w.FrameFn()
			b.renderFrame(w)
			return true
		}, nil)
	defer sdl.DelEventWatch(watchHandle)

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

		mc := w.MouseCursorState()
		if int(mc) < len(b.cursors) && b.cursors[mc] != nil {
			sdl.SetCursor(b.cursors[mc])
		}
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

	// Set initial pipeline state.
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

func (b *Backend) handleResize() {
	var dw, dh C.int
	cWin := (*C.SDL_Window)(unsafe.Pointer(b.window))
	C.SDL_Metal_GetDrawableSize(cWin, &dw, &dh)
	b.physW = int32(dw)
	b.physH = int32(dh)
	C.metalResize(dw, dh)
	b.updateProjection()
}

func (b *Backend) updateProjection() {
	ortho(&b.mvp,
		0, float32(b.physW),
		float32(b.physH), 0,
		-1, 1)
}

// Run initializes the Metal backend, runs the event loop, and
// cleans up on exit.
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
	b.textures.destroyAll()
	if b.glyphBack != nil {
		b.glyphBack.destroy()
	}
	if b.textSys != nil {
		b.textSys.Free()
	}
	for i, c := range b.cursors {
		if c != nil {
			sdl.FreeCursor(c)
			b.cursors[i] = nil
		}
	}

	C.metalDestroy()

	if b.metalView != nil {
		C.SDL_Metal_DestroyView(
			C.SDL_MetalView(b.metalView))
	}
	if b.window != nil {
		_ = b.window.Destroy()
	}
	sdl.Quit()
	log.Println("metal: backend destroyed")
}

// useGlyphPipeline sets up Metal state for glyph text rendering.
func (b *Backend) useGlyphPipeline() {
	C.metalSetPipeline(C.int(pipeGlyphTex))
	C.metalSetMVP((*C.float)(&b.mvp[0]))
}

