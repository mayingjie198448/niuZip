; ====================================================================
; niuZip Windows 64位 原生安装程序脚本 (NSIS 3.0+)
; ====================================================================

!include "MUI2.nsh"
!include "FileFunc.nsh"

; 基础定义
!define PRODUCT_NAME "niuZip"
!define PRODUCT_VERSION "1.0.2"
!define PRODUCT_PUBLISHER "niuZip"
!define PRODUCT_WEB_SITE "https://github.com/mayingjie198448/niuZip"
!define PRODUCT_DIR_REGKEY "Software\niuZip"
!define PRODUCT_UNINST_KEY "Software\Microsoft\Windows\CurrentVersion\Uninstall\${PRODUCT_NAME}"
!define PRODUCT_UNINST_ROOT_KEY "HKLM"

; 设置压缩引擎
SetCompressor lzma

; 版本信息与属性 (降低杀软误报，提供标准文件属性)
VIProductVersion "1.0.2.0"
VIAddVersionKey "ProductName" "${PRODUCT_NAME}"
VIAddVersionKey "Comments" "Cross-platform high-performance modern archive manager"
VIAddVersionKey "CompanyName" "${PRODUCT_PUBLISHER}"
VIAddVersionKey "LegalCopyright" "Copyright © 2026 niuZip. All rights reserved."
VIAddVersionKey "FileDescription" "${PRODUCT_NAME} Setup Installer"
VIAddVersionKey "FileVersion" "${PRODUCT_VERSION}.0"
VIAddVersionKey "ProductVersion" "${PRODUCT_VERSION}.0"

; 界面设置
Name "${PRODUCT_NAME} ${PRODUCT_VERSION}"
OutFile "..\..\build\niuZip-Setup-x64.exe"
InstallDir "$PROGRAMFILES64\niuZip"
InstallDirRegKey HKLM "${PRODUCT_DIR_REGKEY}" ""
ShowInstDetails show
ShowUnInstDetails show
RequestExecutionLevel admin

; 图标设置
!define MUI_ICON "..\icons\appicon.ico"
!define MUI_UNICON "..\icons\appicon.ico"

; 页面设置 (向导)
!insertmacro MUI_PAGE_WELCOME
!insertmacro MUI_PAGE_LICENSE "..\..\LICENSE"
!insertmacro MUI_PAGE_COMPONENTS
!insertmacro MUI_PAGE_DIRECTORY
!insertmacro MUI_PAGE_INSTFILES

; 安装完成页 - 支持直接启动程序
!define MUI_FINISHPAGE_RUN "$INSTDIR\niuZip.exe"
!define MUI_FINISHPAGE_RUN_TEXT "立即运行 ${PRODUCT_NAME}"
!insertmacro MUI_PAGE_FINISH

; 卸载页面设置
!insertmacro MUI_UNPAGE_CONFIRM
!insertmacro MUI_UNPAGE_INSTFILES
!insertmacro MUI_UNPAGE_FINISH

; 语言设置
!insertmacro MUI_LANGUAGE "SimpChinese"

; ====================================================================
; 组件区 (Sections)
; ====================================================================

; 1. 核心程序组件 (必选)
Section "核心程序文件 (必需)" SecCore
  SectionIn RO
  SetOutPath "$INSTDIR"
  SetOverwrite ifnewer

  ; 写入主程序及图标
  File "/oname=niuZip.exe" "..\..\build\bin\niuZip-windows-amd64.exe"
  File "/oname=appicon.ico" "..\icons\appicon.ico"

  ; 写入反安装程序
  WriteUninstaller "$INSTDIR\Uninstall.exe"

  ; 注册应用路径与卸载信息到系统控制面板 (添加/删除程序)
  WriteRegStr HKLM "${PRODUCT_DIR_REGKEY}" "" "$INSTDIR"
  WriteRegStr ${PRODUCT_UNINST_ROOT_KEY} "${PRODUCT_UNINST_KEY}" "DisplayName" "${PRODUCT_NAME}"
  WriteRegStr ${PRODUCT_UNINST_ROOT_KEY} "${PRODUCT_UNINST_KEY}" "UninstallString" '"$INSTDIR\Uninstall.exe"'
  WriteRegStr ${PRODUCT_UNINST_ROOT_KEY} "${PRODUCT_UNINST_KEY}" "DisplayIcon" '"$INSTDIR\appicon.ico"'
  WriteRegStr ${PRODUCT_UNINST_ROOT_KEY} "${PRODUCT_UNINST_KEY}" "DisplayVersion" "${PRODUCT_VERSION}"
  WriteRegStr ${PRODUCT_UNINST_ROOT_KEY} "${PRODUCT_UNINST_KEY}" "Publisher" "${PRODUCT_PUBLISHER}"
  WriteRegStr ${PRODUCT_UNINST_ROOT_KEY} "${PRODUCT_UNINST_KEY}" "URLInfoAbout" "${PRODUCT_WEB_SITE}"

  ${GetSize} "$INSTDIR" "/S=0K" $0 $1 $2
  IntFmt $0 "0x%08X" $0
  WriteRegDWORD ${PRODUCT_UNINST_ROOT_KEY} "${PRODUCT_UNINST_KEY}" "EstimatedSize" "$0"
