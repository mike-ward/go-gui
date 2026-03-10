package gui

import (
	"time"

	"github.com/mike-ward/go-glyph"
)

const animIDDragScroll = "input-drag-scroll"

// InputCfg configures a text input field.
type InputCfg struct {
	ID          string
	Text        string
	Placeholder string
	Mask        string
	MaskPreset  InputMaskPreset
	MaskTokens  []MaskTokenDef
	Mode        InputMode
	IsPassword  bool
	Disabled    bool
	Invisible   bool

	// Sizing
	Sizing    Sizing
	Width     float32
	Height    float32
	MinWidth  float32
	MaxWidth  float32
	MinHeight float32
	MaxHeight float32

	// Appearance
	Padding          Opt[Padding]
	Radius           Opt[float32]
	SizeBorder       Opt[float32]
	Color            Color
	ColorHover       Color
	ColorBorder      Color
	ColorBorderFocus Color
	TextStyle        TextStyle
	PlaceholderStyle TextStyle

	// Focus
	IDFocus  uint32
	IDScroll uint32

	// Callbacks
	OnTextChanged      func(*Layout, string, *Window)
	OnTextCommit       func(*Layout, string, InputCommitReason, *Window)
	OnEnter            func(*Layout, *Event, *Window)
	OnKeyDown          func(*Layout, *Event, *Window)
	OnBlur             func(*Layout, *Window)
	PreTextChange      func(current, proposed string) (string, bool)
	PostCommitNormalize func(text string, reason InputCommitReason) string

	// Accessibility
	A11YLabel       string
	A11YDescription string
}

