package gui

import "testing"

func noop(_ *Layout, _ *Event, _ *Window) {}

// --- Radio ---

func TestRadioGeneratesLayout(t *testing.T) {
	w := newTestWindow()
	v := Radio(RadioCfg{
		Label:   "Option A",
		OnClick: noop,
		IDFocus: 1,
	})
	layout := GenerateViewLayout(v, w)
	if layout.Shape.A11YRole != AccessRoleRadioButton {
		t.Fatalf("got role %d, want RadioButton", layout.Shape.A11YRole)
	}
	// Children: circle + label row
	if len(layout.Children) != 2 {
		t.Fatalf("got %d children, want 2", len(layout.Children))
	}
}

func TestRadioSelectedState(t *testing.T) {
	w := newTestWindow()
	v := Radio(RadioCfg{
		OnClick:  noop,
		Selected: true,
	})
	layout := GenerateViewLayout(v, w)
	if layout.Shape.A11YState != AccessStateSelected {
		t.Fatalf("got state %d, want Selected", layout.Shape.A11YState)
	}
}

func TestRadioNoLabel(t *testing.T) {
	w := newTestWindow()
	v := Radio(RadioCfg{OnClick: noop})
	layout := GenerateViewLayout(v, w)
	// Only circle child.
	if len(layout.Children) != 1 {
		t.Fatalf("got %d children, want 1", len(layout.Children))
	}
}

// --- Toggle / Checkbox ---

func TestToggleGeneratesLayout(t *testing.T) {
	w := newTestWindow()
	v := Toggle(ToggleCfg{
		Label:   "Accept",
		OnClick: noop,
		IDFocus: 2,
	})
	layout := GenerateViewLayout(v, w)
	if layout.Shape.A11YRole != AccessRoleCheckbox {
		t.Fatalf("got role %d, want Checkbox", layout.Shape.A11YRole)
	}
	// Children: toggle box + label text
	if len(layout.Children) != 2 {
		t.Fatalf("got %d children, want 2", len(layout.Children))
	}
}

func TestCheckboxIsToggleAlias(t *testing.T) {
	w := newTestWindow()
	v := Checkbox(ToggleCfg{OnClick: noop})
	layout := GenerateViewLayout(v, w)
	if layout.Shape.A11YRole != AccessRoleCheckbox {
		t.Fatalf("got role %d, want Checkbox", layout.Shape.A11YRole)
	}
}

func TestToggleCheckedState(t *testing.T) {
	w := newTestWindow()
	v := Toggle(ToggleCfg{OnClick: noop, Selected: true})
	layout := GenerateViewLayout(v, w)
	if layout.Shape.A11YState != AccessStateChecked {
		t.Fatalf("got state %d, want Checked", layout.Shape.A11YState)
	}
}

// --- Switch ---

func TestSwitchGeneratesLayout(t *testing.T) {
	w := newTestWindow()
	v := Switch(SwitchCfg{
		Label:   "Dark mode",
		OnClick: noop,
		IDFocus: 3,
	})
	layout := GenerateViewLayout(v, w)
	if layout.Shape.A11YRole != AccessRoleSwitchToggle {
		t.Fatalf("got role %d, want SwitchToggle", layout.Shape.A11YRole)
	}
	// Children: switch body + label text
	if len(layout.Children) != 2 {
		t.Fatalf("got %d children, want 2", len(layout.Children))
	}
}

func TestSwitchSelectedState(t *testing.T) {
	w := newTestWindow()
	v := Switch(SwitchCfg{OnClick: noop, Selected: true})
	layout := GenerateViewLayout(v, w)
	if layout.Shape.A11YState != AccessStateChecked {
		t.Fatalf("got state %d, want Checked", layout.Shape.A11YState)
	}
}

func TestSwitchNoLabel(t *testing.T) {
	w := newTestWindow()
	v := Switch(SwitchCfg{OnClick: noop})
	layout := GenerateViewLayout(v, w)
	// Only switch body child.
	if len(layout.Children) != 1 {
		t.Fatalf("got %d children, want 1", len(layout.Children))
	}
}

