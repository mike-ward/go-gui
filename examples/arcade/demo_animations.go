package main

import (
	"fmt"

	"github.com/mike-ward/go-gui/gui"
)

func demoAnimations(w *gui.Window) gui.View {
	t := gui.CurrentTheme()
	app := gui.State[ArcadeApp](w)

	return gui.Column(gui.ContainerCfg{
		Sizing:  gui.FillFit,
		Spacing: gui.Some(float32(16)),
		Padding: gui.Some(gui.PaddingNone),
		Content: []gui.View{
			animTweenDemo(t, app),
			line(),
			animSpringDemo(t, app),
			line(),
			animKeyframeDemo(t, app),
		},
	})
}

func animTweenDemo(t gui.Theme, app *ArcadeApp) gui.View {
	return gui.Column(gui.ContainerCfg{
		Sizing:  gui.FillFit,
		Spacing: gui.Some(float32(8)),
		Padding: gui.Some(gui.PaddingNone),
		Content: []gui.View{
			gui.Text(gui.TextCfg{Text: "Tween Animation", TextStyle: t.B4}),
			gui.Text(gui.TextCfg{
				Text:      fmt.Sprintf("Position: %.0f", app.AnimTweenX),
				TextStyle: t.N3,
			}),
			gui.Row(gui.ContainerCfg{
				Sizing:  gui.FillFit,
				Padding: gui.Some(gui.PaddingNone),
				Content: []gui.View{
					gui.Column(gui.ContainerCfg{
						Width:  app.AnimTweenX,
						Sizing: gui.FixedFit,
					}),
					gui.Column(gui.ContainerCfg{
						Width:   24,
						Height:  24,
						Sizing:  gui.FixedFixed,
						Color:   t.ColorActive,
						Radius:  gui.Some(float32(12)),
						Content: []gui.View{},
					}),
				},
			}),
			gui.Row(gui.ContainerCfg{
				Sizing:  gui.FillFit,
				Spacing: gui.Some(float32(8)),
				Padding: gui.Some(gui.PaddingNone),
				Content: []gui.View{
					gui.Button(gui.ButtonCfg{
						ID:      "btn-tween-go",
						Padding: gui.Some(gui.NewPadding(6, 16, 6, 16)),
						Content: []gui.View{gui.Text(gui.TextCfg{Text: "Animate", TextStyle: t.N3})},
						OnClick: func(_ *gui.Layout, e *gui.Event, w *gui.Window) {
							app := gui.State[ArcadeApp](w)
							target := float32(300)
							if app.AnimTweenX > 100 {
								target = 0
							}
							a := gui.NewTweenAnimation("arcade-tween", app.AnimTweenX, target,
								func(v float32, w *gui.Window) {
									gui.State[ArcadeApp](w).AnimTweenX = v
								})
							w.AnimationAdd(a)
							e.IsHandled = true
						},
					}),
				},
			}),
		},
	})
}

func animSpringDemo(t gui.Theme, app *ArcadeApp) gui.View {
	return gui.Column(gui.ContainerCfg{
		Sizing:  gui.FillFit,
		Spacing: gui.Some(float32(8)),
		Padding: gui.Some(gui.PaddingNone),
		Content: []gui.View{
			gui.Text(gui.TextCfg{Text: "Spring Animation", TextStyle: t.B4}),
			gui.Text(gui.TextCfg{
				Text:      fmt.Sprintf("Position: %.0f", app.AnimSpringX),
				TextStyle: t.N3,
			}),
			gui.Row(gui.ContainerCfg{
				Sizing:  gui.FillFit,
				Padding: gui.Some(gui.PaddingNone),
				Content: []gui.View{
					gui.Column(gui.ContainerCfg{
						Width:  app.AnimSpringX,
						Sizing: gui.FixedFit,
					}),
					gui.Column(gui.ContainerCfg{
						Width:   24,
						Height:  24,
						Sizing:  gui.FixedFixed,
						Color:   gui.ColorFromString("#10b981"),
						Radius:  gui.Some(float32(4)),
						Content: []gui.View{},
					}),
				},
			}),
			gui.Row(gui.ContainerCfg{
				Sizing:  gui.FillFit,
				Spacing: gui.Some(float32(8)),
				Padding: gui.Some(gui.PaddingNone),
				Content: []gui.View{
					gui.Button(gui.ButtonCfg{
						ID:      "btn-spring-go",
						Padding: gui.Some(gui.NewPadding(6, 16, 6, 16)),
						Content: []gui.View{gui.Text(gui.TextCfg{Text: "Spring", TextStyle: t.N3})},
						OnClick: func(_ *gui.Layout, e *gui.Event, w *gui.Window) {
							app := gui.State[ArcadeApp](w)
							target := float32(300)
							if app.AnimSpringX > 100 {
								target = 0
							}
							a := gui.NewSpringAnimation("arcade-spring",
								func(v float32, w *gui.Window) {
									gui.State[ArcadeApp](w).AnimSpringX = v
								})
							a.SpringTo(app.AnimSpringX, target)
							w.AnimationAdd(a)
							e.IsHandled = true
						},
					}),
				},
			}),
		},
	})
}

func animKeyframeDemo(t gui.Theme, app *ArcadeApp) gui.View {
	return gui.Column(gui.ContainerCfg{
		Sizing:  gui.FillFit,
		Spacing: gui.Some(float32(8)),
		Padding: gui.Some(gui.PaddingNone),
		Content: []gui.View{
			gui.Text(gui.TextCfg{Text: "Keyframe Animation", TextStyle: t.B4}),
			gui.Text(gui.TextCfg{
				Text:      fmt.Sprintf("Position: %.0f", app.AnimKeyframeX),
				TextStyle: t.N3,
			}),
			gui.Row(gui.ContainerCfg{
				Sizing:  gui.FillFit,
				Padding: gui.Some(gui.PaddingNone),
				Content: []gui.View{
					gui.Column(gui.ContainerCfg{
						Width:  app.AnimKeyframeX,
						Sizing: gui.FixedFit,
					}),
					gui.Column(gui.ContainerCfg{
						Width:   24,
						Height:  24,
						Sizing:  gui.FixedFixed,
						Color:   gui.ColorFromString("#f59e0b"),
						Radius:  gui.Some(float32(12)),
						Content: []gui.View{},
					}),
				},
			}),
			gui.Row(gui.ContainerCfg{
				Sizing:  gui.FillFit,
				Spacing: gui.Some(float32(8)),
				Padding: gui.Some(gui.PaddingNone),
				Content: []gui.View{
					gui.Button(gui.ButtonCfg{
						ID:      "btn-keyframe-go",
						Padding: gui.Some(gui.NewPadding(6, 16, 6, 16)),
						Content: []gui.View{gui.Text(gui.TextCfg{Text: "Keyframes", TextStyle: t.N3})},
						OnClick: func(_ *gui.Layout, e *gui.Event, w *gui.Window) {
							a := gui.NewKeyframeAnimation("arcade-keyframe",
								[]gui.Keyframe{
									{At: 0, Value: 0, Easing: gui.EaseLinear},
									{At: 0.25, Value: 300, Easing: gui.EaseOutCubic},
									{At: 0.5, Value: 100, Easing: gui.EaseInOutQuad},
									{At: 0.75, Value: 250, Easing: gui.EaseOutBounce},
									{At: 1.0, Value: 0, Easing: gui.EaseOutCubic},
								},
								func(v float32, w *gui.Window) {
									gui.State[ArcadeApp](w).AnimKeyframeX = v
								})
							w.AnimationAdd(a)
							e.IsHandled = true
						},
					}),
				},
			}),
		},
	})
}
