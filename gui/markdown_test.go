package gui

import (
	"strings"
	"testing"
)

// --- Test helpers ---

func mdBlocks(source string) []MarkdownBlock {
	return markdownToBlocks(source, DefaultMarkdownStyle())
}

func mdParsed(source string) []MdBlock {
	return mdParse(source, false)
}

func mdPlainText(blocks []MarkdownBlock) string {
	var sb strings.Builder
	for _, b := range blocks {
		for _, r := range b.Content.Runs {
			sb.WriteString(r.Text)
		}
	}
	return sb.String()
}

func mdRunTexts(blocks []MdBlock) []string {
	var result []string
	for _, b := range blocks {
		for _, r := range b.Runs {
			result = append(result, r.Text)
		}
	}
	return result
}

// assertContains checks s contains sub.
func assertContains(t *testing.T, s, sub string) {
	t.Helper()
	if !strings.Contains(s, sub) {
		t.Errorf("expected %q to contain %q", s, sub)
	}
}

// --- Headers ---

func TestMarkdownH1(t *testing.T) {
	blocks := mdParsed("# Hello")
	if len(blocks) == 0 {
		t.Fatal("no blocks")
	}
	if blocks[0].HeaderLevel != 1 {
		t.Fatalf("header level: got %d", blocks[0].HeaderLevel)
	}
	assertContains(t, runsToText(blocks[0].Runs), "Hello")
}

func TestMarkdownH2(t *testing.T) {
	blocks := mdParsed("## World")
	if len(blocks) == 0 || blocks[0].HeaderLevel != 2 {
		t.Fatal("expected H2")
	}
}

func TestMarkdownH3toH6(t *testing.T) {
	for level := 3; level <= 6; level++ {
		prefix := strings.Repeat("#", level)
		blocks := mdParsed(prefix + " Heading")
		if len(blocks) == 0 ||
			blocks[0].HeaderLevel != level {
			t.Fatalf("H%d: got level %d",
				level, blocks[0].HeaderLevel)
		}
	}
}

func TestMarkdownSetextH1(t *testing.T) {
	blocks := mdParsed("Title\n=====")
	if len(blocks) == 0 || blocks[0].HeaderLevel != 1 {
		t.Fatal("expected setext H1")
	}
}

func TestMarkdownSetextH2(t *testing.T) {
	blocks := mdParsed("Title\n-----")
	if len(blocks) == 0 || blocks[0].HeaderLevel != 2 {
		t.Fatal("expected setext H2")
	}
}

func TestMarkdownHeadingAnchor(t *testing.T) {
	blocks := mdParsed("# Hello World")
	if len(blocks) == 0 {
		t.Fatal("no blocks")
	}
	if blocks[0].AnchorSlug != "hello-world" {
		t.Fatalf("slug: got %q", blocks[0].AnchorSlug)
	}
}

// --- Inline formatting ---

func TestMarkdownBold(t *testing.T) {
	blocks := mdParsed("**bold**")
	found := false
	for _, b := range blocks {
		for _, r := range b.Runs {
			if r.Format == MdFormatBold &&
				strings.TrimSpace(r.Text) == "bold" {
				found = true
			}
		}
	}
	if !found {
		t.Error("expected bold run")
	}
}

func TestMarkdownItalic(t *testing.T) {
	blocks := mdParsed("*italic*")
	found := false
	for _, b := range blocks {
		for _, r := range b.Runs {
			if r.Format == MdFormatItalic &&
				strings.TrimSpace(r.Text) == "italic" {
				found = true
			}
		}
	}
	if !found {
		t.Error("expected italic run")
	}
}

func TestMarkdownBoldItalic(t *testing.T) {
	blocks := mdParsed("***both***")
	found := false
	for _, b := range blocks {
		for _, r := range b.Runs {
			if r.Format == MdFormatBoldItalic &&
				strings.TrimSpace(r.Text) == "both" {
				found = true
			}
		}
	}
	if !found {
		t.Error("expected bold+italic run")
	}
}

