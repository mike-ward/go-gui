package gui

// GridSortDir specifies ascending or descending sort order.
type GridSortDir uint8

// GridSortDir constants.
const (
	GridSortAsc GridSortDir = iota
	GridSortDesc
)

// GridPaginationKind selects cursor- or offset-based paging.
type GridPaginationKind uint8

// GridPaginationKind constants.
const (
	GridPaginationCursor GridPaginationKind = iota
	GridPaginationOffset
)

// GridMutationKind identifies a create, update, or delete.
type GridMutationKind uint8

// GridMutationKind constants.
const (
	GridMutationCreate GridMutationKind = iota
	GridMutationUpdate
	GridMutationDelete
)

// GridSort describes a single sort criterion.
type GridSort struct {
	ColID string
	Dir   GridSortDir
}

// GridFilter describes a single column filter.
type GridFilter struct {
	ColID string
	Op    string // "contains", "equals", "starts_with", "ends_with"
	Value string
}

// GridQueryState holds the active sorts, filters, and quick
// filter for a data grid query.
type GridQueryState struct {
	Sorts       []GridSort
	Filters     []GridFilter
	QuickFilter string
}

// GridSelection tracks the selected rows in a data grid.
type GridSelection struct {
	AnchorRowID    string
	ActiveRowID    string
	SelectedRowIDs map[string]bool
}

// GridRow represents a single data row with an ID and
// column-keyed cell values.
type GridRow struct {
	ID    string
	Cells map[string]string
}

// GridCellEdit describes a single cell edit operation.
type GridCellEdit struct {
	RowID  string
	RowIdx int
	ColID  string
	Value  string
}
