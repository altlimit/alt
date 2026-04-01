# alt — One-line installer for Windows (PowerShell)
# Usage: powershell -Command "iwr https://raw.githubusercontent.com/altlimit/alt/main/scripts/install.ps1 -useb | iex"

$ErrorActionPreference = "Stop"

$Repo = "altlimit/alt"
$DataDir = "$env:LOCALAPPDATA\alt"
$InstallDir = "$DataDir\internal"
$BinDir = "$DataDir\bin"

# Detect Architecture
$Arch = if ([System.Environment]::Is64BitOperatingSystem) {
    if ($env:PROCESSOR_ARCHITECTURE -eq "ARM64") { "arm64" } else { "amd64" }
} else {
    "386"
}

Write-Host "Detected: windows/$Arch"

# Get latest release tag
try {
    $Headers = @{
        "User-Agent" = "alt-cli/1.0"
        "Accept"     = "application/vnd.github+json"
    }
    if ($env:GITHUB_TOKEN) {
        $Headers["Authorization"] = "Bearer $env:GITHUB_TOKEN"
    }
    $Release = Invoke-RestMethod -Uri "https://api.github.com/repos/$Repo/releases/latest" -Headers $Headers
    $Latest = $Release.tag_name
} catch {
    Write-Host "Error: Could not determine latest release."
    $StatusCode = $_.Exception.Response.StatusCode.value__
    if ($StatusCode -eq 403 -or $_.Exception.Message -match "rate limit") {
        Write-Host ""
        Write-Host "GitHub API rate limit exceeded. Set GITHUB_TOKEN to increase your limit:"
        Write-Host "  `$env:GITHUB_TOKEN = 'your_token'"
    } elseif ($StatusCode -eq 404) {
        Write-Host ""
        Write-Host "No releases found for $Repo."
        Write-Host "Check: https://github.com/$Repo/releases"
    } else {
        Write-Host $_.Exception.Message
    }
    exit 1
}

Write-Host "Latest version: $Latest"

# Build download URL
$BinaryName = "alt_windows_${Arch}.exe"
$DownloadUrl = "https://github.com/$Repo/releases/download/$Latest/$BinaryName"

# Create directories
foreach ($Dir in @($InstallDir, $BinDir)) {
    if (-not (Test-Path $Dir)) {
        New-Item -ItemType Directory -Path $Dir -Force | Out-Null
    }
}

# Download to internal dir
$AltPath = Join-Path $InstallDir "alt.exe"
Write-Host "Downloading $DownloadUrl..."
try {
    Invoke-WebRequest -Uri $DownloadUrl -OutFile $AltPath -UseBasicParsing
} catch {
    Write-Host "Error: Download failed."
    Write-Host $_.Exception.Message
    Write-Host ""
    Write-Host "Check that a release exists at: https://github.com/$Repo/releases"
    exit 1
}

# Copy to bin dir so alt is on PATH alongside installed tools
$BinAlt = Join-Path $BinDir "alt.exe"
Copy-Item -Path $AltPath -Destination $BinAlt -Force
Write-Host "Installed alt to $BinAlt"

# Add bin dir to User PATH
$CurrentPath = [Environment]::GetEnvironmentVariable("Path", "User")
if ($CurrentPath -notlike "*$BinDir*") {
    $NewPath = "$BinDir;$CurrentPath"
    [Environment]::SetEnvironmentVariable("Path", $NewPath, "User")
    Write-Host "Added $BinDir to User PATH"
    Write-Host ""
    Write-Host "NOTE: Restart your terminal for PATH changes to take effect."
} else {
    Write-Host "$BinDir is already in PATH"
}

Write-Host ""
Write-Host "alt $Latest installed successfully!"
Write-Host ""
Write-Host "Get started: alt install user/repo"
