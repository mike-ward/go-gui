package gui

import "testing"

func TestTreeNodeIDFallback(t *testing.T) {
	if got := treeNodeID(TreeNodeCfg{Text: "alpha"}); got != "alpha" {
		t.Errorf("treeNodeID(Text=alpha) = %q, want %q", got, "alpha")
	}
	if got := treeNodeID(TreeNodeCfg{ID: "id-alpha", Text: "alpha"}); got != "id-alpha" {
		t.Errorf("treeNodeID(ID=id-alpha, Text=alpha) = %q, want %q", got, "id-alpha")
	}
}

func TestTreeNodeIDFallbackCanCollide(t *testing.T) {
	w := newTestWindow()
	var rows []treeFlatRow

	treeCollectFlatRows(
		[]TreeNodeCfg{{Text: "dup"}, {Text: "dup"}},
		nil,
		"tree",
		StateMap[string, bool](w, nsTreeLazy, capMany),
		&rows,
		0,
		"",
	)

	if got := len(rows); got != 2 {
		t.Fatalf("len(rows) = %d, want 2", got)
	}
	if rows[0].ID != rows[1].ID {
		t.Fatalf("rows[0].ID = %q, rows[1].ID = %q, want duplicate IDs to document caller uniqueness requirements", rows[0].ID, rows[1].ID)
	}
}

func TestTreeCollectFlatRowsExpanded(t *testing.T) {
	w := newTestWindow()
	var rows []treeFlatRow

	treeCollectFlatRows(
		[]TreeNodeCfg{
			{
				ID:   "root",
				Text: "Root",
				Nodes: []TreeNodeCfg{
					{ID: "child", Text: "Child"},
				},
			},
			{ID: "sibling", Text: "Sibling"},
		},
		map[string]bool{"root": true},
		"tree",
		StateMap[string, bool](w, nsTreeLazy, capMany),
		&rows,
		0,
		"",
	)

	if got := len(rows); got != 3 {
		t.Fatalf("len(rows) = %d, want 3", got)
	}
	if rows[0].ID != "root" || rows[0].Depth != 0 {
		t.Fatalf("rows[0] = %+v, want root depth 0", rows[0])
	}
	if rows[1].ID != "child" || rows[1].Depth != 1 || rows[1].ParentID != "root" {
		t.Fatalf("rows[1] = %+v, want child depth 1 parent root", rows[1])
	}
	if rows[2].ID != "sibling" || rows[2].Depth != 0 {
		t.Fatalf("rows[2] = %+v, want sibling depth 0", rows[2])
	}
}

func TestTreeCollectFlatRowsCollapsed(t *testing.T) {
	w := newTestWindow()
	var rows []treeFlatRow

	treeCollectFlatRows(
		[]TreeNodeCfg{
			{
				ID:   "root",
				Text: "Root",
				Nodes: []TreeNodeCfg{
					{ID: "child", Text: "Child"},
				},
			},
		},
		nil,
		"tree",
		StateMap[string, bool](w, nsTreeLazy, capMany),
		&rows,
		0,
		"",
	)

	if got := len(rows); got != 1 {
		t.Fatalf("len(rows) = %d, want 1", got)
	}
	if rows[0].ID != "root" {
		t.Fatalf("rows[0].ID = %q, want %q", rows[0].ID, "root")
	}
}

func TestTreeCollectFlatRowsLazyLoading(t *testing.T) {
	w := newTestWindow()
	var rows []treeFlatRow

	lazyState := StateMap[string, bool](w, nsTreeLazy, capMany)
	lazyState.Set(treeLazyKey("tree", "remote"), true)

	treeCollectFlatRows(
		[]TreeNodeCfg{{ID: "remote", Text: "Remote", Lazy: true}},
		map[string]bool{"remote": true},
		"tree",
		lazyState,
		&rows,
		0,
		"",
	)

	if got := len(rows); got != 2 {
		t.Fatalf("len(rows) = %d, want 2", got)
	}
	if rows[1].ID != "remote"+treeLoadingSuffix || !rows[1].IsLoading {
		t.Fatalf("rows[1] = %+v, want loading sentinel", rows[1])
	}
}

func TestTreeCollectFlatRowsLazyAutoClear(t *testing.T) {
	w := newTestWindow()
	var rows []treeFlatRow

	lazyState := StateMap[string, bool](w, nsTreeLazy, capMany)
	key := treeLazyKey("tree", "remote")
	lazyState.Set(key, true)

	treeCollectFlatRows(
		[]TreeNodeCfg{
			{
				ID:   "remote",
				Text: "Remote",
				Lazy: true,
				Nodes: []TreeNodeCfg{
					{ID: "child", Text: "Child"},
				},
			},
		},
		map[string]bool{"remote": true},
		"tree",
		lazyState,
		&rows,
		0,
		"",
	)

	if _, ok := lazyState.Get(key); ok {
		t.Fatalf("lazyState.Get(%q) = present, want entry cleared", key)
	}
	if got := len(rows); got != 2 {
		t.Fatalf("len(rows) = %d, want 2", got)
	}
	if rows[1].IsLoading {
		t.Fatalf("rows[1].IsLoading = %t, want false", rows[1].IsLoading)
	}
}

