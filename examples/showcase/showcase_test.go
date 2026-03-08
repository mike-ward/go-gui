package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mike-ward/go-gui/gui"
)

func TestEntryMatchesQuery(t *testing.T) {
	entry := DemoEntry{
		ID:      "numeric_input",
		Label:   "Numeric Input",
		Group:   groupInput,
		Summary: "Locale-aware number input with step controls.",
		Tags:    []string{"number", "decimal"},
	}

	tests := []string{"numeric", "locale-aware", "input", "decimal"}
	for _, query := range tests {
		if !entryMatchesQuery(entry, query) {
			t.Fatalf("expected query %q to match entry", query)
		}
	}
}

func TestFilteredEntriesByGroupAndQuery(t *testing.T) {
	app := &ShowcaseApp{
		SelectedGroup: groupWelcome,
		NavQuery:      "splitter",
	}
	entries := filteredEntries(app)
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].ID != "doc_splitter" {
		t.Fatalf("expected doc_splitter, got %s", entries[0].ID)
	}
}

func TestPreferredComponentForGroupPinsWelcome(t *testing.T) {
	entries := []DemoEntry{
		{ID: "doc_tables", Label: "Tables"},
		{ID: "welcome", Label: "Welcome"},
		{ID: "doc_get_started", Label: "Get Started"},
	}
	if got := preferredComponentForGroup(groupWelcome, entries); got != "welcome" {
		t.Fatalf("expected welcome, got %s", got)
	}
}

func TestRelatedExamplesMap(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd failed: %v", err)
	}
	repoRoot := filepath.Clean(filepath.Join(wd, "..", ".."))

	for _, entry := range demoEntries() {
		got := relatedExamplePaths(entry.ID)
		if len(got) == 0 {
			t.Fatalf("expected related examples for %s", entry.ID)
		}
		for _, path := range got {
			if !strings.HasPrefix(path, "examples/") {
				t.Fatalf("expected examples/* path for %s, got %q", entry.ID, path)
			}
			if strings.HasPrefix(path, "/") {
				t.Fatalf("expected relative example path for %s, got %q", entry.ID, path)
			}
			info, err := os.Stat(filepath.Join(repoRoot, path))
			if err != nil {
				t.Fatalf("expected related example path for %s to exist: %s (%v)", entry.ID, path, err)
			}
			if info.IsDir() {
				t.Fatalf("expected related example path for %s to be a file, got directory %s", entry.ID, path)
			}
		}
	}
}

func TestComponentDocsExist(t *testing.T) {
	for _, id := range []string{"data_source", "tree", "drag_reorder"} {
		if doc := componentDoc(id); doc == "" {
			t.Fatalf("expected docs for %s", id)
		}
	}
}

func TestDocPagesExist(t *testing.T) {
	for _, id := range []string{"doc_data_grid", "doc_tree"} {
		if doc := docPageSource(id); doc == "" {
			t.Fatalf("expected doc page for %s", id)
		}
	}
}

func TestTreeTitleBarShowsDocToggle(t *testing.T) {
	layout := gui.GenerateViewLayout(viewTitleBar(DemoEntry{
		ID:    "tree",
		Label: "Tree View",
	}, false), &gui.Window{})

	if _, ok := layout.FindByID("btn-doc-toggle"); !ok {
		t.Fatal("expected tree title bar to include btn-doc-toggle")
	}
}

func TestDemoTreeWrapsIntroText(t *testing.T) {
	w := &gui.Window{}
	w.SetState(newShowcaseApp())

	layout := gui.GenerateViewLayout(demoTree(w), w)
	if len(layout.Children) < 2 {
		t.Fatalf("len(layout.Children) = %d, want >= 2", len(layout.Children))
	}

	for idx := 0; idx < 2; idx++ {
		tc := layout.Children[idx].Shape.TC
		if tc == nil {
			t.Fatalf("layout.Children[%d].Shape.TC = nil, want text config", idx)
		}
		if tc.TextMode != gui.TextModeWrap {
			t.Fatalf("layout.Children[%d].Shape.TC.TextMode = %v, want %v", idx, tc.TextMode, gui.TextModeWrap)
		}
	}
}

