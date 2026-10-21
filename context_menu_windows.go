//go:build windows

package main

import (
	"fmt"
	"os"
	"path/filepath"

	"golang.org/x/sys/windows/registry"
)

// RegisterWindowsContextMenu adds "用 niuZip 打开" to Windows Explorer context menu for current user.
func (a *App) RegisterWindowsContextMenu() error {
	exePath, err := os.Executable()
	if err != nil {
		return err
	}
	exePath = filepath.Clean(exePath)

	extensions := []string{".zip", ".7z", ".rar", ".tar", ".gz", ".bz2", ".xz", ".iso", ".cab"}

	for _, ext := range extensions {
		subKey := fmt.Sprintf(`Software\Classes\SystemFileAssociations\%s\shell\niuZip`, ext)
		k, _, err := registry.CreateKey(registry.CURRENT_USER, subKey, registry.ALL_ACCESS)
		if err != nil {
			return err
		}
		_ = k.SetStringValue("", "用 niuZip 打开")
		_ = k.SetStringValue("Icon", fmt.Sprintf(`"%s",0`, exePath))
		k.Close()

		cmdKey, _, err := registry.CreateKey(registry.CURRENT_USER, subKey+`\command`, registry.ALL_ACCESS)
		if err != nil {
			return err
		}
		_ = cmdKey.SetStringValue("", fmt.Sprintf(`"%s" "%%1"`, exePath))
		cmdKey.Close()
	}

	// Also add for Directory (folder)
	dirSubKey := `Software\Classes\Directory\shell\niuZip`
	k, _, err := registry.CreateKey(registry.CURRENT_USER, dirSubKey, registry.ALL_ACCESS)
	if err == nil {
		_ = k.SetStringValue("", "用 niuZip 浏览/压缩")
		_ = k.SetStringValue("Icon", fmt.Sprintf(`"%s",0`, exePath))
		k.Close()

		cmdKey, _, err := registry.CreateKey(registry.CURRENT_USER, dirSubKey+`\command`, registry.ALL_ACCESS)
		if err == nil {
			_ = cmdKey.SetStringValue("", fmt.Sprintf(`"%s" "%%1"`, exePath))
			cmdKey.Close()
		}
	}

	return nil
}

// UnregisterWindowsContextMenu removes niuZip from Windows Explorer context menu.
func (a *App) UnregisterWindowsContextMenu() error {
	extensions := []string{".zip", ".7z", ".rar", ".tar", ".gz", ".bz2", ".xz", ".iso", ".cab"}

	for _, ext := range extensions {
		subKey := fmt.Sprintf(`Software\Classes\SystemFileAssociations\%s\shell\niuZip`, ext)
		_ = registry.DeleteKey(registry.CURRENT_USER, subKey+`\command`)
		_ = registry.DeleteKey(registry.CURRENT_USER, subKey)
	}

	dirSubKey := `Software\Classes\Directory\shell\niuZip`
	_ = registry.DeleteKey(registry.CURRENT_USER, dirSubKey+`\command`)
	_ = registry.DeleteKey(registry.CURRENT_USER, dirSubKey)

	return nil
}
