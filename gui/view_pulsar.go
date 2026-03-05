package gui

// PulsarCfg configures a blinking text indicator.
type PulsarCfg struct {
	ID    string
	Text1 string
	Text2 string
	Color Color
	Size  float32
	Width float32
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
	if cfg.Size == 0 {
		cfg.Size = guiTheme.SizeTextMedium
	}
	if !cfg.Color.IsSet() {
		cfg.Color = guiTheme.TextStyleDef.Color
	}

	ts := TextStyle{
		Color: cfg.Color,
		Size:  cfg.Size,
	}

	txt := cfg.Text2
	if w.InputCursorOn() {
		txt = cfg.Text1
	}

	width := cfg.Width
	if width <= 0 {
		// Placeholder width based on Text1 length.
		width = cfg.Size * 0.6 * float32(len(cfg.Text1))
	}

	return Column(ContainerCfg{
		ID:       cfg.ID,
		MinWidth: width,
		Padding:  PaddingNone,
		Content: []View{
			Text(TextCfg{
				Text:      txt,
				TextStyle: ts,
			}),
		},
	})
}
