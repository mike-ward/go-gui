package gui

import (
	"crypto/rand"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

const dataGridSourceMaxPageLimit = 10000

// FNV-1a 64-bit constants.
const (
	dataGridFnv64Offset = uint64(14695981039346656037)
	dataGridFnv64Prime  = uint64(1099511628211)
)

// GridCursorPageReq requests a cursor-based page.
type GridCursorPageReq struct {
	Cursor string
	Limit  int
}

func (GridCursorPageReq) gridPageRequest() {}

// GridOffsetPageReq requests an offset-based page.
type GridOffsetPageReq struct {
	StartIndex int
	EndIndex   int
}

func (GridOffsetPageReq) gridPageRequest() {}

// GridPageRequest is satisfied by GridCursorPageReq or
// GridOffsetPageReq.
type GridPageRequest interface {
	gridPageRequest()
}

// GridAbortSignal communicates cancellation via an atomic bool.
type GridAbortSignal struct {
	aborted atomic.Bool
}

// IsAborted reports cancellation status.
func (s *GridAbortSignal) IsAborted() bool {
	if s == nil {
		return false
	}
	return s.aborted.Load()
}

// GridAbortController manages an abort signal.
type GridAbortController struct {
	Signal *GridAbortSignal
}

// NewGridAbortController allocates a fresh abort controller.
func NewGridAbortController() *GridAbortController {
	return &GridAbortController{Signal: &GridAbortSignal{}}
}

// Abort marks the request as cancelled.
func (c *GridAbortController) Abort() {
	if c.Signal == nil {
		return
	}
	c.Signal.aborted.Store(true)
}

// GridDataRequest is the request payload for FetchData.
type GridDataRequest struct {
	GridID    string
	Query     GridQueryState
	Page      GridPageRequest
	Signal    *GridAbortSignal
	RequestID uint64
}

// GridDataResult is the response from FetchData.
type GridDataResult struct {
	Rows          []GridRow
	NextCursor    string
	PrevCursor    string
	RowCount      int  // -1 when unknown
	HasMore       bool
	ReceivedCount int
}

// GridDataCapabilities describes what a data source supports.
type GridDataCapabilities struct {
	SupportsCursorPagination bool
	SupportsOffsetPagination bool
	SupportsNumberedPages    bool
	RowCountKnown            bool
	SupportsCreate           bool
	SupportsUpdate           bool
	SupportsDelete           bool
	SupportsBatchDelete      bool
}

// GridMutationRequest is the request payload for MutateData.
type GridMutationRequest struct {
	GridID    string
	Kind      GridMutationKind
	Query     GridQueryState
	Rows      []GridRow
	RowIDs    []string
	Edits     []GridCellEdit
	Signal    *GridAbortSignal
	RequestID uint64
}

// GridMutationResult is the response from MutateData.
type GridMutationResult struct {
	Created    []GridRow
	Updated    []GridRow
	DeletedIDs []string
	FailedIDs  []string
	Errors     map[string]string
	RowCount   int // -1 when unknown
}

// DataGridDataSource is the interface for grid data providers.
type DataGridDataSource interface {
	Capabilities() GridDataCapabilities
	FetchData(req GridDataRequest) (GridDataResult, error)
	MutateData(req GridMutationRequest) (GridMutationResult, error)
}

// InMemoryDataSource implements DataGridDataSource using an
// in-memory row slice.
type InMemoryDataSource struct {
	mu            sync.RWMutex
	Rows          []GridRow
	DefaultLimit  int
	LatencyMs     int
	RowCountKnown bool
	SupportsCursor bool
	SupportsOffset bool
}

// NewInMemoryDataSource creates an InMemoryDataSource with
// sensible defaults.
func NewInMemoryDataSource(rows []GridRow) *InMemoryDataSource {
	return &InMemoryDataSource{
		Rows:           rows,
		DefaultLimit:   100,
		RowCountKnown:  true,
		SupportsCursor: true,
		SupportsOffset: true,
	}
}

func (s *InMemoryDataSource) Capabilities() GridDataCapabilities {
	return GridDataCapabilities{
		SupportsCursorPagination: s.SupportsCursor,
		SupportsOffsetPagination: s.SupportsOffset,
		SupportsNumberedPages:    s.SupportsOffset,
		RowCountKnown:            s.RowCountKnown,
		SupportsCreate:           true,
		SupportsUpdate:           true,
		SupportsDelete:           true,
		SupportsBatchDelete:      true,
	}
}

func (s *InMemoryDataSource) FetchData(req GridDataRequest) (GridDataResult, error) {
	if err := dataGridSourceSleepWithAbort(req.Signal, s.LatencyMs); err != nil {
		return GridDataResult{}, err
	}
	s.mu.RLock()
	rows := make([]GridRow, len(s.Rows))
	copy(rows, s.Rows)
	defaultLimit := s.DefaultLimit
	rowCountKnown := s.RowCountKnown
	s.mu.RUnlock()
	return dataGridSourceInMemoryFetch(
		rows, defaultLimit, 0, rowCountKnown, req)
}

func (s *InMemoryDataSource) MutateData(req GridMutationRequest) (GridMutationResult, error) {
	if err := dataGridSourceSleepWithAbort(req.Signal, s.LatencyMs); err != nil {
		return GridMutationResult{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	return dataGridSourceInMemoryMutate(
		&s.Rows, 0, s.RowCountKnown, req)
}

func dataGridSourceInMemoryFetch(
	rows []GridRow, defaultLimit, latencyMs int,
	rowCountKnown bool, req GridDataRequest,
) (GridDataResult, error) {
	if err := dataGridSourceSleepWithAbort(req.Signal, latencyMs); err != nil {
		return GridDataResult{}, err
	}
	filtered := dataGridSourceApplyQuery(rows, req.Query)
	limit := intClamp(
		nonZero(defaultLimit, 100), 1, dataGridSourceMaxPageLimit)
	var start, end int
	switch p := req.Page.(type) {
	case GridCursorPageReq:
		s := intClamp(dataGridSourceCursorToIndex(p.Cursor),
			0, len(filtered))
		chunk := intClamp(nonZero(p.Limit, limit),
			1, dataGridSourceMaxPageLimit)
		start, end = s, intMin(len(filtered), s+chunk)
	case GridOffsetPageReq:
		start, end = dataGridSourceOffsetBounds(
			p.StartIndex, p.EndIndex, len(filtered), limit)
	default:
		start, end = 0, intMin(len(filtered), limit)
	}
	page := make([]GridRow, end-start)
	copy(page, filtered[start:end])
	if err := gridAbortCheck(req.Signal); err != nil {
		return GridDataResult{}, err
	}
	_, isCursor := req.Page.(GridCursorPageReq)
	rc := -1
	if rowCountKnown {
		rc = len(filtered)
	}
	var nextCursor, prevCursor string
	if isCursor && end < len(filtered) {
		nextCursor = dataGridSourceCursorFromIndex(end)
	}
	if isCursor {
		prevCursor = dataGridSourcePrevCursor(start, end-start)
	}
	return GridDataResult{
		Rows:          page,
		NextCursor:    nextCursor,
		PrevCursor:    prevCursor,
		RowCount:      rc,
		HasMore:       end < len(filtered),
		ReceivedCount: len(page),
	}, nil
}

func dataGridSourceInMemoryMutate(
	rows *[]GridRow, latencyMs int,
	rowCountKnown bool, req GridMutationRequest,
) (GridMutationResult, error) {
	if err := dataGridSourceSleepWithAbort(req.Signal, latencyMs); err != nil {
		return GridMutationResult{}, err
	}
	work := make([]GridRow, len(*rows))
	copy(work, *rows)
	result, err := dataGridSourceApplyMutation(
		&work, req.Kind, req.Rows, req.RowIDs, req.Edits)
	if err != nil {
		return GridMutationResult{}, err
	}
	if err := gridAbortCheck(req.Signal); err != nil {
		return GridMutationResult{}, err
	}
	*rows = work
	rc := -1
	if rowCountKnown {
		rc = len(*rows)
	}
	return GridMutationResult{
		Created:    result.created,
		Updated:    result.updated,
		DeletedIDs: result.deletedIDs,
		RowCount:   rc,
	}, nil
}

// dataGridSourceOffsetBounds clamps start/end to [0,total]
// and falls back to defaultLimit when the range is empty.
func dataGridSourceOffsetBounds(
	startIndex, endIndex, total, defaultLimit int,
) (int, int) {
	start := intClamp(startIndex, 0, total)
	end := intClamp(endIndex, start, total)
	if end <= start {
		end = intMin(total, start+defaultLimit)
	}
	return start, end
}

var errGridAborted = errors.New("grid: request aborted")

func gridAbortCheck(signal *GridAbortSignal) error {
	if signal.IsAborted() {
		return errGridAborted
	}
	return nil
}

func dataGridSourceSleepWithAbort(
	signal *GridAbortSignal, ms int,
) error {
	if ms <= 0 {
		return gridAbortCheck(signal)
	}
	remaining := ms
	for remaining > 0 {
		if err := gridAbortCheck(signal); err != nil {
			return err
		}
		step := intMin(remaining, 20)
		time.Sleep(time.Duration(step) * time.Millisecond)
		remaining -= step
	}
	return gridAbortCheck(signal)
}

func dataGridSourceCursorFromIndex(index int) string {
	return fmt.Sprintf("i:%d", intMax(0, index))
}

func dataGridSourcePrevCursor(start, pageSize int) string {
	if start <= 0 {
		return ""
	}
	return dataGridSourceCursorFromIndex(
		intMax(0, start-pageSize))
}

func dataGridSourceCursorToIndex(cursor string) int {
	idx, ok := dataGridSourceCursorToIndexOpt(cursor)
	if !ok {
		return 0
	}
	return idx
}

func dataGridSourceCursorToIndexOpt(cursor string) (int, bool) {
	trimmed := strings.TrimSpace(cursor)
	if trimmed == "" {
		return 0, true
	}
	if strings.HasPrefix(trimmed, "i:") {
		val := trimmed[2:]
		if !dataGridSourceIsDecimal(val) {
			return 0, false
		}
		n, _ := strconv.Atoi(val)
		return intMax(0, n), true
	}
	if !dataGridSourceIsDecimal(trimmed) {
		return 0, false
	}
	n, _ := strconv.Atoi(trimmed)
	return intMax(0, n), true
}

func dataGridSourceIsDecimal(input string) bool {
	if input == "" {
		return false
	}
	for i := 0; i < len(input); i++ {
		if input[i] < '0' || input[i] > '9' {
			return false
		}
	}
	return true
}

// dataGridSourceApplyQuery filters and sorts rows in memory.
func dataGridSourceApplyQuery(
	rows []GridRow, query GridQueryState,
) []GridRow {
	if query.QuickFilter == "" && len(query.Filters) == 0 &&
		len(query.Sorts) == 0 {
		return rows
	}
	hasFilters := query.QuickFilter != "" ||
		len(query.Filters) > 0
	var filtered []GridRow
	if hasFilters {
		needle := strings.ToLower(query.QuickFilter)
		lowered := make([]gridFilterLowered, len(query.Filters))
		for i, f := range query.Filters {
			lowered[i] = gridFilterLowered{
				colID: f.ColID,
				op:    f.Op,
				value: strings.ToLower(f.Value),
			}
		}
		filtered = make([]GridRow, 0, len(rows))
		for _, row := range rows {
			if dataGridSourceRowMatchesQuery(
				row, needle, lowered) {
				filtered = append(filtered, row)
			}
		}
	} else {
		filtered = rows
	}
	if len(query.Sorts) == 0 {
		return filtered
	}
	n := len(filtered)
	if n <= 1 {
		return filtered
	}
	idxs := make([]int, n)
	for i := range idxs {
		idxs[i] = i
	}
	if len(query.Sorts) == 1 {
		s0 := query.Sorts[0]
		keys := make([]string, n)
		for i, row := range filtered {
			keys[i] = row.Cells[s0.ColID]
		}
		dir := 1
		if s0.Dir == GridSortDesc {
			dir = -1
		}
		sort.Slice(idxs, func(a, b int) bool {
			ka, kb := keys[idxs[a]], keys[idxs[b]]
			if ka == kb {
				return idxs[a] < idxs[b]
			}
			if ka < kb {
				return dir > 0
			}
			return dir < 0
		})
	} else {
		sorts := query.Sorts
		keyCols := make([][]string, len(sorts))
		for si, s := range sorts {
			col := make([]string, n)
			for i, row := range filtered {
				col[i] = row.Cells[s.ColID]
			}
			keyCols[si] = col
		}
		sort.Slice(idxs, func(a, b int) bool {
			ia, ib := idxs[a], idxs[b]
			for si, s := range sorts {
				ka := keyCols[si][ia]
				kb := keyCols[si][ib]
				if ka == kb {
					continue
				}
				if s.Dir == GridSortAsc {
					return ka < kb
				}
				return ka > kb
			}
			return ia < ib
		})
	}
	result := make([]GridRow, n)
	for i, idx := range idxs {
		result[i] = filtered[idx]
	}
	return result
}

type gridFilterLowered struct {
	colID string
	op    string
	value string
}

func dataGridSourceRowMatchesQuery(
	row GridRow, needle string, filters []gridFilterLowered,
) bool {
	if needle != "" {
		matched := false
		for _, value := range row.Cells {
			if gridContainsLower(value, needle) {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}
	for _, f := range filters {
		cell := row.Cells[f.colID]
		var matched bool
		switch f.op {
		case "equals":
			matched = gridEqualsLower(cell, f.value)
		case "starts_with":
			matched = gridStartsWithLower(cell, f.value)
		case "ends_with":
			matched = gridEndsWithLower(cell, f.value)
		default:
			matched = gridContainsLower(cell, f.value)
		}
		if !matched {
			return false
		}
	}
	return true
}

// gridLowerByte returns ASCII lowercase (a-z, A-Z only).
func gridLowerByte(c byte) byte {
	if c >= 'A' && c <= 'Z' {
		return c | 0x20
	}
	return c
}

// gridContainsLower checks haystack.toLower().contains(needle)
// without allocating. needle must already be lowered.
func gridContainsLower(haystack, needle string) bool {
	if len(needle) == 0 {
		return true
	}
	if len(haystack) < len(needle) {
		return false
	}
	limit := len(haystack) - len(needle)
	for i := 0; i <= limit; i++ {
		found := true
		for j := 0; j < len(needle); j++ {
			if gridLowerByte(haystack[i+j]) != needle[j] {
				found = false
				break
			}
		}
		if found {
			return true
		}
	}
	return false
}

// gridEqualsLower checks haystack.toLower() == needle
// without allocating. needle must already be lowered.
func gridEqualsLower(haystack, needle string) bool {
	if len(haystack) != len(needle) {
		return false
	}
	for i := 0; i < len(haystack); i++ {
		if gridLowerByte(haystack[i]) != needle[i] {
			return false
		}
	}
	return true
}

// gridStartsWithLower checks haystack.toLower().hasPrefix(needle)
// without allocating. needle must already be lowered.
func gridStartsWithLower(haystack, needle string) bool {
	if len(haystack) < len(needle) {
		return false
	}
	for i := 0; i < len(needle); i++ {
		if gridLowerByte(haystack[i]) != needle[i] {
			return false
		}
	}
	return true
}

// gridEndsWithLower checks haystack.toLower().hasSuffix(needle)
// without allocating. needle must already be lowered.
func gridEndsWithLower(haystack, needle string) bool {
	if len(haystack) < len(needle) {
		return false
	}
	off := len(haystack) - len(needle)
	for i := 0; i < len(needle); i++ {
		if gridLowerByte(haystack[i+off]) != needle[i] {
			return false
		}
	}
	return true
}

// GridQuerySignature returns a stable FNV-1a 64-bit hash of
// the query state. Filters are sorted by col_id for order
// independence.
func GridQuerySignature(query GridQueryState) uint64 {
	h := dataGridFnv64Offset
	h = dataGridFnv64Str(h, query.QuickFilter)
	h = dataGridFnv64Byte(h, '|')
	h = dataGridFnv64Byte(h, 's')
	for _, s := range query.Sorts {
		h = dataGridFnv64Byte(h, 0x1e)
		h = dataGridFnv64Str(h, s.ColID)
		if s.Dir == GridSortDesc {
			h = dataGridFnv64Byte(h, 'd')
		} else {
			h = dataGridFnv64Byte(h, 'a')
		}
	}
	h = dataGridFnv64Byte(h, '|')
	h = dataGridFnv64Byte(h, 'f')
	filters := query.Filters
	if len(filters) <= 1 {
		for _, f := range filters {
			h = gridHashFilter(h, f)
		}
		return h
	}
	idxs := make([]int, len(filters))
	for i := range idxs {
		idxs[i] = i
	}
	sort.Slice(idxs, func(a, b int) bool {
		fa, fb := filters[idxs[a]], filters[idxs[b]]
		if fa.ColID != fb.ColID {
			return fa.ColID < fb.ColID
		}
		if fa.Op != fb.Op {
			return fa.Op < fb.Op
		}
		return fa.Value < fb.Value
	})
	for _, i := range idxs {
		h = gridHashFilter(h, filters[i])
	}
	return h
}

func dataGridFnv64Str(h uint64, s string) uint64 {
	for i := 0; i < len(s); i++ {
		h = (h ^ uint64(s[i])) * dataGridFnv64Prime
	}
	return h
}

func dataGridFnv64Byte(h uint64, b byte) uint64 {
	return (h ^ uint64(b)) * dataGridFnv64Prime
}

func gridHashFilter(h uint64, f GridFilter) uint64 {
	hash := dataGridFnv64Byte(h, 0x1e)
	hash = dataGridFnv64Str(hash, f.ColID)
	hash = dataGridFnv64Byte(hash, 0x1f)
	hash = dataGridFnv64Str(hash, f.Op)
	hash = dataGridFnv64Byte(hash, 0x1f)
	hash = dataGridFnv64Str(hash, f.Value)
	return hash
}

type gridMutationApplyResult struct {
	created    []GridRow
	updated    []GridRow
	deletedIDs []string
}

func dataGridSourceApplyMutation(
	rows *[]GridRow, kind GridMutationKind,
	reqRows []GridRow, reqRowIDs []string,
	edits []GridCellEdit,
) (gridMutationApplyResult, error) {
	switch kind {
	case GridMutationCreate:
		return dataGridSourceApplyCreate(rows, reqRows)
	case GridMutationUpdate:
		return dataGridSourceApplyUpdate(rows, reqRows, edits)
	case GridMutationDelete:
		return dataGridSourceApplyDelete(rows, reqRows, reqRowIDs)
	}
	return gridMutationApplyResult{}, errors.New(
		"grid: unknown mutation kind")
}

func dataGridSourceApplyCreate(
	rows *[]GridRow, reqRows []GridRow,
) (gridMutationApplyResult, error) {
	if len(reqRows) == 0 {
		return gridMutationApplyResult{}, nil
	}
	existing := make(map[string]bool, len(*rows))
	for idx, row := range *rows {
		existing[dataGridRowID(row, idx)] = true
	}
	created := make([]GridRow, 0, len(reqRows))
	for _, row := range reqRows {
		nextID, err := dataGridSourceNextCreateRowID(
			*rows, existing, row.ID)
		if err != nil {
			return gridMutationApplyResult{}, err
		}
		cells := make(map[string]string, len(row.Cells))
		for k, v := range row.Cells {
			cells[k] = v
		}
		nextRow := GridRow{ID: nextID, Cells: cells}
		*rows = append(*rows, nextRow)
		existing[nextID] = true
		created = append(created, nextRow)
	}
	return gridMutationApplyResult{created: created}, nil
}

func dataGridSourceApplyUpdate(
	rows *[]GridRow, reqRows []GridRow, edits []GridCellEdit,
) (gridMutationApplyResult, error) {
	updatedIDs := make(map[string]bool)
	editsByRow := make(map[string][]GridCellEdit)
	for _, edit := range edits {
		if edit.RowID == "" {
			return gridMutationApplyResult{},
				errors.New("grid: row id is required")
		}
		if edit.ColID == "" {
			return gridMutationApplyResult{},
				errors.New("grid: edit has empty col id")
		}
		editsByRow[edit.RowID] = append(
			editsByRow[edit.RowID], edit)
	}
	updated := make([]GridRow, 0,
		len(reqRows)+len(editsByRow))
	rowIdx := make(map[string]int, len(*rows))
	for idx, row := range *rows {
		rowIdx[dataGridRowID(row, idx)] = idx
	}
	for _, reqRow := range reqRows {
		if reqRow.ID == "" {
			return gridMutationApplyResult{},
				errors.New("grid: row id is required")
		}
		idx, ok := rowIdx[reqRow.ID]
		if !ok {
			return gridMutationApplyResult{},
				fmt.Errorf("grid: update row not found: %s",
					reqRow.ID)
		}
		cells := make(map[string]string,
			len((*rows)[idx].Cells))
		for k, v := range (*rows)[idx].Cells {
			cells[k] = v
		}
		for k, v := range reqRow.Cells {
			cells[k] = v
		}
		if rowEdits, ok := editsByRow[reqRow.ID]; ok {
			for _, edit := range rowEdits {
				cells[edit.ColID] = edit.Value
			}
		}
		(*rows)[idx] = GridRow{ID: (*rows)[idx].ID, Cells: cells}
		updated = append(updated, (*rows)[idx])
		updatedIDs[reqRow.ID] = true
	}
	pendingIDs := make([]string, 0, len(editsByRow))
	for rowID := range editsByRow {
		if updatedIDs[rowID] {
			continue
		}
		pendingIDs = append(pendingIDs, rowID)
	}
	sort.Strings(pendingIDs)
	for _, rowID := range pendingIDs {
		rowEdits := editsByRow[rowID]
		if updatedIDs[rowID] {
			continue
		}
		idx, ok := rowIdx[rowID]
		if !ok {
			return gridMutationApplyResult{},
				fmt.Errorf("grid: edit row not found: %s", rowID)
		}
		cells := make(map[string]string,
			len((*rows)[idx].Cells))
		for k, v := range (*rows)[idx].Cells {
			cells[k] = v
		}
		for _, edit := range rowEdits {
			cells[edit.ColID] = edit.Value
		}
		(*rows)[idx] = GridRow{
			ID: (*rows)[idx].ID, Cells: cells,
		}
		updated = append(updated, (*rows)[idx])
	}
	return gridMutationApplyResult{updated: updated}, nil
}

func dataGridSourceApplyDelete(
	rows *[]GridRow, reqRows []GridRow, reqRowIDs []string,
) (gridMutationApplyResult, error) {
	idSet := gridDeduplicateRowIDs(reqRows, reqRowIDs)
	if len(idSet) == 0 {
		return gridMutationApplyResult{}, nil
	}
	kept := make([]GridRow, 0, len(*rows))
	deletedIDs := make([]string, 0, len(idSet))
	for idx, row := range *rows {
		rowID := dataGridRowID(row, idx)
		if idSet[rowID] {
			deletedIDs = append(deletedIDs, rowID)
			continue
		}
		kept = append(kept, row)
	}
	*rows = kept
	return gridMutationApplyResult{deletedIDs: deletedIDs}, nil
}

// gridDeduplicateRowIDs collects unique non-empty IDs from
// GridRow.ID values and raw ID strings.
func gridDeduplicateRowIDs(
	rows []GridRow, rowIDs []string,
) map[string]bool {
	seen := make(map[string]bool)
	for _, row := range rows {
		if row.ID != "" {
			seen[row.ID] = true
		}
	}
	for _, rowID := range rowIDs {
		id := strings.TrimSpace(rowID)
		if id != "" {
			seen[id] = true
		}
	}
	return seen
}

// dataGridRowID is in view_data_grid.go (auto-hash fallback).

func dataGridSourceNextCreateRowID(
	rows []GridRow, existing map[string]bool, preferredID string,
) (string, error) {
	id := strings.TrimSpace(preferredID)
	if id != "" && !existing[id] {
		return id, nil
	}
	cap := len(rows) + 1000
	next := len(rows) + 1
	for next <= cap {
		candidate := strconv.Itoa(next)
		if !existing[candidate] {
			return candidate, nil
		}
		next++
	}
	// Numeric range exhausted; try random hex IDs.
	for range 10 {
		var buf [8]byte
		if _, err := rand.Read(buf[:]); err != nil {
			return "", fmt.Errorf(
				"grid: random id generation failed: %w", err)
		}
		candidate := fmt.Sprintf("__gen_%016x",
			uint64(buf[0])<<56|uint64(buf[1])<<48|
				uint64(buf[2])<<40|uint64(buf[3])<<32|
				uint64(buf[4])<<24|uint64(buf[5])<<16|
				uint64(buf[6])<<8|uint64(buf[7]))
		if !existing[candidate] {
			return candidate, nil
		}
	}
	return "", errors.New("grid: unable to generate unique row id")
}

func nonZero(v, fallback int) int {
	if v > 0 {
		return v
	}
	return fallback
}
