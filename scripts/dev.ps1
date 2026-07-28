param(
    [string]$DataDir = "",
    [switch]$SkipInstall
)

$ErrorActionPreference = "Stop"
. (Join-Path $PSScriptRoot "common.ps1")

$repoRoot = Get-DDivinationRepoRoot
$runtimeDir = Join-Path $repoRoot ".tmp\dev-runtime"
$runtimeFile = Join-Path $runtimeDir "processes.json"
$backendLog = Join-Path $runtimeDir "backend.log"
$backendErrorLog = Join-Path $runtimeDir "backend.error.log"
$webLog = Join-Path $runtimeDir "web.log"
$webErrorLog = Join-Path $runtimeDir "web.error.log"
$backendBinary = Join-Path $runtimeDir "ddivination-dev.exe"
$backendProcess = $null
$webProcess = $null
$runtimeFileCreated = $false
$previousDataDir = [Environment]::GetEnvironmentVariable("DDIVINATION_DATA_DIR", "Process")

Push-Location $repoRoot
try {
    $goExe = Resolve-DDivinationGo
    $nodeTools = Resolve-DDivinationNodeTools
    if (Test-Path -LiteralPath $runtimeFile) {
        throw "Ja existe uma execucao registrada. Execute .\scripts\stop.ps1 antes de iniciar novamente."
    }
    Assert-DDivinationPortAvailable -Port 8080
    Assert-DDivinationPortAvailable -Port 5173

    if (-not (Test-Path -LiteralPath (Join-Path $repoRoot "node_modules"))) {
        if ($SkipInstall) {
            throw "node_modules nao existe e -SkipInstall foi informado. Execute npm ci."
        }
        Write-Host "Instalando dependencias npm..."
        & $nodeTools.Npm ci
        if ($LASTEXITCODE -ne 0) {
            throw "npm ci falhou com codigo $LASTEXITCODE."
        }
    }

    New-Item -ItemType Directory -Path $runtimeDir -Force | Out-Null

    $resolvedDataDir = if ([string]::IsNullOrWhiteSpace($DataDir)) {
        Join-Path $repoRoot ".tmp\dev-data"
    } elseif ([System.IO.Path]::IsPathRooted($DataDir)) {
        [System.IO.Path]::GetFullPath($DataDir)
    } else {
        [System.IO.Path]::GetFullPath((Join-Path $repoRoot $DataDir))
    }
    New-Item -ItemType Directory -Path $resolvedDataDir -Force | Out-Null

    Write-Host "Compilando servidor Go..."
    & $goExe -C (Join-Path $repoRoot "apps\server") build -o $backendBinary ./cmd/ddivination
    if ($LASTEXITCODE -ne 0) {
        throw "Build do servidor falhou com codigo $LASTEXITCODE."
    }

    $env:DDIVINATION_DATA_DIR = $resolvedDataDir
    $backendProcess = Start-Process `
        -FilePath $backendBinary `
        -WorkingDirectory $repoRoot `
        -RedirectStandardOutput $backendLog `
        -RedirectStandardError $backendErrorLog `
        -WindowStyle Hidden `
        -PassThru

    $webProcess = Start-Process `
        -FilePath $nodeTools.Npm `
        -ArgumentList @("run", "dev:web") `
        -WorkingDirectory $repoRoot `
        -RedirectStandardOutput $webLog `
        -RedirectStandardError $webErrorLog `
        -WindowStyle Hidden `
        -PassThru

    $runtimeRecord = [ordered]@{
        repoRoot = $repoRoot
        dataDir = $resolvedDataDir
        startedAtUtc = [DateTime]::UtcNow.ToString("o")
        backend = [ordered]@{
            id = $backendProcess.Id
            startedAtUtc = $backendProcess.StartTime.ToUniversalTime().ToString("o")
            executable = $backendBinary
        }
        web = [ordered]@{
            id = $webProcess.Id
            startedAtUtc = $webProcess.StartTime.ToUniversalTime().ToString("o")
            executable = $nodeTools.Npm
        }
    }
    $runtimeRecord | ConvertTo-Json -Depth 4 | Set-Content -LiteralPath $runtimeFile -Encoding utf8
    $runtimeFileCreated = $true

    Wait-DDivinationEndpoint -Uri "http://127.0.0.1:8080/api/v1/health"
    Wait-DDivinationEndpoint -Uri "http://127.0.0.1:5173"

    Write-Host ""
    Write-Host "DDivination esta rodando." -ForegroundColor Green
    Write-Host "Frontend: http://127.0.0.1:5173"
    Write-Host "Backend:  http://127.0.0.1:8080"
    Write-Host "Dados:    $resolvedDataDir"
    Write-Host "Logs:     $runtimeDir"
    Write-Host ""
    Write-Host "Para encerrar: .\scripts\stop.ps1"
} catch {
    if ($null -ne $webProcess) {
        Stop-DDivinationProcessTree -Process $webProcess
    }
    if ($null -ne $backendProcess) {
        Stop-DDivinationProcessTree -Process $backendProcess
    }
    if ($runtimeFileCreated -and (Test-Path -LiteralPath $runtimeFile)) {
        Remove-Item -LiteralPath $runtimeFile -Force
    }
    Write-Error $_
} finally {
    if ($null -eq $previousDataDir) {
        Remove-Item Env:DDIVINATION_DATA_DIR -ErrorAction SilentlyContinue
    } else {
        $env:DDIVINATION_DATA_DIR = $previousDataDir
    }
    Pop-Location
}
