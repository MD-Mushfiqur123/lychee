@echo off
echo.
echo 🍒  Lychee v0.1.1-alpha — Windows Build
echo ========================================
echo.

setlocal
set "GOOS=windows"
set "GOARCH=amd64"
set "CGO_ENABLED=1"

echo 🔨 Building lychee.exe for windows/amd64...
go build -ldflags="-s -w -X main.version=0.1.1-alpha" -o dist\lychee.exe .
if %ERRORLEVEL% NEQ 0 (
    echo ❌ Build failed!
    exit /b 1
)

echo.
echo ✅ Build complete!
echo 📦 Output: dist\lychee.exe
dir dist\lychee.exe 2>nul
echo.
echo 🚀 Run: dist\lychee.exe serve
endlocal
