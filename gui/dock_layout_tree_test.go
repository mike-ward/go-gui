package gui

import (
	"slices"
	"testing"
)

// --- helpers ---

func makeTestTree() *DockNode {
	// left group: panels A, B (selected A)
	// right split:
	//   top group: panel C (selected C)
	//   bottom group: panels D, E (selected D)
	left := DockPanelGroup("g1", []string{"A", "B"}, "A")
	topRight := DockPanelGroup("g2", []string{"C"}, "C")
	bottomRight := DockPanelGroup("g3", []string{"D", "E"}, "D")
	right := DockSplit("s2", DockSplitVertical, 0.6, topRight, bottomRight)
	return DockSplit("s1", DockSplitHorizontal, 0.3, left, right)
}

// --- DockSplit / DockPanelGroup constructors ---

func TestDockSplitConstructor(t *testing.T) {
	a := DockPanelGroup("a", []string{"p1"}, "p1")
	b := DockPanelGroup("b", []string{"p2"}, "p2")
	s := DockSplit("s", DockSplitHorizontal, 0.4, a, b)
	if s.Kind != DockNodeSplit {
		t.Fatal("expected split")
	}
	if s.Dir != DockSplitHorizontal {
		t.Fatal("expected horizontal")
	}
	if s.Ratio != 0.4 {
		t.Fatalf("ratio = %f, want 0.4", s.Ratio)
	}
	if s.First != a || s.Second != b {
		t.Fatal("children mismatch")
	}
}

func TestDockPanelGroupConstructor(t *testing.T) {
	g := DockPanelGroup("g", []string{"x", "y"}, "y")
	if g.Kind != DockNodePanelGroup {
		t.Fatal("expected panel_group")
	}
	if g.ID != "g" {
		t.Fatal("wrong id")
	}
	if len(g.PanelIDs) != 2 || g.PanelIDs[0] != "x" {
		t.Fatal("wrong panel_ids")
	}
	if g.SelectedID != "y" {
		t.Fatal("wrong selected_id")
	}
}

// --- CollectPanelNodes ---

func TestCollectPanelNodes(t *testing.T) {
	root := makeTestTree()
	nodes := DockTreeCollectPanelNodes(root)
	if len(nodes) != 3 {
		t.Fatalf("got %d nodes, want 3", len(nodes))
	}
	ids := []string{nodes[0].ID, nodes[1].ID, nodes[2].ID}
	slices.Sort(ids)
	if ids[0] != "g1" || ids[1] != "g2" || ids[2] != "g3" {
		t.Fatalf("unexpected ids: %v", ids)
	}
}

func TestCollectPanelNodesSingle(t *testing.T) {
	g := DockPanelGroup("only", []string{"p"}, "p")
	nodes := DockTreeCollectPanelNodes(g)
	if len(nodes) != 1 || nodes[0].ID != "only" {
		t.Fatal("expected single node")
	}
}

// --- FindGroupByPanel ---

func TestFindGroupByPanel(t *testing.T) {
	root := makeTestTree()
	g, ok := DockTreeFindGroupByPanel(root, "D")
	if !ok || g.ID != "g3" {
		t.Fatal("expected g3")
	}
}

func TestFindGroupByPanelNotFound(t *testing.T) {
	root := makeTestTree()
	_, ok := DockTreeFindGroupByPanel(root, "Z")
	if ok {
		t.Fatal("should not find Z")
	}
}

func TestFindGroupByPanelFirst(t *testing.T) {
	root := makeTestTree()
	g, ok := DockTreeFindGroupByPanel(root, "A")
	if !ok || g.ID != "g1" {
		t.Fatal("expected g1")
	}
}

// --- FindGroupByID ---

func TestFindGroupByID(t *testing.T) {
	root := makeTestTree()
	g, ok := DockTreeFindGroupByID(root, "g2")
	if !ok {
		t.Fatal("not found")
	}
	if g.PanelIDs[0] != "C" {
		t.Fatal("wrong group")
	}
}

