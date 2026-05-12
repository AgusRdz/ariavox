#Requires -Version 5.1
<#
.SYNOPSIS
    Installs ariavox — accessible PTY wrapper for AI coding agents.
.PARAMETER Version
    Version to install (e.g. "v0.1.0"). Defaults to latest release.
.PARAMETER InstallDir
    Directory to install the binary. Defaults to %LOCALAPPDATA%\Programs\ariavox.
.PARAMETER NoPathUpdate
    Skip adding the install directory to the user PATH.
.PARAMETER SkipChecksumVerify
    Skip SHA-256 checksum verification.
.EXAMPLE
    irm https://raw.githubusercontent.com/AgusRdz/ariavox/main/install.ps1 | iex
.EXAMPLE
    & ([scriptblock]::Create((irm https://raw.githubusercontent.com/AgusRdz/ariavox/main/install.ps1))) -Version v0.1.0
#>
[CmdletBinding()]
param(
    [string]$Version,
    [string]$InstallDir,
    [switch]$NoPathUpdate,
    [switch]$SkipChecksumVerify
)

$ErrorActionPreference = "Stop"

$Repo = "AgusRdz/ariavox"

# ── Helpers ──────────────────────────────────────────────────────────────────

function Write-Step  { param([string]$msg) Write-Host "  $msg" }
function Write-Ok    { param([string]$msg) Write-Host "  ok  $msg" -ForegroundColor Green }
function Write-Warn  { param([string]$msg) Write-Host "  warn  $msg" -ForegroundColor Yellow }
function Write-Fail  { param([string]$msg) Write-Error "  FAIL  $msg" }

# ── Architecture ──────────────────────────────────────────────────────────────

$Arch = if (
    [System.Runtime.InteropServices.RuntimeInformation]::ProcessArchitecture -eq
    [System.Runtime.InteropServices.Architecture]::Arm64
) { "arm64" } else { "amd64" }

$Binary = "ariavox-windows-$Arch.exe"

# ── Install directory ─────────────────────────────────────────────────────────

if (-not $InstallDir) {
    $InstallDir = if ($env:ARIAVOX_INSTALL_DIR) {
        $env:ARIAVOX_INSTALL_DIR
    } else {
        Join-Path $env:LOCALAPPDATA "Programs\ariavox"
    }
}

# ── Version ───────────────────────────────────────────────────────────────────

if (-not $Version) {
    $Version = $env:ARIAVOX_VERSION
}
if (-not $Version) {
    Write-Step "fetching latest release..."
    try {
        $release = Invoke-RestMethod `
            -Uri "https://api.github.com/repos/$Repo/releases/latest" `
            -Headers @{ "User-Agent" = "ariavox-installer" }
        $Version = $release.tag_name
    } catch {
        Write-Fail "could not fetch latest release: $_"
    }
}
if (-not $Version) {
    Write-Fail "failed to determine version to install"
}

# ── URLs ──────────────────────────────────────────────────────────────────────

$BaseUrl      = "https://github.com/$Repo/releases/download/$Version"
$BinaryUrl    = "$BaseUrl/$Binary"
$ChecksumsUrl = "$BaseUrl/checksums.txt"
$BundleUrl    = "$BaseUrl/checksums.txt.bundle"

Write-Host ""
Write-Host "ariavox installer"
Write-Host "-----------------"
Write-Step "version:  $Version"
Write-Step "platform: windows/$Arch"
Write-Step "dest:     $InstallDir"
Write-Host ""

# ── Download ──────────────────────────────────────────────────────────────────

New-Item -ItemType Directory -Force -Path $InstallDir | Out-Null
$TmpDir     = New-Item -ItemType Directory -Path (Join-Path $env:TEMP "ariavox-install-$([System.IO.Path]::GetRandomFileName())")
$TmpBinary  = Join-Path $TmpDir $Binary
$TmpChecks  = Join-Path $TmpDir "checksums.txt"
$TmpBundle  = Join-Path $TmpDir "checksums.txt.bundle"
$Destination = Join-Path $InstallDir "ariavox.exe"

try {
    Write-Step "downloading $Binary..."
    Invoke-WebRequest -Uri $BinaryUrl -OutFile $TmpBinary -UseBasicParsing

    # ── Checksum verification ─────────────────────────────────────────────────

    if (-not $SkipChecksumVerify) {
        Write-Step "verifying checksum..."
        try {
            Invoke-WebRequest -Uri $ChecksumsUrl -OutFile $TmpChecks -UseBasicParsing

            $expected = (Get-Content $TmpChecks |
                Where-Object { $_ -match [regex]::Escape($Binary) } |
                Select-Object -First 1) -split '\s+' | Select-Object -First 1

            if (-not $expected) {
                Write-Warn "binary not found in checksums.txt — skipping checksum verification"
            } else {
                $actual = (Get-FileHash -Algorithm SHA256 -Path $TmpBinary).Hash.ToLower()
                if ($actual -ne $expected.ToLower()) {
                    Remove-Item -Recurse -Force $TmpDir -ErrorAction SilentlyContinue
                    Write-Fail "checksum mismatch!`n  expected: $expected`n  actual:   $actual"
                }
                Write-Ok "checksum verified"
            }
        } catch {
            Write-Warn "could not download checksums.txt — skipping checksum verification ($_)"
        }
    }

    # ── Cosign verification (optional) ────────────────────────────────────────

    $cosign = Get-Command cosign -ErrorAction SilentlyContinue
    if ($cosign) {
        Write-Step "verifying cosign signature..."
        try {
            Invoke-WebRequest -Uri $BundleUrl -OutFile $TmpBundle -UseBasicParsing
            $result = & cosign verify-blob `
                --bundle $TmpBundle `
                --certificate-identity-regexp "https://github.com/$Repo/" `
                --certificate-oidc-issuer "https://token.actions.githubusercontent.com" `
                $TmpChecks 2>&1
            if ($LASTEXITCODE -eq 0) {
                Write-Ok "cosign signature verified"
            } else {
                Write-Warn "cosign verification failed: $result"
            }
        } catch {
            Write-Warn "cosign verification error: $_"
        }
    }

    # ── Install ───────────────────────────────────────────────────────────────

    Copy-Item -Force $TmpBinary $Destination
    Write-Ok "installed to $Destination"

} finally {
    Remove-Item -Recurse -Force $TmpDir -ErrorAction SilentlyContinue
}

# ── PATH ──────────────────────────────────────────────────────────────────────

if (-not $NoPathUpdate) {
    $CleanDir  = $InstallDir.TrimEnd("\")
    $UserPath  = [Environment]::GetEnvironmentVariable("PATH", "User")
    $PathParts = $UserPath -split ";" | ForEach-Object { $_.TrimEnd("\") }

    if ($PathParts -notcontains $CleanDir) {
        [Environment]::SetEnvironmentVariable("PATH", "$InstallDir;$UserPath", "User")
        Write-Ok "added $InstallDir to User PATH"
    }

    # Make available in the current session immediately
    $SessionParts = $env:PATH -split ";" | ForEach-Object { $_.TrimEnd("\") }
    if ($SessionParts -notcontains $CleanDir) {
        $env:PATH = "$InstallDir;$env:PATH"
    }

    # Broadcast WM_SETTINGCHANGE so Explorer and other processes pick up the new PATH
    try {
        $sig = @'
[DllImport("user32.dll", SetLastError=true, CharSet=CharSet.Auto)]
public static extern IntPtr SendMessageTimeout(
    IntPtr hWnd, uint Msg, IntPtr wParam, string lParam,
    uint fuFlags, uint uTimeout, out IntPtr lpdwResult);
'@
        $u32 = Add-Type -MemberDefinition $sig -Name "User32" -Namespace "Win32" -PassThru -ErrorAction SilentlyContinue
        $r   = [IntPtr]::Zero
        $u32::SendMessageTimeout([IntPtr]0xffff, 0x001a, [IntPtr]::Zero, "Environment", 2, 500, [ref]$r) | Out-Null
    } catch { <# non-critical #> }
}

# ── Done ──────────────────────────────────────────────────────────────────────

Write-Host ""
Write-Host "next steps:"
Write-Host ""
Write-Host "  ariavox run -- claude"
Write-Host "  ariavox run --sr -- claude          # screen reader mode"
Write-Host "  ariavox run --tts -- claude         # TTS mode"
Write-Host "  ariavox config show"
Write-Host "  ariavox doctor                      # verify system dependencies"
Write-Host ""
Write-Host "installation complete!" -ForegroundColor Green
