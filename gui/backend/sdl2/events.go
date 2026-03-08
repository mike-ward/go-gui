package sdl2

import (
	"unicode/utf8"

	"github.com/mike-ward/go-gui/gui"
	"github.com/veandco/go-sdl2/sdl"
)

// mapEvent converts an SDL2 event to a gui.Event.
// Returns the event and true to continue, or false to quit.
func mapEvent(ev sdl.Event, _ *Backend) (gui.Event, bool) {
	switch e := ev.(type) {
	case *sdl.QuitEvent:
		return gui.Event{}, false

	case *sdl.MouseButtonEvent:
		btn := mapMouseButton(e.Button)
		if e.Type == sdl.MOUSEBUTTONDOWN {
			return gui.Event{
				Type:        gui.EventMouseDown,
				MouseX:      float32(e.X),
				MouseY:      float32(e.Y),
				MouseButton: btn,
				Modifiers:   mapKeyMod(sdl.GetModState()),
			}, true
		}
		return gui.Event{
			Type:        gui.EventMouseUp,
			MouseX:      float32(e.X),
			MouseY:      float32(e.Y),
			MouseButton: btn,
			Modifiers:   mapKeyMod(sdl.GetModState()),
		}, true

	case *sdl.MouseMotionEvent:
		return gui.Event{
			Type:      gui.EventMouseMove,
			MouseX:    float32(e.X),
			MouseY:    float32(e.Y),
			MouseDX:   float32(e.XRel),
			MouseDY:   float32(e.YRel),
			Modifiers: mapKeyMod(sdl.GetModState()),
		}, true

	case *sdl.MouseWheelEvent:
		mx, my, _ := sdl.GetMouseState()
		return gui.Event{
			Type:      gui.EventMouseScroll,
			ScrollX:   float32(e.X),
			ScrollY:   float32(e.Y),
			MouseX:    float32(mx),
			MouseY:    float32(my),
			Modifiers: mapKeyMod(sdl.GetModState()),
		}, true

	case *sdl.KeyboardEvent:
		if e.Type == sdl.KEYDOWN {
			return gui.Event{
				Type:      gui.EventKeyDown,
				KeyCode:   mapKeyCode(e.Keysym.Sym),
				Modifiers: mapKeyMod(sdl.Keymod(e.Keysym.Mod)),
				KeyRepeat: e.Repeat > 0,
			}, true
		}
		return gui.Event{
			Type:      gui.EventKeyUp,
			KeyCode:   mapKeyCode(e.Keysym.Sym),
			Modifiers: mapKeyMod(sdl.Keymod(e.Keysym.Mod)),
		}, true

	case *sdl.TextInputEvent:
		text := e.GetText()
		if len(text) == 0 {
			return gui.Event{}, true
		}
		r, sz := utf8.DecodeRuneInString(text)
		if r == utf8.RuneError && sz == 1 {
			return gui.Event{}, true
		}
		return gui.Event{
			Type:      gui.EventChar,
			CharCode:  uint32(r),
			IMEText:   text,
			Modifiers: mapKeyMod(sdl.GetModState()),
		}, true

	case *sdl.TextEditingEvent:
		return gui.Event{
			Type:      gui.EventIMEComposition,
			IMEText:   e.GetText(),
			IMEStart:  e.Start,
			IMELength: e.Length,
		}, true

	case *sdl.WindowEvent:
		switch e.Event {
		case sdl.WINDOWEVENT_RESIZED,
			sdl.WINDOWEVENT_SIZE_CHANGED:
			return gui.Event{
				Type:         gui.EventResized,
				WindowWidth:  int(e.Data1),
				WindowHeight: int(e.Data2),
			}, true
		case sdl.WINDOWEVENT_FOCUS_GAINED:
			return gui.Event{Type: gui.EventFocused}, true
		case sdl.WINDOWEVENT_FOCUS_LOST:
			return gui.Event{Type: gui.EventUnfocused}, true
		}
	}
	return gui.Event{}, true
}

func mapMouseButton(b uint8) gui.MouseButton {
	switch b {
	case sdl.BUTTON_LEFT:
		return gui.MouseLeft
	case sdl.BUTTON_RIGHT:
		return gui.MouseRight
	case sdl.BUTTON_MIDDLE:
		return gui.MouseMiddle
	default:
		return gui.MouseInvalid
	}
}

