//go:build windows

package main

import (
	"fmt"
	"os"
	"path/filepath"

	"golang.org/x/sys/windows/registry"
)

// RegisterWindowsContextMenu adds niuZip shortcut actions to Windows Explorer context menu for current user.
func (a *App) RegisterWindowsContextMenu() error {
	exePath, err := os.Executable()
	if err != nil {
		return err
	}
	exePath = filepath.Clean(exePath)

	extensions := []string{".zip", ".7z", ".rar", ".tar", ".gz", ".bz2", ".xz", ".iso", ".cab", ".apk", ".jar"}

	// 1. Cascading menu for archives: Open / Extract Here / Extract to Dir
	for _, ext := range extensions {
		subKey := fmt.Sprintf(`Software\Classes\SystemFileAssociations\%s\shell\niuZip`, ext)
		// Clean up any legacy flat registration from older versions (e.g. v1.0.1 command subkey)
		_ = registry.DeleteKey(registry.CURRENT_USER, subKey+`\command`)
		_ = registry.DeleteKey(registry.CURRENT_USER, subKey)

		k, _, err := registry.CreateKey(registry.CURRENT_USER, subKey, registry.ALL_ACCESS)
		if err != nil {
			continue
		}
		_ = k.SetStringValue("MUIVerb", "niuZip")
		_ = k.SetStringValue("Icon", fmt.Sprintf(`"%s",0`, exePath))
		_ = k.SetStringValue("SubCommands", "")
		k.Close()

		// Shell\open
		openKey, _, err := registry.CreateKey(registry.CURRENT_USER, subKey+`\Shell\open\command`, registry.ALL_ACCESS)
		if err == nil {
			_ = openKey.SetStringValue("", fmt.Sprintf(`"%s" "%%1"`, exePath))
			openKey.Close()
			if p, _, err := registry.CreateKey(registry.CURRENT_USER, subKey+`\Shell\open`, registry.ALL_ACCESS); err == nil {
				_ = p.SetStringValue("", "用 niuZip 打开")
				_ = p.SetStringValue("Icon", fmt.Sprintf(`"%s",0`, exePath))
				p.Close()
			}
		}

		// Shell\extracthere
		hereKey, _, err := registry.CreateKey(registry.CURRENT_USER, subKey+`\Shell\extracthere\command`, registry.ALL_ACCESS)
		if err == nil {
			_ = hereKey.SetStringValue("", fmt.Sprintf(`"%s" --extract-here "%%1"`, exePath))
			hereKey.Close()
			if p, _, err := registry.CreateKey(registry.CURRENT_USER, subKey+`\Shell\extracthere`, registry.ALL_ACCESS); err == nil {
				_ = p.SetStringValue("", "解压到当前文件夹")
				_ = p.SetStringValue("Icon", fmt.Sprintf(`"%s",0`, exePath))
				p.Close()
			}
		}

		// Shell\extracttodir
		dirKey, _, err := registry.CreateKey(registry.CURRENT_USER, subKey+`\Shell\extracttodir\command`, registry.ALL_ACCESS)
		if err == nil {
			_ = dirKey.SetStringValue("", fmt.Sprintf(`"%s" --extract-to-dir "%%1"`, exePath))
			dirKey.Close()
			if p, _, err := registry.CreateKey(registry.CURRENT_USER, subKey+`\Shell\extracttodir`, registry.ALL_ACCESS); err == nil {
				_ = p.SetStringValue("", "解压到同名文件夹")
				_ = p.SetStringValue("Icon", fmt.Sprintf(`"%s",0`, exePath))
				p.Close()
			}
		}
	}

	// 2. Cascading menu for Folders (Directory) and Files (*): Compress to ZIP / Add to Archive...
	for _, target := range []string{`Directory`, `*`} {
		targetKey := fmt.Sprintf(`Software\Classes\%s\shell\niuZip`, target)
		// Clean up any legacy flat registration from older versions (e.g. v1.0.1 command subkey)
		_ = registry.DeleteKey(registry.CURRENT_USER, targetKey+`\command`)
		_ = registry.DeleteKey(registry.CURRENT_USER, targetKey)

		k, _, err := registry.CreateKey(registry.CURRENT_USER, targetKey, registry.ALL_ACCESS)
		if err != nil {
			continue
		}
		_ = k.SetStringValue("MUIVerb", "niuZip")
		_ = k.SetStringValue("Icon", fmt.Sprintf(`"%s",0`, exePath))
		_ = k.SetStringValue("SubCommands", "")
		k.Close()

		// Shell\addtozip
		addZipKey, _, err := registry.CreateKey(registry.CURRENT_USER, targetKey+`\Shell\addtozip\command`, registry.ALL_ACCESS)
		if err == nil {
			_ = addZipKey.SetStringValue("", fmt.Sprintf(`"%s" --add-to-zip "%%1"`, exePath))
			addZipKey.Close()
			if p, _, err := registry.CreateKey(registry.CURRENT_USER, targetKey+`\Shell\addtozip`, registry.ALL_ACCESS); err == nil {
				_ = p.SetStringValue("", "立即压缩为 ZIP")
				_ = p.SetStringValue("Icon", fmt.Sprintf(`"%s",0`, exePath))
				p.Close()
			}
		}

		// Shell\adddialog
		addDlgKey, _, err := registry.CreateKey(registry.CURRENT_USER, targetKey+`\Shell\adddialog\command`, registry.ALL_ACCESS)
		if err == nil {
			_ = addDlgKey.SetStringValue("", fmt.Sprintf(`"%s" --add-dialog "%%1"`, exePath))
			addDlgKey.Close()
			if p, _, err := registry.CreateKey(registry.CURRENT_USER, targetKey+`\Shell\adddialog`, registry.ALL_ACCESS); err == nil {
				_ = p.SetStringValue("", "添加到压缩文件...")
				_ = p.SetStringValue("Icon", fmt.Sprintf(`"%s",0`, exePath))
				p.Close()
			}
		}
	}

	return nil
}

