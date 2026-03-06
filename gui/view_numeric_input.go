package gui

// NumericInputCfg configures a locale-aware numeric input with
// optional step controls.
type NumericInputCfg struct {
	ID          string
	IDFocus     uint32
	Text        string
	Value       *float64
	Placeholder string
	Locale      NumericLocaleCfg
	StepCfg     NumericStepCfg
	Mode        NumericInputMode
	CurrencyCfg NumericCurrencyModeCfg
	PercentCfg  NumericPercentModeCfg
	Decimals    int
	Min         *float64
	Max         *float64

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

	Disabled  bool
	Invisible bool

	// Callbacks
	OnTextChanged func(*Layout, string, *Window)
	OnValueCommit func(*Layout, *float64, string, *Window)

	// Accessibility
	A11YLabel       string
	A11YDescription string
}

// DefaultNumericInputStyle holds defaults for NumericInputCfg Opt fields.
var DefaultNumericInputStyle = struct {
	SizeBorder float32
	Radius     float32
}{
	SizeBorder: SizeBorderDef,
	Radius:     RadiusMedium,
}

// NumericInput creates a locale-aware numeric input.
func NumericInput(cfg NumericInputCfg) View {
	applyNumericInputDefaults(&cfg)

	dn := &DefaultNumericInputStyle
	sizeBorder := cfg.SizeBorder.Get(dn.SizeBorder)
	radius := cfg.Radius.Get(dn.Radius)
	locale := numericLocaleNormalize(cfg.Locale)
	stepCfg := numericStepCfgNormalize(cfg.StepCfg)

	field := numericInputField(cfg, locale, stepCfg, stepCfg.ShowButtons)
	if !stepCfg.ShowButtons {
		return field
	}

	colorHover := cfg.ColorHover
	colorBorderFocus := cfg.ColorBorderFocus
	idFocus := cfg.IDFocus

	content := []View{
		field,
		numericInputStepButtons(cfg, locale, stepCfg),
	}

	return Row(ContainerCfg{
		ID:          cfg.ID,
		IDFocus:     cfg.IDFocus,
		A11YRole:    AccessRoleTextField,
		A11YLabel:   a11yLabel(cfg.A11YLabel, cfg.Placeholder),
		Width:       cfg.Width,
		Height:      cfg.Height,
		MinWidth:    cfg.MinWidth,
		MaxWidth:    cfg.MaxWidth,
		MinHeight:   cfg.MinHeight,
		MaxHeight:   cfg.MaxHeight,
		Sizing:      cfg.Sizing,
		Clip:        true,
		Color:       cfg.Color,
		ColorBorder: cfg.ColorBorder,
		SizeBorder:  Some(sizeBorder),
		Radius:      Some(radius),
		Padding:     Some(PaddingNone),
		Invisible:   cfg.Invisible,
		Disabled:    cfg.Disabled,
		VAlign:      VAlignMiddle,
		Spacing:     Some(float32(0)),
		OnClick: func(_ *Layout, _ *Event, w *Window) {
			if idFocus > 0 {
				w.SetIDFocus(idFocus)
			}
		},
		OnHover: func(layout *Layout, _ *Event, w *Window) {
			if w.IsFocus(idFocus) {
				w.SetMouseCursor(CursorIBeam)
			} else {
				layout.Shape.Color = colorHover
			}
		},
		AmendLayout: func(layout *Layout, w *Window) {
			if layout.Shape.Disabled {
				return
			}
			if idFocus > 0 && w.IsFocus(idFocus) {
				layout.Shape.ColorBorder = colorBorderFocus
			}
		},
		Content: content,
	})
}

func numericInputField(cfg NumericInputCfg, locale NumericLocaleCfg, stepCfg NumericStepCfg, fillParent bool) View {
	sizing := cfg.Sizing
	if fillParent {
		sizing = FillFill
	}
	inputID := cfg.ID
	if fillParent && len(cfg.ID) > 0 {
		inputID = cfg.ID + "_field"
	}
	color := cfg.Color
	colorHover := cfg.ColorHover
	colorBorder := cfg.ColorBorder
	colorBorderFocus := cfg.ColorBorderFocus
	sizeBorder := cfg.SizeBorder
	radius := cfg.Radius
	if fillParent {
		color = ColorTransparent
		colorHover = ColorTransparent
		colorBorder = ColorTransparent
		colorBorderFocus = ColorTransparent
		sizeBorder = Opt[float32]{}
		radius = Opt[float32]{}
	}

	return Input(InputCfg{
		ID:               inputID,
		IDFocus:          cfg.IDFocus,
		Text:             cfg.Text,
		Placeholder:      cfg.Placeholder,
		Sizing:           sizing,
		Padding:          cfg.Padding,
		Radius:           radius,
		SizeBorder:       sizeBorder,
		Color:            color,
		ColorHover:       colorHover,
		ColorBorder:      colorBorder,
		ColorBorderFocus: colorBorderFocus,
		TextStyle:        cfg.TextStyle,
		PlaceholderStyle: cfg.PlaceholderStyle,
		Disabled:         cfg.Disabled,
		Invisible:        cfg.Invisible,
		OnTextChanged:    cfg.OnTextChanged,
		OnTextCommit: func(layout *Layout, text string, reason InputCommitReason, w *Window) {
			modeCfg := numericModeCfgFromInput(cfg)
			value, committed := numericInputCommitResultMode(
				text, cfg.Value, cfg.Min, cfg.Max,
				cfg.Decimals, locale, modeCfg)
			if cfg.OnValueCommit != nil {
				cfg.OnValueCommit(layout, value, committed, w)
			}
		},
	})
}

