# smoke-test.ps1
# Lychee Smoke Test — Quick API sanity check
# Usage: .\scripts\smoke-test.ps1 [port]
# Defaults to port 11434

param(
    [int]$Port = 11434,
    [string]$Binary = ".\lychee.exe",
    [string]$Model = "orca-mini:3b"   # tiny model for fast smoke test
)

$ErrorActionPreference = "Stop"
$BaseUrl = "http://localhost:$Port"
$Pass = 0
$Fail = 0

Write-Host "==============================" -ForegroundColor Cyan
Write-Host "  Lychee Smoke Test" -ForegroundColor Cyan
Write-Host "  Binary : $Binary" -ForegroundColor Cyan
Write-Host "  Port   : $Port" -ForegroundColor Cyan
Write-Host "  Model  : $Model" -ForegroundColor Cyan
Write-Host "==============================" -ForegroundColor Cyan
Write-Host ""

# ── 1. Start lychee server ──────────────────────────────────────────
Write-Host "[1/5] Starting lychee serve..." -ForegroundColor Yellow

if (-not (Test-Path $Binary)) {
    Write-Host "  FAIL — binary not found at $Binary" -ForegroundColor Red
    Write-Host "  Run: go build -o lychee.exe ." -ForegroundColor Gray
    exit 1
}

$LycheeProcess = Start-Process -FilePath $Binary `
    -ArgumentList "serve","--port",$Port `
    -PassThru `
    -WindowStyle Hidden

Write-Host "  Started (PID $($LycheeProcess.Id)), waiting for server to be ready..." -ForegroundColor Gray
Start-Sleep -Seconds 3

# ── 2. Warm-up / health check ───────────────────────────────────────
Write-Host "[2/5] Health check..." -ForegroundColor Yellow
try {
    # Use GET /api/tags as a basic health-check endpoint
    $health = Invoke-RestMethod -Uri "$BaseUrl/api/tags" -Method Get -TimeoutSec 10 -ErrorAction Stop
    Write-Host "  PASS — server is responding" -ForegroundColor Green
    $Pass++
} catch {
    Write-Host "  FAIL — server not reachable at $BaseUrl" -ForegroundColor Red
    Write-Host "  Error: $_" -ForegroundColor Red
    $Fail++

    Write-Host "  Stopping lychee..." -ForegroundColor Gray
    Stop-Process -Id $LycheeProcess.Id -Force -ErrorAction SilentlyContinue
    exit 1
}

# ── 3. Test GET /api/tags ───────────────────────────────────────────
Write-Host "[3/5] Testing GET /api/tags..." -ForegroundColor Yellow
try {
    $tags = Invoke-RestMethod -Uri "$BaseUrl/api/tags" -Method Get -TimeoutSec 10 -ErrorAction Stop

    # Handle both array and object-with-models-key responses
    if ($tags -is [array]) {
        $modelCount = $tags.Count
    } elseif ($tags.models) {
        $modelCount = $tags.models.Count
    } else {
        $modelCount = 0
    }

    Write-Host "  PASS — /api/tags returned $modelCount model(s)" -ForegroundColor Green
    $Pass++
} catch {
    Write-Host "  FAIL — /api/tags: $_" -ForegroundColor Red
    $Fail++
}

# ── 4. Test POST /api/generate ──────────────────────────────────────
Write-Host "[4/5] Testing POST /api/generate..." -ForegroundColor Yellow

# First pull the model if we have a model to test with
$pullSuccess = $false
try {
    Write-Host "  Pulling model '$Model' (may take a minute)..." -ForegroundColor Gray
    $pullBody = @{ name = $Model; stream = $false } | ConvertTo-Json
    $pullResult = Invoke-RestMethod -Uri "$BaseUrl/api/pull" -Method Post `
        -Body $pullBody -ContentType "application/json" -TimeoutSec 120 -ErrorAction Stop
    Write-Host "  Model pulled OK" -ForegroundColor Gray
    $pullSuccess = $true
} catch {
    Write-Host "  WARNING — could not pull model '$Model': $_" -ForegroundColor DarkYellow
    Write-Host "  Skipping generate test (no model available)" -ForegroundColor DarkYellow
}

if ($pullSuccess) {
    try {
        $genBody = @{
            model  = $Model
            prompt = "Say hello in one word"
            stream = $false
        } | ConvertTo-Json

        $genResult = Invoke-RestMethod -Uri "$BaseUrl/api/generate" -Method Post `
            -Body $genBody -ContentType "application/json" -TimeoutSec 120 -ErrorAction Stop

        Write-Host "  PASS — /api/generate returned: $($genResult.response)" -ForegroundColor Green
        $Pass++
    } catch {
        Write-Host "  FAIL — /api/generate: $_" -ForegroundColor Red
        $Fail++
    }
} else {
    Write-Host "  SKIP — no model to test with" -ForegroundColor DarkYellow
}

# ── 5. Tear down ────────────────────────────────────────────────────
Write-Host "[5/5] Tearing down..." -ForegroundColor Yellow
Stop-Process -Id $LycheeProcess.Id -Force -ErrorAction SilentlyContinue
Write-Host "  Lychee process stopped" -ForegroundColor Gray

# ── Summary ─────────────────────────────────────────────────────────
Write-Host ""
Write-Host "==============================" -ForegroundColor Cyan
Write-Host "  SMOKE TEST RESULTS" -ForegroundColor Cyan
Write-Host "  Passed : $Pass" -ForegroundColor Green
Write-Host "  Failed : $Fail" -ForegroundColor $(if ($Fail -gt 0) { "Red" } else { "Gray" })
Write-Host "==============================" -ForegroundColor Cyan

if ($Fail -gt 0) {
    exit 1
} else {
    Write-Host "  All checks passed!" -ForegroundColor Green
    exit 0
}
