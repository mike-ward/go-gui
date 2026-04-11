//go:build darwin && !ios

#ifndef FILEDIALOG_DARWIN_H
#define FILEDIALOG_DARWIN_H

#include <stdint.h>

// Status codes matching gui.NativeDialogStatus.
enum {
    DIALOG_OK      = 0,
    DIALOG_CANCEL  = 1,
    DIALOG_ERROR   = 2,
    DIALOG_DISCARD = 3,
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

// Alert level constants matching gui.NativeAlertLevel.
enum {
    ALERT_INFO     = 0,
    ALERT_WARNING  = 1,
    ALERT_CRITICAL = 2,
};

// AlertResult returned by message/confirm dialogs.
typedef struct {
    int   status;  // DIALOG_OK, DIALOG_CANCEL, or DIALOG_ERROR
    char *errorMessage;
} AlertResult;

AlertResult filedialogMessage(const char *title, const char *body,
    int level);

AlertResult filedialogConfirm(const char *title, const char *body,
    int level);

AlertResult filedialogSaveDiscard(const char *title, const char *body,
    int level);

#endif
