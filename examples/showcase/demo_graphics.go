package main

import (
	"sort"
	"strings"

	"github.com/mike-ward/go-gui/gui"
)

func demoRectangle(_ *gui.Window) gui.View {
	t := gui.CurrentTheme()
	return gui.Row(gui.ContainerCfg{
		Sizing:  gui.FillFit,
		Spacing: gui.SomeF(12),
		Padding: gui.NoPadding,
		Content: []gui.View{
			gui.Column(gui.ContainerCfg{
				Width:   80,
				Height:  60,
				Sizing:  gui.FixedFixed,
				Color:   t.ColorActive,
				Radius:  gui.NoRadius,
				HAlign:  gui.HAlignCenter,
				VAlign:  gui.VAlignMiddle,
				Content: []gui.View{gui.Text(gui.TextCfg{Text: "Sharp", TextStyle: t.N2})},
			}),
			gui.Column(gui.ContainerCfg{
				Width:   80,
				Height:  60,
				Sizing:  gui.FixedFixed,
				Color:   t.ColorSelect,
				Radius:  gui.SomeF(8),
				HAlign:  gui.HAlignCenter,
				VAlign:  gui.VAlignMiddle,
				Content: []gui.View{gui.Text(gui.TextCfg{Text: "Rounded", TextStyle: t.N2})},
			}),
			gui.Column(gui.ContainerCfg{
				Width:       80,
				Height:      60,
				Sizing:      gui.FixedFixed,
				Color:       gui.ColorTransparent,
				ColorBorder: t.ColorActive,
				SizeBorder:  gui.SomeF(2),
				Radius:      gui.SomeF(4),
				HAlign:      gui.HAlignCenter,
				VAlign:      gui.VAlignMiddle,
				Content:     []gui.View{gui.Text(gui.TextCfg{Text: "Border", TextStyle: t.N2})},
			}),
			gui.Column(gui.ContainerCfg{
				Width:  80,
				Height: 60,
				Sizing: gui.FixedFixed,
				Color:  t.ColorHover,
				Radius: gui.SomeF(30),
				HAlign: gui.HAlignCenter,
				VAlign: gui.VAlignMiddle,
				Content: []gui.View{
					gui.Text(gui.TextCfg{Text: "Pill", TextStyle: t.N2}),
				},
			}),
		},
	})
}

