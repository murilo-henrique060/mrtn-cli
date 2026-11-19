<#
.SYNOPSIS
    Instalador automatizado do mrtn para Windows (PowerShell)
    https://github.com/murilo-henrique060/mrtn-cli
#>

[CmdletBinding()]
param(
    [string]$Version = "v1.1.0",
    [string]$InstallDir = ""
)

$ErrorActionPreference = "Stop"

Write-Host "============================================================" -ForegroundColor Cyan
Write-Host "           Instalador do mrtn ($Version) - Windows           " -ForegroundColor Cyan
Write-Host "============================================================" -ForegroundColor Cyan

# 1. Detect Architecture
$arch = "amd64"
if ([System.Environment]::Is64BitOperatingSystem) {
    if ($env:PROCESSOR_ARCHITECTURE -eq "ARM64") {
        $arch = "arm64"
    } else {
        $arch = "amd64"
    }
}
Write-Host "Detectado: Windows ($arch)" -ForegroundColor Gray

# 2. Define Directories
if (-not $InstallDir) {
    $InstallDir = Join-Path $env:LOCALAPPDATA "Programs\mrtn"
}

if (-not (Test-Path -Path $InstallDir)) {
    New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null
}

$targetExe = Join-Path $InstallDir "mrtn.exe"
$binaryName = "mrtn-$Version-windows-$arch.exe"

# 3. Locate local binary or download
$localBinary = $null
if (-not [string]::IsNullOrEmpty($PSScriptRoot)) {
    $candidate = Join-Path $PSScriptRoot "dist\$Version\$binaryName"
    if (Test-Path -Path $candidate) {
        $localBinary = $candidate
    }
}

if ($localBinary) {
    Write-Host "Usando binário local: $localBinary" -ForegroundColor Yellow
    Copy-Item -Path $localBinary -Destination $targetExe -Force
} else {
    $repo = "murilo-henrique060/mrtn-cli"
    $downloadUrl = "https://github.com/$repo/releases/download/$Version/$binaryName"
    Write-Host "Baixando $binaryName de $downloadUrl..." -ForegroundColor Yellow
    
    $ProgressPreference = 'SilentlyContinue'
    [Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12
    $tempFile = [System.IO.Path]::GetTempFileName()
    try {
        Invoke-WebRequest -Uri $downloadUrl -OutFile $tempFile -UseBasicParsing
        Move-Item -Path $tempFile -Destination $targetExe -Force
    } catch {
        Write-Error "Falha ao baixar o executável: $_"
        if (Test-Path -Path $tempFile) { Remove-Item -Path $tempFile -Force }
        exit 1
    }
}

Write-Host "Executável instalado em: $targetExe" -ForegroundColor Green

# 4. Configure PATH
$userPath = [Environment]::GetEnvironmentVariable("Path", "User")
if ($null -eq $userPath) { $userPath = "" }
$paths = @($userPath -split ';' | Where-Object { -not [string]::IsNullOrWhiteSpace($_) })
if ($paths -notcontains $InstallDir) {
    Write-Host "Adicionando $InstallDir à variável PATH do usuário..." -ForegroundColor Yellow
    $newPath = ($paths + $InstallDir) -join ';'
    [Environment]::SetEnvironmentVariable("Path", $newPath, "User")
    Write-Host "-> PATH atualizado com sucesso!" -ForegroundColor Green
} else {
    Write-Host "-> $InstallDir já está no PATH do usuário." -ForegroundColor Gray
}

# Update current session PATH so it works right away
if ($env:Path -notlike "*$InstallDir*") {
    $env:Path += ";$InstallDir"
}

# 5. Verify Installation
Write-Host ""
Write-Host "Verificação da instalação:" -ForegroundColor Cyan
try {
    & $targetExe --version
} catch {
    Write-Host "Aviso: Não foi possível executar o binário diretamente agora." -ForegroundColor Yellow
}

Write-Host ""
Write-Host "Próximos passos:" -ForegroundColor Cyan
Write-Host "  1. mrtn sync     # Sincroniza os casos de teste oficiais"
Write-Host "  2. mrtn info     # Lista as provas e questões disponíveis"
Write-Host "  3. mrtn run      # Executa e valida sua solução contra o juiz"
Write-Host ""
Write-Host "Instalação concluída com sucesso! 🎈" -ForegroundColor Green
Write-Host "(Nota: Caso o comando 'mrtn' não seja reconhecido em outros terminais abertos, basta reabri-los)." -ForegroundColor Gray
