//go:build darwin && !js

package sdl2

/*
#cgo LDFLAGS: -framework Cocoa
void setDockIcon(const void *data, int len);
*/
import "C"
import "unsafe"

// setAppIcon sets the macOS Dock icon from PNG data.
func setAppIcon(png []byte) {
	C.setDockIcon(unsafe.Pointer(&png[0]), C.int(len(png)))
}
