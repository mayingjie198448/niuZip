package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/mayingjie198448/niuZip/core/engine"
	"github.com/mayingjie198448/niuZip/core/vfs"
	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

// ArchiveSummary contains overview statistics, the virtual tree, and recommended flows.
type ArchiveSummary struct {
	Path                  string        `json:"path"`
	Filename              string        `json:"filename"`
	Format                string        `json:"format"`
	TotalFiles            int           `json:"total_files"`
	TotalFolders          int           `json:"total_folders"`
	TotalUncompressedSize int64         `json:"total_uncompressed_size"`
	TotalPackedSize       int64         `json:"total_packed_size"`
	HasEncryption         bool          `json:"has_encryption"`
	Tree                  *vfs.TreeNode `json:"tree"`
	Flows                 []FlowAction  `json:"flows"`
}

// FlowAction represents an intelligent one-click action recommended for the opened archive.
type FlowAction struct {
	ID       string `json:"id"`
	Type     string `json:"type"`      // "install_app", "install_font", "preview_media"
	Title    string `json:"title"`
	Subtitle string `json:"subtitle"`
	Icon     string `json:"icon"`
	Payload  string `json:"payload"`   // Target entry path
}

// ExtractOptions specifies extraction destinations and file filters.
type ExtractOptions struct {
	ArchivePath   string   `json:"archive_path"`
	DestDir       string   `json:"dest_dir"`
	SelectedFiles []string `json:"selected_files"`
	Password      string   `json:"password"`
}

// App represents the main application backend binding service.
type App struct {
	ctx            context.Context
	broker         *engine.Broker
	mu             sync.RWMutex
	currentPath    string
	cachedEntries  []vfs.FileEntry
	tempPreviewDir string
}

func NewApp() *App {
	tempBase, err := os.UserCacheDir()
	if err != nil {
		tempBase = os.TempDir()
	}
	previewDir := filepath.Join(tempBase, "niuzip", "previews")
	_ = os.MkdirAll(previewDir, 0755)

	return &App{
		broker:         engine.NewBroker(),
		tempPreviewDir: previewDir,
	}
}

// Startup is called by Wails when the desktop application launches.
func (a *App) Startup(ctx context.Context) {
	a.ctx = ctx
	initPlatformServices()

	wailsRuntime.EventsOn(ctx, "wails:file-drop", func(optionalData ...interface{}) {
		if len(optionalData) > 0 {
			if paths, ok := optionalData[0].([]string); ok && len(paths) > 0 {
				target := paths[0]
				// If it's an archive, open it immediately
				if a.broker.DetectFormatWithFile(target) != "" {
					_, _ = a.OpenArchive(target)
				} else {
					// Notify frontend that external files were dropped
					wailsRuntime.EventsEmit(ctx, "files-dropped", paths)
				}
			}
		}
	})

	// Check if an archive was passed as CLI argument (e.g. from context menu / "Open with")
	go func() {
		time.Sleep(350 * time.Millisecond)
		initial := a.GetInitialArchive()
		if initial != "" {
			_, _ = a.OpenArchive(initial)
		}
	}()
}

// GetInitialArchive returns any archive path passed via command-line arguments (e.g. from right-click context menu).
func (a *App) GetInitialArchive() string {
	for i := 1; i < len(os.Args); i++ {
		arg := os.Args[i]
		if strings.HasPrefix(arg, "-") {
			continue
		}
		if i > 1 && os.Args[i-1] == "--add-dialog" {
			continue
		}
		if fi, err := os.Stat(arg); err == nil && !fi.IsDir() {
			return arg
		}
	}
	return ""
}

// GetInitialCreatePaths returns any file or folder paths passed via command-line arguments to be compressed.
func (a *App) GetInitialCreatePaths() []string {
	var paths []string
	addDialogMode := false

	for i := 1; i < len(os.Args); i++ {
		arg := os.Args[i]
		if arg == "--add-dialog" {
			addDialogMode = true
			continue
		}
		if strings.HasPrefix(arg, "-") {
			continue
		}
		if fi, err := os.Stat(arg); err == nil {
			if addDialogMode || fi.IsDir() {
				paths = append(paths, arg)
			}
		}
	}
	return paths
}

