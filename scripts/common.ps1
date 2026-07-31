Set-StrictMode -Version Latest

$script:DDivinationRepoRoot = [System.IO.Path]::GetFullPath(
    (Join-Path $PSScriptRoot "..")
)

function Get-DDivinationRepoRoot {
    return $script:DDivinationRepoRoot
}

function Resolve-DDivinationGo {
    $candidate = Get-Command go -ErrorAction SilentlyContinue
    $goPath = if ($null -ne $candidate) { $candidate.Source } else { $null }

    if ([string]::IsNullOrWhiteSpace($goPath)) {
        $toolsRoot = Join-Path $script:DDivinationRepoRoot ".tools"
        if (Test-Path -LiteralPath $toolsRoot) {
            $goPath = Get-ChildItem -LiteralPath $toolsRoot -Directory |
                Where-Object { $_.Name -like "go*" } |
                Sort-Object LastWriteTimeUtc -Descending |
                ForEach-Object {
                    $portable = Join-Path $_.FullName "go\bin\go.exe"
                    if (Test-Path -LiteralPath $portable) {
                        $portable
                    }
                } |
                Select-Object -First 1
        }
    }

    if ([string]::IsNullOrWhiteSpace($goPath)) {
        throw "Go 1.26+ nao foi encontrado no PATH nem em .tools. Instale o Go ou restaure o runtime portatil."
    }

    $versionOutput = & $goPath version
    if ($LASTEXITCODE -ne 0 -or $versionOutput -notmatch "go version go(\d+)\.(\d+)") {
        throw "Nao foi possivel determinar a versao do Go em '$goPath'."
    }
    $major = [int]$Matches[1]
    $minor = [int]$Matches[2]
    if ($major -lt 1 -or ($major -eq 1 -and $minor -lt 26)) {
        throw "DDivination requer Go 1.26+. Encontrado: $versionOutput"
    }

    return $goPath
}

function Resolve-DDivinationNodeTools {
    $node = Get-Command node -ErrorAction SilentlyContinue
    # Prefer the Windows command shim. Start-Process cannot launch npm.ps1
    # reliably when the user's execution policy blocks unsigned scripts.
    $npm = Get-Command npm.cmd -ErrorAction SilentlyContinue
    if ($null -eq $npm) {
        $npm = Get-Command npm -ErrorAction SilentlyContinue
    }
    if ($null -eq $node -or $null -eq $npm) {
        throw "Node.js 24+ e npm 11+ sao obrigatorios. Instale-os antes de continuar."
    }

    $nodeVersion = (& $node.Source --version).TrimStart("v")
    $npmVersion = (& $npm.Source --version)
    if ($LASTEXITCODE -ne 0) {
        throw "Nao foi possivel executar npm."
    }
    $nodeMajor = [int]($nodeVersion.Split(".")[0])
    $npmMajor = [int]($npmVersion.Split(".")[0])
    if ($nodeMajor -lt 24) {
        throw "DDivination requer Node.js 24+. Encontrado: $nodeVersion"
    }
    if ($npmMajor -lt 11) {
        throw "DDivination requer npm 11+. Encontrado: $npmVersion"
    }

    return [pscustomobject]@{
        Node = $node.Source
        Npm = $npm.Source
        NodeVersion = $nodeVersion
        NpmVersion = $npmVersion
    }
}

function Assert-DDivinationPortAvailable {
    param(
        [Parameter(Mandatory = $true)]
        [int]$Port
    )

    $listeners = [System.Net.NetworkInformation.IPGlobalProperties]::GetIPGlobalProperties().GetActiveTcpListeners()
    if (-not ($listeners | Where-Object { $_.Port -eq $Port })) {
        return
    }
    throw "A porta $Port ja esta em uso. Execute .\scripts\stop.ps1 ou encerre o processo responsavel."
}

function Wait-DDivinationEndpoint {
    param(
        [Parameter(Mandatory = $true)]
        [string]$Uri,
        [int]$TimeoutSeconds = 60
    )

    $deadline = [DateTime]::UtcNow.AddSeconds($TimeoutSeconds)
    do {
        try {
            $response = Invoke-WebRequest -UseBasicParsing -Uri $Uri -TimeoutSec 2
            if ($response.StatusCode -ge 200 -and $response.StatusCode -lt 500) {
                return
            }
        } catch {
            Start-Sleep -Milliseconds 300
        }
    } while ([DateTime]::UtcNow -lt $deadline)

    throw "Tempo esgotado aguardando $Uri."
}

function Stop-DDivinationProcessTree {
    param(
        [Parameter(Mandatory = $true)]
        [System.Diagnostics.Process]$Process
    )

    if ($Process.HasExited) {
        return
    }
    if ($env:OS -eq "Windows_NT") {
        & taskkill.exe /PID $Process.Id /T /F 2>$null | Out-Null
    } else {
        $pkill = Get-Command pkill -ErrorAction SilentlyContinue
        if ($null -ne $pkill) {
            & $pkill.Source -TERM -P $Process.Id 2>$null
        }
        Stop-Process -Id $Process.Id -Force -ErrorAction SilentlyContinue
    }
}
