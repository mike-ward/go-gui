package gui

import "unicode/utf8"

const inputMaxInsertRunes = 65_536

// InputState manages cursor, selection, and undo/redo for
// an input field. Stored in StateRegistry keyed by IDFocus.
type InputState struct {
	CursorPos    int
	SelectBeg    uint32
	SelectEnd    uint32
	Undo         *BoundedStack[InputMemento]
	Redo         *BoundedStack[InputMemento]
	CursorOffset float32
}

// InputMemento stores a snapshot for undo/redo.
type InputMemento struct {
	Text         string
	CursorPos    int
	SelectBeg    uint32
	SelectEnd    uint32
	CursorOffset float32
}

// InputMode selects single-line or multiline behavior.
type InputMode uint8

const (
	InputSingleLine InputMode = iota
	InputMultiline
)

// InputCommitReason identifies why text was committed.
type InputCommitReason uint8

const (
	CommitEnter InputCommitReason = iota
	CommitBlur
)

const undoMaxSize = 50

func inputStateOrDefault(idFocus uint32, w *Window) InputState {
	m := StateMap[uint32, InputState](w, nsInput, capMany)
	if v, ok := m.Get(idFocus); ok {
		return v
	}
	return InputState{}
}

func inputMementoFromState(text string, is InputState) InputMemento {
	return InputMemento{
		Text:         text,
		CursorPos:    is.CursorPos,
		SelectBeg:    is.SelectBeg,
		SelectEnd:    is.SelectEnd,
		CursorOffset: is.CursorOffset,
	}
}

func inputPushUndo(is InputState, text string) *BoundedStack[InputMemento] {
	stack := is.Undo
	if stack == nil {
		stack = NewBoundedStack[InputMemento](undoMaxSize)
	}
	stack.Push(inputMementoFromState(text, is))
	return stack
}

func inputStateFromMemento(m InputMemento, undo, redo *BoundedStack[InputMemento]) InputState {
	return InputState{
		CursorPos:    m.CursorPos,
		SelectBeg:    m.SelectBeg,
		SelectEnd:    m.SelectEnd,
		CursorOffset: m.CursorOffset,
		Undo:         undo,
		Redo:         redo,
	}
}

// inputInsert inserts text at cursor or replaces selection.
// Returns resulting text.
func inputInsert(text string, insertText string, idFocus uint32, w *Window) string {
	if len(insertText) == 0 {
		return text
	}
	insertRunes := []rune(insertText)
	if len(insertRunes) > inputMaxInsertRunes {
		insertRunes = insertRunes[:inputMaxInsertRunes]
	}

	runes := []rune(text)
	is := inputStateOrDefault(idFocus, w)
	cursorPos := min(is.CursorPos, len(runes))
	if cursorPos < 0 {
		runes = append([]rune(text), insertRunes...)
		cursorPos = len(runes)
	} else if is.SelectBeg != is.SelectEnd {
		beg, end := u32Sort(is.SelectBeg, is.SelectEnd)
		if int(beg) >= len(runes) || int(end) > len(runes) {
			return text
		}
		result := make([]rune, 0, int(beg)+len(insertRunes)+(len(runes)-int(end)))
		result = append(result, runes[:beg]...)
		result = append(result, insertRunes...)
		result = append(result, runes[end:]...)
		runes = result
		cursorPos = min(int(beg)+len(insertRunes), len(runes))
	} else {
		result := make([]rune, 0, cursorPos+len(insertRunes)+(len(runes)-cursorPos))
		result = append(result, runes[:cursorPos]...)
		result = append(result, insertRunes...)
		result = append(result, runes[cursorPos:]...)
		runes = result
		cursorPos = min(cursorPos+len(insertRunes), len(runes))
	}

	nextText := string(runes)
	undo := inputPushUndo(is, text)
	imap := StateMap[uint32, InputState](w, nsInput, capMany)
	imap.Set(idFocus, InputState{
		CursorPos:    cursorPos,
		CursorOffset: -1,
		Undo:         undo,
	})
	return nextText
}

