# Lychee Windows Install Script
# Run this script in PowerShell to install Lychee on Windows.

$ErrorActionPreference = "Stop"

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "  Lychee Windows Installer" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

# Check if Go is installed
Write-Host "[1/2] Checking for Go installation..." -ForegroundColor Yellow
try {
    $goVersion = go version 2>&1
    Write-Host "  Found: $goVersion" -ForegroundColor Green
} catch {
    Write-Host "  ERROR: Go is not installed or not in PATH." -ForegroundColor Red
    Write-Host "  Please install Go from https://go.dev/dl/ and try again." -ForegroundColor Red
    exit 1
}

# Install Lychee
Write-Host "[2/2] Installing Lychee..." -ForegroundColor Yellow
go install github.com/lychee/lychee@latest

if ($LASTEXITCODE -ne 0) {
    Write-Host "  ERROR: Installation failed." -ForegroundColor Red
    exit 1
}

Write-Host ""
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "  Lychee installed successfully!" -ForegroundColor Green
Write-Host ""
Write-Host "  Run 'lychee serve' to start the server."
Write-Host "  Run 'lychee --help' for more commands."
Write-Host "========================================" -ForegroundColor Cyan
