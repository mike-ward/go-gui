package main

import (
	"fmt"

	"github.com/mike-ward/go-gui/gui"
)

func demoToggle(w *gui.Window) gui.View {
	app := gui.State[ShowcaseApp](w)
	return gui.Column(gui.ContainerCfg{
		Sizing:  gui.FillFit,
		Spacing: gui.Some(float32(12)),
		Padding: gui.Some(gui.PaddingNone),
		Content: []gui.View{
			gui.Toggle(gui.ToggleCfg{
				ID:       "toggle-a",
				Label:    "Toggle",
				Selected: app.ToggleA,
				OnClick: func(_ *gui.Layout, _ *gui.Event, w *gui.Window) {
					a := gui.State[ShowcaseApp](w)
					a.ToggleA = !a.ToggleA
				},
			}),
			gui.Checkbox(gui.ToggleCfg{
				ID:       "checkbox-a",
				Label:    "Checkbox",
				Selected: app.CheckboxA,
				OnClick: func(_ *gui.Layout, _ *gui.Event, w *gui.Window) {
					a := gui.State[ShowcaseApp](w)
					a.CheckboxA = !a.CheckboxA
				},
			}),
		},
	})
}

func demoSwitch(w *gui.Window) gui.View {
	app := gui.State[ShowcaseApp](w)
	return gui.Row(gui.ContainerCfg{
		Sizing:  gui.FillFit,
		Spacing: gui.Some(float32(12)),
		Padding: gui.Some(gui.PaddingNone),
		VAlign:  gui.VAlignMiddle,
		Content: []gui.View{
			gui.Switch(gui.SwitchCfg{
				ID:       "switch-a",
				Label:    "Enable feature",
				Selected: app.SwitchA,
				OnClick: func(_ *gui.Layout, _ *gui.Event, w *gui.Window) {
					a := gui.State[ShowcaseApp](w)
					a.SwitchA = !a.SwitchA
				},
			}),
		},
	})
}

func demoRadio(w *gui.Window) gui.View {
	app := gui.State[ShowcaseApp](w)
	options := []struct{ label, value string }{
		{"Go", "go"},
		{"Rust", "rust"},
		{"Zig", "zig"},
	}
	views := make([]gui.View, len(options))
	for i, opt := range options {
		v := opt.value
		views[i] = gui.Radio(gui.RadioCfg{
			ID:       "radio-" + v,
			Label:    opt.label,
			Selected: app.RadioValue == v,
			OnClick: func(_ *gui.Layout, _ *gui.Event, w *gui.Window) {
				gui.State[ShowcaseApp](w).RadioValue = v
			},
		})
	}
	return gui.Column(gui.ContainerCfg{
		Sizing:  gui.FillFit,
		Spacing: gui.Some(float32(8)),
		Padding: gui.Some(gui.PaddingNone),
		Content: views,
	})
}

func demoRadioGroup(w *gui.Window) gui.View {
	app := gui.State[ShowcaseApp](w)
	t := gui.CurrentTheme()
	return gui.Column(gui.ContainerCfg{
		Sizing:  gui.FillFit,
		Spacing: gui.Some(float32(16)),
		Padding: gui.Some(gui.PaddingNone),
		Content: []gui.View{
			gui.Text(gui.TextCfg{Text: "Column layout", TextStyle: t.B3}),
			gui.RadioButtonGroupColumn(gui.RadioButtonGroupCfg{
				Value: app.RadioValue,
				Options: []gui.RadioOption{
					gui.NewRadioOption("Go", "go"),
					gui.NewRadioOption("Rust", "rust"),
					gui.NewRadioOption("Zig", "zig"),
				},
				OnSelect: func(v string, w *gui.Window) {
					gui.State[ShowcaseApp](w).RadioValue = v
				},
			}),
			gui.Text(gui.TextCfg{Text: "Row layout", TextStyle: t.B3}),
			gui.RadioButtonGroupRow(gui.RadioButtonGroupCfg{
				Value: app.RadioValue,
				Options: []gui.RadioOption{
					gui.NewRadioOption("Go", "go"),
					gui.NewRadioOption("Rust", "rust"),
					gui.NewRadioOption("Zig", "zig"),
				},
				OnSelect: func(v string, w *gui.Window) {
					gui.State[ShowcaseApp](w).RadioValue = v
				},
			}),
		},
	})
}

func demoSelect(w *gui.Window) gui.View {
	app := gui.State[ShowcaseApp](w)
	t := gui.CurrentTheme()
	return gui.Column(gui.ContainerCfg{
		Sizing:  gui.FillFit,
		Spacing: gui.Some(float32(12)),
		Padding: gui.Some(gui.PaddingNone),
		Content: []gui.View{
			gui.Text(gui.TextCfg{
				Text:      fmt.Sprintf("Selected: %v", app.SelectValue),
				TextStyle: t.N3,
			}),
			sectionLabel(t, "Single Select"),
			gui.Select(gui.SelectCfg{
				ID:          "select-single",
				Placeholder: "Pick a language",
				Selected:    app.SelectValue,
				Options:     []string{"Go", "Rust", "Zig", "C", "Python", "TypeScript"},
				OnSelect: func(sel []string, _ *gui.Event, w *gui.Window) {
					gui.State[ShowcaseApp](w).SelectValue = sel
				},
			}),
			sectionLabel(t, "Multi-Select"),
			gui.Select(gui.SelectCfg{
				ID:             "select-multi",
				Placeholder:    "Pick languages",
				Selected:       app.SelectValue,
				SelectMultiple: true,
				Options:        []string{"Go", "Rust", "Zig", "C", "Python", "TypeScript"},
				OnSelect: func(sel []string, _ *gui.Event, w *gui.Window) {
					gui.State[ShowcaseApp](w).SelectValue = sel
				},
			}),
		},
	})
}

