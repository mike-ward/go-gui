package gui

import "slices"

// layoutArrange is the top-level layout orchestrator. It sets
// parents, extracts floats, injects toast/dialog overlays, runs
// the pipeline on each layer, and processes hover in reverse
// layer order.
func layoutArrange(layout *Layout, w *Window) []Layout {
	ensureLayoutShape(layout)

	// Set parent pointers.
	layoutParents(layout, nil)

	// Extract floating layouts from main tree.
	floatingLayouts := w.scratch.takeFloatingLayouts(len(layout.Children))
	defer func() {
		w.scratch.putFloatingLayouts(floatingLayouts)
	}()
	layoutRemoveFloatingLayouts(layout, w, &floatingLayouts)

	slices.SortStableFunc(floatingLayouts,
		func(a, b *Layout) int {
			return a.Shape.FloatZIndex - b.Shape.FloatZIndex
		})

	// Inject inspector overlay as a floating layer.
	if inspectorSupported && w.inspectorEnabled {
		iv := inspectorFloatingPanel(w)
		if iv != nil {
			il := GenerateViewLayout(iv, w)
			heapLayout := w.scratch.allocFloatingLayout(il)
			layoutParents(heapLayout, nil)
			floatingLayouts = append(floatingLayouts, heapLayout)
		}
	}

	// Inject toast container as floating layer.
	if len(w.toasts) > 0 {
		tv := toastContainerView(w)
		if tv != nil {
			tl := GenerateViewLayout(tv, w)
			heapLayout := w.scratch.allocFloatingLayout(tl)
			layoutParents(heapLayout, nil)
			floatingLayouts = append(floatingLayouts, heapLayout)
		}
	}

	// Inject dialog as last floating layer (always on top).
	if w.dialogCfg.visible {
		dv := dialogViewGenerator(w.dialogCfg)
		if dv != nil {
			dl := GenerateViewLayout(dv, w)
			heapLayout := w.scratch.allocFloatingLayout(dl)
			layoutParents(heapLayout, nil)
			floatingLayouts = append(floatingLayouts, heapLayout)
		}
	}

	// Run pipeline on main layout.
	layoutPipeline(layout, w)
	layouts := w.scratch.layerLayouts.take(1 + len(floatingLayouts))
	layouts = append(layouts, *layout)

	// Run pipeline on each floating layout.
	for _, fl := range floatingLayouts {
		// Note: clips haven't been set yet at this point, so
		// invisible floats still need the pipeline.
		layoutPipeline(fl, w)
		layouts = append(layouts, *fl)
	}

	// Hover processing: topmost first.
	for i := len(layouts) - 1; i >= 0; i-- {
		handled := layoutHover(&layouts[i], w)
		if handled && i > 0 {
			// Cursor inside a floating layer blocks lower layers.
			break
		}
	}

	return layouts
}