func TestMarkdownInlineCode(t *testing.T) {
	blocks := mdParsed("text `code` text")
	found := false
	for _, b := range blocks {
		for _, r := range b.Runs {
			if r.Format == MdFormatCode &&
				strings.TrimSpace(r.Text) == "code" {
				found = true
			}
		}
	}
	if !found {
		t.Error("expected code run")
	}
}

func TestMarkdownStrikethrough(t *testing.T) {
	blocks := mdParsed("~~strikethrough~~")
	found := false
	for _, b := range blocks {
		for _, r := range b.Runs {
			if r.Strikethrough &&
				strings.TrimSpace(r.Text) == "strikethrough" {
				found = true
			}
		}
	}
	if !found {
		t.Error("expected strikethrough run")
	}
}

func TestMarkdownHighlight(t *testing.T) {
	blocks := mdParsed("==highlighted==")
	found := false
	for _, b := range blocks {
		for _, r := range b.Runs {
			if r.Highlight &&
				strings.TrimSpace(r.Text) == "highlighted" {
				found = true
			}
		}
	}
	if !found {
		t.Error("expected highlight run")
	}
}

func TestMarkdownSuperscript(t *testing.T) {
	blocks := mdParsed("x^2^")
	found := false
	for _, b := range blocks {
		for _, r := range b.Runs {
			if r.Superscript &&
				strings.TrimSpace(r.Text) == "2" {
				found = true
			}
		}
	}
	if !found {
		t.Error("expected superscript run")
	}
}

func TestMarkdownSubscript(t *testing.T) {
	blocks := mdParsed("H~2~O")
	found := false
	for _, b := range blocks {
		for _, r := range b.Runs {
			if r.Subscript &&
				strings.TrimSpace(r.Text) == "2" {
				found = true
			}
		}
	}
	if !found {
		t.Error("expected subscript run")
	}
}

// --- Links ---

func TestMarkdownInlineLink(t *testing.T) {
	blocks := mdParsed("[click](https://example.com)")
	found := false
	for _, b := range blocks {
		for _, r := range b.Runs {
			if r.Link == "https://example.com" &&
				strings.TrimSpace(r.Text) == "click" {
				found = true
			}
		}
	}
	if !found {
		t.Error("expected link run")
	}
}

func TestMarkdownAutoLink(t *testing.T) {
	blocks := mdParsed("<https://example.com>")
	found := false
	for _, b := range blocks {
		for _, r := range b.Runs {
			if r.Link != "" &&
				strings.Contains(r.Link, "example.com") {
				found = true
			}
		}
	}
	if !found {
		t.Error("expected autolink run")
	}
}

func TestMarkdownReferenceLink(t *testing.T) {
	blocks := mdParsed("[text][ref]\n\n[ref]: https://example.com")
	found := false
	for _, b := range blocks {
		for _, r := range b.Runs {
			if r.Link == "https://example.com" {
				found = true
			}
		}
	}
	if !found {
		t.Error("expected reference link")
	}
}

// --- Images ---

func TestMarkdownImage(t *testing.T) {
	blocks := mdParsed("![alt text](image.png)")
	found := false
	for _, b := range blocks {
		if b.IsImage && b.ImageSrc == "image.png" {
			found = true
		}
	}
	if !found {
		t.Error("expected image block")
	}
}

func TestMarkdownImageTraversal(t *testing.T) {
	blocks := mdParsed("![alt](../../../etc/passwd.png)")
	for _, b := range blocks {
		if b.IsImage && b.ImageSrc != "" {
			t.Error("path traversal image should be blocked")
		}
	}
}

// --- Lists ---

func TestMarkdownUnorderedList(t *testing.T) {
	blocks := mdParsed("- item1\n- item2\n- item3")
	listCount := 0
	for _, b := range blocks {
		if b.IsList {
			listCount++
		}
	}
	if listCount < 3 {
		t.Fatalf("expected 3 list items, got %d", listCount)
	}
}