// Input creates a text input field view.
func Input(cfg InputCfg) View {
	applyInputDefaults(&cfg)

	d := &DefaultInputStyle
	sizeBorder := cfg.SizeBorder.Get(d.SizeBorder)
	radius := cfg.Radius.Get(d.Radius)

	placeholderActive := len(cfg.Text) == 0
	txt := cfg.Text
	if placeholderActive {
		txt = cfg.Placeholder
	}
	txtStyle := cfg.TextStyle
	if placeholderActive {
		txtStyle = cfg.PlaceholderStyle
	}
	mode := TextModeSingleLine
	if cfg.Mode == InputMultiline {
		mode = TextModeWrapKeepSpaces
	}

	colorBorderFocus := cfg.ColorBorderFocus
	idFocus := cfg.IDFocus
	onBlur := cfg.OnBlur

	hcfg := inputHandlerCfg{
		IDFocus:       cfg.IDFocus,
		IDScroll:      cfg.IDScroll,
		IsPassword:    cfg.IsPassword,
		Mode:          cfg.Mode,
		Mask:          cfg.Mask,
		MaskPreset:    cfg.MaskPreset,
		MaskTokens:    cfg.MaskTokens,
		OnTextChanged:       cfg.OnTextChanged,
		OnTextCommit:        cfg.OnTextCommit,
		OnEnter:             cfg.OnEnter,
		OnKeyDown:           cfg.OnKeyDown,
		PreTextChange:       cfg.PreTextChange,
		PostCommitNormalize: cfg.PostCommitNormalize,
	}

	txtSizing := Sizing(FillFill)
	innerSizing := Sizing(FillFill)
	if cfg.Mode == InputMultiline && cfg.IDScroll > 0 {
		txtSizing = FillFit
		innerSizing = FillFit
	}

	txtContent := []View{
		Text(TextCfg{
			IDFocus:           cfg.IDFocus,
			Sizing:            txtSizing,
			Text:              txt,
			TextStyle:         txtStyle,
			Mode:              mode,
			IsPassword:        cfg.IsPassword,
			PlaceholderActive: placeholderActive,
		}),
	}

	a11yRole := AccessRoleTextField
	if cfg.Mode == InputMultiline {
		a11yRole = AccessRoleTextArea
	}
	a11yState := AccessStateNone
	if cfg.IDFocus == 0 {
		a11yState = AccessStateReadOnly
	}

	vAlign := VAlignMiddle
	if cfg.Mode == InputMultiline {
		vAlign = VAlignTop
	}

	idScroll := cfg.IDScroll
	innerCfg := ContainerCfg{
		Padding: NoPadding,
		Sizing:  innerSizing,
		VAlign:  vAlign,
		OnClick: func(layout *Layout, e *Event, w *Window) {
			if len(layout.Children) < 1 {
				return
			}
			ly := layout.Children[0]
			if ly.Shape.IDFocus > 0 {
				w.SetIDFocus(ly.Shape.IDFocus)
			}
			if ly.Shape.TC == nil {
				return
			}
			if ly.Shape.TC.TextIsPlaceholder {
				imap := StateMap[uint32, InputState](
					w, nsInput, capMany,
				)
				is, _ := imap.Get(ly.Shape.IDFocus)
				is.CursorPos = 0
				is.SelectBeg = 0
				is.SelectEnd = 0
				is.CursorOffset = -1
				imap.Set(ly.Shape.IDFocus, is)
				resetBlinkCursorVisible(w)
				e.IsHandled = true
				return
			}
			text := ly.Shape.TC.Text
			style := textStyleOrDefault(ly.Shape)
			gl, ok := inputGlyphLayout(
				text, ly.Shape, style, w,
			)
			if !ok {
				return
			}
			relX := e.MouseX - (ly.Shape.X - layout.Shape.X)
			relY := e.MouseY - (ly.Shape.Y - layout.Shape.Y)
			byteIdx := gl.GetClosestOffset(relX, relY)
			displayText := text
			if ly.Shape.TC.TextIsPassword {
				displayText = passwordMask(text)
			}
			runePos := byteToRuneIndex(displayText, byteIdx)
			imap := StateMap[uint32, InputState](
				w, nsInput, capMany,
			)
			is, _ := imap.Get(ly.Shape.IDFocus)

			// Double-click selects word.
			now := time.Now().UnixMilli()
			doubleClick := is.LastClickTime > 0 &&
				now-is.LastClickTime <= 400
			is.LastClickTime = now

			runes := []rune(displayText)
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
			imap.Set(ly.Shape.IDFocus, is)
			resetBlinkCursorVisible(w)
			if idScroll > 0 && layout.Parent != nil {
				inputScrollCursorIntoView(
					idScroll, text, layout.Parent, w,
				)
			}
			e.IsHandled = true

			// Drag-to-select via MouseLock.
			anchorPos := is.SelectBeg
			anchorEnd := is.SelectEnd
			dragGL := gl
			dragDisplayText := displayText
			dragTxtOffX := ly.Shape.X - layout.Shape.X
			dragTxtOffY := ly.Shape.Y - layout.Shape.Y
			dragIDFocus := ly.Shape.IDFocus

			var lastMouseX, lastMouseY float32
			dragScrollY0 := float32(0)
			viewTop := float32(0)
			viewBot := float32(0)
			maxScrollNeg := float32(0)
			if idScroll > 0 && layout.Parent != nil {
				sy := StateMap[uint32, float32](
					w, nsScrollY, capScroll)
				dragScrollY0, _ = sy.Get(idScroll)
				p := layout.Parent.Shape
				viewTop = p.Y + p.Padding.Top
				viewH := p.Height - p.PaddingHeight()
				viewBot = viewTop + viewH
				maxScrollNeg = f32Min(0,
					viewH-layout.Shape.Height)
			}

			computeRunePos := func(mx, my float32, w *Window) int {
				scrollDelta := float32(0)
				if idScroll > 0 {
					sy := StateMap[uint32, float32](
						w, nsScrollY, capScroll)
					sNow, _ := sy.Get(idScroll)
					scrollDelta = sNow - dragScrollY0
				}
				relX := mx - dragTxtOffX
				relY := my - (dragTxtOffY + scrollDelta)
				byteIdx := dragGL.GetClosestOffset(relX, relY)
				return byteToRuneIndex(dragDisplayText, byteIdx)
			}

			updateDragSelection := func(rp int, w *Window) {
				imap := StateMap[uint32, InputState](
					w, nsInput, capMany)
				is, _ := imap.Get(dragIDFocus)
				if doubleClick {
					wb, we := wordBoundsAt(runes, rp)
					if rp < int(anchorPos) {
						is.SelectBeg = anchorEnd
						is.SelectEnd = uint32(wb)
						is.CursorPos = wb
					} else {
						is.SelectBeg = anchorPos
						is.SelectEnd = uint32(we)
						is.CursorPos = we
					}
				} else {
					is.CursorPos = rp
					is.SelectBeg = anchorPos
					is.SelectEnd = uint32(rp)
				}
				is.CursorOffset = -1
				imap.Set(dragIDFocus, is)
				resetBlinkCursorVisible(w)
			}

			dragScrollCB := func(_ *Animate, w *Window) {
				var delta float32
				if lastMouseY < viewTop {
					delta = (viewTop - lastMouseY) * 0.3
				} else if lastMouseY > viewBot {
					delta = -((lastMouseY - viewBot) * 0.3)
				} else {
					w.AnimationRemove(animIDDragScroll)
					return
				}
				sy := StateMap[uint32, float32](
					w, nsScrollY, capScroll)
				cur, _ := sy.Get(idScroll)
				newScroll := f32Clamp(cur+delta, maxScrollNeg, 0)
				if newScroll == cur {
					return
				}
				sy.Set(idScroll, newScroll)
				rp := computeRunePos(lastMouseX, lastMouseY, w)
				updateDragSelection(rp, w)
			}

			w.MouseLock(MouseLockCfg{
				MouseMove: func(_ *Layout, e *Event, w *Window) {
					lastMouseX = e.MouseX
					lastMouseY = e.MouseY
					rp := computeRunePos(e.MouseX, e.MouseY, w)
					updateDragSelection(rp, w)
					if idScroll > 0 {
						outside := e.MouseY < viewTop ||
							e.MouseY > viewBot
						if outside && !w.HasAnimation(
							animIDDragScroll) {
							w.AnimationAdd(&Animate{
								AnimateID: animIDDragScroll,
								Delay:     32 * time.Millisecond,
								Repeat:    true,
								Refresh:   AnimationRefreshLayout,
								Callback:  dragScrollCB,
							})
						} else if !outside {
							w.AnimationRemove(animIDDragScroll)
						}
					}
				},
				MouseUp: func(_ *Layout, _ *Event, w *Window) {
					w.AnimationRemove(animIDDragScroll)
					w.MouseUnlock()
				},
			})
		},
		Content: txtContent,
	}
	var inner View
	if cfg.Mode == InputMultiline {
		inner = Column(innerCfg)
	} else {
		inner = Row(innerCfg)
	}

	return Column(ContainerCfg{
		ID:              cfg.ID,
		IDFocus:         cfg.IDFocus,
		A11YRole:        a11yRole,
		A11YState:       a11yState,
		A11YLabel:       a11yLabel(cfg.A11YLabel, cfg.Placeholder),
		A11YDescription: cfg.A11YDescription,
		Width:           cfg.Width,
		Height:          cfg.Height,
		MinWidth:        cfg.MinWidth,
		MaxWidth:        cfg.MaxWidth,
		MinHeight:       cfg.MinHeight,
		MaxHeight:       cfg.MaxHeight,
		Disabled:        cfg.Disabled,
		Clip:            true,
		Color:           cfg.Color,
		ColorBorder:     cfg.ColorBorder,
		SizeBorder:      Some(sizeBorder),
		Invisible:       cfg.Invisible,
		Padding:         cfg.Padding,
		Radius:          Some(radius),
		Sizing:          cfg.Sizing,
		IDScroll:        cfg.IDScroll,
		Spacing:         SomeF(0),
		OnChar:          makeInputOnChar(hcfg),
		OnKeyDown:       makeInputOnKeyDown(hcfg),
		OnHover: func(_ *Layout, _ *Event, w *Window) {
			if w.IsFocus(idFocus) {
				w.SetMouseCursor(CursorIBeam)
			}
		},
		AmendLayout: func(layout *Layout, w *Window) {
			if layout.Shape.IDFocus == 0 {
				return
			}
			focused := !layout.Shape.Disabled &&
				layout.Shape.IDFocus == w.IDFocus()
			if focused {
				layout.Shape.ColorBorder = colorBorderFocus
			}

			// Blur detection: fire commit on focus loss.
			focusMap := StateMap[uint32, bool](
				w, nsInputFocus, capMany)
			wasFocused, _ := focusMap.Get(layout.Shape.IDFocus)
			focusMap.Set(layout.Shape.IDFocus, focused)
			if wasFocused && !focused {
				text := inputTextFromLayout(layout)
				if hcfg.PostCommitNormalize != nil {
					normalized := hcfg.PostCommitNormalize(
						text, CommitBlur)
					if normalized != text {
						text = normalized
						if hcfg.OnTextChanged != nil {
							hcfg.OnTextChanged(
								layout, text, w)
						}
					}
				}
				if hcfg.OnTextCommit != nil {
					hcfg.OnTextCommit(
						layout, text, CommitBlur, w)
				}
				if onBlur != nil {
					onBlur(layout, w)
				}
			}

			// Propagate selection to inner text shape.
			if len(layout.Children) > 0 {
				inner := &layout.Children[0]
				if len(inner.Children) > 0 {
					txt := &inner.Children[0]
					if txt.Shape.TC != nil {
						is := StateReadOr(w, nsInput,
							layout.Shape.IDFocus,
							InputState{})
						txt.Shape.TC.TextSelBeg = is.SelectBeg
						txt.Shape.TC.TextSelEnd = is.SelectEnd
					}
				}
			}
		},
		Content: []View{inner},
	})
}

