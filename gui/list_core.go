package gui

// list_core.go provides pure functions shared by ListBox, Select,
// and other list widgets. No state, no Window dependency.

const listCoreVirtualBufferRows = 2

// ListCoreAction represents a keyboard navigation action.
type ListCoreAction uint8

const (
	ListCoreNone       ListCoreAction = iota
	ListCoreMoveUp
	ListCoreMoveDown
	ListCoreSelectItem
	ListCoreDismiss
	ListCoreFirst
	ListCoreLast
)

// listCoreNavigate maps a key code to a list navigation action.
func listCoreNavigate(key KeyCode, itemCount int) ListCoreAction {
	if itemCount == 0 {
		return ListCoreNone
	}
	switch key {
	case KeyUp:
		return ListCoreMoveUp
	case KeyDown:
		return ListCoreMoveDown
	case KeyEnter:
		return ListCoreSelectItem
	case KeyEscape:
		return ListCoreDismiss
	case KeyHome:
		return ListCoreFirst
	case KeyEnd:
		return ListCoreLast
	default:
		return ListCoreNone
	}
}

// listCoreApplyNav applies a navigation action to a highlight
// index. Returns new index and whether it changed.
func listCoreApplyNav(action ListCoreAction, cur, itemCount int) (int, bool) {
	switch action {
	case ListCoreMoveUp:
		next := cur - 1
		if next < 0 {
			next = 0
		}
		return next, next != cur
	case ListCoreMoveDown:
		next := cur + 1
		if next >= itemCount {
			next = itemCount - 1
		}
		return next, next != cur
	case ListCoreFirst:
		return 0, cur != 0
	case ListCoreLast:
		last := itemCount - 1
		if last < 0 {
			last = 0
		}
		return last, cur != last
	default:
		return cur, false
	}
}

// listCoreVisibleRange computes the visible index range from
// scroll offset. Pure arithmetic.
func listCoreVisibleRange(itemCount int, rowHeight, listHeight, scrollY float32) (int, int) {
	if itemCount == 0 || rowHeight <= 0 || listHeight <= 0 {
		return 0, -1
	}
	maxIdx := itemCount - 1
	absScroll := scrollY
	if absScroll < 0 {
		absScroll = -absScroll
	}
	first := intClamp(int(absScroll/rowHeight), 0, maxIdx)
	visibleRows := int(listHeight/rowHeight) + 1
	buf := listCoreVirtualBufferRows
	firstVisible := intMax(0, first-buf)
	lastVisible := min(maxIdx, first+visibleRows+buf)
	if firstVisible > lastVisible {
		lastVisible = firstVisible
	}
	return firstVisible, lastVisible
}

// listCoreRowHeightEstimate estimates row height from text
// style + padding. No Window needed.
func listCoreRowHeightEstimate(style TextStyle, pad Padding) float32 {
	return style.Size + pad.Height()
}

// listCoreFuzzyScore scores a candidate against a query.
// Returns -1 (no match) or 0+ (lower = better).
func listCoreFuzzyScore(candidate, query string) int {
	if len(query) == 0 {
		return 0
	}
	if len(candidate) == 0 {
		return -1
	}
	qi := 0
	score := 0
	prevMatch := -1
	for ci := 0; ci < len(candidate); ci++ {
		if qi >= len(query) {
			break
		}
		cb := toLowerByte(candidate[ci])
		qb := toLowerByte(query[qi])
		if cb == qb {
			if prevMatch >= 0 {
				gap := ci - prevMatch - 1
				score += gap
			}
			prevMatch = ci
			qi++
		}
	}
	if qi < len(query) {
		return -1
	}
	return score
}

func toLowerByte(b byte) byte {
	if b >= 0x41 && b <= 0x5A {
		return b + 32
	}
	return b
}

// listBoxNextSelectedIDs computes the next selection set after
// toggling datID.
func listBoxNextSelectedIDs(selectedIDs []string, datID string, isMultiple bool) []string {
	if !isMultiple {
		return []string{datID}
	}
	for _, id := range selectedIDs {
		if id == datID {
			// Remove it.
			next := make([]string, 0, len(selectedIDs)-1)
			for _, id2 := range selectedIDs {
				if id2 != datID {
					next = append(next, id2)
				}
			}
			return next
		}
	}
	// Add it.
	next := make([]string, 0, len(selectedIDs)+1)
	next = append(next, selectedIDs...)
	next = append(next, datID)
	return next
}
