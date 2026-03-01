package gui

import "testing"

func TestProgressBarDefaultLayout(t *testing.T) {
	v := ProgressBar(ProgressBarCfg{
		Percent: 0.5,
	})
	layout := GenerateViewLayout(v, &Window{})
	// Row with fill bar child + text label child
	if layout.Shape.Axis != AxisLeftToRight {
		t.Error("default should be horizontal (row)")
	}
	if len(layout.Children) < 1 {
		t.Fatal("should have at least 1 child (fill bar)")
	}
}

func TestProgressBarVertical(t *testing.T) {
	v := ProgressBar(ProgressBarCfg{
		Percent:  0.5,
		Vertical: true,
	})
	layout := GenerateViewLayout(v, &Window{})
	if layout.Shape.Axis != AxisTopToBottom {
		t.Error("vertical bar should use column axis")
	}
}

func TestProgressBarTextShow(t *testing.T) {
	v := ProgressBar(ProgressBarCfg{
		Percent:  0.5,
		TextShow: true,
	})
	layout := GenerateViewLayout(v, &Window{})
	if len(layout.Children) != 2 {
		t.Fatalf("with text: got %d children, want 2",
			len(layout.Children))
	}
}

func TestProgressBarNoText(t *testing.T) {
	v := ProgressBar(ProgressBarCfg{
		Percent:  0.5,
		TextShow: false,
	})
	layout := GenerateViewLayout(v, &Window{})
	if len(layout.Children) != 1 {
		t.Fatalf("without text: got %d children, want 1",
			len(layout.Children))
	}
}

func TestProgressBarA11YRole(t *testing.T) {
	v := ProgressBar(ProgressBarCfg{
		Percent: 0.3,
	})
	layout := GenerateViewLayout(v, &Window{})
	if layout.Shape.A11YRole != AccessRoleProgressBar {
		t.Errorf("role = %d, want ProgressBar", layout.Shape.A11YRole)
	}
	if layout.Shape.A11Y == nil {
		t.Fatal("a11y should be set")
	}
	if layout.Shape.A11Y.ValueMax != 1 {
		t.Error("value_max should be 1")
	}
}

func TestProgressBarIndefiniteA11Y(t *testing.T) {
	v := ProgressBar(ProgressBarCfg{
		Indefinite: true,
	})
	layout := GenerateViewLayout(v, &Window{})
	want := AccessStateBusy | AccessStateLive
	if layout.Shape.A11YState != want {
		t.Errorf("state = %d, want busy|live (%d)",
			layout.Shape.A11YState, want)
	}
}
