package gui

import (
	"time"
)

// Data grid constants.
const (
	dataGridVirtualBufferRows       = 2
	dataGridResizeDoubleClickFrames = uint64(24) // ~400ms at 60fps
	dataGridEditDoubleClickFrames   = uint64(36) // ~600ms at 60fps
	dataGridResizeHandleWidth       = float32(6)
	dataGridAutofitPadding          = float32(18)
	dataGridAutofitMaxRows          = 1000
	dataGridIndicatorAlpha          = uint8(140)
	dataGridHeaderControlWidth      = float32(12)
	dataGridHeaderReorderSpacing    = float32(1)
	dataGridHeaderLabelMinWidth     = float32(24)
	dataGridGroupIndentStep         = float32(14)
	dataGridDetailIndentGap         = float32(4)
	dataGridRecordSep               = "\x1e"
	dataGridUnitSep                 = "\x1f"
	dataGridGroupSep                = "\x1d"
	dataGridDefaultRowHeight        = float32(30)
	dataGridDefaultHeaderHeight     = float32(34)
	dataGridDefaultPageLimit        = 100
	dataGridJumpInputWidth          = float32(68)
)

// GridColumnPin specifies column pinning position.
type GridColumnPin uint8

// GridColumnPin values.
const (
	GridColumnPinNone GridColumnPin = iota
	GridColumnPinLeft
	GridColumnPinRight
)

// GridAggregateOp specifies the aggregation operation.
type GridAggregateOp uint8

// GridAggregateOp values.
const (
	GridAggregateCount GridAggregateOp = iota
	GridAggregateSum
	GridAggregateAvg
	GridAggregateMin
	GridAggregateMax
)

func (op GridAggregateOp) String() string {
	switch op {
	case GridAggregateCount:
		return "count"
	case GridAggregateSum:
		return "sum"
	case GridAggregateAvg:
		return "avg"
	case GridAggregateMin:
		return "min"
	case GridAggregateMax:
		return "max"
	}
	return "count"
}

// GridCellEditorKind specifies the type of inline cell editor.
type GridCellEditorKind uint8

// GridCellEditorKind values.
const (
	GridCellEditorText GridCellEditorKind = iota
	GridCellEditorSelect
	GridCellEditorDate
	GridCellEditorCheckbox
)

// dataGridDisplayRowKind classifies display rows.
type dataGridDisplayRowKind uint8

const (
	dataGridDisplayRowData dataGridDisplayRowKind = iota
	dataGridDisplayRowGroupHeader
	dataGridDisplayRowDetail
)

// GridColumnCfg configures a single data grid column.
type GridColumnCfg struct {
	ID               string
	Title            string
	Width            Opt[float32]
	MinWidth         Opt[float32]
	MaxWidth         Opt[float32]
	Resizable        bool
	Reorderable      bool
	Sortable         bool
	Filterable       bool
	Editable         bool
	Editor           GridCellEditorKind
	EditorOptions    []string
	EditorTrueValue  string
	EditorFalseValue string
	DefaultValue     string
	Pin              GridColumnPin
	Align            HorizontalAlign
	TextStyle        *TextStyle
}

// gridColumnCfgDefaults applies V-style defaults to a
// GridColumnCfg zero value. Called once per cfg construction.
func gridColumnCfgDefaults(c *GridColumnCfg) {
	if !c.Width.IsSet() {
		c.Width = SomeF(120)
	}
	if !c.MinWidth.IsSet() {
		c.MinWidth = SomeF(60)
	}
	if !c.MaxWidth.IsSet() {
		c.MaxWidth = SomeF(600)
	}
	if c.EditorTrueValue == "" {
		c.EditorTrueValue = "true"
	}
	if c.EditorFalseValue == "" {
		c.EditorFalseValue = "false"
	}
}

// GridAggregateCfg configures an aggregate operation.
type GridAggregateCfg struct {
	ColID string
	Op    GridAggregateOp
	Label string
}

// GridCsvData holds parsed CSV data.
type GridCsvData struct {
	Columns []GridColumnCfg
	Rows    []GridRow
}

// GridExportCfg configures export behavior.
type GridExportCfg struct {
	SanitizeSpreadsheetFormulas bool
	XLSXAutoType                bool
}

// GridCellFormat describes conditional cell formatting.
type GridCellFormat struct {
	HasBGColor   bool
	BGColor      Color
	HasTextColor bool
	TextColor    Color
}

