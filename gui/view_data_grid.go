package gui

import (
	"fmt"
	"math"
	"strconv"
	"strings"
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
	dataGridResizeKeyStep           = float32(8)
	dataGridResizeKeyStepLarge      = float32(24)
	dataGridHeaderControlWidth      = float32(12)
	dataGridHeaderReorderSpacing    = float32(1)
	dataGridHeaderLabelMinWidth     = float32(24)
	dataGridGroupIndentStep         = float32(14)
	dataGridDetailIndentGap         = float32(4)
	dataGridPDFPageWidth            = float32(612)
	dataGridPDFPageHeight           = float32(792)
	dataGridPDFMargin               = float32(40)
	dataGridPDFFontSize             = float32(10)
	dataGridPDFLineHeight           = float32(12)
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

const (
	GridColumnPinNone GridColumnPin = iota
	GridColumnPinLeft
	GridColumnPinRight
)

// GridAggregateOp specifies the aggregation operation.
type GridAggregateOp uint8

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
	Width            float32
	MinWidth         float32
	MaxWidth         float32
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
	if c.Width == 0 {
		c.Width = 120
	}
	if c.MinWidth == 0 {
		c.MinWidth = 60
	}
	if c.MaxWidth == 0 {
		c.MaxWidth = 600
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
	PaddingFilter          Padding
	TextStyle              TextStyle
	TextStyleHeader        TextStyle
	TextStyleFilter        TextStyle
	Radius                 float32
	SizeBorder             float32
	Scrollbar              ScrollbarOverflow
	Sizing                 Sizing
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
	if cfg.Sizing == (Sizing{}) {
		cfg.Sizing = FillFill
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
	if cfg.Radius == 0 {
		cfg.Radius = s.Radius
	}
	if cfg.SizeBorder == 0 {
		cfg.SizeBorder = s.SizeBorder
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

// --- Helper functions ---

func dataGridIndicatorTextStyle(base TextStyle) TextStyle {
	s := base
	s.Color = dataGridDimColor(base.Color)
	return s
}

func dataGridDimColor(c Color) Color {
	return Color{R: c.R, G: c.G, B: c.B, A: dataGridIndicatorAlpha, set: true}
}

func dataGridRowID(row GridRow, idx int) string {
	if row.ID != "" {
		return row.ID
	}
	autoID := dataGridRowAutoID(row)
	if autoID != "" {
		return autoID
	}
	return strconv.Itoa(idx)
}

func dataGridRowAutoID(row GridRow) string {
	if len(row.Cells) == 0 {
		return ""
	}
	keys := make([]string, 0, len(row.Cells))
	for k := range row.Cells {
		keys = append(keys, k)
	}
	// Sort for stability.
	for i := 1; i < len(keys); i++ {
		for j := i; j > 0 && keys[j] < keys[j-1]; j-- {
			keys[j], keys[j-1] = keys[j-1], keys[j]
		}
	}
	h := dataGridFnv64Offset
	for i, key := range keys {
		if i > 0 {
			h = dataGridFnv64Byte(h, dataGridUnitSep[0])
		}
		h = dataGridFnv64Str(h, key)
		h = dataGridFnv64Byte(h, '=')
		h = dataGridFnv64Str(h, row.Cells[key])
	}
	return fmt.Sprintf("__auto_%016x", h)
}

func dataGridHeight(cfg *DataGridCfg) float32 {
	if cfg.Height > 0 {
		return cfg.Height
	}
	if cfg.MaxHeight > 0 {
		return cfg.MaxHeight
	}
	return 0
}

func dataGridPagerEnabled(cfg *DataGridCfg, pageCount int) bool {
	return cfg.PageSize > 0 && pageCount > 1
}

func dataGridPagerHeight(cfg *DataGridCfg) float32 {
	if cfg.RowHeight > 0 {
		return cfg.RowHeight
	}
	return dataGridHeaderHeight(cfg)
}

func dataGridPagerPadding(cfg *DataGridCfg) Padding {
	pc := cfg.PaddingCell.Get(Padding{})
	left := f32Max(cfg.PaddingFilter.Left, pc.Left)
	right := f32Max(cfg.PaddingFilter.Right, pc.Right)
	return NewPadding(cfg.PaddingFilter.Top, right, cfg.PaddingFilter.Bottom, left)
}

func dataGridHeaderHeight(cfg *DataGridCfg) float32 {
	if cfg.HeaderHeight > 0 {
		return cfg.HeaderHeight
	}
	return cfg.RowHeight
}

func dataGridFilterHeight(cfg *DataGridCfg) float32 {
	return dataGridHeaderHeight(cfg)
}

func dataGridQuickFilterHeight(cfg *DataGridCfg) float32 {
	return dataGridHeaderHeight(cfg)
}

func dataGridRowHeight(cfg *DataGridCfg, w *Window) float32 {
	if cfg.RowHeight > 0 {
		return cfg.RowHeight
	}
	return cfg.TextStyle.Size + cfg.PaddingCell.Get(Padding{}).Height() + cfg.SizeBorder
}

func dataGridStaticTopHeight(cfg *DataGridCfg, rowHeight float32, chooserOpen bool, includeHeader bool) float32 {
	top := float32(0)
	if cfg.ShowColumnChooser {
		top += dataGridColumnChooserHeight(cfg, chooserOpen)
	}
	if includeHeader {
		top += dataGridHeaderHeight(cfg)
	}
	if cfg.ShowFilterRow {
		top += dataGridFilterHeight(cfg)
	}
	return top
}

func dataGridFocusID(cfg *DataGridCfg) uint32 {
	if cfg.IDFocus > 0 {
		return cfg.IDFocus
	}
	return fnvSum32(cfg.ID + ":focus")
}

func dataGridScrollID(cfg *DataGridCfg) uint32 {
	if cfg.IDScroll > 0 {
		return cfg.IDScroll
	}
	return fnvSum32(cfg.ID + ":scroll")
}

// dataGridVisibleRangeForScroll converts scroll position to
// the range of row indices to render.
func dataGridVisibleRangeForScroll(scrollY, viewportHeight, rowHeight float32, rowCount int, staticTop float32, buffer int) (int, int) {
	if rowCount == 0 || rowHeight <= 0 || viewportHeight <= 0 {
		return 0, -1
	}
	bodyScroll := scrollY - staticTop
	if bodyScroll < 0 {
		bodyScroll = 0
	}
	first := int(bodyScroll / rowHeight)
	visibleRows := int(viewportHeight/rowHeight) + 1
	firstVisible := first - buffer
	if firstVisible < 0 {
		firstVisible = 0
	}
	lastVisible := first + visibleRows + buffer
	if lastVisible > rowCount-1 {
		lastVisible = rowCount - 1
	}
	if firstVisible > lastVisible {
		firstVisible = lastVisible
	}
	return firstVisible, lastVisible
}

// --- Presentation building ---

func dataGridPresentation_(cfg *DataGridCfg, columns []GridColumnCfg) dataGridPresentation {
	return dataGridPresentationRows(cfg, columns, dataGridVisibleRowIndices(len(cfg.Rows), nil))
}

func dataGridCachedPresentation(cfg *DataGridCfg, columns []GridColumnCfg, rowIndices []int, w *Window) dataGridPresentation {
	groupCols := dataGridGroupColumns(cfg.GroupBy, columns)
	valueCols := dataGridPresentationValueCols(groupCols, cfg.Aggregates)
	signature := dataGridPresentationSignature(cfg, columns, rowIndices, groupCols, valueCols)
	dgPC := StateMap[string, dataGridPresentationCache](w, nsDgPresentation, capModerate)
	if cached, ok := dgPC.Get(cfg.ID); ok {
		if cached.Signature == signature {
			return dataGridPresentation{
				Rows:          cached.Rows,
				DataToDisplay: cached.DataToDisplay,
			}
		}
	}
	visibleIndices := dataGridVisibleRowIndices(len(cfg.Rows), rowIndices)
	var groupRanges map[string]int
	if len(groupCols) > 0 && len(visibleIndices) > 0 {
		groupRanges = dataGridGroupRanges(cfg.Rows, visibleIndices, groupCols)
	} else {
		groupRanges = map[string]int{}
	}
	pres := dataGridPresentationRowsWithGroupRanges(cfg, columns, visibleIndices, groupCols, groupRanges)
	dgPC.Set(cfg.ID, dataGridPresentationCache{
		Signature:     signature,
		Rows:          pres.Rows,
		DataToDisplay: pres.DataToDisplay,
		GroupRanges:   groupRanges,
		GroupCols:     groupCols,
	})
	return pres
}

func dataGridPresentationSignature(cfg *DataGridCfg, columns []GridColumnCfg, rowIndices []int, groupCols []string, valueCols []string) uint64 {
	h := dataGridFnv64Offset
	visibleIndices := dataGridVisibleRowIndices(len(cfg.Rows), rowIndices)
	if len(groupCols) == 0 && len(cfg.Aggregates) == 0 && cfg.OnDetailRowView == nil {
		h = dataGridFnv64Str(h, cfg.ID)
		h = dataGridFnv64Byte(h, 0x1e)
		for _, idx := range visibleIndices {
			h = dataGridFnv64U64(h, uint64(idx))
			h = dataGridFnv64Byte(h, 0x1f)
			row := cfg.Rows[idx]
			if row.ID != "" {
				h = dataGridFnv64Str(h, row.ID)
			}
			h = dataGridFnv64Byte(h, 0x1f)
		}
		return h
	}
	groupTitles := dataGridGroupTitles(columns)
	h = dataGridFnv64Str(h, cfg.ID)
	h = dataGridFnv64Byte(h, 0x1e)
	for _, idx := range rowIndices {
		h = dataGridFnv64U64(h, uint64(idx))
		h = dataGridFnv64Byte(h, 0x1f)
	}
	h = dataGridFnv64Byte(h, 0x1e)
	for _, colID := range groupCols {
		h = dataGridFnv64Str(h, colID)
		h = dataGridFnv64Byte(h, 0x1f)
		h = dataGridFnv64Str(h, groupTitles[colID])
		h = dataGridFnv64Byte(h, 0x1f)
	}
	h = dataGridFnv64Byte(h, 0x1e)
	for _, agg := range cfg.Aggregates {
		h = dataGridFnv64Str(h, agg.ColID)
		h = dataGridFnv64Byte(h, 0x1f)
		h = dataGridFnv64Byte(h, byte(agg.Op))
		h = dataGridFnv64Byte(h, 0x1f)
		h = dataGridFnv64Str(h, agg.Label)
		h = dataGridFnv64Byte(h, 0x1f)
	}
	detailEnabled := cfg.OnDetailRowView != nil
	if detailEnabled {
		h = dataGridFnv64Byte(h, '1')
	} else {
		h = dataGridFnv64Byte(h, '0')
	}
	for _, rowIdx := range visibleIndices {
		row := cfg.Rows[rowIdx]
		rowID := dataGridRowID(row, rowIdx)
		h = dataGridFnv64Byte(h, 0x1e)
		h = dataGridFnv64U64(h, uint64(rowIdx))
		h = dataGridFnv64Byte(h, 0x1f)
		h = dataGridFnv64Str(h, rowID)
		h = dataGridFnv64Byte(h, 0x1f)
		if detailEnabled && dataGridDetailRowExpanded(cfg, rowID) {
			h = dataGridFnv64Byte(h, '1')
		} else {
			h = dataGridFnv64Byte(h, '0')
		}
		for _, colID := range valueCols {
			h = dataGridFnv64Byte(h, 0x1f)
			h = dataGridFnv64Str(h, colID)
			h = dataGridFnv64Byte(h, '=')
			h = dataGridFnv64Str(h, row.Cells[colID])
		}
	}
	return h
}

func dataGridPresentationValueCols(groupCols []string, aggregates []GridAggregateCfg) []string {
	seen := map[string]bool{}
	cols := make([]string, 0, len(groupCols)+len(aggregates))
	for _, colID := range groupCols {
		if colID == "" || seen[colID] {
			continue
		}
		seen[colID] = true
		cols = append(cols, colID)
	}
	for _, agg := range aggregates {
		if agg.Op == GridAggregateCount || agg.ColID == "" || seen[agg.ColID] {
			continue
		}
		seen[agg.ColID] = true
		cols = append(cols, agg.ColID)
	}
	// Sort for stability.
	for i := 1; i < len(cols); i++ {
		for j := i; j > 0 && cols[j] < cols[j-1]; j-- {
			cols[j], cols[j-1] = cols[j-1], cols[j]
		}
	}
	return cols
}

func dataGridPresentationRows(cfg *DataGridCfg, columns []GridColumnCfg, rowIndices []int) dataGridPresentation {
	visibleIndices := dataGridVisibleRowIndices(len(cfg.Rows), rowIndices)
	groupCols := dataGridGroupColumns(cfg.GroupBy, columns)
	var groupRanges map[string]int
	if len(groupCols) > 0 && len(visibleIndices) > 0 {
		groupRanges = dataGridGroupRanges(cfg.Rows, visibleIndices, groupCols)
	} else {
		groupRanges = map[string]int{}
	}
	return dataGridPresentationRowsWithGroupRanges(cfg, columns, visibleIndices, groupCols, groupRanges)
}

func dataGridPresentationRowsWithGroupRanges(cfg *DataGridCfg, columns []GridColumnCfg, visibleIndices []int, groupCols []string, groupRanges map[string]int) dataGridPresentation {
	rows := make([]dataGridDisplayRow, 0, len(cfg.Rows)+8)
	dataToDisplay := map[int]int{}
	if len(groupCols) == 0 || len(visibleIndices) == 0 {
		for _, rowIdx := range visibleIndices {
			row := cfg.Rows[rowIdx]
			dataToDisplay[rowIdx] = len(rows)
			rows = append(rows, dataGridDisplayRow{
				Kind:       dataGridDisplayRowData,
				DataRowIdx: rowIdx,
			})
			if cfg.OnDetailRowView != nil && dataGridDetailRowExpanded(cfg, dataGridRowID(row, rowIdx)) {
				rows = append(rows, dataGridDisplayRow{
					Kind:       dataGridDisplayRowDetail,
					DataRowIdx: rowIdx,
				})
			}
		}
		return dataGridPresentation{Rows: rows, DataToDisplay: dataToDisplay}
	}

	groupTitles := dataGridGroupTitles(columns)
	prevValues := make([]string, len(groupCols))
	hasPrev := false

	for localIdx, rowIdx := range visibleIndices {
		row := cfg.Rows[rowIdx]
		values := make([]string, len(groupCols))
		for i, colID := range groupCols {
			values[i] = row.Cells[colID]
		}
		changeDepth := -1
		if !hasPrev {
			changeDepth = 0
		} else {
			for depth, value := range values {
				if value != prevValues[depth] {
					changeDepth = depth
					break
				}
			}
		}
		if changeDepth >= 0 {
			for depth := changeDepth; depth < len(groupCols); depth++ {
				colID := groupCols[depth]
				rangeEndLocal, ok := groupRanges[dataGridGroupRangeKey(depth, localIdx)]
				if !ok {
					rangeEndLocal = localIdx
				}
				count := intMax(0, rangeEndLocal-localIdx+1)
				rangeEnd := visibleIndices[rangeEndLocal]
				rows = append(rows, dataGridDisplayRow{
					Kind:          dataGridDisplayRowGroupHeader,
					GroupColID:    colID,
					GroupValue:    values[depth],
					GroupColTitle: groupTitles[colID],
					GroupDepth:    depth,
					GroupCount:    count,
					AggregateText: dataGridGroupAggregateText(cfg, rowIdx, rangeEnd),
				})
			}
		}
		dataToDisplay[rowIdx] = len(rows)
		rows = append(rows, dataGridDisplayRow{
			Kind:       dataGridDisplayRowData,
			DataRowIdx: rowIdx,
		})
		if cfg.OnDetailRowView != nil && dataGridDetailRowExpanded(cfg, dataGridRowID(row, rowIdx)) {
			rows = append(rows, dataGridDisplayRow{
				Kind:       dataGridDisplayRowDetail,
				DataRowIdx: rowIdx,
			})
		}
		copy(prevValues, values)
		hasPrev = true
	}
	return dataGridPresentation{Rows: rows, DataToDisplay: dataToDisplay}
}

func dataGridGroupColumns(groupBy []string, columns []GridColumnCfg) []string {
	if len(groupBy) == 0 {
		return nil
	}
	available := map[string]bool{}
	for _, col := range columns {
		if col.ID != "" {
			available[col.ID] = true
		}
	}
	seen := map[string]bool{}
	cols := make([]string, 0, len(groupBy))
	for _, colID := range groupBy {
		if colID == "" || seen[colID] || !available[colID] {
			continue
		}
		seen[colID] = true
		cols = append(cols, colID)
	}
	return cols
}

func dataGridGroupTitles(columns []GridColumnCfg) map[string]string {
	titles := make(map[string]string, len(columns))
	for _, col := range columns {
		if col.ID != "" {
			titles[col.ID] = col.Title
		}
	}
	return titles
}

func dataGridGroupRangeKey(depth, startIdx int) string {
	return fmt.Sprintf("%d:%d", depth, startIdx)
}

func dataGridGroupRanges(rows []GridRow, indices []int, groupCols []string) map[string]int {
	ranges := map[string]int{}
	if len(indices) == 0 || len(groupCols) == 0 {
		return ranges
	}
	starts := make([]int, len(groupCols))
	values := make([]string, len(groupCols))
	for depth, colID := range groupCols {
		values[depth] = rows[indices[0]].Cells[colID]
	}
	for i := 1; i < len(indices); i++ {
		row := rows[indices[i]]
		changeDepth := -1
		for depth, colID := range groupCols {
			if row.Cells[colID] != values[depth] {
				changeDepth = depth
				break
			}
		}
		if changeDepth < 0 {
			continue
		}
		for depth := len(groupCols) - 1; depth >= changeDepth; depth-- {
			ranges[dataGridGroupRangeKey(depth, starts[depth])] = i - 1
		}
		for dep := changeDepth; dep < len(groupCols); dep++ {
			starts[dep] = i
			values[dep] = row.Cells[groupCols[dep]]
		}
	}
	last := len(indices) - 1
	for depth := len(groupCols) - 1; depth >= 0; depth-- {
		ranges[dataGridGroupRangeKey(depth, starts[depth])] = last
	}
	return ranges
}

func dataGridGroupAggregateText(cfg *DataGridCfg, startIdx, endIdx int) string {
	if len(cfg.Aggregates) == 0 || startIdx < 0 || endIdx < startIdx || endIdx >= len(cfg.Rows) {
		return ""
	}
	parts := make([]string, 0, len(cfg.Aggregates))
	for _, agg := range cfg.Aggregates {
		value, ok := dataGridAggregateValue(cfg.Rows, startIdx, endIdx, agg)
		if !ok {
			continue
		}
		parts = append(parts, dataGridAggregateLabel(agg)+": "+value)
	}
	return strings.Join(parts, "  ")
}

func dataGridAggregateLabel(agg GridAggregateCfg) string {
	if agg.Label != "" {
		return agg.Label
	}
	if agg.Op == GridAggregateCount {
		return "count"
	}
	if agg.ColID == "" {
		return agg.Op.String()
	}
	return agg.Op.String() + " " + agg.ColID
}

func dataGridAggregateValue(rows []GridRow, startIdx, endIdx int, agg GridAggregateCfg) (string, bool) {
	if agg.Op == GridAggregateCount {
		return strconv.Itoa(endIdx - startIdx + 1), true
	}
	if agg.ColID == "" {
		return "", false
	}
	var values []float64
	for idx := startIdx; idx <= endIdx; idx++ {
		raw := rows[idx].Cells[agg.ColID]
		n, ok := dataGridParseNumber(raw)
		if !ok {
			continue
		}
		values = append(values, n)
	}
	if len(values) == 0 {
		return "", false
	}
	var result float64
	switch agg.Op {
	case GridAggregateSum, GridAggregateAvg:
		for _, v := range values {
			result += v
		}
		if agg.Op == GridAggregateAvg {
			result /= float64(len(values))
		}
	case GridAggregateMin:
		result = values[0]
		for _, v := range values[1:] {
			if v < result {
				result = v
			}
		}
	case GridAggregateMax:
		result = values[0]
		for _, v := range values[1:] {
			if v > result {
				result = v
			}
		}
	}
	return dataGridFormatNumber(result), true
}

func dataGridParseNumber(value string) (float64, bool) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return 0, false
	}
	n, err := strconv.ParseFloat(trimmed, 64)
	if err != nil {
		return 0, false
	}
	if math.IsNaN(n) || math.IsInf(n, 0) {
		return 0, false
	}
	return n, true
}

func dataGridFormatNumber(value float64) string {
	text := fmt.Sprintf("%.4f", value)
	for strings.Contains(text, ".") && strings.HasSuffix(text, "0") {
		text = text[:len(text)-1]
	}
	if strings.HasSuffix(text, ".") {
		text = text[:len(text)-1]
	}
	return text
}

func dataGridDetailRowExpanded(cfg *DataGridCfg, rowID string) bool {
	return rowID != "" && cfg.DetailExpandedRowIDs[rowID]
}

func dataGridHasSource(cfg *DataGridCfg) bool {
	return cfg.DataSource != nil
}

func dataGridColumnChooserHeight(cfg *DataGridCfg, isOpen bool) float32 {
	base := cfg.RowHeight
	if base <= 0 {
		base = dataGridHeaderHeight(cfg)
	}
	if isOpen {
		return base * 2
	}
	return base
}

// --- Page bounds ---

func dataGridPageBounds(totalRows, pageSize, requestedPage int) (start, end, pageIndex, pageCount int) {
	if totalRows <= 0 {
		return 0, 0, 0, 1
	}
	if pageSize <= 0 {
		return 0, totalRows, 0, 1
	}
	ps := intMin(pageSize, totalRows)
	pageCount = intMax(1, (totalRows+ps-1)/ps)
	pageIndex = intClamp(requestedPage, 0, pageCount-1)
	start = pageIndex * ps
	end = intMin(totalRows, start+ps)
	return
}

func dataGridPageRowIndices(start, end int) []int {
	if end <= start || start < 0 {
		return nil
	}
	indices := make([]int, 0, end-start)
	for idx := start; idx < end; idx++ {
		indices = append(indices, idx)
	}
	return indices
}

func dataGridVisibleRowIndices(rowCount int, pageIndices []int) []int {
	if len(pageIndices) > 0 {
		return pageIndices
	}
	return dataGridPageRowIndices(0, intMax(0, rowCount))
}

func dataGridIndexInList(values []int, target int) int {
	for idx, v := range values {
		if v == target {
			return idx
		}
	}
	return -1
}

func dataGridHasRowID(rows []GridRow, rowID string) bool {
	if rowID == "" {
		return false
	}
	for idx, row := range rows {
		if dataGridRowID(row, idx) == rowID {
			return true
		}
	}
	return false
}

func dataGridActiveRowIndex(rows []GridRow, selection GridSelection) int {
	res := dataGridActiveRowIndexStrict(rows, selection)
	if res >= 0 {
		return res
	}
	if len(rows) > 0 {
		return 0
	}
	return -1
}

func dataGridActiveRowIndexStrict(rows []GridRow, selection GridSelection) int {
	if len(rows) == 0 {
		return -1
	}
	hasActive := selection.ActiveRowID != ""
	hasSelected := len(selection.SelectedRowIDs) > 0
	if !hasActive && !hasSelected {
		return -1
	}
	firstSelected := -1
	for idx, row := range rows {
		id := dataGridRowID(row, idx)
		if hasActive && id == selection.ActiveRowID {
			return idx
		}
		if firstSelected < 0 && hasSelected && selection.SelectedRowIDs[id] {
			firstSelected = idx
		}
	}
	return firstSelected
}

func dataGridSelectedRows(rows []GridRow, selection GridSelection) []GridRow {
	if len(selection.SelectedRowIDs) == 0 {
		return nil
	}
	var selected []GridRow
	for idx, row := range rows {
		if selection.SelectedRowIDs[dataGridRowID(row, idx)] {
			selected = append(selected, row)
		}
	}
	return selected
}

func dataGridPageRows(cfg *DataGridCfg, rowHeight float32) int {
	if rowHeight <= 0 {
		return 1
	}
	page := int(dataGridHeight(cfg) / rowHeight)
	if page < 1 {
		return 1
	}
	return page
}

func dataGridEditingEnabled(cfg *DataGridCfg) bool {
	if cfg.OnCellEdit == nil && !cfg.ShowCRUDToolbar {
		return false
	}
	for _, col := range cfg.Columns {
		if col.Editable {
			return true
		}
	}
	return false
}

func dataGridCrudEnabled(cfg *DataGridCfg) bool {
	return cfg.ShowCRUDToolbar
}

// --- FNV-1a helpers (64-bit, presentation cache) ---

func dataGridFnv64U64(h, val uint64) uint64 {
	h = (h ^ (val & 0xff)) * dataGridFnv64Prime
	h = (h ^ ((val >> 8) & 0xff)) * dataGridFnv64Prime
	h = (h ^ ((val >> 16) & 0xff)) * dataGridFnv64Prime
	h = (h ^ ((val >> 24) & 0xff)) * dataGridFnv64Prime
	h = (h ^ ((val >> 32) & 0xff)) * dataGridFnv64Prime
	h = (h ^ ((val >> 40) & 0xff)) * dataGridFnv64Prime
	h = (h ^ ((val >> 48) & 0xff)) * dataGridFnv64Prime
	h = (h ^ ((val >> 56) & 0xff)) * dataGridFnv64Prime
	return h
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
	totalWidth := dataGridColumnsTotalWidth(columns, columnWidths)
	headerView := dataGridHeaderRow(&resolvedCfg, columns, columnWidths, focusID,
		hoveredColID, resizingColID, focusedColID)
	headerHeight := dataGridHeaderHeight(&resolvedCfg)
	frozenTopViews, frozenTopDisplayRows := dataGridFrozenTopViews(&resolvedCfg,
		frozenTopIndices, columns, columnWidths, rowHeight, focusID, editingRowID,
		rowDeleteEnabled, w)
	sx := StateMap[uint32, float32](w, nsScrollX, capScroll)
	scrollX, _ := sx.Get(scrollID)
	lastRowIdx := len(presentation.Rows) - 1

	// Visible range for virtualization.
	firstVisible, lastVisible := 0, lastRowIdx
	if virtualize {
		firstVisible, lastVisible = dataGridVisibleRangeForScroll(scrollY, gridHeight,
			rowHeight, len(presentation.Rows), staticTop, dataGridVirtualBufferRows)
	}

	// Assemble scroll body rows.
	rows := make([]View, 0, len(presentation.Rows)+8)
	if resolvedCfg.ShowColumnChooser {
		rows = append(rows, dataGridColumnChooserRow(&resolvedCfg, chooserOpen, focusID))
	}
	if headerInScrollBody {
		rows = append(rows, headerView)
	}
	if resolvedCfg.ShowFilterRow {
		rows = append(rows, dataGridFilterRow(&resolvedCfg, columns, columnWidths))
	}
	if hasSource && resolvedCfg.Loading && len(presentation.Rows) == 0 {
		rows = append(rows, dataGridSourceStatusRow(&resolvedCfg, guiLocale.StrLoading))
	}
	if hasSource && resolvedCfg.LoadError != "" && len(presentation.Rows) == 0 {
		rows = append(rows, dataGridSourceStatusRow(&resolvedCfg,
			guiLocale.StrLoadError+": "+resolvedCfg.LoadError))
	}

	if virtualize && firstVisible > 0 {
		rows = append(rows, Rectangle(RectangleCfg{
			Color:  ColorTransparent,
			Height: float32(firstVisible) * rowHeight,
			Sizing: FillFixed,
		}))
	}

	// Emit visible rows.
	for rowIdx := firstVisible; rowIdx <= lastVisible; rowIdx++ {
		if rowIdx < 0 || rowIdx >= len(presentation.Rows) {
			continue
		}
		entry := presentation.Rows[rowIdx]
		if entry.Kind == dataGridDisplayRowGroupHeader {
			rows = append(rows, dataGridGroupHeaderRowView(&resolvedCfg, entry, rowHeight))
			continue
		}
		if entry.Kind == dataGridDisplayRowDetail {
			if entry.DataRowIdx < 0 || entry.DataRowIdx >= len(resolvedCfg.Rows) {
				continue
			}
			rows = append(rows, dataGridDetailRowView(&resolvedCfg,
				resolvedCfg.Rows[entry.DataRowIdx], entry.DataRowIdx, columns,
				columnWidths, rowHeight, focusID, w))
			continue
		}
		if entry.DataRowIdx < 0 || entry.DataRowIdx >= len(resolvedCfg.Rows) {
			continue
		}
		rows = append(rows, dataGridRowView(&resolvedCfg,
			resolvedCfg.Rows[entry.DataRowIdx], entry.DataRowIdx, columns,
			columnWidths, rowHeight, focusID, editingRowID, rowDeleteEnabled, w))
	}

	if virtualize && lastVisible < lastRowIdx {
		remaining := lastRowIdx - lastVisible
		rows = append(rows, Rectangle(RectangleCfg{
			Color:  ColorTransparent,
			Height: float32(remaining) * rowHeight,
			Sizing: FillFixed,
		}))
	}

	// Scrollable body.
	scrollbarCfg := ScrollbarCfg{Overflow: resolvedCfg.Scrollbar}
	scrollBody := Column(ContainerCfg{
		ID:            resolvedCfg.ID + ":scroll",
		IDScroll:      scrollID,
		ScrollbarCfgX: &scrollbarCfg,
		ScrollbarCfgY: &scrollbarCfg,
		Color:         resolvedCfg.ColorBackground,
		Padding:       Some(dataGridScrollPadding(&resolvedCfg)),
		Spacing:       Some(float32(0)),
		Sizing:        FillFill,
		Content:       rows,
	})

	// Final assembly.
	content := make([]View, 0, 6)
	if crudEnabled {
		content = append(content, dataGridCrudToolbarRow(&resolvedCfg, crudState,
			sourceCaps, hasSource, focusID))
	}
	if resolvedCfg.ShowQuickFilter {
		qfHeight := dataGridQuickFilterHeight(&resolvedCfg)
		content = append(content, dataGridFrozenTopZone(&resolvedCfg,
			[]View{dataGridQuickFilterRow(&resolvedCfg)},
			qfHeight, totalWidth, scrollX))
	}
	if boolDefault(resolvedCfg.ShowHeader, true) && resolvedCfg.FreezeHeader {
		content = append(content, dataGridFrozenTopZone(&resolvedCfg,
			[]View{headerView}, headerHeight, totalWidth, scrollX))
	}
	if frozenTopDisplayRows > 0 {
		frozenHeight := float32(frozenTopDisplayRows) * rowHeight
		content = append(content, dataGridFrozenTopZone(&resolvedCfg,
			frozenTopViews, frozenHeight, totalWidth, scrollX))
	}
	content = append(content, scrollBody)
	if pagerEnabled {
		totalRows := len(resolvedCfg.Rows)
		if resolvedCfg.RowCount != nil {
			totalRows = *resolvedCfg.RowCount
		}
		dgJump := StateMap[string, string](w, nsDgJump, capModerate)
		jumpText, _ := dgJump.Get(resolvedCfg.ID)
		content = append(content, dataGridPagerRow(&resolvedCfg, focusID, pageIndex,
			pageCount, pageStart, pageEnd, totalRows, gridHeight, rowHeight, staticTop,
			scrollID, presentation.DataToDisplay, jumpText))
	}
	if sourcePagerEnabled {
		dgJump := StateMap[string, string](w, nsDgJump, capModerate)
		jumpText, _ := dgJump.Get(resolvedCfg.ID)
		content = append(content, dataGridSourcePagerRow(&resolvedCfg, focusID,
			sourceState, sourceCaps, jumpText))
	}

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
		SizeBorder:  Some(resolvedCfg.SizeBorder),
		Radius:      Some(resolvedCfg.Radius),
		Padding:     Some(PaddingNone),
		Spacing:     Some(float32(0)),
		Sizing:      resolvedCfg.Sizing,
		Width:       resolvedCfg.Width,
		Height:      resolvedCfg.Height,
		MinWidth:    resolvedCfg.MinWidth,
		MaxWidth:    resolvedCfg.MaxWidth,
		MinHeight:   resolvedCfg.MinHeight,
		MaxHeight:   resolvedCfg.MaxHeight,
		Content:     content,
	})
}
