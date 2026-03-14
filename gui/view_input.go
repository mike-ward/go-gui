package gui

import (
	"log"
	"strings"
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
	SpellCheck  bool
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
	// PreTextChange is called before text changes. Return (adjusted, true)
	// to accept (adjusted may differ from proposed), or ("", false) to
	// reject. Undo/redo bypass this callback by design — if security
	// invariants (max length, forbidden chars) must be enforced
	// unconditionally, use OnTextChanged instead.
	PreTextChange func(current, proposed string) (string, bool)
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
	colorHover := cfg.ColorHover
	idFocus := cfg.IDFocus
	spellChk := cfg.SpellCheck && !cfg.IsPassword
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
	hcfg.CompiledMask = hcfg.compiledMask()

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

			var runes []rune
			if doubleClick {
				runes = []rune(displayText)
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
			ds := &inputDragState{
				anchorPos:   is.SelectBeg,
				anchorEnd:   is.SelectEnd,
				gl:          gl,
				displayText: displayText,
				txtOffX:     ly.Shape.X - layout.Shape.X,
				txtOffY:     ly.Shape.Y - layout.Shape.Y,
				idFocus:     ly.Shape.IDFocus,
				idScroll:    idScroll,
			}
			if doubleClick {
				ds.runes = runes
			}
			if idScroll > 0 && layout.Parent != nil {
				sy := StateMap[uint32, float32](
					w, nsScrollY, capScroll)
				ds.scrollY0, _ = sy.Get(idScroll)
				p := layout.Parent.Shape
				ds.viewTop = p.Y + p.Padding.Top
				viewH := p.Height - p.PaddingHeight()
				ds.viewBot = ds.viewTop + viewH
				ds.maxScrollNeg = f32Min(0,
					viewH-layout.Shape.Height)
			}
			startInputDrag(ds, w)
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
		OnHover: func(layout *Layout, _ *Event, w *Window) {
			w.SetMouseCursor(CursorIBeam)
			if !w.IsFocus(idFocus) {
				layout.Shape.Color = colorHover
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

			// Spell check debounce trigger.
			if spellChk && focused {
				text := inputTextFromLayout(layout)
				spellCheckTrigger(
					layout.Shape.IDFocus, text, w)
			}
		},
		Content: []View{inner},
	})
}

func applyInputDefaults(cfg *InputCfg) {
	d := &DefaultInputStyle
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
	CompiledMask        *CompiledInputMask
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
		log.Printf("input: mask compile failed: %v", err)
		return nil
	}
	return &c
}

