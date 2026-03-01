package gui

// TextCfg configures a text view. Use for labels or larger
// blocks of multiline text. Giving it an IDFocus allows mark
// and copy operations.
type TextCfg struct {
	ID               string
	Text             string
	TextStyle        TextStyle
	IDFocus          uint32
	TabSize          uint32
	MinWidth         float32
	Mode             TextMode
	Invisible        bool
	Clip             bool
	FocusSkip        bool
	Disabled         bool
	IsPassword       bool
	PlaceholderActive bool
	Hero             bool
	Opacity          float32
	Sizing           Sizing

	// Accessibility
	A11YLabel       string
	A11YDescription string
}

// textView implements View for text rendering.
type textView struct {
	cfg     TextCfg
	sizing  Sizing
}

func (tv *textView) Content() []View { return nil }

func (tv *textView) GenerateLayout(w *Window) Layout {
	c := &tv.cfg
	ts := &c.TextStyle

	layout := Layout{
		Shape: &Shape{
			ShapeType: ShapeText,
			ID:        c.ID,
			IDFocus:   c.IDFocus,
			A11YRole:  AccessRoleStaticText,
			A11Y: makeA11YInfo(
				a11yLabel(c.A11YLabel, c.Text), c.A11YDescription,
			),
			Clip:      c.Clip,
			FocusSkip: c.FocusSkip,
			Disabled:  c.Disabled,
			MinWidth:  c.MinWidth,
			Sizing:    tv.sizing,
			Hero:      c.Hero,
			Opacity:   c.Opacity,
			TC: &ShapeTextConfig{
				Text:              c.Text,
				TextStyle:         ts,
				TextIsPassword:    c.IsPassword,
				TextIsPlaceholder: c.PlaceholderActive,
				TextMode:          c.Mode,
				TextTabSize:       c.TabSize,
			},
		},
	}

	if ts.Size == 0 {
		ts.Size = SizeTextMedium
	}
	if w.textMeasurer != nil {
		layout.Shape.Width = w.textMeasurer.TextWidth(c.Text, *ts)
		layout.Shape.Height = w.textMeasurer.FontHeight(*ts)
	} else {
		// Fallback for tests (no backend).
		charWidth := ts.Size * 0.6
		layout.Shape.Width = float32(len(c.Text)) * charWidth
		layout.Shape.Height = ts.Size * 1.4
	}

	if c.Mode == TextModeSingleLine ||
		layout.Shape.Sizing.Width == SizingFixed {
		layout.Shape.MinWidth = f32Max(
			layout.Shape.Width, layout.Shape.MinWidth,
		)
		layout.Shape.Width = layout.Shape.MinWidth
	}
	if c.Mode == TextModeSingleLine ||
		layout.Shape.Sizing.Height == SizingFixed {
		layout.Shape.MinHeight = f32Max(
			layout.Shape.Height, layout.Shape.MinHeight,
		)
		layout.Shape.Height = layout.Shape.MinHeight
	}
	if layout.Shape.Opacity == 0 {
		layout.Shape.Opacity = 1.0
	}
	ApplyFixedSizingConstraints(layout.Shape)
	return layout
}

// Text creates a text view for displaying text content.
func Text(cfg TextCfg) View {
	if cfg.Invisible {
		return invisibleContainerView()
	}
	sizing := cfg.Sizing
	if sizing == (Sizing{}) {
		if cfg.Mode == TextModeWrap ||
			cfg.Mode == TextModeWrapKeepSpaces {
			sizing = FillFit
		} else {
			sizing = FitFit
		}
	}
	if cfg.TabSize == 0 {
		cfg.TabSize = 4
	}
	if cfg.TextStyle == (TextStyle{}) {
		cfg.TextStyle = DefaultTextStyle
	}
	return &textView{cfg: cfg, sizing: sizing}
}
