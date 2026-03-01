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

	b, err := sdl2.New(w)
	if err != nil {
		panic(err)
	}
	defer b.Destroy()
	b.Run(w)
}

func mainView(w *gui.Window) gui.View {
	ww, wh := w.WindowSize()
	app := gui.State[App](w)

	return gui.Column(gui.ContainerCfg{
		Width:   float32(ww),
		Height:  float32(wh),
		Sizing:  gui.FixedFixed,
		HAlign:  gui.HAlignCenter,
		VAlign:  gui.VAlignMiddle,
		Spacing: 10,
		Content: []gui.View{
			gui.Text(gui.TextCfg{
				Text:      "Hello go-gui!",
				TextStyle: gui.CurrentTheme().B1,
			}),
			gui.Button(gui.ButtonCfg{
				IDFocus: 1,
				Content: []gui.View{
					gui.Text(gui.TextCfg{
						Text: fmt.Sprintf("%d Clicks", app.Clicks),
					}),
				},
				OnClick: func(_ *gui.Layout, _ *gui.Event, w *gui.Window) {
					gui.State[App](w).Clicks++
				},
			}),
		},
	})
}
