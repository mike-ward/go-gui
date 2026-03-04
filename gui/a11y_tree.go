package gui

import "fmt"

// Accessibility action constants.
const (
	A11yActionPress     = 0
	A11yActionIncrement = 1
	A11yActionDecrement = 2
	A11yActionConfirm   = 3
	A11yActionCancel    = 4
)

// A11yNode is a flat accessibility node pushed to the native
// accessibility backend each frame.
type A11yNode struct {
	Role          AccessRole
	State         AccessState
	Label         string
	Value         string
	Description   string
	X, Y, W, H    float32
	ParentIdx     int
	ChildrenStart int
	ChildrenCount int
}

type liveNode struct {
	label string
	value string
}

// a11y holds per-window accessibility backend state.
type a11y struct {
	initialized    bool
	prevIDFocus    uint32
	prevLiveValues map[string]string
	nodes          []A11yNode // reused across frames
	liveNodes      []liveNode // reused across frames
}

// initA11y lazily creates the native accessibility container.
// Called from frame loop, same pattern as IME.
func (w *Window) initA11y() {
	if w.a11y.initialized {
		return
	}
	w.a11y.initialized = true

	if w.nativePlatform != nil {
		w.nativePlatform.A11yInit(func(action, index int) {
			a11yActionCallback(w, action, index)
		})
	}
}

// syncA11y walks the layout tree, builds a flat node array,
// and pushes it to the native accessibility backend.
func (w *Window) syncA11y() {
	if w.nativePlatform == nil || !w.a11y.initialized {
		return
	}
	if w.layout.Shape == nil {
		return
	}

	// Reuse slices across frames.
	w.a11y.nodes = w.a11y.nodes[:0]
	w.a11y.liveNodes = w.a11y.liveNodes[:0]

	focusedIdx := a11yCollect(
		&w.layout, -1,
		&w.a11y.nodes,
		w.viewState.idFocus,
		&w.a11y.liveNodes,
	)

	if len(w.a11y.nodes) == 0 {
		return
	}

	w.nativePlatform.A11ySync(w.a11y.nodes, len(w.a11y.nodes), focusedIdx)
	_ = w.viewState.idFocus // suppress unused warning on focusedIdx

	// Live region change detection.
	for _, ln := range w.a11y.liveNodes {
		if prev, ok := w.a11y.prevLiveValues[ln.label]; ok {
			if prev != ln.value {
				w.nativePlatform.A11yAnnounce(ln.value)
			}
		}
	}
	// Update previous values.
	if w.a11y.prevLiveValues == nil {
		w.a11y.prevLiveValues = make(map[string]string)
	}
	clear(w.a11y.prevLiveValues)
	for _, ln := range w.a11y.liveNodes {
		w.a11y.prevLiveValues[ln.label] = ln.value
	}
	w.a11y.prevIDFocus = w.viewState.idFocus
}

// a11yCollect recursively walks the layout tree and appends
// nodes to the flat array. Returns the index of the focused
// node, or -1 if not found.
func a11yCollect(
	layout *Layout,
	parentIdx int,
	nodes *[]A11yNode,
	idFocus uint32,
	live *[]liveNode,
) int {
	focusedIdx := -1
	if layout.Shape == nil {
		return focusedIdx
	}
	s := layout.Shape

	// Skip shapes without a11y role but recurse children.
	if s.A11YRole == AccessRoleNone {
		for i := range layout.Children {
			if fi := a11yCollect(&layout.Children[i], parentIdx, nodes, idFocus, live); fi >= 0 {
				focusedIdx = fi
			}
		}
		return focusedIdx
	}

	nodeIdx := len(*nodes)

	// Build label from AccessInfo or shape text.
	label := ""
	value := ""
	description := ""
	if s.A11Y != nil {
		if s.A11Y.Label != "" {
			label = s.A11Y.Label
		}
		description = s.A11Y.Description
		value = a11yValueText(s.A11Y)
	}
	if label == "" {
		label = shapeA11yLabel(s)
	}

	*nodes = append(*nodes, A11yNode{
		Role:        s.A11YRole,
		State:       s.A11YState,
		Label:       label,
		Value:       value,
		Description: description,
		X:           s.X,
		Y:           s.Y,
		W:           s.Width,
		H:           s.Height,
		ParentIdx:   parentIdx,
	})

	if idFocus > 0 && s.IDFocus == idFocus {
		focusedIdx = nodeIdx
	}

	// Track live regions.
	if s.A11YState.Has(AccessStateLive) {
		*live = append(*live, liveNode{label: label, value: value})
	}

	// Process children.
	childrenStart := len(*nodes)
	for i := range layout.Children {
		if fi := a11yCollect(&layout.Children[i], nodeIdx, nodes, idFocus, live); fi >= 0 {
			focusedIdx = fi
		}
	}
	childrenCount := len(*nodes) - childrenStart

	// Update node's children info.
	(*nodes)[nodeIdx].ChildrenStart = childrenStart
	(*nodes)[nodeIdx].ChildrenCount = childrenCount

	return focusedIdx
}

// a11yValueText formats AccessInfo numeric values as text.
func a11yValueText(info *AccessInfo) string {
	if info.ValueNum == 0 && info.ValueMin == 0 && info.ValueMax == 0 {
		return ""
	}
	return fmt.Sprintf("%g", info.ValueNum)
}

// shapeA11yLabel extracts an accessibility label from shape text.
func shapeA11yLabel(s *Shape) string {
	if s.TC != nil && s.TC.Text != "" {
		return s.TC.Text
	}
	return ""
}

// a11yActionCallback routes native accessibility actions to
// the focused layout node's event handlers.
func a11yActionCallback(w *Window, action, index int) {
	// Action dispatch is a stub — the full implementation
	// routes actions to event handlers on the focused shape.
	_ = w
	_ = action
	_ = index
}

// WindowCleanup releases resources. Called by the backend
// during window destruction.
func (w *Window) WindowCleanup() {
	w.cleanupOnce.Do(func() {
		w.stopAnimationLoop()
		w.ReleaseAllFileAccess()
		if w.nativePlatform != nil {
			w.nativePlatform.A11yDestroy()
		}
	})
}
