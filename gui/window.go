package gui

// Window is the main application window. Phase 1 provides a minimal
// stub sufficient for the layout engine and state registry tests.
type Window struct {
	viewState     ViewState
	refreshLayout bool
}

// ViewState holds per-window UI state.
type ViewState struct {
	registry StateRegistry
	idFocus  uint32
}

// ClearViewState resets all view state.
func (w *Window) ClearViewState() {
	w.viewState.registry.Clear()
	w.viewState.idFocus = 0
}
