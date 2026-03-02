package gui

import "testing"

func TestA11yCollectEmpty(t *testing.T) {
	layout := Layout{}
	var nodes []A11yNode
	var live []liveNode
	idx := a11yCollect(&layout, -1, &nodes, 0, &live)
	if idx != -1 {
		t.Errorf("focusedIdx: got %d, want -1", idx)
	}
	if len(nodes) != 0 {
		t.Errorf("nodes: got %d, want 0", len(nodes))
	}
}

func TestA11yCollectNilShape(t *testing.T) {
	layout := Layout{Shape: nil}
	var nodes []A11yNode
	var live []liveNode
	idx := a11yCollect(&layout, -1, &nodes, 0, &live)
	if idx != -1 || len(nodes) != 0 {
		t.Error("nil shape should produce no nodes")
	}
}

func TestA11yCollectSkipNoneRole(t *testing.T) {
	layout := Layout{
		Shape: &Shape{A11YRole: AccessRoleNone},
	}
	var nodes []A11yNode
	var live []liveNode
	a11yCollect(&layout, -1, &nodes, 0, &live)
	if len(nodes) != 0 {
		t.Errorf("AccessRoleNone should not emit a node, got %d", len(nodes))
	}
}

func TestA11yCollectSingleButton(t *testing.T) {
	layout := Layout{
		Shape: &Shape{
			A11YRole: AccessRoleButton,
			A11Y:     &AccessInfo{Label: "OK"},
			X:        10, Y: 20, Width: 100, Height: 30,
		},
	}
	var nodes []A11yNode
	var live []liveNode
	a11yCollect(&layout, -1, &nodes, 0, &live)
	if len(nodes) != 1 {
		t.Fatalf("expected 1 node, got %d", len(nodes))
	}
	n := nodes[0]
	if n.Role != AccessRoleButton {
		t.Errorf("Role: got %d", n.Role)
	}
	if n.Label != "OK" {
		t.Errorf("Label: got %q", n.Label)
	}
	if n.X != 10 || n.Y != 20 || n.W != 100 || n.H != 30 {
		t.Errorf("bounds: %g,%g %gx%g", n.X, n.Y, n.W, n.H)
	}
	if n.ParentIdx != -1 {
		t.Errorf("ParentIdx: got %d, want -1", n.ParentIdx)
	}
}

func TestA11yCollectLabelFromText(t *testing.T) {
	layout := Layout{
		Shape: &Shape{
			A11YRole: AccessRoleButton,
			TC:       &ShapeTextConfig{Text: "Submit"},
		},
	}
	var nodes []A11yNode
	var live []liveNode
	a11yCollect(&layout, -1, &nodes, 0, &live)
	if len(nodes) != 1 {
		t.Fatalf("expected 1 node, got %d", len(nodes))
	}
	if nodes[0].Label != "Submit" {
		t.Errorf("Label: got %q, want %q", nodes[0].Label, "Submit")
	}
}

func TestA11yCollectA11YLabelOverridesText(t *testing.T) {
	layout := Layout{
		Shape: &Shape{
			A11YRole: AccessRoleButton,
			A11Y:     &AccessInfo{Label: "Close dialog"},
			TC:       &ShapeTextConfig{Text: "X"},
		},
	}
	var nodes []A11yNode
	var live []liveNode
	a11yCollect(&layout, -1, &nodes, 0, &live)
	if nodes[0].Label != "Close dialog" {
		t.Errorf("Label: got %q", nodes[0].Label)
	}
}

func TestA11yCollectWithChildren(t *testing.T) {
	layout := Layout{
		Shape: &Shape{A11YRole: AccessRoleGroup},
		Children: []Layout{
			{Shape: &Shape{A11YRole: AccessRoleButton, A11Y: &AccessInfo{Label: "A"}}},
			{Shape: &Shape{A11YRole: AccessRoleButton, A11Y: &AccessInfo{Label: "B"}}},
		},
	}
	var nodes []A11yNode
	var live []liveNode
	a11yCollect(&layout, -1, &nodes, 0, &live)
	if len(nodes) != 3 {
		t.Fatalf("expected 3 nodes, got %d", len(nodes))
	}
	// Parent is first node (group).
	if nodes[0].ChildrenStart != 1 || nodes[0].ChildrenCount != 2 {
		t.Errorf("parent children: start=%d count=%d",
			nodes[0].ChildrenStart, nodes[0].ChildrenCount)
	}
	// Children have parent_idx = 0.
	if nodes[1].ParentIdx != 0 || nodes[2].ParentIdx != 0 {
		t.Errorf("child parent: %d, %d", nodes[1].ParentIdx, nodes[2].ParentIdx)
	}
}

