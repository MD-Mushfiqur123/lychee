# ──────────────────────────────────────────────
# 🍒 Lychee — Windows PowerShell Installer
# Installs the Lychee CLI binary on Windows.
#
# Usage (one-liner):
#   irm https://raw.githubusercontent.com/MD-Mushfiqur123/lychee/main/scripts/install.ps1 | iex
#
# Or specify a version:
#   $env:LYCHEE_VERSION = "v0.1.0"; irm https://raw.../install.ps1 | iex
# ──────────────────────────────────────────────

$ErrorActionPreference = "Stop"

Write-Host ""
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "  Lychee Windows Installer" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

$Repo = "MD-Mushfiqur123/lychee"
$Arch = if ([Environment]::Is64BitOperatingSystem) { "amd64" } else { err "32-bit Windows is not supported." }
$GoModulePath = "github.com/MD-Mushfiqur123/lychee"

# ── Version ────────────────────────────────────
$Version = if ($env:LYCHEE_VERSION) { $env:LYCHEE_VERSION } else { "latest" }

# ── Helper functions ───────────────────────────
function Log    { Write-Host ">>> $args" -ForegroundColor Gray }
function Success { Write-Host "✔ $args" -ForegroundColor Green }
function Warn   { Write-Host "⚠ $args" -ForegroundColor Yellow }
function Err    { Write-Host "✘ $args" -ForegroundColor Red; exit 1 }

# ── Check for Go 1.22+ first ──────────────────
$goInstalled = $false
try {
    $goOutput = go version 2>&1
    if ($goOutput -match 'go(\d+)\.(\d+)') {
        $major = [int]$Matches[1]
        $minor = [int]$Matches[2]
        if (($major -gt 1) -or ($major -eq 1 -and $minor -ge 22)) {
            Log "Go $($major).$($minor) detected — installing via 'go install'..."
            $goInstalled = $true
        } else {
            Warn "Go $($major).$($minor) detected but 1.22+ required — will download binary instead."
        }
    }
} catch {
    Warn "Go not detected — will download pre-built binary."
}

if ($goInstalled) {
    # Install via Go
    $versionFlag = "@latest"
    if ($Version -ne "latest") { $versionFlag = "@$Version" }

    go install "$GoModulePath$versionFlag"

    if ($LASTEXITCODE -eq 0) {
        $gopath = go env GOPATH 2>$null
        if (-not $gopath) { $gopath = "$env:USERPROFILE\go" }
        $lycheeBin = Join-Path $gopath "bin\lychee.exe"

        if (Test-Path $lycheeBin) {
            # Add GOPATH/bin to PATH if needed
            $gopathBin = Join-Path $gopath "bin"
            $userPath = [Environment]::GetEnvironmentVariable("PATH", "User")
            if ($userPath -notlike "*$gopathBin*") {
                [Environment]::SetEnvironmentVariable("PATH", "$userPath;$gopathBin", "User")
                $env:PATH = "$env:PATH;$gopathBin"
                Log "Added $gopathBin to user PATH."
            }

            Success "Lychee installed via Go!"
            try { lychee version 2>$null } catch {}
            Write-Host ""
            Log "Run 'lychee serve' to start the server."
            Log "Run 'lychee --help' for all commands."
            exit 0
        } else {
            Warn "go install completed but binary not found — falling back to binary download..."
        }
    } else {
        Warn "go install failed (exit code $LASTEXITCODE) — falling back to binary download..."
    }
}

# ── Install via pre-built binary ───────────────
Log "Downloading pre-built binary..."

$BinaryName = "lychee-windows-$Arch.exe"
$downloadUrl = if ($Version -eq "latest") {
    "https://github.com/$Repo/releases/latest/download/$BinaryName"
} else {
    "https://github.com/$Repo/releases/download/$Version/$BinaryName"
}

$installDir = if ($env:LYCHEE_INSTALL_DIR) { $env:LYCHEE_INSTALL_DIR } else { "$env:LOCALAPPDATA\Programs\lychee" }
if (-not (Test-Path $installDir)) {
    New-Item -ItemType Directory -Path $installDir -Force | Out-Null
}

$destPath = Join-Path $installDir "lychee.exe"
$tempPath = Join-Path $env:TEMP "lychee-$([System.Guid]::NewGuid()).exe"

try {
    $ProgressPreference = "SilentlyContinue"
    Invoke-WebRequest -Uri $downloadUrl -OutFile $tempPath -ErrorAction Stop

    if (-not (Test-Path $tempPath) -or (Get-Item $tempPath).Length -eq 0) {
        Err "Download failed — no binary at $downloadUrl`n       Try installing via Go:  go install $GoModulePath@latest"
    }

    Move-Item -Force $tempPath $destPath
    Success "Lychee downloaded to $destPath"
} catch {
    Err "Download failed: $_`n       Try installing via Go:  go install $GoModulePath@latest"
}

# ── Add to PATH ────────────────────────────────
$userPath = [Environment]::GetEnvironmentVariable("PATH", "User")
if ($userPath -notlike "*$installDir*") {
    [Environment]::SetEnvironmentVariable("PATH", "$userPath;$installDir", "User")
    $env:PATH = "$env:PATH;$installDir"
    Log "Added $installDir to user PATH."
}

# ── Verify ─────────────────────────────────────
try {
    $lycheeVersion = & $destPath version 2>$null
    Success "Lychee installed successfully! $lycheeVersion"
} catch {
    Warn "Lychee installed to $installDir but may not be accessible yet."
    Warn "Open a new terminal and run 'lychee --help' to verify."
}

Write-Host ""
Log "Run 'lychee serve' to start the server."
Log "Run 'lychee --help' for all commands."

exit 0
