package gui

// TableBorderStyle controls which borders are drawn in a table.
type TableBorderStyle uint8

const (
	TableBorderNone       TableBorderStyle = iota // no borders
	TableBorderAll                                // full grid
	TableBorderHorizontal                         // horizontal lines between rows
	TableBorderHeaderOnly                         // single line under header row
)

// TableRowCfg configures a table row.
type TableRowCfg struct {
	ID      string
	Cells   []TableCellCfg
	OnClick func(*Layout, *Event, *Window)
}

// TableCellCfg configures a table cell.
type TableCellCfg struct {
	ID        string
	Value     string
	HeadCell  bool
	HAlign    *HorizontalAlign
	TextStyle *TextStyle
	Content   View
	OnClick   func(*Layout, *Event, *Window)
}

// TableCfg configures a table layout.
type TableCfg struct {
	ID                 string
	ColorBorder        Color
	ColorSelect        Color
	ColorHover         Color
	ColorRowAlt        *Color
	CellPadding        Padding
	TextStyle          TextStyle
	TextStyleHead      TextStyle
	AlignHead          HorizontalAlign
	ColumnAlignments   []HorizontalAlign
	ColumnWidthDefault float32
	ColumnWidthMin     float32
	SizeBorder         float32
	SizeBorderHeader   float32
	BorderStyle        TableBorderStyle
	MultiSelect        bool
	Selected           map[int]bool
	OnSelect           func(map[int]bool, int, *Event, *Window)
	Data               []TableRowCfg

	// Text measurement — set by caller if *Window available.
	// When non-nil, column widths auto-size to content.
	TextMeasurer TextMeasurer

	// Sizing
	Sizing    Sizing
	Width     float32
	Height    float32
	MinWidth  float32
	MaxWidth  float32
	MinHeight float32
	MaxHeight float32
}

func applyTableDefaults(cfg *TableCfg) {
	s := &DefaultTableStyle
	if !cfg.ColorBorder.IsSet() {
		cfg.ColorBorder = s.ColorBorder
	}
	if !cfg.ColorSelect.IsSet() {
		cfg.ColorSelect = s.ColorSelect
	}
	if !cfg.ColorHover.IsSet() {
		cfg.ColorHover = s.ColorHover
	}
	if !cfg.CellPadding.IsSet() {
		cfg.CellPadding = s.CellPadding
	}
	if cfg.TextStyle == (TextStyle{}) {
		cfg.TextStyle = s.TextStyle
	}
	if cfg.TextStyleHead == (TextStyle{}) {
		cfg.TextStyleHead = s.TextStyleHead
	}
	if cfg.AlignHead == HAlignStart {
		cfg.AlignHead = s.AlignHead
	}
	if cfg.ColumnWidthDefault == 0 {
		cfg.ColumnWidthDefault = s.ColumnWidthDefault
	}
	if cfg.ColumnWidthMin == 0 {
		cfg.ColumnWidthMin = s.ColumnWidthMin
	}
}

