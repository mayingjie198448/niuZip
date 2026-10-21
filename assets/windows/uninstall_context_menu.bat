@echo off
chcp 65001 >nul
echo ========================================================
echo   正在从 Windows 鼠标右键菜单中移除 niuZip...
echo ========================================================

for %%E in (.zip .7z .rar .tar .gz .bz2 .xz .iso .cab .apk .jar) do (
    reg delete "HKCU\Software\Classes\SystemFileAssociations\%%E\shell\niuZip" /f >nul 2>&1
)

reg delete "HKCU\Software\Classes\Directory\shell\niuZip" /f >nul 2>&1

echo.
echo ========================================================
echo   [成功] 已从鼠标右键菜单中移除 niuZip。
echo ========================================================
pause
