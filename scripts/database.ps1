param(
    [ValidateSet("Up", "Down", "Reset", "Status")]
    [string]$Action = "Up"
)

$ErrorActionPreference = "Stop"
. (Join-Path $PSScriptRoot "common.ps1")

$repoRoot = Get-DDivinationRepoRoot
$docker = Get-Command docker -ErrorAction SilentlyContinue
$wsl = Get-Command wsl -ErrorAction SilentlyContinue
$useWSLDocker = $false
$composeRoot = $repoRoot
if ($null -eq $docker -and $null -ne $wsl) {
    & $wsl.Source -e docker version 2>$null | Out-Null
    if ($LASTEXITCODE -eq 0) {
        $useWSLDocker = $true
        $composeRoot = (& $wsl.Source -e wslpath -a $repoRoot).Trim()
    }
}
if ($null -eq $docker -and -not $useWSLDocker) {
    throw "Docker nao foi encontrado. Instale Docker Desktop ou configure DATABASE_URL para um PostgreSQL existente."
}
$composeFile = if ($useWSLDocker) {
    "$composeRoot/compose.postgres.yaml"
} else {
    Join-Path $composeRoot "compose.postgres.yaml"
}

function Invoke-DDivinationDocker {
    param([Parameter(ValueFromRemainingArguments = $true)][string[]]$Arguments)
    $previousPreference = $ErrorActionPreference
    $ErrorActionPreference = "Continue"
    try {
        if ($useWSLDocker) {
            $output = & $wsl.Source -e docker @Arguments 2>&1
        } else {
            $output = & $docker.Source @Arguments 2>&1
        }
        $exitCode = $LASTEXITCODE
    } finally {
        $ErrorActionPreference = $previousPreference
    }
    foreach ($line in $output) {
        Write-Output ([string]$line)
    }
    $global:LASTEXITCODE = $exitCode
}

function Invoke-DDivinationCompose {
    param([Parameter(ValueFromRemainingArguments = $true)][string[]]$Arguments)
    Invoke-DDivinationDocker compose --project-directory $composeRoot -f $composeFile @Arguments
    if ($LASTEXITCODE -ne 0) {
        throw "docker compose falhou com codigo $LASTEXITCODE."
    }
}

function Wait-DDivinationPostgres {
    $deadline = [DateTime]::UtcNow.AddSeconds(60)
    do {
        Invoke-DDivinationDocker -Arguments @(
            "compose", "--project-directory", $composeRoot, "-f", $composeFile,
            "exec", "-T", "postgres", "pg_isready", "-U", "ddivination", "-d", "ddivination"
        ) 2>$null | Out-Null
        if ($LASTEXITCODE -eq 0) {
            Write-Host "PostgreSQL pronto em 127.0.0.1:54329." -ForegroundColor Green
            return
        }
        Start-Sleep -Milliseconds 500
    } while ([DateTime]::UtcNow -lt $deadline)
    throw "Tempo esgotado aguardando o PostgreSQL local."
}

Push-Location $repoRoot
try {
    switch ($Action) {
        "Up" {
            Invoke-DDivinationCompose -Arguments @("up", "-d", "postgres")
            Wait-DDivinationPostgres
        }
        "Down" {
            Invoke-DDivinationCompose -Arguments @("stop", "postgres")
            Write-Host "PostgreSQL local encerrado; dados preservados." -ForegroundColor Green
        }
        "Reset" {
            Invoke-DDivinationCompose -Arguments @("down", "--volumes", "--remove-orphans")
            Invoke-DDivinationCompose -Arguments @("up", "-d", "postgres")
            Wait-DDivinationPostgres
            Write-Host "PostgreSQL local recriado. Os dados anteriores foram removidos." -ForegroundColor Yellow
        }
        "Status" {
            Invoke-DDivinationCompose -Arguments @("ps", "postgres")
        }
    }
} finally {
    Pop-Location
}
