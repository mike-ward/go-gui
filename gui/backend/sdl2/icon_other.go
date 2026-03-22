//go:build !darwin || js

package sdl2

// setAppIcon is a no-op on non-macOS platforms; the window icon
// set via SDL is sufficient.
func setAppIcon([]byte) {}
