package gui

// view_rtf.go defines the Rich Text Format (RTF) view.
// Renders text with multiple typefaces, sizes, and styles.
// Supports text wrapping, clickable links, and custom runs.

import (
	"math"
	"time"

	"github.com/mike-ward/go-glyph"

	"github.com/mike-ward/go-gui/gui/markdown"
)

// RtfCfg configures a Rich Text View.
type RtfCfg struct {
	ID              string
	A11YLabel       string
	A11YDescription string
	RichText        RichText
	MinWidth        float32
	IDFocus         uint32
	Mode            TextMode
	Invisible       bool
	Clip            bool
	FocusSkip       bool
	Disabled        bool
	HangingIndent   float32 // negative indent for wrapped lines
	BaseTextStyle   *TextStyle
}

type rtfView struct {
	RtfCfg
	sizing Sizing
}

func (v *rtfView) Content() []View { return nil }

// rtfSuppressInlineObjectGlyphs prevents object placeholder glyphs from
// painting when a later render pass draws the actual inline object.
func rtfSuppressInlineObjectGlyphs(layout *glyph.Layout) {
	if layout == nil {
		return
	}
	for i := range layout.Items {
		if !layout.Items[i].IsObject {
			continue
		}
		layout.Items[i].GlyphCount = 0
	}
}

func (v *rtfView) GenerateLayout(w *Window) Layout {
	// Convert RichText to glyph.RichText.
	vgRT, mathHashes := v.RichText.toGlyphRichTextWithMath(
		w.viewState.diagramCache)

	// Determine base style.
	var baseStyle glyph.TextStyle
	if v.BaseTextStyle != nil {
		baseStyle = v.BaseTextStyle.ToGlyphStyle()
	} else if len(vgRT.Runs) > 0 {
		baseStyle = vgRT.Runs[0].Style
	}

	// For wrapped modes, skip the initial LayoutRichText — Width is
	// overridden by Fill sizing and Height by layoutWrapRTF. The
	// expensive glyph shaping runs once in layoutWrapRTF instead.
	isWrap := v.Mode == TextModeWrap ||
		v.Mode == TextModeWrapKeepSpaces

	var layout glyph.Layout
	if !isWrap {
		cfg := glyph.TextConfig{
			Style: baseStyle,
			Block: glyph.BlockStyle{
				Wrap:   glyph.WrapWord,
				Width:  -1.0,
				Indent: -v.HangingIndent,
			},
		}
		if w.textMeasurer != nil {
			if tm, ok := w.textMeasurer.(interface {
				LayoutRichText(glyph.RichText, glyph.TextConfig) (glyph.Layout, error)
			}); ok {
				if l, err := tm.LayoutRichText(vgRT, cfg); err == nil {
					layout = l
					rtfSuppressInlineObjectGlyphs(&layout)
				}
			}
		}
	}

	shape := &Shape{
		ShapeType: ShapeRTF,
		ID:        v.ID,
		IDFocus:   v.IDFocus,
		A11YRole:  AccessRoleStaticText,
		A11Y:      makeA11YInfo(v.A11YLabel, v.A11YDescription),
		Width:     layout.Width,
		Height:    layout.Height,
		Clip:      v.Clip,
		FocusSkip: v.FocusSkip,
		Disabled:  v.Disabled,
		MinWidth:  v.MinWidth,
		Sizing:    v.sizing,
		Events: &EventHandlers{
			OnClick:     rtfOnClick,
			OnMouseMove: rtfMouseMove,
			AmendLayout: rtfAmendTooltip,
		},
		TC: &ShapeTextConfig{
			TextMode:      v.Mode,
			HangingIndent: v.HangingIndent,
			RtfBaseStyle:  baseStyle,
			RtfLayout:     &layout,
			RtfRuns:       &v.RichText,
			rtfGlyphRT:    &vgRT,
			rtfMathHashes: mathHashes,
		},
	}
	l := Layout{Shape: shape}
	blockKey := rtfRunsKey(shape.TC.RtfRuns)
	if ts := &w.viewState.tooltip; ts.id != "" &&
		ts.text != "" && ts.blockKey != 0 &&
		blockKey == ts.blockKey {
		l.Children = []Layout{
			GenerateViewLayout(rtfTooltipView(ts), w),
		}
	}
	// Link context menu popup — only on the owning RTF block.
	if st := StateReadOr[string, rtfLinkMenuState](
		w, nsRtfLinkMenu, nsRtfLinkMenu,
		rtfLinkMenuState{}); st.Open &&
		st.BlockKey == blockKey {
		l.Children = append(l.Children,
			GenerateViewLayout(rtfLinkMenuView(w, st), w))
	}
	return l
}

// RTF creates a rich text view.
func RTF(cfg RtfCfg) View {
	if cfg.Invisible {
		return invisibleContainerView()
	}
	sizing := FitFit
	if cfg.Mode == TextModeWrap ||
		cfg.Mode == TextModeWrapKeepSpaces {
		sizing = FillFit
	}
	return &rtfView{RtfCfg: cfg, sizing: sizing}
}

