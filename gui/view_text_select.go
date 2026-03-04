package gui

import "time"

const animIDTextDragScroll = "text-drag-scroll"

// makeTextOnClick returns a click handler for text selection.
// Single-click places cursor, double-click (400ms) selects word,
// drag-to-select via MouseLock.
func makeTextOnClick() func(*Layout, *Event, *Window) {
	return func(layout *Layout, e *Event, w *Window) {
		shape := layout.Shape
		if shape.TC == nil || shape.IDFocus == 0 {
			return
		}
		w.SetIDFocus(shape.IDFocus)

		text := shape.TC.Text
		style := textStyleOrDefault(shape)
		gl, glOK := inputGlyphLayout(text, shape, style, w)

		// e.MouseX/Y are already relative to shape via
		// eventRelativeTo in executeMouseCallback.
		relX := e.MouseX
		relY := e.MouseY

		var runePos int
		if glOK {
			byteIdx := gl.GetClosestOffset(relX, relY)
			runePos = byteToRuneIndex(text, byteIdx)
		} else {
			// Fallback for tests (no glyph backend).
			charWidth := style.Size * 0.6
			if charWidth <= 0 {
				charWidth = 14 * 0.6
			}
			runePos = int(relX / charWidth)
			runeLen := utf8RuneCount(text)
			if runePos < 0 {
				runePos = 0
			}
			if runePos > runeLen {
				runePos = runeLen
			}
		}

		idFocus := shape.IDFocus
		imap := StateMap[uint32, InputState](
			w, nsInput, capMany,
		)
		is, _ := imap.Get(idFocus)

		// Double-click detection (400ms threshold).
		now := time.Now().UnixMilli()
		doubleClick := is.LastClickTime > 0 &&
			now-is.LastClickTime <= 400
		is.LastClickTime = now

		runes := []rune(text)
		if doubleClick {
			beg, end := wordBoundsAt(runes, runePos)
			is.CursorPos = end
			is.SelectBeg = uint32(beg)
			is.SelectEnd = uint32(end)
		} else {
			is.CursorPos = runePos
			is.SelectBeg = uint32(runePos)
			is.SelectEnd = uint32(runePos)
		}
		is.CursorOffset = -1
		imap.Set(idFocus, is)
		resetBlinkCursorVisible(w)
		e.IsHandled = true

		// Drag-to-select via MouseLock.
		anchorPos := is.SelectBeg
		anchorEnd := is.SelectEnd
		dragGL := gl
		dragGLOK := glOK
		dragIDFocus := idFocus
		dragShapeX := shape.X
		dragShapeY := shape.Y

		// Find nearest scroll ancestor.
		var lastMouseX, lastMouseY float32
		scrollID := uint32(0)
		dragScrollY0 := float32(0)
		viewTop := float32(0)
		viewBot := float32(0)
		maxScrollNeg := float32(0)
		for p := layout.Parent; p != nil; p = p.Parent {
			if p.Shape != nil && p.Shape.IDScroll > 0 {
				scrollID = p.Shape.IDScroll
				sy := StateMap[uint32, float32](
					w, nsScrollY, capScroll)
				dragScrollY0, _ = sy.Get(scrollID)
				sp := p.Shape
				viewTop = sp.Y + sp.Padding.Top
				viewH := sp.Height - sp.PaddingHeight()
				viewBot = viewTop + viewH
				maxScrollNeg = f32Min(0,
					viewH-contentHeight(p))
				break
			}
		}

		computeRunePos := func(mx, my float32,
			w *Window,
		) int {
			if dragGLOK {
				scrollDelta := float32(0)
				if scrollID > 0 {
					sy := StateMap[uint32, float32](
						w, nsScrollY, capScroll)
					sNow, _ := sy.Get(scrollID)
					scrollDelta = sNow - dragScrollY0
				}
				rx := mx - dragShapeX
				ry := my - (dragShapeY + scrollDelta)
				bi := dragGL.GetClosestOffset(rx, ry)
				return byteToRuneIndex(text, bi)
			}
			cw := style.Size * 0.6
			if cw <= 0 {
				cw = 14 * 0.6
			}
			rp := int((mx - dragShapeX) / cw)
			rl := utf8RuneCount(text)
			if rp < 0 {
				rp = 0
			}
			if rp > rl {
				rp = rl
			}
			return rp
		}

		updateDragSelection := func(rp int, w *Window) {
			dim := StateMap[uint32, InputState](
				w, nsInput, capMany,
			)
			dis, _ := dim.Get(dragIDFocus)
			if doubleClick {
				wb, we := wordBoundsAt(runes, rp)
				if rp < int(anchorPos) {
					dis.SelectBeg = anchorEnd
					dis.SelectEnd = uint32(wb)
					dis.CursorPos = wb
				} else {
					dis.SelectBeg = anchorPos
					dis.SelectEnd = uint32(we)
					dis.CursorPos = we
				}
			} else {
				dis.CursorPos = rp
				dis.SelectBeg = anchorPos
				dis.SelectEnd = uint32(rp)
			}
			dis.CursorOffset = -1
			dim.Set(dragIDFocus, dis)
			resetBlinkCursorVisible(w)
		}

		dragScrollCB := func(_ *Animate, w *Window) {
			var delta float32
			if lastMouseY < viewTop {
				delta = (viewTop - lastMouseY) * 0.3
			} else if lastMouseY > viewBot {
				delta = -((lastMouseY - viewBot) * 0.3)
			} else {
				w.AnimationRemove(animIDTextDragScroll)
				return
			}
			sy := StateMap[uint32, float32](
				w, nsScrollY, capScroll)
			cur, _ := sy.Get(scrollID)
			newScroll := f32Clamp(
				cur+delta, maxScrollNeg, 0)
			if newScroll == cur {
				return
			}
			sy.Set(scrollID, newScroll)
			rp := computeRunePos(
				lastMouseX, lastMouseY, w)
			updateDragSelection(rp, w)
		}

		w.MouseLock(MouseLockCfg{
			MouseMove: func(_ *Layout, e *Event, w *Window) {
				lastMouseX = e.MouseX
				lastMouseY = e.MouseY
				rp := computeRunePos(
					e.MouseX, e.MouseY, w)
				updateDragSelection(rp, w)
				if scrollID > 0 {
					outside := e.MouseY < viewTop ||
						e.MouseY > viewBot
					if outside && !w.HasAnimation(
						animIDTextDragScroll) {
						w.AnimationAdd(&Animate{
							AnimateID: animIDTextDragScroll,
							Delay:     32 * time.Millisecond,
							Repeat:    true,
							Refresh:   AnimationRefreshLayout,
							Callback:  dragScrollCB,
						})
					} else if !outside {
						w.AnimationRemove(
							animIDTextDragScroll)
					}
				}
			},
			MouseUp: func(_ *Layout, _ *Event, w *Window) {
				w.AnimationRemove(animIDTextDragScroll)
				w.MouseUnlock()
			},
		})
	}
}