func TestFindGroupByIDNotFound(t *testing.T) {
	root := makeTestTree()
	_, ok := DockTreeFindGroupByID(root, "missing")
	if ok {
		t.Fatal("should not find")
	}
}

// --- RemovePanel ---

func TestRemovePanelFromMulti(t *testing.T) {
	root := makeTestTree()
	newRoot := DockTreeRemovePanel(root, "B")
	g, ok := DockTreeFindGroupByID(newRoot, "g1")
	if !ok {
		t.Fatal("g1 missing")
	}
	if len(g.PanelIDs) != 1 || g.PanelIDs[0] != "A" {
		t.Fatal("B not removed")
	}
}

func TestRemovePanelCollapsesEmptyGroup(t *testing.T) {
	root := makeTestTree()
	newRoot := DockTreeRemovePanel(root, "C")
	// g2 had only C, parent split s2 should collapse to g3
	_, ok := DockTreeFindGroupByID(newRoot, "g2")
	if ok {
		t.Fatal("g2 should be gone")
	}
	// g3 should still exist
	g3, ok := DockTreeFindGroupByID(newRoot, "g3")
	if !ok {
		t.Fatal("g3 missing")
	}
	if len(g3.PanelIDs) != 2 {
		t.Fatal("g3 should still have D, E")
	}
}

func TestRemovePanelUpdatesSelected(t *testing.T) {
	root := makeTestTree()
	// Select A, then remove A — should select B
	newRoot := DockTreeRemovePanel(root, "A")
	g, ok := DockTreeFindGroupByID(newRoot, "g1")
	if !ok {
		t.Fatal("g1 missing")
	}
	if g.SelectedID != "B" {
		t.Fatalf("selected = %s, want B", g.SelectedID)
	}
}

func TestRemovePanelNotFound(t *testing.T) {
	root := makeTestTree()
	newRoot := DockTreeRemovePanel(root, "Z")
	if newRoot != root {
		t.Fatal("should return same root")
	}
}

func TestRemovePanelNilChildren(t *testing.T) {
	nd := &DockNode{Kind: DockNodeSplit, ID: "s", First: nil, Second: nil}
	result := DockTreeRemovePanel(nd, "x")
	if result != nd {
		t.Fatal("should return same node")
	}
}

// --- AddTab ---

func TestAddTab(t *testing.T) {
	root := makeTestTree()
	newRoot := DockTreeAddTab(root, "g2", "F")
	g, ok := DockTreeFindGroupByID(newRoot, "g2")
	if !ok {
		t.Fatal("g2 missing")
	}
	if len(g.PanelIDs) != 2 || g.PanelIDs[1] != "F" {
		t.Fatal("F not added")
	}
	if g.SelectedID != "F" {
		t.Fatal("F should be selected")
	}
}

func TestAddTabWrongGroup(t *testing.T) {
	root := makeTestTree()
	newRoot := DockTreeAddTab(root, "nonexistent", "F")
	if newRoot != root {
		t.Fatal("should return same root")
	}
}

// --- SplitAt ---

func TestSplitAtLeft(t *testing.T) {
	root := makeTestTree()
	newRoot := DockTreeSplitAt(root, "g2", "F", DockDropLeft)
	// g2 should now be inside a split
	_, ok := DockTreeFindGroupByPanel(newRoot, "F")
	if !ok {
		t.Fatal("F not found")
	}
	_, ok = DockTreeFindGroupByPanel(newRoot, "C")
	if !ok {
		t.Fatal("C not found")
	}
}

func TestSplitAtBottom(t *testing.T) {
	root := makeTestTree()
	newRoot := DockTreeSplitAt(root, "g1", "F", DockDropBottom)
	fGroup, ok := DockTreeFindGroupByPanel(newRoot, "F")
	if !ok {
		t.Fatal("F not found")
	}
	if fGroup.PanelIDs[0] != "F" {
		t.Fatal("wrong group for F")
	}
}

func TestSplitAtWrongGroup(t *testing.T) {
	root := makeTestTree()
	newRoot := DockTreeSplitAt(root, "missing", "F", DockDropLeft)
	if newRoot != root {
		t.Fatal("should return same root")
	}
}

// --- WrapRoot ---

