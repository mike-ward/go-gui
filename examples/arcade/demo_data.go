package main

import "github.com/mike-ward/go-gui/gui"

func demoTable(_ *gui.Window) gui.View {
	cfg := gui.TableCfgFromData([][]string{
		{"Language", "Year", "Creator", "Type", "Paradigm"},
		{"Go", "2009", "Rob Pike", "Compiled", "Concurrent"},
		{"Rust", "2010", "Graydon Hoare", "Compiled", "Systems"},
		{"Zig", "2016", "Andrew Kelley", "Compiled", "Systems"},
		{"Python", "1991", "Guido van Rossum", "Interpreted", "Multi"},
		{"TypeScript", "2012", "Anders Hejlsberg", "Transpiled", "OOP"},
		{"Ruby", "1995", "Yukihiro Matsumoto", "Interpreted", "OOP"},
		{"C", "1972", "Dennis Ritchie", "Compiled", "Procedural"},
		{"Java", "1995", "James Gosling", "Compiled", "OOP"},
		{"Kotlin", "2011", "JetBrains", "Compiled", "Multi"},
		{"Swift", "2014", "Chris Lattner", "Compiled", "Multi"},
	})
	cfg.ID = "table-demo"
	cfg.Sizing = gui.FillFit
	return gui.Table(cfg)
}

func demoDataGrid(w *gui.Window) gui.View {
	return w.DataGrid(gui.DataGridCfg{
		ID:                "datagrid-demo",
		PageSize:          5,
		ShowQuickFilter:   true,
		ShowColumnChooser: true,
		Columns: []gui.GridColumnCfg{
			{ID: "name", Title: "Name", Width: 150, Sortable: true, Filterable: true, Reorderable: true},
			{ID: "lang", Title: "Language", Width: 120, Sortable: true, Filterable: true, Reorderable: true},
			{ID: "stars", Title: "Stars", Width: 80, Sortable: true, Reorderable: true},
			{ID: "license", Title: "License", Width: 100, Sortable: true, Filterable: true, Reorderable: true},
			{ID: "category", Title: "Category", Width: 120, Sortable: true, Filterable: true, Reorderable: true},
		},
		Rows: []gui.GridRow{
			{ID: "1", Cells: map[string]string{"name": "go-gui", "lang": "Go", "stars": "1200", "license": "MIT", "category": "GUI"}},
			{ID: "2", Cells: map[string]string{"name": "egui", "lang": "Rust", "stars": "18000", "license": "MIT", "category": "GUI"}},
			{ID: "3", Cells: map[string]string{"name": "Dear ImGui", "lang": "C++", "stars": "52000", "license": "MIT", "category": "GUI"}},
			{ID: "4", Cells: map[string]string{"name": "Fyne", "lang": "Go", "stars": "23000", "license": "BSD", "category": "GUI"}},
			{ID: "5", Cells: map[string]string{"name": "Gio", "lang": "Go", "stars": "1500", "license": "MIT", "category": "GUI"}},
			{ID: "6", Cells: map[string]string{"name": "Slint", "lang": "Rust", "stars": "15000", "license": "GPL", "category": "GUI"}},
			{ID: "7", Cells: map[string]string{"name": "Flutter", "lang": "Dart", "stars": "160000", "license": "BSD", "category": "Mobile"}},
			{ID: "8", Cells: map[string]string{"name": "React", "lang": "JavaScript", "stars": "220000", "license": "MIT", "category": "Web"}},
			{ID: "9", Cells: map[string]string{"name": "Vue", "lang": "JavaScript", "stars": "205000", "license": "MIT", "category": "Web"}},
			{ID: "10", Cells: map[string]string{"name": "Svelte", "lang": "JavaScript", "stars": "75000", "license": "MIT", "category": "Web"}},
			{ID: "11", Cells: map[string]string{"name": "Qt", "lang": "C++", "stars": "0", "license": "LGPL", "category": "GUI"}},
			{ID: "12", Cells: map[string]string{"name": "GTK", "lang": "C", "stars": "0", "license": "LGPL", "category": "GUI"}},
			{ID: "13", Cells: map[string]string{"name": "Electron", "lang": "JavaScript", "stars": "110000", "license": "MIT", "category": "Desktop"}},
			{ID: "14", Cells: map[string]string{"name": "Tauri", "lang": "Rust", "stars": "75000", "license": "MIT", "category": "Desktop"}},
		},
		OnColumnOrderChange: func(_ []string, _ *gui.Event, _ *gui.Window) {},
		Sizing:              gui.FillFit,
	})
}
