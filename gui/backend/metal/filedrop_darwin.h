#ifndef FILEDROP_DARWIN_H
#define FILEDROP_DARWIN_H

#include <SDL.h>

// Register the SDL window's content view as a file drop target.
void fileDropInit(SDL_Window *win);

// Remove file drop registration for the given window.
void fileDropDestroy(SDL_Window *win);

// Implemented in Go — called from ObjC when a file is dropped.
extern void goFileDrop(SDL_Window *win, char *path);

#endif
