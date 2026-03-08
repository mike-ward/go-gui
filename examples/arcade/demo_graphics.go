package main

import "github.com/mike-ward/go-gui/gui"

func demoRectangle(_ *gui.Window) gui.View {
	t := gui.CurrentTheme()
	return gui.Row(gui.ContainerCfg{
		Sizing:  gui.FillFit,
		Spacing: gui.Some(float32(12)),
		Padding: gui.Some(gui.PaddingNone),
		Content: []gui.View{
			gui.Column(gui.ContainerCfg{
				Width:   80,
				Height:  60,
				Sizing:  gui.FixedFixed,
				Color:   t.ColorActive,
				Radius:  gui.Some(float32(0)),
				HAlign:  gui.HAlignCenter,
				VAlign:  gui.VAlignMiddle,
				Content: []gui.View{gui.Text(gui.TextCfg{Text: "Sharp", TextStyle: t.N2})},
			}),
			gui.Column(gui.ContainerCfg{
				Width:   80,
				Height:  60,
				Sizing:  gui.FixedFixed,
				Color:   t.ColorSelect,
				Radius:  gui.Some(float32(8)),
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
				SizeBorder:  gui.Some(float32(2)),
				Radius:      gui.Some(float32(4)),
				HAlign:      gui.HAlignCenter,
				VAlign:      gui.VAlignMiddle,
				Content:     []gui.View{gui.Text(gui.TextCfg{Text: "Border", TextStyle: t.N2})},
			}),
			gui.Column(gui.ContainerCfg{
				Width:  80,
				Height: 60,
				Sizing: gui.FixedFixed,
				Color:  t.ColorHover,
				Radius: gui.Some(float32(30)),
				HAlign: gui.HAlignCenter,
				VAlign: gui.VAlignMiddle,
				Content: []gui.View{
					gui.Text(gui.TextCfg{Text: "Pill", TextStyle: t.N2}),
				},
			}),
		},
	})
}

func demoIcons(_ *gui.Window) gui.View {
	t := gui.CurrentTheme()
	icons := []struct{ name, icon string }{
		{"Home", gui.IconHome},
		{"Search", gui.IconSearch},
		{"Heart", gui.IconHeart},
		{"Star", gui.IconStar},
		{"Bell", gui.IconBell},
		{"Calendar", gui.IconCalendar},
		{"Camera", gui.IconCamera},
		{"Clock", gui.IconClock},
		{"Cloud", gui.IconCloud},
		{"Code", gui.IconCode},
		{"Download", gui.IconDownload},
		{"Edit", gui.IconEdit},
		{"Eye", gui.IconEye},
		{"Filter", gui.IconFilter},
		{"Globe", gui.IconGlobe},
		{"Info", gui.IconInfo},
		{"Layout", gui.IconLayout},
		{"Plus", gui.IconPlus},
		{"Tag", gui.IconTag},
		{"Trash", gui.IconTrash},
	}

	views := make([]gui.View, len(icons))
	for i, ic := range icons {
		views[i] = gui.Column(gui.ContainerCfg{
			Width:   64,
			Height:  64,
			Sizing:  gui.FixedFixed,
			HAlign:  gui.HAlignCenter,
			VAlign:  gui.VAlignMiddle,
			Spacing: gui.Some(float32(4)),
			Content: []gui.View{
				gui.Text(gui.TextCfg{Text: ic.icon, TextStyle: t.Icon4}),
				gui.Text(gui.TextCfg{Text: ic.name, TextStyle: t.N1}),
			},
		})
	}
	return gui.Wrap(gui.ContainerCfg{
		Sizing:  gui.FillFit,
		Spacing: gui.Some(float32(4)),
		Padding: gui.Some(gui.PaddingNone),
		Content: views,
	})
}

