//go:build !windows

package main

// RegisterWindowsContextMenu is a no-op on non-Windows platforms.
func (a *App) RegisterWindowsContextMenu() error {
	return nil
}

// UnregisterWindowsContextMenu is a no-op on non-Windows platforms.
func (a *App) UnregisterWindowsContextMenu() error {
	return nil
}
