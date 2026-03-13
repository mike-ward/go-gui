package gui

// PulsarCfg configures a blinking text indicator.
type PulsarCfg struct {
	ID        string
	Text1     string
	Text2     string
	Color     Color
	TextStyle TextStyle
	Size      Opt[float32]
	Width     float32
}

// Pulsar creates a blinking text view that toggles between
// Text1 and Text2 based on the window's input cursor state.
func Pulsar(cfg PulsarCfg, w *Window) View {
	if cfg.Text1 == "" {
		cfg.Text1 = "..."
	}
	if cfg.Text2 == "" {
		cfg.Text2 = ".."
	}

	ts := cfg.TextStyle
	if ts == (TextStyle{}) {
		ts = guiTheme.TextStyleDef
	}
	if cfg.Color.IsSet() {
		ts.Color = cfg.Color
	}
	if cfg.Size.IsSet() {
		ts.Size = cfg.Size.Get(0)
	}

	if !w.hasAnimationLocked(blinkCursorAnimationID) {
		w.animationAdd(NewBlinkCursorAnimation())
	}

	txt := cfg.Text2
	if w.InputCursorOn() {
		txt = cfg.Text1
	}

	width := cfg.Width
	if width <= 0 {
		width = w.TextWidth(cfg.Text1, ts)
	}

	return Column(ContainerCfg{
		ID:         cfg.ID,
		MinWidth:   width,
		Padding:    NoPadding,
		SizeBorder: NoBorder,
		Sizing:     FitFit,
		Content: []View{
			Text(TextCfg{
				Text:      txt,
				TextStyle: ts,
			}),
		},
	})
}
