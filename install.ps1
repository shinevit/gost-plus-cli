# PowerShell Installer for gost-tunnel
# Usage: irm https://github.com/shinevit/gost-plus-cli/raw/cli/install.ps1 | iex
#
# Prerequisites:
# If running from a downloaded file, you may need to enable script execution:
# Set-ExecutionPolicy RemoteSigned -Scope CurrentUser
#
param(
    [switch]$Latest
)

$ErrorActionPreference = "Stop"

# Check / Elevate Permissions
$CurrentPrincipal = [Security.Principal.WindowsPrincipal][Security.Principal.WindowsIdentity]::GetCurrent()
$IsAdmin = $CurrentPrincipal.IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)

if (-not $IsAdmin) {
    if ($PSScriptRoot) {
        # Script is running from a file, attempt to self-elevate
        Write-Host "Elevating permissions..." -ForegroundColor Yellow
        $Arguments = "& '" + $MyInvocation.MyCommand.Definition + "' " + $args
        Start-Process powershell -Verb RunAs -ArgumentList $Arguments
        exit
    } else {
        # Script is running from IEX (memory)
        Write-Warning "Script is not running as Administrator."
        Write-Warning "If you encounter permission errors, please run PowerShell as Administrator."
    }
}

$Repo = "shinevit/gost-plus-cli"
$BaseUrl = "https://api.github.com/repos/$Repo/releases"

function Install-GostTunnel {
    param (
        [string]$Version
    )

    # 1. Detect Architecture
    if ($env:PROCESSOR_ARCHITECTURE -eq "AMD64") {
        $Arch = "amd64"
    } elseif ($env:PROCESSOR_ARCHITECTURE -eq "x86") {
        $Arch = "386"
    } elseif ($env:PROCESSOR_ARCHITECTURE -eq "ARM64") {
        $Arch = "arm64"
    } else {
        Write-Error "Unsupported architecture: $env:PROCESSOR_ARCHITECTURE"
        exit 1
    }

    # 2. Get Download URL
    Write-Host "Fetching release info for $Repo ($Version)..." -ForegroundColor Cyan
    if ($Version -eq "latest") {
        $ReleaseUrl = "$BaseUrl/latest"
    } else {
        $ReleaseUrl = "$BaseUrl/tags/$Version"
    }

    try {
        $ReleaseJson = Invoke-RestMethod -Uri $ReleaseUrl
    } catch {
        Write-Error "Failed to fetch release info. Check version or network."
        exit 1
    }

    $Asset = $ReleaseJson.assets | Where-Object { $_.name -match "windows_$Arch.zip" } | Select-Object -First 1

    if (-not $Asset) {
        Write-Error "Could not find asset for windows_$Arch in release $($ReleaseJson.tag_name)"
        exit 1
    }

    $DownloadUrl = $Asset.browser_download_url
    $TagName = $ReleaseJson.tag_name
    Write-Host "Found version: $TagName" -ForegroundColor Green

    # 3. Prepare Install Directory
    # Using LocalAppData\Programs is standard for per-user installations
    $InstallDir = "$env:LOCALAPPDATA\Programs\gost.plus"
    if (-not (Test-Path $InstallDir)) {
        New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null
    }

    $ZipFile = "$InstallDir\gost-tunnel.zip"
    $BinName = "gost-tunnel.exe"

    # 4. Download
    Write-Host "Downloading from $DownloadUrl..." -ForegroundColor Cyan
    Invoke-WebRequest -Uri $DownloadUrl -OutFile $ZipFile

    # 5. Extract
    Write-Host "Extracting..." -ForegroundColor Cyan
    Expand-Archive -Path $ZipFile -DestinationPath "$InstallDir\temp_extract" -Force

    # 6. Install Binary
    $ExtractedBin = Get-ChildItem -Path "$InstallDir\temp_extract" -Filter $BinName -Recurse | Select-Object -First 1

    if ($ExtractedBin) {
        Copy-Item -Path $ExtractedBin.FullName -Destination "$InstallDir\$BinName" -Force
        Write-Host "Installed binary to $InstallDir\$BinName" -ForegroundColor Green
    } else {
        Write-Error "Binary '$BinName' not found in archive."
        exit 1
    }

    # 7. Cleanup
    Remove-Item $ZipFile -Force
    Remove-Item "$InstallDir\temp_extract" -Recurse -Force

    # 8. Add to PATH
    $CurrentPath = [System.Environment]::GetEnvironmentVariable("Path", [System.EnvironmentVariableTarget]::User)
    if ($CurrentPath -notlike "*$InstallDir*") {
        Write-Host "Adding $InstallDir to User PATH..." -ForegroundColor Cyan
        [System.Environment]::SetEnvironmentVariable("Path", $CurrentPath + ";$InstallDir", [System.EnvironmentVariableTarget]::User)
        Write-Host "Added to PATH. You may need to restart your terminal." -ForegroundColor Green
    } else {
        Write-Host "$InstallDir is already in PATH." -ForegroundColor Yellow
    }

    # 9. Final Instructions
    Write-Host "`nInstallation complete!" -ForegroundColor Green
    Write-Host "You can run the tool using:"
    Write-Host "  gost-tunnel -h" -ForegroundColor Yellow
}

if ($Latest) {
    Install-GostTunnel -Version "latest"
} else {
    Write-Host "Fetching available versions..." -ForegroundColor Cyan
    try {
        $Releases = Invoke-RestMethod -Uri $BaseUrl
        $Tags = $Releases.tag_name
    } catch {
        Write-Error "Failed to fetch releases."
        exit 1
    }

    Write-Host "Available gost-tunnel versions:" -ForegroundColor Cyan
    if ($Tags -is [string]) {
        $Tags = @($Tags)
    }
    
    for ($i = 0; $i -lt $Tags.Count; $i++) {
        Write-Host "$($i + 1)) $($Tags[$i])"
    }

    $Selection = Read-Host "Select a version (enter number)"
    if ($Selection -match "^\d+$" -and $Selection -gt 0 -and $Selection -le $Tags.Count) {
        $SelectedVersion = $Tags[$Selection - 1]
        Install-GostTunnel -Version $SelectedVersion
    } else {
        Write-Error "Invalid selection."
        exit 1
    }
}
