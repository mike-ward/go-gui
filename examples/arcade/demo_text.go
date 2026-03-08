package main

import "github.com/mike-ward/go-gui/gui"

func demoText(_ *gui.Window) gui.View {
	t := gui.CurrentTheme()
	return gui.Column(gui.ContainerCfg{
		Sizing:  gui.FillFit,
		Spacing: gui.Some(float32(8)),
		Padding: gui.Some(gui.PaddingNone),
		Content: []gui.View{
			gui.Text(gui.TextCfg{Text: "N1 — extra small", TextStyle: t.N1}),
			gui.Text(gui.TextCfg{Text: "N2 — small", TextStyle: t.N2}),
			gui.Text(gui.TextCfg{Text: "N3 — normal", TextStyle: t.N3}),
			gui.Text(gui.TextCfg{Text: "N4 — medium", TextStyle: t.N4}),
			gui.Text(gui.TextCfg{Text: "N5 — large", TextStyle: t.N5}),
			gui.Text(gui.TextCfg{Text: "N6 — extra large", TextStyle: t.N6}),
			line(),
			gui.Text(gui.TextCfg{Text: "B3 — bold normal", TextStyle: t.B3}),
			gui.Text(gui.TextCfg{Text: "I3 — italic normal", TextStyle: t.I3}),
			gui.Text(gui.TextCfg{Text: "BI3 — bold italic normal", TextStyle: t.BI3}),
			gui.Text(gui.TextCfg{Text: "M3 — monospace normal", TextStyle: t.M3}),
		},
	})
}

func demoRtf(_ *gui.Window) gui.View {
	t := gui.CurrentTheme()
	return gui.Column(gui.ContainerCfg{
		Sizing:  gui.FillFit,
		Spacing: gui.Some(float32(8)),
		Padding: gui.Some(gui.PaddingNone),
		Content: []gui.View{
			gui.Text(gui.TextCfg{
				Text:      "Rich text supports mixed styles within a single text block.",
				TextStyle: t.N3,
			}),
			gui.Text(gui.TextCfg{Text: "Bold text", TextStyle: t.B4}),
			gui.Text(gui.TextCfg{Text: "Italic text", TextStyle: t.I4}),
			gui.Text(gui.TextCfg{Text: "Monospace text", TextStyle: t.M3}),
			gui.Text(gui.TextCfg{
				Text: "Strikethrough",
				TextStyle: gui.TextStyle{
					Color:         t.N3.Color,
					Size:          t.N3.Size,
					Strikethrough: true,
				},
			}),
			gui.Text(gui.TextCfg{
				Text: "Underlined",
				TextStyle: gui.TextStyle{
					Color:     t.N3.Color,
					Size:      t.N3.Size,
					Underline: true,
				},
			}),
		},
	})
}

func demoMarkdown(w *gui.Window) gui.View {
	return w.Markdown(gui.MarkdownCfg{
		Style: gui.DefaultMarkdownStyle(),
		Source: `## Markdown Demo

This is rendered from a **markdown** string. Supports:

- **Bold**, *italic*, ` + "`code`" + `, ~~strikethrough~~
- [Links](https://github.com)
- Ordered and unordered lists
- Headings (H1–H6)
- Code blocks with syntax highlighting

` + "```go" + `
func main() {
    fmt.Println("Hello, Arcade!")
}
` + "```" + `

> Blockquotes are supported too.

| Column A | Column B | Column C |
|----------|----------|----------|
| Cell 1   | Cell 2   | Cell 3   |
| Cell 4   | Cell 5   | Cell 6   |
`,
	})
}