SectionEnd

; 2. 注册系统鼠标右键菜单 (默认勾选)
Section "添加到鼠标右键快捷菜单 (解压/压缩)" SecContextMenu
  ; 关联主流压缩文件格式 (级联菜单: 用 niuZip 打开 / 解压到当前文件夹 / 解压到同名文件夹)
  !macro RegisterExtContextMenu EXT
    ; 先清理旧版本可能残留的单项关联 (特别是旧版的 command 子键与旧默认值)，确保覆盖安装时级联子菜单干净生效
    DeleteRegKey HKCR "SystemFileAssociations\${EXT}\shell\niuZip"

    WriteRegStr HKCR "SystemFileAssociations\${EXT}\shell\niuZip" "MUIVerb" "niuZip"
    WriteRegStr HKCR "SystemFileAssociations\${EXT}\shell\niuZip" "Icon" '"$INSTDIR\niuZip.exe",0'
    WriteRegStr HKCR "SystemFileAssociations\${EXT}\shell\niuZip" "SubCommands" ""

    WriteRegStr HKCR "SystemFileAssociations\${EXT}\shell\niuZip\Shell\open" "" "用 niuZip 打开"
    WriteRegStr HKCR "SystemFileAssociations\${EXT}\shell\niuZip\Shell\open" "Icon" '"$INSTDIR\niuZip.exe",0'
    WriteRegStr HKCR "SystemFileAssociations\${EXT}\shell\niuZip\Shell\open\command" "" '"$INSTDIR\niuZip.exe" "%1"'

    WriteRegStr HKCR "SystemFileAssociations\${EXT}\shell\niuZip\Shell\extracthere" "" "解压到当前文件夹"
    WriteRegStr HKCR "SystemFileAssociations\${EXT}\shell\niuZip\Shell\extracthere" "Icon" '"$INSTDIR\niuZip.exe",0'
    WriteRegStr HKCR "SystemFileAssociations\${EXT}\shell\niuZip\Shell\extracthere\command" "" '"$INSTDIR\niuZip.exe" --extract-here "%1"'

    WriteRegStr HKCR "SystemFileAssociations\${EXT}\shell\niuZip\Shell\extracttodir" "" "解压到同名文件夹"
    WriteRegStr HKCR "SystemFileAssociations\${EXT}\shell\niuZip\Shell\extracttodir" "Icon" '"$INSTDIR\niuZip.exe",0'
    WriteRegStr HKCR "SystemFileAssociations\${EXT}\shell\niuZip\Shell\extracttodir\command" "" '"$INSTDIR\niuZip.exe" --extract-to-dir "%1"'
  !macroend

  !insertmacro RegisterExtContextMenu ".zip"
  !insertmacro RegisterExtContextMenu ".7z"
  !insertmacro RegisterExtContextMenu ".rar"
  !insertmacro RegisterExtContextMenu ".tar"
  !insertmacro RegisterExtContextMenu ".gz"
  !insertmacro RegisterExtContextMenu ".bz2"
  !insertmacro RegisterExtContextMenu ".xz"
  !insertmacro RegisterExtContextMenu ".iso"
  !insertmacro RegisterExtContextMenu ".cab"
  !insertmacro RegisterExtContextMenu ".apk"
  !insertmacro RegisterExtContextMenu ".jar"

  ; 关联文件夹右键菜单 (立即压缩为 ZIP / 添加到压缩文件...)
  DeleteRegKey HKCR "Directory\shell\niuZip"
  WriteRegStr HKCR "Directory\shell\niuZip" "MUIVerb" "niuZip"
  WriteRegStr HKCR "Directory\shell\niuZip" "Icon" '"$INSTDIR\niuZip.exe",0'
  WriteRegStr HKCR "Directory\shell\niuZip" "SubCommands" ""

  WriteRegStr HKCR "Directory\shell\niuZip\Shell\addtozip" "" "立即压缩为 ZIP"
  WriteRegStr HKCR "Directory\shell\niuZip\Shell\addtozip" "Icon" '"$INSTDIR\niuZip.exe",0'
  WriteRegStr HKCR "Directory\shell\niuZip\Shell\addtozip\command" "" '"$INSTDIR\niuZip.exe" --add-to-zip "%1"'

  WriteRegStr HKCR "Directory\shell\niuZip\Shell\adddialog" "" "添加到压缩文件..."
  WriteRegStr HKCR "Directory\shell\niuZip\Shell\adddialog" "Icon" '"$INSTDIR\niuZip.exe",0'
  WriteRegStr HKCR "Directory\shell\niuZip\Shell\adddialog\command" "" '"$INSTDIR\niuZip.exe" --add-dialog "%1"'

  ; 关联任意文件右键菜单 (立即压缩为 ZIP / 添加到压缩文件...)
  DeleteRegKey HKCR "*\shell\niuZip"
  WriteRegStr HKCR "*\shell\niuZip" "MUIVerb" "niuZip"
  WriteRegStr HKCR "*\shell\niuZip" "Icon" '"$INSTDIR\niuZip.exe",0'
  WriteRegStr HKCR "*\shell\niuZip" "SubCommands" ""

  WriteRegStr HKCR "*\shell\niuZip\Shell\addtozip" "" "立即压缩为 ZIP"
  WriteRegStr HKCR "*\shell\niuZip\Shell\addtozip" "Icon" '"$INSTDIR\niuZip.exe",0'
  WriteRegStr HKCR "*\shell\niuZip\Shell\addtozip\command" "" '"$INSTDIR\niuZip.exe" --add-to-zip "%1"'

  WriteRegStr HKCR "*\shell\niuZip\Shell\adddialog" "" "添加到压缩文件..."
  WriteRegStr HKCR "*\shell\niuZip\Shell\adddialog" "Icon" '"$INSTDIR\niuZip.exe",0'
  WriteRegStr HKCR "*\shell\niuZip\Shell\adddialog\command" "" '"$INSTDIR\niuZip.exe" --add-dialog "%1"'