// UnregisterWindowsContextMenu removes niuZip from Windows Explorer context menu.
func (a *App) UnregisterWindowsContextMenu() error {
	extensions := []string{".zip", ".7z", ".rar", ".tar", ".gz", ".bz2", ".xz", ".iso", ".cab", ".apk", ".jar"}

	for _, ext := range extensions {
		subKey := fmt.Sprintf(`Software\Classes\SystemFileAssociations\%s\shell\niuZip`, ext)
		_ = registry.DeleteKey(registry.CURRENT_USER, subKey+`\Shell\open\command`)
		_ = registry.DeleteKey(registry.CURRENT_USER, subKey+`\Shell\open`)
		_ = registry.DeleteKey(registry.CURRENT_USER, subKey+`\Shell\extracthere\command`)
		_ = registry.DeleteKey(registry.CURRENT_USER, subKey+`\Shell\extracthere`)
		_ = registry.DeleteKey(registry.CURRENT_USER, subKey+`\Shell\extracttodir\command`)
		_ = registry.DeleteKey(registry.CURRENT_USER, subKey+`\Shell\extracttodir`)
		_ = registry.DeleteKey(registry.CURRENT_USER, subKey+`\Shell`)
		_ = registry.DeleteKey(registry.CURRENT_USER, subKey+`\command`)
		_ = registry.DeleteKey(registry.CURRENT_USER, subKey)
	}

	for _, target := range []string{`Directory`, `*`} {
		targetKey := fmt.Sprintf(`Software\Classes\%s\shell\niuZip`, target)
		_ = registry.DeleteKey(registry.CURRENT_USER, targetKey+`\Shell\addtozip\command`)
		_ = registry.DeleteKey(registry.CURRENT_USER, targetKey+`\Shell\addtozip`)
		_ = registry.DeleteKey(registry.CURRENT_USER, targetKey+`\Shell\adddialog\command`)
		_ = registry.DeleteKey(registry.CURRENT_USER, targetKey+`\Shell\adddialog`)
		_ = registry.DeleteKey(registry.CURRENT_USER, targetKey+`\Shell`)
		_ = registry.DeleteKey(registry.CURRENT_USER, targetKey+`\command`)
		_ = registry.DeleteKey(registry.CURRENT_USER, targetKey)
	}

	return nil
}
