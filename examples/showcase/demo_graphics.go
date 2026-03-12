package main

import (
	"sort"
	"strings"
	"time"

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
	return gui.Column(gui.ContainerCfg{
		Sizing:     gui.FillFit,
		Spacing:    gui.SomeF(12),
		Padding:    gui.NoPadding,
		SizeBorder: gui.NoBorder,
		Content: []gui.View{
			gui.Row(gui.ContainerCfg{
				Sizing:     gui.FillFit,
				Spacing:    gui.SomeF(12),
				Padding:    gui.NoPadding,
				SizeBorder: gui.NoBorder,
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
			}),
			gui.Row(gui.ContainerCfg{
				Sizing:     gui.FillFit,
				Spacing:    gui.SomeF(12),
				Padding:    gui.NoPadding,
				SizeBorder: gui.NoBorder,
				Content: []gui.View{
					gui.Column(gui.ContainerCfg{
						Width:  120,
						Height: 80,
						Sizing: gui.FixedFixed,
						Gradient: &gui.GradientDef{
							Type: gui.GradientRadial,
							Stops: []gui.GradientStop{
								{Pos: 0, Color: gui.ColorFromString("#facc15")},
								{Pos: 1, Color: gui.ColorFromString("#f97316")},
							},
						},
						Radius: gui.SomeF(8),
						HAlign: gui.HAlignCenter,
						VAlign: gui.VAlignMiddle,
						Content: []gui.View{
							gui.Text(gui.TextCfg{
								Text:      "Radial",
								TextStyle: gui.TextStyle{Color: gui.RGB(255, 255, 255), Size: 14},
							}),
						},
					}),
					gui.Column(gui.ContainerCfg{
						Width:  120,
						Height: 80,
						Sizing: gui.FixedFixed,
						Gradient: &gui.GradientDef{
							Type: gui.GradientRadial,
							Stops: []gui.GradientStop{
								{Pos: 0, Color: gui.ColorFromString("#ffffff")},
								{Pos: 0.4, Color: gui.ColorFromString("#ec4899")},
								{Pos: 1, Color: gui.ColorFromString("#7c3aed")},
							},
						},
						Radius: gui.SomeF(8),
						HAlign: gui.HAlignCenter,
						VAlign: gui.VAlignMiddle,
						Content: []gui.View{
							gui.Text(gui.TextCfg{
								Text:      "Radial Multi",
								TextStyle: gui.TextStyle{Color: gui.RGB(255, 255, 255), Size: 14},
							}),
						},
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
			gui.Row(gui.ContainerCfg{
				Sizing:  gui.FillFit,
				Spacing: gui.SomeF(40),
				Padding: gui.NoPadding,
				Content: []gui.View{
					showcaseShadowCard("Soft depth", "Blur 12, Y 3", cardColor, gui.RGBA(0, 0, 0, 40), 0, 3, 12),
					showcaseShadowCard("Elevated", "Blur 22, Y 10", cardColor, gui.RGBA(0, 0, 0, 55), 0, 10, 22),
				},
			}),
			gui.Row(gui.ContainerCfg{
				Sizing:  gui.FillFit,
				Spacing: gui.SomeF(40),
				Padding: gui.NoPadding,
				Content: []gui.View{
					showcaseShadowCard("Directional", "Blur 10, X 8, Y 8", cardColor, gui.RGBA(0, 0, 0, 65), 8, 8, 10),
					showcaseShadowCard("Blue glow", "Blur 24, no offset", cardColor, gui.RGBA(80, 120, 255, 85), 0, 0, 24),
				},
			}),
		},
	})
}

func showcaseShadowCard(title, note string, bg, shadowColor gui.Color, shadowOffsetX, shadowOffsetY, shadowBlur float32) gui.View {
	t := gui.CurrentTheme()
	return gui.Column(gui.ContainerCfg{
		Width:       170,
		Height:      96,
		Sizing:      gui.FixedFixed,
		Padding:     gui.SomeP(10, 10, 10, 10),
		Spacing:     gui.SomeF(2),
		Radius:      gui.SomeF(10),
		Color:       bg,
		ColorBorder: t.ColorBorder,
		SizeBorder:  gui.SomeF(1),
		Shadow: &gui.BoxShadow{
			Color:      shadowColor,
			OffsetX:    shadowOffsetX,
			OffsetY:    shadowOffsetY,
			BlurRadius: shadowBlur,
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

func demoDrawCanvas(_ *gui.Window) gui.View {
	chartData := []float32{2, 5, 3, 8, 6, 4, 7, 9, 5, 10, 8, 6, 11, 7}
	t := gui.CurrentTheme()

	return gui.Column(gui.ContainerCfg{
		Sizing:     gui.FillFit,
		Spacing:    gui.SomeF(8),
		Padding:    gui.NoPadding,
		SizeBorder: gui.NoBorder,
		Content: []gui.View{
			gui.Text(gui.TextCfg{
				Text:      "Line chart drawn with DrawCanvas (polyline, filled polygon, circles).",
				TextStyle: t.N3,
				Mode:      gui.TextModeWrap,
			}),
			gui.DrawCanvas(gui.DrawCanvasCfg{
				ID:      "showcase-draw-canvas",
				Version: 1,
				Width:   480,
				Height:  280,
				Color:   gui.RGBA(30, 30, 40, 255),
				Radius:  8,
				Padding: gui.Some(gui.Padding{
					Top: 24, Right: 24, Bottom: 24, Left: 24,
				}),
				OnDraw: func(dc *gui.DrawContext) {
					cw := dc.Width
					ch := dc.Height
					gridColor := gui.RGBA(80, 80, 100, 255)

					// Horizontal grid.
					rows := 5
					for i := range rows + 1 {
						y := ch * float32(i) / float32(rows)
						dc.Line(0, y, cw, y, gridColor, 1)
					}
					// Vertical grid.
					cols := len(chartData) - 1
					for i := range cols + 1 {
						x := cw * float32(i) / float32(cols)
						dc.Line(x, 0, x, ch, gridColor, 1)
					}

					// Data range.
					mn, mx := chartData[0], chartData[0]
					for _, v := range chartData {
						if v < mn {
							mn = v
						}
						if v > mx {
							mx = v
						}
					}
					span := mx - mn
					if span == 0 {
						span = 1
					}

					// Build polyline points.
					pts := make([]float32, 0, len(chartData)*2)
					for i, v := range chartData {
						x := cw * float32(i) / float32(len(chartData)-1)
						y := ch - ch*(v-mn)/span
						pts = append(pts, x, y)
					}

					// Filled area under curve.
					fillColor := gui.RGBA(70, 130, 220, 60)
					for i := 0; i+3 < len(pts); i += 2 {
						dc.FilledPolygon([]float32{
							pts[i], pts[i+1],
							pts[i+2], pts[i+3],
							pts[i+2], ch,
							pts[i], ch,
						}, fillColor)
					}

					// Polyline.
					dc.Polyline(pts, gui.RGBA(70, 130, 220, 255), 2.5)

					// Dot markers.
					for i := 0; i < len(pts); i += 2 {
						dc.FilledCircle(pts[i], pts[i+1], 4,
							gui.RGBA(220, 220, 255, 255))
					}
				},
			}),
		},
	})
}

func demoShader(w *gui.Window) gui.View {
	t := gui.CurrentTheme()
	app := gui.State[ShowcaseApp](w)
	elapsed := float32(time.Since(app.ShaderStartTime).Milliseconds()) / 1000.0

	return gui.Column(gui.ContainerCfg{
		Sizing:  gui.FillFit,
		Spacing: gui.SomeF(12),
		Padding: gui.NoPadding,
		Content: []gui.View{
			gui.Text(gui.TextCfg{
				Text:      "Custom fragment shaders (Metal + GLSL). Params[0] is animated time.",
				TextStyle: t.N3,
			}),
			gui.Row(gui.ContainerCfg{
				Sizing:  gui.FillFit,
				Spacing: gui.SomeF(20),
				Padding: gui.NoPadding,
				Content: []gui.View{
					gui.Column(gui.ContainerCfg{
						Width:  200,
						Height: 200,
						Sizing: gui.FixedFixed,
						Radius: gui.SomeF(16),
						HAlign: gui.HAlignCenter,
						VAlign: gui.VAlignMiddle,
						Shader: &gui.Shader{
							Metal: `
								float t = in.p0.x;
								float2 st = in.uv * 0.5 + 0.5;
								float3 c = 0.5 + 0.5 * cos(t + st.xyx + float3(0,2,4));
								float4 frag_color = float4(c, 1.0);
							`,
							GLSL: `
								float t = p0.x;
								vec2 st = uv * 0.5 + 0.5;
								vec3 c = 0.5 + 0.5 * cos(t + st.xyx + vec3(0,2,4));
								vec4 frag_color = vec4(c, 1.0);
							`,
							Params: []float32{elapsed},
						},
						Content: []gui.View{gui.Text(gui.TextCfg{Text: "Rainbow", TextStyle: t.N2})},
					}),
					gui.Column(gui.ContainerCfg{
						Width:  200,
						Height: 200,
						Sizing: gui.FixedFixed,
						Radius: gui.SomeF(16),
						HAlign: gui.HAlignCenter,
						VAlign: gui.VAlignMiddle,
						Shader: &gui.Shader{
							Metal: `
								float t = in.p0.x;
								float2 st = in.uv * 3.0;
								float v = sin(st.x + t) + sin(st.y + t)
									+ sin(st.x + st.y + t)
									+ sin(length(st) + 1.5 * t);
								v = v * 0.25 + 0.5;
								float3 c = float3(
									sin(v * 3.14159),
									sin(v * 3.14159 + 2.094),
									sin(v * 3.14159 + 4.188));
								c = c * 0.5 + 0.5;
								float4 frag_color = float4(c, 1.0);
							`,
							GLSL: `
								float t = p0.x;
								vec2 st = uv * 3.0;
								float v = sin(st.x + t) + sin(st.y + t)
									+ sin(st.x + st.y + t)
									+ sin(length(st) + 1.5 * t);
								v = v * 0.25 + 0.5;
								vec3 c = vec3(
									sin(v * 3.14159),
									sin(v * 3.14159 + 2.094),
									sin(v * 3.14159 + 4.188));
								c = c * 0.5 + 0.5;
								vec4 frag_color = vec4(c, 1.0);
							`,
							Params: []float32{elapsed},
						},
						Content: []gui.View{gui.Text(gui.TextCfg{Text: "Plasma", TextStyle: t.N2})},
					}),
				},
			}),
		},
	})
}
