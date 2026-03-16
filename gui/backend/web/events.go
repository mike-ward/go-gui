//go:build js && wasm

package web

import (
	"syscall/js"
	"unicode/utf8"

	"github.com/mike-ward/go-gui/gui"
)

// registerEvents attaches DOM event listeners to the canvas and
// window. Registered callbacks are appended to b.callbacks to
// prevent garbage collection.
func (b *Backend) registerEvents(w *gui.Window) {
	doc := js.Global().Get("document")
	canvas := b.canvas

	reg := func(target js.Value, name string,
		fn func(js.Value, []js.Value) any) {
		f := js.FuncOf(fn)
		b.callbacks = append(b.callbacks, f)
		target.Call("addEventListener", name, f)
	}

	evt := new(gui.Event)

	reg(canvas, "mousedown", func(_ js.Value, args []js.Value) any {
		e := args[0]
		*evt = gui.Event{
			Type:        gui.EventMouseDown,
			MouseX:      float32(e.Get("offsetX").Float()),
			MouseY:      float32(e.Get("offsetY").Float()),
			MouseButton: mapMouseButton(e.Get("button").Int()),
			Modifiers:   mapModifiers(e),
		}
		w.EventFn(evt)
		return nil
	})

	reg(canvas, "mouseup", func(_ js.Value, args []js.Value) any {
		e := args[0]
		*evt = gui.Event{
			Type:        gui.EventMouseUp,
			MouseX:      float32(e.Get("offsetX").Float()),
			MouseY:      float32(e.Get("offsetY").Float()),
			MouseButton: mapMouseButton(e.Get("button").Int()),
			Modifiers:   mapModifiers(e),
		}
		w.EventFn(evt)
		return nil
	})

	reg(canvas, "mousemove", func(_ js.Value, args []js.Value) any {
		e := args[0]
		*evt = gui.Event{
			Type:      gui.EventMouseMove,
			MouseX:    float32(e.Get("offsetX").Float()),
			MouseY:    float32(e.Get("offsetY").Float()),
			MouseDX:   float32(e.Get("movementX").Float()),
			MouseDY:   float32(e.Get("movementY").Float()),
			Modifiers: mapModifiers(e),
		}
		w.EventFn(evt)
		return nil
	})

	reg(canvas, "wheel", func(_ js.Value, args []js.Value) any {
		e := args[0]
		e.Call("preventDefault")
		*evt = gui.Event{
			Type:      gui.EventMouseScroll,
			ScrollX:   -float32(e.Get("deltaX").Float()) / 120,
			ScrollY:   -float32(e.Get("deltaY").Float()) / 120,
			MouseX:    float32(e.Get("offsetX").Float()),
			MouseY:    float32(e.Get("offsetY").Float()),
			Modifiers: mapModifiers(e),
		}
		w.EventFn(evt)
		return nil
	})

	reg(canvas, "mouseenter", func(_ js.Value, _ []js.Value) any {
		*evt = gui.Event{Type: gui.EventMouseEnter}
		w.EventFn(evt)
		return nil
	})

	reg(canvas, "mouseleave", func(_ js.Value, _ []js.Value) any {
		*evt = gui.Event{Type: gui.EventMouseLeave}
		w.EventFn(evt)
		return nil
	})

	reg(js.Global(), "keydown", func(_ js.Value, args []js.Value) any {
		e := args[0]
		code := e.Get("code").String()
		key := e.Get("key").String()
		mods := mapModifiers(e)

		// Prevent browser defaults for navigation keys.
		if shouldPreventDefault(code) {
			e.Call("preventDefault")
		}

		// Key event.
		kc := mapKeyCode(code)
		*evt = gui.Event{
			Type:      gui.EventKeyDown,
			KeyCode:   kc,
			Modifiers: mods,
			KeyRepeat: e.Get("repeat").Bool(),
		}
		w.EventFn(evt)

		// Generate char event for printable keys.
		if len(key) > 0 && !e.Get("ctrlKey").Bool() &&
			!e.Get("metaKey").Bool() {
			r, sz := utf8.DecodeRuneInString(key)
			if r != utf8.RuneError && sz == len(key) &&
				r >= 32 && r != 127 {
				*evt = gui.Event{
					Type:      gui.EventChar,
					CharCode:  uint32(r),
					IMEText:   key,
					Modifiers: mods,
				}
				w.EventFn(evt)
			}
		}
		return nil
	})

	reg(js.Global(), "keyup", func(_ js.Value, args []js.Value) any {
		e := args[0]
		*evt = gui.Event{
			Type:      gui.EventKeyUp,
			KeyCode:   mapKeyCode(e.Get("code").String()),
			Modifiers: mapModifiers(e),
		}
		w.EventFn(evt)
		return nil
	})

	reg(js.Global(), "compositionupdate",
		func(_ js.Value, args []js.Value) any {
			e := args[0]
			*evt = gui.Event{
				Type:    gui.EventIMEComposition,
				IMEText: e.Get("data").String(),
			}
			w.EventFn(evt)
			return nil
		})

	reg(js.Global(), "resize", func(_ js.Value, _ []js.Value) any {
		ww := js.Global().Get("innerWidth").Int()
		wh := js.Global().Get("innerHeight").Int()
		b.resizeCanvas(ww, wh)
		*evt = gui.Event{
			Type:         gui.EventResized,
			WindowWidth:  ww,
			WindowHeight: wh,
		}
		w.EventFn(evt)
		return nil
	})

	reg(js.Global(), "focus", func(_ js.Value, _ []js.Value) any {
		*evt = gui.Event{Type: gui.EventFocused}
		w.EventFn(evt)
		return nil
	})

	reg(js.Global(), "blur", func(_ js.Value, _ []js.Value) any {
		*evt = gui.Event{Type: gui.EventUnfocused}
		w.EventFn(evt)
		return nil
	})

	reg(doc, "paste", func(_ js.Value, args []js.Value) any {
		e := args[0]
		cd := e.Get("clipboardData")
		if !cd.IsNull() && !cd.IsUndefined() {
			b.lastPasteText = cd.Call("getData", "text/plain").String()
		}
		*evt = gui.Event{Type: gui.EventClipboardPasted}
		w.EventFn(evt)
		return nil
	})

	// Prevent default context menu on canvas.
	reg(canvas, "contextmenu", func(_ js.Value, args []js.Value) any {
		args[0].Call("preventDefault")
		return nil
	})
}

