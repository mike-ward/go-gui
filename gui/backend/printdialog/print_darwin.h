#ifndef PRINTDIALOG_DARWIN_H
#define PRINTDIALOG_DARWIN_H

// Status codes matching gui.PrintRunStatus.
enum {
    PRINT_OK     = 0,
    PRINT_CANCEL = 1,
    PRINT_ERROR  = 2,
};

// PrintParams mirrors gui.NativePrintParams for the C bridge.
typedef struct {
    const char *title;
    const char *jobName;
    const char *pdfPath;
    double      paperWidth;   // points
    double      paperHeight;  // points
    double      marginTop;    // points
    double      marginRight;
    double      marginBottom;
    double      marginLeft;
    int         orientation;  // 0=portrait, 1=landscape
    int         copies;
    const char *pageRanges;   // "1-3,5" or ""
    int         duplexMode;
    int         colorMode;
    int         scaleMode;
} PrintParams;

// PrintResult returned by the bridge.
typedef struct {
    int   status;
    char *errorMessage;
    char *pdfPath;
} PrintResult;

PrintResult printdialogShow(PrintParams p);
void        printdialogFreeResult(PrintResult r);

#endif