func applyInputDefaults(cfg *InputCfg) {
	d := &DefaultButtonStyle
	if !cfg.Color.IsSet() {
		cfg.Color = d.Color
	}
	if !cfg.ColorHover.IsSet() {
		cfg.ColorHover = d.ColorHover
	}
	if !cfg.ColorBorder.IsSet() {
		cfg.ColorBorder = d.ColorBorder
	}
	if !cfg.ColorBorderFocus.IsSet() {
		cfg.ColorBorderFocus = d.ColorBorderFocus
	}
	if !cfg.Padding.IsSet() {
		cfg.Padding = Some(PaddingTwoFour)
	}
	if cfg.TextStyle == (TextStyle{}) {
		cfg.TextStyle = DefaultTextStyle
	}
	if cfg.PlaceholderStyle == (TextStyle{}) {
		cfg.PlaceholderStyle = DefaultInputStyle.PlaceholderStyle
	}
	if !cfg.Radius.IsSet() {
		cfg.Radius = Some(d.Radius)
	}
	if !cfg.SizeBorder.IsSet() {
		cfg.SizeBorder = Some(d.SizeBorder)
	}
}

// inputHandlerCfg captures the fields shared by OnChar and
// OnKeyDown handler factories.
type inputHandlerCfg struct {
	IDFocus             uint32
	IDScroll            uint32
	IsPassword          bool
	Mode                InputMode
	Mask                string
	MaskPreset          InputMaskPreset
	MaskTokens          []MaskTokenDef
	OnTextChanged       func(*Layout, string, *Window)
	OnTextCommit        func(*Layout, string, InputCommitReason, *Window)
	OnEnter             func(*Layout, *Event, *Window)
	OnKeyDown           func(*Layout, *Event, *Window)
	PreTextChange       func(current, proposed string) (string, bool)
	PostCommitNormalize func(text string, reason InputCommitReason) string
}