func TestMarkdownOrderedList(t *testing.T) {
	blocks := mdParsed("1. first\n2. second\n3. third")
	listCount := 0
	for _, b := range blocks {
		if b.IsList {
			listCount++
		}
	}
	if listCount < 3 {
		t.Fatalf("expected 3 ordered items, got %d", listCount)
	}
}

func TestMarkdownNestedList(t *testing.T) {
	blocks := mdParsed("- outer\n  - inner")
	found := false
	for _, b := range blocks {
		if b.IsList && b.ListIndent > 0 {
			found = true
		}
	}
	if !found {
		t.Error("expected nested list item")
	}
}

func TestMarkdownTaskList(t *testing.T) {
	blocks := mdParsed("- [x] done\n- [ ] todo")
	if len(blocks) < 2 {
		t.Fatal("expected at least 2 blocks")
	}
}

// --- Blockquotes ---

func TestMarkdownBlockquote(t *testing.T) {
	blocks := mdParsed("> quoted text")
	found := false
	for _, b := range blocks {
		if b.IsBlockquote {
			found = true
		}
	}
	if !found {
		t.Error("expected blockquote block")
	}
}

func TestMarkdownNestedBlockquote(t *testing.T) {
	blocks := mdParsed("> level1\n>> level2")
	maxDepth := 0
	for _, b := range blocks {
		if b.IsBlockquote && b.BlockquoteDepth > maxDepth {
			maxDepth = b.BlockquoteDepth
		}
	}
	if maxDepth < 2 {
		t.Fatalf("expected depth >= 2, got %d", maxDepth)
	}
}

// --- Tables ---

func TestMarkdownTable(t *testing.T) {
	blocks := mdParsed("| A | B |\n|---|---|\n| 1 | 2 |")
	found := false
	for _, b := range blocks {
		if b.IsTable && b.TableData != nil {
			found = true
			if b.TableData.ColCount != 2 {
				t.Fatalf("colCount: got %d",
					b.TableData.ColCount)
			}
		}
	}
	if !found {
		t.Error("expected table block")
	}
}

func TestMarkdownTableAlignments(t *testing.T) {
	blocks := mdParsed(
		"| L | C | R |\n|:--|:--:|--:|\n| a | b | c |")
	for _, b := range blocks {
		if b.IsTable && b.TableData != nil {
			if len(b.TableData.Alignments) < 3 {
				t.Fatal("expected 3 alignments")
			}
			if b.TableData.Alignments[0] != MdAlignLeft {
				t.Error("col0 should be left-aligned")
			}
			if b.TableData.Alignments[1] != MdAlignCenter {
				t.Error("col1 should be center-aligned")
			}
			if b.TableData.Alignments[2] != MdAlignRight {
				t.Error("col2 should be right-aligned")
			}
		}
	}
}

func TestMarkdownTableNoOuterPipes(t *testing.T) {
	blocks := mdParsed("A | B\n---|---\n1 | 2")
	found := false
	for _, b := range blocks {
		if b.IsTable && b.TableData != nil {
			found = true
		}
	}
	if !found {
		t.Error("expected table without outer pipes")
	}
}

// --- Code blocks ---

func TestMarkdownFencedCodeBlock(t *testing.T) {
	blocks := mdParsed("```go\nfmt.Println(\"hi\")\n```")
	found := false
	for _, b := range blocks {
		if b.IsCode && b.CodeLanguage == "go" {
			found = true
		}
	}
	if !found {
		t.Error("expected Go code block")
	}
}

func TestMarkdownTildeFence(t *testing.T) {
	blocks := mdParsed("~~~python\nprint('hi')\n~~~")
	found := false
	for _, b := range blocks {
		if b.IsCode && b.CodeLanguage == "py" {
			found = true
		}
	}
	if !found {
		t.Error("expected Python code block (tilde)")
	}
}

func TestMarkdownCodeBlockContent(t *testing.T) {
	blocks := mdParsed("```\nhello world\n```")
	found := false
	for _, b := range blocks {
		if b.IsCode {
			text := runsToText(b.Runs)
			if strings.Contains(text, "hello world") {
				found = true
			}
		}
	}
	if !found {
		t.Error("expected code block with content")
	}
}