// Shutdown is called when the application closes; cleans up cached preview files.
func (a *App) Shutdown(ctx context.Context) {
	if a.tempPreviewDir != "" {
		_ = os.RemoveAll(a.tempPreviewDir)
	}
}

// OpenArchive opens and indexes an archive without extracting payload data.
func (a *App) OpenArchive(archivePath string) (*ArchiveSummary, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	entries, err := a.broker.List(context.Background(), archivePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open archive: %w", err)
	}

	a.currentPath = archivePath
	a.cachedEntries = entries

	var totalFiles, totalFolders int
	var totalUncompressed, totalPacked int64
	var hasEncryption bool

	for _, e := range entries {
		if e.IsDir {
			totalFolders++
		} else {
			totalFiles++
			totalUncompressed += e.Size
			totalPacked += e.PackedSize
		}
		if e.IsEncrypted {
			hasEncryption = true
		}
	}

	tree := vfs.BuildTree(entries)

	summary := &ArchiveSummary{
		Path:                  archivePath,
		Filename:              filepath.Base(archivePath),
		Format:                a.broker.DetectFormatWithFile(archivePath),
		TotalFiles:            totalFiles,
		TotalFolders:          totalFolders,
		TotalUncompressedSize: totalUncompressed,
		TotalPackedSize:       totalPacked,
		HasEncryption:         hasEncryption,
		Tree:                  tree,
	}

	summary.Flows = a.DetectFlows(summary)

	if a.ctx != nil {
		wailsRuntime.EventsEmit(a.ctx, "archive-opened", summary)
	}

	return summary, nil
}

// DetectFlows analyzes entries to suggest contextual actions (Smart Flows).
func (a *App) DetectFlows(summary *ArchiveSummary) []FlowAction {
	var flows []FlowAction

	// 1. macOS Application (.app bundle) - ONLY on macOS
	if runtime.GOOS == "darwin" {
		appRoots := make(map[string]bool)
		for _, e := range a.cachedEntries {
			clean := filepath.ToSlash(e.Path)
			// Must have genuine macOS app bundle structure (.app/Contents/Info.plist or .app/Contents/MacOS)
			idx := strings.Index(clean, ".app/")
			if idx != -1 {
				appDir := clean[:idx+4]
				if strings.Contains(clean, ".app/Contents/Info.plist") || strings.Contains(clean, ".app/Contents/MacOS") {
					appRoots[appDir] = true
				}
			}
		}

		for appDir := range appRoots {
			appName := filepath.Base(appDir)
			flows = append(flows, FlowAction{
				ID:       "flow_install_app",
				Type:     "install_app",
				Title:    fmt.Sprintf("一键安装 %s 到应用程序", appName),
				Subtitle: "自动解压并部署到系统的 /Applications 文件夹",
				Icon:     "🚀",
				Payload:  appDir,
			})
			break // Only suggest primary app
		}
	}

	// 2. Windows Installer (.exe / .msi) - ONLY on Windows
	if runtime.GOOS == "windows" {
		for _, e := range a.cachedEntries {
			if e.IsDir {
				continue
			}
			lower := strings.ToLower(e.Path)
			name := strings.ToLower(filepath.Base(lower))
			if name == "setup.exe" || name == "installer.exe" || strings.HasSuffix(name, ".msi") {
				flows = append(flows, FlowAction{
					ID:       "flow_run_installer",
					Type:     "run_installer",
					Title:    fmt.Sprintf("运行安装程序 %s", filepath.Base(e.Path)),
					Subtitle: "解压并在 Windows 中启动安装程序",
					Icon:     "⚙️",
					Payload:  e.Path,
				})
				break
			}
		}
	}

	// 3. Detect Font files (.ttf, .otf) - Cross-platform
	var fontFiles []string
	for _, e := range a.cachedEntries {
		if e.IsDir {
			continue
		}
		ext := strings.ToLower(filepath.Ext(e.Path))
		if ext == ".ttf" || ext == ".otf" {
			fontFiles = append(fontFiles, e.Path)
		}
	}
	if len(fontFiles) > 0 {
		var sub string
		if runtime.GOOS == "darwin" {
			sub = "将字体文件解压并安装至用户字体目录 (~/Library/Fonts)"
		} else {
			sub = "将字体文件解压并安装至 Windows 系统字体目录"
		}
		flows = append(flows, FlowAction{
			ID:       "flow_install_font",
			Type:     "install_font",
			Title:    fmt.Sprintf("一键安装 %d 个字体到系统", len(fontFiles)),
			Subtitle: sub,
			Icon:     "🔤",
			Payload:  strings.Join(fontFiles, ";"),
		})
	}

	return flows
}

