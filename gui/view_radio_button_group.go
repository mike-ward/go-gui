package gui

// RadioOption defines a radio button for a RadioButtonGroupCfg.
type RadioOption struct {
	Label string
	Value string
}

// NewRadioOption creates a RadioOption.
func NewRadioOption(label, value string) RadioOption {
	return RadioOption{Label: label, Value: value}
}

// RadioButtonGroupCfg configures a radio button group.
type RadioButtonGroupCfg struct {
	Value       string
	Options     []RadioOption
	OnSelect    func(string, *Window)
	Sizing      Sizing
	Spacing     Opt[float32]
	Padding     Opt[Padding]
	ColorBorder Color
	SizeBorder  Opt[float32]
	MinWidth    float32
	MinHeight   float32
	IDFocus     uint32

	A11YLabel       string
	A11YDescription string
}

// DefaultRadioGroupStyle holds defaults for RadioButtonGroupCfg Opt fields.
var DefaultRadioGroupStyle = struct {
	SizeBorder float32
}{
	SizeBorder: 0,
}

// RadioButtonGroupColumn creates a vertically stacked radio
// button group.
func RadioButtonGroupColumn(cfg RadioButtonGroupCfg) View {
	applyRadioGroupDefaults(&cfg)
	drg := &DefaultRadioGroupStyle
	sizeBorder := cfg.SizeBorder.Get(drg.SizeBorder)
	return Column(ContainerCfg{
		A11YRole:        AccessRoleRadioGroup,
		A11YLabel:       cfg.A11YLabel,
		A11YDescription: cfg.A11YDescription,
		ColorBorder:     cfg.ColorBorder,
		SizeBorder:      Some(sizeBorder),
		Spacing:         cfg.Spacing,
		Padding:         cfg.Padding,
		MinWidth:        cfg.MinWidth,
		MinHeight:       cfg.MinHeight,
		Sizing:          cfg.Sizing,
		Content:         buildRadioOptions(cfg),
	})
}

// RadioButtonGroupRow creates a horizontally stacked radio
// button group.
func RadioButtonGroupRow(cfg RadioButtonGroupCfg) View {
	applyRadioGroupDefaults(&cfg)
	drg2 := &DefaultRadioGroupStyle
	sizeBorder2 := cfg.SizeBorder.Get(drg2.SizeBorder)
	return Row(ContainerCfg{
		A11YRole:        AccessRoleRadioGroup,
		A11YLabel:       cfg.A11YLabel,
		A11YDescription: cfg.A11YDescription,
		ColorBorder:     cfg.ColorBorder,
		SizeBorder:      Some(sizeBorder2),
		Spacing:         cfg.Spacing,
		Padding:         cfg.Padding,
		MinWidth:        cfg.MinWidth,
		MinHeight:       cfg.MinHeight,
		Sizing:          cfg.Sizing,
		Content:         buildRadioOptions(cfg),
	})
}

func buildRadioOptions(cfg RadioButtonGroupCfg) []View {
	content := make([]View, 0, len(cfg.Options))
	idFocus := cfg.IDFocus
	onSelect := cfg.OnSelect
	for _, opt := range cfg.Options {
		optValue := opt.Value
		content = append(content, Radio(RadioCfg{
			Label:    opt.Label,
			IDFocus:  idFocus,
			Selected: cfg.Value == opt.Value,
			OnClick: func(_ *Layout, _ *Event, w *Window) {
				if onSelect != nil {
					onSelect(optValue, w)
				}
			},
		}))
		if cfg.IDFocus != 0 {
			idFocus++
		}
	}
	return content
}

func applyRadioGroupDefaults(cfg *RadioButtonGroupCfg) {
	if !cfg.ColorBorder.IsSet() {
		cfg.ColorBorder = guiTheme.ColorBorder
	}
	if !cfg.Padding.IsSet() {
		cfg.Padding = Some(guiTheme.PaddingLarge)
	}
	if !cfg.Spacing.IsSet() {
		cfg.Spacing = Some(SpacingSmall)
	}
}