func demoListBox(w *gui.Window) gui.View {
	app := gui.State[ShowcaseApp](w)
	t := gui.CurrentTheme()
	return gui.Column(gui.ContainerCfg{
		Sizing:  gui.FillFit,
		Spacing: gui.Some(float32(12)),
		Padding: gui.Some(gui.PaddingNone),
		Content: []gui.View{
			gui.Text(gui.TextCfg{
				Text:      fmt.Sprintf("Selected: %v", app.ListBoxSelected),
				TextStyle: t.N3,
			}),
			gui.ListBox(gui.ListBoxCfg{
				ID:          "listbox-demo",
				Sizing:      gui.FillFit,
				MaxHeight:   200,
				Multiple:    true,
				SelectedIDs: app.ListBoxSelected,
				Data: []gui.ListBoxOption{
					gui.NewListBoxSubheading("h-compiled", "Compiled"),
					gui.NewListBoxOption("go", "Go", "go"),
					gui.NewListBoxOption("rust", "Rust", "rust"),
					gui.NewListBoxOption("zig", "Zig", "zig"),
					gui.NewListBoxOption("c", "C", "c"),
					gui.NewListBoxSubheading("h-interp", "Interpreted"),
					gui.NewListBoxOption("python", "Python", "python"),
					gui.NewListBoxOption("js", "JavaScript", "js"),
					gui.NewListBoxOption("ruby", "Ruby", "ruby"),
				},
				OnSelect: func(ids []string, _ *gui.Event, w *gui.Window) {
					gui.State[ShowcaseApp](w).ListBoxSelected = ids
				},
			}),
		},
	})
}

func demoCombobox(w *gui.Window) gui.View {
	app := gui.State[ShowcaseApp](w)
	t := gui.CurrentTheme()
	return gui.Column(gui.ContainerCfg{
		Sizing:  gui.FillFit,
		Spacing: gui.Some(float32(12)),
		Padding: gui.Some(gui.PaddingNone),
		Content: []gui.View{
			gui.Text(gui.TextCfg{
				Text:      "Selected: " + app.ComboboxValue,
				TextStyle: t.N3,
			}),
			gui.Combobox(gui.ComboboxCfg{
				ID:          "combobox-demo",
				Placeholder: "Type to search...",
				Value:       app.ComboboxValue,
				Options:     []string{"Go", "Rust", "Zig", "C", "C++", "Python", "TypeScript", "JavaScript", "Ruby", "Elixir"},
				Sizing:      gui.FillFit,
				OnSelect: func(v string, _ *gui.Event, w *gui.Window) {
					gui.State[ShowcaseApp](w).ComboboxValue = v
				},
			}),
		},
	})
}

func demoDragReorder(_ *gui.Window) gui.View {
	return gui.Column(gui.ContainerCfg{
		Sizing:  gui.FillFit,
		Padding: gui.Some(gui.PaddingNone),
		Spacing: gui.Some(float32(16)),
		Content: []gui.View{
			placeholderHeader("Drag reorder is not implemented in go-gui yet."),
			gui.Text(gui.TextCfg{
				Text:      "Drag items to reorder, or use Alt+Arrow keys. Escape cancels.",
				TextStyle: gui.CurrentTheme().N3,
			}),
			dragMockSection("List Box", []string{"Apple", "Banana", "Cherry", "Date", "Elderberry", "Fig"}),
			dragMockSection("Tab Control", []string{"Alpha", "Beta", "Gamma", "Delta"}),
			dragMockSection("Tree View", []string{"src", "  main.v", "  util.v", "  app.v", "docs", "  README.md", "  GUIDE.md", "tests", "build"}),
		},
	})
}

func demoRangeSlider(w *gui.Window) gui.View {
	app := gui.State[ShowcaseApp](w)
	t := gui.CurrentTheme()
	return gui.Column(gui.ContainerCfg{
		Sizing:  gui.FillFit,
		Spacing: gui.Some(float32(12)),
		Padding: gui.Some(gui.PaddingNone),
		Content: []gui.View{
			gui.Text(gui.TextCfg{
				Text:      fmt.Sprintf("Value: %.0f", app.RangeValue),
				TextStyle: t.N3,
			}),
			gui.RangeSlider(gui.RangeSliderCfg{
				ID:     "slider-demo",
				Value:  app.RangeValue,
				Min:    0,
				Max:    100,
				Sizing: gui.FillFit,
				OnChange: func(v float32, _ *gui.Event, w *gui.Window) {
					gui.State[ShowcaseApp](w).RangeValue = v
				},
			}),
		},
	})
}

func dragMockSection(title string, items []string) gui.View {
	content := []gui.View{
		gui.Text(gui.TextCfg{Text: title, TextStyle: gui.CurrentTheme().B3}),
		showcaseWrappedText(
			"Faithful placeholder: the original showcase supports drag and Alt+Arrow reordering here.",
			gui.CurrentTheme().N2,
		),
	}
	for _, item := range items {
		content = append(content, gui.Row(gui.ContainerCfg{
			Sizing:  gui.FillFit,
			Padding: gui.Some(gui.NewPadding(6, 10, 6, 10)),
			Color:   gui.CurrentTheme().ColorPanel,
			Radius:  gui.Some(float32(6)),
			Content: []gui.View{
				gui.Text(gui.TextCfg{Text: "⋮⋮ " + item, TextStyle: gui.CurrentTheme().N3}),
			},
		}))
	}
	return gui.Column(gui.ContainerCfg{
		Sizing:  gui.FillFit,
		Spacing: gui.Some(float32(6)),
		Padding: gui.Some(gui.PaddingNone),
		Content: content,
	})
}
