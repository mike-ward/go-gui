#ifndef SCROLL_PHASE_DARWIN_H
#define SCROLL_PHASE_DARWIN_H

#include <SDL.h>

// Install a local NSEvent scroll-wheel monitor for the given window.
// The monitor fires goScrollBegan when NSEventPhaseMayBegin or
// NSEventPhaseBegan is detected. Safe to call multiple times (one
// monitor is shared across all registered windows).
void scrollPhaseInit(SDL_Window *win);

// Remove the given window from the scroll phase monitor registry.
// The shared monitor is torn down when the last window is removed.
void scrollPhaseDestroy(SDL_Window *win);

// Implemented in Go — called from the NSEvent monitor when a
// trackpad touch-begin phase is detected for the given window.
extern void goScrollBegan(SDL_Window *win);

#endif
