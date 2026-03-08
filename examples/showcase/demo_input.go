package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/mike-ward/go-gui/gui"
)

func demoInput(w *gui.Window) gui.View {
	t := gui.CurrentTheme()
	app := gui.State[ShowcaseApp](w)

	return gui.Column(gui.ContainerCfg{
		Sizing:  gui.FillFit,
		Spacing: gui.Some(float32(12)),
		Padding: gui.Some(gui.PaddingNone),
		Content: []gui.View{
			showcaseWrappedText(
				"Accessibility: supports IME composition, keyboard tab focus, masked input, and multiline editing.",
				t.N3,
			),
			labeledRow(t, "Text", gui.Input(gui.InputCfg{
				ID:          "input-text",
				Sizing:      gui.FillFit,
				Text:        app.InputText,
				Placeholder: "Enter text...",
				OnTextChanged: func(_ *gui.Layout, s string, w *gui.Window) {
					gui.State[ShowcaseApp](w).InputText = s
				},
			})),
			labeledRow(t, "Password", gui.Input(gui.InputCfg{
				ID:          "input-password",
				Sizing:      gui.FillFit,
				Text:        app.InputPassword,
				Placeholder: "Enter password...",
				IsPassword:  true,
				OnTextChanged: func(_ *gui.Layout, s string, w *gui.Window) {
					gui.State[ShowcaseApp](w).InputPassword = s
				},
			})),
			labeledRow(t, "Phone", gui.Input(gui.InputCfg{
				ID:          "input-phone",
				Sizing:      gui.FillFit,
				Text:        app.InputPhone,
				Placeholder: "(555) 000-0000",
				MaskPreset:  gui.MaskPhoneUS,
				OnTextChanged: func(_ *gui.Layout, s string, w *gui.Window) {
					gui.State[ShowcaseApp](w).InputPhone = s
				},
			})),
			labeledRow(t, "Expiry", gui.Input(gui.InputCfg{
				ID:          "input-expiry",
				Sizing:      gui.FillFit,
				Text:        app.InputExpiry,
				Placeholder: "MM/YY",
				Mask:        "##/##",
				OnTextChanged: func(_ *gui.Layout, s string, w *gui.Window) {
					gui.State[ShowcaseApp](w).InputExpiry = s
				},
			})),
			labeledRow(t, "Multiline", gui.Input(gui.InputCfg{
				ID:          "input-multi",
				Sizing:      gui.FillFit,
				Text:        app.InputMultiline,
				Placeholder: "Multiple lines...",
				Mode:        gui.InputMultiline,
				Height:      90,
				OnTextChanged: func(_ *gui.Layout, s string, w *gui.Window) {
					gui.State[ShowcaseApp](w).InputMultiline = s
				},
			})),
		},
	})
}

