package gui

import (
	"testing"
	"time"
)

func TestToastAddCreatesNotification(t *testing.T) {
	w := &Window{}
	id := w.Toast(ToastCfg{Title: "Test", Body: "Body"})
	if id == 0 {
		t.Error("expected non-zero toast id")
	}
	if len(w.toasts) != 1 {
		t.Fatalf("expected 1 toast, got %d", len(w.toasts))
	}
	if w.toasts[0].phase != toastEntering {
		t.Error("expected entering phase")
	}
}

func TestToastCounterIncrements(t *testing.T) {
	w := &Window{}
	id1 := w.Toast(ToastCfg{Title: "A"})
	id2 := w.Toast(ToastCfg{Title: "B"})
	if id2 <= id1 {
		t.Errorf("expected id2 > id1, got %d <= %d", id2, id1)
	}
}

func TestToastRemove(t *testing.T) {
	w := &Window{}
	w.toasts = []toastNotification{
		{id: 1, cfg: ToastCfg{Title: "A"}, animFrac: 1},
		{id: 2, cfg: ToastCfg{Title: "B"}, animFrac: 1},
		{id: 3, cfg: ToastCfg{Title: "C"}, animFrac: 1},
	}
	toastRemove(w, 2)
	if len(w.toasts) != 2 {
		t.Fatalf("expected 2 toasts, got %d", len(w.toasts))
	}
	for _, toast := range w.toasts {
		if toast.id == 2 {
			t.Error("toast 2 should have been removed")
		}
	}
}

func TestToastSetHovered(t *testing.T) {
	w := &Window{}
	w.toasts = []toastNotification{
		{id: 1, cfg: ToastCfg{Title: "A"}},
		{id: 2, cfg: ToastCfg{Title: "B"}},
	}
	toastSetHovered(w, 2, true)
	if !w.toasts[1].hovered {
		t.Error("expected toast 2 to be hovered")
	}
	toastSetHovered(w, 2, false)
	if w.toasts[1].hovered {
		t.Error("expected toast 2 not hovered")
	}
}

func TestToastEnforceMaxVisible(t *testing.T) {
	w := &Window{}
	// Add more than max.
	for i := 0; i < 8; i++ {
		w.toasts = append(w.toasts, toastNotification{
			id:       uint64(i + 1),
			cfg:      ToastCfg{Title: "T"},
			animFrac: 1,
			phase:    toastVisible,
		})
	}
	toastEnforceMaxVisible(w)

	exiting := 0
	for _, toast := range w.toasts {
		if toast.phase == toastExiting {
			exiting++
		}
	}
	if exiting < 3 {
		t.Errorf("expected >= 3 exiting, got %d", exiting)
	}
}

func TestToastContainerViewNilWhenEmpty(t *testing.T) {
	w := &Window{}
	v := toastContainerView(w)
	if v != nil {
		t.Error("expected nil when no toasts")
	}
}

func TestToastContainerViewReturnsView(t *testing.T) {
	w := &Window{}
	w.toasts = []toastNotification{
		{id: 1, cfg: ToastCfg{Title: "Hi"}, animFrac: 1},
	}
	v := toastContainerView(w)
	if v == nil {
		t.Fatal("expected non-nil view")
	}
}

func TestToastDuration(t *testing.T) {
	w := &Window{}
	w.toasts = []toastNotification{
		{id: 1, cfg: ToastCfg{Duration: 5 * time.Second}},
	}
	d := toastDuration(w, 1)
	if d != 5*time.Second {
		t.Errorf("expected 5s, got %v", d)
	}
}

func TestToastDurationDefault(t *testing.T) {
	w := &Window{}
	w.toasts = []toastNotification{
		{id: 1, cfg: ToastCfg{}},
	}
	d := toastDuration(w, 1)
	if d != toastDefaultDelay {
		t.Errorf("expected %v, got %v", toastDefaultDelay, d)
	}
}

func TestToastStartExitGuardsDouble(t *testing.T) {
	w := &Window{}
	w.toasts = []toastNotification{
		{id: 1, cfg: ToastCfg{Title: "T"}, animFrac: 1,
			phase: toastExiting},
	}
	// Should not panic or add another animation.
	toastStartExit(w, 1)
	if w.toasts[0].phase != toastExiting {
		t.Error("expected phase still exiting")
	}
}

func TestToastDismissAll(t *testing.T) {
	w := &Window{}
	w.toasts = []toastNotification{
		{id: 1, cfg: ToastCfg{Title: "A"}, animFrac: 1,
			phase: toastVisible},
		{id: 2, cfg: ToastCfg{Title: "B"}, animFrac: 1,
			phase: toastVisible},
	}
	w.ToastDismissAll()
	for _, toast := range w.toasts {
		if toast.phase != toastExiting {
			t.Errorf("toast %d should be exiting", toast.id)
		}
	}
}

func TestUitoa(t *testing.T) {
	tests := []struct {
		n    uint64
		want string
	}{
		{0, "0"},
		{1, "1"},
		{42, "42"},
		{1234567890, "1234567890"},
	}
	for _, tt := range tests {
		got := uitoa(tt.n)
		if got != tt.want {
			t.Errorf("uitoa(%d) = %q, want %q",
				tt.n, got, tt.want)
		}
	}
}