// compiledMask returns a non-nil *CompiledInputMask if the
// handler config specifies a mask pattern.
func (h *inputHandlerCfg) compiledMask() *CompiledInputMask {
	pattern := h.Mask
	if pattern == "" && h.MaskPreset != MaskNone {
		pattern = InputMaskFromPreset(h.MaskPreset)
	}
	if pattern == "" {
		return nil
	}
	c, err := CompileInputMask(pattern, h.MaskTokens)
	if err != nil {
		return nil
	}
	return &c
}

func makeInputOnChar(hcfg inputHandlerCfg) func(*Layout, *Event, *Window) {
	return func(layout *Layout, e *Event, w *Window) {
		if hcfg.IDFocus == 0 || !w.IsFocus(hcfg.IDFocus) {
			return
		}
		ch := e.CharCode
		id := hcfg.IDFocus

		text := inputTextFromLayout(layout)
		mask := hcfg.compiledMask()
		changed := false

		switch {
		// Skip control characters handled by OnKeyDown.
		case ch < CharSpace && ch != CharBSP && ch != CharDel &&
			ch != CharLF && ch != CharCR:
			e.IsHandled = true
			return
		// Backspace.
		case ch == CharBSP:
			if mask != nil {
				is := inputStateOrDefault(id, w)
				res := InputMaskBackspace(text, is.CursorPos, is.SelectBeg, is.SelectEnd, mask)
				if res.Changed {
					text = res.Text
					StateMap[uint32, InputState](w, nsInput, capMany).Set(id, InputState{
						CursorPos: res.CursorPos, Undo: is.Undo,
					})
					changed = true
				}
			} else {
				newText, delChanged := inputDeleteGrapheme(
					text, id, false, layout, w,
				)
				if delChanged {
					text = newText
					changed = true
				}
			}
		// Delete.
		case ch == CharDel:
			if mask != nil {
				is := inputStateOrDefault(id, w)
				res := InputMaskDelete(text, is.CursorPos, is.SelectBeg, is.SelectEnd, mask)
				if res.Changed {
					text = res.Text
					StateMap[uint32, InputState](w, nsInput, capMany).Set(id, InputState{
						CursorPos: res.CursorPos, Undo: is.Undo,
					})
					changed = true
				}
			} else {
				newText, delChanged := inputDeleteGrapheme(
					text, id, true, layout, w,
				)
				if delChanged {
					text = newText
					changed = true
				}
			}
		// Enter / LF.
		case ch == CharLF || ch == CharCR:
			if hcfg.Mode == InputMultiline {
				text = inputInsert(text, "\n", id, w)
				changed = true
			} else {
				commitText := text
				if hcfg.PostCommitNormalize != nil {
					normalized := hcfg.PostCommitNormalize(
						text, CommitEnter)
					if normalized != text {
						commitText = normalized
						if hcfg.OnTextChanged != nil {
							hcfg.OnTextChanged(
								layout, commitText, w)
						}
					}
				}
				if hcfg.OnTextCommit != nil {
					hcfg.OnTextCommit(
						layout, commitText, CommitEnter, w)
				}
				if hcfg.OnEnter != nil {
					hcfg.OnEnter(layout, e, w)
				}
			}
		// Printable characters.
		case ch >= CharSpace:
			ins := e.IMEText
			if len(ins) == 0 {
				ins = string(rune(ch))
			}
			if mask != nil {
				is := inputStateOrDefault(id, w)
				res := InputMaskInsert(text, is.CursorPos, is.SelectBeg, is.SelectEnd, ins, mask)
				if res.Changed {
					text = res.Text
					StateMap[uint32, InputState](w, nsInput, capMany).Set(id, InputState{
						CursorPos: res.CursorPos, Undo: is.Undo,
					})
					changed = true
				}
			} else if hcfg.PreTextChange != nil {
				proposed := inputProposedText(text, ins, id, w)
				if _, ok := hcfg.PreTextChange(text, proposed); ok {
					text = inputInsert(text, ins, id, w)
					changed = true
				}
			} else {
				text = inputInsert(text, ins, id, w)
				changed = true
			}
		}

		if changed {
			resetBlinkCursorVisible(w)
			if hcfg.OnTextChanged != nil {
				hcfg.OnTextChanged(layout, text, w)
			}
			inputScrollCursorIntoView(
				hcfg.IDScroll, text, layout, w,
			)
		}
		e.IsHandled = true
	}
}