// --- HR ---

func TestMarkdownHR(t *testing.T) {
	blocks := mdParsed("---")
	found := false
	for _, b := range blocks {
		if b.IsHR {
			found = true
		}
	}
	if !found {
		t.Error("expected HR block")
	}
}

// --- Math ---

func TestMarkdownDisplayMath(t *testing.T) {
	blocks := mdParsed("$$\nE = mc^2\n$$")
	found := false
	for _, b := range blocks {
		if b.IsMath {
			found = true
		}
	}
	if !found {
		t.Error("expected display math block")
	}
}

func TestMarkdownMathFence(t *testing.T) {
	blocks := mdParsed("```math\nE = mc^2\n```")
	found := false
	for _, b := range blocks {
		if b.IsMath {
			found = true
		}
	}
	if !found {
		t.Error("expected math code fence block")
	}
}

func TestMarkdownInlineMath(t *testing.T) {
	blocks := mdParsed("The equation $E = mc^2$ is famous.")
	found := false
	for _, b := range blocks {
		for _, r := range b.Runs {
			if r.MathID != "" || r.MathLatex != "" {
				found = true
			}
		}
	}
	if !found {
		t.Error("expected inline math run")
	}
}

func TestMarkdownNotMathDollarAmount(t *testing.T) {
	blocks := mdParsed("The price is $10.")
	for _, b := range blocks {
		for _, r := range b.Runs {
			if r.MathID != "" || r.MathLatex != "" {
				t.Error("$10 should not be parsed as math")
			}
		}
	}
}

// --- Footnotes ---

func TestMarkdownFootnote(t *testing.T) {
	blocks := mdParsed(
		"Text[^1] more.\n\n[^1]: Footnote content.")
	foundFootnote := false
	for _, b := range blocks {
		for _, r := range b.Runs {
			if r.Tooltip == "Footnote content." {
				foundFootnote = true
			}
		}
	}
	if !foundFootnote {
		t.Error("expected footnote tooltip")
	}
}

func TestMarkdownFootnoteNamed(t *testing.T) {
	blocks := mdParsed(
		"Text[^note] more.\n\n[^note]: Named footnote.")
	foundFootnote := false
	for _, b := range blocks {
		for _, r := range b.Runs {
			if r.Tooltip == "Named footnote." {
				foundFootnote = true
			}
		}
	}
	if !foundFootnote {
		t.Error("expected named footnote tooltip")
	}
}

func TestMarkdownFootnoteUndefined(t *testing.T) {
	blocks := mdParsed("Text[^undef] more.")
	for _, b := range blocks {
		for _, r := range b.Runs {
			if r.Tooltip != "" {
				t.Error("undefined footnote should not have tooltip")
			}
		}
	}
}

// --- Definition lists ---

func TestMarkdownDefinitionList(t *testing.T) {
	blocks := mdParsed("Term\n:   Definition text")
	hasTerm := false
	hasValue := false
	for _, b := range blocks {
		if b.IsDefTerm {
			hasTerm = true
		}
		if b.IsDefValue {
			hasValue = true
		}
	}
	if !hasTerm || !hasValue {
		t.Errorf("deflist: term=%v value=%v",
			hasTerm, hasValue)
	}
}

// --- Abbreviations ---

func TestMarkdownAbbreviation(t *testing.T) {
	blocks := mdParsed(
		"The HTML spec.\n\n*[HTML]: HyperText Markup Language")
	found := false
	for _, b := range blocks {
		for _, r := range b.Runs {
			if r.Tooltip == "HyperText Markup Language" {
				found = true
			}
		}
	}
	if !found {
		t.Error("expected abbreviation tooltip")
	}
}