func TestTreeVisibleRange(t *testing.T) {
	w := newTestWindow()
	StateMap[uint32, float32](w, nsScrollY, capScroll).Set(55, -40)

	gotFirst, gotLast := treeVisibleRange(50, 20, 100, 55, w)
	wantFirst, wantLast := listCoreVisibleRange(100, 20, 50, -40)
	if gotFirst != wantFirst || gotLast != wantLast {
		t.Fatalf("treeVisibleRange() = (%d, %d), want (%d, %d)", gotFirst, gotLast, wantFirst, wantLast)
	}
}

func TestTreeRowClickTogglesAndSelects(t *testing.T) {
	w := newTestWindow()
	row := treeFlatRow{
		ID:          "remote",
		HasChildren: true,
		IsLazy:      true,
	}

	selectedID := ""
	selectSawExpanded := false
	lazyCalls := 0
	e := &Event{}
	treeRowClick(
		"tree",
		row,
		88,
		func(id string, _ *Event, w *Window) {
			selectedID = id
			selectSawExpanded = treeExpandedState(w, "tree")["remote"]
		},
		func(treeID, nodeID string, _ *Window) {
			if treeID != "tree" || nodeID != "remote" {
				t.Fatalf("OnLazyLoad(%q, %q), want (%q, %q)", treeID, nodeID, "tree", "remote")
			}
			lazyCalls++
		},
		e,
		w,
	)

	if got := StateReadOr[string, string](w, nsTreeFocus, "tree", ""); got != "remote" {
		t.Fatalf("focused node = %q, want %q", got, "remote")
	}
	if !treeExpandedState(w, "tree")["remote"] {
		t.Fatal("remote should be expanded after click")
	}
	if got := selectedID; got != "remote" {
		t.Fatalf("selectedID = %q, want %q", got, "remote")
	}
	if !selectSawExpanded {
		t.Fatal("OnSelect should observe expanded state already applied")
	}
	if got := lazyCalls; got != 1 {
		t.Fatalf("lazyCalls = %d, want 1", got)
	}
	if loading, ok := StateMap[string, bool](w, nsTreeLazy, capMany).Get(treeLazyKey("tree", "remote")); !ok || !loading {
		t.Fatal("lazy loading state should be set after first expansion")
	}
	if got := w.IDFocus(); got != 88 {
		t.Fatalf("w.IDFocus() = %d, want 88", got)
	}
	if !e.IsHandled {
		t.Fatal("click event should be handled")
	}
}

func TestTreeRowClickCollapseClearsLoading(t *testing.T) {
	w := newTestWindow()
	treeExpandedSet(w, "tree", "remote", true)
	StateMap[string, bool](w, nsTreeLazy, capMany).Set(treeLazyKey("tree", "remote"), true)

	treeRowClick(
		"tree",
		treeFlatRow{ID: "remote", HasChildren: true, IsExpanded: true, IsLazy: true},
		0,
		nil,
		nil,
		&Event{},
		w,
	)

	if treeExpandedState(w, "tree")["remote"] {
		t.Fatal("remote should be collapsed after click")
	}
	if _, ok := StateMap[string, bool](w, nsTreeLazy, capMany).Get(treeLazyKey("tree", "remote")); ok {
		t.Fatal("lazy loading state should be cleared on collapse")
	}
}

func TestTreeOnKeyDownNavigation(t *testing.T) {
	w := newTestWindow()
	visibleIDs := []string{"a", "b", "c"}
	rowByID := map[string]treeFlatRow{
		"a": {ID: "a"},
		"b": {ID: "b"},
		"c": {ID: "c"},
	}
	treeFocusedSet(w, "tree", "b")

	eUp := &Event{KeyCode: KeyUp}
	treeOnKeyDown("tree", visibleIDs, rowByID, nil, nil, 0, 0, 0, eUp, w)
	if got := StateReadOr[string, string](w, nsTreeFocus, "tree", ""); got != "a" {
		t.Fatalf("focus after KeyUp = %q, want %q", got, "a")
	}

	eEnd := &Event{KeyCode: KeyEnd}
	treeOnKeyDown("tree", visibleIDs, rowByID, nil, nil, 0, 0, 0, eEnd, w)
	if got := StateReadOr[string, string](w, nsTreeFocus, "tree", ""); got != "c" {
		t.Fatalf("focus after KeyEnd = %q, want %q", got, "c")
	}

	eHome := &Event{KeyCode: KeyHome}
	treeOnKeyDown("tree", visibleIDs, rowByID, nil, nil, 0, 0, 0, eHome, w)
	if got := StateReadOr[string, string](w, nsTreeFocus, "tree", ""); got != "a" {
		t.Fatalf("focus after KeyHome = %q, want %q", got, "a")
	}

	selectedID := ""
	eEnter := &Event{KeyCode: KeyEnter}
	treeOnKeyDown(
		"tree",
		visibleIDs,
		rowByID,
		func(id string, _ *Event, _ *Window) { selectedID = id },
		nil,
		0, 0, 0,
		eEnter,
		w,
	)
	if got := selectedID; got != "a" {
		t.Fatalf("selectedID after KeyEnter = %q, want %q", got, "a")
	}
	if !eEnter.IsHandled {
		t.Fatal("KeyEnter event should be handled")
	}
}

