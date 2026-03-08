package main

import (
	"time"

	"github.com/mike-ward/go-gui/gui"
)

func demoLocale(_ *gui.Window) gui.View {
	t := gui.CurrentTheme()
	locale := gui.CurrentLocale()
	now := time.Now()

	localeNames := gui.LocaleRegisteredNames()
	localeViews := make([]gui.View, len(localeNames))
	for i, name := range localeNames {
		selected := locale.ID == name
		color := t.ColorInterior
		ts := t.N2
		if selected {
			color = t.ColorActive
			ts.Color = gui.RGB(255, 255, 255)
		}
		n := name
		localeViews[i] = gui.Button(gui.ButtonCfg{
			ID:      "locale-" + n,
			Color:   color,
			Padding: gui.Some(gui.NewPadding(4, 10, 4, 10)),
			Radius:  gui.Some(float32(12)),
			Content: []gui.View{gui.Text(gui.TextCfg{Text: n, TextStyle: ts})},
			OnClick: func(_ *gui.Layout, e *gui.Event, w *gui.Window) {
				_ = w.SetLocaleID(n)
				e.IsHandled = true
			},
		})
	}

	shortDate := gui.LocaleFormatDate(now, locale.Date.ShortDate)
	longDate := gui.LocaleFormatDate(now, locale.Date.LongDate)

	decSep := string(locale.Number.DecimalSep)
	grpSep := string(locale.Number.GroupSep)

	return gui.Column(gui.ContainerCfg{
		Sizing:  gui.FillFit,
		Spacing: gui.Some(float32(12)),
		Padding: gui.Some(gui.PaddingNone),
		Content: []gui.View{
			gui.Text(gui.TextCfg{Text: "Select a locale to see formatting changes.", TextStyle: t.N3}),
			gui.Wrap(gui.ContainerCfg{
				Sizing:  gui.FillFit,
				Spacing: gui.Some(float32(4)),
				Padding: gui.Some(gui.PaddingNone),
				Content: localeViews,
			}),
			line(),
			gui.Text(gui.TextCfg{Text: "Current: " + locale.ID, TextStyle: t.B3}),
			gui.Text(gui.TextCfg{Text: "Short date: " + shortDate, TextStyle: t.N3}),
			gui.Text(gui.TextCfg{Text: "Long date: " + longDate, TextStyle: t.N3}),
			gui.Text(gui.TextCfg{Text: "Decimal sep: " + decSep + "  Group sep: " + grpSep, TextStyle: t.N3}),
			showcaseWrappedText(
				"OK: "+locale.StrOK+"  Yes: "+locale.StrYes+"  No: "+locale.StrNo+"  Cancel: "+locale.StrCancel,
				t.N3,
			),
		},
	})
}
