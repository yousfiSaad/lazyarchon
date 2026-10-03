# LazyArchon Windows Installation Script
# This script downloads and installs the latest release of LazyArchon on Windows

param(
    [string]$Version = "",
    [string]$InstallDir = "$env:USERPROFILE\.local\bin"
)

# Configuration
$Repo = "yousfisaad/lazyarchon"
$BinaryName = "lazyarchon"

# Error handling
$ErrorActionPreference = "Stop"

# Helper functions
function Write-ColorOutput($ForegroundColor, $Message) {
    $originalColor = $host.UI.RawUI.ForegroundColor
    $host.UI.RawUI.ForegroundColor = $ForegroundColor
    Write-Output $Message
    $host.UI.RawUI.ForegroundColor = $originalColor
}

function Write-Info($Message) {
    Write-ColorOutput Blue "Info: $Message"
}

function Write-Success($Message) {
    Write-ColorOutput Green "Success: $Message"
}

function Write-Error($Message) {
    Write-ColorOutput Red "Error: $Message"
    exit 1
}

function Write-Warning($Message) {
    Write-ColorOutput Yellow "Warning: $Message"
}

# Check if command exists
function Test-Command($Command) {
    try {
        Get-Command $Command -ErrorAction Stop | Out-Null
        return $true
    }
    catch {
        return $false
    }
}

# Detect platform
function Get-Platform {
    $arch = if ([Environment]::Is64BitOperatingSystem) { "amd64" } else { "386" }
    return "windows-$arch"
}

# Get latest release version
function Get-LatestVersion {
    try {
        $response = Invoke-RestMethod -Uri "https://api.github.com/repos/$Repo/releases/latest"
        return $response.tag_name
    }
    catch {
        Write-Error "Failed to fetch latest version: $_"
    }
}

# Download and extract binary
function Get-Binary($Version, $Platform) {
    $filename = "$BinaryName-$Version-$Platform.exe.zip"
    $downloadUrl = "https://github.com/$Repo/releases/download/$Version/$filename"
    
    $tempDir = [System.IO.Path]::GetTempPath()
    $tempFile = Join-Path $tempDir $filename
    $extractDir = Join-Path $tempDir "lazyarchon-extract"
    
    Write-Info "Downloading $filename..."
    
    try {
        # Create extraction directory
        New-Item -ItemType Directory -Path $extractDir -Force | Out-Null
        
        # Download file
        Invoke-WebRequest -Uri $downloadUrl -OutFile $tempFile
        
        Write-Info "Extracting binary..."
        
        # Extract zip file
        Add-Type -AssemblyName System.IO.Compression.FileSystem
        [System.IO.Compression.ZipFile]::ExtractToDirectory($tempFile, $extractDir)
        
        # Find the extracted binary
        $binaryPath = Join-Path $extractDir "$BinaryName-$Version-$Platform.exe"
        
        if (-not (Test-Path $binaryPath)) {
            Write-Error "Binary not found after extraction: $binaryPath"
        }
        
        return $binaryPath
    }
    catch {
        Write-Error "Failed to download or extract binary: $_"
    }
    finally {
        # Clean up download file
        if (Test-Path $tempFile) {
            Remove-Item $tempFile -Force
        }
    }
}

# Install binary
function Install-Binary($BinaryPath, $InstallPath) {
    Write-Info "Installing to $InstallPath..."
    
    try {
        # Create install directory if it doesn't exist
        $installDir = Split-Path $InstallPath -Parent
        New-Item -ItemType Directory -Path $installDir -Force | Out-Null
        
        # Copy binary to install directory
        Copy-Item $BinaryPath $InstallPath -Force
        
        # Verify installation
        $output = & $InstallPath --version 2>&1
        if ($LASTEXITCODE -ne 0) {
            Write-Error "Installation verification failed"
        }
        
        Write-Success "LazyArchon installed successfully!"
        
        # Check if install directory is in PATH
        $pathDirs = $env:PATH -split ';'
        $installDir = Split-Path $InstallPath -Parent
        
        if ($installDir -notin $pathDirs) {
            Write-Warning "Install directory $installDir is not in your PATH"
            Write-Info "You may need to add it to your PATH environment variable"
            Write-Info "Or restart your terminal to pick up changes"
        }
    }
    catch {
        Write-Error "Failed to install binary: $_"
    }
}

# Main installation function
function Main {
    Write-Output "LazyArchon Windows Installation Script"
    Write-Output "====================================="
    
    # Detect platform
    $platform = Get-Platform
    Write-Info "Detected platform: $platform"
    
    # Get version to install
    if (-not $Version) {
        Write-Info "Fetching latest release..."
        $Version = Get-LatestVersion
    }
    
    Write-Info "Installing LazyArchon $Version"
    
    # Download and extract
    $binaryPath = Get-Binary $Version $platform
    
    # Install
    $installPath = Join-Path $InstallDir "$BinaryName.exe"
    Install-Binary $binaryPath $installPath
    
    # Cleanup
    $extractDir = Split-Path $binaryPath -Parent
    if (Test-Path $extractDir) {
        Remove-Item $extractDir -Recurse -Force
    }
    
    Write-Output ""
    Write-Success "LazyArchon $Version installed successfully!"
    Write-Info "Run 'lazyarchon --help' to get started"
}

# Show help if requested
if ($args -contains "--help" -or $args -contains "-h") {
    Write-Output @"
LazyArchon Windows Installation Script

USAGE:
    .\install.ps1 [OPTIONS]

OPTIONS:
    -Version <version>      Install specific version (default: latest)
    -InstallDir <directory> Install directory (default: `$env:USERPROFILE\.local\bin)
    -h, --help             Show this help message

EXAMPLES:
    # Install latest version
    .\install.ps1

    # Install specific version
    .\install.ps1 -Version v0.1.0

    # Install to custom directory
    .\install.ps1 -InstallDir "C:\Program Files\LazyArchon"

"@
    exit 0
}

# Run main function
Main