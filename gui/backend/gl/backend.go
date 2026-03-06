// Package gl provides an OpenGL 3.3 backend for go-gui.
package gl

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sync"

	"github.com/go-gl/gl/v3.3-core/gl"
	"github.com/mike-ward/go-glyph"
	"github.com/veandco/go-sdl2/sdl"

	"github.com/mike-ward/go-gui/gui"
	"github.com/mike-ward/go-gui/gui/svg"
)

// Backend is the OpenGL 3.3 backend for go-gui.
type Backend struct {
	window   *sdl.Window
	glCtx    sdl.GLContext
	textSys  *glyph.TextSystem
	dpiScale float32
	physW    int32
	physH    int32
	cursors  [11]*sdl.Cursor

	pipelines pipelineSet
	quadVAO   uint32
	quadVBO   uint32
	quadIBO   uint32
	mvp       [16]float32

	// Reusable buffers.
	svgVAO             uint32
	svgVBO             uint32
	svgCap             int
	textPathPlacements []glyph.GlyphPlacement
	svgVerts           []vertex
	normBuf            []gui.GradientStop
	sampledBuf         []gui.GradientStop

	textures    glTexCache
	filterFBO   uint32
	filterTexA  uint32
	filterTexB  uint32
	filterW     int32
	filterH     int32
	filterBlur  float32
	filterLayer int

	glyphBack  *glyphBackend
	customOnce sync.Once

	allowedImageRoots []string
	imagePathCache    map[string]string
	maxImageBytes     int64
	maxImagePixels    int64
}

