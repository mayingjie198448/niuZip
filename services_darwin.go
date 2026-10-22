//go:build darwin

package main

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework Cocoa -framework UniformTypeIdentifiers

extern void initMacServices(void);
*/
import "C"
import (
	"os/exec"
	"path/filepath"
	"strings"
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

func autoEnableServicesInPBS() {
	out, _ := exec.Command("defaults", "read", "pbs", "NSServicesStatus").Output()
	statusStr := string(out)

	modified := false
	extractKey := "com.niuzip.app - niuZip: 解压到当前文件夹 - extractHereService"
	compressKey := "com.niuzip.app - niuZip: 快速压缩为 ZIP - compressToZipService"

	if !strings.Contains(statusStr, extractKey) {
		_ = exec.Command("defaults", "write", "pbs", "NSServicesStatus", "-dict-add",
			extractKey,
			`{ "enabled_context_menu" = 1; "enabled_services_menu" = 1; "presentation_modes" = { "ContextMenu" = 1; "ServicesMenu" = 1; }; }`,
		).Run()
		modified = true
	}

	if !strings.Contains(statusStr, compressKey) {
		_ = exec.Command("defaults", "write", "pbs", "NSServicesStatus", "-dict-add",
			compressKey,
			`{ "enabled_context_menu" = 1; "enabled_services_menu" = 1; "presentation_modes" = { "ContextMenu" = 1; "ServicesMenu" = 1; }; }`,
		).Run()
		modified = true
	}

	if modified {
		_ = exec.Command("killall", "pbs").Run()
	}
}

func initPlatformServices() {
	C.initMacServices()
	go autoEnableServicesInPBS()
}
