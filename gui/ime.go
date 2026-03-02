package gui

// ime holds per-window Input Method Editor state.
// Created lazily because the native window is not ready during init.
type ime struct {
	initialized bool
}

// initIME lazily creates the IME overlay and registers
// composition callbacks. Called from the frame loop.
func (w *Window) initIME() {
	if w.ime.initialized {
		return
	}
	w.ime.initialized = true
	// Platform-specific IME initialization is handled by
	// the backend. The gui package only tracks init state.
}
