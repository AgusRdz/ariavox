$ErrorActionPreference = "Stop"

$Repo       = "AgusRdz/ariavox"
$InstallDir = if ($env:ARIAVOX_INSTALL_DIR) { $env:ARIAVOX_INSTALL_DIR } else { "$env:LOCALAPPDATA\Programs\ariavox" }

# Detect architecture
$Arch = if ([System.Runtime.InteropServices.RuntimeInformation]::ProcessArchitecture -eq [System.Runtime.InteropServices.Architecture]::Arm64) {
    "arm64"
} else {
    "amd64"
}

$Binary = "ariavox-windows-$Arch.exe"

# Get latest version
if (-not $env:ARIAVOX_VERSION) {
    $Release = Invoke-RestMethod "https://api.github.com/repos/$Repo/releases/latest"
    $env:ARIAVOX_VERSION = $Release.tag_name
}

if (-not $env:ARIAVOX_VERSION) {
    Write-Error "failed to determine latest version"
    exit 1
}

$Url         = "https://github.com/$Repo/releases/download/$($env:ARIAVOX_VERSION)/$Binary"
$Destination = Join-Path $InstallDir "ariavox.exe"

Write-Host "installing ariavox $($env:ARIAVOX_VERSION) (windows/$Arch)..."

New-Item -ItemType Directory -Force -Path $InstallDir | Out-Null
Invoke-WebRequest -Uri $Url -OutFile $Destination

Write-Host "installed ariavox to $Destination"
Write-Host ""

# Add to user PATH (persists across sessions)
$UserPath        = [Environment]::GetEnvironmentVariable("PATH", "User")
$CleanInstallDir = $InstallDir.TrimEnd("\")
$PathParts       = $UserPath -split ";" | ForEach-Object { $_.TrimEnd("\") }

if ($PathParts -notcontains $CleanInstallDir) {
    [Environment]::SetEnvironmentVariable("PATH", "$InstallDir;$UserPath", "User")
    Write-Host "added $InstallDir to User PATH"
}

# Make available in the current session immediately (no new terminal needed)
$CurrentPathParts = $env:PATH -split ";" | ForEach-Object { $_.TrimEnd("\") }
if ($CurrentPathParts -notcontains $CleanInstallDir) {
    $env:PATH = "$InstallDir;$env:PATH"
}

# Broadcast WM_SETTINGCHANGE so other processes pick up the new PATH
$HWND_BROADCAST = [IntPtr]0xffff
$WM_SETTINGCHANGE = 0x001a
$MethodDefinition = @'
[DllImport("user32.dll", SetLastError = true, CharSet = CharSet.Auto)]
public static extern IntPtr SendMessageTimeout(IntPtr hWnd, uint Msg, IntPtr wParam, string lParam, uint fuFlags, uint uTimeout, out IntPtr lpdwResult);
'@
$User32 = Add-Type -MemberDefinition $MethodDefinition -Name "User32" -Namespace "Win32" -PassThru
$result = [IntPtr]::Zero
$User32::SendMessageTimeout($HWND_BROADCAST, $WM_SETTINGCHANGE, [IntPtr]::Zero, "Environment", 2, 100, [ref]$result) | Out-Null

Write-Host ""
Write-Host "next steps:"
Write-Host ""
Write-Host "  ariavox run -- claude"
Write-Host "  ariavox run --sr -- claude          # screen reader mode"
Write-Host "  ariavox run --tts -- claude         # TTS mode"
Write-Host "  ariavox config show"
Write-Host "  ariavox doctor                      # verify system dependencies"
Write-Host ""
Write-Host "installation complete!"