func demoNumericInput(w *gui.Window) gui.View {
	app := gui.State[ShowcaseApp](w)
	titleStyle := gui.CurrentTheme().B1
	bodyStyle := gui.CurrentTheme().N3

	return gui.Column(gui.ContainerCfg{
		Sizing:  gui.FillFit,
		Spacing: gui.Some(float32(12)),
		Padding: gui.Some(gui.PaddingNone),
		Content: []gui.View{
			gui.Text(gui.TextCfg{Text: "Default (en_US)", TextStyle: titleStyle}),
			gui.NumericInput(gui.NumericInputCfg{
				ID:       "num-en",
				IDFocus:  9170,
				Text:     app.NumericENText,
				Value:    app.NumericENValue,
				Decimals: 2,
				Min:      float64p(0),
				Max:      float64p(10000),
				Width:    220,
				Sizing:   gui.FixedFit,
				OnTextChanged: func(_ *gui.Layout, text string, w *gui.Window) {
					gui.State[ShowcaseApp](w).NumericENText = text
				},
				OnValueCommit: func(_ *gui.Layout, value *float64, text string, w *gui.Window) {
					app := gui.State[ShowcaseApp](w)
					app.NumericENValue = cloneFloatPtr(value)
					app.NumericENText = text
				},
			}),
			gui.Text(gui.TextCfg{Text: "Committed: " + numericValueText(app.NumericENValue), TextStyle: bodyStyle}),

			gui.Text(gui.TextCfg{Text: "German (de_DE)", TextStyle: titleStyle}),
			gui.NumericInput(gui.NumericInputCfg{
				ID:       "num-de",
				IDFocus:  9171,
				Text:     app.NumericDEText,
				Value:    app.NumericDEValue,
				Decimals: 2,
				Locale: gui.NumericLocaleCfg{
					DecimalSep: ',',
					GroupSep:   '.',
				},
				Min:    float64p(0),
				Max:    float64p(10000),
				Width:  220,
				Sizing: gui.FixedFit,
				OnTextChanged: func(_ *gui.Layout, text string, w *gui.Window) {
					gui.State[ShowcaseApp](w).NumericDEText = text
				},
				OnValueCommit: func(_ *gui.Layout, value *float64, text string, w *gui.Window) {
					app := gui.State[ShowcaseApp](w)
					app.NumericDEValue = cloneFloatPtr(value)
					app.NumericDEText = text
				},
			}),
			gui.Text(gui.TextCfg{Text: "Committed: " + numericValueText(app.NumericDEValue), TextStyle: bodyStyle}),

			gui.Text(gui.TextCfg{Text: "Currency mode", TextStyle: titleStyle}),
			gui.NumericInput(gui.NumericInputCfg{
				ID:       "num-currency",
				IDFocus:  9173,
				Text:     app.NumericCurrencyText,
				Value:    app.NumericCurrencyValue,
				Mode:     gui.NumericCurrency,
				Decimals: 2,
				Min:      float64p(0),
				Max:      float64p(10000),
				Width:    220,
				Sizing:   gui.FixedFit,
				OnTextChanged: func(_ *gui.Layout, text string, w *gui.Window) {
					gui.State[ShowcaseApp](w).NumericCurrencyText = text
				},
				OnValueCommit: func(_ *gui.Layout, value *float64, text string, w *gui.Window) {
					app := gui.State[ShowcaseApp](w)
					app.NumericCurrencyValue = cloneFloatPtr(value)
					app.NumericCurrencyText = text
				},
			}),
			gui.Text(gui.TextCfg{Text: "Committed: " + numericValueText(app.NumericCurrencyValue), TextStyle: bodyStyle}),

			gui.Text(gui.TextCfg{Text: "Percent mode (ratio value)", TextStyle: titleStyle}),
			gui.NumericInput(gui.NumericInputCfg{
				ID:       "num-percent",
				IDFocus:  9174,
				Text:     app.NumericPercentText,
				Value:    app.NumericPercentValue,
				Mode:     gui.NumericPercent,
				Decimals: 2,
				Min:      float64p(0),
				Max:      float64p(1),
				Width:    220,
				Sizing:   gui.FixedFit,
				OnTextChanged: func(_ *gui.Layout, text string, w *gui.Window) {
					gui.State[ShowcaseApp](w).NumericPercentText = text
				},
				OnValueCommit: func(_ *gui.Layout, value *float64, text string, w *gui.Window) {
					app := gui.State[ShowcaseApp](w)
					app.NumericPercentValue = cloneFloatPtr(value)
					app.NumericPercentText = text
				},
			}),
			gui.Text(gui.TextCfg{Text: "Committed ratio: " + numericValueText(app.NumericPercentValue), TextStyle: bodyStyle}),

			gui.Text(gui.TextCfg{Text: "No buttons (integer)", TextStyle: titleStyle}),
			gui.NumericInput(gui.NumericInputCfg{
				ID:       "num-plain",
				IDFocus:  9172,
				Text:     app.NumericPlainText,
				Value:    app.NumericPlainValue,
				Decimals: 0,
				StepCfg: gui.NumericStepCfg{
					ShowButtons: false,
				},
				Width:  220,
				Sizing: gui.FixedFit,
				OnTextChanged: func(_ *gui.Layout, text string, w *gui.Window) {
					gui.State[ShowcaseApp](w).NumericPlainText = text
				},
				OnValueCommit: func(_ *gui.Layout, value *float64, text string, w *gui.Window) {
					app := gui.State[ShowcaseApp](w)
					app.NumericPlainValue = cloneFloatPtr(value)
					app.NumericPlainText = text
				},
			}),
			gui.Text(gui.TextCfg{Text: "Committed: " + numericValueText(app.NumericPlainValue), TextStyle: bodyStyle}),
		},
	})
}