// inputDelete removes text at cursor or selected range.
// forwardDelete=true for Delete key, false for Backspace.
func inputDelete(text string, idFocus uint32, forwardDelete bool, w *Window) (string, bool) {
	runes := []rune(text)
	is := inputStateOrDefault(idFocus, w)
	cursorPos := min(is.CursorPos, len(runes))
	if cursorPos < 0 {
		cursorPos = len(runes)
	}

	if is.SelectBeg != is.SelectEnd {
		beg, end := u32Sort(is.SelectBeg, is.SelectEnd)
		if int(beg) >= len(runes) || int(end) > len(runes) {
			return text, false
		}
		result := make([]rune, 0, int(beg)+(len(runes)-int(end)))
		result = append(result, runes[:beg]...)
		result = append(result, runes[end:]...)
		runes = result
		cursorPos = min(int(beg), len(runes))
	} else {
		if cursorPos == 0 && !forwardDelete {
			return text, true
		}
		if cursorPos == len(runes) && forwardDelete {
			return text, true
		}
		delPos := cursorPos
		if !forwardDelete {
			delPos = cursorPos - 1
		}
		if delPos < 0 || delPos >= len(runes) {
			return text, false
		}
		result := make([]rune, 0, len(runes)-1)
		result = append(result, runes[:delPos]...)
		result = append(result, runes[delPos+1:]...)
		runes = result
		if !forwardDelete {
			cursorPos--
		}
	}

	nextText := string(runes)
	undo := inputPushUndo(is, text)
	imap := StateMap[uint32, InputState](w, nsInput, capMany)
	imap.Set(idFocus, InputState{
		CursorPos:    cursorPos,
		CursorOffset: -1,
		Undo:         undo,
	})
	return nextText, true
}

// inputCopy returns the selected text. Returns ("", false) if
// no selection or password mode.
func inputCopy(text string, idFocus uint32, isPassword bool, w *Window) (string, bool) {
	if isPassword {
		return "", false
	}
	is := StateReadOr(w, nsInput, idFocus, InputState{})
	if is.SelectBeg == is.SelectEnd {
		return "", false
	}
	beg, end := u32Sort(is.SelectBeg, is.SelectEnd)
	runeCount := utf8.RuneCountInString(text)
	if int(beg) > runeCount || int(end) > runeCount || beg >= end {
		return "", false
	}
	runes := []rune(text)
	return string(runes[beg:end]), true
}

// inputCut copies selected text then deletes it.
func inputCut(text string, idFocus uint32, isPassword bool, w *Window) (string, string, bool) {
	if isPassword {
		return text, "", false
	}
	copied, ok := inputCopy(text, idFocus, false, w)
	if !ok {
		return text, "", false
	}
	newText, _ := inputDelete(text, idFocus, false, w)
	return newText, copied, true
}

// inputUndo reverts to previous state. Returns restored text.
func inputUndo(text string, idFocus uint32, w *Window) string {
	imap := StateMap[uint32, InputState](w, nsInput, capMany)
	is, _ := imap.Get(idFocus)
	if is.Undo == nil || is.Undo.IsEmpty() {
		return text
	}
	memento, ok := is.Undo.Pop()
	if !ok {
		return text
	}
	redo := is.Redo
	if redo == nil {
		redo = NewBoundedStack[InputMemento](undoMaxSize)
	}
	redo.Push(inputMementoFromState(text, is))
	imap.Set(idFocus, inputStateFromMemento(memento, is.Undo, redo))
	return memento.Text
}

// inputRedo reapplies a previously undone operation.
func inputRedo(text string, idFocus uint32, w *Window) string {
	imap := StateMap[uint32, InputState](w, nsInput, capMany)
	is, _ := imap.Get(idFocus)
	if is.Redo == nil || is.Redo.IsEmpty() {
		return text
	}
	memento, ok := is.Redo.Pop()
	if !ok {
		return text
	}
	undo := is.Undo
	if undo == nil {
		undo = NewBoundedStack[InputMemento](undoMaxSize)
	}
	undo.Push(inputMementoFromState(text, is))
	imap.Set(idFocus, inputStateFromMemento(memento, undo, is.Redo))
	return memento.Text
}

// inputSelectAll selects all text.
func inputSelectAll(text string, idFocus uint32, w *Window) {
	runeCount := utf8.RuneCountInString(text)
	imap := StateMap[uint32, InputState](w, nsInput, capMany)
	is, _ := imap.Get(idFocus)
	is.SelectBeg = 0
	is.SelectEnd = uint32(runeCount)
	is.CursorPos = runeCount
	imap.Set(idFocus, is)
}

// inputHasSelection returns true if text is selected.
func inputHasSelection(idFocus uint32, w *Window) bool {
	is := StateReadOr(w, nsInput, idFocus, InputState{})
	return is.SelectBeg != is.SelectEnd
}

// inputSelectedText returns the selected text.
func inputSelectedText(text string, idFocus uint32, w *Window) string {
	is := StateReadOr(w, nsInput, idFocus, InputState{})
	if is.SelectBeg == is.SelectEnd {
		return ""
	}
	beg, end := u32Sort(is.SelectBeg, is.SelectEnd)
	runes := []rune(text)
	if int(beg) >= len(runes) || int(end) > len(runes) {
		return ""
	}
	return string(runes[beg:end])
}
