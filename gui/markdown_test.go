package gui

import (
	"strings"
	"testing"
)

// --- Test helpers ---

func mdBlocks(source string) []MarkdownBlock {
	return markdownToBlocks(source, DefaultMarkdownStyle())
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

// --- Styled output ---

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

func TestMarkdownToRichText(t *testing.T) {
	style := DefaultMarkdownStyle()
	rt := MarkdownToRichText("**bold** and *italic*", style)
	if len(rt.Runs) == 0 {
		t.Fatal("expected runs")
	}
	text := richTextPlain(rt)
	if !strings.Contains(text, "bold") {
		t.Error("expected bold text")
	}
	if !strings.Contains(text, "italic") {
		t.Error("expected italic text")
	}
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

func TestMarkdownDefaultsToWrap(t *testing.T) {
	w := &Window{}
	layout := GenerateViewLayout(w.Markdown(MarkdownCfg{
		Source: "A paragraph that should wrap by default.",
		Style:  DefaultMarkdownStyle(),
	}), w)

	if layout.Shape == nil || layout.Shape.Sizing != FillFit {
		t.Fatalf("markdown sizing = %v, want FillFit", layout.Shape.Sizing)
	}

	mode, ok := firstMarkdownTextMode(layout)
	if !ok {
		t.Fatal("expected markdown text node")
	}
	if mode != TextModeWrap {
		t.Fatalf("markdown text mode = %v, want TextModeWrap", mode)
	}
}

func TestMarkdownCanExplicitlyUseSingleLine(t *testing.T) {
	w := &Window{}
	layout := GenerateViewLayout(w.Markdown(MarkdownCfg{
		Source: "single line",
		Style:  DefaultMarkdownStyle(),
		Mode:   Some(TextModeSingleLine),
	}), w)

	if layout.Shape == nil || layout.Shape.Sizing != FitFit {
		t.Fatalf("markdown sizing = %v, want FitFit", layout.Shape.Sizing)
	}

	mode, ok := firstMarkdownTextMode(layout)
	if !ok {
		t.Fatal("expected markdown text node")
	}
	if mode != TextModeSingleLine {
		t.Fatalf("markdown text mode = %v, want TextModeSingleLine", mode)
	}
}

func firstMarkdownTextMode(layout Layout) (TextMode, bool) {
	if layout.Shape != nil && layout.Shape.TC != nil {
		return layout.Shape.TC.TextMode, true
	}
	for _, child := range layout.Children {
		if mode, ok := firstMarkdownTextMode(child); ok {
			return mode, true
		}
	}
	return 0, false
}