func mapMouseButton(b int) gui.MouseButton {
	switch b {
	case 0:
		return gui.MouseLeft
	case 1:
		return gui.MouseMiddle
	case 2:
		return gui.MouseRight
	default:
		return gui.MouseInvalid
	}
}

func mapModifiers(e js.Value) gui.Modifier {
	var m gui.Modifier
	if e.Get("shiftKey").Bool() {
		m |= gui.ModShift
	}
	if e.Get("ctrlKey").Bool() {
		m |= gui.ModCtrl
	}
	if e.Get("altKey").Bool() {
		m |= gui.ModAlt
	}
	if e.Get("metaKey").Bool() {
		m |= gui.ModSuper
	}
	return m
}

func shouldPreventDefault(code string) bool {
	switch code {
	case "Tab", "ArrowUp", "ArrowDown", "ArrowLeft", "ArrowRight",
		"Backspace", "Space":
		return true
	}
	return false
}

func mapKeyCode(code string) gui.KeyCode {
	// Single-letter keys: KeyA..KeyZ.
	if len(code) == 4 && code[:3] == "Key" {
		ch := code[3]
		if ch >= 'A' && ch <= 'Z' {
			return gui.KeyCode(ch)
		}
	}
	// Digit keys: Digit0..Digit9.
	if len(code) == 6 && code[:5] == "Digit" {
		ch := code[5]
		if ch >= '0' && ch <= '9' {
			return gui.KeyCode(ch)
		}
	}

	switch code {
	case "Space":
		return gui.KeySpace
	case "Enter", "NumpadEnter":
		return gui.KeyEnter
	case "Escape":
		return gui.KeyEscape
	case "Tab":
		return gui.KeyTab
	case "Backspace":
		return gui.KeyBackspace
	case "Delete":
		return gui.KeyDelete
	case "Insert":
		return gui.KeyInsert
	case "ArrowRight":
		return gui.KeyRight
	case "ArrowLeft":
		return gui.KeyLeft
	case "ArrowDown":
		return gui.KeyDown
	case "ArrowUp":
		return gui.KeyUp
	case "PageUp":
		return gui.KeyPageUp
	case "PageDown":
		return gui.KeyPageDown
	case "Home":
		return gui.KeyHome
	case "End":
		return gui.KeyEnd
	case "ShiftLeft":
		return gui.KeyLeftShift
	case "ShiftRight":
		return gui.KeyRightShift
	case "ControlLeft":
		return gui.KeyLeftControl
	case "ControlRight":
		return gui.KeyRightControl
	case "AltLeft":
		return gui.KeyLeftAlt
	case "AltRight":
		return gui.KeyRightAlt
	case "MetaLeft":
		return gui.KeyLeftSuper
	case "MetaRight":
		return gui.KeyRightSuper
	case "Comma":
		return gui.KeyComma
	case "Minus":
		return gui.KeyMinus
	case "Period":
		return gui.KeyPeriod
	case "Slash":
		return gui.KeySlash
	case "Semicolon":
		return gui.KeySemicolon
	case "Equal":
		return gui.KeyEqual
	case "BracketLeft":
		return gui.KeyLeftBracket
	case "Backslash":
		return gui.KeyBackslash
	case "BracketRight":
		return gui.KeyRightBracket
	case "Backquote":
		return gui.KeyGraveAccent
	case "CapsLock":
		return gui.KeyCapsLock
	case "F1":
		return gui.KeyF1
	case "F2":
		return gui.KeyF2
	case "F3":
		return gui.KeyF3
	case "F4":
		return gui.KeyF4
	case "F5":
		return gui.KeyF5
	case "F6":
		return gui.KeyF6
	case "F7":
		return gui.KeyF7
	case "F8":
		return gui.KeyF8
	case "F9":
		return gui.KeyF9
	case "F10":
		return gui.KeyF10
	case "F11":
		return gui.KeyF11
	case "F12":
		return gui.KeyF12
	default:
		return gui.KeyInvalid
	}
}

// cursorCSS maps gui.MouseCursor to CSS cursor values.
var cursorCSS = [11]string{
	gui.CursorDefault:      "default",
	gui.CursorArrow:        "default",
	gui.CursorIBeam:        "text",
	gui.CursorCrosshair:    "crosshair",
	gui.CursorPointingHand: "pointer",
	gui.CursorResizeEW:     "ew-resize",
	gui.CursorResizeNS:     "ns-resize",
	gui.CursorResizeNWSE:   "nwse-resize",
	gui.CursorResizeNESW:   "nesw-resize",
	gui.CursorResizeAll:    "move",
	gui.CursorNotAllowed:   "not-allowed",
}
