package main

import "github.com/mike-ward/go-gui/gui"

func demoWelcome(w *gui.Window) gui.View {
	return w.Markdown(gui.MarkdownCfg{
		Source: `# Welcome to Arcade

Arcade is the **Go GUI** framework showcase. Browse the component
catalog on the left to explore every widget, layout, and feature.

## Quick Start

` + "```" + `go
w := gui.NewWindow(gui.WindowCfg{
    State: &MyApp{},
    Title: "Hello",
    Width: 800, Height: 600,
    OnInit: func(w *gui.Window) {
        w.UpdateView(mainView)
    },
})
backend.Run(w)
` + "```" + `

## Features

- **Immediate-mode** pipeline — no virtual DOM, no diffing
- **Typed state** slot per window — no globals, no closures
- **Flexible layout** — Row, Column, Wrap, Splitter, Sidebar
- **Rich widgets** — buttons, inputs, tables, date pickers, and more
- **Theming** — dark/light presets or generate your own
- **Accessibility** — ARIA roles, focus management, keyboard navigation

Toggle the theme with the icon button at the bottom of the catalog.
`,
	})
}
