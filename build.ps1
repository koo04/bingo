#!/usr/bin/env pwsh

# Build script for the Bingo application with embedded frontend
# This script builds the frontend and embeds it in the backend

param(
    [Parameter(HelpMessage="Target architecture (e.g., amd64, arm64)")]
    [string]$Arch = "amd64",
    
    [Parameter(HelpMessage="Target operating system (e.g., windows, linux, darwin)")]
    [string]$OS = "windows",
    
    [Parameter(HelpMessage="Use production environment (.env.production)")]
    [switch]$Production,
    
    [Parameter(HelpMessage="Show help information")]
    [switch]$Help
)

# Show usage information
if ($Help) {
    Write-Host "Build script for the Bingo application with embedded frontend" -ForegroundColor Green
    Write-Host ""
    Write-Host "SYNTAX:" -ForegroundColor Yellow
    Write-Host "    .\build.ps1 [-Arch <string>] [-OS <string>] [-Help]"
    Write-Host ""
    Write-Host "PARAMETERS:" -ForegroundColor Yellow
    Write-Host "    -Arch <string>"
    Write-Host "        Target architecture (default: amd64)"
    Write-Host "        Valid values: amd64, arm64, 386, etc."
    Write-Host ""
    Write-Host "    -OS <string>"
    Write-Host "        Target operating system (default: windows)"
    Write-Host "        Valid values: windows, linux, darwin, etc."
    Write-Host ""
    Write-Host "    -Production"
    Write-Host "        Use production environment (.env.production)"
    Write-Host ""
    Write-Host "    -Help"
    Write-Host "        Show this help message"
    Write-Host ""
    Write-Host "EXAMPLES:" -ForegroundColor Yellow
    Write-Host "    .\build.ps1"
    Write-Host "    .\build.ps1 -Arch arm64 -OS darwin"
    Write-Host "    .\build.ps1 -OS linux -Production"
    Write-Host "    .\build.ps1 -Production"
    exit 0
}

Write-Host "Building Bingo application with embedded frontend..." -ForegroundColor Green

# Change to project root
Set-Location $PSScriptRoot

# Load the appropriate .env file if it exists
if ($Production) {
    if (Test-Path ".env.production") {
        Write-Host "Loading environment variables from .env.production file..." -ForegroundColor Cyan
        Get-Content ".env.production" | Where-Object { $_ -notmatch '^#' -and $_ -notmatch '^\s*$' } | ForEach-Object {
            $name, $value = $_ -split '=', 2
            [Environment]::SetEnvironmentVariable($name.Trim(), $value.Trim(), "Process")
        }
    } else {
        Write-Host "Warning: .env.production file not found, continuing without environment variables..." -ForegroundColor Yellow
    }
} else {
    if (Test-Path ".env") {
        Write-Host "Loading environment variables from .env file..." -ForegroundColor Cyan
        Get-Content ".env" | Where-Object { $_ -notmatch '^#' -and $_ -notmatch '^\s*$' } | ForEach-Object {
            $name, $value = $_ -split '=', 2
            [Environment]::SetEnvironmentVariable($name.Trim(), $value.Trim(), "Process")
        }
    }
}

# Build the frontend
Write-Host "Building frontend..." -ForegroundColor Yellow
Set-Location frontend
npm run build
if ($LASTEXITCODE -ne 0) {
    Write-Host "Frontend build failed!" -ForegroundColor Red
    exit 1
}

# Frontend build automatically outputs to backend\dist via Vite config
Set-Location ..

# Build the backend with embedded files
Write-Host "Building backend with embedded frontend for $OS/$Arch..." -ForegroundColor Yellow
Set-Location backend

# Set the output filename based on OS
$outputFile = if ($OS -eq "windows") { "../server.exe" } else { "../server" }

# Set environment variables and build
$env:CGO_ENABLED = "0"
$env:GOARCH = $Arch
$env:GOOS = $OS
go build -o $outputFile .
if ($LASTEXITCODE -ne 0) {
    Write-Host "Backend build failed!" -ForegroundColor Red
    exit 1
}

Set-Location ..

Write-Host "Build completed successfully!" -ForegroundColor Green
$serverFile = if ($OS -eq "windows") { "server.exe" } else { "server" }
Write-Host "The $serverFile file now contains the embedded frontend." -ForegroundColor Green
$runCommand = if ($OS -eq "windows") { ".\server.exe" } else { "./server" }
Write-Host "You can run it with: $runCommand" -ForegroundColor Cyan
