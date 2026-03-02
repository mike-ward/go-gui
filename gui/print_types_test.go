package gui

import "testing"

func TestPrintPageSizeA4Portrait(t *testing.T) {
	w, h := PrintPageSize(PaperA4, PrintPortrait)
	if w != 595 || h != 842 {
		t.Errorf("A4 portrait: got %gx%g, want 595x842", w, h)
	}
}

func TestPrintPageSizeA4Landscape(t *testing.T) {
	w, h := PrintPageSize(PaperA4, PrintLandscape)
	if w != 842 || h != 595 {
		t.Errorf("A4 landscape: got %gx%g, want 842x595", w, h)
	}
}

func TestPrintPageSizeLetter(t *testing.T) {
	w, h := PrintPageSize(PaperLetter, PrintPortrait)
	if w != 612 || h != 792 {
		t.Errorf("Letter: got %gx%g, want 612x792", w, h)
	}
}

func TestPrintPageSizeLegal(t *testing.T) {
	w, h := PrintPageSize(PaperLegal, PrintPortrait)
	if w != 612 || h != 1008 {
		t.Errorf("Legal: got %gx%g, want 612x1008", w, h)
	}
}

func TestPrintPageSizeA3(t *testing.T) {
	w, h := PrintPageSize(PaperA3, PrintPortrait)
	if w != 842 || h != 1191 {
		t.Errorf("A3: got %gx%g, want 842x1191", w, h)
	}
}

func TestDefaultPrintMargins(t *testing.T) {
	m := DefaultPrintMargins()
	if m.Top != 36 || m.Right != 36 || m.Bottom != 36 || m.Left != 36 {
		t.Errorf("defaults: %+v", m)
	}
}

func TestValidatePrintMarginsNegative(t *testing.T) {
	if err := validatePrintMargins(595, 842, PrintMargins{Left: -1}); err == nil {
		t.Error("expected error for negative margin")
	}
}

func TestValidatePrintMarginsExceedWidth(t *testing.T) {
	if err := validatePrintMargins(100, 842, PrintMargins{Left: 60, Right: 60}); err == nil {
		t.Error("expected error for margins exceeding width")
	}
}

func TestValidatePrintMarginsExceedHeight(t *testing.T) {
	if err := validatePrintMargins(595, 100, PrintMargins{Top: 60, Bottom: 60}); err == nil {
		t.Error("expected error for margins exceeding height")
	}
}