// --- Select ---

func TestSelectGeneratesLayout(t *testing.T) {
	w := newTestWindow()
	v := Select(SelectCfg{
		ID:       "country",
		Options:  []string{"US", "UK", "DE"},
		OnSelect: func(_ []string, _ *Event, _ *Window) {},
		IDFocus:  10,
	})
	layout := GenerateViewLayout(v, w)
	if layout.Shape.A11YRole != AccessRoleComboBox {
		t.Fatalf("got role %d, want ComboBox", layout.Shape.A11YRole)
	}
	// Content: text + spacer + arrow
	if len(layout.Children) < 3 {
		t.Fatalf("got %d children, want >= 3", len(layout.Children))
	}
}

func TestSelectPlaceholder(t *testing.T) {
	w := newTestWindow()
	v := Select(SelectCfg{
		ID:          "sel",
		Placeholder: "Choose...",
		Options:     []string{"A", "B"},
		OnSelect:    func(_ []string, _ *Event, _ *Window) {},
	})
	layout := GenerateViewLayout(v, w)
	if len(layout.Children) == 0 {
		t.Fatal("no children")
	}
	txt := layout.Children[0]
	if txt.Shape.TC == nil || txt.Shape.TC.Text != "Choose..." {
		t.Fatalf("placeholder not rendered")
	}
}

func TestSelectShowsSelected(t *testing.T) {
	w := newTestWindow()
	v := Select(SelectCfg{
		ID:       "sel",
		Selected: []string{"B"},
		Options:  []string{"A", "B", "C"},
		OnSelect: func(_ []string, _ *Event, _ *Window) {},
	})
	layout := GenerateViewLayout(v, w)
	if len(layout.Children) == 0 {
		t.Fatal("no children")
	}
	txt := layout.Children[0]
	if txt.Shape.TC == nil || txt.Shape.TC.Text != "B" {
		t.Fatalf("got %q, want B", txt.Shape.TC.Text)
	}
}

// --- NumericInput ---

func TestNumericInputGeneratesLayout(t *testing.T) {
	w := newTestWindow()
	v := NumericInput(NumericInputCfg{
		ID:      "qty",
		Text:    "42",
		IDFocus: 20,
	})
	layout := GenerateViewLayout(v, w)
	if layout.Shape.A11YRole != AccessRoleTextField {
		t.Fatalf("got role %d, want TextField", layout.Shape.A11YRole)
	}
}

func TestNumericInputNoButtons(t *testing.T) {
	w := newTestWindow()
	v := NumericInput(NumericInputCfg{
		ID:      "qty",
		Text:    "10",
		StepCfg: NumericStepCfg{ShowButtons: false},
		IDFocus: 21,
	})
	layout := GenerateViewLayout(v, w)
	// Should be a Column (Input view), not Row with step buttons.
	if layout.Shape.Axis != AxisTopToBottom {
		t.Fatalf("expected column axis, got %d", layout.Shape.Axis)
	}
}

func TestNumericInputWithButtons(t *testing.T) {
	w := newTestWindow()
	v := NumericInput(NumericInputCfg{
		ID:      "qty",
		Text:    "10",
		StepCfg: NumericStepCfg{ShowButtons: true, Step: 1},
		IDFocus: 22,
	})
	layout := GenerateViewLayout(v, w)
	// Row with field + step button column.
	if len(layout.Children) != 2 {
		t.Fatalf("got %d children, want 2", len(layout.Children))
	}
}

// --- ListBox ---

func TestListBoxGeneratesLayout(t *testing.T) {
	w := newTestWindow()
	v := ListBox(ListBoxCfg{
		ID: "fruits",
		Data: []ListBoxOption{
			{ID: "a", Name: "Apple"},
			{ID: "b", Name: "Banana"},
			{ID: "c", Name: "Cherry"},
		},
		OnSelect: func(_ []string, _ *Event, _ *Window) {},
		IDFocus:  30,
	})
	layout := GenerateViewLayout(v, w)
	if layout.Shape.A11YRole != AccessRoleList {
		t.Fatalf("got role %d, want List", layout.Shape.A11YRole)
	}
	if len(layout.Children) != 3 {
		t.Fatalf("got %d children, want 3", len(layout.Children))
	}
}

