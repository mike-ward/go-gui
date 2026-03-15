package gui

import (
	"runtime"
	"testing"
)

func TestShortcutString(t *testing.T) {
	s := Shortcut{Key: KeyS, Modifiers: ModCtrl}
	got := s.String()
	if runtime.GOOS == "darwin" {
		want := "⌃S"
		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	} else {
		want := "Ctrl+S"
		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	}
}

func TestShortcutStringMultiModifier(t *testing.T) {
	s := Shortcut{Key: KeyZ, Modifiers: ModCtrlShift}
	got := s.String()
	if runtime.GOOS == "darwin" {
		want := "⌃⇧Z"
		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	} else {
		want := "Ctrl+Shift+Z"
		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	}
}

func TestShortcutStringNoKey(t *testing.T) {
	s := Shortcut{}
	if got := s.String(); got != "" {
		t.Errorf("got %q, want empty", got)
	}
}

func TestShortcutIsSet(t *testing.T) {
	if (Shortcut{}).IsSet() {
		t.Error("zero shortcut should not be set")
	}
	if !(Shortcut{Key: KeyA}).IsSet() {
		t.Error("shortcut with key should be set")
	}
}

func TestShortcutMatches(t *testing.T) {
	s := Shortcut{Key: KeyS, Modifiers: ModCtrl}
	hit := &Event{KeyCode: KeyS, Modifiers: ModCtrl}
	miss := &Event{KeyCode: KeyS, Modifiers: ModNone}
	if !s.matches(hit) {
		t.Error("should match")
	}
	if s.matches(miss) {
		t.Error("should not match without modifier")
	}
}

func TestRegisterCommand(t *testing.T) {
	w := NewWindow(WindowCfg{State: new(int)})
	w.RegisterCommand(Command{ID: "test", Label: "Test"})
	cmd, ok := w.CommandByID("test")
	if !ok || cmd.Label != "Test" {
		t.Error("command not found")
	}
}

func TestRegisterCommandDuplicatePanics(t *testing.T) {
	w := NewWindow(WindowCfg{State: new(int)})
	w.RegisterCommand(Command{ID: "x"})
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic on duplicate ID")
		}
	}()
	w.RegisterCommand(Command{ID: "x"})
}

func TestRegisterCommandDuplicateShortcutPanics(t *testing.T) {
	w := NewWindow(WindowCfg{State: new(int)})
	s := Shortcut{Key: KeyS, Modifiers: ModCtrl}
	w.RegisterCommand(Command{ID: "a", Shortcut: s})
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic on duplicate shortcut")
		}
	}()
	w.RegisterCommand(Command{ID: "b", Shortcut: s})
}

func TestUnregisterCommand(t *testing.T) {
	w := NewWindow(WindowCfg{State: new(int)})
	w.RegisterCommand(Command{ID: "rm"})
	w.UnregisterCommand("rm")
	if _, ok := w.CommandByID("rm"); ok {
		t.Error("command should be removed")
	}
}

func TestCommandCanExecute(t *testing.T) {
	w := NewWindow(WindowCfg{State: new(int)})
	w.RegisterCommand(Command{
		ID:         "c",
		CanExecute: func(_ *Window) bool { return false },
	})
	if w.CommandCanExecute("c") {
		t.Error("should be disabled")
	}
	if w.CommandCanExecute("nonexistent") {
		t.Error("nonexistent should return false")
	}
}

func TestCommandCanExecuteNil(t *testing.T) {
	w := NewWindow(WindowCfg{State: new(int)})
	w.RegisterCommand(Command{ID: "d"})
	if !w.CommandCanExecute("d") {
		t.Error("nil CanExecute = always enabled")
	}
}

func TestCommandDispatchGlobal(t *testing.T) {
	w := NewWindow(WindowCfg{State: new(int)})
	called := false
	w.RegisterCommand(Command{
		ID:       "g",
		Shortcut: Shortcut{Key: KeyN, Modifiers: ModCtrl},
		Global:   true,
		Execute:  func(_ *Event, _ *Window) { called = true },
	})
	e := &Event{KeyCode: KeyN, Modifiers: ModCtrl}
	w.commandDispatch(e, true)
	if !called {
		t.Error("global command not dispatched")
	}
	if !e.IsHandled {
		t.Error("event should be handled")
	}
}

