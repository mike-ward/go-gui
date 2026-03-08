package main

import "github.com/mike-ward/go-gui/gui"

func line() gui.View {
	t := gui.CurrentTheme()
	return gui.Row(gui.ContainerCfg{
		Sizing:     gui.FillFit,
		Padding:    gui.Some(gui.NewPadding(2, 5, 0, 0)),
		SizeBorder: gui.Some(float32(0)),
		Content: []gui.View{
			gui.Row(gui.ContainerCfg{
				Height:     1,
				Sizing:     gui.FillFit,
				Padding:    gui.Some(gui.PaddingNone),
				SizeBorder: gui.Some(float32(0)),
				Color:      t.ColorActive,
			}),
		},
	})
}

func demoBox(label string, color gui.Color) gui.View {
	return demoBoxSized(label, color, 60, 40)
}

func demoBoxSized(label string, color gui.Color, w, h float32) gui.View {
	t := gui.CurrentTheme()
	return gui.Column(gui.ContainerCfg{
		Width:  w,
		Height: h,
		Sizing: gui.FixedFixed,
		Color:  color,
		Radius: gui.Some(float32(4)),
		HAlign: gui.HAlignCenter,
		VAlign: gui.VAlignMiddle,
		Content: []gui.View{
			gui.Text(gui.TextCfg{Text: label, TextStyle: t.N2}),
		},
	})
}

func float64p(v float64) *float64 {
	return &v
}

func placeholderHeader(text string) gui.View {
	return gui.Column(gui.ContainerCfg{
		Sizing:  gui.FillFit,
		Padding: gui.Some(gui.NewPadding(12, 12, 12, 12)),
		Color:   gui.CurrentTheme().ColorPanel,
		Radius:  gui.Some(float32(8)),
		Content: []gui.View{
			showcaseWrappedText(text, gui.CurrentTheme().N3),
		},
	})
}

func showcaseWrappedText(text string, style gui.TextStyle) gui.View {
	return gui.Text(gui.TextCfg{
		Text:      text,
		TextStyle: style,
		Mode:      gui.TextModeWrap,
	})
}
