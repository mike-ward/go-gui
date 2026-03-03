package main

import (
	"fmt"

	"github.com/mike-ward/go-gui/gui"
	sdl2 "github.com/mike-ward/go-gui/gui/backend/sdl2"
)

type App struct {
	MenuOpen bool
	MenuX    float32
	MenuY    float32
	Status   string
}

const idContextMenu uint32 = 100

func main() {
	gui.SetTheme(gui.ThemeDarkBordered)

	w := gui.NewWindow(gui.WindowCfg{
		State:  &App{Status: "Right-click anywhere"},
		Title:  "Context Menu Example",
		Width:  500,
		Height: 400,
		OnInit: func(w *gui.Window) {
			w.UpdateView(mainView)
		},
	})

	sdl2.Run(w)
}

func mainView(w *gui.Window) gui.View {
	ww, wh := w.WindowSize()
	app := gui.State[App](w)

	content := []gui.View{
		gui.Text(gui.TextCfg{
			Text:      "Right-click anywhere for a context menu",
			TextStyle: gui.CurrentTheme().B1,
		}),
		gui.Text(gui.TextCfg{
			Text: app.Status,
		}),
	}

	if app.MenuOpen {
		content = append(content,
			contextMenu(w, app.MenuX, app.MenuY))
	}

	return gui.Column(gui.ContainerCfg{
		Width:  float32(ww),
		Height: float32(wh),
		Sizing: gui.FixedFixed,
		HAlign: gui.HAlignCenter,
		VAlign: gui.VAlignMiddle,
		OnAnyClick: func(l *gui.Layout, e *gui.Event, w *gui.Window) {
			app := gui.State[App](w)
			if e.MouseButton == gui.MouseRight {
				app.MenuOpen = true
				app.MenuX = e.MouseX + l.Shape.X
				app.MenuY = e.MouseY + l.Shape.Y
				w.SetIDFocus(idContextMenu)
			} else {
				app.MenuOpen = false
			}
			e.IsHandled = true
		},
		Content: content,
	})
}

func contextMenu(w *gui.Window, mx, my float32) gui.View {
	return gui.Menu(w, gui.MenubarCfg{
		ID:      "ctx",
		IDFocus: idContextMenu,
		Items: []gui.MenuItemCfg{
			gui.MenuSubtitle("Actions"),
			{ID: "cut", Text: "Cut"},
			{ID: "copy", Text: "Copy"},
			{ID: "paste", Text: "Paste"},
			gui.MenuSeparator(),
			gui.MenuSubmenu("more", "More", []gui.MenuItemCfg{
				{ID: "selectall", Text: "Select All"},
				{ID: "find", Text: "Find"},
			}),
			gui.MenuSeparator(),
			{ID: "delete", Text: "Delete"},
		},
		Action: func(id string, e *gui.Event, w *gui.Window) {
			app := gui.State[App](w)
			app.Status = fmt.Sprintf("Selected: %s", id)
			app.MenuOpen = false
			e.IsHandled = true
		},
		Float:       true,
		FloatAnchor: gui.FloatTopLeft,
		FloatTieOff: gui.FloatTopLeft,
		FloatOffsetX: mx,
		FloatOffsetY: my,
	})
}
