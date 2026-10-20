# Copyright the open-caption authors
# SPDX-License-Identifier: Apache-2.0

<#
.SYNOPSIS
    Builds open-caption for Windows: bin\open-caption.exe. The same job as `make build` (Windows has no make).

.DESCRIPTION
    Run it through build.bat, which works whatever PowerShell's script policy is:

        .\build.bat
        .\build.bat -AcmeProviders none
        .\build.bat -AcmeProviders "easydns route53"

    -AcmeProviders  Which DNS providers for automatic HTTPS certificates are built in, exactly as ACME_PROVIDERS in the
                    Makefile (docs/10-acme-lets-encrypt.md): a list of names, "all", or "none". The default is
                    "cloudflare godaddy easydns". Anything but the default needs Git for Windows (it runs
                    tools/acme-go.sh) and an internet connection.
    -Output         Where to put the program (default bin\open-caption.exe).

    The build command is kept here, so the instructions for building on Windows stay "run build.bat" if building
    ever needs more steps.
#>
[CmdletBinding()]
param(
    [string]$AcmeProviders = 'cloudflare godaddy easydns',
    [string]$Output = 'bin\open-caption.exe'
)
$ErrorActionPreference = 'Stop'
Set-Location -LiteralPath $PSScriptRoot

function Fail([string]$message) { Write-Host "build: $message" -ForegroundColor Red; exit 1 }

if (-not (Get-Command go -ErrorAction SilentlyContinue)) {
    Fail "Go is not installed (or this window was opened before it was). Install it with 'winget install --id GoLang.Go -e', then close and reopen PowerShell. See docs\build-windows.md."
}
Write-Host ("build: " + (& go version))

# The version shown by `open-caption --version`: the git tag or commit, as in the Makefile.
$version = 'dev'
if (Get-Command git -ErrorAction SilentlyContinue) {
    $described = & git describe --tags --always --dirty 2>$null
    if ($LASTEXITCODE -eq 0 -and $described) { $version = "$described" }
}
$ldflags = "-s -w -X main.version=$version"

New-Item -ItemType Directory -Force -Path (Split-Path -Parent $Output) | Out-Null

$default = 'cloudflare easydns godaddy'
$wanted = (($AcmeProviders -replace '[,\s]+', ' ').Trim().ToLower() -split ' ' | Where-Object { $_ } | Sort-Object -Unique) -join ' '
if ($wanted -eq '' -or $wanted -eq $default) {
    # The ordinary build: nothing but Go is needed.
    & go build -ldflags $ldflags -o $Output .\cmd\open-caption
} else {
    # A different choice of providers: the same helper script as the Makefile, which needs a Unix shell (Git for Windows has one).
    $sh = $null
    $git = Get-Command git -ErrorAction SilentlyContinue
    if ($git) {
        $root = Split-Path -Parent (Split-Path -Parent $git.Source)
        foreach ($candidate in @("$root\usr\bin\sh.exe", "$root\bin\sh.exe")) { if (Test-Path $candidate) { $sh = $candidate; break } }
    }
    if (-not $sh) { Fail "-AcmeProviders '$AcmeProviders' needs Git for Windows (for its shell). Install it with 'winget install --id Git.Git -e', or build the default providers." }
    & $sh tools/acme-go.sh $AcmeProviders build -ldflags $ldflags -o $Output ./cmd/open-caption
}
if ($LASTEXITCODE -ne 0) { Fail "the build failed (see the messages above)." }
Write-Host "build: done. Run it with:  .\$Output" -ForegroundColor Green
