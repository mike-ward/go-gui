package gui

import "testing"

// --- cursor_right / cursor_end ---

func TestCursorRightClamp3Byte(t *testing.T) {
	text := "€" // 3 bytes, 1 rune
	assertEqual(t, cursorRight(text, 0), 1)
	assertEqual(t, cursorRight(text, 1), 1) // clamped
}

func TestCursorRightClamp4Byte(t *testing.T) {
	text := "𝐀" // U+1D400, 4 bytes, 1 rune
	assertEqual(t, cursorRight(text, 0), 1)
	assertEqual(t, cursorRight(text, 1), 1)
}

func TestCursorRightCJK(t *testing.T) {
	text := "日本語" // 3 runes, 9 bytes
	assertEqual(t, cursorRight(text, 2), 3)
	assertEqual(t, cursorRight(text, 3), 3)
}

func TestCursorRightEmptyText(t *testing.T) {
	assertEqual(t, cursorRight("", 0), 0)
}

func TestCursorEnd3Byte(t *testing.T) {
	assertEqual(t, cursorEnd("€€€"), 3)
}

func TestCursorEnd4Byte(t *testing.T) {
	assertEqual(t, cursorEnd("𝐀𝐁𝐂"), 3)
}

func TestCursorEndEmojiOnly(t *testing.T) {
	assertEqual(t, cursorEnd("😀😀"), 2)
}

func TestCursorEndEmpty(t *testing.T) {
	assertEqual(t, cursorEnd(""), 0)
}

func TestCursorEndCombiningChar(t *testing.T) {
	// e + combining acute = 2 runes
	assertEqual(t, cursorEnd("e\u0301"), 2)
}

func TestCursorEndSingle4Byte(t *testing.T) {
	assertEqual(t, cursorEnd("😀"), 1)
}

// --- cursor_left ---

func TestCursorLeftAtZero(t *testing.T) {
	assertEqual(t, cursorLeft(0), 0)
}

func TestCursorLeftFromOne(t *testing.T) {
	assertEqual(t, cursorLeft(1), 0)
}

// --- word boundaries ---

func TestCursorEndOfWordCJKMixed(t *testing.T) {
	text := "hello 日本語 world"
	// From pos 5 (space): skip space, then non-spaces → end of 日本語
	assertEqual(t, cursorEndOfWord(text, 5), 9)
}

func TestCursorStartOfWordCJKMixed(t *testing.T) {
	text := "hello 日本語 world"
	// From pos 9 (space after 語): skip space back, then
	// non-spaces back → start of 日本語 at rune 6
	assertEqual(t, cursorStartOfWord(text, 9), 6)
}

// --- paragraph navigation ---

func TestCursorStartOfParagraphMidLine(t *testing.T) {
	text := "abc\ndef\nghi"
	// pos 5 is in "def" → start of "def" is after first \n
	assertEqual(t, cursorStartOfParagraph(text, 5), 4)
}

func TestCursorStartOfParagraphBeginning(t *testing.T) {
	text := "hello world"
	assertEqual(t, cursorStartOfParagraph(text, 3), 0)
}

func TestCursorEndOfParagraphMidLine(t *testing.T) {
	text := "abc\ndef\nghi"
	// pos 1 → find \n at byte 3 → rune 3
	assertEqual(t, cursorEndOfParagraph(text, 1), 3)
}

func TestCursorEndOfParagraphLastLine(t *testing.T) {
	text := "abc\ndef"
	assertEqual(t, cursorEndOfParagraph(text, 5), 7)
}

// --- helpers ---

func TestRuneToByteIndex(t *testing.T) {
	text := "a€b" // a=1byte, €=3bytes, b=1byte
	assertEqual(t, runeToByteIndex(text, 0), 0)
	assertEqual(t, runeToByteIndex(text, 1), 1) // after 'a'
	assertEqual(t, runeToByteIndex(text, 2), 4) // after '€'
}

func TestByteToRuneIndex(t *testing.T) {
	text := "a€b"
	assertEqual(t, byteToRuneIndex(text, 0), 0)
	assertEqual(t, byteToRuneIndex(text, 1), 1)
	assertEqual(t, byteToRuneIndex(text, 4), 2)
}

func TestSelectionRange(t *testing.T) {
	beg, end := selectionRange(5, 2)
	assertEqual(t, int(beg), 2)
	assertEqual(t, int(end), 5)
}

func assertEqual(t *testing.T, got, want int) {
	t.Helper()
	if got != want {
		t.Errorf("got %d, want %d", got, want)
	}
}