func makeInputOnKeyDown(hcfg inputHandlerCfg) func(*Layout, *Event, *Window) {
	return func(layout *Layout, e *Event, w *Window) {
		if hcfg.IDFocus == 0 || !w.IsFocus(hcfg.IDFocus) {
			return
		}
		id := hcfg.IDFocus
		imap := StateMap[uint32, InputState](w, nsInput, capMany)
		is, _ := imap.Get(id)
		savedOffset := is.CursorOffset
		savedTrailing := is.CursorTrailing
		is.CursorOffset = -1
		is.CursorTrailing = false
		text := inputTextFromLayout(layout)
		runeLen := utf8RuneCount(text)
		pos := is.CursorPos
		if pos > runeLen {
			pos = runeLen
		}
		isShift := e.Modifiers.Has(ModShift)
		isWordMod := e.Modifiers.HasAny(ModCtrl, ModAlt, ModSuper)
		handled := true
		textChanged := false

		// Use glyph layout for cursor navigation when available.
		gl, glOK := inputGlyphLayoutFor(layout, w)

		switch e.KeyCode {
		case KeyLeft:
			if isWordMod {
				var newPos int
				if glOK {
					byteIdx := runeToByteIndex(text, pos)
					newPos = byteToRuneIndex(text, gl.MoveCursorWordLeft(byteIdx))
				} else {
					newPos = moveCursorWordLeft([]rune(text), pos)
				}
				updateCursorAndSelection(imap, id, is,
					newPos, isShift)
			} else if !isShift && is.SelectBeg != is.SelectEnd {
				beg, _ := u32Sort(is.SelectBeg, is.SelectEnd)
				updateCursorAndSelection(imap, id, is,
					int(beg), false)
			} else {
				var newPos int
				if glOK {
					byteIdx := runeToByteIndex(text, pos)
					newPos = byteToRuneIndex(text, gl.MoveCursorLeft(byteIdx))
				} else {
					newPos = pos - 1
					if newPos < 0 {
						newPos = 0
					}
				}
				updateCursorAndSelection(imap, id, is,
					newPos, isShift)
			}
		case KeyRight:
			if isWordMod {
				var newPos int
				if glOK {
					byteIdx := runeToByteIndex(text, pos)
					newPos = byteToRuneIndex(text, gl.MoveCursorWordRight(byteIdx))
				} else {
					newPos = moveCursorWordRight([]rune(text), pos)
				}
				updateCursorAndSelection(imap, id, is,
					newPos, isShift)
			} else if !isShift && is.SelectBeg != is.SelectEnd {
				_, end := u32Sort(is.SelectBeg, is.SelectEnd)
				updateCursorAndSelection(imap, id, is,
					int(end), false)
			} else {
				var newPos int
				if glOK {
					byteIdx := runeToByteIndex(text, pos)
					newPos = byteToRuneIndex(text, gl.MoveCursorRight(byteIdx))
				} else {
					newPos = pos + 1
					if newPos > runeLen {
						newPos = runeLen
					}
				}
				updateCursorAndSelection(imap, id, is,
					newPos, isShift)
			}
		case KeyHome:
			var newPos int
			if glOK {
				byteIdx := runeToByteIndex(text, pos)
				startByte := gl.MoveCursorLineStart(byteIdx)
				if savedTrailing {
					startByte = trailingLineStart(gl.Lines, byteIdx, startByte)
				}
				lineStart := byteToRuneIndex(text, startByte)
				if pos != lineStart {
					newPos = lineStart
				} else {
					paraStart := cursorStartOfParagraph(text, pos)
					if pos != paraStart {
						newPos = paraStart
					} else {
						newPos = cursorHome()
					}
				}
			} else {
				lineStart := moveCursorLineStart([]rune(text), pos)
				if pos != lineStart {
					newPos = lineStart
				} else {
					newPos = cursorHome()
				}
			}
			updateCursorAndSelection(imap, id, is,
				newPos, isShift)
		case KeyEnd:
			var newPos int
			trailingLevel1 := false
			if glOK {
				byteIdx := runeToByteIndex(text, pos)
				endByte := gl.MoveCursorLineEnd(byteIdx)
				if savedTrailing {
					endByte = trailingLineEnd(gl.Lines, byteIdx, endByte)
				}
				lineEnd := byteToRuneIndex(text, endByte)
				if pos != lineEnd {
					newPos = lineEnd
					trailingLevel1 = true
				} else {
					paraEnd := cursorEndOfParagraph(text, pos)
					if pos != paraEnd {
						newPos = paraEnd
					} else {
						newPos = cursorEnd(text)
					}
				}
			} else {
				lineEnd := moveCursorLineEnd([]rune(text), pos)
				if pos != lineEnd {
					newPos = lineEnd
					trailingLevel1 = true
				} else {
					newPos = cursorEnd(text)
				}
			}
			is.CursorTrailing = trailingLevel1
			updateCursorAndSelection(imap, id, is,
				newPos, isShift)
		case KeyUp:
			if hcfg.Mode == InputMultiline {
				var newPos int
				if glOK {
					byteIdx := runeToByteIndex(text, pos)
					preferredX := savedOffset
					if preferredX < 0 {
						if cp, ok := gl.GetCursorPos(byteIdx); ok {
							preferredX = cp.X
						}
					}
					is.CursorOffset = preferredX
					newPos = byteToRuneIndex(text,
						gl.MoveCursorUp(byteIdx, preferredX))
				} else {
					newPos = moveCursorUp([]rune(text), pos)
				}
				updateCursorAndSelection(imap, id, is,
					newPos, isShift)
			}
		case KeyDown:
			if hcfg.Mode == InputMultiline {
				var newPos int
				if glOK {
					byteIdx := runeToByteIndex(text, pos)
					preferredX := savedOffset
					if preferredX < 0 {
						if cp, ok := gl.GetCursorPos(byteIdx); ok {
							preferredX = cp.X
						}
					}
					is.CursorOffset = preferredX
					newPos = byteToRuneIndex(text,
						gl.MoveCursorDown(byteIdx, preferredX))
				} else {
					newPos = moveCursorDown([]rune(text), pos)
				}
				updateCursorAndSelection(imap, id, is,
					newPos, isShift)
			}
		case KeyEnter:
			if hcfg.Mode == InputMultiline {
				text = inputInsert(text, "\n", id, w)
				textChanged = true
			} else {
				commitText := text
				if hcfg.PostCommitNormalize != nil {
					normalized := hcfg.PostCommitNormalize(
						text, CommitEnter)
					if normalized != text {
						commitText = normalized
						if hcfg.OnTextChanged != nil {
							hcfg.OnTextChanged(
								layout, commitText, w)
						}
					}
				}
				if hcfg.OnTextCommit != nil {
					hcfg.OnTextCommit(
						layout, commitText, CommitEnter, w)
				}
				if hcfg.OnEnter != nil {
					hcfg.OnEnter(layout, e, w)
				}
			}
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
				if copied, ok := inputCopy(text, id,
					hcfg.IsPassword, w); ok {
					w.SetClipboard(copied)
				}
			} else {
				handled = false
			}
		case KeyV:
			if e.Modifiers.HasAny(ModCtrl, ModSuper) {
				clip := w.GetClipboard()
				if len(clip) > 0 {
					mask := hcfg.compiledMask()
					if mask != nil {
						cis := inputStateOrDefault(id, w)
						res := InputMaskInsert(text,
							cis.CursorPos,
							cis.SelectBeg,
							cis.SelectEnd, clip, mask)
						if res.Changed {
							text = res.Text
							StateMap[uint32, InputState](
								w, nsInput, capMany,
							).Set(id, InputState{
								CursorPos: res.CursorPos,
								Undo:      cis.Undo,
							})
							textChanged = true
						}
					} else if hcfg.PreTextChange != nil {
						proposed := inputProposedText(
							text, clip, id, w)
						if _, ok := hcfg.PreTextChange(
							text, proposed); ok {
							text = inputInsert(
								text, clip, id, w)
							textChanged = true
						}
					} else {
						text = inputInsert(text, clip, id, w)
						textChanged = true
					}
				}
			} else {
				handled = false
			}
		case KeyX:
			if e.Modifiers.HasAny(ModCtrl, ModSuper) {
				newText, copied, ok := inputCut(text, id,
					hcfg.IsPassword, w)
				if ok {
					w.SetClipboard(copied)
					text = newText
					textChanged = true
				}
			} else {
				handled = false
			}
		case KeyZ:
			if e.Modifiers.HasAny(ModCtrl, ModSuper) {
				if e.Modifiers.Has(ModShift) {
					if nt := inputRedo(text, id, w); nt != text {
						text = nt
						textChanged = true
					}
				} else {
					if nt := inputUndo(text, id, w); nt != text {
						text = nt
						textChanged = true
					}
				}
			} else {
				handled = false
			}
		case KeyBackspace:
			mask := hcfg.compiledMask()
			if mask != nil {
				res := InputMaskBackspace(text, is.CursorPos,
					is.SelectBeg, is.SelectEnd, mask)
				if res.Changed {
					text = res.Text
					imap.Set(id, InputState{
						CursorPos: res.CursorPos, Undo: is.Undo,
					})
					textChanged = true
				}
			} else {
				newText, delChanged := inputDeleteGrapheme(
					text, id, false, layout, w,
				)
				if delChanged {
					text = newText
					textChanged = true
				}
			}
		case KeyDelete:
			mask := hcfg.compiledMask()
			if mask != nil {
				res := InputMaskDelete(text, is.CursorPos,
					is.SelectBeg, is.SelectEnd, mask)
				if res.Changed {
					text = res.Text
					imap.Set(id, InputState{
						CursorPos: res.CursorPos, Undo: is.Undo,
					})
					textChanged = true
				}
			} else {
				newText, delChanged := inputDeleteGrapheme(
					text, id, true, layout, w,
				)
				if delChanged {
					text = newText
					textChanged = true
				}
			}
		default:
			handled = false
		}

		if handled {
			resetBlinkCursorVisible(w)
			if textChanged && hcfg.OnTextChanged != nil {
				hcfg.OnTextChanged(layout, text, w)
			}
			inputScrollCursorIntoView(
				hcfg.IDScroll, text, layout, w,
			)
			e.IsHandled = true
		} else if hcfg.OnKeyDown != nil {
			hcfg.OnKeyDown(layout, e, w)
		}
	}
}

