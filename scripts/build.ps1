param(
    [ValidateSet("windows", "linux", "darwin")]
    [string]$TargetOS = "windows",
    [ValidateSet("amd64", "arm64")]
    [string]$TargetArch = "amd64"
)

$ErrorActionPreference = "Stop"
$repoRoot = Split-Path -Parent $PSScriptRoot
$webDist = Join-Path $repoRoot "apps\web\dist"
$embedDist = Join-Path $repoRoot "apps\server\internal\webapp\dist"
$releaseDir = Join-Path $repoRoot "release"

Push-Location $repoRoot
try {
    npm ci
    npm run build:web

    $resolvedEmbed = [System.IO.Path]::GetFullPath($embedDist)
    $expectedParent = [System.IO.Path]::GetFullPath((Join-Path $repoRoot "apps\server\internal\webapp"))
    if (-not $resolvedEmbed.StartsWith($expectedParent, [System.StringComparison]::OrdinalIgnoreCase)) {
        throw "Unsafe embedded frontend path: $resolvedEmbed"
    }
    if (Test-Path -LiteralPath $embedDist) {
        Remove-Item -LiteralPath $embedDist -Recurse -Force
    }
    Copy-Item -LiteralPath $webDist -Destination $embedDist -Recurse

    New-Item -ItemType Directory -Path $releaseDir -Force | Out-Null
    $extension = if ($TargetOS -eq "windows") { ".exe" } else { "" }
    $output = Join-Path $releaseDir ("ddivination-" + $TargetOS + "-" + $TargetArch + $extension)
    $env:GOOS = $TargetOS
    $env:GOARCH = $TargetArch
    $env:CGO_ENABLED = "0"
    Push-Location (Join-Path $repoRoot "apps\server")
    try {
        go test ./...
        go build -trimpath -ldflags="-s -w" -o $output ./cmd/ddivination
    }
    finally {
        Pop-Location
    }

    $packTarget = Join-Path $releaseDir "assets\base-pack"
    New-Item -ItemType Directory -Path $packTarget -Force | Out-Null
    Copy-Item -Path (Join-Path $repoRoot "assets\base-pack\*") -Destination $packTarget -Recurse -Force
    Write-Host "Built $output"
}
finally {
    $resolvedEmbedForCleanup = [System.IO.Path]::GetFullPath($embedDist)
    $expectedParentForCleanup = [System.IO.Path]::GetFullPath((Join-Path $repoRoot "apps\server\internal\webapp"))
    if ($resolvedEmbedForCleanup.StartsWith($expectedParentForCleanup, [System.StringComparison]::OrdinalIgnoreCase)) {
        if (Test-Path -LiteralPath $embedDist) {
            Remove-Item -LiteralPath $embedDist -Recurse -Force
        }
        New-Item -ItemType Directory -Path $embedDist -Force | Out-Null
        Set-Content -LiteralPath (Join-Path $embedDist "placeholder.txt") -Encoding utf8 -Value "Frontend production files are copied here temporarily by scripts/build.ps1."
    }
    Pop-Location
}
