package cocoa

// #include "gui.h"
// #include <stdlib.h>
import "C"
import "unsafe"

func Run() {
	C.initialize()
	C.pollEvents()
}

func GetScreenSize() (int, int) {
	sz := C.getScreenSize()
	return int(sz.width), int(sz.height)
}

func AddRecentOpenedFile(str string) {
	cstr := C.CString(str)
	defer C.free(unsafe.Pointer(cstr))

	C.addRecentOpenedFile(cstr)
}

func (w NativeWindow) Alert(str string) {
	cstr := C.CString(str)
	defer C.free(unsafe.Pointer(cstr))

	C.alert(unsafe.Pointer(w), cstr)
}

type NativeWindow uintptr

func (w NativeWindow) RefreshContent() {
	C.refreshWindowContent(unsafe.Pointer(w))
}

func (w NativeWindow) GetSize() (int, int) {
	sz := C.getWindowSize(unsafe.Pointer(w))
	return int(sz.width), int(sz.height)
}

func (w NativeWindow) SetTitle(title string) {
	ctitle := C.CString(title)
	defer C.free(unsafe.Pointer(ctitle))

	C.setWindowTitle(unsafe.Pointer(w), ctitle)
}

func (w NativeWindow) SetSize(width, height int) {
	C.setWindowSize(unsafe.Pointer(w), C.int(width), C.int(height))
}

func NewWindow(title string, width, height int) NativeWindow {
	ctitle := C.CString(title)
	defer C.free(unsafe.Pointer(ctitle))

	return NativeWindow(unsafe.Pointer(C.newWindow(ctitle, C.int(width), C.int(height))))
}

func (w NativeWindow) Show() {
	C.showWindow(unsafe.Pointer(w))
}

func (w NativeWindow) MakeCurrentContext() {
	C.initWindowCurrentContext(unsafe.Pointer(w))
}

func (w NativeWindow) ToggleFullScreen() {
	C.toggleFullScreen(unsafe.Pointer(w))
}

func (w NativeWindow) SetStartupViewVisible(b bool) {
	var i C.int
	if b {
		i = 1
	} else {
		i = 0
	}
	C.setStartupViewVisible(unsafe.Pointer(w), i)
}
func (w NativeWindow) UpdatePlaybackInfo(left, right string, percent float64) {
	cleft := C.CString(left)
	defer C.free(unsafe.Pointer(cleft))

	cright := C.CString(right)
	defer C.free(unsafe.Pointer(cright))

	C.updatePlaybackInfo(unsafe.Pointer(w), cleft, cright, C.double(percent))
}
func (w NativeWindow) UpdateBufferInfo(speed string, percent float64) {
	cspeed := C.CString(speed)
	defer C.free(unsafe.Pointer(cspeed))

	C.updateBufferInfo(unsafe.Pointer(w), cspeed, C.double(percent))
}

func (w NativeWindow) ShowSubtitle(items []struct {
	Content string
	Style   int //0 -normal, 1 -italic, 2 -bold, 3 italic and bold
	Color   uint
}, posType int, x, y float64) uintptr {

	if len(items) == 0 {
		return 0
	}

	ctexts := make([]C.AttributedString, 0)
	for _, str := range items {
		cstr := C.CString(str.Content)
		defer C.free(unsafe.Pointer(cstr))

		ctexts = append(ctexts, C.AttributedString{cstr, C.int(str.Style), C.uint(str.Color)})
	}

	citem := &C.SubItem{&ctexts[0], C.int(len(ctexts)), C.int(posType), C.double(x), C.double(y)}

	return uintptr(C.showSubtitle(unsafe.Pointer(w), citem))
}

func (w NativeWindow) HideSubtitle(ptr uintptr) {
	C.hideSubtitle(unsafe.Pointer(w), unsafe.Pointer(ptr))
}

func (w NativeWindow) SetControlsVisible(b bool, autoHide bool) {
	C.setControlsVisible(unsafe.Pointer(w), b2i(b), b2i(autoHide))
}

func (w NativeWindow) SetSpinningVisible(b bool) {
	C.setSpinningVisible(unsafe.Pointer(w), b2i(b))
}

func (w NativeWindow) SetVolume(volume int) {
	C.setVolume(unsafe.Pointer(w), C.int(volume))
}

func (w NativeWindow) SetVolumeVisible(b bool) {
	C.setVolumeVisible(unsafe.Pointer(w), b2i(b))
}

func (w NativeWindow) Close() {
	C.closeWindow(unsafe.Pointer(w))
}