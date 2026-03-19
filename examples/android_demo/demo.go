//go:build android

// Package androidapp is an Android demo app for go-gui.
// Built with gomobile bind to generate an AAR for Kotlin host.
package androidapp

import (
	"fmt"

	"github.com/mike-ward/go-gui/gui"
	"github.com/mike-ward/go-gui/gui/backend/android"
)

type App struct {
	Clicks int
}

// Init creates the gui.Window and registers it with the backend.
func Init() {
	w := gui.NewWindow(gui.WindowCfg{
		State: &App{},
		OnInit: func(w *gui.Window) {
			w.UpdateView(view)
		},
	})
	android.SetWindow(w)
}

// Start initializes the GLES backend.
func Start(width, height int, scale float32) {
	android.Start(width, height, scale)
}

// Render runs one frame.
func Render() {
	android.Render()
}

// TouchBegan maps a touch-down event.
func TouchBegan(x, y float32) { android.TouchBegan(x, y) }

// TouchMoved maps a touch-move event.
func TouchMoved(x, y float32) { android.TouchMoved(x, y) }

// TouchEnded maps a touch-up event.
func TouchEnded(x, y float32) { android.TouchEnded(x, y) }

// Resize updates the viewport.
func Resize(width, height int, scale float32) {
	android.Resize(width, height, scale)
}

// CleanUp releases all backend resources.
func CleanUp() {
	android.CleanUp()
}

func view(w *gui.Window) gui.View {
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
				Text:      "Go-Gui on Android",
				TextStyle: gui.CurrentTheme().B1,
			}),
			gui.Text(gui.TextCfg{
				Text: "Tap the button to increment.",
			}),
			gui.Button(gui.ButtonCfg{
				IDFocus: 1,
				Content: []gui.View{
					gui.Text(gui.TextCfg{
						Text: fmt.Sprintf("%d Clicks",
							app.Clicks),
					}),
				},
				OnClick: func(_ *gui.Layout,
					e *gui.Event, w *gui.Window) {
					gui.State[App](w).Clicks++
					e.IsHandled = true
				},
			}),
		},
	})
}
