package gui

// inputScrollCursorIntoView adjusts the vertical scroll of a
// multiline input so the cursor remains visible.
// layout must be the outer scroll container (Column with IDScroll).
func inputScrollCursorIntoView(
	idScroll uint32, text string, layout *Layout, w *Window,
) {
	if idScroll == 0 || w.textMeasurer == nil {
		return
	}
	if len(layout.Children) == 0 {
		return
	}
	inner := &layout.Children[0]
	if len(inner.Children) == 0 {
		return
	}
	txtShape := inner.Children[0].Shape
	if txtShape == nil || txtShape.TC == nil {
		return
	}
	style := textStyleOrDefault(txtShape)
	gl, ok := inputGlyphLayout(text, txtShape, style, w)
	if !ok {
		return
	}

	is := StateReadOr(w, nsInput,
		layout.Shape.IDFocus, InputState{})
	runeLen := utf8RuneCount(text)
	pos := is.CursorPos
	if pos > runeLen {
		pos = runeLen
	}
	byteIdx := runeToByteIndex(text, pos)

	cp, ok := gl.GetCursorPos(byteIdx)
	if !ok {
		return
	}
	if is.CursorTrailing {
		for i, line := range gl.Lines {
			if i > 0 && byteIdx == line.StartIndex {
				prev := gl.Lines[i-1]
				cp.X = prev.Rect.X + prev.Rect.Width
				cp.Y = prev.Rect.Y
				cp.Height = prev.Rect.Height
				break
			}
		}
	}

	sy := StateMap[uint32, float32](w, nsScrollY, capScroll)
	scrollOffset, _ := sy.Get(idScroll)
	viewportH := layout.Shape.Height - layout.Shape.PaddingHeight()

	cursorTop := cp.Y
	cursorBot := cp.Y + cp.Height
	visibleTop := -scrollOffset
	visibleBot := visibleTop + viewportH

	if cursorTop < visibleTop {
		sy.Set(idScroll, -cursorTop)
	} else if cursorBot > visibleBot {
		sy.Set(idScroll, -(cursorBot - viewportH))
	}
}

// scrollHorizontal adjusts the horizontal scroll offset of a
// scrollable layout. Returns true if offset was adjusted.
func scrollHorizontal(layout *Layout, delta float32, w *Window) bool {
	idScroll := layout.Shape.IDScroll
	if idScroll == 0 {
		return false
	}
	maxOffset := f32Min(0,
		layout.Shape.Width-layout.Shape.PaddingWidth()-
			contentWidth(layout))
	sx := StateMap[uint32, float32](w, nsScrollX, capScroll)
	old, _ := sx.Get(idScroll)
	offsetX := old + delta*guiTheme.ScrollMultiplier
	sx.Set(idScroll, f32Clamp(offsetX, maxOffset, 0))
	if layout.Shape.HasEvents() &&
		layout.Shape.Events.OnScroll != nil {
		layout.Shape.Events.OnScroll(layout, w)
	}
	return true
}

// scrollVertical adjusts the vertical scroll offset of a
// scrollable layout. Returns true if offset was adjusted.
func scrollVertical(layout *Layout, delta float32, w *Window) bool {
	idScroll := layout.Shape.IDScroll
	if idScroll == 0 {
		return false
	}
	maxOffset := f32Min(0,
		layout.Shape.Height-layout.Shape.PaddingHeight()-
			contentHeight(layout))
	sy := StateMap[uint32, float32](w, nsScrollY, capScroll)
	old, _ := sy.Get(idScroll)
	offsetY := old + delta*guiTheme.ScrollMultiplier
	sy.Set(idScroll, f32Clamp(offsetY, maxOffset, 0))
	if layout.Shape.HasEvents() &&
		layout.Shape.Events.OnScroll != nil {
		layout.Shape.Events.OnScroll(layout, w)
	}
	return true
}

// ScrollToView scrolls the parent scroll container to make
// the view with the given id visible.
func (w *Window) ScrollToView(id string) {
	target, ok := w.layout.FindByID(id)
	if !ok {
		return
	}
	p := target
	for p.Parent != nil {
		p = p.Parent
		if p.Shape.IDScroll > 0 {
			scrollID := p.Shape.IDScroll
			sy := StateMap[uint32, float32](w, nsScrollY, capScroll)
			current, _ := sy.Get(scrollID)
			baseY := p.Shape.Y + p.Shape.Padding.Top
			newScroll := baseY - target.Shape.Y + current
			sy.Set(scrollID, newScroll)
			w.UpdateWindow()
			return
		}
	}
}

