package gui

func runQueuedCommands(cmds []queuedCommand) {
	w := &Window{}
	for i := range cmds {
		cmd := cmds[i]
		switch cmd.kind {
		case queuedCommandWindowFn:
			if cmd.windowFn != nil {
				cmd.windowFn(w)
			}
		case queuedCommandValueFn:
			if cmd.valueFn != nil {
				cmd.valueFn(cmd.value, w)
			}
		case queuedCommandAnimateFn:
			if cmd.animateFn != nil {
				cmd.animateFn(cmd.animate, w)
			}
		}
	}
}
