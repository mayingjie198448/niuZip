//go:build darwin

package main

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework Cocoa -framework UniformTypeIdentifiers

extern void initMacServices(void);
*/
import "C"
import (
	"path/filepath"
)

//export goHandleExtractService
func goHandleExtractService(cPath *C.char) {
	path := C.GoString(cPath)
	if path == "" {
		return
	}
	destDir := filepath.Dir(path)
	_ = runExtract(path, destDir)
}

//export goHandleCompressService
func goHandleCompressService(cPath *C.char) {
	path := C.GoString(cPath)
	if path == "" {
		return
	}
	_ = runAddToZip(path)
}

func initPlatformServices() {
	C.initMacServices()
}
