package gui

import "testing"

func TestMarkLayoutRefreshClearsRenderOnly(t *testing.T) {
	w := &Window{refreshRenderOnly: true}
	w.markLayoutRefresh()
	if !w.refreshLayout {
		t.Error("refreshLayout should be true")
	}
	if w.refreshRenderOnly {
		t.Error("refreshRenderOnly should be false")
	}
}

func TestMarkRenderOnlyRefreshSetsWhenLayoutNotPending(t *testing.T) {
	w := &Window{}
	w.markRenderOnlyRefresh()
	if w.refreshLayout {
		t.Error("refreshLayout should be false")
	}
	if !w.refreshRenderOnly {
		t.Error("refreshRenderOnly should be true")
	}
}

func TestMarkRenderOnlyRefreshSkipsWhenLayoutPending(t *testing.T) {
	w := &Window{refreshLayout: true}
	w.markRenderOnlyRefresh()
	if !w.refreshLayout {
		t.Error("refreshLayout should remain true")
	}
	if w.refreshRenderOnly {
		t.Error("refreshRenderOnly should remain false")
	}
}

func TestMaxAnimationRefreshKindPrefersLayout(t *testing.T) {
	kind := maxAnimationRefreshKind(AnimationRefreshRenderOnly, AnimationRefreshLayout)
	if kind != AnimationRefreshLayout {
		t.Errorf("got %d, want layout", kind)
	}
}

func TestMaxAnimationRefreshKindPrefersRenderOnlyOverNone(t *testing.T) {
	kind := maxAnimationRefreshKind(AnimationRefreshNone, AnimationRefreshRenderOnly)
	if kind != AnimationRefreshRenderOnly {
		t.Errorf("got %d, want render_only", kind)
	}
}

func TestBlinkCursorAnimationRefreshKindIsRenderOnly(t *testing.T) {
	a := NewBlinkCursorAnimation()
	if a.RefreshKind() != AnimationRefreshRenderOnly {
		t.Errorf("got %d, want render_only", a.RefreshKind())
	}
}

func TestAnimateRefreshKindIsLayout(t *testing.T) {
	a := &Animate{
		AnimateID: "test",
		Callback:  func(*Animate, *Window) {},
	}
	if a.RefreshKind() != AnimationRefreshLayout {
		t.Errorf("got %d, want layout", a.RefreshKind())
	}
}

func TestAnimateRefreshKindOverride(t *testing.T) {
	a := &Animate{
		AnimateID: "test",
		Callback:  func(*Animate, *Window) {},
		Refresh:   AnimationRefreshRenderOnly,
	}
	if a.RefreshKind() != AnimationRefreshRenderOnly {
		t.Errorf("got %d, want render_only", a.RefreshKind())
	}
}