func demoGradient(_ *gui.Window) gui.View {
	return gui.Row(gui.ContainerCfg{
		Sizing:  gui.FillFit,
		Spacing: gui.Some(float32(12)),
		Padding: gui.Some(gui.PaddingNone),
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
				Radius: gui.Some(float32(8)),
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
				Radius: gui.Some(float32(8)),
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
				Radius: gui.Some(float32(8)),
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
	return gui.Row(gui.ContainerCfg{
		Sizing:  gui.FillFit,
		Spacing: gui.Some(float32(24)),
		Padding: gui.Some(gui.NewPadding(16, 16, 16, 16)),
		Content: []gui.View{
			gui.Column(gui.ContainerCfg{
				Width:   100,
				Height:  60,
				Sizing:  gui.FixedFixed,
				Color:   t.ColorPanel,
				Radius:  gui.Some(float32(8)),
				HAlign:  gui.HAlignCenter,
				VAlign:  gui.VAlignMiddle,
				Shadow:  &gui.BoxShadow{OffsetX: 2, OffsetY: 2, BlurRadius: 8, Color: gui.RGBA(0, 0, 0, 60)},
				Content: []gui.View{gui.Text(gui.TextCfg{Text: "Small", TextStyle: t.N3})},
			}),
			gui.Column(gui.ContainerCfg{
				Width:   100,
				Height:  60,
				Sizing:  gui.FixedFixed,
				Color:   t.ColorPanel,
				Radius:  gui.Some(float32(8)),
				HAlign:  gui.HAlignCenter,
				VAlign:  gui.VAlignMiddle,
				Shadow:  &gui.BoxShadow{OffsetX: 4, OffsetY: 4, BlurRadius: 16, Color: gui.RGBA(0, 0, 0, 80)},
				Content: []gui.View{gui.Text(gui.TextCfg{Text: "Medium", TextStyle: t.N3})},
			}),
			gui.Column(gui.ContainerCfg{
				Width:   100,
				Height:  60,
				Sizing:  gui.FixedFixed,
				Color:   t.ColorPanel,
				Radius:  gui.Some(float32(8)),
				HAlign:  gui.HAlignCenter,
				VAlign:  gui.VAlignMiddle,
				Shadow:  &gui.BoxShadow{OffsetX: 0, OffsetY: 8, BlurRadius: 24, Color: gui.RGBA(0, 0, 0, 100)},
				Content: []gui.View{gui.Text(gui.TextCfg{Text: "Large", TextStyle: t.N3})},
			}),
		},
	})
}

func demoSvg(_ *gui.Window) gui.View {
	return gui.Row(gui.ContainerCfg{
		Sizing:  gui.FillFit,
		Spacing: gui.Some(float32(16)),
		Padding: gui.Some(gui.PaddingNone),
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
		},
	})
}

func demoImage(_ *gui.Window) gui.View {
	t := gui.CurrentTheme()
	return gui.Column(gui.ContainerCfg{
		Sizing:  gui.FillFit,
		Spacing: gui.Some(float32(12)),
		Padding: gui.Some(gui.PaddingNone),
		Content: []gui.View{
			gui.Text(gui.TextCfg{
				Text:      "Images load from HTTP URLs (auto-cached by the framework).",
				TextStyle: t.N3,
			}),
			gui.Row(gui.ContainerCfg{
				Sizing:  gui.FillFit,
				Spacing: gui.Some(float32(12)),
				Padding: gui.Some(gui.PaddingNone),
				VAlign:  gui.VAlignBottom,
				Content: []gui.View{
					gui.Column(gui.ContainerCfg{
						Sizing:  gui.FitFit,
						Spacing: gui.Some(float32(4)),
						Padding: gui.Some(gui.PaddingNone),
						HAlign:  gui.HAlignCenter,
						Content: []gui.View{
							gui.Image(gui.ImageCfg{
								Src:    "https://picsum.photos/id/237/200/150",
								Width:  200,
								Height: 150,
							}),
							gui.Text(gui.TextCfg{Text: "200 x 150", TextStyle: t.N2}),
						},
					}),
					gui.Column(gui.ContainerCfg{
						Sizing:  gui.FitFit,
						Spacing: gui.Some(float32(4)),
						Padding: gui.Some(gui.PaddingNone),
						HAlign:  gui.HAlignCenter,
						Content: []gui.View{
							gui.Image(gui.ImageCfg{
								Src:    "https://picsum.photos/id/1015/150/150",
								Width:  150,
								Height: 150,
							}),
							gui.Text(gui.TextCfg{Text: "150 x 150", TextStyle: t.N2}),
						},
					}),
					gui.Column(gui.ContainerCfg{
						Sizing:  gui.FitFit,
						Spacing: gui.Some(float32(4)),
						Padding: gui.Some(gui.PaddingNone),
						HAlign:  gui.HAlignCenter,
						Content: []gui.View{
							gui.Image(gui.ImageCfg{
								Src:    "https://picsum.photos/id/1025/100/100",
								Width:  100,
								Height: 100,
							}),
							gui.Text(gui.TextCfg{Text: "100 x 100", TextStyle: t.N2}),
						},
					}),
				},
			}),
		},
	})
}

func demoShader(_ *gui.Window) gui.View {
	t := gui.CurrentTheme()
	return gui.Column(gui.ContainerCfg{
		Sizing:  gui.FillFit,
		Spacing: gui.Some(float32(12)),
		Padding: gui.Some(gui.PaddingNone),
		Content: []gui.View{
			gui.Text(gui.TextCfg{
				Text:      "Custom fragment shader (Metal + GLSL). Params[0] is animated time.",
				TextStyle: t.N3,
			}),
			gui.Column(gui.ContainerCfg{
				Width:  300,
				Height: 200,
				Sizing: gui.FixedFixed,
				Radius: gui.Some(float32(8)),
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