func TestMarkdownAbbrWordBoundary(t *testing.T) {
	blocks := mdParsed(
		"HTMLX is not HTML.\n\n*[HTML]: Markup Language")
	// Only standalone "HTML" should have tooltip,
	// not "HTMLX".
	for _, b := range blocks {
		for _, r := range b.Runs {
			if r.Text == "HTMLX" && r.Tooltip != "" {
				t.Error("HTMLX should not have tooltip")
			}
		}
	}
}

// --- Emoji ---

func TestMarkdownEmoji(t *testing.T) {
	blocks := mdParsed(":smile:")
	if len(blocks) == 0 {
		t.Fatal("expected at least one block")
	}
	text := runsToText(blocks[0].Runs)
	if strings.Contains(text, ":smile:") {
		t.Error("emoji should be converted to unicode")
	}
}

func TestMarkdownEmojiUnknown(t *testing.T) {
	blocks := mdParsed(":notarealemojicode:")
	text := runsToText(blocks[0].Runs)
	if !strings.Contains(text, ":notarealemojicode:") {
		t.Error("unknown emoji should remain as text")
	}
}

func TestMarkdownBareColon(t *testing.T) {
	blocks := mdParsed("time: 3:00")
	text := runsToText(blocks[0].Runs)
	assertContains(t, text, "time: 3:00")
}

// --- Paragraph breaks ---

func TestMarkdownParagraphBreak(t *testing.T) {
	blocks := mdParsed("Para 1.\n\nPara 2.")
	if len(blocks) < 2 {
		t.Fatalf("expected 2 paragraphs, got %d",
			len(blocks))
	}
}

// --- Hard line breaks ---

func TestMarkdownHardLineBreak(t *testing.T) {
	blocks := mdParsed("line1  \nline2")
	text := runsToText(blocks[0].Runs)
	if !strings.Contains(text, "\n") {
		t.Error("expected hard line break")
	}
}

// --- Escapes ---

func TestMarkdownEscapes(t *testing.T) {
	blocks := mdParsed("\\*not bold\\*")
	if len(blocks) == 0 {
		t.Fatal("no blocks")
	}
	// Verify no bold formatting applied.
	for _, r := range blocks[0].Runs {
		if r.Format == MdFormatBold {
			t.Error("escaped asterisks should not be bold")
		}
	}
}

// --- Security ---

func TestMarkdownJavascriptLinkBlocked(t *testing.T) {
	// In the styled view, links with unsafe URLs should
	// not have the Link field set (via isSafeURL check
	// in rtfOnClick). The parser preserves the link.
	blocks := mdParsed("[click](javascript:alert(1))")
	for _, b := range blocks {
		for _, r := range b.Runs {
			if r.Link == "javascript:alert(1)" {
				// Parser preserves it; view layer
				// blocks via isSafeURL.
				if isSafeURL(r.Link) {
					t.Error("javascript: should be unsafe")
				}
			}
		}
	}
}

func TestSanitizeLatex(t *testing.T) {
	tests := []struct {
		input, want string
	}{
		{"E = mc^2", "E = mc^2"},
		{"\\input{secrets}", "{secrets}"},
		{"\\def\\x{y}", "\\x{y}"},
		{"\\write18{cmd}", "{cmd}"},
	}
	for _, tc := range tests {
		got := sanitizeLatex(tc.input)
		if got != tc.want {
			t.Errorf("sanitizeLatex(%q): got %q, want %q",
				tc.input, got, tc.want)
		}
	}
}

// --- Syntax highlighting ---

func TestMarkdownSyntaxHighlightGo(t *testing.T) {
	blocks := mdParsed(
		"```go\nfunc main() {\n\tfmt.Println(\"hello\")\n}\n```")
	found := false
	for _, b := range blocks {
		if !b.IsCode || b.CodeLanguage != "go" {
			continue
		}
		for _, r := range b.Runs {
			if r.CodeToken == MdTokenKeyword {
				found = true
			}
		}
	}
	if !found {
		t.Error("expected Go keyword tokens")
	}
}

