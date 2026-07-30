param(
    [switch]$SkipE2E,
    [switch]$SkipInstall
)

$ErrorActionPreference = "Stop"
. (Join-Path $PSScriptRoot "common.ps1")

$repoRoot = Get-DDivinationRepoRoot
$testOutputDir = Join-Path $repoRoot ".tmp\test-bin"
$testBinary = Join-Path $testOutputDir "ddivination-test.exe"
$previousPath = $env:PATH

function Invoke-DDivinationStep {
    param(
        [Parameter(Mandatory = $true)]
        [string]$Name,
        [Parameter(Mandatory = $true)]
        [scriptblock]$Action
    )

    Write-Host ""
    Write-Host "==> $Name" -ForegroundColor Cyan
    & $Action
    if ($LASTEXITCODE -ne 0) {
        throw "'$Name' falhou com codigo $LASTEXITCODE."
    }
}

Push-Location $repoRoot
try {
    $goExe = Resolve-DDivinationGo
    $nodeTools = Resolve-DDivinationNodeTools
    $goDir = Split-Path -Parent $goExe
    $env:PATH = $goDir + [System.IO.Path]::PathSeparator + $previousPath

    if (-not (Test-Path -LiteralPath (Join-Path $repoRoot "node_modules"))) {
        if ($SkipInstall) {
            throw "node_modules nao existe e -SkipInstall foi informado. Execute npm ci."
        }
        Invoke-DDivinationStep "Instalar dependencias" {
            & $nodeTools.Npm ci
        }
    }

    New-Item -ItemType Directory -Path $testOutputDir -Force | Out-Null

    Invoke-DDivinationStep "Testes Go" {
        & $goExe -C (Join-Path $repoRoot "apps\server") test ./...
    }
    Invoke-DDivinationStep "Go vet" {
        & $goExe -C (Join-Path $repoRoot "apps\server") vet ./...
    }
    Invoke-DDivinationStep "Build Go" {
        & $goExe -C (Join-Path $repoRoot "apps\server") build -o $testBinary ./cmd/ddivination
    }
    Invoke-DDivinationStep "Contrato OpenAPI e cliente TypeScript" {
        & (Join-Path $repoRoot "scripts\check-contract.ps1") -SkipInstall
    }
    Invoke-DDivinationStep "TypeScript strict" {
        & $nodeTools.Npm run lint:web
    }
    Invoke-DDivinationStep "Testes frontend" {
        & $nodeTools.Npm run test:web
    }
    Invoke-DDivinationStep "Build frontend" {
        & $nodeTools.Npm run build:web
    }
    Invoke-DDivinationStep "Budgets do bundle web" {
        & $nodeTools.Npm run check:bundle
    }

    if (-not $SkipE2E) {
        Assert-DDivinationPortAvailable -Port 8080
        Invoke-DDivinationStep "Playwright E2E" {
            & $nodeTools.Npm run test:e2e --workspace "@ddivination/web"
        }
    }

    Write-Host ""
    Write-Host "Todos os testes selecionados passaram." -ForegroundColor Green
} finally {
    $env:PATH = $previousPath
    Pop-Location
}