func mapKeyMod(mod sdl.Keymod) gui.Modifier {
	var m gui.Modifier
	if mod&sdl.KMOD_SHIFT != 0 {
		m |= gui.ModShift
	}
	if mod&sdl.KMOD_CTRL != 0 {
		m |= gui.ModCtrl
	}
	if mod&sdl.KMOD_ALT != 0 {
		m |= gui.ModAlt
	}
	if mod&sdl.KMOD_GUI != 0 {
		m |= gui.ModSuper
	}
	return m
}

// mapKeyCode maps SDL keycodes to GLFW-style gui.KeyCode values.
func mapKeyCode(sym sdl.Keycode) gui.KeyCode {
	// Printable ASCII range (a-z → A-Z for GLFW compat).
	if sym >= 'a' && sym <= 'z' {
		return gui.KeyCode(sym - 32)
	}
	if sym >= '0' && sym <= '9' {
		return gui.KeyCode(sym)
	}

	switch sym {
	case sdl.K_SPACE:
		return gui.KeySpace
	case sdl.K_RETURN, sdl.K_RETURN2, sdl.K_KP_ENTER:
		return gui.KeyEnter
	case sdl.K_ESCAPE:
		return gui.KeyEscape
	case sdl.K_TAB:
		return gui.KeyTab
	case sdl.K_BACKSPACE:
		return gui.KeyBackspace
	case sdl.K_DELETE:
		return gui.KeyDelete
	case sdl.K_INSERT:
		return gui.KeyInsert
	case sdl.K_RIGHT:
		return gui.KeyRight
	case sdl.K_LEFT:
		return gui.KeyLeft
	case sdl.K_DOWN:
		return gui.KeyDown
	case sdl.K_UP:
		return gui.KeyUp
	case sdl.K_PAGEUP:
		return gui.KeyPageUp
	case sdl.K_PAGEDOWN:
		return gui.KeyPageDown
	case sdl.K_HOME:
		return gui.KeyHome
	case sdl.K_END:
		return gui.KeyEnd
	case sdl.K_LSHIFT:
		return gui.KeyLeftShift
	case sdl.K_RSHIFT:
		return gui.KeyRightShift
	case sdl.K_LCTRL:
		return gui.KeyLeftControl
	case sdl.K_RCTRL:
		return gui.KeyRightControl
	case sdl.K_LALT:
		return gui.KeyLeftAlt
	case sdl.K_RALT:
		return gui.KeyRightAlt
	case sdl.K_LGUI:
		return gui.KeyLeftSuper
	case sdl.K_RGUI:
		return gui.KeyRightSuper
	case sdl.K_COMMA:
		return gui.KeyComma
	case sdl.K_MINUS:
		return gui.KeyMinus
	case sdl.K_PERIOD:
		return gui.KeyPeriod
	case sdl.K_SLASH:
		return gui.KeySlash
	case sdl.K_SEMICOLON:
		return gui.KeySemicolon
	case sdl.K_EQUALS:
		return gui.KeyEqual
	case sdl.K_LEFTBRACKET:
		return gui.KeyLeftBracket
	case sdl.K_BACKSLASH:
		return gui.KeyBackslash
	case sdl.K_RIGHTBRACKET:
		return gui.KeyRightBracket
	case sdl.K_BACKQUOTE:
		return gui.KeyGraveAccent
	case sdl.K_CAPSLOCK:
		return gui.KeyCapsLock
	case sdl.K_F1:
		return gui.KeyF1
	case sdl.K_F2:
		return gui.KeyF2
	case sdl.K_F3:
		return gui.KeyF3
	case sdl.K_F4:
		return gui.KeyF4
	case sdl.K_F5:
		return gui.KeyF5
	case sdl.K_F6:
		return gui.KeyF6
	case sdl.K_F7:
		return gui.KeyF7
	case sdl.K_F8:
		return gui.KeyF8
	case sdl.K_F9:
		return gui.KeyF9
	case sdl.K_F10:
		return gui.KeyF10
	case sdl.K_F11:
		return gui.KeyF11
	case sdl.K_F12:
		return gui.KeyF12
	default:
		return gui.KeyInvalid
	}
}
