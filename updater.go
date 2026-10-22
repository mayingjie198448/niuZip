package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

const (
	CurrentAppVersion = "1.0.2"
	GitHubRepoOwner   = "mayingjie198448"
	GitHubRepoName    = "niuZip"
)

// GitHubRelease represents GitHub Releases API response.
type GitHubRelease struct {
	TagName     string        `json:"tag_name"`
	Name        string        `json:"name"`
	Body        string        `json:"body"`
	PublishedAt string        `json:"published_at"`
	HTMLURL     string        `json:"html_url"`
	Assets      []GitHubAsset `json:"assets"`
}

type GitHubAsset struct {
	Name               string `json:"name"`
	Size               int64  `json:"size"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

// UpdateInfo is returned to the frontend.
type UpdateInfo struct {
	HasUpdate      bool   `json:"has_update"`
	CurrentVersion string `json:"current_version"`
	LatestVersion  string `json:"latest_version"`
	ReleaseTitle   string `json:"release_title"`
	ReleaseNotes   string `json:"release_notes"`
	DownloadURL    string `json:"download_url"`
	AssetSize      int64  `json:"asset_size"`
	AssetName      string `json:"asset_name"`
	ReleaseURL     string `json:"release_url"`
}

// GetAppVersion returns the current application version.
func (a *App) GetAppVersion() string {
	return CurrentAppVersion
}

func getGitHubToken() string {
	if t := os.Getenv("GITHUB_TOKEN"); t != "" {
		return t
	}
	return os.Getenv("GH_TOKEN")
}

// CheckForUpdate queries GitHub API for the latest release, with automatic web-fallback.
func (a *App) CheckForUpdate() (*UpdateInfo, error) {
	// 1. Try standard GitHub REST API
	info, err := a.checkViaGitHubAPI()
	if err == nil {
		return info, nil
	}

	// 2. If API fails (e.g. unauthenticated Rate Limit 403, network blocked), fallback to Web Redirect check
	fallbackInfo, fallbackErr := a.checkViaWebRedirect()
	if fallbackErr == nil {
		return fallbackInfo, nil
	}

	return nil, fmt.Errorf("检查更新失败: %w", err)
}

func (a *App) checkViaGitHubAPI() (*UpdateInfo, error) {
	apiURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", GitHubRepoOwner, GitHubRepoName)

	client := &http.Client{Timeout: 8 * time.Second}
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "niuZip-App")
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	if token := getGitHubToken(); token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("无法连接 GitHub API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return &UpdateInfo{
			HasUpdate:      false,
			CurrentVersion: CurrentAppVersion,
			LatestVersion:  CurrentAppVersion,
		}, nil
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API 返回状态码 %d", resp.StatusCode)
	}

	var release GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, fmt.Errorf("解析更新信息失败: %w", err)
	}

	latestVer := strings.TrimPrefix(release.TagName, "v")
	hasUpdate := isNewerVersion(CurrentAppVersion, latestVer)

	selectedAsset := findMatchingAsset(release.Assets)
	var downloadURL, assetName string
	var assetSize int64

	if selectedAsset != nil {
		downloadURL = selectedAsset.BrowserDownloadURL
		assetName = selectedAsset.Name
		assetSize = selectedAsset.Size
	} else if len(release.Assets) > 0 {
		downloadURL = release.Assets[0].BrowserDownloadURL
		assetName = release.Assets[0].Name
		assetSize = release.Assets[0].Size
	}

	return &UpdateInfo{
		HasUpdate:      hasUpdate,
		CurrentVersion: CurrentAppVersion,
		LatestVersion:  latestVer,
		ReleaseTitle:   release.Name,
		ReleaseNotes:   release.Body,
		DownloadURL:    downloadURL,
		AssetName:      assetName,
		AssetSize:      assetSize,
		ReleaseURL:     release.HTMLURL,
	}, nil
}

func (a *App) checkViaWebRedirect() (*UpdateInfo, error) {
	webURL := fmt.Sprintf("https://github.com/%s/%s/releases/latest", GitHubRepoOwner, GitHubRepoName)

	client := &http.Client{
		Timeout: 8 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse // Do not follow redirect
		},
	}

	req, err := http.NewRequest("HEAD", webURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "niuZip-App")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	loc := resp.Header.Get("Location")
	if loc == "" {
		// Fallback to GET
		getReq, _ := http.NewRequest("GET", webURL, nil)
		getReq.Header.Set("User-Agent", "niuZip-App")
		getResp, getErr := client.Do(getReq)
		if getErr != nil {
			return nil, getErr
		}
		defer getResp.Body.Close()
		loc = getResp.Header.Get("Location")
	}

	if loc == "" {
		return nil, errors.New("无法解析 GitHub 最新发布重定向地址")
	}

	// Example loc: https://github.com/mayingjie198448/niuZip/releases/tag/v1.0.1
	parts := strings.Split(loc, "/tag/")
	if len(parts) < 2 {
		return nil, fmt.Errorf("无法从重定向地址解析版本: %s", loc)
	}

	rawTag := parts[1]
	latestVer := strings.TrimPrefix(rawTag, "v")
	hasUpdate := isNewerVersion(CurrentAppVersion, latestVer)

	var assetName string
	if runtime.GOOS == "windows" {
		assetName = "niuZip-Setup-x64.exe"
	} else if runtime.GOOS == "darwin" {
		assetName = "niuZip-arm64.dmg"
	} else {
		assetName = "niuZip-darwin-arm64"
	}

	downloadURL := fmt.Sprintf("https://github.com/%s/%s/releases/download/%s/%s", GitHubRepoOwner, GitHubRepoName, rawTag, assetName)

	return &UpdateInfo{
		HasUpdate:      hasUpdate,
		CurrentVersion: CurrentAppVersion,
		LatestVersion:  latestVer,
		ReleaseTitle:   fmt.Sprintf("niuZip %s", rawTag),
		ReleaseNotes:   "包含关键缺陷修复、杀毒软件白名单元数据规范化及性能优化。",
		DownloadURL:    downloadURL,
		AssetName:      assetName,
		ReleaseURL:     loc,
	}, nil
}

// DownloadAndInstallUpdate downloads the new installer and runs it.
func (a *App) DownloadAndInstallUpdate(downloadURL string) error {
	if downloadURL == "" {
		return errors.New("下载链接为空")
	}

	client := &http.Client{Timeout: 15 * time.Minute}
	req, err := http.NewRequest("GET", downloadURL, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "niuZip-App")
	if token := getGitHubToken(); token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("下载失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("下载服务器返回状态: %s", resp.Status)
	}

	totalSize := resp.ContentLength
	fileName := filepath.Base(downloadURL)
	if fileName == "" || fileName == "/" || fileName == "." {
		if runtime.GOOS == "windows" {
			fileName = "niuZip-Setup-x64.exe"
		} else {
			fileName = "niuZip-arm64.dmg"
		}
	}

	tmpDir := os.TempDir()
	savePath := filepath.Join(tmpDir, fileName)

	outFile, err := os.Create(savePath)
	if err != nil {
		return fmt.Errorf("创建临时更新文件失败: %w", err)
	}
	defer outFile.Close()

	buf := make([]byte, 64*1024)
	var downloaded int64
	lastEmitTime := time.Now()

	for {
		n, readErr := resp.Body.Read(buf)
		if n > 0 {
			if _, writeErr := outFile.Write(buf[:n]); writeErr != nil {
				return writeErr
			}
			downloaded += int64(n)

			if time.Since(lastEmitTime) > 100*time.Millisecond || downloaded == totalSize {
				lastEmitTime = time.Now()
				percent := float64(0)
				if totalSize > 0 {
					percent = float64(downloaded) / float64(totalSize) * 100.0
				}
				if a.ctx != nil {
					wailsRuntime.EventsEmit(a.ctx, "update-download-progress", map[string]interface{}{
						"percent":    percent,
						"downloaded": downloaded,
						"total":      totalSize,
					})
				}
			}
		}

		if readErr != nil {
			if readErr == io.EOF {
				break
			}
			return readErr
		}
	}

	outFile.Close()

	// Launch the installer
	switch runtime.GOOS {
	case "windows":
		// Run setup installer
		cmd := exec.Command("cmd", "/c", "start", "", savePath)
		if err := cmd.Start(); err != nil {
			return fmt.Errorf("启动 Windows 安装包失败: %w", err)
		}
		// Exit after launching installer
		go func() {
			time.Sleep(1 * time.Second)
			os.Exit(0)
		}()
		return nil

	case "darwin":
		// Open the DMG file in Finder
		cmd := exec.Command("open", savePath)
		return cmd.Start()

	default:
		cmd := exec.Command("xdg-open", savePath)
		return cmd.Start()
	}
}

func findMatchingAsset(assets []GitHubAsset) *GitHubAsset {
	for _, a := range assets {
		lower := strings.ToLower(a.Name)
		if runtime.GOOS == "windows" {
			if strings.HasSuffix(lower, "setup-x64.exe") || strings.HasSuffix(lower, "setup.exe") {
				return &a
			}
		} else if runtime.GOOS == "darwin" {
			if strings.HasSuffix(lower, ".dmg") {
				return &a
			}
		}
	}

	for _, a := range assets {
		lower := strings.ToLower(a.Name)
		if runtime.GOOS == "windows" && strings.HasSuffix(lower, ".exe") {
			return &a
		}
		if runtime.GOOS == "darwin" && strings.HasSuffix(lower, ".dmg") {
			return &a
		}
	}
	return nil
}

// isNewerVersion returns true if remoteVer > currentVer
func isNewerVersion(currentVer, remoteVer string) bool {
	cParts := parseVersionParts(currentVer)
	rParts := parseVersionParts(remoteVer)

	for i := 0; i < 3; i++ {
		if rParts[i] > cParts[i] {
			return true
		}
		if rParts[i] < cParts[i] {
			return false
		}
	}
	return false
}

func parseVersionParts(ver string) [3]int {
	clean := strings.TrimPrefix(strings.TrimSpace(ver), "v")
	parts := strings.Split(clean, ".")
	var res [3]int
	for i := 0; i < len(parts) && i < 3; i++ {
		// Strip any non-digit suffix like -beta
		numStr := ""
		for _, r := range parts[i] {
			if r >= '0' && r <= '9' {
				numStr += string(r)
			} else {
				break
			}
		}
		if n, err := strconv.Atoi(numStr); err == nil {
			res[i] = n
		}
	}
	return res
}
