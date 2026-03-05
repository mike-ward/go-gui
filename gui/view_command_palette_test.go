package gui

import "testing"

func TestCommandPaletteHidden(t *testing.T) {
	w := &Window{}
	v := CommandPalette(CommandPaletteCfg{
		Items: []CommandPaletteItem{
			{ID: "save", Label: "Save"},
		},
		OnAction: func(_ string, _ *Event, _ *Window) {},
		IDFocus:  1,
	})
	layout := GenerateViewLayout(v, w)
	// Hidden state: should produce minimal layout.
	if layout.Shape.ShapeType == ShapeDrawCanvas {
		t.Error("hidden palette should not be canvas")
	}
}

func TestCommandPaletteVisible(t *testing.T) {
	w := &Window{}
	CommandPaletteShow("__cmd_palette__", 1, w)

	v := CommandPalette(CommandPaletteCfg{
		Items: []CommandPaletteItem{
			{ID: "save", Label: "Save", Detail: "Ctrl+S"},
			{ID: "open", Label: "Open", Detail: "Ctrl+O"},
		},
		OnAction: func(_ string, _ *Event, _ *Window) {},
		IDFocus:  1,
	})
	layout := GenerateViewLayout(v, w)
	// Should produce a non-trivial layout tree.
	if len(layout.Children) == 0 {
		t.Error("visible palette should have children")
	}
}

func TestCommandPaletteShowDismiss(t *testing.T) {
	w := &Window{}
	id := "cp-sd"
	CommandPaletteShow(id, 1, w)
	if !CommandPaletteIsVisible(w, id) {
		t.Error("expected visible after show")
	}
	CommandPaletteDismiss(id, w)
	if CommandPaletteIsVisible(w, id) {
		t.Error("expected hidden after dismiss")
	}
}

func TestCommandPaletteToggle(t *testing.T) {
	w := &Window{}
	id := "cp-tog"
	CommandPaletteToggle(id, 1, w)
	if !CommandPaletteIsVisible(w, id) {
		t.Error("first toggle should show")
	}
	CommandPaletteToggle(id, 1, w)
	if CommandPaletteIsVisible(w, id) {
		t.Error("second toggle should hide")
	}
}

func TestPaletteOnKeyDownEscape(t *testing.T) {
	w := &Window{}
	CommandPaletteShow("cp-esc", 1, w)
	dismissed := false
	e := &Event{KeyCode: KeyEscape}
	paletteOnKeyDown("cp-esc", nil,
		func(_ *Window) { dismissed = true },
		[]string{"a"}, e, w)
	if !dismissed {
		t.Error("escape should dismiss")
	}
	if CommandPaletteIsVisible(w, "cp-esc") {
		t.Error("should be hidden after escape")
	}
}

func TestPaletteOnKeyDownSelect(t *testing.T) {
	w := &Window{}
	CommandPaletteShow("cp-sel", 1, w)
	selected := ""
	e := &Event{KeyCode: KeyEnter}
	paletteOnKeyDown("cp-sel",
		func(id string, _ *Event, _ *Window) { selected = id },
		nil,
		[]string{"cmd1", "cmd2"}, e, w)
	if selected != "cmd1" {
		t.Errorf("selected = %q, want cmd1", selected)
	}
}

func TestCommandPaletteItemsCacheInvalidatesOnItemsChange(t *testing.T) {
	w := &Window{}
	id := "cp-cache"
	CommandPaletteShow(id, 1, w)

	v := CommandPalette(CommandPaletteCfg{
		ID: id,
		Items: []CommandPaletteItem{
			{ID: "a", Label: "A"},
		},
		OnAction: func(_ string, _ *Event, _ *Window) {},
		IDFocus:  1,
	})
	_ = GenerateViewLayout(v, w)

	cm := StateMapRead[string, *cmdPaletteItemsCache](w, nsCmdPaletteItems)
	if cm == nil {
		t.Fatal("expected command palette items cache map")
	}
	cache, ok := cm.Get(id)
	if !ok || cache == nil {
		t.Fatal("expected command palette cache entry")
	}
	if got := len(cache.items); got != 1 {
		t.Fatalf("cache items len = %d, want 1", got)
	}

	v = CommandPalette(CommandPaletteCfg{
		ID: id,
		Items: []CommandPaletteItem{
			{ID: "a", Label: "A"},
			{ID: "b", Label: "B"},
		},
		OnAction: func(_ string, _ *Event, _ *Window) {},
		IDFocus:  1,
	})
	_ = GenerateViewLayout(v, w)
	cache, _ = cm.Get(id)
	if got := len(cache.items); got != 2 {
		t.Fatalf("cache items len = %d, want 2", got)
	}
}
