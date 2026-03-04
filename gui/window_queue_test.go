package gui

import "testing"

type queueTestState struct {
	count int
}

func TestFlushCommandsRunsAllAndClears(t *testing.T) {
	state := &queueTestState{}
	w := &Window{}
	w.state = state
	w.commandsMu.Lock()
	w.commands = append(w.commands, queuedCommand{
		kind: queuedCommandWindowFn,
		windowFn: func(win *Window) {
			State[queueTestState](win).count++
		},
	})
	w.commands = append(w.commands, queuedCommand{
		kind: queuedCommandWindowFn,
		windowFn: func(win *Window) {
			State[queueTestState](win).count++
		},
	})
	w.commandsMu.Unlock()

	w.flushCommands()

	if state.count != 2 {
		t.Errorf("count: got %d, want 2", state.count)
	}
	if len(w.commands) != 0 {
		t.Errorf("commands not cleared: len=%d", len(w.commands))
	}
}

func TestFlushCommandsEmptyQueueIsNoop(t *testing.T) {
	w := &Window{}
	w.flushCommands()
	if len(w.commands) != 0 {
		t.Errorf("commands: got %d, want 0", len(w.commands))
	}
}
