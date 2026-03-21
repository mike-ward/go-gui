package gui

import "testing"

func TestNativeMenuItemsFromMenuItems_Basic(t *testing.T) {
	items := []MenuItemCfg{
		{ID: "a", Text: "Alpha", CommandID: "cmd-a"},
		{ID: "b", Text: "Beta"},
	}
	got := NativeMenuItemsFromMenuItems(items)
	if len(got) != 2 {
		t.Fatalf("len = %d, want 2", len(got))
	}
	if got[0].ID != "a" || got[0].Text != "Alpha" {
		t.Errorf("item 0: got %+v", got[0])
	}
	if got[0].CommandID != "cmd-a" {
		t.Errorf("CommandID = %q, want cmd-a", got[0].CommandID)
	}
	if got[1].ID != "b" || got[1].Text != "Beta" {
		t.Errorf("item 1: got %+v", got[1])
	}
}

func TestNativeMenuItemsFromMenuItems_Separator(t *testing.T) {
	items := []MenuItemCfg{
		MenuSeparator(),
	}
	got := NativeMenuItemsFromMenuItems(items)
	if len(got) != 1 {
		t.Fatalf("len = %d, want 1", len(got))
	}
	if !got[0].Separator {
		t.Error("expected Separator=true")
	}
}

func TestNativeMenuItemsFromMenuItems_Submenu(t *testing.T) {
	items := []MenuItemCfg{
		MenuSubmenu("file", "File", []MenuItemCfg{
			{ID: "new", Text: "New"},
			{ID: "open", Text: "Open"},
		}),
	}
	got := NativeMenuItemsFromMenuItems(items)
	if len(got) != 1 {
		t.Fatalf("len = %d, want 1", len(got))
	}
	if len(got[0].Submenu) != 2 {
		t.Fatalf("submenu len = %d, want 2",
			len(got[0].Submenu))
	}
	if got[0].Submenu[0].ID != "new" {
		t.Errorf("submenu[0].ID = %q", got[0].Submenu[0].ID)
	}
}

func TestNativeMenuItemsFromMenuItems_Empty(t *testing.T) {
	got := NativeMenuItemsFromMenuItems(nil)
	if len(got) != 0 {
		t.Fatalf("len = %d, want 0", len(got))
	}
}

func TestNativeMenuItemsFromMenuItems_Disabled(t *testing.T) {
	items := []MenuItemCfg{
		MenuSubtitle("Section"),
	}
	got := NativeMenuItemsFromMenuItems(items)
	if !got[0].Disabled {
		t.Error("expected Disabled=true for subtitle")
	}
}

func TestExitOnTrayRemoved(t *testing.T) {
	app := NewApp()
	app.ExitMode = ExitOnTrayRemoved

	w := NewWindow(WindowCfg{State: new(struct{})})
	app.Register(1, w)

	// Add a tray.
	app.mu.Lock()
	app.trays = map[int]*SystemTrayHandle{1: {id: 1}}
	app.mu.Unlock()

	// Unregister last window — should NOT exit (tray exists).
	if app.Unregister(1) {
		t.Error("should not exit: tray still active")
	}

	// Remove tray, register+unregister — should exit.
	app.mu.Lock()
	delete(app.trays, 1)
	app.mu.Unlock()

	w2 := NewWindow(WindowCfg{State: new(struct{})})
	app.Register(2, w2)
	if !app.Unregister(2) {
		t.Error("should exit: no windows and no trays")
	}
}
