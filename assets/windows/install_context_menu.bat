@echo off
chcp 65001 >nul
echo ========================================================
echo   正在将 niuZip 添加到 Windows 资源管理器鼠标右键菜单...
echo ========================================================

set "APP_DIR=%~dp0"
set "EXE_PATH=%APP_DIR%niuZip-windows-amd64.exe"

if not exist "%EXE_PATH%" (
    set "EXE_PATH=%APP_DIR%..\..\build\bin\niuZip-windows-amd64.exe"
)

if not exist "%EXE_PATH%" (
    echo [错误] 未找到 niuZip-windows-amd64.exe，请确保本脚本与程序放在同级目录。
    pause
    exit /b 1
)

echo 程序路径: %EXE_PATH%

:: 1. 注册通用文件与常见压缩包右键菜单 (仅需当前用户权限，无需管理员提权)
for %%E in (.zip .7z .rar .tar .gz .bz2 .xz .iso .cab .apk .jar) do (
    reg add "HKCU\Software\Classes\SystemFileAssociations\%%E\shell\niuZip" /ve /t REG_SZ /d "用 niuZip 打开" /f >nul
    reg add "HKCU\Software\Classes\SystemFileAssociations\%%E\shell\niuZip" /v "Icon" /t REG_SZ /d "\"%EXE_PATH%\",0" /f >nul
    reg add "HKCU\Software\Classes\SystemFileAssociations\%%E\shell\niuZip\command" /ve /t REG_SZ /d "\"%EXE_PATH%\" \"%%1\"" /f >nul
)

:: 2. 注册文件夹右键菜单 (浏览/压缩)
reg add "HKCU\Software\Classes\Directory\shell\niuZip" /ve /t REG_SZ /d "用 niuZip 浏览/压缩" /f >nul
reg add "HKCU\Software\Classes\Directory\shell\niuZip" /v "Icon" /t REG_SZ /d "\"%EXE_PATH%\",0" /f >nul
reg add "HKCU\Software\Classes\Directory\shell\niuZip\command" /ve /t REG_SZ /d "\"%EXE_PATH%\" \"%%1\"" /f >nul

echo.
echo ========================================================
echo   [成功] 已成功将「用 niuZip 打开」添加到鼠标右键菜单！
echo   现在右键点击任意 .zip / .7z / .rar 等文件即可快速打开。
echo ========================================================
pause