SectionEnd

; 3. 创建开始菜单快捷方式 (默认勾选)
Section "开始菜单快捷方式" SecStartMenu
  CreateDirectory "$SMPROGRAMS\${PRODUCT_NAME}"
  CreateShortcut "$SMPROGRAMS\${PRODUCT_NAME}\${PRODUCT_NAME}.lnk" "$INSTDIR\niuZip.exe" "" "$INSTDIR\appicon.ico" 0
  CreateShortcut "$SMPROGRAMS\${PRODUCT_NAME}\卸载 ${PRODUCT_NAME}.lnk" "$INSTDIR\Uninstall.exe"
SectionEnd

; 4. 创建桌面快捷方式 (默认勾选)
Section "桌面快捷方式" SecDesktopShortcut
  CreateShortcut "$DESKTOP\${PRODUCT_NAME}.lnk" "$INSTDIR\niuZip.exe" "" "$INSTDIR\appicon.ico" 0
SectionEnd

; 组件描述语言包
!insertmacro MUI_FUNCTION_DESCRIPTION_BEGIN
  !insertmacro MUI_DESCRIPTION_TEXT ${SecCore} "niuZip 主执行程序及必要的核心组件。"
  !insertmacro MUI_DESCRIPTION_TEXT ${SecContextMenu} "在 Windows 资源管理器右键添加快捷菜单 (解压到当前文件夹、解压到同名文件夹、立即压缩为 ZIP、添加到压缩文件...)。"
  !insertmacro MUI_DESCRIPTION_TEXT ${SecStartMenu} "在 Windows 开始菜单中生成应用程序入口及卸载项。"
  !insertmacro MUI_DESCRIPTION_TEXT ${SecDesktopShortcut} "在当前用户桌面上生成 niuZip 启动快捷方式。"
!insertmacro MUI_FUNCTION_DESCRIPTION_END

; ====================================================================
; 卸载程序 (Uninstaller)
; ====================================================================
Section "Uninstall"
  ; 1. 彻底清除注册的鼠标右键菜单
  !macro UnregisterExtContextMenu EXT
    DeleteRegKey HKCR "SystemFileAssociations\${EXT}\shell\niuZip"
  !macroend

  !insertmacro UnregisterExtContextMenu ".zip"
  !insertmacro UnregisterExtContextMenu ".7z"
  !insertmacro UnregisterExtContextMenu ".rar"
  !insertmacro UnregisterExtContextMenu ".tar"
  !insertmacro UnregisterExtContextMenu ".gz"
  !insertmacro UnregisterExtContextMenu ".bz2"
  !insertmacro UnregisterExtContextMenu ".xz"
  !insertmacro UnregisterExtContextMenu ".iso"
  !insertmacro UnregisterExtContextMenu ".cab"
  !insertmacro UnregisterExtContextMenu ".apk"
  !insertmacro UnregisterExtContextMenu ".jar"

  DeleteRegKey HKCR "Directory\shell\niuZip"
  DeleteRegKey HKCR "*\shell\niuZip"

  ; 2. 清除控制面板卸载项与程序目录注册表
  DeleteRegKey ${PRODUCT_UNINST_ROOT_KEY} "${PRODUCT_UNINST_KEY}"
  DeleteRegKey HKLM "${PRODUCT_DIR_REGKEY}"

  ; 3. 删除桌面与开始菜单快捷方式
  Delete "$DESKTOP\${PRODUCT_NAME}.lnk"
  RMDir /r "$SMPROGRAMS\${PRODUCT_NAME}"

  ; 4. 删除安装目录下的文件与目录
  Delete "$INSTDIR\niuZip.exe"
  Delete "$INSTDIR\appicon.ico"
  Delete "$INSTDIR\Uninstall.exe"
  RMDir "$INSTDIR"

  SetAutoClose true
SectionEnd