func TestA11yCollectNoneRolePassesChildren(t *testing.T) {
	// Parent has no role but child does.
	layout := Layout{
		Shape: &Shape{A11YRole: AccessRoleNone},
		Children: []Layout{
			{Shape: &Shape{A11YRole: AccessRoleButton, A11Y: &AccessInfo{Label: "Child"}}},
		},
	}
	var nodes []A11yNode
	var live []liveNode
	a11yCollect(&layout, -1, &nodes, 0, &live)
	if len(nodes) != 1 {
		t.Fatalf("expected 1 node from child, got %d", len(nodes))
	}
	if nodes[0].Label != "Child" {
		t.Errorf("Label: got %q", nodes[0].Label)
	}
	// Parent was -1 (root context), child inherits that.
	if nodes[0].ParentIdx != -1 {
		t.Errorf("ParentIdx: got %d, want -1", nodes[0].ParentIdx)
	}
}

func TestA11yCollectFocusTracking(t *testing.T) {
	layout := Layout{
		Shape: &Shape{A11YRole: AccessRoleGroup},
		Children: []Layout{
			{Shape: &Shape{A11YRole: AccessRoleButton, IDFocus: 1}},
			{Shape: &Shape{A11YRole: AccessRoleButton, IDFocus: 2}},
		},
	}
	var nodes []A11yNode
	var live []liveNode
	idx := a11yCollect(&layout, -1, &nodes, 2, &live)
	if idx != 2 { // third node (index 2) has IDFocus=2
		t.Errorf("focusedIdx: got %d, want 2", idx)
	}
}

func TestA11yCollectNoFocus(t *testing.T) {
	layout := Layout{
		Shape: &Shape{A11YRole: AccessRoleButton, IDFocus: 1},
	}
	var nodes []A11yNode
	var live []liveNode
	idx := a11yCollect(&layout, -1, &nodes, 99, &live)
	if idx != -1 {
		t.Errorf("focusedIdx: got %d, want -1", idx)
	}
}

func TestA11yCollectLiveRegion(t *testing.T) {
	layout := Layout{
		Shape: &Shape{
			A11YRole:  AccessRoleStaticText,
			A11YState: AccessStateLive,
			A11Y:      &AccessInfo{Label: "status", ValueNum: 42},
		},
	}
	var nodes []A11yNode
	var live []liveNode
	a11yCollect(&layout, -1, &nodes, 0, &live)
	if len(live) != 1 {
		t.Fatalf("expected 1 live node, got %d", len(live))
	}
	if live[0].label != "status" {
		t.Errorf("label: got %q", live[0].label)
	}
}

func TestA11yValueText(t *testing.T) {
	tests := []struct {
		info *AccessInfo
		want string
	}{
		{&AccessInfo{}, ""},
		{&AccessInfo{ValueNum: 42}, "42"},
		{&AccessInfo{ValueNum: 0.5, ValueMin: 0, ValueMax: 1}, "0.5"},
		{&AccessInfo{ValueNum: 0, ValueMin: 0, ValueMax: 0}, ""},
	}
	for _, tt := range tests {
		got := a11yValueText(tt.info)
		if got != tt.want {
			t.Errorf("a11yValueText(%+v) = %q, want %q", tt.info, got, tt.want)
		}
	}
}

func TestShapeA11yLabel(t *testing.T) {
	s := &Shape{TC: &ShapeTextConfig{Text: "Hello"}}
	if got := shapeA11yLabel(s); got != "Hello" {
		t.Errorf("got %q", got)
	}
	s2 := &Shape{}
	if got := shapeA11yLabel(s2); got != "" {
		t.Errorf("got %q, want empty", got)
	}
}

func TestA11yDeepNesting(t *testing.T) {
	// 3-level nesting: root > group > button.
	layout := Layout{
		Shape: &Shape{A11YRole: AccessRoleGroup},
		Children: []Layout{
			{
				Shape: &Shape{A11YRole: AccessRoleGroup},
				Children: []Layout{
					{Shape: &Shape{A11YRole: AccessRoleButton, A11Y: &AccessInfo{Label: "Deep"}}},
				},
			},
		},
	}
	var nodes []A11yNode
	var live []liveNode
	a11yCollect(&layout, -1, &nodes, 0, &live)
	if len(nodes) != 3 {
		t.Fatalf("expected 3 nodes, got %d", len(nodes))
	}
	// Root's child is at index 1.
	if nodes[2].ParentIdx != 1 {
		t.Errorf("deep child parent: got %d, want 1", nodes[2].ParentIdx)
	}
}

func TestWindowCleanup(t *testing.T) {
	w := &Window{}
	w.storeBookmark("/a", nil)
	w.storeBookmark("/b", nil)
	w.WindowCleanup()
	if w.FileAccessGrantCount() != 0 {
		t.Errorf("grants not released: %d", w.FileAccessGrantCount())
	}
}