func TestWrapRootLeft(t *testing.T) {
	root := makeTestTree()
	newRoot := DockTreeWrapRoot(root, "F", DockDropWindowLeft)
	if newRoot.Kind != DockNodeSplit {
		t.Fatal("expected split")
	}
	if newRoot.Dir != DockSplitHorizontal {
		t.Fatal("expected horizontal")
	}
	if newRoot.Ratio != 0.2 {
		t.Fatalf("ratio = %f, want 0.2", newRoot.Ratio)
	}
	// First should be new group with F
	if newRoot.First.Kind != DockNodePanelGroup {
		t.Fatal("first should be panel group")
	}
	if newRoot.First.PanelIDs[0] != "F" {
		t.Fatal("first should contain F")
	}
}

func TestWrapRootBottom(t *testing.T) {
	root := makeTestTree()
	newRoot := DockTreeWrapRoot(root, "F", DockDropWindowBottom)
	if newRoot.Kind != DockNodeSplit {
		t.Fatal("expected split")
	}
	if newRoot.Dir != DockSplitVertical {
		t.Fatal("expected vertical")
	}
	if newRoot.Ratio != 0.8 {
		t.Fatalf("ratio = %f, want 0.8", newRoot.Ratio)
	}
	if newRoot.Second.PanelIDs[0] != "F" {
		t.Fatal("second should contain F")
	}
}

// --- MovePanel ---

func TestMovePanelCenter(t *testing.T) {
	root := makeTestTree()
	newRoot := DockTreeMovePanel(root, "A", "g2", DockDropCenter)
	g2, ok := DockTreeFindGroupByID(newRoot, "g2")
	if !ok {
		t.Fatal("g2 missing")
	}
	if !slices.Contains(g2.PanelIDs, "A") {
		t.Fatal("A not in g2")
	}
}

func TestMovePanelEdge(t *testing.T) {
	root := makeTestTree()
	newRoot := DockTreeMovePanel(root, "A", "g2", DockDropRight)
	_, ok := DockTreeFindGroupByPanel(newRoot, "A")
	if !ok {
		t.Fatal("A missing")
	}
}

func TestMovePanelWindowEdge(t *testing.T) {
	root := makeTestTree()
	newRoot := DockTreeMovePanel(root, "D", "", DockDropWindowTop)
	dg, ok := DockTreeFindGroupByPanel(newRoot, "D")
	if !ok {
		t.Fatal("D missing")
	}
	if dg.ID != "dock_edge_D" {
		t.Fatalf("expected dock_edge_D, got %s", dg.ID)
	}
}

// --- SelectPanel ---

func TestSelectPanel(t *testing.T) {
	root := makeTestTree()
	newRoot := DockTreeSelectPanel(root, "g1", "B")
	g, ok := DockTreeFindGroupByID(newRoot, "g1")
	if !ok {
		t.Fatal("g1 missing")
	}
	if g.SelectedID != "B" {
		t.Fatalf("selected = %s, want B", g.SelectedID)
	}
}

func TestSelectPanelNoOp(t *testing.T) {
	root := makeTestTree()
	newRoot := DockTreeSelectPanel(root, "g1", "A")
	if newRoot != root {
		t.Fatal("should return same root when already selected")
	}
}

func TestSelectPanelWrongGroup(t *testing.T) {
	root := makeTestTree()
	newRoot := DockTreeSelectPanel(root, "missing", "A")
	if newRoot != root {
		t.Fatal("should return same root")
	}
}

// --- dockZoneToSplitDir ---

func TestZoneToSplitDirHorizontal(t *testing.T) {
	for _, z := range []DockDropZone{DockDropLeft, DockDropRight, DockDropWindowLeft, DockDropWindowRight} {
		if dockZoneToSplitDir(z) != DockSplitHorizontal {
			t.Fatalf("zone %d should map to horizontal", z)
		}
	}
}

func TestZoneToSplitDirVertical(t *testing.T) {
	for _, z := range []DockDropZone{DockDropTop, DockDropBottom, DockDropWindowTop, DockDropWindowBottom, DockDropCenter} {
		if dockZoneToSplitDir(z) != DockSplitVertical {
			t.Fatalf("zone %d should map to vertical", z)
		}
	}
}