// --- Hit testing ---

func rtfRunRect(run glyph.Item) DrawClip {
	return DrawClip{
		X:      float32(run.X),
		Y:      float32(run.Y - run.Ascent),
		Width:  float32(run.Width),
		Height: float32(run.Ascent + run.Descent),
	}
}

func rtfHitTest(run glyph.Item, mx, my float32) bool {
	r := rtfRunRect(run)
	return mx >= r.X && my >= r.Y &&
		mx < r.X+r.Width && my < r.Y+r.Height
}

func rtfFindRunAtIndex(
	l *Layout, startIndex int,
) RichTextRun {
	if l.Shape == nil || l.Shape.TC == nil ||
		l.Shape.TC.RtfRuns == nil {
		return RichTextRun{}
	}
	idx := 0
	for _, r := range l.Shape.TC.RtfRuns.Runs {
		runLen := len(r.Text)
		if startIndex >= idx &&
			startIndex < idx+runLen {
			return r
		}
		idx += runLen
	}
	return RichTextRun{}
}

// --- Event handlers ---

func rtfMouseMove(l *Layout, e *Event, w *Window) {
	if !l.Shape.HasRtfLayout() {
		return
	}
	ts := &w.viewState.tooltip
	layout := l.Shape.TC.RtfLayout
	for _, run := range layout.Items {
		if run.IsObject {
			continue
		}
		if rtfHitTest(run, e.MouseX, e.MouseY) {
			found := rtfFindRunAtIndex(l, run.StartIndex)
			if found.Tooltip != "" {
				tipID := found.Tooltip
				if ts.hoverID == tipID {
					e.IsHandled = true
					return
				}
				r := rtfRunRect(run)
				ts.hoverID = tipID
				ts.text = found.Tooltip
				ts.bounds = DrawClip{
					X:      l.Shape.X + r.X,
					Y:      l.Shape.Y + r.Y,
					Width:  r.Width,
					Height: r.Height,
				}
				ts.floatOffsetX = r.X + r.Width/2
				ts.floatOffsetY = r.Y - 3
				ts.blockKey = rtfRunsKey(
					l.Shape.TC.RtfRuns)
				ts.hoverStart = time.Now()
				w.AnimationAdd(rtfTooltipAnimation(tipID))
				e.IsHandled = true
				return
			}
			if found.Link != "" {
				w.SetMouseCursorPointingHand()
				e.IsHandled = true
				return
			}
		}
	}
	ts.clearText()
}

// rtfTooltipAnimation returns an Animate that activates
// the RTF tooltip after the configured delay.
func rtfTooltipAnimation(tipID string) *Animate {
	return &Animate{
		AnimID: "___tooltip___",
		Delay:     DefaultTooltipStyle.Delay,
		Callback: func(_ *Animate, w *Window) {
			ts := &w.viewState.tooltip
			if ts.hoverID == tipID && ts.text != "" {
				ts.id = tipID
				ts.popupID = tipID + "_rtf_popup"
			}
		},
	}
}

// rtfAmendTooltip clears RTF tooltip state when the mouse
// leaves the stored bounds, and dismisses the link context
// menu when focus is lost.
func rtfAmendTooltip(_ *Layout, w *Window) {
	ts := &w.viewState.tooltip
	if ts.text != "" {
		mx := w.viewState.mousePosX
		my := w.viewState.mousePosY
		b := ts.bounds
		if mx < b.X || my < b.Y ||
			mx >= b.X+b.Width || my >= b.Y+b.Height {
			ts.clearText()
		}
	}
	// Dismiss link context menu when focus moves away.
	if !w.IsFocus(rtfLinkMenuIDFocus) {
		sm := StateMapRead[string, rtfLinkMenuState](
			w, nsRtfLinkMenu)
		if sm != nil {
			sm.Delete(nsRtfLinkMenu)
		}
	}
}

// rtfRunsKey computes an FNV-1a hash of RichText content
// including Link, Tooltip, MathID, and MathLatex for
// tooltip/menu block matching and cross-frame caching.
func rtfRunsKey(rt *RichText) uint64 {
	h := uint64(14695981039346656037)
	for _, r := range rt.Runs {
		for i := 0; i < len(r.Text); i++ {
			h ^= uint64(r.Text[i])
			h *= 1099511628211
		}
		h ^= 0x1F // field separator
		h *= 1099511628211
		for i := 0; i < len(r.Link); i++ {
			h ^= uint64(r.Link[i])
			h *= 1099511628211
		}
		h ^= 0x1F
		h *= 1099511628211
		for i := 0; i < len(r.Tooltip); i++ {
			h ^= uint64(r.Tooltip[i])
			h *= 1099511628211
		}
		h ^= 0x1F
		h *= 1099511628211
		for i := 0; i < len(r.MathID); i++ {
			h ^= uint64(r.MathID[i])
			h *= 1099511628211
		}
		h ^= 0x1F
		h *= 1099511628211
		for i := 0; i < len(r.MathLatex); i++ {
			h ^= uint64(r.MathLatex[i])
			h *= 1099511628211
		}
		h ^= 0x1F // run separator
		h *= 1099511628211
	}
	return h
}

