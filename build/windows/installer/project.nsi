Unicode true

####
## Please note: Template replacements don't work in this file. They are provided with default defines like
## mentioned underneath.
## If the keyword is not defined, "wails_tools.nsh" will populate them with the values from ProjectInfo.
## If they are defined here, "wails_tools.nsh" will not touch them. This allows to use this project.nsi manually
## from outside of Wails for debugging and development of the installer.
##
## For development first make a wails nsis build to populate the "wails_tools.nsh":
## > wails build --target windows/amd64 --nsis
## Then you can call makensis on this file with specifying the path to your binary:
## For a AMD64 only installer:
## > makensis -DARG_WAILS_AMD64_BINARY=..\..\bin\app.exe
## For a ARM64 only installer:
## > makensis -DARG_WAILS_ARM64_BINARY=..\..\bin\app.exe
## For a installer with both architectures:
## > makensis -DARG_WAILS_AMD64_BINARY=..\..\bin\app-amd64.exe -DARG_WAILS_ARM64_BINARY=..\..\bin\app-arm64.exe
####
## The following information is taken from the ProjectInfo file, but they can be overwritten here.
####
## !define INFO_PROJECTNAME    "MyProject" # Default "{{.Name}}"
## !define INFO_COMPANYNAME    "MyCompany" # Default "{{.Info.CompanyName}}"
## !define INFO_PRODUCTNAME    "MyProduct" # Default "{{.Info.ProductName}}"
## !define INFO_PRODUCTVERSION "1.0.0"     # Default "{{.Info.ProductVersion}}"
## !define INFO_COPYRIGHT      "Copyright" # Default "{{.Info.Copyright}}"
###
## !define PRODUCT_EXECUTABLE  "Application.exe"      # Default "${INFO_PROJECTNAME}.exe"
## !define UNINST_KEY_NAME     "UninstKeyInRegistry"  # Default "${INFO_COMPANYNAME}${INFO_PRODUCTNAME}"
####
## !define REQUEST_EXECUTION_LEVEL "admin"            # Default "admin"  see also https://nsis.sourceforge.io/Docs/Chapter4.html
####
## Include the wails tools
####
## Sortly: instalação por usuário, sem pedir administrador, como a versão 1.0
## (electron-builder), em %LOCALAPPDATA%\Programs\Sortly.
!define WAILS_INSTALL_SCOPE "user"
!define REQUEST_EXECUTION_LEVEL "user"

!include "wails_tools.nsh"
!include "StrFunc.nsh"
${StrStr}

# The version information for this two must consist of 4 parts
VIProductVersion "${INFO_PRODUCTVERSION}.0"
VIFileVersion    "${INFO_PRODUCTVERSION}.0"

VIAddVersionKey "CompanyName"     "${INFO_COMPANYNAME}"
VIAddVersionKey "FileDescription" "${INFO_PRODUCTNAME} Installer"
VIAddVersionKey "ProductVersion"  "${INFO_PRODUCTVERSION}"
VIAddVersionKey "FileVersion"     "${INFO_PRODUCTVERSION}"
VIAddVersionKey "LegalCopyright"  "${INFO_COPYRIGHT}"
VIAddVersionKey "ProductName"     "${INFO_PRODUCTNAME}"

# Enable HiDPI support. https://nsis.sourceforge.io/Reference/ManifestDPIAware
ManifestDPIAware true

!include "MUI.nsh"

!define MUI_ICON "..\icon.ico"
!define MUI_UNICON "..\icon.ico"
# !define MUI_WELCOMEFINISHPAGE_BITMAP "resources\leftimage.bmp" #Include this to add a bitmap on the left side of the Welcome Page. Must be a size of 164x314
!define MUI_FINISHPAGE_NOAUTOCLOSE # Wait on the INSTFILES page so the user can take a look into the details of the installation steps
!define MUI_ABORTWARNING # This will warn the user if they exit from the installer.
!define MUI_FINISHPAGE_RUN "$INSTDIR\${PRODUCT_EXECUTABLE}" # Sortly: opção de abrir o app ao concluir, como na 1.0

!insertmacro MUI_PAGE_WELCOME # Welcome to the installer page.
# !insertmacro MUI_PAGE_LICENSE "resources\eula.txt" # Adds a EULA page to the installer
!insertmacro MUI_PAGE_DIRECTORY # In which folder install page.
!insertmacro MUI_PAGE_INSTFILES # Installing page.
!insertmacro MUI_PAGE_FINISH # Finished installation page.

!insertmacro MUI_UNPAGE_INSTFILES # Uinstalling page