// dataGridDisplayRow is a flat display entry (data, group
// header, or detail expansion).
type dataGridDisplayRow struct {
	Kind          dataGridDisplayRowKind
	DataRowIdx    int
	GroupColID    string
	GroupValue    string
	GroupColTitle string
	GroupDepth    int
	GroupCount    int
	AggregateText string
}

// dataGridPresentation is the flattened display row list
// with a data-row-index → display-index map.
type dataGridPresentation struct {
	Rows          []dataGridDisplayRow
	DataToDisplay map[int]int
}

// DataGridCfg configures a data grid widget.
type DataGridCfg struct {
	ID                     string
	IDFocus                uint32
	IDScroll               uint32
	Columns                []GridColumnCfg
	ColumnOrder            []string
	GroupBy                []string
	Aggregates             []GridAggregateCfg
	Rows                   []GridRow
	DataSource             DataGridDataSource
	PaginationKind         GridPaginationKind
	Cursor                 string
	PageLimit              int
	RowCount               *int
	Loading                bool
	LoadError              string
	ShowCRUDToolbar        bool
	AllowCreate            *bool
	AllowDelete            *bool
	Query                  GridQueryState
	Selection              GridSelection
	MultiSort              *bool
	MultiSelect            *bool
	RangeSelect            *bool
	ShowHeader             *bool
	FreezeHeader           bool
	ShowFilterRow          bool
	ShowQuickFilter        bool
	ShowColumnChooser      bool
	ShowGroupCounts        *bool
	PageSize               int
	PageIndex              int
	HiddenColumnIDs        map[string]bool
	FrozenTopRowIDs        []string
	DetailExpandedRowIDs   map[string]bool
	QuickFilterPlaceholder string
	QuickFilterDebounce    time.Duration
	RowHeight              float32
	HeaderHeight           float32
	ColorBackground        Color
	ColorHeader            Color
	ColorHeaderHover       Color
	ColorFilter            Color
	ColorQuickFilter       Color
	ColorRowHover          Color
	ColorRowAlt            Color
	ColorRowSelected       Color
	ColorBorder            Color
	ColorResizeHandle      Color
	ColorResizeActive      Color
	PaddingCell            Opt[Padding]
	PaddingHeader          Opt[Padding]
	PaddingFilter          Opt[Padding]
	TextStyle              TextStyle
	TextStyleHeader        TextStyle
	TextStyleFilter        TextStyle
	Radius                 Opt[float32]
	SizeBorder             Opt[float32]
	Scrollbar              ScrollbarOverflow
	Sizing                 Opt[Sizing]
	Width                  float32
	Height                 float32
	MinWidth               float32
	MaxWidth               float32
	MinHeight              float32
	MaxHeight              float32
	OnQueryChange          func(GridQueryState, *Event, *Window)
	OnSelectionChange      func(GridSelection, *Event, *Window)
	OnColumnOrderChange    func([]string, *Event, *Window)
	OnColumnPinChange      func(string, GridColumnPin, *Event, *Window)
	OnHiddenColumnsChange  func(map[string]bool, *Event, *Window)
	OnPageChange           func(int, *Event, *Window)
	OnDetailExpandedChange func(map[string]bool, *Event, *Window)
	OnCellEdit             func(GridCellEdit, *Event, *Window)
	OnRowsChange           func([]GridRow, *Event, *Window)
	OnCRUDError            func(string, *Event, *Window)
	OnCellFormat           func(GridRow, int, GridColumnCfg, string, *Window) GridCellFormat
	OnDetailRowView        func(GridRow, *Window) View
	OnCopyRows             func([]GridRow, *Event, *Window) (string, bool)
	OnRowActivate          func(GridRow, *Event, *Window)
	Disabled               bool
	Invisible              bool
	A11YLabel              string
	A11YDescription        string
}

// boolDefault returns *p if non-nil, else def.
func boolDefault(p *bool, def bool) bool {
	if p != nil {
		return *p
	}
	return def
}

