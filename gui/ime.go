package gui

// ime holds per-window Input Method Editor composition state.
type ime struct {
	compText   string
	compCursor int
	compSelLen int
	composing  bool
}

// imeUpdate sets composition state from an IME composition event.
func (w *Window) imeUpdate(e *Event) {
	if len(e.IMEText) == 0 {
		w.imeClear()
		return
	}
	w.ime.composing = true
	w.ime.compText = e.IMEText
	w.ime.compCursor = int(e.IMEStart)
	w.ime.compSelLen = int(e.IMELength)
}

// imeClear resets composition state (called on commit or focus
// change).
func (w *Window) imeClear() {
	w.ime.composing = false
	w.ime.compText = ""
	w.ime.compCursor = 0
	w.ime.compSelLen = 0
}

// IMEComposing returns true if an IME composition is in progress.
func (w *Window) IMEComposing() bool {
	return w.ime.composing
}

// IMECompText returns the current IME preedit string.
func (w *Window) IMECompText() string {
	return w.ime.compText
}

// IMESetRect reports the cursor rect to the platform so the
// candidate window positions correctly. No-ops if no backend.
func (w *Window) IMESetRect(x, y, width, height float32) {
	if np := w.nativePlatform; np != nil {
		np.IMESetRect(
			int32(x), int32(y), int32(width), int32(height),
		)
	}
}