func TestValidatePrintMarginsOK(t *testing.T) {
	if err := validatePrintMargins(595, 842, DefaultPrintMargins()); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestValidatePrintJobDefaults(t *testing.T) {
	job := NewPrintJob()
	if err := validatePrintJob(job); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestValidatePrintJobZeroCopies(t *testing.T) {
	job := NewPrintJob()
	job.Copies = 0
	if err := validatePrintJob(job); err == nil {
		t.Error("expected error for zero copies")
	}
}

func TestValidatePrintJobBadPageRange(t *testing.T) {
	job := NewPrintJob()
	job.PageRanges = []PrintPageRange{{From: 0, To: 5}}
	if err := validatePrintJob(job); err == nil {
		t.Error("expected error for page range starting at 0")
	}
}

func TestValidatePrintJobReversedRange(t *testing.T) {
	job := NewPrintJob()
	job.PageRanges = []PrintPageRange{{From: 5, To: 3}}
	if err := validatePrintJob(job); err == nil {
		t.Error("expected error for reversed page range")
	}
}

func TestValidatePrintJobBadDPI(t *testing.T) {
	job := NewPrintJob()
	job.RasterDPI = 50
	if err := validatePrintJob(job); err == nil {
		t.Error("expected error for low DPI")
	}
}

func TestValidatePrintJobBadQuality(t *testing.T) {
	job := NewPrintJob()
	job.JPEGQuality = 5
	if err := validatePrintJob(job); err == nil {
		t.Error("expected error for low quality")
	}
}

func TestValidatePrintJobPDFPathRequired(t *testing.T) {
	job := NewPrintJob()
	job.Source = PrintJobSource{Kind: PrintSourcePDFPath}
	if err := validatePrintJob(job); err == nil {
		t.Error("expected error for empty pdf_path")
	}
}

func TestValidateExportPrintJobNoOutputPath(t *testing.T) {
	job := NewPrintJob()
	if err := validateExportPrintJob(job); err == nil {
		t.Error("expected error for empty output_path")
	}
}

func TestValidateExportPrintJobOK(t *testing.T) {
	job := NewPrintJob()
	job.OutputPath = "/tmp/out.pdf"
	if err := validateExportPrintJob(job); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestExtractPrintTokens(t *testing.T) {
	tests := []struct {
		input string
		want  int
	}{
		{"Page {page} of {pages}", 2},
		{"{date}", 1},
		{"no tokens", 0},
		{"{page}{title}", 2},
		{"{}", 0},
		{"{{nested}}", 1},
	}
	for _, tt := range tests {
		tokens := extractPrintTokens(tt.input)
		if len(tokens) != tt.want {
			t.Errorf("extractPrintTokens(%q): got %d tokens %v, want %d", tt.input, len(tokens), tokens, tt.want)
		}
	}
}

func TestValidatePrintTokenOK(t *testing.T) {
	for _, tok := range []string{"page", "pages", "date", "title", "job"} {
		if err := validatePrintToken(tok); err != nil {
			t.Errorf("validatePrintToken(%q) unexpected error: %v", tok, err)
		}
	}
}

func TestValidatePrintTokenBad(t *testing.T) {
	if err := validatePrintToken("foo"); err == nil {
		t.Error("expected error for unsupported token")
	}
}

func TestValidateHeaderFooterOK(t *testing.T) {
	cfg := PrintHeaderFooterCfg{Enabled: true, Left: "Page {page}"}
	if err := validateHeaderFooterCfg(cfg); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestValidateHeaderFooterBadToken(t *testing.T) {
	cfg := PrintHeaderFooterCfg{Enabled: true, Center: "{invalid}"}
	if err := validateHeaderFooterCfg(cfg); err == nil {
		t.Error("expected error for bad token")
	}
}

func TestValidateHeaderFooterDisabled(t *testing.T) {
	cfg := PrintHeaderFooterCfg{Enabled: false, Center: "{invalid}"}
	if err := validateHeaderFooterCfg(cfg); err != nil {
		t.Errorf("disabled header should skip validation: %v", err)
	}
}

func TestNormalizePrintPageRangesEmpty(t *testing.T) {
	result := NormalizePrintPageRanges(nil)
	if result != nil {
		t.Errorf("expected nil, got %v", result)
	}
}

func TestNormalizePrintPageRangesSingle(t *testing.T) {
	result := NormalizePrintPageRanges([]PrintPageRange{{From: 1, To: 5}})
	if len(result) != 1 || result[0].From != 1 || result[0].To != 5 {
		t.Errorf("got %v", result)
	}
}

func TestNormalizePrintPageRangesMerge(t *testing.T) {
	input := []PrintPageRange{{From: 1, To: 3}, {From: 4, To: 6}}
	result := NormalizePrintPageRanges(input)
	if len(result) != 1 || result[0].From != 1 || result[0].To != 6 {
		t.Errorf("expected merged [1-6], got %v", result)
	}
}

func TestNormalizePrintPageRangesOverlap(t *testing.T) {
	input := []PrintPageRange{{From: 1, To: 5}, {From: 3, To: 8}}
	result := NormalizePrintPageRanges(input)
	if len(result) != 1 || result[0].From != 1 || result[0].To != 8 {
		t.Errorf("expected merged [1-8], got %v", result)
	}
}

func TestNormalizePrintPageRangesDisjoint(t *testing.T) {
	input := []PrintPageRange{{From: 1, To: 3}, {From: 10, To: 12}}
	result := NormalizePrintPageRanges(input)
	if len(result) != 2 {
		t.Errorf("expected 2 ranges, got %v", result)
	}
}

func TestNormalizePrintPageRangesSort(t *testing.T) {
	input := []PrintPageRange{{From: 10, To: 12}, {From: 1, To: 3}}
	result := NormalizePrintPageRanges(input)
	if len(result) != 2 || result[0].From != 1 || result[1].From != 10 {
		t.Errorf("expected sorted, got %v", result)
	}
}

func TestPrintPageRangesToString(t *testing.T) {
	tests := []struct {
		input []PrintPageRange
		want  string
	}{
		{nil, ""},
		{[]PrintPageRange{{1, 1}}, "1"},
		{[]PrintPageRange{{1, 5}}, "1-5"},
		{[]PrintPageRange{{1, 3}, {5, 7}}, "1-3,5-7"},
	}
	for _, tt := range tests {
		got := printPageRangesToString(tt.input)
		if got != tt.want {
			t.Errorf("printPageRangesToString(%v) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestPrintExportResultIsOk(t *testing.T) {
	ok := printExportOKResult("/tmp/out.pdf")
	if !ok.IsOk() {
		t.Error("expected IsOk true")
	}
	fail := printExportErrorResult("/tmp/out.pdf", "err", "msg")
	if fail.IsOk() {
		t.Error("expected IsOk false")
	}
}

func TestNewPrintJob(t *testing.T) {
	job := NewPrintJob()
	if job.Paper != PaperA4 {
		t.Errorf("Paper: got %d, want %d", job.Paper, PaperA4)
	}
	if job.Copies != 1 {
		t.Errorf("Copies: got %d, want 1", job.Copies)
	}
	if job.RasterDPI != 300 {
		t.Errorf("RasterDPI: got %d, want 300", job.RasterDPI)
	}
	if job.JPEGQuality != 85 {
		t.Errorf("JPEGQuality: got %d, want 85", job.JPEGQuality)
	}
}