// Table generates a table from the given TableCfg.
// When TextMeasurer is set, column widths auto-size to content.
func Table(cfg TableCfg) View {
	applyTableDefaults(&cfg)

	if len(cfg.Data) == 0 {
		return Column(ContainerCfg{
			ID:      cfg.ID,
			Padding: PaddingNone,
		})
	}

	lastRowIdx := len(cfg.Data) - 1

	var cellBorder float32
	if cfg.BorderStyle == TableBorderAll {
		cellBorder = cfg.SizeBorder
	}

	var rowSpacing float32
	if cfg.BorderStyle == TableBorderAll {
		rowSpacing = -cfg.SizeBorder
	}

	// Compute column widths.
	numCols := 0
	for _, r := range cfg.Data {
		if len(r.Cells) > numCols {
			numCols = len(r.Cells)
		}
	}
	columnWidths := make([]float32, numCols)
	if cfg.TextMeasurer != nil {
		pad := cfg.CellPadding.Width()
		for _, r := range cfg.Data {
			for ci, cell := range r.Cells {
				style := cfg.TextStyle
				if cell.TextStyle != nil {
					style = *cell.TextStyle
				} else if cell.HeadCell {
					style = cfg.TextStyleHead
				}
				tw := cfg.TextMeasurer.TextWidth(
					cell.Value, style) + pad
				if tw > columnWidths[ci] {
					columnWidths[ci] = tw
				}
			}
		}
		for i := range columnWidths {
			if columnWidths[i] < cfg.ColumnWidthMin {
				columnWidths[i] = cfg.ColumnWidthMin
			}
		}
	} else {
		w := cfg.ColumnWidthDefault + cfg.CellPadding.Width()
		for i := range columnWidths {
			columnWidths[i] = w
		}
	}

	// Hoist loop-invariant values.
	onSelect := cfg.OnSelect
	selected := cfg.Selected
	multiSelect := cfg.MultiSelect
	colorHover := cfg.ColorHover

	rows := make([]View, 0, len(cfg.Data)*2)

	for rowIdx, r := range cfg.Data {
		cells := make([]View, 0, len(r.Cells))
		for colIdx, cell := range r.Cells {
			cellTextStyle := cfg.TextStyle
			if cell.TextStyle != nil {
				cellTextStyle = *cell.TextStyle
			} else if cell.HeadCell {
				cellTextStyle = cfg.TextStyleHead
			}

			var colWidth float32
			if colIdx < len(columnWidths) {
				colWidth = columnWidths[colIdx]
			}

			hAlign := HAlignStart
			if cell.HAlign != nil {
				hAlign = *cell.HAlign
			} else if cell.HeadCell {
				hAlign = cfg.AlignHead
			} else if colIdx < len(cfg.ColumnAlignments) {
				hAlign = cfg.ColumnAlignments[colIdx]
			}

			var cellContent []View
			if cell.Content != nil {
				cellContent = []View{cell.Content}
			} else {
				cellContent = []View{
					Text(TextCfg{
						Text:      cell.Value,
						TextStyle: cellTextStyle,
					}),
				}
			}

			cellOnClick := cell.OnClick
			var cellOnHover func(*Layout, *Event, *Window)
			if cellOnClick != nil {
				cellOnHover = func(layout *Layout, _ *Event, w *Window) {
					w.SetMouseCursorPointingHand()
					layout.Shape.Color = colorHover
				}
			}

			cells = append(cells, Column(ContainerCfg{
				Color:       ColorTransparent,
				ColorBorder: cfg.ColorBorder,
				SizeBorder:  Some(cellBorder),
				Padding:     cfg.CellPadding,
				HAlign:      hAlign,
				Sizing:      FixedFill,
				Width:       colWidth,
				OnClick:     cellOnClick,
				OnHover:     cellOnHover,
				Content:     cellContent,
			}))
		}

		isSelected := selected[rowIdx]
		rowColor := ColorTransparent
		if isSelected {
			rowColor = cfg.ColorSelect
		} else if cfg.ColorRowAlt != nil && rowIdx%2 == 1 {
			rowColor = *cfg.ColorRowAlt
		}

		rowOnClick := r.OnClick
		ri := rowIdx

		rows = append(rows, Row(ContainerCfg{
			Color:   rowColor,
			Spacing: Some(-cellBorder),
			Padding: PaddingNone,
			Content: cells,
			OnClick: func(layout *Layout, e *Event, w *Window) {
				if rowOnClick != nil {
					rowOnClick(layout, e, w)
				}
				if onSelect != nil {
					var newSel map[int]bool
					if multiSelect {
						newSel = copySelected(selected)
					} else {
						newSel = make(map[int]bool)
					}
					if newSel[ri] {
						delete(newSel, ri)
					} else {
						newSel[ri] = true
					}
					onSelect(newSel, ri, e, w)
				}
			},
			OnHover: func(layout *Layout, _ *Event, w *Window) {
				if onSelect != nil {
					w.SetMouseCursorPointingHand()
					if !isSelected {
						layout.Shape.Color = colorHover
					}
				}
			},
		}))

		// Separator.
		sepHeight := cfg.SizeBorder
		if rowIdx == 0 && cfg.SizeBorderHeader > 0 {
			sepHeight = cfg.SizeBorderHeader
		}

		needsSep := false
		switch cfg.BorderStyle {
		case TableBorderHorizontal:
			needsSep = rowIdx != lastRowIdx
		case TableBorderHeaderOnly:
			needsSep = rowIdx == 0
		}

		if needsSep {
			rows = append(rows, Rectangle(RectangleCfg{
				Color:  cfg.ColorBorder,
				Height: sepHeight,
				Sizing: FillFixed,
			}))
		}
	}

	return Column(ContainerCfg{
		ID:        cfg.ID,
		Color:     ColorTransparent,
		Padding:   PaddingNone,
		Spacing:   Some(rowSpacing),
		Sizing:    cfg.Sizing,
		Width:     cfg.Width,
		Height:    cfg.Height,
		MinWidth:  cfg.MinWidth,
		MaxWidth:  cfg.MaxWidth,
		MinHeight: cfg.MinHeight,
		MaxHeight: cfg.MaxHeight,
		Content:   rows,
	})
}

// TR creates a table row from the given cells.
func TR(cols []TableCellCfg) TableRowCfg {
	return TableRowCfg{Cells: cols}
}

// TH creates a header cell.
func TH(value string) TableCellCfg {
	return TableCellCfg{Value: value, HeadCell: true}
}

// TD creates a data cell.
func TD(value string) TableCellCfg {
	return TableCellCfg{Value: value}
}

// TableCfgFromData creates a TableCfg from [][]string.
// First row is treated as a header row.
func TableCfgFromData(data [][]string) TableCfg {
	rows := make([]TableRowCfg, 0, len(data))
	for i, r := range data {
		cells := make([]TableCellCfg, 0, len(r))
		for _, cell := range r {
			cells = append(cells, TableCellCfg{
				Value:    cell,
				HeadCell: i == 0,
			})
		}
		rows = append(rows, TableRowCfg{Cells: cells})
	}
	return TableCfg{Data: rows}
}

// TableCfgError creates a TableCfg with an error message.
func TableCfgError(message string) TableCfg {
	return TableCfg{Data: []TableRowCfg{TR([]TableCellCfg{TD(message)})}}
}

func copySelected(m map[int]bool) map[int]bool {
	if m == nil {
		return nil
	}
	cp := make(map[int]bool, len(m))
	for k, v := range m {
		cp[k] = v
	}
	return cp
}