func TestCommandDispatchNonGlobal(t *testing.T) {
	w := NewWindow(WindowCfg{State: new(int)})
	called := false
	w.RegisterCommand(Command{
		ID:       "ng",
		Shortcut: Shortcut{Key: KeyZ, Modifiers: ModCtrl},
		Execute:  func(_ *Event, _ *Window) { called = true },
	})
	e := &Event{KeyCode: KeyZ, Modifiers: ModCtrl}
	// Should not fire when checking global only.
	w.commandDispatch(e, true)
	if called {
		t.Error("non-global should not fire in global pass")
	}
	// Should fire in non-global pass.
	w.commandDispatch(e, false)
	if !called {
		t.Error("non-global command not dispatched")
	}
}

func TestCommandDispatchCanExecuteFalse(t *testing.T) {
	w := NewWindow(WindowCfg{State: new(int)})
	called := false
	w.RegisterCommand(Command{
		ID:         "ce",
		Shortcut:   Shortcut{Key: KeyS, Modifiers: ModCtrl},
		Execute:    func(_ *Event, _ *Window) { called = true },
		CanExecute: func(_ *Window) bool { return false },
	})
	e := &Event{KeyCode: KeyS, Modifiers: ModCtrl}
	w.commandDispatch(e, false)
	if called {
		t.Error("should not execute when CanExecute=false")
	}
}

func TestCommandDispatchNoShortcut(t *testing.T) {
	w := NewWindow(WindowCfg{State: new(int)})
	called := false
	w.RegisterCommand(Command{
		ID:      "ns",
		Execute: func(_ *Event, _ *Window) { called = true },
	})
	e := &Event{KeyCode: KeyA}
	w.commandDispatch(e, false)
	if called {
		t.Error("command with no shortcut should not dispatch")
	}
}

func TestCommandPaletteItems(t *testing.T) {
	w := NewWindow(WindowCfg{State: new(int)})
	w.RegisterCommands(
		Command{ID: "a", Label: "Alpha", Group: "G1",
			Shortcut: Shortcut{Key: KeyA, Modifiers: ModCtrl}},
		Command{ID: "b", Label: "Beta"},
		Command{ID: "c"}, // no label, excluded
	)
	items := w.CommandPaletteItems()
	if len(items) != 2 {
		t.Fatalf("got %d items, want 2", len(items))
	}
	if items[0].ID != "a" || items[0].Group != "G1" {
		t.Error("first item mismatch")
	}
	if items[0].Detail == "" {
		t.Error("expected shortcut detail text")
	}
}

func TestKeyNameLetters(t *testing.T) {
	if got := keyName(KeyA); got != "A" {
		t.Errorf("got %q, want A", got)
	}
	if got := keyName(KeyZ); got != "Z" {
		t.Errorf("got %q, want Z", got)
	}
}

func TestKeyNameNumbers(t *testing.T) {
	if got := keyName(Key0); got != "0" {
		t.Errorf("got %q, want 0", got)
	}
	if got := keyName(Key9); got != "9" {
		t.Errorf("got %q, want 9", got)
	}
}

func TestKeyNameFunctionKeys(t *testing.T) {
	if got := keyName(KeyF1); got != "F1" {
		t.Errorf("got %q, want F1", got)
	}
	if got := keyName(KeyF12); got != "F12" {
		t.Errorf("got %q, want F12", got)
	}
}

func TestKeyNameSpecial(t *testing.T) {
	tests := []struct {
		key  KeyCode
		want string
	}{
		{KeySpace, "Space"},
		{KeyEnter, "Enter"},
		{KeyEscape, "Esc"},
		{KeyDelete, "Del"},
		{KeyTab, "Tab"},
		{KeyHome, "Home"},
		{KeyEnd, "End"},
	}
	for _, tt := range tests {
		if got := keyName(tt.key); got != tt.want {
			t.Errorf("keyName(%d) = %q, want %q",
				tt.key, got, tt.want)
		}
	}
}
