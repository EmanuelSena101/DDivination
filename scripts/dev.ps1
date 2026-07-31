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
$webDist = Join-Path $repoRoot "apps\web\dist"
$backendProcess = $null
$webProcess = $null
$runtimeFileCreated = $false
$databaseManaged = $false
$previousDataDir = [Environment]::GetEnvironmentVariable("DDIVINATION_DATA_DIR", "Process")
$previousWebDir = [Environment]::GetEnvironmentVariable("DDIVINATION_WEB_DIR", "Process")
$previousDatabaseURL = [Environment]::GetEnvironmentVariable("DATABASE_URL", "Process")

Push-Location $repoRoot
try {
    $goExe = Resolve-DDivinationGo
    $nodeTools = Resolve-DDivinationNodeTools
    if (Test-Path -LiteralPath $runtimeFile) {
        throw "Ja existe uma execucao registrada. Execute .\scripts\stop.ps1 antes de iniciar novamente."
    }
    Assert-DDivinationPortAvailable -Port 8080
    Assert-DDivinationPortAvailable -Port 5173

    if ([string]::IsNullOrWhiteSpace($env:DATABASE_URL)) {
        & (Join-Path $repoRoot "scripts\database.ps1") -Action Up
        if ($LASTEXITCODE -ne 0) {
            throw "Nao foi possivel iniciar o PostgreSQL local."
        }
        $env:DATABASE_URL = "postgres://ddivination:ddivination@127.0.0.1:54329/ddivination?sslmode=disable"
        $databaseManaged = $true
    }

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

    Write-Host "Compilando frontend para o acesso LAN..."
    & $nodeTools.Npm run build:web
    if ($LASTEXITCODE -ne 0) {
        throw "Build do frontend falhou com codigo $LASTEXITCODE."
    }

    Write-Host "Compilando servidor Go..."
    & $goExe -C (Join-Path $repoRoot "apps\server") build -o $backendBinary ./cmd/ddivination
    if ($LASTEXITCODE -ne 0) {
        throw "Build do servidor falhou com codigo $LASTEXITCODE."
    }

    $env:DDIVINATION_DATA_DIR = $resolvedDataDir
    $env:DDIVINATION_WEB_DIR = $webDist
    $backendStart = @{
        FilePath = $backendBinary
        WorkingDirectory = $repoRoot
        RedirectStandardOutput = $backendLog
        RedirectStandardError = $backendErrorLog
        PassThru = $true
    }
    $webStart = @{
        FilePath = $nodeTools.Npm
        ArgumentList = @("run", "dev:web")
        WorkingDirectory = $repoRoot
        RedirectStandardOutput = $webLog
        RedirectStandardError = $webErrorLog
        PassThru = $true
    }
    if ($env:OS -eq "Windows_NT") {
        $backendStart.WindowStyle = "Hidden"
        $webStart.WindowStyle = "Hidden"
    }
    $backendProcess = Start-Process @backendStart
    $webProcess = Start-Process @webStart

    $runtimeRecord = [ordered]@{
        repoRoot = $repoRoot
        dataDir = $resolvedDataDir
        startedAtUtc = [DateTime]::UtcNow.ToString("o")
        databaseManaged = $databaseManaged
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
    Wait-DDivinationEndpoint -Uri "http://127.0.0.1:8080"
    Wait-DDivinationEndpoint -Uri "http://127.0.0.1:5173"

    Write-Host ""
    Write-Host "DDivination esta rodando." -ForegroundColor Green
    Write-Host "Frontend: http://127.0.0.1:5173"
    Write-Host "Backend:  http://127.0.0.1:8080"
    Write-Host "Dados:    $resolvedDataDir"
    Write-Host "Banco:    PostgreSQL"
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
    if ($databaseManaged) {
        & (Join-Path $repoRoot "scripts\database.ps1") -Action Down 2>$null
    }
    Write-Error $_
} finally {
    if ($null -eq $previousDataDir) {
        Remove-Item Env:DDIVINATION_DATA_DIR -ErrorAction SilentlyContinue
    } else {
        $env:DDIVINATION_DATA_DIR = $previousDataDir
    }
    if ($null -eq $previousWebDir) {
        Remove-Item Env:DDIVINATION_WEB_DIR -ErrorAction SilentlyContinue
    } else {
        $env:DDIVINATION_WEB_DIR = $previousWebDir
    }
    if ($null -eq $previousDatabaseURL) {
        Remove-Item Env:DATABASE_URL -ErrorAction SilentlyContinue
    } else {
        $env:DATABASE_URL = $previousDatabaseURL
    }
    Pop-Location
}
