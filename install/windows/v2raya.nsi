;--------------------------------
;Include Dependencies

  !include "MUI2.nsh"
  !include "FileFunc.nsh"
  !include "LogicLib.nsh"

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
;Uninstall

!macro UninstallExisting exitcode uninstcommand
Push `${uninstcommand}`
Call UninstallExisting
Pop ${exitcode}
!macroend
Function UninstallExisting
Exch $1 ; uninstcommand
Push $2 ; Uninstaller
Push $3 ; Len
StrCpy $3 ""
StrCpy $2 $1 1
StrCmp $2 '"' qloop sloop
sloop:
	StrCpy $2 $1 1 $3
	IntOp $3 $3 + 1
	StrCmp $2 "" +2
	StrCmp $2 ' ' 0 sloop
	IntOp $3 $3 - 1
	Goto run
qloop:
	StrCmp $3 "" 0 +2
	StrCpy $1 $1 "" 1 ; Remove initial quote
	IntOp $3 $3 + 1
	StrCpy $2 $1 1 $3
	StrCmp $2 "" +2
	StrCmp $2 '"' 0 qloop
run:
	StrCpy $2 $1 $3 ; Path to uninstaller
	StrCpy $1 161 ; ERROR_BAD_PATHNAME
	GetFullPathName $3 "$2\.." ; $InstDir
	IfFileExists "$2" 0 +4
	ExecWait '"$2" /S _?=$3' $1 ; This assumes the existing uninstaller is a NSIS uninstaller, other uninstallers don't support /S nor _?=
	IntCmp $1 0 "" +2 +2 ; Don't delete the installer if it was aborted
	Delete "$2" ; Delete the uninstaller
	RMDir "$3" ; Try to delete $InstDir
	RMDir "$3\.." ; (Optional) Try to delete the parent of $InstDir
Pop $3
Pop $2
Exch $1 ; exitcode
FunctionEnd

;--------------------------------
;Installer Sections

Function .onInit
ReadRegStr $0 HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\${UninstId}" "UninstallString"
${If} $0 != ""
${AndIf} ${Cmd} `MessageBox MB_YESNO|MB_ICONQUESTION "Uninstall previous version?" /SD IDYES IDYES`
	!insertmacro UninstallExisting $0 $0
	${If} $0 <> 0
		MessageBox MB_YESNO|MB_ICONSTOP "Failed to uninstall, continue anyway?" /SD IDYES IDYES +2
			Abort
	${EndIf}
${EndIf}
FunctionEnd

Section "Install Section" SecInst

  SetOutPath "$INSTDIR"
  
  File "v2rayA\v2raya.exe"                    
  File "v2rayA\v2raya.ico"
  File "v2rayA\v2raya.xml"
  File "v2rayA\v2raya_windows_@ARCH@_@VERSION@.exe"
  File "v2rayA\*.dat"

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
  Delete "$INSTDIR\*.dat"

  Delete "$INSTDIR\v2ray-core\*"
  RMDir "$INSTDIR\v2ray-core"

  RMDir "$INSTDIR"
  RMDir "$SMPROGRAMS\v2rayA"

  DeleteRegKey HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\${UninstId}"
  DeleteRegKey /ifempty HKCU "Software\v2rayA"

SectionEnd
