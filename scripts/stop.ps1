$ErrorActionPreference = "Stop"
. (Join-Path $PSScriptRoot "common.ps1")

$repoRoot = Get-DDivinationRepoRoot
$runtimeDir = Join-Path $repoRoot ".tmp\dev-runtime"
$runtimeFile = Join-Path $runtimeDir "processes.json"

if (-not (Test-Path -LiteralPath $runtimeFile)) {
    Write-Host "Nenhuma execucao iniciada por scripts\dev.ps1 foi encontrada."
    exit 0
}

$record = Get-Content -Raw -LiteralPath $runtimeFile | ConvertFrom-Json
if ([System.IO.Path]::GetFullPath($record.repoRoot) -ne $repoRoot) {
    throw "O registro de execucao pertence a outro workspace: $($record.repoRoot)"
}

function Stop-RecordedProcess {
    param(
        [Parameter(Mandatory = $true)]
        [string]$Name,
        [Parameter(Mandatory = $true)]
        $Entry
    )

    $processId = [int]$Entry.id
    $process = Get-Process -Id $processId -ErrorAction SilentlyContinue
    if ($null -eq $process) {
        Write-Host "$Name ja estava encerrado."
        return
    }

    $expectedStart = [DateTime]::Parse($Entry.startedAtUtc).ToUniversalTime()
    $actualStart = $process.StartTime.ToUniversalTime()
    if ([Math]::Abs(($actualStart - $expectedStart).TotalSeconds) -gt 2) {
        throw "PID $processId foi reutilizado por outro processo; $Name nao sera encerrado."
    }

    Write-Host "Encerrando $Name (PID $processId)..."
    & taskkill.exe /PID $processId /T /F 2>$null | Out-Null
}

Stop-RecordedProcess -Name "frontend" -Entry $record.web
Stop-RecordedProcess -Name "backend" -Entry $record.backend
Remove-Item -LiteralPath $runtimeFile -Force

Write-Host "DDivination encerrado." -ForegroundColor Green
Write-Host "Os logs foram preservados em $runtimeDir."
