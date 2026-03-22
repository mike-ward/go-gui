package gui

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/mike-ward/go-glyph"
)

// TextMeasurer measures text dimensions. Set by the backend
// after initialization; nil in tests (placeholder fallback).
type TextMeasurer interface {
	TextWidth(text string, style TextStyle) float32
	TextHeight(text string, style TextStyle) float32
	FontHeight(style TextStyle) float32
	FontAscent(style TextStyle) float32
	// LayoutText uses wrapWidth > 0 for wrap-enabled block width and
	// wrapWidth < 0 for width-constrained no-wrap alignment/layout.
	LayoutText(text string, style TextStyle, wrapWidth float32) (glyph.Layout, error)
}

// Window is the main application window.
type Window struct {
	// Mutexes.
	mu         sync.Mutex // guards layout/renderer state
	commandsMu sync.Mutex // guards command queue

	// Multi-window: parent App and SDL window ID.
	app        *App
	platformID uint32
	closeReq   atomic.Bool

	// User state — accessed via State[T](w).
	state any

	// View state.
	viewState ViewState

	// Command queue — flushed at frame start.
	commands []queuedCommand
	// Scratch queue used to avoid reallocating command storage each frame.
	commandScratch []queuedCommand

	// Layout tree — current frame.
	layout Layout

	// Renderers — flat draw command list, reused via [:0].
	renderers []RenderCmd

	// Clip radius propagated during render walk.
	clipRadius float32
	// Stencil depth for nested ClipContents.
	stencilDepth uint8
	// Nesting guard for filter brackets.
	inFilter bool

	// Per-frame pipeline timings.
	frameTimings FrameTimings

	// Refresh flags.
	refreshLayout     bool
	refreshRenderOnly bool

	// Window dimensions (logical pixels).
	windowWidth  int
	windowHeight int

	// Render guard — warnings emitted once per kind.
	renderGuardWarned map[string]bool

	// Active animations keyed by ID.
	animations map[string]Animation

	// Dialog state.
	dialogCfg DialogCfg

	// Inspector state.
	inspectorEnabled    bool
	inspectorTreeCache  []TreeNodeCfg
	inspectorPropsCache map[string]inspectorNodeProps

	// Toast state.
	toasts       []toastNotification
	toastCounter uint64

	// Window focus state — backend sets false on unfocus event.
	focused bool

	// OnEvent is called for unhandled events. Nil-safe.
	OnEvent func(*Event, *Window)

	// Command registry — registered commands for shortcut
	// dispatch, menu/button integration.
	cmdRegistry []Command

	// Text measurement — set by backend, nil in tests.
	textMeasurer TextMeasurer

	// SVG parser — set by backend, nil in tests.
	svgParser SvgParser

	// Native platform — set by backend, nil in tests.
	nativePlatform NativePlatform

	// File access / security-scoped bookmarks.
	fileAccess fileAccessState

	// Accessibility backend state.
	a11y a11y

	// Clipboard — set by backend, nil in tests.
	clipboardSetFn func(string)
	clipboardGetFn func() string

	// Input Method Editor state.
	ime ime

	// View generator — produces the root View each frame.
	viewGenerator func(*Window) View

	// Config stores the WindowCfg for backend access.
	Config WindowCfg

	// Reusable per-frame scratch buffers.
	scratch scratchPools

	// Lifecycle context — cancelled in WindowCleanup to abort
	// in-flight async goroutines (HTTP fetches, notifications, etc.).
	ctx       context.Context
	cancelCtx context.CancelFunc

	// Animation loop lifecycle.
	animationStop     chan struct{}
	animationDone     chan struct{}
	animationResumeCh chan struct{} // buffered(1), resumes ticker
	animationStopOnce sync.Once
	cleanupOnce       sync.Once

	// wakeMainFn pushes an SDL user event to wake the main
	// thread from WaitEventTimeout. Set by backend; nil-safe.
	wakeMainFn func()
}

// MouseLockCfg stores callbacks for mouse event handling in a
// locked state (drag operations). When mouse is locked, these
// callbacks intercept normal mouse event processing.
type MouseLockCfg struct {
	CursorPos int
	MouseDown func(*Layout, *Event, *Window)
	MouseMove func(*Layout, *Event, *Window)
	MouseUp   func(*Layout, *Event, *Window)
}

