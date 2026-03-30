package gui

import "testing"

func TestQueueOnDoneAppendsCommand(t *testing.T) {
	var cmds []queuedCommand
	called := false
	queueOnDone(&cmds, func(_ *Window) { called = true })

	if len(cmds) != 1 {
		t.Fatalf("len = %d, want 1", len(cmds))
	}
	if cmds[0].kind != queuedCommandWindowFn {
		t.Errorf("kind = %d, want %d", cmds[0].kind, queuedCommandWindowFn)
	}
	cmds[0].windowFn(nil)
	if !called {
		t.Error("callback not invoked")
	}
}

func TestQueueOnDoneNilSkips(t *testing.T) {
	var cmds []queuedCommand
	queueOnDone(&cmds, nil)

	if len(cmds) != 0 {
		t.Errorf("len = %d, want 0 for nil fn", len(cmds))
	}
}

func TestQueueOnValueAppendsCommand(t *testing.T) {
	var cmds []queuedCommand
	var gotVal float32
	queueOnValue(&cmds, func(v float32, _ *Window) {
		gotVal = v
	}, 42.5)

	if len(cmds) != 1 {
		t.Fatalf("len = %d, want 1", len(cmds))
	}
	if cmds[0].kind != queuedCommandValueFn {
		t.Errorf("kind = %d, want %d", cmds[0].kind, queuedCommandValueFn)
	}
	if cmds[0].value != 42.5 {
		t.Errorf("value = %f, want 42.5", cmds[0].value)
	}
	cmds[0].valueFn(cmds[0].value, nil)
	if gotVal != 42.5 {
		t.Errorf("callback got %f, want 42.5", gotVal)
	}
}

func TestCommandMarkLayoutRefreshSetsFlag(t *testing.T) {
	w := &Window{}
	commandMarkLayoutRefresh(w)
	if !w.refreshLayout {
		t.Error("refreshLayout should be true")
	}
}

func TestCommandMarkRenderOnlyRefreshSetsFlag(t *testing.T) {
	w := &Window{}
	commandMarkRenderOnlyRefresh(w)
	if !w.refreshRenderOnly {
		t.Error("refreshRenderOnly should be true")
	}
}
