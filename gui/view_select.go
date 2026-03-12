package gui

import "strings"

// SelectCfg configures a select (dropdown) view.
type SelectCfg struct {
	ID               string
	Placeholder      string
	Selected         []string // currently selected option text(s)
	Options          []string
	Color            Color
	ColorBorder      Color
	ColorBorderFocus Color
	ColorFocus       Color
	ColorSelect      Color
	Padding          Opt[Padding]
	SizeBorder       Opt[float32]
	TextStyle        TextStyle
	SubheadingStyle  TextStyle
	PlaceholderStyle TextStyle
	OnSelect         func([]string, *Event, *Window)
	MinWidth         float32
	MaxWidth         float32
	Radius           Opt[float32]
	IDFocus          uint32
	SelectMultiple   bool
	NoWrap           bool
	Sizing           Sizing
	FloatZIndex      int
	Disabled         bool
	Invisible        bool

	A11YLabel       string
	A11YDescription string
}

// selectView implements View for select (dropdown).
type selectView struct {
	cfg SelectCfg
}

// Select creates a select (dropdown) view.
func Select(cfg SelectCfg) View {
	applySelectDefaults(&cfg)
	return &selectView{cfg: cfg}
}

func (sv *selectView) Content() []View { return nil }

func (sv *selectView) GenerateLayout(w *Window) Layout {
	cfg := &sv.cfg
	dn := &DefaultSelectStyle
	sizeBorder := cfg.SizeBorder.Get(dn.SizeBorder)
	radius := cfg.Radius.Get(dn.Radius)
	isOpen := StateReadOr[string, bool](w, nsSelect, cfg.ID, false)
	idScroll := fnvSum32(cfg.ID + ".dropdown")

	empty := len(cfg.Selected) == 0 || len(cfg.Selected[0]) == 0
	clip := cfg.SelectMultiple && cfg.NoWrap

	txt := cfg.Placeholder
	if !empty {
		txt = strings.Join(cfg.Selected, ", ")
	}
	txtStyle := cfg.PlaceholderStyle
	if !empty {
		txtStyle = cfg.TextStyle
	}
	wrapMode := TextModeSingleLine
	if cfg.SelectMultiple && !cfg.NoWrap {
		wrapMode = TextModeWrap
	}

	arrowText := "▼"
	if isOpen {
		arrowText = "▲"
	}

	spacerSizing := FillFill
	if wrapMode != TextModeSingleLine {
		spacerSizing = FitFill
	}

	content := make([]View, 0, 4)
	content = append(content,
		Text(TextCfg{
			Text:      txt,
			TextStyle: txtStyle,
			Mode:      wrapMode,
		}),
		Row(ContainerCfg{
			Sizing:  spacerSizing,
			Padding: NoPadding,
		}),
		Text(TextCfg{
			Text:      arrowText,
			TextStyle: cfg.TextStyle,
		}),
	)

	if isOpen {
		highlightedIdx := StateReadOr[string, int](
			w, nsSelectHL, cfg.ID, 0)
		options := make([]View, 0, len(cfg.Options))
		for i, option := range cfg.Options {
			if strings.HasPrefix(option, "---") {
				options = append(options,
					selectSubHeaderView(cfg, option))
			} else {
				options = append(options,
					selectOptionView(cfg, option, i,
						i == highlightedIdx))
			}
		}
		content = append(content, Column(ContainerCfg{
			ID:            cfg.ID + ".dropdown",
			SizeBorder:    Some(sizeBorder),
			Radius:        Some(radius),
			ColorBorder:   cfg.ColorBorder,
			Color:         cfg.Color,
			MinHeight:     50,
			MaxHeight:     200,
			MinWidth:      cfg.MinWidth,
			MaxWidth:      cfg.MaxWidth,
			Float:         true,
			FloatAutoFlip: true,
			FloatAnchor:   FloatBottomLeft,
			FloatTieOff:   FloatTopLeft,
			FloatOffsetY:  -sizeBorder,
			FloatZIndex:   cfg.FloatZIndex,
			IDScroll:      idScroll,
			Padding: SomeP(
				PadSmall, PadMedium, PadSmall, PadSmall),
			Spacing: NoSpacing,
			Content: options,
		}))
	}

	id := cfg.ID
	colorFocus := cfg.ColorFocus
	colorBorderFocus := cfg.ColorBorderFocus

	// Build the outer row layout directly.
	cv := &containerView{
		cfg: ContainerCfg{
			ID:          cfg.ID,
			IDFocus:     cfg.IDFocus,
			Clip:        clip,
			A11YRole:    AccessRoleComboBox,
			A11YLabel:   a11yLabel(cfg.A11YLabel, cfg.Placeholder),
			Color:       cfg.Color,
			ColorBorder: cfg.ColorBorder,
			SizeBorder:  Some(sizeBorder),
			Radius:      Some(radius),
			Padding:     cfg.Padding,
			Sizing:      cfg.Sizing,
			MinWidth:    cfg.MinWidth,
			MaxWidth:    cfg.MaxWidth,
			Disabled:    cfg.Disabled,
			Invisible:   cfg.Invisible,
			axis:        AxisLeftToRight,
			AmendLayout: func(layout *Layout, w *Window) {
				if layout.Shape.Disabled {
					return
				}
				if w.IsFocus(layout.Shape.IDFocus) {
					layout.Shape.Color = colorFocus
					layout.Shape.ColorBorder = colorBorderFocus
				}
			},
			OnKeyDown: makeSelectOnKeyDown(sv.cfg, idScroll),
			OnClick: func(_ *Layout, e *Event, w *Window) {
				ss := StateMap[string, bool](
					w, nsSelect, capModerate)
				cur, _ := ss.Get(id)
				ss.Clear()
				ss.Set(id, !cur)
				e.IsHandled = true
			},
			Opacity: 1.0,
		},
		content:   content,
		shapeType: ShapeRectangle,
	}
	// Resolve click handler.
	cv.cfg.OnClick = leftClickOnly(cv.cfg.OnClick)
	return GenerateViewLayout(cv, w)
}

