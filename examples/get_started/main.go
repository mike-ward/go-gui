package main

import (
	"fmt"

	"github.com/mike-ward/go-gui/gui"
	sdl2 "github.com/mike-ward/go-gui/gui/backend/sdl2"
)

type App struct {
	Clicks int
}

func main() {
	gui.SetTheme(gui.ThemeDarkBordered)

	w := gui.NewWindow(gui.WindowCfg{
		State:  &App{},
		Title:  "get_started",
		Width:  300,
		Height: 300,
		OnInit: func(w *gui.Window) {
			w.UpdateView(mainView)
		},
	})

	sdl2.Run(w)
}

func mainView(w *gui.Window) gui.View {
	ww, wh := w.WindowSize()
	app := gui.State[App](w)

	return gui.Column(gui.ContainerCfg{
		Width:  float32(ww),
		Height: float32(wh),
		Sizing: gui.FixedFixed,
		HAlign: gui.HAlignCenter,
		VAlign: gui.VAlignMiddle,
		Content: []gui.View{
			gui.Text(gui.TextCfg{
				Text:      "Hello GUI! 😀🚀🎉👍",
				TextStyle: gui.CurrentTheme().B1,
			}),
			gui.WithTooltip(w, gui.WithTooltipCfg{
				ID:   "btn_tip",
				Text: "Click to increment counter",
				Content: []gui.View{
					gui.Button(gui.ButtonCfg{
						IDFocus: 1,
						Shadow: &gui.BoxShadow{
							Color:      gui.Color{B: 64, A: 64},
							BlurRadius: 3,
						},
						Content: []gui.View{
							gui.Text(gui.TextCfg{
								Text: fmt.Sprintf("%d Clicks", app.Clicks),
							}),
						},
						OnClick: func(_ *gui.Layout, e *gui.Event, w *gui.Window) {
							gui.State[App](w).Clicks++
							e.IsHandled = true
						},
					}),
				},
			}),
		},
	})
}
