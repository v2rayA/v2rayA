;--------------------------------
;Include Dependencies

  !include "MUI2.nsh"
  !include "FileFunc.nsh"

;--------------------------------
;General

  Name "v2rayA"
  OutFile "installer_windows_@ARCH@_@VERSION@.exe"

  SetCompressor /solid lzma
  ;SetCompress off ;Uncomment for development

  InstallDir "$PROGRAMFILES64\v2rayA"
  !define UninstId "v2rayA"
  
  ;Get installation folder from registry if available
  InstallDirRegKey HKCU "Software\${UninstId}" ""

  ;Request application privileges for Windows Vista
  RequestExecutionLevel admin

;--------------------------------
;Interface Settings

  !define MUI_ABORTWARNING

  !define MUI_ICON "v2rayA\v2raya.ico"
  !define MUI_HEADERIMAGE
  !define MUI_HEADERIMAGE_BITMAP "v2rayA\v2raya.bmp"
  !define MUI_HEADERIMAGE_RIGHT
  !define MUI_HEADERIMAGE_BITMAP_STRETCH "AspectFitHeight"

;--------------------------------
;Pages

  !insertmacro MUI_PAGE_LICENSE "License.txt"
  !insertmacro MUI_PAGE_DIRECTORY
  !insertmacro MUI_PAGE_INSTFILES
  
  !insertmacro MUI_UNPAGE_CONFIRM
  !insertmacro MUI_UNPAGE_INSTFILES
  
;--------------------------------
;Languages
 
  !insertmacro MUI_LANGUAGE "English"
  !insertmacro MUI_LANGUAGE "SimpChinese"
  !insertmacro MUI_LANGUAGE "TradChinese"

;--------------------------------
;Installer Sections


Section "Install Section" SecInst

  SetOutPath "$INSTDIR"

  ExecWait "$INSTDIR\Uninstall.exe /S _?=$INSTDIR"
  
  File "v2rayA\v2raya.exe"                    
  File "v2rayA\v2raya.ico"
  File "v2rayA\v2raya.xml"
  File "v2rayA\v2raya_windows_@ARCH@_@VERSION@.exe"
  
  CreateDirectory "$INSTDIR\v2ray-core"
  SetOutPath "$INSTDIR\v2ray-core"
  File "v2ray-core\*"
  
  ;Store installation folder
  WriteRegStr HKCU "Software\v2rayA" "" $INSTDIR
  
  ;Create uninstaller
  WriteUninstaller "$INSTDIR\Uninstall.exe"

  ;Manage service
  ExecWait '"$INSTDIR\v2raya.exe" "install"'
  ExecWait '"$INSTDIR\v2raya.exe" "start"'

  ;Create entry in Control Panel
  WriteRegStr HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\${UninstId}" "DisplayName" "v2rayA"
  WriteRegStr HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\${UninstId}" "UninstallString" "$\"$INSTDIR\Uninstall.exe$\""
  WriteRegStr HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\${UninstId}" "DisplayIcon" "$\"$INSTDIR\v2raya.ico$\""
  WriteRegStr HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\${UninstId}" "Publisher" "The v2rayA developer community"
  WriteRegStr HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\${UninstId}" "URLInfoAbout" "https://github.com/v2rayA/v2rayA"
  WriteRegDWORD HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\${UninstId}" "NoModify" 1
  WriteRegDWORD HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\${UninstId}" "NoRepair" 1
  WriteRegStr HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\${UninstId}" "DisplayVersion" "@VERSION@"

  ;Create shortcuts
  !macro CreateInternetShortcutWithIcon FILEPATH URL ICONPATH ICONINDEX
  	WriteINIStr "${FILEPATH}" "InternetShortcut" "URL" "${URL}"
  	WriteINIStr "${FILEPATH}" "InternetShortcut" "IconIndex" "${ICONINDEX}"
  	WriteINIStr "${FILEPATH}" "InternetShortcut" "IconFile" "${ICONPATH}"
  !macroend

  !insertmacro CreateInternetShortcutWithIcon "$DESKTOP\v2rayA.url" "http://localhost:2017" "$INSTDIR/v2raya.ico" 0
  CreateDirectory "$SMPROGRAMS\v2rayA"
  !insertmacro CreateInternetShortcutWithIcon "$SMPROGRAMS\v2rayA\v2rayA.url" "http://localhost:2017" "$INSTDIR/v2raya.ico" 0
  createShortCut "$SMPROGRAMS\v2rayA\Uninstall.lnk" "$INSTDIR\Uninstall.exe" "" ""

  ;Calculate size
  ${GetSize} "$INSTDIR" "/S=0K" $0 $1 $2
  IntFmt $0 "0x%08X" $0
  WriteRegDWORD HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\${UninstId}" "EstimatedSize" "$0"

SectionEnd

;--------------------------------
;Uninstaller Section

Section "un.Uninstall"

  Delete "$DESKTOP\v2rayA.url"
  Delete "$SMPROGRAMS\v2rayA\v2rayA.url"

  ;Manage service
  ExecWait '"$INSTDIR\v2raya.exe" "stop"'
  ExecWait '"$INSTDIR\v2raya.exe" "uninstall"'

  Delete "$INSTDIR\Uninstall.exe"
  Delete "$INSTDIR\v2raya.exe"
  Delete "$INSTDIR\v2raya.ico"
  Delete "$INSTDIR\v2raya.xml"
  Delete "$INSTDIR\v2raya_windows_@ARCH@_@VERSION@.exe"
  
  Delete "$INSTDIR\v2ray-core\*"
  RMDir "$INSTDIR\v2ray-core"

  RMDir "$INSTDIR"
  RMDir "$SMPROGRAMS\v2rayA"

  DeleteRegKey HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\${UninstId}"
  DeleteRegKey /ifempty HKCU "Software\v2rayA"

SectionEnd
