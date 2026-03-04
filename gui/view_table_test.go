package gui

import (
	"testing"

	"github.com/mike-ward/go-glyph"
)

func TestTableBasic(t *testing.T) {
	v := Table(TableCfg{
		Data: []TableRowCfg{
			TR([]TableCellCfg{TH("Name"), TH("Age")}),
			TR([]TableCellCfg{TD("Alice"), TD("30")}),
			TR([]TableCellCfg{TD("Bob"), TD("25")}),
		},
	})
	w := &Window{}
	layout := GenerateViewLayout(v, w)
	if len(layout.Children) != 3 {
		t.Fatalf("rows = %d, want 3", len(layout.Children))
	}
}

func TestTableEmpty(t *testing.T) {
	v := Table(TableCfg{})
	w := &Window{}
	layout := GenerateViewLayout(v, w)
	if len(layout.Children) != 0 {
		t.Errorf("children = %d, want 0", len(layout.Children))
	}
}

func TestTableBorderAll(t *testing.T) {
	v := Table(TableCfg{
		BorderStyle: TableBorderAll,
		SizeBorder:  1,
		Data: []TableRowCfg{
			TR([]TableCellCfg{TD("a"), TD("b")}),
			TR([]TableCellCfg{TD("c"), TD("d")}),
		},
	})
	w := &Window{}
	layout := GenerateViewLayout(v, w)
	if len(layout.Children) != 2 {
		t.Errorf("rows = %d, want 2", len(layout.Children))
	}
}

func TestTableBorderHorizontal(t *testing.T) {
	v := Table(TableCfg{
		BorderStyle: TableBorderHorizontal,
		SizeBorder:  1,
		Data: []TableRowCfg{
			TR([]TableCellCfg{TH("H")}),
			TR([]TableCellCfg{TD("a")}),
			TR([]TableCellCfg{TD("b")}),
		},
	})
	w := &Window{}
	layout := GenerateViewLayout(v, w)
	// 3 rows + 2 separators (between 0-1 and 1-2; not after last).
	if len(layout.Children) != 5 {
		t.Errorf("children = %d, want 5", len(layout.Children))
	}
}

func TestTableBorderHeaderOnly(t *testing.T) {
	v := Table(TableCfg{
		BorderStyle: TableBorderHeaderOnly,
		SizeBorder:  2,
		Data: []TableRowCfg{
			TR([]TableCellCfg{TH("H")}),
			TR([]TableCellCfg{TD("a")}),
			TR([]TableCellCfg{TD("b")}),
		},
	})
	w := &Window{}
	layout := GenerateViewLayout(v, w)
	// 3 rows + 1 separator (after header only).
	if len(layout.Children) != 4 {
		t.Errorf("children = %d, want 4", len(layout.Children))
	}
}

func TestTableCfgFromData(t *testing.T) {
	data := [][]string{
		{"Name", "Age"},
		{"Alice", "30"},
	}
	cfg := TableCfgFromData(data)
	if len(cfg.Data) != 2 {
		t.Fatalf("rows = %d, want 2", len(cfg.Data))
	}
	if !cfg.Data[0].Cells[0].HeadCell {
		t.Error("first row should be header")
	}
	if cfg.Data[1].Cells[0].HeadCell {
		t.Error("second row should not be header")
	}
}

func TestTableCfgError(t *testing.T) {
	cfg := TableCfgError("oops")
	if len(cfg.Data) != 1 {
		t.Fatalf("rows = %d, want 1", len(cfg.Data))
	}
	if cfg.Data[0].Cells[0].Value != "oops" {
		t.Errorf("value = %q", cfg.Data[0].Cells[0].Value)
	}
}

func TestTableHelpers(t *testing.T) {
	row := TR([]TableCellCfg{TH("x"), TD("y")})
	if len(row.Cells) != 2 {
		t.Fatalf("cells = %d", len(row.Cells))
	}
	if !row.Cells[0].HeadCell {
		t.Error("TH should be head")
	}
	if row.Cells[1].HeadCell {
		t.Error("TD should not be head")
	}
}

func TestTableSelection(t *testing.T) {
	var selectedRows map[int]bool
	var clickedRow int
	v := Table(TableCfg{
		Data: []TableRowCfg{
			TR([]TableCellCfg{TD("a")}),
			TR([]TableCellCfg{TD("b")}),
		},
		OnSelect: func(sel map[int]bool, row int, _ *Event, _ *Window) {
			selectedRows = sel
			clickedRow = row
		},
	})
	w := &Window{}
	layout := GenerateViewLayout(v, w)
	// Click second row.
	row := &layout.Children[1]
	if row.Shape.HasEvents() && row.Shape.Events.OnClick != nil {
		row.Shape.Events.OnClick(row, &Event{}, w)
	}
	if clickedRow != 1 {
		t.Errorf("clicked row = %d, want 1", clickedRow)
	}
	if !selectedRows[1] {
		t.Error("row 1 should be selected")
	}
}

type tableTestMeasurer struct{}

func (m *tableTestMeasurer) TextWidth(text string, _ TextStyle) float32 {
	return float32(len(text)) * 8
}
func (m *tableTestMeasurer) TextHeight(_ string, _ TextStyle) float32 { return 16 }
func (m *tableTestMeasurer) FontHeight(_ TextStyle) float32            { return 16 }
func (m *tableTestMeasurer) LayoutText(_ string, _ TextStyle, _ float32) (glyph.Layout, error) {
	return glyph.Layout{Height: 16}, nil
}

func TestTableColumnAutoWidth(t *testing.T) {
	v := Table(TableCfg{
		TextMeasurer: &tableTestMeasurer{},
		CellPadding:  NewPadding(4, 4, 4, 4),
		Data: []TableRowCfg{
			TR([]TableCellCfg{TH("Name"), TH("Age")}),
			TR([]TableCellCfg{TD("Alexander"), TD("30")}),
		},
	})
	w := &Window{}
	layout := GenerateViewLayout(v, w)
	// First column cell should be wider than second.
	row0 := layout.Children[0]
	if len(row0.Children) < 2 {
		t.Fatal("expected 2 cells")
	}
	col0W := row0.Children[0].Shape.Width
	col1W := row0.Children[1].Shape.Width
	if col0W <= col1W {
		t.Errorf("col0 width (%.1f) should be > col1 (%.1f)",
			col0W, col1W)
	}
}

func TestTableRowAltColor(t *testing.T) {
	alt := RGB(50, 50, 50)
	v := Table(TableCfg{
		ColorRowAlt: &alt,
		Data: []TableRowCfg{
			TR([]TableCellCfg{TD("a")}),
			TR([]TableCellCfg{TD("b")}),
		},
	})
	w := &Window{}
	layout := GenerateViewLayout(v, w)
	// Row 0: transparent, Row 1: alt color.
	if layout.Children[1].Shape.Color != alt {
		t.Errorf("row 1 color = %v, want %v",
			layout.Children[1].Shape.Color, alt)
	}
}
