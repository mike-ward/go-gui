package main

import "github.com/mike-ward/go-gui/gui"

func demoInput(w *gui.Window) gui.View {
	t := gui.CurrentTheme()
	app := gui.State[ArcadeApp](w)

	return gui.Column(gui.ContainerCfg{
		Sizing:  gui.FillFit,
		Spacing: gui.Some(float32(12)),
		Padding: gui.Some(gui.PaddingNone),
		Content: []gui.View{
			labeledRow(t, "Text", gui.Input(gui.InputCfg{
				ID:          "input-text",
				Sizing:      gui.FillFit,
				Text:        app.InputText,
				Placeholder: "Enter text...",
				OnTextChanged: func(_ *gui.Layout, s string, w *gui.Window) {
					gui.State[ArcadeApp](w).InputText = s
				},
			})),
			labeledRow(t, "Password", gui.Input(gui.InputCfg{
				ID:          "input-password",
				Sizing:      gui.FillFit,
				Text:        app.InputPassword,
				Placeholder: "Enter password...",
				IsPassword:  true,
				OnTextChanged: func(_ *gui.Layout, s string, w *gui.Window) {
					gui.State[ArcadeApp](w).InputPassword = s
				},
			})),
			labeledRow(t, "Phone", gui.Input(gui.InputCfg{
				ID:          "input-phone",
				Sizing:      gui.FillFit,
				Text:        app.InputPhone,
				Placeholder: "(555) 000-0000",
				MaskPreset:  gui.MaskPhoneUS,
				OnTextChanged: func(_ *gui.Layout, s string, w *gui.Window) {
					gui.State[ArcadeApp](w).InputPhone = s
				},
			})),
			labeledRow(t, "Multiline", gui.Input(gui.InputCfg{
				ID:          "input-multi",
				Sizing:      gui.FillFit,
				Text:        app.InputMultiline,
				Placeholder: "Multiple lines...",
				Mode:        gui.InputMultiline,
				Height:      80,
				OnTextChanged: func(_ *gui.Layout, s string, w *gui.Window) {
					gui.State[ArcadeApp](w).InputMultiline = s
				},
			})),
		},
	})
}

func demoNumericInput(w *gui.Window) gui.View {
	t := gui.CurrentTheme()
	_ = w
	return gui.Column(gui.ContainerCfg{
		Sizing:  gui.FillFit,
		Spacing: gui.Some(float32(12)),
		Padding: gui.Some(gui.PaddingNone),
		Content: []gui.View{
			gui.NumericInput(gui.NumericInputCfg{
				ID:          "numeric-plain",
				Placeholder: "Enter a number",
				Sizing:      gui.FillFit,
			}),
			gui.Text(gui.TextCfg{
				Text:      "Adjust with arrow keys or step buttons",
				TextStyle: t.N2,
			}),
		},
	})
}

func demoColorPicker(w *gui.Window) gui.View {
	app := gui.State[ArcadeApp](w)
	return gui.ColorPicker(gui.ColorPickerCfg{
		ID:    "color-picker",
		Color: app.ColorPickerColor,
		OnColorChange: func(c gui.Color, _ *gui.Event, w *gui.Window) {
			gui.State[ArcadeApp](w).ColorPickerColor = c
		},
	})
}

func demoDatePicker(w *gui.Window) gui.View {
	_ = w
	return gui.DatePicker(gui.DatePickerCfg{
		ID: "date-picker",
	})
}

func demoDatePickerRoller(w *gui.Window) gui.View {
	_ = w
	return gui.DatePickerRoller(gui.DatePickerRollerCfg{
		ID: "date-roller",
	})
}

func demoInputDate(w *gui.Window) gui.View {
	_ = w
	return gui.InputDate(gui.InputDateCfg{
		ID:     "input-date",
		Sizing: gui.FillFit,
	})
}

func demoForms(w *gui.Window) gui.View {
	t := gui.CurrentTheme()
	app := gui.State[ArcadeApp](w)
	return gui.Column(gui.ContainerCfg{
		Sizing:      gui.FillFit,
		Spacing:     gui.Some(float32(12)),
		Padding:     gui.Some(gui.NewPadding(16, 16, 16, 16)),
		Color:       t.ColorPanel,
		Radius:      gui.Some(float32(8)),
		ColorBorder: t.ColorBorder,
		SizeBorder:  gui.Some(float32(1)),
		Content: []gui.View{
			gui.Text(gui.TextCfg{Text: "Registration Form", TextStyle: t.B4}),
			labeledRow(t, "Name", gui.Input(gui.InputCfg{
				ID:          "form-name",
				Sizing:      gui.FillFit,
				Text:        app.InputText,
				Placeholder: "Your name",
				OnTextChanged: func(_ *gui.Layout, s string, w *gui.Window) {
					gui.State[ArcadeApp](w).InputText = s
				},
			})),
			labeledRow(t, "Email", gui.Input(gui.InputCfg{
				ID:          "form-email",
				Sizing:      gui.FillFit,
				Text:        app.InputPassword,
				Placeholder: "you@example.com",
				OnTextChanged: func(_ *gui.Layout, s string, w *gui.Window) {
					gui.State[ArcadeApp](w).InputPassword = s
				},
			})),
			gui.Button(gui.ButtonCfg{
				ID:      "form-submit",
				Padding: gui.Some(gui.NewPadding(8, 24, 8, 24)),
				Content: []gui.View{
					gui.Text(gui.TextCfg{Text: "Submit", TextStyle: t.B3}),
				},
				OnClick: func(_ *gui.Layout, e *gui.Event, w *gui.Window) {
					w.Toast(gui.ToastCfg{Title: "Form", Body: "Submitted!"})
					e.IsHandled = true
				},
			}),
		},
	})
}

func labeledRow(t gui.Theme, label string, content gui.View) gui.View {
	return gui.Row(gui.ContainerCfg{
		Sizing:  gui.FillFit,
		Spacing: gui.Some(float32(8)),
		Padding: gui.Some(gui.PaddingNone),
		VAlign:  gui.VAlignMiddle,
		Content: []gui.View{
			gui.Text(gui.TextCfg{
				Text:      label,
				TextStyle: t.B3,
				Sizing:    gui.FixedFit,
			}),
			content,
		},
	})
}