// --- UpdateRatio ---

func TestUpdateRatio(t *testing.T) {
	root := makeTestTree()
	newRoot := dockTreeUpdateRatio(root, "s2", 0.75)
	// s2 should have new ratio
	if newRoot.Second.Ratio != 0.75 {
		t.Fatalf("ratio = %f, want 0.75", newRoot.Second.Ratio)
	}
	// s1 should keep original ratio
	if newRoot.Ratio != 0.3 {
		t.Fatalf("s1 ratio = %f, want 0.3", newRoot.Ratio)
	}
}

func TestUpdateRatioRoot(t *testing.T) {
	root := makeTestTree()
	newRoot := dockTreeUpdateRatio(root, "s1", 0.5)
	if newRoot.Ratio != 0.5 {
		t.Fatalf("ratio = %f, want 0.5", newRoot.Ratio)
	}
}

func TestUpdateRatioMissing(t *testing.T) {
	root := makeTestTree()
	newRoot := dockTreeUpdateRatio(root, "missing", 0.9)
	if newRoot != root {
		t.Fatal("should return same root")
	}
}

// --- dockTreeIsEmpty ---

func TestDockTreeIsEmpty(t *testing.T) {
	empty := DockPanelGroup("e", nil, "")
	if !dockTreeIsEmpty(empty) {
		t.Fatal("should be empty")
	}
	full := DockPanelGroup("f", []string{"x"}, "x")
	if dockTreeIsEmpty(full) {
		t.Fatal("should not be empty")
	}
	split := DockSplit("s", DockSplitHorizontal, 0.5, empty, full)
	if dockTreeIsEmpty(split) {
		t.Fatal("split should not be empty")
	}
}

// --- Edge cases ---

func TestRemoveAllPanelsCollapses(t *testing.T) {
	// Two groups, remove all from one
	g1 := DockPanelGroup("g1", []string{"A"}, "A")
	g2 := DockPanelGroup("g2", []string{"B"}, "B")
	root := DockSplit("s", DockSplitHorizontal, 0.5, g1, g2)

	newRoot := DockTreeRemovePanel(root, "A")
	// Should collapse to just g2
	if newRoot.Kind != DockNodePanelGroup {
		t.Fatal("should collapse to panel group")
	}
	if newRoot.ID != "g2" {
		t.Fatalf("expected g2, got %s", newRoot.ID)
	}
}

func TestSplitAtDirections(t *testing.T) {
	g := DockPanelGroup("g", []string{"A"}, "A")

	for _, tc := range []struct {
		zone DockDropZone
		dir  DockSplitDir
	}{
		{DockDropLeft, DockSplitHorizontal},
		{DockDropRight, DockSplitHorizontal},
		{DockDropTop, DockSplitVertical},
		{DockDropBottom, DockSplitVertical},
	} {
		result := DockTreeSplitAt(g, "g", "F", tc.zone)
		if result.Kind != DockNodeSplit {
			t.Fatalf("zone %d: expected split", tc.zone)
		}
		if result.Dir != tc.dir {
			t.Fatalf("zone %d: dir = %d, want %d", tc.zone, result.Dir, tc.dir)
		}
	}
}

func TestMovePanelPreservesOtherPanels(t *testing.T) {
	root := makeTestTree()
	newRoot := DockTreeMovePanel(root, "B", "g3", DockDropCenter)

	// A should still be in g1
	g1, ok := DockTreeFindGroupByID(newRoot, "g1")
	if !ok {
		t.Fatal("g1 missing")
	}
	if len(g1.PanelIDs) != 1 || g1.PanelIDs[0] != "A" {
		t.Fatal("g1 should only have A")
	}

	// B should be in g3
	g3, ok := DockTreeFindGroupByID(newRoot, "g3")
	if !ok {
		t.Fatal("g3 missing")
	}
	if !slices.Contains(g3.PanelIDs, "B") {
		t.Fatal("B not in g3")
	}
}