// inputTextFromLayout extracts the current text from the input's
// inner layout structure (Column → Row → Text).
func inputTextFromLayout(layout *Layout) string {
	if len(layout.Children) == 0 {
		return ""
	}
	row := &layout.Children[0]
	if len(row.Children) == 0 {
		return ""
	}
	txt := &row.Children[0]
	if txt.Shape.TC == nil {
		return ""
	}
	if txt.Shape.TC.TextIsPlaceholder {
		return ""
	}
	return txt.Shape.TC.Text
}

// inputGlyphLayoutFor navigates to the inner text shape of an
// input layout and returns a glyph Layout for cursor navigation.
func inputGlyphLayoutFor(layout *Layout, w *Window) (glyph.Layout, bool) {
	if w.textMeasurer == nil {
		return glyph.Layout{}, false
	}
	if len(layout.Children) == 0 {
		return glyph.Layout{}, false
	}
	row := &layout.Children[0]
	if len(row.Children) == 0 {
		return glyph.Layout{}, false
	}
	txt := &row.Children[0]
	if txt.Shape == nil || txt.Shape.TC == nil {
		return glyph.Layout{}, false
	}
	style := textStyleOrDefault(txt.Shape)
	return inputGlyphLayout(
		inputTextFromLayout(layout), txt.Shape, style, w,
	)
}

