package gui

import "testing"

func TestDialogCfgDefaults(t *testing.T) {
	cfg := DialogCfg{}
	applyDialogDefaults(&cfg)

	if !cfg.Color.IsSet() {
		t.Error("expected non-zero Color")
	}
	if cfg.IDFocus != dialogBaseIDFocus {
		t.Errorf("expected IDFocus=%d, got %d",
			dialogBaseIDFocus, cfg.IDFocus)
	}
	if cfg.MinWidth != 200 {
		t.Errorf("expected MinWidth=200, got %f", cfg.MinWidth)
	}
	if cfg.MaxWidth != 300 {
		t.Errorf("expected MaxWidth=300, got %f", cfg.MaxWidth)
	}
}

func TestDialogViewGeneratorReturnsView(t *testing.T) {
	cfg := DialogCfg{
		Title:      "Test Dialog",
		Body:       "Some body text",
		DialogType: DialogMessage,
	}
	v := dialogViewGenerator(cfg)
	if v == nil {
		t.Fatal("expected non-nil view")
	}
	w := &Window{}
	layout := GenerateViewLayout(v, w)
	if layout.Shape == nil {
		t.Fatal("expected non-nil shape")
	}
	if layout.Shape.ID != reservedDialogID {
		t.Errorf("expected ID=%q, got %q",
			reservedDialogID, layout.Shape.ID)
	}
}

func TestDialogShowDismissLifecycle(t *testing.T) {
	w := &Window{}
	w.SetIDFocus(42)

	w.Dialog(DialogCfg{
		Title:      "Test",
		DialogType: DialogMessage,
	})
	if !w.DialogIsVisible() {
		t.Error("expected dialog visible after Dialog()")
	}
	if w.IDFocus() != dialogBaseIDFocus {
		t.Errorf("expected focus=%d, got %d",
			dialogBaseIDFocus, w.IDFocus())
	}

	w.DialogDismiss()
	if w.DialogIsVisible() {
		t.Error("expected dialog hidden after Dismiss()")
	}
	if w.IDFocus() != 42 {
		t.Errorf("expected focus restored to 42, got %d",
			w.IDFocus())
	}
}

func TestDialogKeyDownEscape(t *testing.T) {
	w := &Window{}
	cancelled := false
	cfg := DialogCfg{
		OnCancelNo: func(_ *Window) { cancelled = true },
	}
	handler := dialogKeyDown(cfg)
	e := &Event{KeyCode: KeyEscape}
	handler(nil, e, w)

	if !cancelled {
		t.Error("expected OnCancelNo to fire")
	}
	if !e.IsHandled {
		t.Error("expected event handled")
	}
}

func TestDialogKeyDownCtrlCCopiesBody(t *testing.T) {
	w := newTestWindow()
	var clipped string
	w.SetClipboardFn(func(s string) { clipped = s })

	cfg := DialogCfg{Body: "hello world"}
	handler := dialogKeyDown(cfg)
	e := &Event{KeyCode: KeyC, Modifiers: ModCtrl}
	handler(nil, e, w)

	if clipped != "hello world" {
		t.Fatalf("expected clipboard=%q got %q",
			"hello world", clipped)
	}
	if !e.IsHandled {
		t.Fatal("expected IsHandled=true")
	}
}

func TestDialogKeyDownSuperCCopiesBody(t *testing.T) {
	w := newTestWindow()
	var clipped string
	w.SetClipboardFn(func(s string) { clipped = s })

	cfg := DialogCfg{Body: "mac copy"}
	handler := dialogKeyDown(cfg)
	e := &Event{KeyCode: KeyC, Modifiers: ModSuper}
	handler(nil, e, w)

	if clipped != "mac copy" {
		t.Fatalf("expected clipboard=%q got %q",
			"mac copy", clipped)
	}
}

func TestDialogKeyDownCtrlCNoOpWhenBodyEmpty(t *testing.T) {
	w := newTestWindow()
	called := false
	w.SetClipboardFn(func(string) { called = true })

	handler := dialogKeyDown(DialogCfg{})
	e := &Event{KeyCode: KeyC, Modifiers: ModCtrl}
	handler(nil, e, w)

	if called {
		t.Fatal("clipboard should not be set when body empty")
	}
	if e.IsHandled {
		t.Fatal("expected IsHandled=false for empty body")
	}
}

func TestDialogConfirmView(t *testing.T) {
	cfg := DialogCfg{
		Title:      "Confirm?",
		Body:       "Are you sure?",
		DialogType: DialogConfirm,
	}
	v := dialogViewGenerator(cfg)
	if v == nil {
		t.Fatal("expected non-nil view")
	}
	w := &Window{}
	layout := GenerateViewLayout(v, w)
	if len(layout.Children) == 0 {
		t.Error("expected children for confirm dialog")
	}
}
