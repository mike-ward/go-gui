package gui

import "testing"

func testColumns() []GridOrmColumnSpec {
	return []GridOrmColumnSpec{
		{ID: "name", DBField: "name", QuickFilter: true,
			Filterable: true, Sortable: true,
			CaseInsensitive: true},
		{ID: "email", DBField: "email", QuickFilter: true,
			Filterable: true, Sortable: true,
			CaseInsensitive: true},
		{ID: "status", DBField: "status", QuickFilter: false,
			Filterable: true, Sortable: true,
			CaseInsensitive: false,
			AllowedOps: []string{"equals"}},
	}
}

func testOrmSource(t *testing.T) *GridOrmDataSource {
	t.Helper()
	src, err := NewGridOrmDataSource(GridOrmDataSource{
		Columns:      testColumns(),
		DefaultLimit: 50,
		FetchFn: func(
			spec GridOrmQuerySpec,
			signal *GridAbortSignal,
		) (GridOrmPage, error) {
			return GridOrmPage{
				Rows:    []GridRow{{ID: "1"}},
				HasMore: false,
			}, nil
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	return src
}

func TestNewGridOrmDataSourceValidation(t *testing.T) {
	_, err := NewGridOrmDataSource(GridOrmDataSource{
		Columns: []GridOrmColumnSpec{
			{ID: "", DBField: "x"},
		},
		FetchFn: func(GridOrmQuerySpec, *GridAbortSignal) (GridOrmPage, error) {
			return GridOrmPage{}, nil
		},
	})
	if err == nil {
		t.Fatal("expected error for empty column id")
	}
}

func TestNewGridOrmDataSourceDuplicateColumn(t *testing.T) {
	_, err := NewGridOrmDataSource(GridOrmDataSource{
		Columns: []GridOrmColumnSpec{
			{ID: "x", DBField: "x", Filterable: true, Sortable: true},
			{ID: "x", DBField: "y", Filterable: true, Sortable: true},
		},
		FetchFn: func(GridOrmQuerySpec, *GridAbortSignal) (GridOrmPage, error) {
			return GridOrmPage{}, nil
		},
	})
	if err == nil {
		t.Fatal("expected error for duplicate column id")
	}
}

func TestNewGridOrmDataSourceInvalidDBField(t *testing.T) {
	_, err := NewGridOrmDataSource(GridOrmDataSource{
		Columns: []GridOrmColumnSpec{
			{ID: "x", DBField: "1bad"},
		},
		FetchFn: func(GridOrmQuerySpec, *GridAbortSignal) (GridOrmPage, error) {
			return GridOrmPage{}, nil
		},
	})
	if err == nil {
		t.Fatal("expected error for invalid db_field")
	}
}

func TestNewGridOrmDataSourceNoFetchFn(t *testing.T) {
	_, err := NewGridOrmDataSource(GridOrmDataSource{
		Columns: testColumns(),
	})
	if err == nil {
		t.Fatal("expected error for nil FetchFn")
	}
}

func TestGridOrmCapabilities(t *testing.T) {
	src := testOrmSource(t)
	caps := src.Capabilities()
	if !caps.SupportsCursorPagination {
		t.Fatal("cursor should be supported")
	}
	if caps.SupportsCreate {
		t.Fatal("create should not be supported")
	}
}

func TestGridOrmFetchData(t *testing.T) {
	src := testOrmSource(t)
	res, err := src.FetchData(GridDataRequest{
		Page: GridCursorPageReq{Limit: 10},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Rows) != 1 {
		t.Fatalf("rows = %d, want 1", len(res.Rows))
	}
}

func TestGridOrmValidateQuery(t *testing.T) {
	q, err := GridOrmValidateQuery(GridQueryState{
		Sorts: []GridSort{
			{ColID: "name", Dir: GridSortAsc},
			{ColID: "unknown"},
		},
		Filters: []GridFilter{
			{ColID: "name", Op: "contains", Value: "test"},
			{ColID: "status", Op: "equals", Value: "a"},
			{ColID: "status", Op: "contains", Value: "b"},
		},
	}, testColumns())
	if err != nil {
		t.Fatal(err)
	}
	// Unknown sort dropped.
	if len(q.Sorts) != 1 {
		t.Fatalf("sorts = %d, want 1", len(q.Sorts))
	}
	// status only allows "equals", "contains" is dropped.
	if len(q.Filters) != 2 {
		t.Fatalf("filters = %d, want 2", len(q.Filters))
	}
}

func TestGridOrmValidateQueryQuickFilterTooLong(t *testing.T) {
	long := make([]byte, gridOrmMaxFilterValueLen+1)
	for i := range long {
		long[i] = 'a'
	}
	_, err := GridOrmValidateQuery(GridQueryState{
		QuickFilter: string(long),
	}, testColumns())
	if err == nil {
		t.Fatal("expected error for long quick_filter")
	}
}

func TestGridOrmBuildSQL(t *testing.T) {
	src := testOrmSource(t)
	b, err := src.BuildSQL(GridOrmQuerySpec{
		QuickFilter: "test",
		Sorts: []GridSort{
			{ColID: "name", Dir: GridSortDesc},
		},
		Filters: []GridFilter{
			{ColID: "status", Op: "equals", Value: "active"},
		},
		Limit:  25,
		Offset: 0,
	})
	if err != nil {
		t.Fatal(err)
	}
	if b.WhereSQL == "" {
		t.Fatal("WhereSQL should not be empty")
	}
	if b.OrderSQL == "" {
		t.Fatal("OrderSQL should not be empty")
	}
	if len(b.Params) == 0 {
		t.Fatal("Params should not be empty")
	}
}

func TestGridOrmEscapeLike(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"hello", "hello"},
		{"100%", `100\%`},
		{"a_b", `a\_b`},
		{`a\b`, `a\\b`},
		{`%_\`, `\%\_\\`},
	}
	for _, tt := range tests {
		got := GridOrmEscapeLike(tt.input)
		if got != tt.want {
			t.Errorf("EscapeLike(%q) = %q, want %q",
				tt.input, got, tt.want)
		}
	}
}

func TestGridOrmValidDBField(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"name", true},
		{"_id", true},
		{"t.name", true},
		{"", false},
		{"1bad", false},
		{"a..b", false},
		{"a.", false},
		{"a-b", false},
		{"a b", false},
		{"T1", true},
		{"users.email_addr", true},
	}
	for _, tt := range tests {
		got := GridOrmValidDBField(tt.input)
		if got != tt.want {
			t.Errorf("ValidDBField(%q) = %v, want %v",
				tt.input, got, tt.want)
		}
	}
}

func TestGridOrmMutateMissingFn(t *testing.T) {
	src := testOrmSource(t)
	_, err := src.MutateData(GridMutationRequest{
		Kind: GridMutationCreate,
		Rows: []GridRow{{ID: "x"}},
	})
	if err == nil {
		t.Fatal("expected error for nil CreateFn")
	}
}

func TestGridOrmMutateUnknownColumn(t *testing.T) {
	src, err := NewGridOrmDataSource(GridOrmDataSource{
		Columns: testColumns(),
		FetchFn: func(GridOrmQuerySpec, *GridAbortSignal) (GridOrmPage, error) {
			return GridOrmPage{}, nil
		},
		CreateFn: func(rows []GridRow, _ *GridAbortSignal) ([]GridRow, error) {
			return rows, nil
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	_, err = src.MutateData(GridMutationRequest{
		Kind: GridMutationCreate,
		Rows: []GridRow{
			{ID: "x", Cells: map[string]string{"bogus": "v"}},
		},
	})
	if err == nil {
		t.Fatal("expected error for unknown column")
	}
}

func TestGridOrmDeleteMany(t *testing.T) {
	src, err := NewGridOrmDataSource(GridOrmDataSource{
		Columns: testColumns(),
		FetchFn: func(GridOrmQuerySpec, *GridAbortSignal) (GridOrmPage, error) {
			return GridOrmPage{}, nil
		},
		DeleteManyFn: func(ids []string, _ *GridAbortSignal) ([]string, error) {
			return ids, nil
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	res, err := src.MutateData(GridMutationRequest{
		Kind:   GridMutationDelete,
		RowIDs: []string{"a", "b"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(res.DeletedIDs) != 2 {
		t.Fatalf("deleted = %d, want 2", len(res.DeletedIDs))
	}
}

func TestGridOrmDeleteSingle(t *testing.T) {
	src, err := NewGridOrmDataSource(GridOrmDataSource{
		Columns: testColumns(),
		FetchFn: func(GridOrmQuerySpec, *GridAbortSignal) (GridOrmPage, error) {
			return GridOrmPage{}, nil
		},
		DeleteFn: func(id string, _ *GridAbortSignal) (string, error) {
			return id, nil
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	res, err := src.MutateData(GridMutationRequest{
		Kind:   GridMutationDelete,
		RowIDs: []string{"a"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(res.DeletedIDs) != 1 {
		t.Fatalf("deleted = %d, want 1", len(res.DeletedIDs))
	}
}

func TestGridOrmFilterDedup(t *testing.T) {
	q, err := GridOrmValidateQuery(GridQueryState{
		Filters: []GridFilter{
			{ColID: "name", Op: "contains", Value: "x"},
			{ColID: "name", Op: "contains", Value: "x"},
		},
	}, testColumns())
	if err != nil {
		t.Fatal(err)
	}
	if len(q.Filters) != 1 {
		t.Fatalf("filters = %d, want 1", len(q.Filters))
	}
}
