package main

import "github.com/mike-ward/go-gui/gui"

func line() gui.View {
	t := gui.CurrentTheme()
	return gui.Column(gui.ContainerCfg{
		Sizing:     gui.FillFit,
		Padding:    gui.SomeP(3, 0, 0, 0),
		SizeBorder: gui.NoBorder,
		Radius:     gui.NoRadius,
		Content: []gui.View{
			gui.Row(gui.ContainerCfg{
				Sizing:     gui.FillFit,
				Padding:    gui.NoPadding,
				SizeBorder: gui.NoBorder,
				Radius:     gui.NoRadius,
				Color:      t.ColorActive,
				Height:     1,
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
		Radius: gui.SomeF(4),
		HAlign: gui.HAlignCenter,
		VAlign: gui.VAlignMiddle,
		Content: []gui.View{
			gui.Text(gui.TextCfg{Text: label, TextStyle: t.N2}),
		},
	})
}

func placeholderHeader(text string) gui.View {
	return gui.Column(gui.ContainerCfg{
		Sizing:  gui.FillFit,
		Padding: gui.SomeP(12, 12, 12, 12),
		Color:   gui.CurrentTheme().ColorPanel,
		Radius:  gui.SomeF(8),
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