// applyDataGridDefaults fills zero-valued fields with theme
// defaults and sensible fallbacks.
func applyDataGridDefaults(cfg *DataGridCfg) {
	s := guiTheme.DataGridStyle
	if !cfg.Sizing.IsSet() {
		cfg.Sizing = Some(FillFill)
	}
	if cfg.RowHeight == 0 {
		cfg.RowHeight = dataGridDefaultRowHeight
	}
	if cfg.HeaderHeight == 0 {
		cfg.HeaderHeight = dataGridDefaultHeaderHeight
	}
	if cfg.PageLimit == 0 {
		cfg.PageLimit = dataGridDefaultPageLimit
	}
	if cfg.QuickFilterDebounce == 0 {
		cfg.QuickFilterDebounce = 200 * time.Millisecond
	}
	if !cfg.ColorBackground.IsSet() {
		cfg.ColorBackground = s.ColorBackground
	}
	if !cfg.ColorHeader.IsSet() {
		cfg.ColorHeader = s.ColorHeader
	}
	if !cfg.ColorHeaderHover.IsSet() {
		cfg.ColorHeaderHover = s.ColorHeaderHover
	}
	if !cfg.ColorFilter.IsSet() {
		cfg.ColorFilter = s.ColorFilter
	}
	if !cfg.ColorQuickFilter.IsSet() {
		cfg.ColorQuickFilter = s.ColorQuickFilter
	}
	if !cfg.ColorRowHover.IsSet() {
		cfg.ColorRowHover = s.ColorRowHover
	}
	if !cfg.ColorRowAlt.IsSet() {
		cfg.ColorRowAlt = s.ColorRowAlt
	}
	if !cfg.ColorRowSelected.IsSet() {
		cfg.ColorRowSelected = s.ColorRowSelected
	}
	if !cfg.ColorBorder.IsSet() {
		cfg.ColorBorder = s.ColorBorder
	}
	if !cfg.ColorResizeHandle.IsSet() {
		cfg.ColorResizeHandle = s.ColorResizeHandle
	}
	if !cfg.ColorResizeActive.IsSet() {
		cfg.ColorResizeActive = s.ColorResizeActive
	}
	if !cfg.PaddingCell.IsSet() {
		cfg.PaddingCell = Some(s.PaddingCell)
	}
	if !cfg.PaddingHeader.IsSet() {
		cfg.PaddingHeader = Some(s.PaddingHeader)
	}
	if cfg.TextStyle == (TextStyle{}) {
		cfg.TextStyle = s.TextStyle
	}
	if cfg.TextStyleHeader == (TextStyle{}) {
		cfg.TextStyleHeader = s.TextStyleHeader
	}
	if cfg.TextStyleFilter == (TextStyle{}) {
		cfg.TextStyleFilter = s.TextStyleFilter
	}
	if !cfg.PaddingFilter.IsSet() {
		cfg.PaddingFilter = Some(s.PaddingFilter)
	}
	if !cfg.Radius.IsSet() {
		cfg.Radius = SomeF(s.Radius)
	}
	if !cfg.SizeBorder.IsSet() {
		cfg.SizeBorder = SomeF(s.SizeBorder)
	}
	for i := range cfg.Columns {
		gridColumnCfgDefaults(&cfg.Columns[i])
	}
}

// --- Internal state structs ---

type dataGridResizeState struct {
	Active         bool
	ColID          string
	StartMouseX    float32
	StartWidth     float32
	LastClickFrame uint64
	LastClickColID string
}

type dataGridColWidths struct {
	Widths map[string]float32
}

type dataGridPresentationCache struct {
	Signature     uint64
	Rows          []dataGridDisplayRow
	DataToDisplay map[int]int
	GroupRanges   map[string]int
	GroupCols     []string
}

type dataGridRangeState struct {
	AnchorRowID string
}

type dataGridEditState struct {
	EditingRowID   string
	LastClickRowID string
	LastClickFrame uint64
}

type dataGridCrudState struct {
	SourceSignature         uint64
	LocalRowsLen            int
	LocalRowsIDSignature    uint64
	LocalRowsSignatureValid bool
	CommittedRows           []GridRow
	WorkingRows             []GridRow
	DirtyRowIDs             map[string]bool
	DraftRowIDs             map[string]bool
	DeletedRowIDs           map[string]bool
	NextDraftSeq            int
	Saving                  bool
	SaveError               string
}

type dataGridSourceState struct {
	Rows           []GridRow
	Loading        bool
	LoadError      string
	HasLoaded      bool
	RequestID      uint64
	RequestKey     string
	QuerySignature uint64
	CurrentCursor  string
	NextCursor     string
	PrevCursor     string
	OffsetStart    int
	RowCount       *int
	HasMore        bool
	ReceivedCount  int
	RequestCount   int
	CancelledCount int
	StaleDropCount int
	ActiveAbort    *GridAbortController
	PaginationKind GridPaginationKind
	ConfigCursor   string
	PendingJumpRow int
	CachedCaps     GridDataCapabilities
	CapsCached     bool
	RowsDirty      bool
	RowsSignature  uint64
}

