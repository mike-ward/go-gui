#import <Cocoa/Cocoa.h>
#include <SDL.h>
#include <SDL_syswm.h>
#include "filedrop_darwin.h"

// ─── Drop target view ───────────────────────────────────────────

// GUIDropView is a transparent overlay that registers as a
// NSDraggingDestination on the SDL window's content view.
// It forwards performDragOperation to goFileDrop for each
// dropped file URL.
@interface GUIDropView : NSView <NSDraggingDestination>
@property (nonatomic, assign) SDL_Window *sdlWindow;
@end

@implementation GUIDropView

- (NSDragOperation)draggingEntered:(id<NSDraggingInfo>)sender {
    NSPasteboard *pb = [sender draggingPasteboard];
    if ([pb canReadObjectForClasses:@[[NSURL class]]
                            options:@{NSPasteboardURLReadingFileURLsOnlyKey: @YES}]) {
        return NSDragOperationCopy;
    }
    return NSDragOperationNone;
}

- (BOOL)performDragOperation:(id<NSDraggingInfo>)sender {
    NSPasteboard *pb = [sender draggingPasteboard];
    NSArray<NSURL *> *urls =
        [pb readObjectsForClasses:@[[NSURL class]]
                          options:@{NSPasteboardURLReadingFileURLsOnlyKey: @YES}];
    if (!urls || urls.count == 0) {
        return NO;
    }
    for (NSURL *url in urls) {
        if (![url isFileURL]) continue;
        const char *utf8 = [[url path] UTF8String];
        if (utf8) {
            goFileDrop(self.sdlWindow, (char *)utf8);
        }
    }
    return YES;
}

- (BOOL)prepareForDragOperation:(id<NSDraggingInfo>)sender {
    return YES;
}

@end

// ─── Public API ─────────────────────────────────────────────────

static NSMutableArray<GUIDropView *> *_dropViews;

void fileDropInit(SDL_Window *win) {
    SDL_SysWMinfo info;
    SDL_VERSION(&info.version);
    if (!SDL_GetWindowWMInfo(win, &info)) {
        return;
    }
    NSWindow *nsWindow = info.info.cocoa.window;
    NSView *contentView = [nsWindow contentView];
    if (!contentView) {
        return;
    }

    if (!_dropViews) {
        _dropViews = [NSMutableArray new];
    }

    GUIDropView *dropView =
        [[GUIDropView alloc] initWithFrame:contentView.bounds];
    dropView.sdlWindow = win;
    dropView.autoresizingMask = NSViewWidthSizable | NSViewHeightSizable;
    [contentView addSubview:dropView];
    [dropView registerForDraggedTypes:@[NSPasteboardTypeFileURL]];
    [_dropViews addObject:dropView];
}

void fileDropDestroy(SDL_Window *win) {
    if (!_dropViews) return;
    for (NSUInteger i = 0; i < _dropViews.count; i++) {
        GUIDropView *dv = _dropViews[i];
        if (dv.sdlWindow == win) {
            [dv unregisterDraggedTypes];
            [dv removeFromSuperview];
            [_dropViews removeObjectAtIndex:i];
            return;
        }
    }
}
