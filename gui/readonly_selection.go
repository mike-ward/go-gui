package gui

import (
	"strings"
)

// ReadOnlySelectionState represents text selection in a read-only widget
// (e.g., Markdown or RTF view). Similar to InputState but without editing.
type ReadOnlySelectionState struct {
	SelectBeg uint32  // selection start (rune index)
	SelectEnd uint32  // selection end (rune index)
	CursorPos int     // cursor position for keyboard navigation
}

// rtfSelectedText extracts the selected text from a RichText
// by concatenating the Text fields of selected runs.
func rtfSelectedText(rt *RichText, beg, end uint32) string {
	if rt == nil || beg >= end {
		return ""
	}
	var b strings.Builder
	runeIdx := uint32(0)
	for _, run := range rt.Runs {
		runLen := uint32(len([]rune(run.Text)))
		runEnd := runeIdx + runLen

		// Check if this run overlaps with selection.
		if runEnd > beg && runeIdx < end {
			// Find the substring within this run.
			runBeg := beg - runeIdx
			runEnd := end - runeIdx
			if runBeg < 0 {
				runBeg = 0
			}
			if runEnd > runLen {
				runEnd = runLen
			}
			runes := []rune(run.Text)
			b.WriteString(string(runes[runBeg:runEnd]))
		}
		runeIdx = runeIdx + runLen
	}
	return b.String()
}

// markdownSelectedText extracts selected text from a flattened markdown string.
// Since markdown is rendered as nested RTF widgets, this works on the complete
// text representation of all blocks concatenated.
func markdownSelectedText(blocks []MarkdownBlock, beg, end uint32) string {
	if beg >= end {
		return ""
	}

	// Build complete text from all blocks (including line breaks between blocks).
	var allText strings.Builder
	for i, block := range blocks {
		blockText := richTextPlain(block.Content)
		allText.WriteString(blockText)
		if i < len(blocks)-1 {
			allText.WriteString("\n")
		}
	}

	text := allText.String()
	runes := []rune(text)
	if int(beg) >= len(runes) || int(end) > len(runes) {
		return ""
	}
	return string(runes[beg:end])
}

// richTextPlain extracts plain text from RichText by concatenating all runs.
func richTextPlain(rt RichText) string {
	var b strings.Builder
	for _, run := range rt.Runs {
		b.WriteString(run.Text)
	}
	return b.String()
}

// readOnlySelectionCopy copies the selected text to clipboard for read-only widgets.
func readOnlySelectionCopy(idFocus uint32, w *Window, getText func(beg, end uint32) string) bool {
	is := StateReadOr(w, nsInput, idFocus, InputState{})
	if is.SelectBeg == is.SelectEnd {
		return false
	}
	beg, end := u32Sort(is.SelectBeg, is.SelectEnd)
	text := getText(beg, end)
	if text == "" {
		return false
	}
	w.SetClipboard(text)
	return true
}

// readOnlySelectAll selects all text in a read-only widget.
func readOnlySelectAll(idFocus uint32, totalRunes int, w *Window) {
	imap := StateMap[uint32, InputState](w, nsInput, capMany)
	is, _ := imap.Get(idFocus)
	is.SelectBeg = 0
	is.SelectEnd = uint32(totalRunes)
	is.CursorPos = totalRunes
	imap.Set(idFocus, is)
}

// readOnlyUpdateSelection updates selection and cursor position for a read-only widget.
func readOnlyUpdateSelection(
	imap *BoundedMap[uint32, InputState],
	idFocus uint32,
	is InputState,
	newPos int,
	isShift bool,
) {
	if isShift {
		if is.SelectBeg == is.SelectEnd {
			is.SelectBeg = uint32(is.CursorPos)
			is.SelectEnd = uint32(newPos)
		} else {
			if uint32(is.CursorPos) == is.SelectEnd {
				is.SelectEnd = uint32(newPos)
			} else {
				is.SelectBeg = uint32(newPos)
			}
		}
	} else {
		is.SelectBeg = 0
		is.SelectEnd = 0
	}
	is.CursorPos = newPos
	imap.Set(idFocus, is)
}