// makeTextOnKeyDown returns a read-only key handler for text
// navigation and copy. No editing keys (paste, cut, delete).
func makeTextOnKeyDown() func(*Layout, *Event, *Window) {
	return func(layout *Layout, e *Event, w *Window) {
		shape := layout.Shape
		if shape.TC == nil || shape.IDFocus == 0 ||
			!w.IsFocus(shape.IDFocus) {
			return
		}
		id := shape.IDFocus
		text := shape.TC.Text
		imap := StateMap[uint32, InputState](
			w, nsInput, capMany,
		)
		is, _ := imap.Get(id)
		savedOffset := is.CursorOffset
		savedTrailing := is.CursorTrailing
		is.CursorOffset = -1
		is.CursorTrailing = false
		runeLen := utf8RuneCount(text)
		pos := is.CursorPos
		if pos > runeLen {
			pos = runeLen
		}
		isShift := e.Modifiers.Has(ModShift)
		isWordMod := e.Modifiers.HasAny(
			ModCtrl, ModAlt, ModSuper,
		)
		handled := true

		gl, glOK := inputGlyphLayout(
			text, shape, textStyleOrDefault(shape), w,
		)

		switch e.KeyCode {
		case KeyLeft:
			if isWordMod {
				var newPos int
				if glOK {
					bi := runeToByteIndex(text, pos)
					newPos = byteToRuneIndex(text,
						gl.MoveCursorWordLeft(bi))
				} else {
					newPos = moveCursorWordLeft(
						[]rune(text), pos)
				}
				updateCursorAndSelection(
					imap, id, is, newPos, isShift)
			} else if !isShift &&
				is.SelectBeg != is.SelectEnd {
				beg, _ := u32Sort(
					is.SelectBeg, is.SelectEnd)
				updateCursorAndSelection(
					imap, id, is, int(beg), false)
			} else {
				newPos := pos - 1
				if newPos < 0 {
					newPos = 0
				}
				updateCursorAndSelection(
					imap, id, is, newPos, isShift)
			}
		case KeyRight:
			if isWordMod {
				var newPos int
				if glOK {
					bi := runeToByteIndex(text, pos)
					newPos = byteToRuneIndex(text,
						gl.MoveCursorWordRight(bi))
				} else {
					newPos = moveCursorWordRight(
						[]rune(text), pos)
				}
				updateCursorAndSelection(
					imap, id, is, newPos, isShift)
			} else if !isShift &&
				is.SelectBeg != is.SelectEnd {
				_, end := u32Sort(
					is.SelectBeg, is.SelectEnd)
				updateCursorAndSelection(
					imap, id, is, int(end), false)
			} else {
				newPos := pos + 1
				if newPos > runeLen {
					newPos = runeLen
				}
				updateCursorAndSelection(
					imap, id, is, newPos, isShift)
			}
		case KeyHome:
			var newPos int
			if glOK {
				bi := runeToByteIndex(text, pos)
				sb := gl.MoveCursorLineStart(bi)
				if savedTrailing {
					sb = trailingLineStart(
						gl.Lines, bi, sb)
				}
				lineStart := byteToRuneIndex(text, sb)
				if pos != lineStart {
					newPos = lineStart
				} else {
					ps := cursorStartOfParagraph(
						text, pos)
					if pos != ps {
						newPos = ps
					} else {
						newPos = cursorHome()
					}
				}
			} else {
				ls := moveCursorLineStart(
					[]rune(text), pos)
				if pos != ls {
					newPos = ls
				} else {
					newPos = cursorHome()
				}
			}
			updateCursorAndSelection(
				imap, id, is, newPos, isShift)
		case KeyEnd:
			var newPos int
			trailingLevel1 := false
			if glOK {
				bi := runeToByteIndex(text, pos)
				eb := gl.MoveCursorLineEnd(bi)
				if savedTrailing {
					eb = trailingLineEnd(
						gl.Lines, bi, eb)
				}
				lineEnd := byteToRuneIndex(text, eb)
				if pos != lineEnd {
					newPos = lineEnd
					trailingLevel1 = true
				} else {
					pe := cursorEndOfParagraph(
						text, pos)
					if pos != pe {
						newPos = pe
					} else {
						newPos = cursorEnd(text)
					}
				}
			} else {
				le := moveCursorLineEnd(
					[]rune(text), pos)
				if pos != le {
					newPos = le
					trailingLevel1 = true
				} else {
					newPos = cursorEnd(text)
				}
			}
			is.CursorTrailing = trailingLevel1
			updateCursorAndSelection(
				imap, id, is, newPos, isShift)
		case KeyUp:
			var newPos int
			if glOK {
				bi := runeToByteIndex(text, pos)
				px := savedOffset
				if px < 0 {
					if cp, ok := gl.GetCursorPos(bi); ok {
						px = cp.X
					}
				}
				is.CursorOffset = px
				newPos = byteToRuneIndex(text,
					gl.MoveCursorUp(bi, px))
			} else {
				newPos = moveCursorUp(
					[]rune(text), pos)
			}
			updateCursorAndSelection(
				imap, id, is, newPos, isShift)
		case KeyDown:
			var newPos int
			if glOK {
				bi := runeToByteIndex(text, pos)
				px := savedOffset
				if px < 0 {
					if cp, ok := gl.GetCursorPos(bi); ok {
						px = cp.X
					}
				}
				is.CursorOffset = px
				newPos = byteToRuneIndex(text,
					gl.MoveCursorDown(bi, px))
			} else {
				newPos = moveCursorDown(
					[]rune(text), pos)
			}
			updateCursorAndSelection(
				imap, id, is, newPos, isShift)
		case KeyEscape:
			is.SelectBeg = 0
			is.SelectEnd = 0
			imap.Set(id, is)
		case KeyA:
			if e.Modifiers.HasAny(ModCtrl, ModSuper) {
				inputSelectAll(text, id, w)
			} else {
				handled = false
			}
		case KeyC:
			if e.Modifiers.HasAny(ModCtrl, ModSuper) {
				if copied, ok := inputCopy(
					text, id,
					shape.TC.TextIsPassword, w,
				); ok {
					w.SetClipboard(copied)
				}
			} else {
				handled = false
			}
		default:
			handled = false
		}

		if handled {
			resetBlinkCursorVisible(w)
			textScrollCursorIntoView(layout, w)
			e.IsHandled = true
		}
	}
}

// textAmendLayout copies InputState selection to the shape's
// TextSelBeg/TextSelEnd for rendering. Unlike input's nested
// structure, text is a flat shape — no child traversal needed.
func textAmendLayout(layout *Layout, w *Window) {
	if layout.Shape.IDFocus == 0 || layout.Shape.TC == nil {
		return
	}
	is := StateReadOr(
		w, nsInput, layout.Shape.IDFocus, InputState{},
	)
	layout.Shape.TC.TextSelBeg = is.SelectBeg
	layout.Shape.TC.TextSelEnd = is.SelectEnd
}