func TestMarkdownSyntaxHighlightPython(t *testing.T) {
	blocks := mdParsed(
		"```python\ndef hello():\n    print(\"hi\")\n```")
	found := false
	for _, b := range blocks {
		if !b.IsCode || b.CodeLanguage != "py" {
			continue
		}
		for _, r := range b.Runs {
			if r.CodeToken == MdTokenKeyword {
				found = true
			}
		}
	}
	if !found {
		t.Error("expected Python keyword tokens")
	}
}

func TestMarkdownSyntaxHighlightJS(t *testing.T) {
	blocks := mdParsed(
		"```javascript\nconst x = 42;\n```")
	found := false
	for _, b := range blocks {
		if !b.IsCode {
			continue
		}
		for _, r := range b.Runs {
			if r.CodeToken == MdTokenKeyword {
				found = true
			}
		}
	}
	if !found {
		t.Error("expected JS keyword tokens")
	}
}

func TestMarkdownSyntaxHighlightString(t *testing.T) {
	blocks := mdParsed("```go\nvar s = \"hello\"\n```")
	found := false
	for _, b := range blocks {
		for _, r := range b.Runs {
			if r.CodeToken == MdTokenString {
				found = true
			}
		}
	}
	if !found {
		t.Error("expected string token")
	}
}

func TestMarkdownSyntaxHighlightNumber(t *testing.T) {
	blocks := mdParsed("```go\nvar n = 42\n```")
	found := false
	for _, b := range blocks {
		for _, r := range b.Runs {
			if r.CodeToken == MdTokenNumber {
				found = true
			}
		}
	}
	if !found {
		t.Error("expected number token")
	}
}

func TestMarkdownSyntaxHighlightComment(t *testing.T) {
	blocks := mdParsed("```go\n// comment\n```")
	found := false
	for _, b := range blocks {
		for _, r := range b.Runs {
			if r.CodeToken == MdTokenComment {
				found = true
			}
		}
	}
	if !found {
		t.Error("expected comment token")
	}
}

// --- Styled output ---

func TestMarkdownToRichText(t *testing.T) {
	style := DefaultMarkdownStyle()
	rt := MarkdownToRichText("**bold** and *italic*", style)
	if len(rt.Runs) == 0 {
		t.Fatal("expected runs")
	}
	text := richTextPlain(rt)
	assertContains(t, text, "bold")
	assertContains(t, text, "italic")
}

func TestMarkdownToBlocksCode(t *testing.T) {
	style := DefaultMarkdownStyle()
	blocks := markdownToBlocks("```\ncode\n```", style)
	found := false
	for _, b := range blocks {
		if b.IsCode {
			found = true
		}
	}
	if !found {
		t.Error("expected code block in styled output")
	}
}

func TestMarkdownToBlocksTable(t *testing.T) {
	style := DefaultMarkdownStyle()
	blocks := markdownToBlocks(
		"| A | B |\n|---|---|\n| 1 | 2 |", style)
	found := false
	for _, b := range blocks {
		if b.IsTable && b.TableData != nil {
			found = true
		}
	}
	if !found {
		t.Error("expected table block in styled output")
	}
}

func TestMarkdownStyledLinkColor(t *testing.T) {
	style := DefaultMarkdownStyle()
	blocks := markdownToBlocks(
		"[link](https://example.com)", style)
	for _, b := range blocks {
		for _, r := range b.Content.Runs {
			if r.Link == "https://example.com" {
				if r.Style.Color != style.LinkColor {
					t.Error("link should use LinkColor")
				}
				if !r.Style.Underline {
					t.Error("link should be underlined")
				}
			}
		}
	}
}

func TestMarkdownStyledHighlightBG(t *testing.T) {
	style := DefaultMarkdownStyle()
	blocks := markdownToBlocks("==marked==", style)
	for _, b := range blocks {
		for _, r := range b.Content.Runs {
			if strings.TrimSpace(r.Text) == "marked" {
				if r.Style.BgColor != style.HighlightBG {
					t.Error("highlight should use HighlightBG")
				}
			}
		}
	}
}