// ScrollHorizontalBy scrolls the given scrollable by delta.
func (w *Window) ScrollHorizontalBy(idScroll uint32, delta float32) {
	sx := StateMap[uint32, float32](w, nsScrollX, capScroll)
	current, _ := sx.Get(idScroll)
	sx.Set(idScroll, current+delta)
}

// ScrollHorizontalTo scrolls the given scrollable to offset
// (negative).
func (w *Window) ScrollHorizontalTo(idScroll uint32, offset float32) {
	sx := StateMap[uint32, float32](w, nsScrollX, capScroll)
	sx.Set(idScroll, offset)
}

// ScrollHorizontalToPct scrolls to a horizontal percentage.
// pct: 0.0 = left, 1.0 = right. Clamped to [0, 1].
// No-op if id_scroll not found or content fits viewport.
func (w *Window) ScrollHorizontalToPct(idScroll uint32, pct float32) {
	ly, ok := FindLayoutByIDScroll(&w.layout, idScroll)
	if !ok {
		return
	}
	maxOffset := f32Min(0,
		ly.Shape.Width-ly.Shape.PaddingWidth()-
			contentWidth(ly))
	if maxOffset == 0 {
		return
	}
	sx := StateMap[uint32, float32](w, nsScrollX, capScroll)
	sx.Set(idScroll, maxOffset*f32Clamp(pct, 0, 1))
}

// ScrollHorizontalPct returns the current horizontal scroll
// position as a percentage (0.0 = left, 1.0 = right).
// Returns 0 if not found or content fits viewport.
func (w *Window) ScrollHorizontalPct(idScroll uint32) float32 {
	ly, ok := FindLayoutByIDScroll(&w.layout, idScroll)
	if !ok {
		return 0
	}
	maxOffset := f32Min(0,
		ly.Shape.Width-ly.Shape.PaddingWidth()-
			contentWidth(ly))
	if maxOffset == 0 {
		return 0
	}
	sx := StateMap[uint32, float32](w, nsScrollX, capScroll)
	current, _ := sx.Get(idScroll)
	return f32Clamp(current/maxOffset, 0, 1)
}

// ScrollVerticalBy scrolls the given scrollable by delta.
func (w *Window) ScrollVerticalBy(idScroll uint32, delta float32) {
	sy := StateMap[uint32, float32](w, nsScrollY, capScroll)
	current, _ := sy.Get(idScroll)
	sy.Set(idScroll, current+delta)
}

// ScrollVerticalTo scrolls the given scrollable to offset
// (negative).
func (w *Window) ScrollVerticalTo(idScroll uint32, offset float32) {
	sy := StateMap[uint32, float32](w, nsScrollY, capScroll)
	sy.Set(idScroll, offset)
}

// ScrollVerticalToPct scrolls to a vertical percentage.
// pct: 0.0 = top, 1.0 = bottom. Clamped to [0, 1].
// No-op if id_scroll not found or content fits viewport.
func (w *Window) ScrollVerticalToPct(idScroll uint32, pct float32) {
	ly, ok := FindLayoutByIDScroll(&w.layout, idScroll)
	if !ok {
		return
	}
	maxOffset := f32Min(0,
		ly.Shape.Height-ly.Shape.PaddingHeight()-
			contentHeight(ly))
	if maxOffset == 0 {
		return
	}
	sy := StateMap[uint32, float32](w, nsScrollY, capScroll)
	sy.Set(idScroll, maxOffset*f32Clamp(pct, 0, 1))
}

// ScrollVerticalPct returns the current vertical scroll
// position as a percentage (0.0 = top, 1.0 = bottom).
// Returns 0 if not found or content fits viewport.
func (w *Window) ScrollVerticalPct(idScroll uint32) float32 {
	ly, ok := FindLayoutByIDScroll(&w.layout, idScroll)
	if !ok {
		return 0
	}
	maxOffset := f32Min(0,
		ly.Shape.Height-ly.Shape.PaddingHeight()-
			contentHeight(ly))
	if maxOffset == 0 {
		return 0
	}
	sy := StateMap[uint32, float32](w, nsScrollY, capScroll)
	current, _ := sy.Get(idScroll)
	return f32Clamp(current/maxOffset, 0, 1)
}