// ExecuteFlow executes a recommended Smart Flow.
func (a *App) ExecuteFlow(action FlowAction) error {
	a.mu.RLock()
	archivePath := a.currentPath
	a.mu.RUnlock()

	if archivePath == "" {
		return errors.New("no archive is currently open")
	}

	switch action.Type {
	case "install_app":
		if runtime.GOOS != "darwin" {
			return errors.New("application bundle installation is only supported on macOS")
		}
		dest := "/Applications"
		return a.broker.Extract(context.Background(), archivePath, dest, nil, func(p float64, f string) {
			if a.ctx != nil {
				wailsRuntime.EventsEmit(a.ctx, "progress", map[string]interface{}{
					"percent": p,
					"file":    f,
					"stage":   "installing_app",
				})
			}
		})

	case "run_installer":
		tempDir, err := os.MkdirTemp("", "niuzip_install_*")
		if err != nil {
			return err
		}
		err = a.broker.Extract(context.Background(), archivePath, tempDir, []string{action.Payload}, nil)
		if err != nil {
			return err
		}
		exePath := filepath.Join(tempDir, action.Payload)
		return exec.Command("cmd", "/c", "start", "", exePath).Start()

	case "install_font":
		var fontDest string
		if runtime.GOOS == "darwin" {
			home, _ := os.UserHomeDir()
			fontDest = filepath.Join(home, "Library", "Fonts")
		} else {
			fontDest = filepath.Join(os.Getenv("WINDIR"), "Fonts")
		}
		_ = os.MkdirAll(fontDest, 0755)

		targets := strings.Split(action.Payload, ";")
		return a.broker.Extract(context.Background(), archivePath, fontDest, targets, func(p float64, f string) {
			if a.ctx != nil {
				wailsRuntime.EventsEmit(a.ctx, "progress", map[string]interface{}{
					"percent": p,
					"file":    f,
					"stage":   "installing_fonts",
				})
			}
		})

	case "preview_media":
		return a.QuickLook(action.Payload)

	default:
		return fmt.Errorf("unknown flow action type: %s", action.Type)
	}
}

// SearchEntries performs instant fuzzy search on the currently opened archive.
func (a *App) SearchEntries(query string) []vfs.FileEntry {
	a.mu.RLock()
	defer a.mu.RUnlock()

	return vfs.Search(a.cachedEntries, query)
}

// PreviewEntry extracts a single entry to a local temporary preview file.
func (a *App) PreviewEntry(entryPath string) (string, error) {
	a.mu.RLock()
	curPath := a.currentPath
	a.mu.RUnlock()

	if curPath == "" {
		return "", errors.New("no archive is currently open")
	}

	cleanRel := filepath.ToSlash(entryPath)
	targetDest := filepath.Join(a.tempPreviewDir, filepath.FromSlash(cleanRel))

	if err := os.MkdirAll(filepath.Dir(targetDest), 0755); err != nil {
		return "", err
	}

	if err := a.broker.ExtractSingle(context.Background(), curPath, cleanRel, targetDest); err != nil {
		return "", fmt.Errorf("failed to extract preview: %w", err)
	}

	return targetDest, nil
}

// QuickLook invokes the macOS QuickLook preview (qlmanage -p) or launches default system viewer.
func (a *App) QuickLook(entryPath string) error {
	filePath, err := a.PreviewEntry(entryPath)
	if err != nil {
		return err
	}

	switch runtime.GOOS {
	case "darwin":
		// Run qlmanage non-blocking in background
		go func() {
			cmd := exec.Command("qlmanage", "-p", filePath)
			_ = cmd.Run()
		}()
		return nil
	case "windows":
		cmd := exec.Command("cmd", "/c", "start", "", filePath)
		return cmd.Start()
	default:
		cmd := exec.Command("xdg-open", filePath)
		return cmd.Start()
	}
}

