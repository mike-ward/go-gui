#import <AppKit/AppKit.h>
#import <Quartz/Quartz.h>
#include "print_darwin.h"
#include <stdlib.h>
#include <string.h>

PrintResult printdialogShow(PrintParams p) {
    @autoreleasepool {
        PrintResult result = {0};

        if (p.pdfPath == NULL || p.pdfPath[0] == '\0') {
            result.status = PRINT_ERROR;
            result.errorMessage = strdup("no PDF path provided");
            return result;
        }

        NSString *path = [NSString stringWithUTF8String:p.pdfPath];
        NSURL *url = [NSURL fileURLWithPath:path];
        PDFDocument *doc = [[PDFDocument alloc] initWithURL:url];
        if (doc == nil) {
            result.status = PRINT_ERROR;
            result.errorMessage = strdup("failed to load PDF");
            return result;
        }

        // Configure print info (copy to avoid mutating the
        // process-global singleton).
        NSPrintInfo *info = [[NSPrintInfo sharedPrintInfo] copy];
        [info setTopMargin:p.marginTop];
        [info setRightMargin:p.marginRight];
        [info setBottomMargin:p.marginBottom];
        [info setLeftMargin:p.marginLeft];

        if (p.paperWidth > 0 && p.paperHeight > 0) {
            NSSize paperSize = NSMakeSize(p.paperWidth, p.paperHeight);
            [info setPaperSize:paperSize];
        }

        if (p.orientation == 1) {
            [info setOrientation:NSPaperOrientationLandscape];
        } else {
            [info setOrientation:NSPaperOrientationPortrait];
        }

        // Create print operation from the PDF document.
        NSPrintOperation *op =
            [doc printOperationForPrintInfo:info
                               scalingMode:kPDFPrintPageScaleToFit
                                autoRotate:YES];
        [op setShowsPrintPanel:YES];
        [op setShowsProgressPanel:YES];

        if (p.jobName != NULL && p.jobName[0] != '\0') {
            [op setJobTitle:
                [NSString stringWithUTF8String:p.jobName]];
        }

        if (p.copies > 1) {
            [[info dictionary] setObject:@(p.copies)
                                  forKey:NSPrintCopies];
        }

        BOOL success = [op runOperation];
        if (success) {
            result.status = PRINT_OK;
            result.pdfPath = strdup(p.pdfPath);
        } else {
            // runOperation returns NO for cancel and error both.
            // No way to distinguish reliably; treat as cancel.
            result.status = PRINT_CANCEL;
        }
        return result;
    }
}

void printdialogFreeResult(PrintResult r) {
    free(r.errorMessage);
    free(r.pdfPath);
}
