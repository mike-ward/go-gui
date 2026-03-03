// Package sdl2 provides an SDL2-based backend for go-gui.
package sdl2

import (
	"fmt"
	"log"
	"runtime"

	"github.com/mike-ward/go-glyph"
	glyphsdl "github.com/mike-ward/go-glyph/backend/sdl2"
	"github.com/veandco/go-sdl2/sdl"

	"github.com/mike-ward/go-gui/gui"
	"github.com/mike-ward/go-gui/gui/svg"
)

// Backend is the SDL2 backend for go-gui.
type Backend struct {
	window       *sdl.Window
	renderer     *sdl.Renderer
	textSys      *glyph.TextSystem
	dpiScale     float32
	cursors      [11]*sdl.Cursor
	filterTex    *sdl.Texture // temporary render target for filter groups
	filterBlur   float32      // blur radius in pixels
	filterLayers int          // number of blur layers
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

	win, err := sdl.CreateWindow(
		title,
		sdl.WINDOWPOS_CENTERED, sdl.WINDOWPOS_CENTERED,
		width, height,
		sdl.WINDOW_SHOWN|sdl.WINDOW_RESIZABLE|sdl.WINDOW_ALLOW_HIGHDPI,
	)
	if err != nil {
		sdl.Quit()
		return nil, fmt.Errorf("sdl2: CreateWindow: %w", err)
	}

	ren, err := sdl.CreateRenderer(win, -1,
		sdl.RENDERER_ACCELERATED|sdl.RENDERER_PRESENTVSYNC)
	if err != nil {
		win.Destroy()
		sdl.Quit()
		return nil, fmt.Errorf("sdl2: CreateRenderer: %w", err)
	}
	ren.SetDrawBlendMode(sdl.BLENDMODE_BLEND)

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
		ren.Destroy()
		win.Destroy()
		sdl.Quit()
		return nil, fmt.Errorf("sdl2: NewTextSystem: %w", err)
	}

	b := &Backend{
		window:   win,
		renderer: ren,
		textSys:  textSys,
		dpiScale: dpiScale,
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

	return b, nil
}

// Run starts the event loop. Blocks until quit.
func (b *Backend) Run(w *gui.Window) {
	if w.Config.OnInit != nil {
		w.Config.OnInit(w)
	}

	// Register event watcher for live resize rendering on macOS.
	// During window drag-resize, macOS enters a modal loop that
	// blocks PollEvent. This callback fires from within that loop,
	// allowing re-layout and re-render at the new size.
	watchHandle := sdl.AddEventWatchFunc(
		func(ev sdl.Event, _ interface{}) bool {
			we, ok := ev.(*sdl.WindowEvent)
			if !ok || we.Event != sdl.WINDOWEVENT_SIZE_CHANGED {
				return true
			}
			e := gui.Event{
				Type:         gui.EventResized,
				WindowWidth:  int(we.Data1),
				WindowHeight: int(we.Data2),
			}
			w.EventFn(&e)
			w.FrameFn()
			b.renderFrame(w)
			return true
		}, nil)
	defer sdl.DelEventWatch(watchHandle)

	running := true
	for running {
		for ev := sdl.PollEvent(); ev != nil; ev = sdl.PollEvent() {
			e, cont := mapEvent(ev, b)
			if !cont {
				running = false
				break
			}
			if e.Type != gui.EventInvalid {
				w.EventFn(&e)
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
	b.renderer.SetDrawColor(bg.R, bg.G, bg.B, bg.A)
	b.renderer.Clear()
	b.renderer.SetClipRect(nil)

	w.Lock()
	b.renderersDraw(w)
	w.Unlock()

	b.textSys.Commit()
	b.renderer.Present()
}

// Run initializes the SDL2 backend, runs the event loop,
// and cleans up on exit.
func Run(w *gui.Window) error {
	b, err := New(w)
	if err != nil {
		return err
	}
	defer b.Destroy()
	b.Run(w)
	return nil
}

// Destroy releases all backend resources.
func (b *Backend) Destroy() {
	if b.textSys != nil {
		b.textSys.Free()
	}
	for i, c := range b.cursors {
		if c != nil {
			sdl.FreeCursor(c)
			b.cursors[i] = nil
		}
	}
	if b.renderer != nil {
		b.renderer.Destroy()
	}
	if b.window != nil {
		b.window.Destroy()
	}
	sdl.Quit()
	log.Println("sdl2: backend destroyed")
}