func TestListBoxItemRole(t *testing.T) {
	w := newTestWindow()
	v := ListBox(ListBoxCfg{
		ID: "lb",
		Data: []ListBoxOption{
			{ID: "x", Name: "Item X"},
		},
		OnSelect: func(_ []string, _ *Event, _ *Window) {},
	})
	layout := GenerateViewLayout(v, w)
	if len(layout.Children) == 0 {
		t.Fatal("no children")
	}
	item := layout.Children[0]
	if item.Shape.A11YRole != AccessRoleListItem {
		t.Fatalf("got role %d, want ListItem", item.Shape.A11YRole)
	}
}

func TestListBoxSelectedState(t *testing.T) {
	w := newTestWindow()
	v := ListBox(ListBoxCfg{
		ID: "lb",
		Data: []ListBoxOption{
			{ID: "x", Name: "Item X"},
			{ID: "y", Name: "Item Y"},
		},
		SelectedIDs: []string{"y"},
		OnSelect:    func(_ []string, _ *Event, _ *Window) {},
	})
	layout := GenerateViewLayout(v, w)
	if len(layout.Children) < 2 {
		t.Fatal("too few children")
	}
	if layout.Children[1].Shape.A11YState != AccessStateSelected {
		t.Fatalf("item y not selected")
	}
	if layout.Children[0].Shape.A11YState != AccessStateNone {
		t.Fatalf("item x should be unselected")
	}
}

func TestListBoxSubheading(t *testing.T) {
	w := newTestWindow()
	v := ListBox(ListBoxCfg{
		ID: "lb",
		Data: []ListBoxOption{
			NewListBoxSubheading("h1", "Heading"),
			NewListBoxOption("a", "Alpha", "val"),
		},
		OnSelect: func(_ []string, _ *Event, _ *Window) {},
	})
	layout := GenerateViewLayout(v, w)
	if len(layout.Children) != 2 {
		t.Fatalf("got %d children, want 2", len(layout.Children))
	}
}

// --- ListCore ---

func TestListCoreNavigate(t *testing.T) {
	if listCoreNavigate(KeyUp, 5) != ListCoreMoveUp {
		t.Fatal("expected MoveUp")
	}
	if listCoreNavigate(KeyDown, 5) != ListCoreMoveDown {
		t.Fatal("expected MoveDown")
	}
	if listCoreNavigate(KeyEnter, 5) != ListCoreSelectItem {
		t.Fatal("expected SelectItem")
	}
	if listCoreNavigate(KeyEscape, 5) != ListCoreDismiss {
		t.Fatal("expected Dismiss")
	}
	if listCoreNavigate(KeyUp, 0) != ListCoreNone {
		t.Fatal("expected None for empty list")
	}
}

func TestListCoreApplyNav(t *testing.T) {
	next, changed := listCoreApplyNav(ListCoreMoveDown, 0, 5)
	if !changed || next != 1 {
		t.Fatalf("got %d/%v, want 1/true", next, changed)
	}
	next, changed = listCoreApplyNav(ListCoreMoveUp, 0, 5)
	if changed || next != 0 {
		t.Fatalf("got %d/%v, want 0/false", next, changed)
	}
	next, changed = listCoreApplyNav(ListCoreLast, 0, 5)
	if !changed || next != 4 {
		t.Fatalf("got %d/%v, want 4/true", next, changed)
	}
	next, changed = listCoreApplyNav(ListCoreFirst, 3, 5)
	if !changed || next != 0 {
		t.Fatalf("got %d/%v, want 0/true", next, changed)
	}
}

