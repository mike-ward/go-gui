//go:build ios

// Command ios_demo is an iOS demo app for go-gui.
// Compiled as a c-archive and linked into a native Xcode project.
package main

import "C"
import (
	"fmt"
	"unsafe"

	"github.com/mike-ward/go-gui/gui"
	"github.com/mike-ward/go-gui/gui/backend/ios"
)

type App struct {
	Clicks int
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
				Text:      "Go-Gui on iOS",
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

func init() {
	w := gui.NewWindow(gui.WindowCfg{
		State: &App{},
		OnInit: func(w *gui.Window) {
			w.UpdateView(view)
		},
	})
	ios.SetWindow(w)
}

//export GoGuiStart
func GoGuiStart(layerPtr unsafe.Pointer,
	w, h C.int, scale C.float) {
	ios.Start(layerPtr, int(w), int(h), float32(scale))
}

//export GoGuiRender
func GoGuiRender() { ios.Render() }

//export GoGuiTouchBegan
func GoGuiTouchBegan(x, y C.float) {
	ios.TouchBegan(float32(x), float32(y))
}

//export GoGuiTouchMoved
func GoGuiTouchMoved(x, y C.float) {
	ios.TouchMoved(float32(x), float32(y))
}

//export GoGuiTouchEnded
func GoGuiTouchEnded(x, y C.float) {
	ios.TouchEnded(float32(x), float32(y))
}

//export GoGuiResize
func GoGuiResize(w, h C.int, scale C.float) {
	ios.Resize(int(w), int(h), float32(scale))
}

//export GoGuiDestroy
func GoGuiDestroy() { ios.CleanUp() }

func main() {}
