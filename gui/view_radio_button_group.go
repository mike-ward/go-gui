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
	Padding     Padding
	ColorBorder Color
	SizeBorder  float32
	MinWidth    float32
	MinHeight   float32
	IDFocus     uint32

	A11YLabel       string
	A11YDescription string
}

// RadioButtonGroupColumn creates a vertically stacked radio
// button group.
func RadioButtonGroupColumn(cfg RadioButtonGroupCfg) View {
	applyRadioGroupDefaults(&cfg)
	return Column(ContainerCfg{
		A11YRole:        AccessRoleRadioGroup,
		A11YLabel:       cfg.A11YLabel,
		A11YDescription: cfg.A11YDescription,
		ColorBorder:     cfg.ColorBorder,
		SizeBorder:      cfg.SizeBorder,
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
	return Row(ContainerCfg{
		A11YRole:        AccessRoleRadioGroup,
		A11YLabel:       cfg.A11YLabel,
		A11YDescription: cfg.A11YDescription,
		ColorBorder:     cfg.ColorBorder,
		SizeBorder:      cfg.SizeBorder,
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
	if cfg.ColorBorder == (Color{}) {
		cfg.ColorBorder = guiTheme.ColorBorder
	}
	if cfg.SizeBorder == 0 {
		cfg.SizeBorder = guiTheme.SizeBorder
	}
	if cfg.Padding == (Padding{}) {
		cfg.Padding = guiTheme.PaddingLarge
	}
}