// New creates an OpenGL 3.3 backend and initializes the window.
func New(w *gui.Window) (*Backend, error) {
	runtime.LockOSThread()

	if err := sdl.Init(sdl.INIT_VIDEO | sdl.INIT_EVENTS); err != nil {
		return nil, fmt.Errorf("gl: Init: %w", err)
	}

	// Request OpenGL 3.3 core profile.
	sdl.GLSetAttribute(sdl.GL_CONTEXT_MAJOR_VERSION, 3)
	sdl.GLSetAttribute(sdl.GL_CONTEXT_MINOR_VERSION, 3)
	sdl.GLSetAttribute(sdl.GL_CONTEXT_PROFILE_MASK,
		sdl.GL_CONTEXT_PROFILE_CORE)
	sdl.GLSetAttribute(sdl.GL_CONTEXT_FLAGS,
		sdl.GL_CONTEXT_FORWARD_COMPATIBLE_FLAG)
	sdl.GLSetAttribute(sdl.GL_DOUBLEBUFFER, 1)
	sdl.GLSetAttribute(sdl.GL_STENCIL_SIZE, 8)

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
	flags := uint32(sdl.WINDOW_SHOWN | sdl.WINDOW_ALLOW_HIGHDPI | sdl.WINDOW_OPENGL)
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
		return nil, fmt.Errorf("gl: CreateWindow: %w", err)
	}

	glCtx, err := win.GLCreateContext()
	if err != nil {
		win.Destroy()
		sdl.Quit()
		return nil, fmt.Errorf("gl: GLCreateContext: %w", err)
	}

	if err := gl.Init(); err != nil {
		sdl.GLDeleteContext(glCtx)
		win.Destroy()
		sdl.Quit()
		return nil, fmt.Errorf("gl: gl.Init: %w", err)
	}

	// Enable vsync.
	sdl.GLSetSwapInterval(1)

	// Compute DPI scale.
	glW, glH := win.GLGetDrawableSize()
	winW, _ := win.GetSize()
	dpiScale := float32(1.0)
	if winW > 0 {
		dpiScale = float32(glW) / float32(winW)
	}

	b := &Backend{
		window:         win,
		glCtx:          glCtx,
		dpiScale:       dpiScale,
		physW:          glW,
		physH:          glH,
		textures:       newGLTexCache(128),
		imagePathCache: make(map[string]string, 64),
		maxImageBytes:  cfg.MaxImageBytes,
		maxImagePixels: cfg.MaxImagePixels,
	}
	b.allowedImageRoots = normalizeAllowedRoots(cfg.AllowedImageRoots)

	// Initialize GL state.
	gl.Enable(gl.BLEND)
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
	gl.Disable(gl.DEPTH_TEST)
	gl.Disable(gl.CULL_FACE)
	gl.Viewport(0, 0, glW, glH)

	// Compile shader pipelines.
	if err := b.initPipelines(); err != nil {
		sdl.GLDeleteContext(glCtx)
		win.Destroy()
		sdl.Quit()
		return nil, fmt.Errorf("gl: initPipelines: %w", err)
	}

	// Initialize quad buffers.
	b.initQuadBuffers()
	b.initSvgBuffers()

	// Build ortho projection.
	b.updateProjection()

	// Initialize glyph text system with GL backend.
	b.glyphBack = newGlyphBackend(dpiScale)
	textSys, err := glyph.NewTextSystem(b.glyphBack)
	if err != nil {
		b.Destroy()
		return nil, fmt.Errorf("gl: NewTextSystem: %w", err)
	}
	b.textSys = textSys

	// Load embedded icon font. File must persist because
	// FontConfig registers the path; FreeType reads it lazily.
	if data := gui.IconFontData; len(data) > 0 {
		tmp := filepath.Join(os.TempDir(),
			"go_gui_feathericon.ttf")
		if err := os.WriteFile(tmp, data, 0o644); err != nil {
			log.Printf("gl: write icon font: %v", err)
		} else if err := textSys.AddFontFile(tmp); err != nil {
			log.Printf("gl: load icon font: %v", err)
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
	w.SetClipboardFn(func(text string) { sdl.SetClipboardText(text) })
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
	var watchHandle sdl.EventWatchHandle
	if runtime.GOOS == "darwin" {
		watchHandle = sdl.AddEventWatchFunc(
			func(ev sdl.Event, _ interface{}) bool {
				we, ok := ev.(*sdl.WindowEvent)
				if !ok || we.Event != sdl.WINDOWEVENT_SIZE_CHANGED {
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

		mc := w.MouseCursorState()
		if int(mc) < len(b.cursors) && b.cursors[mc] != nil {
			sdl.SetCursor(b.cursors[mc])
		}
	}
}

// renderFrame clears the screen, draws the current layout, and
// swaps buffers.
func (b *Backend) renderFrame(w *gui.Window) {
	bg := w.Config.BgColor
	if bg == (gui.Color{}) {
		t := gui.CurrentTheme()
		bg = t.ColorBackground
	}
	gl.ClearColor(
		float32(bg.R)/255.0,
		float32(bg.G)/255.0,
		float32(bg.B)/255.0,
		float32(bg.A)/255.0,
	)
	gl.Disable(gl.SCISSOR_TEST)
	gl.Clear(gl.COLOR_BUFFER_BIT)

	w.Lock()
	b.renderersDraw(w)
	w.Unlock()

	b.textSys.Commit()
	b.window.GLSwap()
}

func (b *Backend) handleResize() {
	glW, glH := b.window.GLGetDrawableSize()
	b.physW = glW
	b.physH = glH
	gl.Viewport(0, 0, glW, glH)
	b.updateProjection()
}

func (b *Backend) updateProjection() {
	ortho(&b.mvp,
		0, float32(b.physW),
		float32(b.physH), 0,
		-1, 1)
}

// Run initializes the GL backend, runs the event loop, and cleans
// up on exit.
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
	b.destroyPipelines()
	if b.quadVAO != 0 {
		gl.DeleteVertexArrays(1, &b.quadVAO)
	}
	if b.quadVBO != 0 {
		gl.DeleteBuffers(1, &b.quadVBO)
	}
	if b.quadIBO != 0 {
		gl.DeleteBuffers(1, &b.quadIBO)
	}
	if b.svgVAO != 0 {
		gl.DeleteVertexArrays(1, &b.svgVAO)
	}
	if b.svgVBO != 0 {
		gl.DeleteBuffers(1, &b.svgVBO)
	}
	b.destroyFilterFBO()
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
	if b.glCtx != nil {
		sdl.GLDeleteContext(b.glCtx)
	}
	if b.window != nil {
		b.window.Destroy()
	}
	sdl.Quit()
	log.Println("gl: backend destroyed")
}

// ortho builds an orthographic projection matrix (column-major).
func ortho(m *[16]float32, l, r, b, t, n, f float32) {
	m[0] = 2 / (r - l)
	m[1] = 0
	m[2] = 0
	m[3] = 0

	m[4] = 0
	m[5] = 2 / (t - b)
	m[6] = 0
	m[7] = 0

	m[8] = 0
	m[9] = 0
	m[10] = -2 / (f - n)
	m[11] = 0

	m[12] = -(r + l) / (r - l)
	m[13] = -(t + b) / (t - b)
	m[14] = -(f + n) / (f - n)
	m[15] = 1
}
