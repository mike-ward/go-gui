#ifndef FILEDIALOG_DARWIN_H
#define FILEDIALOG_DARWIN_H

#include <stdint.h>

// Status codes matching gui.NativeDialogStatus.
enum {
    DIALOG_OK     = 0,
    DIALOG_CANCEL = 1,
    DIALOG_ERROR  = 2,
};

// DialogResult returned by all dialog bridge functions.
typedef struct {
    int       status;
    char    **paths;
    int       pathCount;
    void    **bookmarkData;
    int      *bookmarkLens;
    char     *errorMessage;
} DialogResult;

DialogResult filedialogOpen(const char *title, const char *startDir,
    const char **extensions, int extCount, int allowMultiple);

DialogResult filedialogSave(const char *title, const char *startDir,
    const char *defaultName, const char *defaultExt,
    const char **extensions, int extCount, int confirmOverwrite);

DialogResult filedialogFolder(const char *title, const char *startDir);

void filedialogFreeResult(DialogResult r);

#endif
