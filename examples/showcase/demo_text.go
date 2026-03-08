package main

import "github.com/mike-ward/go-gui/gui"

func demoText(_ *gui.Window) gui.View {
	t := gui.CurrentTheme()
	return gui.Column(gui.ContainerCfg{
		Sizing:  gui.FillFit,
		Spacing: gui.Some(float32(16)),
		Padding: gui.Some(gui.PaddingNone),
		Content: []gui.View{
			// Typography scale
			sectionLabel(t, "Typography Scale"),
			gui.Column(gui.ContainerCfg{
				Sizing:  gui.FillFit,
				Spacing: gui.Some(float32(4)),
				Padding: gui.Some(gui.PaddingNone),
				Content: []gui.View{
					gui.Text(gui.TextCfg{Text: "N1 -- extra small", TextStyle: t.N1}),
					gui.Text(gui.TextCfg{Text: "N2 -- small", TextStyle: t.N2}),
					gui.Text(gui.TextCfg{Text: "N3 -- normal", TextStyle: t.N3}),
					gui.Text(gui.TextCfg{Text: "N4 -- medium", TextStyle: t.N4}),
					gui.Text(gui.TextCfg{Text: "N5 -- large", TextStyle: t.N5}),
					gui.Text(gui.TextCfg{Text: "N6 -- extra large", TextStyle: t.N6}),
				},
			}),

			line(),

			// Weight and style variants
			sectionLabel(t, "Weight & Style"),
			gui.Column(gui.ContainerCfg{
				Sizing:  gui.FillFit,
				Spacing: gui.Some(float32(4)),
				Padding: gui.Some(gui.PaddingNone),
				Content: []gui.View{
					gui.Text(gui.TextCfg{Text: "B3 -- bold", TextStyle: t.B3}),
					gui.Text(gui.TextCfg{Text: "I3 -- italic", TextStyle: t.I3}),
					gui.Text(gui.TextCfg{Text: "BI3 -- bold italic", TextStyle: t.BI3}),
					gui.Text(gui.TextCfg{Text: "M3 -- monospace", TextStyle: t.M3}),
				},
			}),

			line(),

			// Text alignment
			sectionLabel(t, "Alignment"),
			gui.Column(gui.ContainerCfg{
				Sizing:      gui.FillFit,
				Padding:     gui.Some(gui.NewPadding(8, 8, 8, 8)),
				Spacing:     gui.Some(float32(4)),
				Color:       t.ColorPanel,
				Radius:      gui.Some(float32(6)),
				ColorBorder: t.ColorBorder,
				SizeBorder:  gui.Some(float32(1)),
				Content: []gui.View{
					gui.Text(gui.TextCfg{
						Text:      "Left aligned (default)",
						TextStyle: gui.TextStyle{Color: t.N3.Color, Size: t.N3.Size, Align: gui.TextAlignLeft},
						Sizing:    gui.FillFit,
					}),
					gui.Text(gui.TextCfg{
						Text:      "Center aligned",
						TextStyle: gui.TextStyle{Color: t.N3.Color, Size: t.N3.Size, Align: gui.TextAlignCenter},
						Sizing:    gui.FillFit,
					}),
					gui.Text(gui.TextCfg{
						Text:      "Right aligned",
						TextStyle: gui.TextStyle{Color: t.N3.Color, Size: t.N3.Size, Align: gui.TextAlignRight},
						Sizing:    gui.FillFit,
					}),
				},
			}),

			line(),

			// Decorations
			sectionLabel(t, "Decorations"),
			gui.Row(gui.ContainerCfg{
				Sizing:  gui.FillFit,
				Spacing: gui.Some(float32(16)),
				Padding: gui.Some(gui.PaddingNone),
				Content: []gui.View{
					gui.Text(gui.TextCfg{
						Text: "Underlined",
						TextStyle: gui.TextStyle{
							Color: t.N3.Color, Size: t.N3.Size,
							Underline: true,
						},
					}),
					gui.Text(gui.TextCfg{
						Text: "Strikethrough",
						TextStyle: gui.TextStyle{
							Color: t.N3.Color, Size: t.N3.Size,
							Strikethrough: true,
						},
					}),
					gui.Text(gui.TextCfg{
						Text: "Outlined",
						TextStyle: gui.TextStyle{
							Color: gui.ColorTransparent, Size: t.N4.Size,
							StrokeWidth: 1,
							StrokeColor: t.ColorActive,
						},
					}),
					gui.Text(gui.TextCfg{
						Text: " Highlighted ",
						TextStyle: gui.TextStyle{
							Color:   t.N3.Color,
							Size:    t.N3.Size,
							BgColor: gui.RGBA(59, 130, 246, 80),
						},
					}),
				},
			}),

			line(),

			// Emoji / i18n
			sectionLabel(t, "Emoji & Multigrapheme"),
			gui.Row(gui.ContainerCfg{
				Sizing:  gui.FillFit,
				Spacing: gui.Some(float32(12)),
				Padding: gui.Some(gui.PaddingNone),
				VAlign:  gui.VAlignMiddle,
				Content: []gui.View{
					gui.Text(gui.TextCfg{Text: "\U0001f680\U0001f30d\U0001f3b5\U0001f525\u2764\ufe0f\U0001f469\u200d\U0001f4bb", TextStyle: t.N5}),
					gui.Text(gui.TextCfg{Text: "\u0928\u092e\u0938\u094d\u0924\u0947", TextStyle: t.N4}),
					gui.Text(gui.TextCfg{Text: "\u0645\u0631\u062d\u0628\u0627", TextStyle: t.N4}),
					gui.Text(gui.TextCfg{Text: "\u3053\u3093\u306b\u3061\u306f", TextStyle: t.N4}),
					gui.Text(gui.TextCfg{Text: "\ud55c\uad6d\uc5b4", TextStyle: t.N4}),
				},
			}),

			line(),

			// Selectable text
			sectionLabel(t, "Selectable Text"),
			gui.Text(gui.TextCfg{
				ID:        "text-selectable",
				IDFocus:   50,
				Text:      "This text block is selectable and copyable. Click to focus, then use Ctrl+A / Cmd+A to select all and Ctrl+C / Cmd+C to copy.",
				TextStyle: t.N3,
				Mode:      gui.TextModeWrap,
				Sizing:    gui.FillFit,
			}),
		},
	})
}