// rtfStyleKey hashes layout-affecting fields of a base style
// for use in the cross-frame RTF layout cache key.
func rtfStyleKey(s glyph.TextStyle) uint64 {
	h := uint64(14695981039346656037)
	for i := 0; i < len(s.FontName); i++ {
		h ^= uint64(s.FontName[i])
		h *= 1099511628211
	}
	h ^= uint64(s.Typeface)
	h *= 1099511628211
	h ^= uint64(math.Float32bits(s.Size))
	h *= 1099511628211
	h ^= uint64(math.Float32bits(s.LetterSpacing))
	h *= 1099511628211
	return h
}

// rtfTooltipView builds a floating tooltip popup positioned
// relative to the owning RTF shape via the float system.
func rtfTooltipView(ts *tooltipState) View {
	d := &DefaultTooltipStyle
	return Column(ContainerCfg{
		ID:            ts.popupID,
		Float:         true,
		FloatAutoFlip: true,
		FloatTieOff:   FloatBottomCenter,
		FloatOffsetX:  ts.floatOffsetX,
		FloatOffsetY:  ts.floatOffsetY,
		Color:         d.Color,
		ColorBorder:   d.ColorBorder,
		SizeBorder:    Some(d.SizeBorder),
		Radius:        Some(d.Radius),
		Padding:       Some(d.Padding),
		MaxWidth:      300,
		Content: []View{
			Text(TextCfg{
				Text:      ts.text,
				TextStyle: d.TextStyle,
				Mode:      TextModeWrap,
			}),
		},
	})
}

func rtfOnClick(l *Layout, e *Event, w *Window) {
	if !l.Shape.HasRtfLayout() {
		return
	}
	layout := l.Shape.TC.RtfLayout
	for _, run := range layout.Items {
		if run.IsObject {
			continue
		}
		if rtfHitTest(run, e.MouseX, e.MouseY) {
			found := rtfFindRunAtIndex(l, run.StartIndex)
			if found.Link != "" && markdown.IsSafeURL(found.Link) {
				if e.MouseButton == MouseRight {
					showLinkContextMenu(w, found.Link,
						e.MouseX,
						e.MouseY,
						rtfRunsKey(l.Shape.TC.RtfRuns))
					e.IsHandled = true
					return
				}
				if len(found.Link) > 0 &&
					found.Link[0] == '#' {
					w.ScrollToView(found.Link[1:])
				} else if w.nativePlatform != nil {
					_ = w.nativePlatform.OpenURI(found.Link)
				}
				e.IsHandled = true
			}
			return
		}
	}
}

// rtfLinkMenuState holds state for the RTF link context menu.
type rtfLinkMenuState struct {
	Open     bool
	Link     string
	X        float32
	Y        float32
	BlockKey uint64 // identifies the owning RTF block
}

const rtfLinkMenuIDFocus uint32 = 8492137

// showLinkContextMenu opens a context menu for an RTF link.
func showLinkContextMenu(
	w *Window, link string, mx, my float32,
	blockKey uint64,
) {
	sm := StateMap[string, rtfLinkMenuState](
		w, nsRtfLinkMenu, capFew)
	sm.Set(nsRtfLinkMenu, rtfLinkMenuState{
		Open:     true,
		Link:     link,
		X:        mx,
		Y:        my,
		BlockKey: blockKey,
	})
	w.SetIDFocus(rtfLinkMenuIDFocus)
}

// rtfLinkMenuDismiss clears the link context menu state.
func rtfLinkMenuDismiss(w *Window) {
	sm := StateMapRead[string, rtfLinkMenuState](
		w, nsRtfLinkMenu)
	if sm != nil {
		sm.Delete(nsRtfLinkMenu)
	}
	w.SetIDFocus(0)
}

// rtfLinkMenuView builds the floating context menu popup
// for RTF link right-click.
func rtfLinkMenuView(w *Window, st rtfLinkMenuState) View {
	link := st.Link
	return Menu(w, MenubarCfg{
		ID:      "rtf_link_menu",
		IDFocus: rtfLinkMenuIDFocus,
		Items: []MenuItemCfg{
			{ID: "open_link", Text: "Open Link"},
			{ID: "copy_link", Text: "Copy Link"},
		},
		Action: func(id string, _ *Event, w *Window) {
			switch id {
			case "open_link":
				if w.nativePlatform != nil &&
					markdown.IsSafeURL(link) {
					_ = w.nativePlatform.OpenURI(link)
				}
			case "copy_link":
				w.SetClipboard(link)
			}
			rtfLinkMenuDismiss(w)
		},
		Float:         true,
		FloatAutoFlip: true,
		FloatAnchor:   FloatTopLeft,
		FloatTieOff:   FloatTopLeft,
		FloatOffsetX:  st.X,
		FloatOffsetY:  st.Y,
	})
}
