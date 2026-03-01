package gui

import "sync"

// Window is the main application window.
type Window struct {
	// Mutexes.
	mu         sync.Mutex // guards layout/renderer state
	commandsMu sync.Mutex // guards command queue

	// User state — accessed via State[T](w).
	state any

	// View state.
	viewState ViewState

	// Command queue — flushed at frame start.
	commands []func(*Window)

	// Layout tree — current frame.
	layout Layout

	// Renderers — flat draw command list, reused via [:0].
	renderers []RenderCmd

	// Clip radius propagated during render walk.
	clipRadius float32

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

	// Toast state.
	toasts       []toastNotification
	toastCounter uint64
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
	registry       StateRegistry
	idFocus        uint32
	mouseCursor    MouseCursor
	mouseLock      MouseLockCfg
	cursorOnSticky bool
	mousePosX      float32
	mousePosY      float32
	tooltip        tooltipState
}

// State returns a typed pointer to the user-supplied state.
func State[T any](w *Window) *T {
	return w.state.(*T)
}

// SetState sets the user state for the window.
func (w *Window) SetState(state any) {
	w.state = state
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

// SetIDFocus sets the focus id.
func (w *Window) SetIDFocus(id uint32) {
	w.viewState.idFocus = id
}

// IsFocus tests if the given id_focus equals the window's id_focus.
func (w *Window) IsFocus(idFocus uint32) bool {
	return w.viewState.idFocus > 0 && w.viewState.idFocus == idFocus
}

// SetMouseCursor sets the mouse cursor shape.
func (w *Window) SetMouseCursor(cursor MouseCursor) {
	w.viewState.mouseCursor = cursor
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