func makeInputOnChar(hcfg inputHandlerCfg) func(*Layout, *Event, *Window) {
	mask := hcfg.CompiledMask
	return func(layout *Layout, e *Event, w *Window) {
		if hcfg.IDFocus == 0 || !w.IsFocus(hcfg.IDFocus) {
			return
		}
		ch := e.CharCode
		id := hcfg.IDFocus

		// Control characters are handled by OnKeyDown.
		if ch < CharSpace {
			e.IsHandled = true
			return
		}

		text := inputTextFromLayout(layout)
		changed := false

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
			if adjusted, ok := hcfg.PreTextChange(text, proposed); ok {
				if adjusted == proposed {
					text = inputInsert(text, ins, id, w)
				} else {
					inputSetTextAndCursorAtEnd(
						text, adjusted, id, w)
					text = adjusted
				}
				changed = true
			}
		} else {
			text = inputInsert(text, ins, id, w)
			changed = true
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
	mask := hcfg.CompiledMask
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
		gl, glOK := inputGlyphLayoutWithText(text, layout, w)

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
			} else {
				handled = false
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
			} else {
				handled = false
			}
		case KeyEnter:
			if hcfg.Mode == InputMultiline {
				text = inputInsert(text, "\n", id, w)
				textChanged = true
			} else {
				inputCommitEnter(hcfg, layout, text, e, w)
			}
		case KeyEscape:
			is.SelectBeg = 0
			is.SelectEnd = 0
			imap.Set(id, is)
			handled = false
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
						if adjusted, ok := hcfg.PreTextChange(
							text, proposed); ok {
							if adjusted == proposed {
								text = inputInsert(
									text, clip, id, w)
							} else {
								inputSetTextAndCursorAtEnd(
									text, adjusted, id, w)
								text = adjusted
							}
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
			if newText, ok := inputHandleDelete(
				text, id, false, mask, layout, w,
			); ok {
				text = newText
				textChanged = true
			}
		case KeyDelete:
			if newText, ok := inputHandleDelete(
				text, id, true, mask, layout, w,
			); ok {
				text = newText
				textChanged = true
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

// inputCommitEnter handles single-line Enter: normalize, commit,
// fire OnEnter.
func inputCommitEnter(
	hcfg inputHandlerCfg,
	layout *Layout, text string, e *Event, w *Window,
) {
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

// inputHandleDelete handles Backspace/Delete for both masked and
// unmasked inputs.
func inputHandleDelete(
	text string, id uint32, forward bool,
	mask *CompiledInputMask,
	layout *Layout, w *Window,
) (string, bool) {
	if mask != nil {
		is := inputStateOrDefault(id, w)
		var res MaskEditResult
		if forward {
			res = InputMaskDelete(text, is.CursorPos,
				is.SelectBeg, is.SelectEnd, mask)
		} else {
			res = InputMaskBackspace(text, is.CursorPos,
				is.SelectBeg, is.SelectEnd, mask)
		}
		if !res.Changed {
			return text, false
		}
		StateMap[uint32, InputState](
			w, nsInput, capMany,
		).Set(id, InputState{
			CursorPos: res.CursorPos, Undo: is.Undo,
		})
		return res.Text, true
	}
	return inputDeleteGrapheme(text, id, forward, layout, w)
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
	return inputGlyphLayoutWithText(
		inputTextFromLayout(layout), layout, w,
	)
}

// inputGlyphLayoutWithText returns a glyph Layout using
// pre-extracted text, avoiding redundant layout traversal.
func inputGlyphLayoutWithText(
	text string, layout *Layout, w *Window,
) (glyph.Layout, bool) {
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
	return inputGlyphLayout(text, txt.Shape, style, w)
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
// Uses a stack-local buffer for short strings to avoid heap
// allocation from strings.Repeat.
func passwordMask(text string) string {
	n := utf8RuneCount(text)
	// "•" (U+2022) = 3 bytes UTF-8: 0xE2 0x80 0xA2
	const bLen = 3
	if n <= 64 {
		var buf [64 * bLen]byte
		for i := range n {
			buf[i*bLen] = 0xe2
			buf[i*bLen+1] = 0x80
			buf[i*bLen+2] = 0xa2
		}
		return string(buf[:n*bLen])
	}
	return strings.Repeat("•", n)
}

// inputDragState holds state for drag-to-select in an input.
// Replaces ~10 closure-captured locals with explicit fields;
// runes is non-nil iff the drag started from a double-click.
type inputDragState struct {
	anchorPos, anchorEnd   uint32
	gl                     glyph.Layout
	displayText            string
	txtOffX, txtOffY       float32
	idFocus                uint32
	idScroll               uint32
	lastMouseX, lastMouseY float32
	scrollY0               float32
	viewTop, viewBot       float32
	maxScrollNeg           float32
	runes                  []rune // non-nil = double-click word select
}

func (d *inputDragState) computeRunePos(
	mx, my float32, w *Window,
) int {
	scrollDelta := float32(0)
	if d.idScroll > 0 {
		sy := StateMap[uint32, float32](
			w, nsScrollY, capScroll)
		sNow, _ := sy.Get(d.idScroll)
		scrollDelta = sNow - d.scrollY0
	}
	relX := mx - d.txtOffX
	relY := my - (d.txtOffY + scrollDelta)
	byteIdx := d.gl.GetClosestOffset(relX, relY)
	return byteToRuneIndex(d.displayText, byteIdx)
}

func (d *inputDragState) updateSelection(rp int, w *Window) {
	imap := StateMap[uint32, InputState](w, nsInput, capMany)
	is, _ := imap.Get(d.idFocus)
	if d.runes != nil {
		wb, we := wordBoundsAt(d.runes, rp)
		if rp < int(d.anchorPos) {
			is.SelectBeg = d.anchorEnd
			is.SelectEnd = uint32(wb)
			is.CursorPos = wb
		} else {
			is.SelectBeg = d.anchorPos
			is.SelectEnd = uint32(we)
			is.CursorPos = we
		}
	} else {
		is.CursorPos = rp
		is.SelectBeg = d.anchorPos
		is.SelectEnd = uint32(rp)
	}
	is.CursorOffset = -1
	imap.Set(d.idFocus, is)
	resetBlinkCursorVisible(w)
}

func (d *inputDragState) scrollCallback(
	_ *Animate, w *Window,
) {
	var delta float32
	if d.lastMouseY < d.viewTop {
		delta = (d.viewTop - d.lastMouseY) * 0.3
	} else if d.lastMouseY > d.viewBot {
		delta = -((d.lastMouseY - d.viewBot) * 0.3)
	} else {
		w.AnimationRemove(animIDDragScroll)
		return
	}
	sy := StateMap[uint32, float32](w, nsScrollY, capScroll)
	cur, _ := sy.Get(d.idScroll)
	newScroll := f32Clamp(cur+delta, d.maxScrollNeg, 0)
	if newScroll == cur {
		return
	}
	sy.Set(d.idScroll, newScroll)
	rp := d.computeRunePos(d.lastMouseX, d.lastMouseY, w)
	d.updateSelection(rp, w)
}

// startInputDrag sets up MouseLock drag-to-select for an input.
func startInputDrag(d *inputDragState, w *Window) {
	w.MouseLock(MouseLockCfg{
		MouseMove: func(_ *Layout, e *Event, w *Window) {
			d.lastMouseX = e.MouseX
			d.lastMouseY = e.MouseY
			rp := d.computeRunePos(e.MouseX, e.MouseY, w)
			d.updateSelection(rp, w)
			if d.idScroll > 0 {
				outside := e.MouseY < d.viewTop ||
					e.MouseY > d.viewBot
				if outside && !w.HasAnimation(
					animIDDragScroll) {
					w.AnimationAdd(&Animate{
						AnimID: animIDDragScroll,
						Delay:     32 * time.Millisecond,
						Repeat:    true,
						Refresh:   AnimationRefreshLayout,
						Callback:  d.scrollCallback,
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
}