func demoRtf(_ *gui.Window) gui.View {
	t := gui.CurrentTheme()
	return gui.Column(gui.ContainerCfg{
		Sizing:  gui.FillFit,
		Spacing: gui.Some(float32(16)),
		Padding: gui.Some(gui.PaddingNone),
		Content: []gui.View{
			// Mixed inline styles in a single paragraph
			sectionLabel(t, "Mixed Inline Styles"),
			gui.RTF(gui.RtfCfg{
				RichText: gui.RichText{
					Runs: []gui.RichTextRun{
						gui.RichRun("Rich text supports ", t.N3),
						gui.RichRun("bold", t.B3),
						gui.RichRun(", ", t.N3),
						gui.RichRun("italic", t.I3),
						gui.RichRun(", ", t.N3),
						gui.RichRun("monospace", t.M3),
						gui.RichRun(", and ", t.N3),
						gui.RichRun("colored", gui.TextStyle{
							Color: gui.ColorFromString("#3b82f6"),
							Size:  t.N3.Size,
						}),
						gui.RichRun(" text in a single paragraph.", t.N3),
					},
				},
			}),

			line(),

			// Underline and strikethrough
			sectionLabel(t, "Decorations"),
			gui.RTF(gui.RtfCfg{
				RichText: gui.RichText{
					Runs: []gui.RichTextRun{
						gui.RichRun("Underlined", gui.TextStyle{
							Color: t.N3.Color, Size: t.N3.Size,
							Underline: true,
						}),
						gui.RichRun(" and ", t.N3),
						gui.RichRun("strikethrough", gui.TextStyle{
							Color: t.N3.Color, Size: t.N3.Size,
							Strikethrough: true,
						}),
						gui.RichRun(" within a single text block.", t.N3),
					},
				},
			}),

			line(),

			// Clickable link
			sectionLabel(t, "Links & Abbreviations"),
			gui.RTF(gui.RtfCfg{
				RichText: gui.RichText{
					Runs: []gui.RichTextRun{
						gui.RichRun("Visit the ", t.N3),
						gui.RichLink("go-gui repository", "https://github.com/mike-ward/go-gui", gui.TextStyle{
							Color:     gui.ColorFromString("#3b82f6"),
							Size:      t.N3.Size,
							Underline: true,
						}),
						gui.RichRun(" for more info. ", t.N3),
						gui.RichAbbr("RTF", "Rich Text Format", t.B3),
						gui.RichRun(" stands for Rich Text Format.", t.N3),
					},
				},
			}),

			line(),

			// Multi-line with breaks
			sectionLabel(t, "Line Breaks"),
			gui.RTF(gui.RtfCfg{
				RichText: gui.RichText{
					Runs: []gui.RichTextRun{
						gui.RichRun("First line of text.", t.N3),
						gui.RichBr(),
						gui.RichRun("Second line after a break.", t.N3),
						gui.RichBr(),
						gui.RichRun("Third line with ", t.N3),
						gui.RichRun("mixed styles", t.B3),
						gui.RichRun(".", t.N3),
					},
				},
			}),
		},
	})
}

func demoMarkdown(w *gui.Window) gui.View {
	return w.Markdown(gui.MarkdownCfg{
		Style:  gui.DefaultMarkdownStyle(),
		Source: embeddedText("docs/markdown_demo.md"),
	})
}

func sectionLabel(t gui.Theme, text string) gui.View {
	return gui.Text(gui.TextCfg{Text: text, TextStyle: t.B3})
}
