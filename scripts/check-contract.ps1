param(
    [switch]$SkipInstall
)

$ErrorActionPreference = "Stop"
. (Join-Path $PSScriptRoot "common.ps1")

$repoRoot = Get-DDivinationRepoRoot
$webRoot = Join-Path $repoRoot "apps\web"
$openAPIPath = Join-Path $webRoot "openapi.json"
$clientRoot = Join-Path $webRoot "src\api\generated"

function Get-GeneratedTreeFingerprint {
    param(
        [Parameter(Mandatory = $true)]
        [string]$Root
    )

    if (-not (Test-Path -LiteralPath $Root)) {
        return "<missing>"
    }
    return (
        Get-ChildItem -LiteralPath $Root -File -Recurse |
            Sort-Object FullName |
            ForEach-Object {
                $relative = $_.FullName.Substring($Root.Length).TrimStart("\", "/")
                $content = [System.IO.File]::ReadAllText($_.FullName)
                $normalized = $content.Replace("`r`n", "`n").Replace("`r", "`n")
                $bytes = [System.Text.Encoding]::UTF8.GetBytes($normalized)
                $hasher = [System.Security.Cryptography.SHA256]::Create()
                try {
                    $hashBytes = $hasher.ComputeHash($bytes)
                } finally {
                    $hasher.Dispose()
                }
                $hash = [System.BitConverter]::ToString($hashBytes).Replace("-", "")
                "$relative`:$hash"
            }
    ) -join "`n"
}

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

    $openAPIOutput = & $goExe -C (Join-Path $repoRoot "apps\server") run ./cmd/openapi
    if ($LASTEXITCODE -ne 0) {
        throw "Geracao do OpenAPI falhou com codigo $LASTEXITCODE."
    }
    $expectedOpenAPI = ($openAPIOutput -join "`n").Trim()
    $currentOpenAPI = if (Test-Path -LiteralPath $openAPIPath) {
        [System.IO.File]::ReadAllText($openAPIPath).Trim()
    } else {
        ""
    }
    if ($currentOpenAPI -cne $expectedOpenAPI) {
        throw "OpenAPI esta desatualizado. Execute .\scripts\generate-contract.ps1."
    }

    $before = Get-GeneratedTreeFingerprint -Root $clientRoot
    & $nodeTools.Npm --workspace apps/web run api:generate
    if ($LASTEXITCODE -ne 0) {
        throw "Geracao do cliente TypeScript falhou com codigo $LASTEXITCODE."
    }
    $after = Get-GeneratedTreeFingerprint -Root $clientRoot
    if ($before -cne $after) {
        throw "Cliente TypeScript estava desatualizado. Revise e versione os arquivos regenerados."
    }

    Write-Host "OpenAPI e cliente TypeScript estao sincronizados." -ForegroundColor Green
} finally {
    Pop-Location
}
