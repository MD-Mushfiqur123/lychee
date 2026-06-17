#!/usr/bin/env pwsh
# Lychee API Smoke Test (PowerShell)
# Tests all major endpoints against a running Lychee server

param(
    [string]$HostUrl = $env:LYCHEE_HOST ?? "http://localhost:11434"
)

$Pass = 0
$Fail = 0

function Check {
    param(
        [string]$Name,
        [string]$Method,
        [string]$Url,
        [string]$Data,
        [int]$Expected = 200
    )

    $uri = "$HostUrl$Url"
    $params = @{
        Uri         = $uri
        Method      = $Method
        ContentType = "application/json"
        ErrorAction = "SilentlyContinue"
        SkipHttpErrorCheck = $true
    }

    if ($Data) {
        $params.Body = $Data
    }

    try {
        $response = Invoke-RestMethod @params -StatusCodeVariable "statusCode"
        $actual = $statusCode
    } catch {
        $actual = $_.Exception.Response.StatusCode.value__
    }

    if ($actual -eq $Expected) {
        Write-Host "✅ $Name"
        $script:Pass++
    } else {
        Write-Host "❌ $Name (got $actual, expected $Expected)"
        $script:Fail++
    }
}

Write-Host "Lychee Smoke Test ($HostUrl)"
Write-Host "=========================="

Check "Health"         "GET" "/"
Check "Version"        "GET" "/api/version"
Check "List models"    "GET" "/api/tags"
Check "Running models" "GET" "/api/ps"

Write-Host ""
Write-Host "Results: $Pass passed, $Fail failed"

exit ($Fail -gt 0 ? 1 : 0)