// OpenInSystem opens the extracted file using the operating system's default application and watches for edits.
func (a *App) OpenInSystem(entryPath string) error {
	filePath, err := a.PreviewEntry(entryPath)
	if err != nil {
		return err
	}

	initialFi, err := os.Stat(filePath)
	if err == nil {
		initialMod := initialFi.ModTime()
		// Start background watcher for external editor saves
		go func(targetEntry, localPath string, lastMod time.Time) {
			ticker := time.NewTicker(1 * time.Second)
			defer ticker.Stop()
			timeout := time.After(15 * time.Minute)

			for {
				select {
				case <-timeout:
					return
				case <-ticker.C:
					fi, statErr := os.Stat(localPath)
					if statErr != nil {
						return
					}
					if fi.ModTime().After(lastMod) {
						lastMod = fi.ModTime()
						if a.ctx != nil {
							wailsRuntime.EventsEmit(a.ctx, "archive-file-modified", map[string]string{
								"entryPath": targetEntry,
								"tempPath":  localPath,
							})
						}
					}
				}
			}
		}(entryPath, filePath, initialMod)
	}

	switch runtime.GOOS {
	case "darwin":
		cmd := exec.Command("open", filePath)
		return cmd.Start()
	case "windows":
		cmd := exec.Command("cmd", "/c", "start", "", filePath)
		return cmd.Start()
	default:
		cmd := exec.Command("xdg-open", filePath)
		return cmd.Start()
	}
}

// SelectArchiveDialog opens native OS open file dialog.
func (a *App) SelectArchiveDialog() (string, error) {
	if a.ctx == nil {
		return "", errors.New("wails context not initialized")
	}

	selected, err := wailsRuntime.OpenFileDialog(a.ctx, wailsRuntime.OpenDialogOptions{
		Title: "选择压缩包",
		Filters: []wailsRuntime.FileFilter{
			{
				DisplayName: "所有支持的归档 (*.zip, *.7z, *.tar, *.rar, *.gz, *.iso, *.dmg)",
				Pattern:     "*.zip;*.7z;*.tar;*.tar.gz;*.tgz;*.tbz;*.tbz2;*.tar.bz2;*.tar.xz;*.txz;*.rar;*.iso;*.dmg;*.cab;*.gz;*.bz2;*.xz;*.zipx;*.apk;*.jar",
			},
			{
				DisplayName: "所有文件 (*.*)",
				Pattern:     "*.*",
			},
		},
	})
	if err != nil {
		return "", err
	}
	return selected, nil
}

// SelectDestDirDialog opens native OS directory chooser dialog.
func (a *App) SelectDestDirDialog() (string, error) {
	if a.ctx == nil {
		return "", errors.New("wails context not initialized")
	}

	return wailsRuntime.OpenDirectoryDialog(a.ctx, wailsRuntime.OpenDialogOptions{
		Title: "选择解压目标文件夹",
	})
}

// SaveArchiveDialog opens native OS save file dialog.
func (a *App) SaveArchiveDialog(defaultName string) (string, error) {
	if a.ctx == nil {
		return "", errors.New("wails context not initialized")
	}

	return wailsRuntime.SaveFileDialog(a.ctx, wailsRuntime.SaveDialogOptions{
		Title:           "保存压缩包为",
		DefaultFilename: defaultName,
		Filters: []wailsRuntime.FileFilter{
			{DisplayName: "ZIP 归档 (*.zip)", Pattern: "*.zip"},
			{DisplayName: "7-Zip 归档 (*.7z)", Pattern: "*.7z"},
			{DisplayName: "TAR GZ 归档 (*.tar.gz)", Pattern: "*.tar.gz"},
		},
	})
}

// PathInfo represents metadata for a selected file or folder.
type PathInfo struct {
	Path  string `json:"path"`
	Name  string `json:"name"`
	IsDir bool   `json:"is_dir"`
	Size  int64  `json:"size"`
}

// GetPathInfos inspects file/folder paths and returns metadata.
func (a *App) GetPathInfos(paths []string) []PathInfo {
	var result []PathInfo
	for _, p := range paths {
		clean := filepath.Clean(p)
		fi, err := os.Stat(clean)
		if err != nil {
			continue
		}
		result = append(result, PathInfo{
			Path:  clean,
			Name:  fi.Name(),
			IsDir: fi.IsDir(),
			Size:  fi.Size(),
		})
	}
	return result
}