func demoIcons(w *gui.Window) gui.View {
	t := gui.CurrentTheme()
	keys := make([]string, 0, len(gui.IconLookup))
	for key := range gui.IconLookup {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	cellMinWidth := float32(0)
	for _, key := range keys {
		labelWidth := w.TextWidth(strings.TrimPrefix(key, "icon_"), t.N5)
		if labelWidth > cellMinWidth {
			cellMinWidth = labelWidth
		}
	}

	cols := 5
	rows := make([]gui.View, 0, (len(keys)+cols-1)/cols)
	for i := 0; i < len(keys); i += cols {
		end := i + cols
		if end > len(keys) {
			end = len(keys)
		}

		icons := make([]gui.View, 0, end-i)
		for _, key := range keys[i:end] {
			icons = append(icons, gui.Column(gui.ContainerCfg{
				MinWidth: cellMinWidth,
				Padding:  gui.NoPadding,
				HAlign:   gui.HAlignCenter,
				Content: []gui.View{
					gui.Text(gui.TextCfg{Text: gui.IconLookup[key], TextStyle: t.Icon1}),
					gui.Text(gui.TextCfg{Text: strings.TrimPrefix(key, "icon_"), TextStyle: t.N5}),
				},
			}))
		}

		rows = append(rows, gui.Row(gui.ContainerCfg{
			Sizing:  gui.FillFit,
			Padding: gui.NoPadding,
			Spacing: gui.NoSpacing,
			Content: icons,
		}))
	}

	return gui.Column(gui.ContainerCfg{
		Sizing:  gui.FillFit,
		Spacing: gui.Some(t.SpacingSmall),
		Padding: gui.NoPadding,
		Content: rows,
	})
}

func demoGradient(_ *gui.Window) gui.View {
	return gui.Row(gui.ContainerCfg{
		Sizing:  gui.FillFit,
		Spacing: gui.SomeF(12),
		Padding: gui.NoPadding,
		Content: []gui.View{
			gui.Column(gui.ContainerCfg{
				Width:  120,
				Height: 80,
				Sizing: gui.FixedFixed,
				Gradient: &gui.GradientDef{
					Direction: gui.GradientToRight,
					Stops: []gui.GradientStop{
						{Pos: 0, Color: gui.ColorFromString("#3b82f6")},
						{Pos: 1, Color: gui.ColorFromString("#8b5cf6")},
					},
				},
				Radius: gui.SomeF(8),
				HAlign: gui.HAlignCenter,
				VAlign: gui.VAlignMiddle,
				Content: []gui.View{
					gui.Text(gui.TextCfg{
						Text:      "Horizontal",
						TextStyle: gui.TextStyle{Color: gui.RGB(255, 255, 255), Size: 14},
					}),
				},
			}),
			gui.Column(gui.ContainerCfg{
				Width:  120,
				Height: 80,
				Sizing: gui.FixedFixed,
				Gradient: &gui.GradientDef{
					Direction: gui.GradientToTop,
					Stops: []gui.GradientStop{
						{Pos: 0, Color: gui.ColorFromString("#f97316")},
						{Pos: 1, Color: gui.ColorFromString("#ef4444")},
					},
				},
				Radius: gui.SomeF(8),
				HAlign: gui.HAlignCenter,
				VAlign: gui.VAlignMiddle,
				Content: []gui.View{
					gui.Text(gui.TextCfg{
						Text:      "Vertical",
						TextStyle: gui.TextStyle{Color: gui.RGB(255, 255, 255), Size: 14},
					}),
				},
			}),
			gui.Column(gui.ContainerCfg{
				Width:  120,
				Height: 80,
				Sizing: gui.FixedFixed,
				Gradient: &gui.GradientDef{
					Direction: gui.GradientToTopRight,
					Stops: []gui.GradientStop{
						{Pos: 0, Color: gui.ColorFromString("#10b981")},
						{Pos: 0.5, Color: gui.ColorFromString("#3b82f6")},
						{Pos: 1, Color: gui.ColorFromString("#8b5cf6")},
					},
				},
				Radius: gui.SomeF(8),
				HAlign: gui.HAlignCenter,
				VAlign: gui.VAlignMiddle,
				Content: []gui.View{
					gui.Text(gui.TextCfg{
						Text:      "Diagonal",
						TextStyle: gui.TextStyle{Color: gui.RGB(255, 255, 255), Size: 14},
					}),
				},
			}),
		},
	})
}

func demoBoxShadows(_ *gui.Window) gui.View {
	t := gui.CurrentTheme()
	cardColor := t.ColorBackground
	return gui.Column(gui.ContainerCfg{
		Sizing:  gui.FillFit,
		Spacing: gui.Some(t.SpacingMedium),
		Padding: gui.NoPadding,
		Content: []gui.View{
			gui.Text(gui.TextCfg{
				Text:      "offset_x/offset_y move the shadow. blur_radius controls softness.",
				TextStyle: t.N5,
				Mode:      gui.TextModeWrap,
			}),
			gui.Text(gui.TextCfg{
				Text:      "spread_radius exists in gui.BoxShadow, but this render path does not apply it.",
				TextStyle: t.N5,
				Mode:      gui.TextModeWrap,
			}),
			gui.Row(gui.ContainerCfg{
				Sizing:  gui.FillFit,
				Spacing: gui.SomeF(40),
				Padding: gui.NoPadding,
				Content: []gui.View{
					showcaseShadowCard("Soft depth", "Blur 12, Y 3", cardColor, gui.RGBA(0, 0, 0, 40), 0, 3, 12, 0),
					showcaseShadowCard("Elevated", "Blur 22, Y 10", cardColor, gui.RGBA(0, 0, 0, 55), 0, 10, 22, 0),
				},
			}),
			gui.Row(gui.ContainerCfg{
				Sizing:  gui.FillFit,
				Spacing: gui.SomeF(40),
				Padding: gui.NoPadding,
				Content: []gui.View{
					showcaseShadowCard("Directional", "Blur 10, X 8, Y 8", cardColor, gui.RGBA(0, 0, 0, 65), 8, 8, 10, 0),
					showcaseShadowCard("Blue glow", "Blur 24, no offset", cardColor, gui.RGBA(80, 120, 255, 85), 0, 0, 24, 0),
				},
			}),
			gui.Text(gui.TextCfg{
				Text:      "spread_radius compare: cards below should match today.",
				TextStyle: t.N5,
				Mode:      gui.TextModeWrap,
			}),
			gui.Row(gui.ContainerCfg{
				Sizing:  gui.FillFit,
				Spacing: gui.SomeF(40),
				Padding: gui.NoPadding,
				Content: []gui.View{
					showcaseShadowCard("Spread 0", "spread_radius: 0", cardColor, gui.RGBA(0, 0, 0, 70), 4, 6, 14, 0),
					showcaseShadowCard("Spread 16", "spread_radius: 16", cardColor, gui.RGBA(0, 0, 0, 70), 4, 6, 14, 16),
				},
			}),
		},
	})
}

func showcaseShadowCard(title, note string, bg, shadowColor gui.Color, shadowOffsetX, shadowOffsetY, shadowBlur, shadowSpread float32) gui.View {
	t := gui.CurrentTheme()
	return gui.Column(gui.ContainerCfg{
		Width:       170,
		Height:      96,
		Sizing:      gui.FixedFixed,
		Padding:     gui.Some(gui.NewPadding(10, 10, 10, 10)),
		Spacing:     gui.SomeF(2),
		Radius:      gui.SomeF(10),
		Color:       bg,
		ColorBorder: t.ColorBorder,
		SizeBorder:  gui.SomeF(1),
		Shadow: &gui.BoxShadow{
			Color:        shadowColor,
			OffsetX:      shadowOffsetX,
			OffsetY:      shadowOffsetY,
			BlurRadius:   shadowBlur,
			SpreadRadius: shadowSpread,
		},
		Content: []gui.View{
			gui.Text(gui.TextCfg{Text: title, TextStyle: t.B5}),
			gui.Text(gui.TextCfg{
				Text:      note,
				TextStyle: t.N5,
				Mode:      gui.TextModeWrap,
			}),
		},
	})
}

func demoSvg(_ *gui.Window) gui.View {
	return gui.Row(gui.ContainerCfg{
		Sizing:  gui.FillFit,
		Spacing: gui.SomeF(16),
		Padding: gui.NoPadding,
		Content: []gui.View{
			gui.Svg(gui.SvgCfg{
				ID:     "svg-circle",
				Width:  100,
				Height: 100,
				Sizing: gui.FixedFixed,
				SvgData: `<svg viewBox="0 0 100 100" xmlns="http://www.w3.org/2000/svg">
  <circle cx="50" cy="50" r="40" fill="#3b82f6" opacity="0.8"/>
  <circle cx="50" cy="50" r="25" fill="#8b5cf6" opacity="0.8"/>
  <circle cx="50" cy="50" r="10" fill="#ec4899"/>
</svg>`,
			}),
			gui.Svg(gui.SvgCfg{
				ID:     "svg-star",
				Width:  100,
				Height: 100,
				Sizing: gui.FixedFixed,
				SvgData: `<svg viewBox="0 0 100 100" xmlns="http://www.w3.org/2000/svg">
  <polygon points="50,5 61,35 95,35 68,57 79,91 50,70 21,91 32,57 5,35 39,35"
    fill="#f59e0b" stroke="#d97706" stroke-width="2"/>
</svg>`,
			}),
			gui.Svg(gui.SvgCfg{
				ID:     "svg-bars",
				Width:  120,
				Height: 100,
				Sizing: gui.FixedFixed,
				SvgData: `<svg viewBox="0 0 120 100" xmlns="http://www.w3.org/2000/svg">
  <rect x="10" y="60" width="20" height="35" rx="3" fill="#10b981"/>
  <rect x="35" y="40" width="20" height="55" rx="3" fill="#3b82f6"/>
  <rect x="60" y="20" width="20" height="75" rx="3" fill="#8b5cf6"/>
  <rect x="85" y="50" width="20" height="45" rx="3" fill="#f59e0b"/>
</svg>`,
			}),
			gui.Svg(gui.SvgCfg{
				ID:      "svg-tiger",
				Width:   120,
				Height:  100,
				Sizing:  gui.FixedFixed,
				SvgData: embeddedText("assets/tiger.svg"),
			}),
			gui.Column(gui.ContainerCfg{
				Clip:       true,
				Width:      70,
				Height:     70,
				Sizing:     gui.FixedFixed,
				Padding:    gui.NoPadding,
				SizeBorder: gui.NoBorder,
				Content: []gui.View{
					gui.Svg(gui.SvgCfg{
						ID:     "svg-clip",
						Width:  100,
						Height: 100,
						Sizing: gui.FixedFixed,
						SvgData: `<svg viewBox="0 0 100 100" xmlns="http://www.w3.org/2000/svg">
  <circle cx="50" cy="50" r="45" fill="#3b82f6"/>
  <circle cx="50" cy="50" r="30" fill="#8b5cf6"/>
  <circle cx="50" cy="50" r="15" fill="#ec4899"/>
</svg>`,
					}),
				},
			}),
		},
	})
}

func demoImage(_ *gui.Window) gui.View {
	t := gui.CurrentTheme()
	imgPath := showcaseAssetPath("image_clip_face.jpg")
	return gui.Column(gui.ContainerCfg{
		Sizing:  gui.FillFit,
		Spacing: gui.SomeF(12),
		Padding: gui.NoPadding,
		Content: []gui.View{
			gui.Row(gui.ContainerCfg{
				Sizing:  gui.FillFit,
				Spacing: gui.SomeF(24),
				Padding: gui.NoPadding,
				Content: []gui.View{
					gui.Column(gui.ContainerCfg{
						Sizing:     gui.FitFit,
						Spacing:    gui.SomeF(8),
						Padding:    gui.NoPadding,
						SizeBorder: gui.NoBorder,
						Content: []gui.View{
							gui.Text(gui.TextCfg{Text: "Default", TextStyle: t.B4}),
							gui.Image(gui.ImageCfg{
								Src:    imgPath,
								Width:  120,
								Height: 120,
							}),
						},
					}),
					gui.Column(gui.ContainerCfg{
						Sizing:     gui.FitFit,
						Spacing:    gui.SomeF(8),
						Padding:    gui.NoPadding,
						SizeBorder: gui.NoBorder,
						Content: []gui.View{
							gui.Text(gui.TextCfg{Text: "Rounded (radius: 10)", TextStyle: t.B4}),
							gui.Column(gui.ContainerCfg{
								Clip:       true,
								Radius:     gui.SomeF(10),
								Width:      120,
								Height:     120,
								Sizing:     gui.FixedFixed,
								Padding:    gui.NoPadding,
								SizeBorder: gui.NoBorder,
								Content: []gui.View{
									gui.Image(gui.ImageCfg{
										Src:    imgPath,
										Width:  120,
										Height: 120,
									}),
								},
							}),
						},
					}),
					gui.Column(gui.ContainerCfg{
						Sizing:     gui.FitFit,
						Spacing:    gui.SomeF(8),
						Padding:    gui.NoPadding,
						SizeBorder: gui.NoBorder,
						Content: []gui.View{
							gui.Text(gui.TextCfg{Text: "Circle", TextStyle: t.B4}),
							gui.Circle(gui.ContainerCfg{
								Clip:       true,
								Width:      120,
								Height:     120,
								Sizing:     gui.FixedFixed,
								Padding:    gui.NoPadding,
								SizeBorder: gui.NoBorder,
								Content: []gui.View{
									gui.Image(gui.ImageCfg{
										Src:    imgPath,
										Width:  120,
										Height: 120,
									}),
								},
							}),
						},
					}),
				},
			}),
			gui.Text(gui.TextCfg{
				Text:      "Embedded: assets/image_clip_face.jpg",
				TextStyle: t.N4,
			}),
		},
	})
}

func demoShader(_ *gui.Window) gui.View {
	t := gui.CurrentTheme()
	return gui.Column(gui.ContainerCfg{
		Sizing:  gui.FillFit,
		Spacing: gui.SomeF(12),
		Padding: gui.NoPadding,
		Content: []gui.View{
			gui.Text(gui.TextCfg{
				Text:      "Custom fragment shader (Metal + GLSL). Params[0] is animated time.",
				TextStyle: t.N3,
			}),
			gui.Column(gui.ContainerCfg{
				Width:  300,
				Height: 200,
				Sizing: gui.FixedFixed,
				Radius: gui.SomeF(8),
				Shader: &gui.Shader{
					Metal: `
float2 uv = in.position.xy / uniforms.size;
float t = params[0];
float r = 0.5 + 0.5 * sin(uv.x * 6.28 + t);
float g = 0.5 + 0.5 * sin(uv.y * 6.28 + t * 1.3);
float b = 0.5 + 0.5 * sin((uv.x + uv.y) * 3.14 + t * 0.7);
return float4(r, g, b, 1.0);
`,
					GLSL: `
vec2 uv = gl_FragCoord.xy / u_size;
float t = u_params[0];
float r = 0.5 + 0.5 * sin(uv.x * 6.28 + t);
float g = 0.5 + 0.5 * sin(uv.y * 6.28 + t * 1.3);
float b = 0.5 + 0.5 * sin((uv.x + uv.y) * 3.14 + t * 0.7);
fragColor = vec4(r, g, b, 1.0);
`,
					Params: []float32{0},
				},
			}),
		},
	})
}
