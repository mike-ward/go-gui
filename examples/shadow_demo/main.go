package main

import (
	"github.com/mike-ward/go-gui/gui"
	"github.com/mike-ward/go-gui/gui/backend"
)

type App struct {
	LightTheme bool
}

func main() {
	gui.SetTheme(gui.ThemeLightNoPadding)

	w := gui.NewWindow(gui.WindowCfg{
		State:  &App{LightTheme: true},
		Title:  "Drop Shadow Demo",
		Width:  800,
		Height: 800,
		OnInit: func(w *gui.Window) {
			w.UpdateView(mainView)
		},
	})

	backend.Run(w)
}

func mainView(w *gui.Window) gui.View {
	app := gui.State[App](w)

	return gui.Column(gui.ContainerCfg{
		Sizing:  gui.FitFit,
		Spacing: gui.Some[float32](40),
		Padding: gui.Some(gui.Padding{Top: 10, Right: 40, Bottom: 40, Left: 40}),
		HAlign:  gui.HAlignCenter,
		Content: []gui.View{
			gui.Row(gui.ContainerCfg{
				Padding: gui.Some(gui.PaddingNone),
				Content: []gui.View{
					gui.Text(gui.TextCfg{
						Text: "Drop Shadow Demo",
						TextStyle: gui.TextStyle{
							Size: 30,
						},
					}),
					gui.Rectangle(gui.RectangleCfg{Width: 100}),
					toggleTheme(app),
				},
			}),
			gui.Row(gui.ContainerCfg{
				Spacing: gui.Some[float32](40),
				Content: []gui.View{
					shadowCard("Soft Shadow\n(Blur: 10, OffsetY: 4)", gui.Color{}, &gui.BoxShadow{
						BlurRadius: 10,
						OffsetY:    4,
						Color:      gui.RGBA(0, 0, 0, 30),
					}),
					shadowCard("Material Elevation\n(Blur: 20, OffsetY: 10)", gui.Color{}, &gui.BoxShadow{
						BlurRadius: 20,
						OffsetY:    10,
						Color:      gui.RGBA(0, 0, 0, 40),
					}),
				},
			}),
			gui.Row(gui.ContainerCfg{
				Spacing: gui.Some[float32](40),
				Content: []gui.View{
					shadowCard("Blue Glow\n(Blur: 30, Color: Blue)", gui.Color{}, &gui.BoxShadow{
						BlurRadius: 30,
						Color:      gui.RGBA(100, 100, 255, 100),
					}),
					shadowCard("Hard Offset\n(Blur: 0, X: 10, Y: 10)", gui.Color{}, &gui.BoxShadow{
						OffsetX: 10,
						OffsetY: 10,
						Color:   gui.RGBA(0, 0, 0, 100),
					}),
				},
			}),
			gui.Row(gui.ContainerCfg{
				Spacing: gui.Some[float32](40),
				Content: []gui.View{
					shadowCard("Blue BG\n(Blur: 15, OffsetY: 5)", gui.LightBlue, &gui.BoxShadow{
						BlurRadius: 15,
						OffsetY:    5,
						Color:      gui.RGBA(0, 0, 0, 50),
					}),
					shadowCard("Orange BG\n(Blur: 20, OffsetY: 8)", gui.Orange, &gui.BoxShadow{
						BlurRadius: 20,
						OffsetY:    8,
						Color:      gui.RGBA(0, 0, 0, 60),
					}),
				},
			}),
		},
	})
}

func shadowCard(text string, bg gui.Color, shadow *gui.BoxShadow) gui.View {
	return gui.Column(gui.ContainerCfg{
		Width:       200,
		Height:      150,
		Radius:      gui.Some[float32](10),
		ColorBorder: gui.Black,
		SizeBorder:  gui.Some[float32](1.5),
		Color:       bg,
		Shadow:      shadow,
		HAlign:      gui.HAlignCenter,
		VAlign:      gui.VAlignMiddle,
		Content: []gui.View{
			gui.Text(gui.TextCfg{
				Text: text,
				TextStyle: gui.TextStyle{
					Color: gui.Black,
					Align: gui.TextAlignCenter,
				},
			}),
		},
	})
}

func toggleTheme(app *App) gui.View {
	return gui.Row(gui.ContainerCfg{
		HAlign:  gui.HAlignEnd,
		Sizing:  gui.FillFit,
		Padding: gui.Some(gui.PaddingNone),
		Spacing: gui.Some[float32](10),
		VAlign:  gui.VAlignMiddle,
		Content: []gui.View{
			gui.Toggle(gui.ToggleCfg{
				TextSelect:   gui.IconMoon,
				TextUnselect: gui.IconSunnyO,
				TextStyle:    gui.CurrentTheme().Icon3,
				Selected:     app.LightTheme,
				ColorSelect:  gui.RGBA(0, 0, 0, 0),
				Padding:      gui.PaddingSmall,
				OnClick: func(_ *gui.Layout, e *gui.Event, w *gui.Window) {
					a := gui.State[App](w)
					if a.LightTheme {
						w.SetTheme(gui.ThemeDarkNoPadding)
					} else {
						w.SetTheme(gui.ThemeLightNoPadding)
					}
					a.LightTheme = !a.LightTheme
					e.IsHandled = true
				},
			}),
		},
	})
}