func TestDetailPanelSummaryWraps(t *testing.T) {
	w := &gui.Window{}
	app := newShowcaseApp()
	app.SelectedGroup = groupData
	app.SelectedComponent = "tree"
	w.SetState(app)

	layout := gui.GenerateViewLayout(detailPanel(w), w)
	if len(layout.Children) < 2 {
		t.Fatalf("len(layout.Children) = %d, want >= 2", len(layout.Children))
	}

	tc := layout.Children[1].Shape.TC
	if tc == nil {
		t.Fatal("layout.Children[1].Shape.TC = nil, want summary text")
	}
	if tc.TextMode != gui.TextModeWrap {
		t.Fatalf("layout.Children[1].Shape.TC.TextMode = %v, want %v", tc.TextMode, gui.TextModeWrap)
	}
}

func TestComponentDemoDocTreeRoute(t *testing.T) {
	layout := gui.GenerateViewLayout(componentDemo(&gui.Window{}, "doc_tree"), &gui.Window{})
	if _, ok := layout.FindByID("showcase-doc-doc_tree"); !ok {
		t.Fatal("expected doc_tree route to render showcase-doc-doc_tree")
	}
}

func TestShowcaseDataGridApplyQuery(t *testing.T) {
	rows := showcaseDataGridRows()
	query := gui.GridQueryState{
		QuickFilter: "core",
		Sorts: []gui.GridSort{
			{ColID: "name", Dir: gui.GridSortDesc},
		},
	}

	filtered := showcaseDataGridApplyQuery(rows, query)
	if len(filtered) != 2 {
		t.Fatalf("expected 2 filtered rows, got %d", len(filtered))
	}
	if filtered[0].Cells["name"] != "Priya" {
		t.Fatalf("expected Priya first after desc sort, got %s", filtered[0].Cells["name"])
	}
}

func TestThemeCfgSaveLoadRoundTrip(t *testing.T) {
	cfg := generateThemeCfg(
		gui.RGB(255, 85, 0),
		"analogous",
		true,
		35,
		gui.White,
		7,
		2,
	)

	path := filepath.Join(t.TempDir(), "theme.json")
	if err := themeCfgSave(path, cfg); err != nil {
		t.Fatalf("themeCfgSave failed: %v", err)
	}

	got, err := themeCfgLoad(path)
	if err != nil {
		t.Fatalf("themeCfgLoad failed: %v", err)
	}

	if got.ColorSelect != cfg.ColorSelect {
		t.Fatalf("expected color select %v, got %v", cfg.ColorSelect, got.ColorSelect)
	}
	if got.Radius != cfg.Radius {
		t.Fatalf("expected radius %.1f, got %.1f", cfg.Radius, got.Radius)
	}
	if got.SizeBorder != cfg.SizeBorder {
		t.Fatalf("expected border %.1f, got %.1f", cfg.SizeBorder, got.SizeBorder)
	}
}

func TestFormValidationHelpers(t *testing.T) {
	if got := validateUsernameSync(""); got != "username required" {
		t.Fatalf("unexpected username required result: %q", got)
	}
	if got := validateUsernameSync("ab"); got != "username min length is 3" {
		t.Fatalf("unexpected username length result: %q", got)
	}
	if got := validateEmailSync("userexample.com"); got != "email must contain @" {
		t.Fatalf("unexpected email validation result: %q", got)
	}
	if got := validateAgeSync(""); got != "age required" {
		t.Fatalf("unexpected age validation result: %q", got)
	}
}

func TestValidateUsernameReserved(t *testing.T) {
	if got := validateUsernameReserved("admin"); got != "username already taken" {
		t.Fatalf("unexpected reserved username result: %q", got)
	}
	if got := validateUsernameReserved("available"); got != "" {
		t.Fatalf("expected no issue for available, got %q", got)
	}
}
