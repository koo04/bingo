# Bingo Application Startup Script (PowerShell)
# Usage: .\start.ps1 [development|dev]
#   - No arguments: Uses Docker if available, fallback to development mode
#   - development/dev: Forces local development mode (bypasses Docker)

Write-Host "🎯 Starting Bingo Application..."

# Check for development mode parameter
$DEVELOPMENT_MODE = $false
if ($args.Count -gt 0 -and ($args[0] -eq "development" -or $args[0] -eq "dev")) {
    $DEVELOPMENT_MODE = $true
    Write-Host "🔧 Development mode requested - starting locally..."
}

# Check if .env file exists
if (-not (Test-Path .env)) {
    Write-Host "⚠️  Warning: .env file not found. Creating from .env.example..."
    Copy-Item .env.example .env
    Write-Host "📝 Please edit .env file with your Discord OAuth credentials before running the application."
    Write-Host "   1. Go to https://discord.com/developers/applications"
    Write-Host "   2. Create a new application"
    Write-Host "   3. Add OAuth2 redirect URL: http://localhost:8080/auth/discord/callback"
    Write-Host "   4. Copy Client ID and Client Secret to .env file"
    Write-Host ""
    Read-Host "Press Enter to continue once you've configured .env file..."
}

# Load environment variables from .env file
if (Test-Path .env) {
    Write-Host "📋 Loading environment variables from .env file..."
    Get-Content .env | Where-Object { $_ -notmatch '^#' -and $_ -ne '' } | ForEach-Object {
        $parts = $_ -split '=', 2
        if ($parts.Length -eq 2) {
            [System.Environment]::SetEnvironmentVariable($parts[0], $parts[1])
        }
    }
    Write-Host "✅ Environment variables loaded"
}

# Start in development mode if requested, otherwise use Docker if available
if ($DEVELOPMENT_MODE) {
    Write-Host "📦 Starting in development mode..."
} elseif (Get-Command docker-compose -ErrorAction SilentlyContinue) {
    Write-Host "🐳 Starting with Docker Compose..."
    docker-compose up --build
    exit 0
} elseif (Get-Command docker -ErrorAction SilentlyContinue) {
    $dockerComposeVersion = docker compose version 2>$null
    if ($LASTEXITCODE -eq 0) {
        Write-Host "🐳 Starting with Docker Compose (newer syntax)..."
        docker compose up --build
        exit 0
    } else {
        Write-Host "📦 Docker not found. Falling back to development mode..."
    }
} else {
    Write-Host "📦 Docker not found. Falling back to development mode..."
}

# Development mode execution (either requested or fallback)
# Check if Go is installed
if (-not (Get-Command go -ErrorAction SilentlyContinue)) {
    Write-Host "❌ Go is not installed. Please install Go or use Docker."
    exit 1
}

# Check if Node.js/Yarn is installed
if (-not (Get-Command node -ErrorAction SilentlyContinue)) {
    Write-Host "❌ Node.js is not installed. Please install Node.js or use Docker."
    exit 1
}

if (-not (Get-Command yarn -ErrorAction SilentlyContinue)) {
    Write-Host "❌ Yarn is not installed. Please install Yarn or use Docker."
    exit 1
}

# Install frontend dependencies if needed
if (-not (Test-Path "frontend/node_modules")) {
    Write-Host "📦 Installing frontend dependencies..."
    Push-Location frontend
    yarn install
    Pop-Location
}

# Install backend dependencies if needed
Push-Location backend
if (-not (Test-Path "go.sum")) {
    Write-Host "📦 Installing backend dependencies..."
    go mod tidy
}
Pop-Location

Write-Host "🚀 Starting backend and frontend..."

# Start backend in background
$backendJob = Start-Job -ScriptBlock {
    Set-Location "$PWD/backend"
    go run .
}
Start-Sleep -Seconds 2

# Start frontend in background
$frontendJob = Start-Job -ScriptBlock {
    Set-Location "$PWD/frontend"
    yarn dev
}

Write-Host "✅ Application started!"
Write-Host "   Frontend: http://localhost:3000"
Write-Host "   Backend:  http://localhost:8080"
Write-Host ""
Write-Host "Press Ctrl+C to stop both services..."

# Function to cleanup processes
function Cleanup {
    Write-Host ""
    Write-Host "🛑 Stopping services..."
    if ($backendJob) { Stop-Job $backendJob; Remove-Job $backendJob }
    if ($frontendJob) { Stop-Job $frontendJob; Remove-Job $frontendJob }
    Write-Host "✅ Services stopped."
    exit 0
}

# Trap Ctrl+C
$null = Register-EngineEvent PowerShell.Exiting -Action { Cleanup }

# Wait for both jobs
Wait-Job $backendJob, $frontendJob