// dataGridCtx bundles commonly repeated DataGrid parameters
// to reduce function signatures. Constructed once per frame
// in DataGrid() and passed by value.
type dataGridCtx struct {
	cfg          *DataGridCfg
	columns      []GridColumnCfg
	columnWidths map[string]float32
	rowHeight    float32
	focusID      uint32
	scrollID     uint32
	editingRowID string
	w            *Window
}

// DataGrid renders a controlled, virtualized data grid view.
func (w *Window) DataGrid(cfg DataGridCfg) View {
	applyDataGridDefaults(&cfg)

	// Resolve data source and apply pending jump/selection.
	resolvedCfg, sourceState, hasSource, sourceCaps := dataGridResolveSourceCfg(cfg, w)
	if hasSource {
		dataGridSourceApplyPendingJumpSelection(&resolvedCfg, sourceState, w)
	}

	// Overlay CRUD working copy.
	var crudState dataGridCrudState
	crudEnabled := dataGridCrudEnabled(&resolvedCfg)
	if crudEnabled {
		nextCfg, nextCrudState := dataGridCrudResolveCfg(resolvedCfg, w)
		resolvedCfg = nextCfg
		crudState = nextCrudState
		if hasSource {
			dgSrc := StateMap[string, dataGridSourceState](w, nsDgSource, capModerate)
			if latestState, ok := dgSrc.Get(resolvedCfg.ID); ok {
				sourceState = latestState
			}
		}
	}

	// Interaction state.
	rowDeleteEnabled := dataGridCrudRowDeleteEnabled(&resolvedCfg, hasSource, sourceCaps)
	focusID := dataGridFocusID(&resolvedCfg)
	scrollID := dataGridScrollID(&resolvedCfg)
	dgHH := StateMap[string, string](w, nsDgHeaderHover, capModerate)
	hoveredColID, _ := dgHH.Get(resolvedCfg.ID)
	resizingColID := dataGridActiveResizeColID(resolvedCfg.ID, w)
	dgCO := StateMap[string, bool](w, nsDgChooserOpen, capModerate)
	chooserOpen, _ := dgCO.Get(resolvedCfg.ID)

	// Height/layout waterfall.
	rowHeight := dataGridRowHeight(&resolvedCfg, w)
	headerInScrollBody := boolDefault(resolvedCfg.ShowHeader, true) && !resolvedCfg.FreezeHeader
	staticTop := dataGridStaticTopHeight(&resolvedCfg, rowHeight, chooserOpen, headerInScrollBody)
	pageStart, pageEnd, pageIndex, pageCount := dataGridPageBounds(len(resolvedCfg.Rows),
		resolvedCfg.PageSize, resolvedCfg.PageIndex)
	pageIndices := dataGridPageRowIndices(pageStart, pageEnd)
	frozenTopIndices, bodyPageIndices := dataGridSplitFrozenTopIndices(&resolvedCfg, pageIndices)
	frozenTopIDs := dataGridFrozenTopIDSet(&resolvedCfg)
	pagerEnabled := dataGridPagerEnabled(&resolvedCfg, pageCount)
	sourcePagerEnabled := hasSource
	gridHeight := dataGridHeight(&resolvedCfg)
	if (pagerEnabled || sourcePagerEnabled) && gridHeight > 0 {
		gridHeight = f32Max(0, gridHeight-dataGridPagerHeight(&resolvedCfg))
	}
	if crudEnabled {
		toolbarHeight := dataGridCrudToolbarHeight(&resolvedCfg)
		if gridHeight > 0 {
			gridHeight = f32Max(0, gridHeight-toolbarHeight)
		}
	}
	virtualize := gridHeight > 0 && len(resolvedCfg.Rows) > 0
	scrollY := float32(0)
	if virtualize {
		sy := StateMap[uint32, float32](w, nsScrollY, capScroll)
		if v, ok := sy.Get(scrollID); ok {
			scrollY = -v
		}
	}

	// Build columns and presentation.
	columns := dataGridEffectiveColumns(resolvedCfg.Columns, resolvedCfg.ColumnOrder,
		resolvedCfg.HiddenColumnIDs)
	presentation := dataGridCachedPresentation(&resolvedCfg, columns, bodyPageIndices, w)
	if !hasSource {
		dataGridApplyPendingLocalJumpScroll(&resolvedCfg, gridHeight, rowHeight,
			staticTop, scrollID, presentation.DataToDisplay, w)
	}

	// Clear stale editing state.
	editingRowID := dataGridEditingRowID(resolvedCfg.ID, w)
	if editingRowID != "" && !dataGridHasRowID(resolvedCfg.Rows, editingRowID) {
		dataGridClearEditingRow(resolvedCfg.ID, w)
		editingRowID = ""
	}
	focusedColID := dataGridHeaderFocusedColID(&resolvedCfg, columns, w.IDFocus())

	// Column widths and header.
	columnWidths := dataGridColumnWidths(resolvedCfg.ID, resolvedCfg.Columns, w)
	dctx := dataGridCtx{
		cfg:          &resolvedCfg,
		columns:      columns,
		columnWidths: columnWidths,
		rowHeight:    rowHeight,
		focusID:      focusID,
		scrollID:     scrollID,
		editingRowID: editingRowID,
		w:            w,
	}
	totalWidth := dataGridColumnsTotalWidth(columns, columnWidths)
	headerView := dataGridHeaderRow(&resolvedCfg, columns, columnWidths, focusID,
		hoveredColID, resizingColID, focusedColID)
	headerHeight := dataGridHeaderHeight(&resolvedCfg)
	frozenTopViews, frozenTopDisplayRows := dataGridFrozenTopViews(dctx,
		frozenTopIndices, rowDeleteEnabled)
	sx := StateMap[uint32, float32](w, nsScrollX, capScroll)
	scrollX, _ := sx.Get(scrollID)

	// Visible range for virtualization.
	firstVisible, lastVisible := 0, len(presentation.Rows)-1
	if virtualize {
		firstVisible, lastVisible = dataGridVisibleRangeForScroll(scrollY, gridHeight,
			rowHeight, len(presentation.Rows), staticTop, dataGridVirtualBufferRows)
	}

	// Assemble scroll body rows.
	rows := dataGridScrollBodyRows(dctx, presentation,
		rowDeleteEnabled, headerInScrollBody, headerView,
		chooserOpen, hasSource, virtualize,
		firstVisible, lastVisible)

	// Scrollable body.
	scrollbarCfg := ScrollbarCfg{Overflow: resolvedCfg.Scrollbar}
	scrollBody := Column(ContainerCfg{
		ID:            resolvedCfg.ID + ":scroll",
		IDScroll:      scrollID,
		ScrollbarCfgX: &scrollbarCfg,
		ScrollbarCfgY: &scrollbarCfg,
		Color:         resolvedCfg.ColorBackground,
		Padding:       Some(dataGridScrollPadding(&resolvedCfg)),
		Spacing:       SomeF(0),
		Sizing:        FillFill,
		Content:       rows,
	})

	// Final assembly.
	content := dataGridFinalContent(dctx, scrollBody, headerView,
		headerHeight, totalWidth, scrollX, gridHeight, staticTop,
		frozenTopViews, frozenTopDisplayRows, crudEnabled, crudState,
		sourceCaps, hasSource, pagerEnabled, sourcePagerEnabled,
		pageIndex, pageCount, pageStart, pageEnd, presentation,
		sourceState)

	return Column(ContainerCfg{
		ID:              resolvedCfg.ID,
		IDFocus:         focusID,
		A11YRole:        AccessRoleGrid,
		A11YLabel:       resolvedCfg.A11YLabel,
		A11YDescription: resolvedCfg.A11YDescription,
		OnKeyDown: dataGridMakeOnKeydown(&resolvedCfg, columns, rowHeight,
			staticTop, scrollID, pageIndices, frozenTopIDs, presentation.DataToDisplay),
		OnChar:      dataGridMakeOnChar(&resolvedCfg, columns),
		OnMouseMove: dataGridMakeOnMouseMove(resolvedCfg.ID),
		Color:       resolvedCfg.ColorBackground,
		ColorBorder: resolvedCfg.ColorBorder,
		SizeBorder:  resolvedCfg.SizeBorder,
		Radius:      resolvedCfg.Radius,
		Padding:     NoPadding,
		Spacing:     SomeF(0),
		Disabled:    resolvedCfg.Disabled,
		Invisible:   resolvedCfg.Invisible,
		Sizing:      resolvedCfg.Sizing.Get(FillFill),
		Width:       resolvedCfg.Width,
		Height:      resolvedCfg.Height,
		MinWidth:    resolvedCfg.MinWidth,
		MaxWidth:    resolvedCfg.MaxWidth,
		MinHeight:   resolvedCfg.MinHeight,
		MaxHeight:   resolvedCfg.MaxHeight,
		Content:     content,
	})
}
