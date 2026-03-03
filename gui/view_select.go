package gui

import (
	"hash/fnv"
	"strings"
)

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
	Padding          Padding
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
	idScroll := fnvSum32(cfg.ID + "dropdown")

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
			Padding: Some(PaddingNone),
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
						i == highlightedIdx, idScroll))
			}
		}
		content = append(content, Column(ContainerCfg{
			ID:          cfg.ID + "dropdown",
			SizeBorder:  Some(sizeBorder),
			Radius:      Some(radius),
			ColorBorder: cfg.ColorBorder,
			Color:       cfg.Color,
			MinHeight:   50,
			MaxHeight:   200,
			MinWidth:    cfg.MinWidth,
			MaxWidth:    cfg.MaxWidth,
			Float:         true,
			FloatAutoFlip: true,
			FloatAnchor:   FloatBottomLeft,
			FloatTieOff: FloatTopLeft,
			FloatOffsetY: -sizeBorder,
			IDScroll:    idScroll,
			Padding: Some(NewPadding(
				PadSmall, PadMedium, PadSmall, PadSmall)),
			Spacing: Some(float32(0)),
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
			Padding:     Some(cfg.Padding),
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
			OnKeyDown: makeSelectOnKeyDown(sv.cfg),
			OnClick: func(_ *Layout, e *Event, w *Window) {
				ss := StateMap[string, bool](
					w, nsSelect, capModerate)
				ss.Clear()
				cur := StateReadOr[string, bool](
					w, nsSelect, id, false)
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
func selectOptionView(cfg *SelectCfg, option string, index int, highlighted bool, idScroll uint32) View {
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
		Padding: Some(NewPadding(0, PadSmall, 0, 1)),
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
		Padding: Some(NewPadding(guiTheme.PaddingMedium.Top, 0, 0, 0)),
		Sizing:  FillFit,
		Content: []View{
				Row(ContainerCfg{
					Padding: Some(PaddingNone),
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

func makeSelectOnKeyDown(cfg SelectCfg) func(*Layout, *Event, *Window) {
	return func(_ *Layout, e *Event, w *Window) {
		selectOnKeyDown(cfg, e, w)
	}
}

func selectOnKeyDown(cfg SelectCfg, e *Event, w *Window) {
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
		return
	}

	// Close on escape.
	if e.KeyCode == KeyEscape && isOpen {
		ss.Clear()
		return
	}

	if isOpen {
		currentIdx, _ := sh.Get(cfg.ID)
		idScroll := fnvSum32(cfg.ID + "dropdown")
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
				scrollSY := StateMap[uint32, float32](
					w, nsScrollY, capScroll)
				scrollSY.Set(idScroll,
					float32(nextIdx)*rowH)
			}
		}
	}
}

func fnvSum32(s string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(s))
	return h.Sum32()
}

func applySelectDefaults(cfg *SelectCfg) {
	d := &DefaultButtonStyle
	if cfg.Color == (Color{}) {
		cfg.Color = d.Color
	}
	if cfg.ColorBorder == (Color{}) {
		cfg.ColorBorder = d.ColorBorder
	}
	if cfg.ColorBorderFocus == (Color{}) {
		cfg.ColorBorderFocus = d.ColorBorderFocus
	}
	if cfg.ColorFocus == (Color{}) {
		cfg.ColorFocus = d.ColorFocus
	}
	if cfg.ColorSelect == (Color{}) {
		cfg.ColorSelect = colorSelectDark
	}
	if cfg.Padding == (Padding{}) {
		cfg.Padding = PaddingTwoFour
	}

	if cfg.TextStyle == (TextStyle{}) {
		cfg.TextStyle = DefaultTextStyle
	}
	if cfg.SubheadingStyle == (TextStyle{}) {
		cfg.SubheadingStyle = TextStyle{
			Color: RGB(180, 180, 180),
			Size:  SizeTextSmall,
		}
	}
	if cfg.PlaceholderStyle == (TextStyle{}) {
		cfg.PlaceholderStyle = TextStyle{
			Color: RGB(150, 150, 150),
			Size:  SizeTextMedium,
		}
	}
}