// ViewState holds per-window UI state.
type ViewState struct {
	registry      StateRegistry
	idFocus       uint32
	mouseCursor   MouseCursor
	mouseLock     MouseLockCfg
	inputCursorOn bool
	mousePosX     float32
	mousePosY     float32
	menuKeyNav    bool
	tooltip       tooltipState

	// Markdown caches (lazy-init: nil until first use).
	markdownTheme            string
	markdownCache            *BoundedMap[int64, []MarkdownBlock]
	diagramCache             *BoundedDiagramCache
	diagramRequestSeq        uint64
	externalAPIWarningLogged bool

	// RTF layout cache — avoids re-shaping unchanged content.
	rtfLayoutCache *BoundedMap[uint64, rtfLayoutEntry]
	rtfLayoutTheme string
}

// State returns a typed pointer to the user-supplied state.
func State[T any](w *Window) *T {
	return w.state.(*T)
}

// SetState sets the user state for the window.
func (w *Window) SetState(state any) {
	w.state = state
}

// Ctx returns the window's lifecycle context. The context is
// cancelled when WindowCleanup runs. Use for async operations
// that should abort on window destruction.
func (w *Window) Ctx() context.Context {
	if w.ctx == nil {
		return context.Background()
	}
	return w.ctx
}

// ClearViewState resets all view state.
func (w *Window) ClearViewState() {
	w.viewState.registry.Clear()
	w.viewState.idFocus = 0
}

// Lock locks the window's mutex.
func (w *Window) Lock() {
	w.mu.Lock()
}

// Unlock unlocks the window's mutex.
func (w *Window) Unlock() {
	w.mu.Unlock()
}

// WindowSize returns cached window dimensions.
func (w *Window) WindowSize() (int, int) {
	return w.windowWidth, w.windowHeight
}

// WindowRect returns the window as a DrawClip.
func (w *Window) WindowRect() DrawClip {
	return DrawClip{
		X: 0, Y: 0,
		Width:  float32(w.windowWidth),
		Height: float32(w.windowHeight),
	}
}

// RenderersCount returns the number of active renderers.
func (w *Window) RenderersCount() int {
	return len(w.renderers)
}

// IDFocus returns the current focus id.
func (w *Window) IDFocus() uint32 {
	return w.viewState.idFocus
}

// SetIDFocus sets the focus id and clears input selections.
func (w *Window) SetIDFocus(id uint32) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.setIDFocusLocked(id)
}

func (w *Window) setIDFocusLocked(id uint32) {
	prev := w.viewState.idFocus
	w.clearInputSelections()
	w.imeClear()
	w.viewState.idFocus = id
	if id > 0 {
		w.viewState.inputCursorOn = true
		if !w.hasAnimationLocked(blinkCursorAnimationID) {
			w.animationAdd(NewBlinkCursorAnimation())
		}
	}
	if np := w.nativePlatform; np != nil {
		if prev > 0 && id != prev {
			np.IMEStop()
		}
		if id > 0 {
			np.IMEStart()
		}
	}
}

// resetBlinkCursorVisible resets the blink timer so the cursor
// stays visible during typing and cursor movement.
func resetBlinkCursorVisible(w *Window) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.viewState.inputCursorOn = true
	if a, ok := w.animations[blinkCursorAnimationID]; ok {
		a.SetStart(time.Now())
	}
}

// PointerOverApp returns true if the mouse pointer is within
// the application window bounds.
func (w *Window) PointerOverApp(e *Event) bool {
	return e.MouseX >= 0 && e.MouseY >= 0 &&
		e.MouseX <= float32(w.windowWidth) &&
		e.MouseY <= float32(w.windowHeight)
}

// clearInputSelections zeros SelectBeg/SelectEnd for all
// input states.
func (w *Window) clearInputSelections() {
	imap := StateMapRead[uint32, InputState](w, nsInput)
	if imap == nil {
		return
	}
	imap.Range(func(key uint32, v InputState) bool {
		v.SelectBeg = 0
		v.SelectEnd = 0
		imap.Set(key, v)
		return true
	})
}

// IsFocus tests if the given id_focus equals the window's id_focus.
func (w *Window) IsFocus(idFocus uint32) bool {
	return w.viewState.idFocus > 0 && w.viewState.idFocus == idFocus
}

// SetMouseCursor sets the mouse cursor shape.
func (w *Window) SetMouseCursor(cursor MouseCursor) {
	w.viewState.mouseCursor = cursor
}

// HasFocus returns true if the window has focus.
func (w *Window) HasFocus() bool {
	return w.focused
}

// SetMouseCursorArrow sets the mouse cursor.
func (w *Window) SetMouseCursorArrow() { w.SetMouseCursor(CursorArrow) }

// SetMouseCursorIBeam sets the mouse cursor.
func (w *Window) SetMouseCursorIBeam() { w.SetMouseCursor(CursorIBeam) }

// SetMouseCursorCrosshair sets the mouse cursor.
func (w *Window) SetMouseCursorCrosshair() { w.SetMouseCursor(CursorCrosshair) }