!insertmacro MUI_LANGUAGE "PortugueseBR" # Sortly: o NSIS usa o idioma do sistema quando disponível
!insertmacro MUI_LANGUAGE "English"

## The following two statements can be used to sign the installer and the uninstaller. The path to the binaries are provided in %1
#!uninstfinalize 'signtool --file "%1"'
#!finalize 'signtool --file "%1"'

Name "${INFO_PRODUCTNAME}"
OutFile "..\..\bin\${INFO_PROJECTNAME}-${ARCH}-installer.exe" # Name of the installer's file.
!ifdef WAILS_INSTALL_SCOPE
  !if "${WAILS_INSTALL_SCOPE}" == "user"
    InstallDir "$LOCALAPPDATA\Programs\${INFO_PRODUCTNAME}"
  !else
    InstallDir "$PROGRAMFILES64\${INFO_COMPANYNAME}\${INFO_PRODUCTNAME}"
  !endif
!else
  InstallDir "$PROGRAMFILES64\${INFO_COMPANYNAME}\${INFO_PRODUCTNAME}"
!endif # Default installing folder ($PROGRAMFILES is Program Files folder).
ShowInstDetails show # This will always show the installation details.

Function .onInit
   !insertmacro wails.checkArchitecture
FunctionEnd

## Sortly: remove a versão 1.0 (Electron), que usa a mesma pasta e o mesmo
## Sortly.exe. O registro da última organização (~/.sortly) é preservado, então
## um desfazer pendente continua disponível depois da atualização.
!define UNINSTALL_ROOT "Software\Microsoft\Windows\CurrentVersion\Uninstall"

Function RemoveSortlyElectron
    StrCpy $0 0
    loop:
        EnumRegKey $1 HKCU "${UNINSTALL_ROOT}" $0
        StrCmp $1 "" done
        IntOp $0 $0 + 1
        ## A 1.0 registra UninstallString = "<pasta>\Uninstall Sortly.exe" /currentuser
        ## e deixa InstallLocation vazio: a pasta é extraída do UninstallString.
        ReadRegStr $2 HKCU "${UNINSTALL_ROOT}\$1" "UninstallString"
        ${StrStr} $3 $2 "\Uninstall Sortly.exe"
        StrCmp $3 "" loop
        StrLen $5 $2
        StrLen $6 $3
        IntOp $5 $5 - $6 ## comprimento até antes de "\Uninstall Sortly.exe"
        StrCpy $7 $2 1
        StrCmp $7 '"' 0 +4
            IntOp $5 $5 - 1
            StrCpy $4 $2 $5 1 ## pula a aspa inicial
            Goto +2
        StrCpy $4 $2 $5
        DetailPrint "Removendo a versão anterior do Sortly: $4"
        ## _?= executa o desinstalador no lugar, sem se copiar para a pasta temporária,
        ## para que o ExecWait realmente espere o fim da remoção.
        ExecWait '"$4\Uninstall Sortly.exe" /S /currentuser _?=$4'
        RMDir /r "$4"
        DeleteRegKey HKCU "${UNINSTALL_ROOT}\$1"
        StrCpy $0 0 ## a lista de chaves mudou: recomeça a busca
        Goto loop
    done:
FunctionEnd

Section
    !insertmacro wails.setShellContext
    Call RemoveSortlyElectron

    !insertmacro wails.webview2runtime

    SetOutPath $INSTDIR

    !insertmacro wails.files

    CreateShortcut "$SMPROGRAMS\${INFO_PRODUCTNAME}.lnk" "$INSTDIR\${PRODUCT_EXECUTABLE}"
    CreateShortCut "$DESKTOP\${INFO_PRODUCTNAME}.lnk" "$INSTDIR\${PRODUCT_EXECUTABLE}"

    !insertmacro wails.associateFiles
    !insertmacro wails.associateCustomProtocols

    !insertmacro wails.writeUninstaller
SectionEnd

Section "uninstall"
    !insertmacro wails.setShellContext

    RMDir /r "$AppData\${PRODUCT_EXECUTABLE}" # Remove the WebView2 DataPath
    RMDir /r "$AppData\${INFO_PRODUCTNAME}\logs" # Sortly: logs do app
    RMDir "$AppData\${INFO_PRODUCTNAME}" # Sortly: só remove se ficar vazia

    RMDir /r $INSTDIR

    Delete "$SMPROGRAMS\${INFO_PRODUCTNAME}.lnk"
    Delete "$DESKTOP\${INFO_PRODUCTNAME}.lnk"

    !insertmacro wails.unassociateFiles
    !insertmacro wails.unassociateCustomProtocols

    !insertmacro wails.deleteUninstaller
SectionEnd
