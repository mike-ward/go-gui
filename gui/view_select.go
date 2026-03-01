package gui

import (
	"hash/fnv"
	"strings"
)

// SelectCfg configures a select (dropdown) view.
type SelectCfg struct {
	ID              string
	Placeholder     string
	Selected        []string // currently selected option text(s)
	Options         []string
	Color           Color
	ColorBorder     Color
	ColorBorderFocus Color
	ColorFocus      Color
	ColorSelect     Color
	Padding         Padding
	SizeBorder      float32
	TextStyle       TextStyle
	SubheadingStyle TextStyle
	PlaceholderStyle TextStyle
	OnSelect        func([]string, *Event, *Window)
	MinWidth        float32
	MaxWidth        float32
	Radius          float32
	IDFocus         uint32
	SelectMultiple  bool
	NoWrap          bool
	Sizing          Sizing
	Disabled        bool
	Invisible       bool

	A11YLabel       string
	A11YDescription string
}

// Select creates a select (dropdown) view.
func Select(cfg SelectCfg) View {
	applySelectDefaults(&cfg)

	isOpen := false // read from state at runtime via AmendLayout
	_ = fnvSum32   // dropdown scroll ID deferred to Phase 6

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

	id := cfg.ID
	colorFocus := cfg.ColorFocus
	colorBorderFocus := cfg.ColorBorderFocus

	arrowText := "▼"
	_ = isOpen // open/close managed by state

	spacerSizing := FillFill
	if wrapMode != TextModeSingleLine {
		spacerSizing = FitFill
	}

	content := []View{
		Text(TextCfg{
			Text:      txt,
			TextStyle: txtStyle,
			Mode:      wrapMode,
		}),
		Row(ContainerCfg{
			Sizing:  spacerSizing,
			Padding: PaddingNone,
		}),
		Text(TextCfg{
			Text:      arrowText,
			TextStyle: cfg.TextStyle,
		}),
	}

	return Row(ContainerCfg{
		ID:          cfg.ID,
		IDFocus:     cfg.IDFocus,
		Clip:        clip,
		A11YRole:    AccessRoleComboBox,
		A11YLabel:   a11yLabel(cfg.A11YLabel, cfg.Placeholder),
		Color:       cfg.Color,
		ColorBorder: cfg.ColorBorder,
		SizeBorder:  cfg.SizeBorder,
		Radius:      cfg.Radius,
		Padding:     cfg.Padding,
		Sizing:      cfg.Sizing,
		MinWidth:    cfg.MinWidth,
		MaxWidth:    cfg.MaxWidth,
		Disabled:    cfg.Disabled,
		Invisible:   cfg.Invisible,
		AmendLayout: func(layout *Layout, w *Window) {
			if layout.Shape.Disabled {
				return
			}
			if w.IsFocus(layout.Shape.IDFocus) {
				layout.Shape.Color = colorFocus
				layout.Shape.ColorBorder = colorBorderFocus
			}
		},
		OnKeyDown: makeSelectOnKeyDown(cfg),
		OnClick: func(_ *Layout, e *Event, w *Window) {
			ss := StateMap[string, bool](w, nsSelect, capModerate)
			ss.Clear()
			cur, _ := ss.Get(id)
			ss.Set(id, !cur)
		},
		Content: content,
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
	if cfg.SizeBorder == 0 {
		cfg.SizeBorder = SizeBorderDef
	}
	if cfg.Radius == 0 {
		cfg.Radius = RadiusMedium
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