func numericInputStepButtons(cfg NumericInputCfg, locale NumericLocaleCfg, stepCfg NumericStepCfg) View {
	triangleSize := f32Max(cfg.TextStyle.Size-4, 8)
	triangleStyle := TextStyle{
		Color:  cfg.TextStyle.Color,
		Size:   triangleSize,
		Family: cfg.TextStyle.Family,
	}
	baseColor := cfg.Color

	return Column(ContainerCfg{
		Spacing:   Some(float32(0)),
		Sizing:    FitFill,
		Disabled:  cfg.Disabled,
		Invisible: cfg.Invisible,
		Padding:   Some(NewPadding(0, PadSmall, 0, 0)),
		Content: []View{
			Button(ButtonCfg{
				ID:          cfg.ID + "_step_up",
				Sizing:      FillFill,
				Padding:     Some(PaddingNone),
				Color:       baseColor,
				ColorHover:  cfg.ColorHover,
				ColorFocus:  cfg.ColorHover,
				ColorClick:  cfg.ColorBorderFocus,
				ColorBorder: ColorTransparent,
				SizeBorder:  Some(float32(0)),
				Radius:      Some(float32(0)),
				OnClick: func(layout *Layout, e *Event, w *Window) {
					numericInputApplyStep(
						layout, cfg, locale, stepCfg,
						1.0, e, w)
				},
				Content: []View{
					Text(TextCfg{
						Text:      "\u25B2",
						TextStyle: triangleStyle,
					}),
				},
			}),
			Button(ButtonCfg{
				ID:          cfg.ID + "_step_down",
				Sizing:      FillFill,
				Padding:     Some(PaddingNone),
				Color:       baseColor,
				ColorHover:  cfg.ColorHover,
				ColorFocus:  cfg.ColorHover,
				ColorClick:  cfg.ColorBorderFocus,
				ColorBorder: ColorTransparent,
				SizeBorder:  Some(float32(0)),
				Radius:      Some(float32(0)),
				OnClick: func(layout *Layout, e *Event, w *Window) {
					numericInputApplyStep(
						layout, cfg, locale, stepCfg,
						-1.0, e, w)
				},
				Content: []View{
					Text(TextCfg{
						Text:      "\u25BC",
						TextStyle: triangleStyle,
					}),
				},
			}),
		},
	})
}

func numericInputApplyStep(
	layout *Layout,
	cfg NumericInputCfg,
	locale NumericLocaleCfg,
	stepCfg NumericStepCfg,
	dir float64,
	e *Event,
	w *Window,
) {
	modeCfg := numericModeCfgFromInput(cfg)
	value, committed := numericInputStepResultMode(
		cfg.Text, cfg.Value, cfg.Min, cfg.Max,
		cfg.Decimals, stepCfg, locale, dir,
		e.Modifiers, modeCfg)
	if cfg.OnValueCommit != nil {
		cfg.OnValueCommit(layout, value, committed, w)
	}
}

func numericModeCfgFromInput(cfg NumericInputCfg) numericModeCfg {
	switch cfg.Mode {
	case NumericCurrency:
		return numericModeCfg{
			mode:              NumericCurrency,
			affix:             cfg.CurrencyCfg.Symbol,
			affixPosition:     cfg.CurrencyCfg.Position,
			affixSpacing:      cfg.CurrencyCfg.SymbolSpacing,
			displayMultiplier: 1.0,
		}
	case NumericPercent:
		return numericModeCfg{
			mode:              NumericPercent,
			affix:             cfg.PercentCfg.Symbol,
			affixPosition:     cfg.PercentCfg.Position,
			affixSpacing:      cfg.PercentCfg.SymbolSpacing,
			displayMultiplier: 100.0,
		}
	default:
		return numericModeCfg{
			mode:              NumericNumber,
			displayMultiplier: 1.0,
		}
	}
}

func applyNumericInputDefaults(cfg *NumericInputCfg) {
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
		cfg.PlaceholderStyle = TextStyle{
			Color: RGB(150, 150, 150),
			Size:  SizeTextMedium,
		}
	}
	if cfg.Decimals == 0 {
		cfg.Decimals = 2
	}
	if cfg.CurrencyCfg == (NumericCurrencyModeCfg{}) {
		cfg.CurrencyCfg = NumericCurrencyModeCfg{
			Symbol:   "$",
			Position: AffixPrefix,
		}
	}
	if cfg.PercentCfg == (NumericPercentModeCfg{}) {
		cfg.PercentCfg = NumericPercentModeCfg{
			Symbol:   "%",
			Position: AffixSuffix,
		}
	}
}
