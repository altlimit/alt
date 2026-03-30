# alt — One-line run script for Windows (PowerShell)
# Installs alt (if needed) and runs a tool in one command.
#
# Piped usage (one-liner):
#   powershell -Command "$env:ALT_RUN='user/repo [args...]'; iwr https://raw.githubusercontent.com/altlimit/alt/main/scripts/run.ps1 -useb | iex"
#
# File usage:
#   powershell -File run.ps1 user/repo [args...]

$ErrorActionPreference = "Stop"

$BinDir = "$env:LOCALAPPDATA\alt\bin"
$AltExe = "$BinDir\alt.exe"

# Resolve arguments: prefer $args, fall back to ALT_RUN env var (for piped iex usage)
if ($args.Count -gt 0) {
    $RunArgs = $args
} elseif ($env:ALT_RUN) {
    $RunArgs = $env:ALT_RUN -split '\s+'
    Remove-Item Env:ALT_RUN -ErrorAction SilentlyContinue
} else {
    Write-Host "Usage:"
    Write-Host "  powershell -Command `"`$env:ALT_RUN='user/repo'; iwr .../run.ps1 -useb | iex`""
    Write-Host "  powershell -File run.ps1 user/repo [args...]"
    Write-Host ""
    Write-Host "Examples:"
    Write-Host "  powershell -File run.ps1 altlimit/sitegen"
    Write-Host "  powershell -File run.ps1 altlimit/sitegen -serve"
    exit 1
}

# Install alt if it doesn't exist
if (-not (Test-Path $AltExe)) {
    Write-Host "Installing alt..."
    $InstallScript = (Invoke-WebRequest -Uri "https://raw.githubusercontent.com/altlimit/alt/main/scripts/install.ps1" -UseBasicParsing).Content
    Invoke-Expression $InstallScript
}

# Ensure alt is in PATH for this session
if ($env:Path -notlike "*$BinDir*") {
    $env:Path = "$BinDir;$env:Path"
}

# Run the tool — all arguments are passed through
& $AltExe run @RunArgs
