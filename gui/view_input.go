package gui

// InputCfg configures a text input field.
type InputCfg struct {
	ID            string
	Text          string
	Placeholder   string
	Mask          string
	MaskPreset    InputMaskPreset
	MaskTokens    []MaskTokenDef
	Mode          InputMode
	IsPassword    bool
	Disabled      bool
	Invisible     bool

	// Sizing
	Sizing    Sizing
	Width     float32
	Height    float32
	MinWidth  float32
	MaxWidth  float32
	MinHeight float32
	MaxHeight float32

	// Appearance
	Padding          Padding
	Radius           float32
	SizeBorder       float32
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
	OnTextChanged func(*Layout, string, *Window)
	OnTextCommit  func(*Layout, string, InputCommitReason, *Window)
	OnEnter       func(*Layout, *Event, *Window)
	OnKeyDown     func(*Layout, *Event, *Window)
	OnBlur        func(*Layout, *Window)

	// Accessibility
	A11YLabel       string
	A11YDescription string
}

// Input creates a text input field view.
func Input(cfg InputCfg) View {
	applyInputDefaults(&cfg)

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

	txtContent := []View{
		Text(TextCfg{
			IDFocus:           cfg.IDFocus,
			Sizing:            FillFill,
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
		SizeBorder:      cfg.SizeBorder,
		Invisible:       cfg.Invisible,
		Padding:         cfg.Padding,
		Radius:          cfg.Radius,
		Sizing:          cfg.Sizing,
		IDScroll:        cfg.IDScroll,
		Spacing:         0,
		OnHover: func(_ *Layout, _ *Event, w *Window) {
			if w.IsFocus(idFocus) {
				w.SetMouseCursor(CursorIBeam)
			} else {
				// Layout color change handled by hover
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
		},
		Content: []View{
			Row(ContainerCfg{
				Padding: PaddingNone,
				Sizing:  FillFill,
				VAlign:  vAlign,
				OnClick: func(layout *Layout, _ *Event, w *Window) {
					if len(layout.Children) < 1 {
						return
					}
					ly := layout.Children[0]
					if ly.Shape.IDFocus > 0 {
						w.SetIDFocus(ly.Shape.IDFocus)
					}
				},
				Content: txtContent,
			}),
		},
	})
}

func applyInputDefaults(cfg *InputCfg) {
	d := &DefaultButtonStyle
	if cfg.Color == (Color{}) {
		cfg.Color = d.Color
	}
	if cfg.ColorHover == (Color{}) {
		cfg.ColorHover = d.ColorHover
	}
	if cfg.ColorBorder == (Color{}) {
		cfg.ColorBorder = d.ColorBorder
	}
	if cfg.ColorBorderFocus == (Color{}) {
		cfg.ColorBorderFocus = d.ColorBorderFocus
	}
	if cfg.Padding == (Padding{}) {
		cfg.Padding = PaddingTwoFour
	}
	if cfg.Radius == 0 {
		cfg.Radius = RadiusMedium
	}
	if cfg.SizeBorder == 0 {
		cfg.SizeBorder = SizeBorderDef
	}
	if cfg.TextStyle == (TextStyle{}) {
		cfg.TextStyle = DefaultTextStyle
	}
	if cfg.PlaceholderStyle == (TextStyle{}) {
		cfg.PlaceholderStyle = TextStyle{
			Color: RGB(150, 150, 150),
			Size:  SizeTextMedium,
		}
	}
}

// passwordMask replaces each rune with a bullet character.
func passwordMask(text string) string {
	runes := []rune(text)
	for i := range runes {
		runes[i] = '•'
	}
	return string(runes)
}
