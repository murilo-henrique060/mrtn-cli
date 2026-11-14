@echo off
setlocal
echo ============================================================
echo           Instalador do mrtn para Windows
echo ============================================================
echo.

set "SCRIPT_DIR=%~dp0"
set "PS_SCRIPT=%SCRIPT_DIR%install.ps1"

if exist "%PS_SCRIPT%" (
    powershell.exe -NoProfile -ExecutionPolicy Bypass -File "%PS_SCRIPT%"
) else (
    echo Baixando e executando o instalador do PowerShell...
    powershell.exe -NoProfile -ExecutionPolicy Bypass -Command "Invoke-Expression (Invoke-RestMethod 'https://raw.githubusercontent.com/murilo-henrique060/mrtn-cli/main/install.ps1')"
)

if %ERRORLEVEL% NEQ 0 (
    echo.
    echo [ERRO] Ocorreu uma falha durante a instalacao.
    pause
    exit /b %ERRORLEVEL%
)

echo.
echo Pressione qualquer tecla para fechar esta janela...
pause >nul
