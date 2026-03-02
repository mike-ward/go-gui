package gui

import "testing"

// --- dataGridJumpDigits ---

func TestJumpDigitsExtractsOnlyDigits(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"123", "123"},
		{"abc", ""},
		{"a1b2c3", "123"},
		{"", ""},
		{"  42  ", "42"},
		{"row#7!", "7"},
	}
	for _, tt := range tests {
		got := dataGridJumpDigits(tt.input)
		if got != tt.want {
			t.Errorf("dataGridJumpDigits(%q) = %q, want %q",
				tt.input, got, tt.want)
		}
	}
}

// --- dataGridParseJumpTarget ---

func TestParseJumpTargetValid(t *testing.T) {
	// "5" with 10 rows -> index 4 (1-based to 0-based)
	got, ok := dataGridParseJumpTarget("5", 10)
	if !ok || got != 4 {
		t.Fatalf("got (%d, %v), want (4, true)", got, ok)
	}
}

func TestParseJumpTargetOutOfRange(t *testing.T) {
	// "20" with 10 rows -> clamped to index 9
	got, ok := dataGridParseJumpTarget("20", 10)
	if !ok || got != 9 {
		t.Fatalf("got (%d, %v), want (9, true)", got, ok)
	}
}

func TestParseJumpTargetEmpty(t *testing.T) {
	_, ok := dataGridParseJumpTarget("", 10)
	if ok {
		t.Fatal("expected false for empty input")
	}
}

func TestParseJumpTargetNonNumeric(t *testing.T) {
	_, ok := dataGridParseJumpTarget("abc", 10)
	if ok {
		t.Fatal("expected false for non-numeric input")
	}
}

func TestParseJumpTargetZeroTotalRows(t *testing.T) {
	_, ok := dataGridParseJumpTarget("1", 0)
	if ok {
		t.Fatal("expected false when totalRows <= 0")
	}
}

func TestParseJumpTargetFirstRow(t *testing.T) {
	got, ok := dataGridParseJumpTarget("1", 5)
	if !ok || got != 0 {
		t.Fatalf("got (%d, %v), want (0, true)", got, ok)
	}
}

// --- dataGridPageBounds (local page count calculation) ---

func TestPageBoundsPageCount(t *testing.T) {
	tests := []struct {
		totalRows int
		pageSize  int
		wantCount int
	}{
		{100, 25, 4},
		{101, 25, 5}, // partial last page
		{0, 25, 1},   // no rows -> 1 page
		{10, 0, 1},   // no paging -> 1 page
		{1, 1, 1},    // single row, single page
		{50, 50, 1},  // exact fit
		{50, 100, 1}, // pageSize > totalRows
	}
	for _, tt := range tests {
		_, _, _, pageCount := dataGridPageBounds(tt.totalRows, tt.pageSize, 0)
		if pageCount != tt.wantCount {
			t.Errorf("dataGridPageBounds(%d, %d, 0) pageCount = %d, want %d",
				tt.totalRows, tt.pageSize, pageCount, tt.wantCount)
		}
	}
}

func TestPageBoundsStartEnd(t *testing.T) {
	// Page 1 of 4 (0-indexed page 1) with 100 rows, pageSize 25
	start, end, pageIdx, _ := dataGridPageBounds(100, 25, 1)
	if start != 25 || end != 50 || pageIdx != 1 {
		t.Fatalf("got start=%d end=%d pageIdx=%d, want 25/50/1",
			start, end, pageIdx)
	}
}

func TestPageBoundsClampsPageIndex(t *testing.T) {
	// Requested page 99 with only 4 pages -> clamped to 3
	_, _, pageIdx, _ := dataGridPageBounds(100, 25, 99)
	if pageIdx != 3 {
		t.Fatalf("got pageIdx=%d, want 3", pageIdx)
	}
}

// --- dataGridPagerRowsText (page bounds text) ---