// selectOptionView builds a single option row.
func selectOptionView(cfg *SelectCfg, option string, index int, highlighted bool) View {
	selectMultiple := cfg.SelectMultiple
	onSelect := cfg.OnSelect
	selectArray := cfg.Selected
	colorSelect := cfg.ColorSelect
	cfgID := cfg.ID

	optColor := ColorTransparent
	if highlighted {
		optColor = cfg.ColorSelect
	}

	checkColor := ColorTransparent
	for _, s := range cfg.Selected {
		if s == option {
			checkColor = cfg.TextStyle.Color
			break
		}
	}

	return Row(ContainerCfg{
		Color:   optColor,
		Padding: SomeP(0, PadSmall, 0, 1),
		Sizing:  FillFit,
		Content: []View{
			Row(ContainerCfg{
				Padding: Some(PadTBLR(2, 0)),
				Content: []View{
					Text(TextCfg{
						Text: "✓",
						TextStyle: TextStyle{
							Color: checkColor,
							Size:  cfg.TextStyle.Size,
						},
					}),
					Text(TextCfg{
						Text:      option,
						TextStyle: cfg.TextStyle,
					}),
				},
			}),
		},
		OnClick: func(_ *Layout, e *Event, w *Window) {
			if onSelect == nil {
				return
			}
			ss := StateMap[string, bool](
				w, nsSelect, capModerate)
			if !selectMultiple {
				ss.Clear()
			}
			var s []string
			if selectMultiple {
				s = listBoxNextSelectedIDs(
					selectArray, option, true)
			} else {
				ss.Clear()
				s = []string{option}
			}
			onSelect(s, e, w)
			e.IsHandled = true
		},
		OnHover: func(layout *Layout, _ *Event, w *Window) {
			w.SetMouseCursor(CursorPointingHand)
			layout.Shape.Color = colorSelect
			sh := StateMap[string, int](
				w, nsSelectHL, capModerate)
			cur, _ := sh.Get(cfgID)
			if cur != index {
				sh.Set(cfgID, index)
			}
		},
	})
}

// selectSubHeaderView builds a section header row.
func selectSubHeaderView(cfg *SelectCfg, option string) View {
	label := option
	if len(option) > 3 {
		label = option[3:]
	}
	return Column(ContainerCfg{
		Padding: SomeP(guiTheme.PaddingMedium.Top, 0, 0, 0),
		Sizing:  FillFit,
		Content: []View{
			Row(ContainerCfg{
				Padding: NoPadding,
				Sizing:  FillFit,
				Spacing: Some[float32](PadXSmall),
				Content: []View{
					Text(TextCfg{
						Text: "✓",
						TextStyle: TextStyle{
							Color: ColorTransparent,
							Size:  cfg.SubheadingStyle.Size,
						},
					}),
					Text(TextCfg{
						Text:      label,
						TextStyle: cfg.SubheadingStyle,
					}),
				},
			}),
			Row(ContainerCfg{
				Padding: Some(PadTBLR(0, PadMedium)),
				Sizing:  FillFit,
				Content: []View{
					Rectangle(RectangleCfg{
						Width:  1,
						Height: 1,
						Sizing: FillFit,
						Color:  cfg.SubheadingStyle.Color,
					}),
				},
			}),
		},
	})
}

func makeSelectOnKeyDown(cfg SelectCfg, idScroll uint32) func(*Layout, *Event, *Window) {
	return func(_ *Layout, e *Event, w *Window) {
		selectOnKeyDown(cfg, idScroll, e, w)
	}
}

