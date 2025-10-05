#!/usr/bin/env pwsh

# Build script for the Bingo application with embedded frontend
# This script builds the frontend and embeds it in the backend

Write-Host "Building Bingo application with embedded frontend..." -ForegroundColor Green

# Change to project root
Set-Location $PSScriptRoot

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
Write-Host "Building backend with embedded frontend..." -ForegroundColor Yellow
Set-Location backend
go build -o ../server.exe .
if ($LASTEXITCODE -ne 0) {
    Write-Host "Backend build failed!" -ForegroundColor Red
    exit 1
}

Set-Location ..

Write-Host "Build completed successfully!" -ForegroundColor Green
Write-Host "The server.exe file now contains the embedded frontend." -ForegroundColor Green
Write-Host "You can run it with: .\server.exe" -ForegroundColor Cyan
