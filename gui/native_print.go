package gui

import (
	"os"
	"strings"
)

// ExportPrintJob exports renderer output to PDF using PrintJob settings.
// Returns a PrintExportResult with status and path.
func (w *Window) ExportPrintJob(job PrintJob) PrintExportResult {
	if err := validateExportPrintJob(job); err != nil {
		return printExportErrorResult(job.OutputPath, printErrorInvalidCfg, err.Error())
	}

	sourceW := job.SourceWidth
	sourceH := job.SourceHeight

	w.Lock()
	if sourceW <= 0 {
		sourceW = float32(w.windowWidth)
	}
	if sourceH <= 0 {
		sourceH = float32(w.windowHeight)
	}
	renderersCopy := make([]RenderCmd, len(w.renderers))
	copy(renderersCopy, w.renderers)
	w.Unlock()

	if len(renderersCopy) == 0 {
		return printExportErrorResult(job.OutputPath, printErrorRender, "no renderers available for export")
	}
	if sourceW <= 0 || sourceH <= 0 {
		return printExportErrorResult(job.OutputPath, printErrorInvalidCfg, "source dimensions must be positive")
	}

	if err := renderToPDF(renderersCopy, job, sourceW, sourceH); err != nil {
		return printExportErrorResult(job.OutputPath, printErrorRender, err.Error())
	}
	return printExportOKResult(job.OutputPath)
}

// RunPrintJob runs the native print flow for the provided PrintJob.
func (w *Window) RunPrintJob(job PrintJob) PrintRunResult {
	if err := validatePrintJob(job); err != nil {
		return printRunErrorResult(printErrorInvalidCfg, err.Error())
	}
	if w.nativePlatform == nil {
		return printRunErrorResult("unsupported", "native print requires a platform backend")
	}

	pdfPath, err := printJobResolvePDFPath(w, job)
	if err != nil {
		code := printErrorInternal
		if job.Source.Kind == PrintSourcePDFPath {
			code = printErrorIO
		}
		return printRunErrorResult(code, err.Error())
	}
	// Clean up temp PDF after dialog returns.
	if job.Source.Kind == PrintSourceCurrentView {
		defer os.Remove(pdfPath)
	}

	pw, ph := PrintPageSize(job.Paper, job.Orientation)
	ranges := NormalizePrintPageRanges(job.PageRanges)

	result := w.nativePlatform.ShowPrintDialog(NativePrintParams{
		Title:        job.Title,
		JobName:      job.JobName,
		PDFPath:      pdfPath,
		PaperWidth:   pw,
		PaperHeight:  ph,
		MarginTop:    job.Margins.Top,
		MarginRight:  job.Margins.Right,
		MarginBottom: job.Margins.Bottom,
		MarginLeft:   job.Margins.Left,
		Orientation:  printOrientationToInt(job.Orientation),
		Copies:       job.Copies,
		PageRanges:   printPageRangesToString(ranges),
		DuplexMode:   int(job.Duplex),
		ColorMode:    int(job.ColorMode),
		ScaleMode:    int(job.ScaleMode),
	})
	return result
}

// printJobResolvePDFPath resolves the PDF path for the print job.
// For current_view source, exports to a temp PDF first.
// For pdf_path source, validates the provided path.
func printJobResolvePDFPath(w *Window, job PrintJob) (string, error) {
	switch job.Source.Kind {
	case PrintSourceCurrentView:
		tmp, err := os.CreateTemp("", "go-gui-print-*.pdf")
		if err != nil {
			return "", &printError{"failed to create temp file: " + err.Error()}
		}
		tmp.Close()
		exportJob := job
		exportJob.OutputPath = tmp.Name()
		result := w.ExportPrintJob(exportJob)
		if !result.IsOk() {
			os.Remove(tmp.Name())
			return "", &printError{result.ErrorMessage}
		}
		return tmp.Name(), nil
	case PrintSourcePDFPath:
		path := strings.TrimSpace(job.Source.PDFPath)
		if path == "" {
			return "", &printError{"pdf_path is required"}
		}
		return path, nil
	default:
		return "", &printError{"unknown source kind"}
	}
}

type printError struct{ msg string }

func (e *printError) Error() string { return e.msg }
