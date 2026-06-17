@echo off
REM ──────────────────────────────────────────────
REM 🍒 Lychee — Windows Installer (Batch wrapper)
REM Double-click this file to install Lychee.
REM ──────────────────────────────────────────────

echo.
echo ============================================
echo   Lychee Installer
echo ============================================
echo.

powershell -ExecutionPolicy Bypass -File "%~dp0install.ps1"

if %ERRORLEVEL% NEQ 0 (
    echo.
    echo Installation failed with error code %ERRORLEVEL%.
    pause
    exit /b %ERRORLEVEL%
)

echo.
echo ============================================
echo   Lychee installed successfully!
echo   Run 'lychee serve' to start.
echo ============================================
pause