func TestPagerRowsTextNormal(t *testing.T) {
	got := dataGridPagerRowsText(0, 25, 100)
	want := "Rows 1-25/100"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestPagerRowsTextZeroTotal(t *testing.T) {
	got := dataGridPagerRowsText(0, 0, 0)
	want := "Rows 0/0"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

// --- dataGridQuickFilterMatchesText ---

func TestQuickFilterMatchesTextLocalOnly(t *testing.T) {
	cfg := &DataGridCfg{
		Rows: []GridRow{
			{ID: "r1", Cells: map[string]string{"a": "1"}},
			{ID: "r2", Cells: map[string]string{"a": "2"}},
			{ID: "r3", Cells: map[string]string{"a": "3"}},
		},
	}
	got := dataGridQuickFilterMatchesText(cfg)
	want := "Matches 3"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestQuickFilterMatchesTextWithRowCount(t *testing.T) {
	total := 200
	cfg := &DataGridCfg{
		Rows:     []GridRow{{ID: "r1"}, {ID: "r2"}},
		RowCount: &total,
	}
	got := dataGridQuickFilterMatchesText(cfg)
	want := "Matches 2/200"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

// --- dataGridSourceRowMatchesQuery (quick filter matching) ---

func TestSourceRowMatchesQueryCaseInsensitive(t *testing.T) {
	row := GridRow{
		ID:    "r1",
		Cells: map[string]string{"name": "Alice", "city": "Boston"},
	}
	if !dataGridSourceRowMatchesQuery(row, "alice", nil) {
		t.Fatal("expected match for case-insensitive 'alice'")
	}
	if !dataGridSourceRowMatchesQuery(row, "bos", nil) {
		t.Fatal("expected match for partial 'bos'")
	}
	if dataGridSourceRowMatchesQuery(row, "xyz", nil) {
		t.Fatal("expected no match for 'xyz'")
	}
}

func TestSourceRowMatchesQueryEmptyNeedle(t *testing.T) {
	row := GridRow{
		ID:    "r1",
		Cells: map[string]string{"a": "value"},
	}
	if !dataGridSourceRowMatchesQuery(row, "", nil) {
		t.Fatal("empty needle should match all rows")
	}
}

func TestSourceRowMatchesQueryMultipleCells(t *testing.T) {
	row := GridRow{
		ID: "r1",
		Cells: map[string]string{
			"first": "John",
			"last":  "Doe",
			"email": "john@example.com",
		},
	}
	// Match in any cell
	if !dataGridSourceRowMatchesQuery(row, "example", nil) {
		t.Fatal("expected match in email cell")
	}
	if !dataGridSourceRowMatchesQuery(row, "doe", nil) {
		t.Fatal("expected match in last cell")
	}
}

// --- dataGridJumpEnabledLocal ---

func TestJumpEnabledLocal(t *testing.T) {
	sel := func(GridSelection, *Event, *Window) {}
	page := func(int, *Event, *Window) {}

	tests := []struct {
		name       string
		rowsLen    int
		onSel      func(GridSelection, *Event, *Window)
		onPage     func(int, *Event, *Window)
		pageSize   int
		totalRows  int
		want       bool
	}{
		{"enabled", 10, sel, page, 5, 10, true},
		{"no rows", 0, sel, page, 5, 10, false},
		{"zero total", 10, sel, page, 5, 0, false},
		{"no selection cb", 10, nil, page, 5, 10, false},
		{"paged no page cb", 10, sel, nil, 5, 10, false},
		{"no paging ok", 10, sel, nil, 0, 10, true},
	}
	for _, tt := range tests {
		got := dataGridJumpEnabledLocal(tt.rowsLen, tt.onSel, tt.onPage,
			tt.pageSize, tt.totalRows)
		if got != tt.want {
			t.Errorf("%s: got %v, want %v", tt.name, got, tt.want)
		}
	}
}

// --- dataGridNextPageIndexForKey ---

func TestNextPageIndexForKeyCtrlPageDown(t *testing.T) {
	e := &Event{KeyCode: KeyPageDown, Modifiers: ModCtrl}
	got, ok := dataGridNextPageIndexForKey(1, 5, e)
	if !ok || got != 2 {
		t.Fatalf("got (%d, %v), want (2, true)", got, ok)
	}
}

func TestNextPageIndexForKeyAltHome(t *testing.T) {
	e := &Event{KeyCode: KeyHome, Modifiers: ModAlt}
	got, ok := dataGridNextPageIndexForKey(3, 5, e)
	if !ok || got != 0 {
		t.Fatalf("got (%d, %v), want (0, true)", got, ok)
	}
}

func TestNextPageIndexForKeySinglePage(t *testing.T) {
	e := &Event{KeyCode: KeyPageDown, Modifiers: ModCtrl}
	_, ok := dataGridNextPageIndexForKey(0, 1, e)
	if ok {
		t.Fatal("expected false for single page")
	}
}

// --- dataGridCharIsCopy / dataGridIsSelectAllShortcut ---

func TestCharIsCopy(t *testing.T) {
	e := &Event{CharCode: 3, Modifiers: ModCtrl}
	if !dataGridCharIsCopy(e) {
		t.Fatal("Ctrl+C should be copy")
	}
	e2 := &Event{CharCode: 3, Modifiers: ModSuper}
	if !dataGridCharIsCopy(e2) {
		t.Fatal("Cmd+C should be copy")
	}
	e3 := &Event{CharCode: 3}
	if dataGridCharIsCopy(e3) {
		t.Fatal("bare charCode=3 should not be copy")
	}
}

func TestIsSelectAllShortcut(t *testing.T) {
	e := &Event{KeyCode: KeyA, Modifiers: ModCtrl}
	if !dataGridIsSelectAllShortcut(e) {
		t.Fatal("Ctrl+A should be select-all")
	}
	e2 := &Event{KeyCode: KeyA, Modifiers: ModSuper}
	if !dataGridIsSelectAllShortcut(e2) {
		t.Fatal("Cmd+A should be select-all")
	}
	e3 := &Event{KeyCode: KeyA}
	if dataGridIsSelectAllShortcut(e3) {
		t.Fatal("bare 'A' should not be select-all")
	}
}

// --- dataGridRangeSelectedRows ---

func TestRangeSelectedRows(t *testing.T) {
	rows := []GridRow{
		{ID: "a"}, {ID: "b"}, {ID: "c"}, {ID: "d"},
	}
	got := dataGridRangeSelectedRows(rows, 1, 2, "b")
	if len(got) != 2 {
		t.Fatalf("got %d selected, want 2", len(got))
	}
	if !got["b"] || !got["c"] {
		t.Fatalf("expected b and c selected, got %v", got)
	}
}

func TestRangeSelectedRowsFallback(t *testing.T) {
	rows := []GridRow{{ID: "a"}}
	got := dataGridRangeSelectedRows(rows, -1, -1, "x")
	if len(got) != 1 || !got["x"] {
		t.Fatalf("expected fallback to target, got %v", got)
	}
}
