#import <Cocoa/Cocoa.h>
#include <SDL.h>
#include <SDL_syswm.h>
#include "scroll_phase_darwin.h"

// Maps NSWindow * → NSValue * (wrapping SDL_Window *).
// One entry per registered go-gui window.
static NSMapTable *_scrollWindows;

// The single shared local event monitor (nil until first window registered).
static id _scrollMonitor;

void scrollPhaseInit(SDL_Window *win) {
    SDL_SysWMinfo info;
    SDL_VERSION(&info.version);
    if (!SDL_GetWindowWMInfo(win, &info)) {
        return;
    }
    NSWindow *nsWin = info.info.cocoa.window;
    if (!nsWin) {
        return;
    }

    if (!_scrollWindows) {
        _scrollWindows = [NSMapTable strongToStrongObjectsMapTable];
    }
    [_scrollWindows setObject:[NSValue valueWithPointer:win] forKey:nsWin];

    if (!_scrollMonitor) {
        _scrollMonitor = [NSEvent
            addLocalMonitorForEventsMatchingMask:NSEventMaskScrollWheel
            handler:^NSEvent *(NSEvent *event) {
                NSEventPhase phase = event.phase;
                if (phase == NSEventPhaseMayBegin || phase == NSEventPhaseBegan) {
                    NSWindow *evWin = event.window;
                    NSValue *val = evWin ? [_scrollWindows objectForKey:evWin] : nil;
                    if (val) {
                        goScrollBegan((SDL_Window *)[val pointerValue]);
                    }
                }
                return event;
            }];
    }
}

void scrollPhaseDestroy(SDL_Window *win) {
    SDL_SysWMinfo info;
    SDL_VERSION(&info.version);
    if (_scrollWindows && SDL_GetWindowWMInfo(win, &info)) {
        NSWindow *nsWin = info.info.cocoa.window;
        if (nsWin) {
            [_scrollWindows removeObjectForKey:nsWin];
        }
    }
    if (_scrollMonitor && (!_scrollWindows || _scrollWindows.count == 0)) {
        [NSEvent removeMonitor:_scrollMonitor];
        _scrollMonitor = nil;
    }
}
