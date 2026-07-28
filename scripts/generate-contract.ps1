param(
    [switch]$SkipInstall
)

$ErrorActionPreference = "Stop"
. (Join-Path $PSScriptRoot "common.ps1")

$repoRoot = Get-DDivinationRepoRoot
$webRoot = Join-Path $repoRoot "apps\web"
$openAPIPath = Join-Path $webRoot "openapi.json"

Push-Location $repoRoot
try {
    $goExe = Resolve-DDivinationGo
    $nodeTools = Resolve-DDivinationNodeTools

    if (-not (Test-Path -LiteralPath (Join-Path $repoRoot "node_modules"))) {
        if ($SkipInstall) {
            throw "node_modules nao existe e -SkipInstall foi informado. Execute npm ci."
        }
        & $nodeTools.Npm ci
        if ($LASTEXITCODE -ne 0) {
            throw "npm ci falhou com codigo $LASTEXITCODE."
        }
    }

    Write-Host "Gerando OpenAPI..."
    $openAPIOutput = & $goExe -C (Join-Path $repoRoot "apps\server") run ./cmd/openapi
    if ($LASTEXITCODE -ne 0) {
        throw "Geracao do OpenAPI falhou com codigo $LASTEXITCODE."
    }
    $openAPIText = (($openAPIOutput -join "`n").TrimEnd() + "`n")
    $utf8WithoutBOM = New-Object System.Text.UTF8Encoding($false)
    [System.IO.File]::WriteAllText($openAPIPath, $openAPIText, $utf8WithoutBOM)

    Write-Host "Gerando cliente TypeScript..."
    & $nodeTools.Npm --workspace apps/web run api:generate
    if ($LASTEXITCODE -ne 0) {
        throw "Geracao do cliente TypeScript falhou com codigo $LASTEXITCODE."
    }

    Write-Host "Contrato e cliente gerados." -ForegroundColor Green
} finally {
    Pop-Location
}