func demoColorPicker(w *gui.Window) gui.View {
	app := gui.State[ShowcaseApp](w)
	c := app.ColorPickerColor
	return gui.Column(gui.ContainerCfg{
		Sizing:  gui.FillFit,
		Spacing: gui.Some(float32(12)),
		Padding: gui.Some(gui.PaddingNone),
		Content: []gui.View{
			gui.ColorPicker(gui.ColorPickerCfg{
				ID:    "color-picker",
				Color: c,
				OnColorChange: func(color gui.Color, _ *gui.Event, w *gui.Window) {
					gui.State[ShowcaseApp](w).ColorPickerColor = color
				},
			}),
			gui.Text(gui.TextCfg{
				Text:      fmt.Sprintf("RGBA(%d, %d, %d, %d)", c.R, c.G, c.B, c.A),
				TextStyle: gui.CurrentTheme().N3,
			}),
		},
	})
}

func demoDatePicker(w *gui.Window) gui.View {
	app := gui.State[ShowcaseApp](w)
	selected := "none"
	if len(app.DatePickerDates) > 0 {
		parts := make([]string, 0, len(app.DatePickerDates))
		for _, date := range app.DatePickerDates {
			parts = append(parts, gui.LocaleFormatDate(date, gui.CurrentLocale().Date.ShortDate))
		}
		selected = strings.Join(parts, ", ")
	}

	return gui.Column(gui.ContainerCfg{
		Sizing:  gui.FillFit,
		Spacing: gui.Some(float32(10)),
		Padding: gui.Some(gui.PaddingNone),
		Content: []gui.View{
			gui.DatePicker(gui.DatePickerCfg{
				ID:             "date-picker",
				Dates:          app.DatePickerDates,
				SelectMultiple: true,
				OnSelect: func(dates []time.Time, _ *gui.Event, w *gui.Window) {
					gui.State[ShowcaseApp](w).DatePickerDates = append([]time.Time(nil), dates...)
				},
			}),
			gui.Text(gui.TextCfg{Text: "Selected: " + selected, TextStyle: gui.CurrentTheme().N3}),
		},
	})
}

func demoDatePickerRoller(w *gui.Window) gui.View {
	app := gui.State[ShowcaseApp](w)
	return gui.Column(gui.ContainerCfg{
		Sizing:  gui.FillFit,
		Spacing: gui.Some(float32(10)),
		Padding: gui.Some(gui.PaddingNone),
		Content: []gui.View{
			gui.DatePickerRoller(gui.DatePickerRollerCfg{
				ID:           "date-roller",
				SelectedDate: app.RollerDate,
				OnChange: func(date time.Time, w *gui.Window) {
					gui.State[ShowcaseApp](w).RollerDate = date
				},
			}),
			gui.Text(gui.TextCfg{
				Text:      "Selected: " + gui.LocaleFormatDate(app.RollerDate, gui.CurrentLocale().Date.LongDate),
				TextStyle: gui.CurrentTheme().N3,
			}),
		},
	})
}