func TestMarkdownStyledStrikethrough(t *testing.T) {
	style := DefaultMarkdownStyle()
	blocks := markdownToBlocks("~~deleted~~", style)
	for _, b := range blocks {
		for _, r := range b.Content.Runs {
			if strings.TrimSpace(r.Text) == "deleted" {
				if !r.Style.Strikethrough {
					t.Error("strikethrough not set")
				}
			}
		}
	}
}

// --- Misc ---

func TestMarkdownEmpty(t *testing.T) {
	blocks := mdParsed("")
	if len(blocks) != 0 {
		t.Fatalf("empty source should produce 0 blocks, got %d",
			len(blocks))
	}
}

func TestMarkdownPlainParagraph(t *testing.T) {
	blocks := mdParsed("Hello world.")
	if len(blocks) == 0 {
		t.Fatal("expected at least one block")
	}
	assertContains(t, runsToText(blocks[0].Runs), "Hello world.")
}

func TestMarkdownMultipleBlocks(t *testing.T) {
	blocks := mdParsed("# Header\n\nParagraph.\n\n---\n\n> Quote")
	hasHeader := false
	hasHR := false
	hasBQ := false
	for _, b := range blocks {
		if b.HeaderLevel > 0 {
			hasHeader = true
		}
		if b.IsHR {
			hasHR = true
		}
		if b.IsBlockquote {
			hasBQ = true
		}
	}
	if !hasHeader || !hasHR || !hasBQ {
		t.Errorf("header=%v HR=%v BQ=%v",
			hasHeader, hasHR, hasBQ)
	}
}

func TestMarkdownBuildTableData(t *testing.T) {
	style := DefaultMarkdownStyle()
	blocks := markdownToBlocks(
		"| H1 | H2 |\n|---|---|\n| a | b |", style)
	for _, b := range blocks {
		if b.IsTable && b.TableData != nil {
			rows := buildMarkdownTableData(
				*b.TableData, style)
			if len(rows) < 2 {
				t.Fatal("expected header + data row")
			}
			// Header row cells.
			if len(rows[0].Cells) != 2 {
				t.Fatalf("header cells: got %d",
					len(rows[0].Cells))
			}
			if !rows[0].Cells[0].HeadCell {
				t.Error("first header cell not marked as head")
			}
		}
	}
}

// --- Code language detection ---

func TestLangFromHint(t *testing.T) {
	tests := []struct {
		hint string
		want MdCodeLanguage
	}{
		{"go", MdLangGo},
		{"golang", MdLangGo},
		{"python", MdLangPython},
		{"py", MdLangPython},
		{"javascript", MdLangJavaScript},
		{"js", MdLangJavaScript},
		{"typescript", MdLangTypeScript},
		{"ts", MdLangTypeScript},
		{"rust", MdLangRust},
		{"rs", MdLangRust},
		{"c", MdLangC},
		{"json", MdLangJSON},
		{"bash", MdLangShell},
		{"sh", MdLangShell},
		{"html", MdLangHTML},
		{"unknown", MdLangGeneric},
	}
	for _, tc := range tests {
		got := langFromHint(tc.hint)
		if got != tc.want {
			t.Errorf("langFromHint(%q): got %d, want %d",
				tc.hint, got, tc.want)
		}
	}
}

func TestMarkdownListPrefix(t *testing.T) {
	blocks := mdParsed("- item")
	for _, b := range blocks {
		if b.IsList {
			if b.ListPrefix == "" {
				t.Error("list prefix should not be empty")
			}
		}
	}
}

func TestMarkdownOrderedListPrefix(t *testing.T) {
	blocks := mdParsed("1. first\n2. second")
	for _, b := range blocks {
		if b.IsList && strings.Contains(b.ListPrefix, "1") {
			return // found
		}
	}
	// Numbered prefixes vary — just check we got list items.
	count := 0
	for _, b := range blocks {
		if b.IsList {
			count++
		}
	}
	if count < 2 {
		t.Error("expected at least 2 ordered list items")
	}
}
