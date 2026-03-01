package gui

import "testing"

func TestListCoreFilterEmpty(t *testing.T) {
	items := []ListCoreItem{
		{ID: "a", Label: "Alpha"},
		{ID: "b", Label: "Beta"},
	}
	result := listCoreFilter(items, "")
	if len(result) != 2 {
		t.Errorf("empty query: got %d indices, want 2", len(result))
	}
}

func TestListCoreFilterMatch(t *testing.T) {
	items := []ListCoreItem{
		{ID: "a", Label: "Alpha"},
		{ID: "b", Label: "Beta"},
		{ID: "c", Label: "Gamma"},
	}
	result := listCoreFilter(items, "al")
	if len(result) != 1 || result[0] != 0 {
		t.Errorf("filter 'al': got %v", result)
	}
}

func TestListCoreFilterSkipsSubheadings(t *testing.T) {
	items := []ListCoreItem{
		{ID: "h", Label: "Header", IsSubheading: true},
		{ID: "a", Label: "Alpha"},
	}
	result := listCoreFilter(items, "hea")
	// Subheading should be skipped; "Alpha" doesn't match "hea".
	if len(result) != 0 {
		t.Errorf("expected 0 matches, got %v", result)
	}
}

func TestListCorePrepare(t *testing.T) {
	items := []ListCoreItem{
		{ID: "x", Label: "Exit"},
		{ID: "s", Label: "Save"},
	}
	p := listCorePrepare(items, "", 5)
	if len(p.Items) != 2 {
		t.Errorf("items = %d", len(p.Items))
	}
	if p.HL != 1 {
		t.Errorf("hl = %d, want 1 (clamped from 5)", p.HL)
	}
	if len(p.IDs) != 2 {
		t.Errorf("ids = %d", len(p.IDs))
	}
}

func TestListCorePrepareFiltered(t *testing.T) {
	items := []ListCoreItem{
		{ID: "a", Label: "Apple"},
		{ID: "b", Label: "Banana"},
		{ID: "c", Label: "Cherry"},
	}
	p := listCorePrepare(items, "an", 0)
	if len(p.Items) != 1 || p.Items[0].ID != "b" {
		t.Errorf("filtered items = %v", p.Items)
	}
}

func TestListCoreViews(t *testing.T) {
	items := []ListCoreItem{
		{ID: "a", Label: "Alpha"},
		{ID: "b", Label: "Beta"},
		{ID: "c", Label: "Gamma"},
	}
	cfg := ListCoreCfg{
		TextStyle:      DefaultTextStyle,
		ColorHighlight: Red,
		ColorHover:     Blue,
		PaddingItem:    PaddingSmall,
	}
	views := listCoreViews(items, cfg, 0, 2, 1, nil, 20)
	if len(views) != 3 {
		t.Errorf("views = %d, want 3", len(views))
	}
}

func TestListCoreViewsWithSpacers(t *testing.T) {
	items := make([]ListCoreItem, 20)
	for i := range items {
		items[i] = ListCoreItem{ID: string(rune('a' + i)), Label: "Item"}
	}
	cfg := ListCoreCfg{
		TextStyle:   DefaultTextStyle,
		PaddingItem: PaddingSmall,
	}
	views := listCoreViews(items, cfg, 5, 10, 5, nil, 20)
	// Should have top spacer + 6 items + bottom spacer.
	if len(views) != 8 {
		t.Errorf("views = %d, want 8", len(views))
	}
}

func TestListCoreSubheadingView(t *testing.T) {
	item := ListCoreItem{ID: "h", Label: "Group A", IsSubheading: true}
	cfg := ListCoreCfg{
		TextStyle:       DefaultTextStyle,
		SubheadingStyle: DefaultTextStyle,
		PaddingItem:     PaddingSmall,
	}
	views := listCoreViews([]ListCoreItem{item}, cfg, 0, 0, -1, nil, 20)
	if len(views) != 1 {
		t.Fatalf("views = %d", len(views))
	}
}
