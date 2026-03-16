//go:build darwin

package gui

import "C"
import "unsafe"

//export goGetWindowCount
func goGetWindowCount() C.int {
	if activeHost == nil {
		return 0
	}
	return C.int(activeHost.WindowCount())
}

//export goGetWindowID
func goGetWindowID(index C.int) *C.char {
	if activeHost == nil {
		return nil
	}
	list := activeHost.WindowList()
	i := int(index)
	if i < 0 || i >= len(list) {
		return nil
	}
	return C.CString(list[i].ID)
}

//export goGetWindowTitle
func goGetWindowTitle(index C.int) *C.char {
	if activeHost == nil {
		return nil
	}
	list := activeHost.WindowList()
	i := int(index)
	if i < 0 || i >= len(list) {
		return nil
	}
	return C.CString(list[i].Filename)
}

//export goDockMenuActivate
func goDockMenuActivate(windowID *C.char) {
	if activeHost == nil {
		return
	}
	id := C.GoString(windowID)
	activeHost.primaryWV.Dispatch(func() {
		activeHost.ActivateWindow(id)
	})
}

//export goDockMenuClose
func goDockMenuClose(windowID *C.char) {
	if activeHost == nil {
		return
	}
	id := C.GoString(windowID)
	activeHost.primaryWV.Dispatch(func() {
		activeHost.CloseWindow(id)
	})
}

//export goDockMenuOpenFile
func goDockMenuOpenFile(path *C.char) {
	if activeHost == nil {
		return
	}
	p := C.GoString(path)
	go activeHost.OpenFile(p)
}

//export goDockMenuQuit
func goDockMenuQuit() {
	if activeHost == nil {
		return
	}
	activeHost.primaryWV.Dispatch(func() {
		activeHost.closeAllWindows()
	})
}

// Ensure unsafe is used (required for //export files).
var _ = unsafe.Pointer(nil)