// trailingLineStart returns the start of the previous visual line
// when byteIdx is at a soft-wrap boundary and CursorTrailing is set.
// Falls back to fallback if no boundary match is found.
func trailingLineStart(lines []glyph.Line, byteIdx, fallback int) int {
	for i, line := range lines {
		if i > 0 && byteIdx == line.StartIndex {
			return lines[i-1].StartIndex
		}
	}
	return fallback
}

// trailingLineEnd returns the end of the previous visual line
// when byteIdx is at a soft-wrap boundary and CursorTrailing is set.
// Falls back to fallback if no boundary match is found.
func trailingLineEnd(lines []glyph.Line, byteIdx, fallback int) int {
	for i, line := range lines {
		if i > 0 && byteIdx == line.StartIndex {
			return lines[i-1].StartIndex + lines[i-1].Length
		}
	}
	return fallback
}

// inputDeleteGrapheme deletes a grapheme cluster at cursor using
// glyph when available, falling back to rune-based inputDelete.
func inputDeleteGrapheme(
	text string, idFocus uint32, forward bool,
	layout *Layout, w *Window,
) (string, bool) {
	gl, glOK := inputGlyphLayoutFor(layout, w)
	if !glOK {
		newText, _ := inputDelete(text, idFocus, forward, w)
		return newText, newText != text
	}
	is := inputStateOrDefault(idFocus, w)
	if is.SelectBeg != is.SelectEnd {
		newText, _ := inputDelete(text, idFocus, forward, w)
		return newText, newText != text
	}
	pos := min(is.CursorPos, utf8RuneCount(text))
	byteIdx := runeToByteIndex(text, pos)
	var res glyph.MutationResult
	if forward {
		res = glyph.DeleteForward(text, gl, byteIdx)
	} else {
		res = glyph.DeleteBackward(text, gl, byteIdx)
	}
	if res.NewText == text {
		return text, false
	}
	newPos := byteToRuneIndex(res.NewText, res.CursorPos)
	undo := inputPushUndo(is, text)
	imap := StateMap[uint32, InputState](w, nsInput, capMany)
	imap.Set(idFocus, InputState{
		CursorPos:    newPos,
		CursorOffset: -1,
		Undo:         undo,
	})
	return res.NewText, true
}

// passwordMask replaces each rune with a bullet character.
func passwordMask(text string) string {
	runes := []rune(text)
	for i := range runes {
		runes[i] = '•'
	}
	return string(runes)
}