func TestListCoreFuzzyScore(t *testing.T) {
	if listCoreFuzzyScore("Hello World", "hw") != 5 {
		t.Fatalf("got %d, want 5", listCoreFuzzyScore("Hello World", "hw"))
	}
	if listCoreFuzzyScore("abc", "xyz") != -1 {
		t.Fatal("expected no match")
	}
	if listCoreFuzzyScore("test", "") != 0 {
		t.Fatal("empty query should score 0")
	}
}

func TestListCoreVisibleRange(t *testing.T) {
	first, last := listCoreVisibleRange(100, 20, 200, 0)
	if first != 0 {
		t.Fatalf("first: got %d, want 0", first)
	}
	if last > 14 { // ~10 visible + 2 buffer + 1
		t.Fatalf("last: got %d, want <= 14", last)
	}
}

func TestListBoxNextSelectedIDs(t *testing.T) {
	// Single select.
	got := listBoxNextSelectedIDs([]string{"a"}, "b", false)
	if len(got) != 1 || got[0] != "b" {
		t.Fatalf("single: got %v", got)
	}
	// Multi add.
	got = listBoxNextSelectedIDs([]string{"a"}, "b", true)
	if len(got) != 2 {
		t.Fatalf("multi add: got %v", got)
	}
	// Multi remove.
	got = listBoxNextSelectedIDs([]string{"a", "b"}, "a", true)
	if len(got) != 1 || got[0] != "b" {
		t.Fatalf("multi remove: got %v", got)
	}
}

func TestListBoxOnKeyDownHandled(t *testing.T) {
	w := newTestWindow()
	itemIDs := []string{"a", "b"}
	e := &Event{KeyCode: KeyDown}
	listBoxOnKeyDown("lb", itemIDs, false,
		func(_ []string, _ *Event, _ *Window) {}, nil,
		0, 0, 0, nil, e, w)
	if !e.IsHandled {
		t.Fatal("expected key navigation event to be handled")
	}

	e = &Event{KeyCode: KeyEnter}
	called := false
	listBoxOnKeyDown("lb", itemIDs, false,
		func(_ []string, _ *Event, _ *Window) { called = true },
		nil, 0, 0, 0, nil, e, w)
	if !e.IsHandled {
		t.Fatal("expected key select event to be handled")
	}
	if !called {
		t.Fatal("expected select callback to run")
	}
}

func TestListBoxDataIndex(t *testing.T) {
	// Data: [sub, a, b, sub, c] → itemIDs=[a,b,c], indices=[1,2,4]
	indices := []int{1, 2, 4}
	if got := listBoxDataIndex(indices, 0); got != 1 {
		t.Errorf("idx 0 → %d, want 1", got)
	}
	if got := listBoxDataIndex(indices, 2); got != 4 {
		t.Errorf("idx 2 → %d, want 4", got)
	}
	// Out of range falls through.
	if got := listBoxDataIndex(indices, 5); got != 5 {
		t.Errorf("idx 5 → %d, want 5", got)
	}
	// Nil mapping returns idx unchanged.
	if got := listBoxDataIndex(nil, 3); got != 3 {
		t.Errorf("nil idx 3 → %d, want 3", got)
	}
}

func TestListBoxScrollWithSubheadings(t *testing.T) {
	w := &Window{}
	var idScroll uint32 = 90
	rowH := float32(26)
	listH := float32(187)

	// Data: [sub, a, b, c, d, e, f, g, sub, h, i, j]
	// itemIDs index 7 = "h" → data index 9 (after 2 subheadings).
	itemDataIndices := []int{1, 2, 3, 4, 5, 6, 7, 9, 10, 11}

	// Scroll to item at itemIDs index 7 using data index.
	dataIdx := listBoxDataIndex(itemDataIndices, 7)
	scrollEnsureVisible(idScroll, dataIdx, rowH, listH, w)

	sy := StateReadOr[uint32, float32](w, nsScrollY, idScroll, 0)
	// Data index 9: bottom = 10*26 = 260 > 187 → scroll = -(260-187) = -73
	want := -(float32(10)*rowH - listH)
	if sy != want {
		t.Fatalf("scrollY = %f, want %f", sy, want)
	}
}
