package gui

import "testing"

func TestComboboxClosedLayout(t *testing.T) {
	w := &Window{}
	v := Combobox(ComboboxCfg{
		ID:          "cb1",
		Value:       "Apple",
		Options:     []string{"Apple", "Banana", "Cherry"},
		Placeholder: "Pick fruit",
		OnSelect:    func(_ string, _ *Event, _ *Window) {},
	})
	layout := GenerateViewLayout(v, w)
	if layout.Shape.ID != "cb1" {
		t.Errorf("ID = %q", layout.Shape.ID)
	}
	if layout.Shape.ShapeType != ShapeRectangle {
		t.Errorf("type = %d", layout.Shape.ShapeType)
	}
}

func TestComboboxOpenLayout(t *testing.T) {
	w := &Window{}
	// Pre-set open state.
	ss := StateMap[string, bool](w, nsCombobox, capModerate)
	ss.Set("cb-open", true)

	v := Combobox(ComboboxCfg{
		ID:       "cb-open",
		Options:  []string{"A", "B", "C"},
		OnSelect: func(_ string, _ *Event, _ *Window) {},
	})
	layout := GenerateViewLayout(v, w)
	// Should have children (input, spacer, arrow, dropdown).
	if len(layout.Children) < 3 {
		t.Errorf("children = %d, want >= 3", len(layout.Children))
	}
}

func TestComboboxOpenClose(t *testing.T) {
	w := &Window{}
	comboboxOpen("test-oc", 0, w)
	isOpen := StateReadOr[string, bool](w, nsCombobox, "test-oc", false)
	if !isOpen {
		t.Error("expected open")
	}
	comboboxClose("test-oc", w)
	isOpen = StateReadOr[string, bool](w, nsCombobox, "test-oc", false)
	if isOpen {
		t.Error("expected closed")
	}
}

func TestComboboxKeyDownOpenClose(t *testing.T) {
	w := &Window{}
	called := ""
	onSel := func(id string, _ *Event, _ *Window) { called = id }

	// Open via Enter.
	e := &Event{KeyCode: KeyEnter}
	comboboxOnKeyDown("cb-kd", onSel, 0, []string{"x", "y"}, e, w)
	if !StateReadOr[string, bool](w, nsCombobox, "cb-kd", false) {
		t.Error("enter should open")
	}

	// Select via Enter.
	e = &Event{KeyCode: KeyEnter}
	comboboxOnKeyDown("cb-kd", onSel, 0, []string{"x", "y"}, e, w)
	if called != "x" {
		t.Errorf("selected = %q, want x", called)
	}
}

func TestComboboxDefaults(t *testing.T) {
	cfg := ComboboxCfg{}
	applyComboboxDefaults(&cfg)
	if cfg.MaxDropdownHeight != 200 {
		t.Errorf("max dropdown = %f", cfg.MaxDropdownHeight)
	}
	if cfg.MinWidth != 75 {
		t.Errorf("min width = %f", cfg.MinWidth)
	}
}

func TestComboboxTextChanged(t *testing.T) {
	w := &Window{}
	fn := makeComboboxOnTextChanged("cb-tc")
	fn(nil, "hello", w)
	q := StateReadOr[string, string](w, nsComboboxQuery, "cb-tc", "")
	if q != "hello" {
		t.Errorf("query = %q", q)
	}
}