func demoInputDate(w *gui.Window) gui.View {
	app := gui.State[ShowcaseApp](w)
	return gui.Column(gui.ContainerCfg{
		Sizing:  gui.FillFit,
		Spacing: gui.Some(float32(10)),
		Padding: gui.Some(gui.PaddingNone),
		Content: []gui.View{
			gui.InputDate(gui.InputDateCfg{
				ID:          "input-date",
				Date:        app.InputDate,
				Placeholder: "Pick a date",
				Sizing:      gui.FillFit,
				OnSelect: func(dates []time.Time, _ *gui.Event, w *gui.Window) {
					if len(dates) == 0 {
						return
					}
					gui.State[ShowcaseApp](w).InputDate = dates[0]
				},
			}),
			gui.Text(gui.TextCfg{
				Text:      "Selected: " + gui.LocaleFormatDate(app.InputDate, gui.CurrentLocale().Date.ShortDate),
				TextStyle: gui.CurrentTheme().N3,
			}),
		},
	})
}

func demoForms(w *gui.Window) gui.View {
	app := gui.State[ShowcaseApp](w)
	form := &app.Form
	summary := formSummary(*form)
	pendingFields := formPendingFields(*form)
	pendingText := ""
	if len(pendingFields) > 0 {
		pendingText = "Validating: " + strings.Join(pendingFields, ", ")
	}

	return gui.Column(gui.ContainerCfg{
		Sizing:  gui.FillFit,
		Spacing: gui.Some(float32(12)),
		Padding: gui.Some(gui.PaddingNone),
		Content: []gui.View{
			showcaseFormRow("Username", gui.Input(gui.InputCfg{
				ID:          "showcase-form-username",
				IDFocus:     9180,
				Width:       260,
				Sizing:      gui.FixedFit,
				Text:        form.Username,
				Placeholder: "username",
				OnTextChanged: func(_ *gui.Layout, s string, w *gui.Window) {
					app := gui.State[ShowcaseApp](w)
					app.Form.Username = s
					app.Form.UsernameState.Dirty = strings.TrimSpace(s) != ""
					app.Form.UsernameState.Issue = validateUsernameSync(s)
					startUsernameAsyncValidation(s, w)
				},
				OnBlur: func(_ *gui.Layout, w *gui.Window) {
					app := gui.State[ShowcaseApp](w)
					app.Form.UsernameState.Touched = true
					app.Form.UsernameState.Issue = validateUsernameSync(app.Form.Username)
				},
			})),
			showcaseFormState("Username", form.UsernameState),
			showcaseFormIssue("username", form.UsernameState.Issue),

			showcaseFormRow("Email", gui.Input(gui.InputCfg{
				ID:          "showcase-form-email",
				IDFocus:     9181,
				Width:       260,
				Sizing:      gui.FixedFit,
				Text:        form.Email,
				Placeholder: "user@example.com",
				OnTextChanged: func(_ *gui.Layout, s string, w *gui.Window) {
					app := gui.State[ShowcaseApp](w)
					app.Form.Email = s
					app.Form.EmailState.Dirty = strings.TrimSpace(s) != ""
					app.Form.EmailState.Issue = validateEmailSync(s)
				},
				OnBlur: func(_ *gui.Layout, w *gui.Window) {
					app := gui.State[ShowcaseApp](w)
					app.Form.EmailState.Touched = true
					app.Form.EmailState.Issue = validateEmailSync(app.Form.Email)
				},
			})),
			showcaseFormState("Email", form.EmailState),
			showcaseFormIssue("email", form.EmailState.Issue),

			showcaseFormRow("Age", gui.NumericInput(gui.NumericInputCfg{
				ID:       "showcase-form-age",
				IDFocus:  9182,
				Width:    120,
				Sizing:   gui.FixedFit,
				Decimals: 0,
				Min:      float64p(0),
				Max:      float64p(120),
				Text:     form.AgeText,
				Value:    form.AgeValue,
				OnTextChanged: func(_ *gui.Layout, text string, w *gui.Window) {
					app := gui.State[ShowcaseApp](w)
					app.Form.AgeText = text
					app.Form.AgeState.Dirty = strings.TrimSpace(text) != ""
					app.Form.AgeState.Issue = validateAgeSync(text)
				},
				OnValueCommit: func(_ *gui.Layout, value *float64, text string, w *gui.Window) {
					app := gui.State[ShowcaseApp](w)
					app.Form.AgeValue = cloneFloatPtr(value)
					app.Form.AgeText = text
					app.Form.AgeState.Touched = true
					app.Form.AgeState.Issue = validateAgeSync(text)
				},
			})),
			showcaseFormState("Age", form.AgeState),
			showcaseFormIssue("age", form.AgeState.Issue),

			gui.Row(gui.ContainerCfg{
				Sizing:  gui.FillFit,
				Padding: gui.Some(gui.PaddingNone),
				Spacing: gui.Some(float32(8)),
				Content: []gui.View{
					gui.Button(gui.ButtonCfg{
						ID:      "showcase-form-submit",
						Padding: gui.Some(gui.NewPadding(8, 16, 8, 16)),
						Content: []gui.View{gui.Text(gui.TextCfg{Text: gui.CurrentLocale().StrSubmit, TextStyle: gui.CurrentTheme().B3})},
						OnClick: func(_ *gui.Layout, e *gui.Event, w *gui.Window) {
							submitShowcaseForm(w)
							e.IsHandled = true
						},
					}),
					gui.Button(gui.ButtonCfg{
						ID:      "showcase-form-reset",
						Padding: gui.Some(gui.NewPadding(8, 16, 8, 16)),
						Content: []gui.View{gui.Text(gui.TextCfg{Text: gui.CurrentLocale().StrReset, TextStyle: gui.CurrentTheme().N3})},
						OnClick: func(_ *gui.Layout, e *gui.Event, w *gui.Window) {
							resetShowcaseForm(gui.State[ShowcaseApp](w))
							e.IsHandled = true
						},
					}),
				},
			}),
			showcaseFormIssue("", fmt.Sprintf("Validation summary: invalid=%d, pending=%d", summary.InvalidCount, summary.PendingCount)),
			showcaseFormIssue("", pendingText),
			showcaseFormIssue("", func() string {
				if form.SubmitMessage != "" {
					return form.SubmitMessage
				}
				return "Submit form to view committed values"
			}()),
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

func showcaseFormRow(label string, field gui.View) gui.View {
	return gui.Row(gui.ContainerCfg{
		Sizing:  gui.FillFit,
		Padding: gui.Some(gui.PaddingNone),
		Spacing: gui.Some(float32(8)),
		VAlign:  gui.VAlignMiddle,
		Content: []gui.View{
			gui.Text(gui.TextCfg{
				Text:      label,
				TextStyle: gui.CurrentTheme().N3,
				MinWidth:  90,
				Sizing:    gui.FixedFit,
			}),
			field,
		},
	})
}

func showcaseFormState(label string, state ShowcaseFieldState) gui.View {
	return gui.Text(gui.TextCfg{
		Text:      fmt.Sprintf("%s: touched=%t, dirty=%t, pending=%t", label, state.Touched, state.Dirty, state.Pending),
		TextStyle: gui.CurrentTheme().N2,
	})
}

func showcaseFormIssue(fieldID, text string) gui.View {
	style := gui.CurrentTheme().N2
	if fieldID != "" && text != "" {
		style.Color = gui.RGB(219, 87, 87)
		text = fieldID + ": " + text
	}
	return gui.Text(gui.TextCfg{
		Text:      text,
		TextStyle: style,
	})
}

func numericValueText(value *float64) string {
	if value == nil {
		return "none"
	}
	return fmt.Sprintf("%.2f", *value)
}
