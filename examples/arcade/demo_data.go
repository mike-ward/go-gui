package main

import "github.com/mike-ward/go-gui/gui"

func demoTable(_ *gui.Window) gui.View {
	cfg := gui.TableCfgFromData([][]string{
		{"Language", "Year", "Creator", "Type"},
		{"Go", "2009", "Rob Pike", "Compiled"},
		{"Rust", "2010", "Graydon Hoare", "Compiled"},
		{"Zig", "2016", "Andrew Kelley", "Compiled"},
		{"Python", "1991", "Guido van Rossum", "Interpreted"},
		{"TypeScript", "2012", "Anders Hejlsberg", "Transpiled"},
		{"Ruby", "1995", "Yukihiro Matsumoto", "Interpreted"},
	})
	cfg.ID = "table-demo"
	cfg.Sizing = gui.FillFit
	return gui.Table(cfg)
}

func demoDataGrid(w *gui.Window) gui.View {
	return w.DataGrid(gui.DataGridCfg{
		ID:              "datagrid-demo",
		PageSize:        5,
		ShowQuickFilter: true,
		Columns: []gui.GridColumnCfg{
			{ID: "name", Title: "Name", Width: 150, Sortable: true, Filterable: true},
			{ID: "lang", Title: "Language", Width: 120, Sortable: true, Filterable: true},
			{ID: "stars", Title: "Stars", Width: 80, Sortable: true},
			{ID: "license", Title: "License", Width: 100, Sortable: true},
		},
		Rows: []gui.GridRow{
			{ID: "1", Cells: map[string]string{"name": "go-gui", "lang": "Go", "stars": "1200", "license": "MIT"}},
			{ID: "2", Cells: map[string]string{"name": "egui", "lang": "Rust", "stars": "18000", "license": "MIT"}},
			{ID: "3", Cells: map[string]string{"name": "Dear ImGui", "lang": "C++", "stars": "52000", "license": "MIT"}},
			{ID: "4", Cells: map[string]string{"name": "Fyne", "lang": "Go", "stars": "23000", "license": "BSD"}},
			{ID: "5", Cells: map[string]string{"name": "Gio", "lang": "Go", "stars": "1500", "license": "MIT"}},
			{ID: "6", Cells: map[string]string{"name": "Slint", "lang": "Rust", "stars": "15000", "license": "GPL"}},
			{ID: "7", Cells: map[string]string{"name": "Flutter", "lang": "Dart", "stars": "160000", "license": "BSD"}},
			{ID: "8", Cells: map[string]string{"name": "React", "lang": "JavaScript", "stars": "220000", "license": "MIT"}},
		},
		Sizing: gui.FillFit,
	})
}
