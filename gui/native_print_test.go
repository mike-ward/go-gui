package gui

import "testing"

func TestExportPrintJobNoOutputPath(t *testing.T) {
	w := &Window{windowWidth: 800, windowHeight: 600}
	job := NewPrintJob()
	result := w.ExportPrintJob(job)
	if result.IsOk() {
		t.Error("expected error for empty output path")
	}
	if result.ErrorCode != printErrorInvalidCfg {
		t.Errorf("ErrorCode: got %q, want %q", result.ErrorCode, printErrorInvalidCfg)
	}
}

func TestExportPrintJobBadDPI(t *testing.T) {
	w := &Window{windowWidth: 800, windowHeight: 600}
	job := NewPrintJob()
	job.OutputPath = "/tmp/out.pdf"
	job.RasterDPI = 10
	result := w.ExportPrintJob(job)
	if result.IsOk() {
		t.Error("expected error for bad DPI")
	}
}

func TestExportPrintJobNoRenderers(t *testing.T) {
	w := &Window{windowWidth: 800, windowHeight: 600}
	job := NewPrintJob()
	job.OutputPath = "/tmp/out.pdf"
	result := w.ExportPrintJob(job)
	if result.IsOk() {
		t.Error("expected error for no renderers")
	}
	if result.ErrorCode != printErrorRender {
		t.Errorf("ErrorCode: got %q, want %q", result.ErrorCode, printErrorRender)
	}
}

func TestRunPrintJobNoPlatform(t *testing.T) {
	w := &Window{}
	job := NewPrintJob()
	result := w.RunPrintJob(job)
	if result.Status != PrintRunError {
		t.Errorf("Status: got %d, want %d", result.Status, PrintRunError)
	}
	if result.ErrorCode != "unsupported" {
		t.Errorf("ErrorCode: got %q", result.ErrorCode)
	}
}

func TestRunPrintJobBadCopies(t *testing.T) {
	w := &Window{}
	job := NewPrintJob()
	job.Copies = 0
	result := w.RunPrintJob(job)
	if result.Status != PrintRunError {
		t.Errorf("expected error status")
	}
	if result.ErrorCode != printErrorInvalidCfg {
		t.Errorf("ErrorCode: got %q", result.ErrorCode)
	}
}

func TestRunPrintJobPDFPathSource(t *testing.T) {
	w := &Window{}
	job := NewPrintJob()
	job.Source = PrintJobSource{Kind: PrintSourcePDFPath, PDFPath: "/path.pdf"}
	result := w.RunPrintJob(job)
	// No platform → unsupported.
	if result.Status != PrintRunError {
		t.Errorf("expected error")
	}
}

func TestPrintJobResolvePDFPathEmpty(t *testing.T) {
	w := &Window{}
	job := NewPrintJob()
	job.Source = PrintJobSource{Kind: PrintSourcePDFPath}
	_, err := printJobResolvePDFPath(w, job)
	if err == nil {
		t.Error("expected error for empty pdf_path")
	}
}

func TestPrintJobResolvePDFPathOK(t *testing.T) {
	w := &Window{}
	job := NewPrintJob()
	job.Source = PrintJobSource{Kind: PrintSourcePDFPath, PDFPath: "/test.pdf"}
	path, err := printJobResolvePDFPath(w, job)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if path != "/test.pdf" {
		t.Errorf("path: got %q", path)
	}
}

func TestPrintJobResolveCurrentViewNotImpl(t *testing.T) {
	w := &Window{}
	job := NewPrintJob()
	_, err := printJobResolvePDFPath(w, job)
	if err == nil {
		t.Error("expected not-implemented error")
	}
}
