package gui

import "testing"

func TestInputGeneratesLayout(t *testing.T) {
	w := newTestWindow()
	v := Input(InputCfg{
		ID:          "email",
		Text:        "hello",
		Placeholder: "Enter email",
		IDFocus:     10,
	})
	layout := GenerateViewLayout(v, w)
	if layout.Shape.ID != "email" {
		t.Fatalf("got ID %q, want email", layout.Shape.ID)
	}
	if layout.Shape.A11YRole != AccessRoleTextField {
		t.Fatalf("got role %d, want TextField", layout.Shape.A11YRole)
	}
}

func TestInputMultilineRole(t *testing.T) {
	w := newTestWindow()
	v := Input(InputCfg{
		Mode:    InputMultiline,
		IDFocus: 11,
	})
	layout := GenerateViewLayout(v, w)
	if layout.Shape.A11YRole != AccessRoleTextArea {
		t.Fatalf("got role %d, want TextArea", layout.Shape.A11YRole)
	}
}

func TestInputReadOnlyWithoutFocus(t *testing.T) {
	w := newTestWindow()
	v := Input(InputCfg{Text: "readonly"})
	layout := GenerateViewLayout(v, w)
	if layout.Shape.A11YState != AccessStateReadOnly {
		t.Fatalf("got state %d, want ReadOnly", layout.Shape.A11YState)
	}
}

func TestInputPlaceholderWhenEmpty(t *testing.T) {
	w := newTestWindow()
	v := Input(InputCfg{
		Placeholder: "Type here",
		IDFocus:     12,
	})
	layout := GenerateViewLayout(v, w)
	// The inner Row → Text child should use placeholder text.
	if len(layout.Children) == 0 {
		t.Fatal("no children")
	}
	row := layout.Children[0]
	if len(row.Children) == 0 {
		t.Fatal("no row children")
	}
	txt := row.Children[0]
	if txt.Shape.TC == nil {
		t.Fatal("text config is nil")
	}
	if txt.Shape.TC.Text != "Type here" {
		t.Fatalf("got %q, want placeholder", txt.Shape.TC.Text)
	}
}

func TestInputPasswordMask(t *testing.T) {
	got := passwordMask("abc")
	if got != "•••" {
		t.Fatalf("got %q, want •••", got)
	}
}

func TestInputPasswordMaskEmoji(t *testing.T) {
	got := passwordMask("🔑key")
	want := "••••"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestInputDefaults(t *testing.T) {
	cfg := InputCfg{}
	applyInputDefaults(&cfg)
	if cfg.Color == (Color{}) {
		t.Fatal("Color not defaulted")
	}
	if !cfg.Radius.IsSet() {
		t.Fatal("Radius not defaulted")
	}
	if !cfg.SizeBorder.IsSet() {
		t.Fatal("SizeBorder not defaulted")
	}
	if cfg.TextStyle == (TextStyle{}) {
		t.Fatal("TextStyle not defaulted")
	}
	if cfg.PlaceholderStyle == (TextStyle{}) {
		t.Fatal("PlaceholderStyle not defaulted")
	}
}

func TestInputA11YLabelFallback(t *testing.T) {
	w := newTestWindow()
	v := Input(InputCfg{
		Placeholder: "Search...",
		IDFocus:     13,
	})
	layout := GenerateViewLayout(v, w)
	if layout.Shape.A11Y == nil {
		t.Fatal("A11Y nil")
	}
	if layout.Shape.A11Y.Label != "Search..." {
		t.Fatalf("got %q, want Search...", layout.Shape.A11Y.Label)
	}
}