// SelectFilesToCompress opens multi-file chooser dialog.
func (a *App) SelectFilesToCompress() ([]string, error) {
	if a.ctx == nil {
		return nil, errors.New("wails context not initialized")
	}

	return wailsRuntime.OpenMultipleFilesDialog(a.ctx, wailsRuntime.OpenDialogOptions{
		Title: "选择要压缩的文件",
	})
}

// SelectFolderToCompress opens directory chooser dialog.
func (a *App) SelectFolderToCompress() (string, error) {
	if a.ctx == nil {
		return "", errors.New("wails context not initialized")
	}

	return wailsRuntime.OpenDirectoryDialog(a.ctx, wailsRuntime.OpenDialogOptions{
		Title: "选择要压缩的文件夹",
	})
}

// Extract extracts selected or all files from the current archive with progress events.
func (a *App) Extract(opts ExtractOptions) error {
	archivePath := opts.ArchivePath
	if archivePath == "" {
		a.mu.RLock()
		archivePath = a.currentPath
		a.mu.RUnlock()
	}

	if archivePath == "" {
		return errors.New("no archive specified")
	}

	progressHandler := func(percent float64, currentFile string) {
		if a.ctx != nil {
			wailsRuntime.EventsEmit(a.ctx, "progress", map[string]interface{}{
				"percent": percent,
				"file":    currentFile,
				"stage":   "extracting",
			})
		}
	}

	err := a.broker.Extract(context.Background(), archivePath, opts.DestDir, opts.SelectedFiles, progressHandler)
	if err != nil && a.ctx != nil {
		wailsRuntime.EventsEmit(a.ctx, "progress-error", err.Error())
	}
	return err
}

// Compress creates a new archive with progress events.
func (a *App) Compress(opts engine.CompressOptions) error {
	eng, err := a.broker.GetEngineForFile(opts.DestPath)
	if err != nil {
		if strings.HasSuffix(strings.ToLower(opts.DestPath), ".7z") {
			opts.Format = "7z"
		} else {
			opts.Format = "zip"
		}
		eng, err = a.broker.GetEngineForFile(opts.DestPath)
		if err != nil {
			return err
		}
	}

	progressHandler := func(percent float64, currentFile string) {
		if a.ctx != nil {
			wailsRuntime.EventsEmit(a.ctx, "progress", map[string]interface{}{
				"percent": percent,
				"file":    currentFile,
				"stage":   "compressing",
			})
		}
	}

	return eng.Create(context.Background(), opts, progressHandler)
}

// DeleteEntries performs in-place deletion of files from the current ZIP archive.
func (a *App) DeleteEntries(targetPaths []string) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.currentPath == "" {
		return errors.New("no archive is open")
	}

	zipEng := engine.NewPureZipEngine()
	if err := zipEng.DeleteEntries(context.Background(), a.currentPath, targetPaths); err != nil {
		return err
	}

	entries, err := zipEng.List(context.Background(), a.currentPath)
	if err == nil {
		a.cachedEntries = entries
	}
	return err
}

// RenameEntry performs in-place renaming of an entry in the current ZIP archive.
func (a *App) RenameEntry(oldPath, newPath string) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.currentPath == "" {
		return errors.New("no archive is open")
	}

	zipEng := engine.NewPureZipEngine()
	if err := zipEng.RenameEntry(context.Background(), a.currentPath, oldPath, newPath); err != nil {
		return err
	}

	entries, err := zipEng.List(context.Background(), a.currentPath)
	if err == nil {
		a.cachedEntries = entries
	}
	return err
}

// AppendEntry appends a new file into the currently open ZIP archive.
func (a *App) AppendEntry(sourceFilePath, targetEntryPath string) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.currentPath == "" {
		return errors.New("no archive is open")
	}

	zipEng := engine.NewPureZipEngine()
	if err := zipEng.AppendEntry(context.Background(), a.currentPath, sourceFilePath, targetEntryPath); err != nil {
		return err
	}

	entries, err := zipEng.List(context.Background(), a.currentPath)
	if err == nil {
		a.cachedEntries = entries
	}
	return err
}