func TestTreeOnKeyDownLeftCollapses(t *testing.T) {
	w := newTestWindow()
	treeFocusedSet(w, "tree", "root")
	treeExpandedSet(w, "tree", "root", true)
	StateMap[string, bool](w, nsTreeLazy, capMany).Set(treeLazyKey("tree", "root"), true)

	treeOnKeyDown(
		"tree",
		[]string{"root"},
		map[string]treeFlatRow{
			"root": {
				ID:          "root",
				HasChildren: true,
				IsExpanded:  true,
			},
		},
		nil,
		nil,
		0, 0, 0,
		&Event{KeyCode: KeyLeft},
		w,
	)

	if treeExpandedState(w, "tree")["root"] {
		t.Fatal("root should be collapsed after KeyLeft")
	}
	if _, ok := StateMap[string, bool](w, nsTreeLazy, capMany).Get(treeLazyKey("tree", "root")); ok {
		t.Fatal("KeyLeft should clear lazy loading state")
	}
}

func TestTreeOnKeyDownRightTriggersLazyLoad(t *testing.T) {
	w := newTestWindow()
	treeFocusedSet(w, "tree", "remote")
	visibleIDs := []string{"remote"}
	rowByID := map[string]treeFlatRow{
		"remote": {
			ID:          "remote",
			HasChildren: true,
			IsLazy:      true,
		},
	}

	lazyCalls := 0
	onLazyLoad := func(treeID, nodeID string, _ *Window) {
		if treeID != "tree" || nodeID != "remote" {
			t.Fatalf("OnLazyLoad(%q, %q), want (%q, %q)", treeID, nodeID, "tree", "remote")
		}
		lazyCalls++
	}

	eRight := &Event{KeyCode: KeyRight}
	treeOnKeyDown("tree", visibleIDs, rowByID, nil, onLazyLoad, 0, 0, 0, eRight, w)
	treeOnKeyDown("tree", visibleIDs, rowByID, nil, onLazyLoad, 0, 0, 0, &Event{KeyCode: KeyRight}, w)

	if !treeExpandedState(w, "tree")["remote"] {
		t.Fatal("remote should be expanded after KeyRight")
	}
	if got := lazyCalls; got != 1 {
		t.Fatalf("lazyCalls = %d, want 1", got)
	}
	if loading, ok := StateMap[string, bool](w, nsTreeLazy, capMany).Get(treeLazyKey("tree", "remote")); !ok || !loading {
		t.Fatal("lazy loading state should be set after KeyRight")
	}
	if !eRight.IsHandled {
		t.Fatal("first KeyRight event should be handled")
	}
}

func TestTreeGenerateLayoutA11Y(t *testing.T) {
	w := newTestWindow()
	treeExpandedSet(w, "tree", "root", true)

	layout := GenerateViewLayout(Tree(TreeCfg{
		ID: "tree",
		Nodes: []TreeNodeCfg{
			{
				ID:   "root",
				Text: "Root",
				Nodes: []TreeNodeCfg{
					{ID: "child", Text: "Child"},
				},
			},
		},
	}), w)

	if layout.Shape.A11YRole != AccessRoleTree {
		t.Fatalf("layout.Shape.A11YRole = %d, want %d", layout.Shape.A11YRole, AccessRoleTree)
	}
	if got := len(layout.Children); got != 2 {
		t.Fatalf("len(layout.Children) = %d, want 2", got)
	}

	rootRow := layout.Children[0]
	if rootRow.Shape.A11YRole != AccessRoleTreeItem {
		t.Fatalf("rootRow.Shape.A11YRole = %d, want %d", rootRow.Shape.A11YRole, AccessRoleTreeItem)
	}
	if !rootRow.Shape.A11YState.Has(AccessStateExpanded) {
		t.Fatal("expanded root row should expose AccessStateExpanded")
	}

	childRow := layout.Children[1]
	if childRow.Shape.A11YRole != AccessRoleTreeItem {
		t.Fatalf("childRow.Shape.A11YRole = %d, want %d", childRow.Shape.A11YRole, AccessRoleTreeItem)
	}
}