func selectOnKeyDown(cfg SelectCfg, idScroll uint32, e *Event, w *Window) {
	if len(cfg.Options) == 0 {
		return
	}

	ss := StateMap[string, bool](w, nsSelect, capModerate)
	sh := StateMap[string, int](w, nsSelectHL, capModerate)
	isOpen, _ := ss.Get(cfg.ID)

	// Open on space/enter.
	if (e.KeyCode == KeySpace || e.KeyCode == KeyEnter) && !isOpen {
		ss.Set(cfg.ID, true)
		initialIdx := 0
		if len(cfg.Selected) > 0 {
			for i, opt := range cfg.Options {
				if opt == cfg.Selected[0] {
					initialIdx = i
					break
				}
			}
		}
		sh.Set(cfg.ID, initialIdx)
		e.IsHandled = true
		return
	}

	// Close on escape or tab.
	if (e.KeyCode == KeyEscape || e.KeyCode == KeyTab) && isOpen {
		ss.Clear()
		e.IsHandled = true
		return
	}

	if isOpen {
		currentIdx, _ := sh.Get(cfg.ID)
		action := listCoreNavigate(e.KeyCode, len(cfg.Options))

		if action == ListCoreSelectItem {
			if currentIdx >= 0 && currentIdx < len(cfg.Options) {
				option := cfg.Options[currentIdx]
				if !strings.HasPrefix(option, "---") {
					if !cfg.SelectMultiple {
						ss.Clear()
					}
					var s []string
					if cfg.SelectMultiple {
						s = listBoxNextSelectedIDs(
							cfg.Selected, option,
							true)
					} else {
						ss.Clear()
						s = []string{option}
					}
					if cfg.OnSelect != nil {
						cfg.OnSelect(s, e, w)
					}
				}
			}
			e.IsHandled = true
			return
		}

		if action == ListCoreMoveUp || action == ListCoreMoveDown {
			dir := 1
			if action == ListCoreMoveUp {
				dir = -1
			}
			nextIdx := currentIdx + dir
			// Skip subheaders.
			for nextIdx >= 0 && nextIdx < len(cfg.Options) {
				if !strings.HasPrefix(cfg.Options[nextIdx], "---") {
					break
				}
				nextIdx += dir
			}
			// Clamp.
			if nextIdx < 0 {
				nextIdx = 0
				for nextIdx < len(cfg.Options) &&
					strings.HasPrefix(cfg.Options[nextIdx], "---") {
					nextIdx++
				}
			} else if nextIdx >= len(cfg.Options) {
				nextIdx = len(cfg.Options) - 1
				for nextIdx >= 0 &&
					strings.HasPrefix(cfg.Options[nextIdx], "---") {
					nextIdx--
				}
			}
			if nextIdx >= 0 && nextIdx < len(cfg.Options) &&
				!strings.HasPrefix(cfg.Options[nextIdx], "---") {
				sh.Set(cfg.ID, nextIdx)
				rowH := cfg.TextStyle.Size + 4
				listH := float32(200) - 2*cfg.SizeBorder.Get(
					DefaultSelectStyle.SizeBorder)
				scrollEnsureVisible(
					idScroll, nextIdx, rowH, listH, w)
				e.IsHandled = true
			}
		}
	}
}

func fnvSum32(s string) uint32 {
	const offset uint32 = 2166136261
	const prime uint32 = 16777619
	h := offset
	for i := 0; i < len(s); i++ {
		h ^= uint32(s[i])
		h *= prime
	}
	return h
}

func applySelectDefaults(cfg *SelectCfg) {
	d := &DefaultSelectStyle
	if !cfg.Color.IsSet() {
		cfg.Color = d.Color
	}
	if !cfg.ColorBorder.IsSet() {
		cfg.ColorBorder = d.ColorBorder
	}
	if !cfg.ColorBorderFocus.IsSet() {
		cfg.ColorBorderFocus = d.ColorBorderFocus
	}
	if !cfg.ColorFocus.IsSet() {
		cfg.ColorFocus = d.ColorFocus
	}
	if !cfg.ColorSelect.IsSet() {
		cfg.ColorSelect = d.ColorSelect
	}
	if !cfg.Padding.IsSet() {
		cfg.Padding = Some(d.Padding)
	}

	if cfg.TextStyle == (TextStyle{}) {
		cfg.TextStyle = d.TextStyleNormal
	}
	if cfg.SubheadingStyle == (TextStyle{}) {
		cfg.SubheadingStyle = d.SubheadingStyle
	}
	if cfg.PlaceholderStyle == (TextStyle{}) {
		cfg.PlaceholderStyle = d.PlaceholderStyle
	}
}