// SetMouseCursorPointingHand sets the mouse cursor.
func (w *Window) SetMouseCursorPointingHand() { w.SetMouseCursor(CursorPointingHand) }

// SetMouseCursorAll sets the mouse cursor.
func (w *Window) SetMouseCursorAll() { w.SetMouseCursor(CursorResizeAll) }

// SetMouseCursorNS sets the mouse cursor.
func (w *Window) SetMouseCursorNS() { w.SetMouseCursor(CursorResizeNS) }

// SetMouseCursorEW sets the mouse cursor.
func (w *Window) SetMouseCursorEW() { w.SetMouseCursor(CursorResizeEW) }

// SetMouseCursorResizeNESW sets the mouse cursor.
func (w *Window) SetMouseCursorResizeNESW() { w.SetMouseCursor(CursorResizeNESW) }

// SetMouseCursorResizeNWSE sets the mouse cursor.
func (w *Window) SetMouseCursorResizeNWSE() { w.SetMouseCursor(CursorResizeNWSE) }

// SetMouseCursorNotAllowed sets the mouse cursor.
func (w *Window) SetMouseCursorNotAllowed() { w.SetMouseCursor(CursorNotAllowed) }

// InputCursorOn returns the input cursor blink state.
func (w *Window) InputCursorOn() bool {
	return w.viewState.inputCursorOn
}

// MouseIsLocked returns true if the mouse is locked (drag).
func (w *Window) MouseIsLocked() bool {
	ml := &w.viewState.mouseLock
	return ml.MouseDown != nil ||
		ml.MouseMove != nil || ml.MouseUp != nil
}

// MouseLock locks the mouse so all mouse events go to the
// handlers in MouseLockCfg.
func (w *Window) MouseLock(cfg MouseLockCfg) {
	w.viewState.mouseLock = cfg
}

// MouseUnlock returns mouse handling events to normal behavior.
func (w *Window) MouseUnlock() {
	w.viewState.mouseLock = MouseLockCfg{}
}

// SetTextMeasurer sets the text measurement backend.
func (w *Window) SetTextMeasurer(tm TextMeasurer) {
	w.textMeasurer = tm
}

// SetWakeMainFn sets the function called to wake the main event
// loop from WaitEventTimeout. The backend sets this at init time.
func (w *Window) SetWakeMainFn(fn func()) {
	w.wakeMainFn = fn
}

// TextWidth measures the rendered width of text for the supplied style.
// When no backend measurer is available, it uses the same approximation
// as text layout generation.
func (w *Window) TextWidth(text string, style TextStyle) float32 {
	if style.Size == 0 {
		style.Size = SizeTextMedium
	}
	if w == nil || w.textMeasurer == nil {
		return float32(utf8RuneCount(text)) * style.Size * 0.6
	}
	return w.textMeasurer.TextWidth(text, style)
}

// SetClipboardFn sets the function used to copy text to the clipboard.
func (w *Window) SetClipboardFn(fn func(string)) {
	w.clipboardSetFn = fn
}

// SetClipboard copies text to the system clipboard.
func (w *Window) SetClipboard(text string) {
	if w.clipboardSetFn != nil {
		w.clipboardSetFn(text)
	}
}

// SetClipboardGetFn sets the function used to read from the clipboard.
func (w *Window) SetClipboardGetFn(fn func() string) {
	w.clipboardGetFn = fn
}

// GetClipboard returns text from the system clipboard.
func (w *Window) GetClipboard() string {
	if w.clipboardGetFn != nil {
		return w.clipboardGetFn()
	}
	return ""
}

// Renderers returns the current render command slice.
func (w *Window) Renderers() []RenderCmd {
	return w.renderers
}

// Timings returns the most recent frame's pipeline timings.
func (w *Window) Timings() FrameTimings { return w.frameTimings }

// MouseCursorState returns the current mouse cursor shape.
func (w *Window) MouseCursorState() MouseCursor {
	return w.viewState.mouseCursor
}

// SetTheme sets the active theme and updates the window.
func (w *Window) SetTheme(t Theme) {
	SetTheme(t)
	w.UpdateWindow()
}

// App returns the parent App, or nil for single-window mode.
func (w *Window) App() *App { return w.app }

// PlatformID returns the SDL window ID (0 if not yet registered).
func (w *Window) PlatformID() uint32 { return w.platformID }

// Close requests the window be closed on the next frame.
// Safe to call from any goroutine.
func (w *Window) Close() { w.closeReq.Store(true) }

// CloseRequested returns true if Close() was called.
func (w *Window) CloseRequested() bool { return w.closeReq.Load() }
