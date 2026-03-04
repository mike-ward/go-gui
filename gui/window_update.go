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

// UpdateView sets the view generator and triggers a full refresh.
func (w *Window) UpdateView(gen func(*Window) View) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.viewState.registry.Clear()
	w.viewGenerator = gen
	w.markLayoutRefresh()
}

// FrameFn is called by the backend each frame. It flushes
// queued commands and rebuilds layout/renderers as needed.
func (w *Window) FrameFn() {
	w.flushCommands()
	if w.refreshLayout {
		w.Update()
	} else if w.refreshRenderOnly {
		w.UpdateRenderOnly()
	}
}

// Update performs a full layout rebuild and re-renders.
func (w *Window) Update() {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.refreshLayout = false
	w.refreshRenderOnly = false

	if w.viewGenerator == nil {
		return
	}

	if len(w.layout.Children) > 0 {
		w.scratch.putLayerLayouts(w.layout.Children)
	}

	view := w.viewGenerator(w)
	rootLayout := GenerateViewLayout(view, w)
	layers := layoutArrange(&rootLayout, w)
	w.layout = composeLayout(layers, w)
	w.buildRenderers(w.Config.BgColor, w.WindowRect())
}

// UpdateRenderOnly rebuilds renderers from the existing layout.
func (w *Window) UpdateRenderOnly() {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.refreshRenderOnly = false
	w.buildRenderers(w.Config.BgColor, w.WindowRect())
}

// composeLayout wraps layer layouts into a single root.
func composeLayout(layers []Layout, w *Window) Layout {
	root := Layout{
		Shape: &Shape{
			Width:  float32(w.windowWidth),
			Height: float32(w.windowHeight),
		},
		Children: layers,
	}
	return root
}

// buildRenderers resets and rebuilds the render command list.
func (w *Window) buildRenderers(bgColor Color, clip DrawClip) {
	w.renderers = w.renderers[:0]
	renderLayout(&w.layout, bgColor, clip, w)
}
