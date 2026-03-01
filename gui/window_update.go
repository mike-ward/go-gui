package gui

// QueueCommand adds a command to the window's atomic command queue.
// Commands execute on the main thread during the next frame update.
// Preferred way to update UI state from other threads.
func (w *Window) QueueCommand(cb func(*Window)) {
	w.commandsMu.Lock()
	w.commands = append(w.commands, cb)
	w.commandsMu.Unlock()
}

// flushCommands executes all pending commands. Called by the main
// loop at frame start.
func (w *Window) flushCommands() {
	w.commandsMu.Lock()
	if len(w.commands) == 0 {
		w.commandsMu.Unlock()
		return
	}
	// Swap to avoid holding lock during execution.
	toRun := w.commands
	w.commands = nil
	w.commandsMu.Unlock()

	for _, cb := range toRun {
		cb(w)
	}
}

// markLayoutRefresh requests a full layout rebuild next frame.
// Overrides any pending render-only refresh.
func (w *Window) markLayoutRefresh() {
	w.refreshLayout = true
	w.refreshRenderOnly = false
}

// markRenderOnlyRefresh requests a renderer-only rebuild from the
// existing layout tree. No-op if a full layout refresh is pending.
func (w *Window) markRenderOnlyRefresh() {
	if !w.refreshLayout {
		w.refreshRenderOnly = true
	}
}

// UpdateWindow marks the window as needing a full layout update.
func (w *Window) UpdateWindow() {
	w.markLayoutRefresh()
}

// RequestRenderOnly marks the window for render-only refresh.
func (w *Window) RequestRenderOnly() {
	w.markRenderOnlyRefresh()
}
