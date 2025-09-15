#!/bin/bash

# Bingo Application Startup Script
# Usage: ./start.sh [development|dev]
#   - No arguments: Uses Docker if available, fallback to development mode
#   - development/dev: Forces local development mode (bypasses Docker)

echo "🎯 Starting Bingo Application..."

# Check for development mode parameter
DEVELOPMENT_MODE=false
if [ "$1" = "development" ] || [ "$1" = "dev" ]; then
    DEVELOPMENT_MODE=true
    echo "🔧 Development mode requested - starting locally..."
fi

# Check if .env file exists
if [ ! -f .env ]; then
    echo "⚠️  Warning: .env file not found. Creating from .env.example..."
    cp .env.example .env
    echo "📝 Please edit .env file with your Discord OAuth credentials before running the application."
    echo "   1. Go to https://discord.com/developers/applications"
    echo "   2. Create a new application"
    echo "   3. Add OAuth2 redirect URL: http://localhost:8080/auth/discord/callback"
    echo "   4. Copy Client ID and Client Secret to .env file"
    echo ""
    read -p "Press Enter to continue once you've configured .env file..."
fi

# Load environment variables from .env file
if [ -f .env ]; then
    echo "📋 Loading environment variables from .env file..."
    # Export variables from .env file (skip empty lines and comments)
    export $(grep -v '^#' .env | grep -v '^$' | xargs)
    echo "✅ Environment variables loaded"
fi

# Start in development mode if requested, otherwise use Docker if available
if [ "$DEVELOPMENT_MODE" = true ]; then
    echo "📦 Starting in development mode..."
elif command -v docker-compose &> /dev/null; then
    echo "🐳 Starting with Docker Compose..."
    docker-compose up --build
    exit 0
elif command -v docker &> /dev/null && docker compose version &> /dev/null; then
    echo "🐳 Starting with Docker Compose (newer syntax)..."
    docker compose up --build
    exit 0
else
    echo "📦 Docker not found. Falling back to development mode..."
fi

# Development mode execution (either requested or fallback)
# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "❌ Go is not installed. Please install Go or use Docker."
    exit 1
fi

# Check if Node.js/Yarn is installed
if ! command -v node &> /dev/null; then
    echo "❌ Node.js is not installed. Please install Node.js or use Docker."
    exit 1
fi

if ! command -v yarn &> /dev/null; then
    echo "❌ Yarn is not installed. Please install Yarn or use Docker."
    exit 1
fi
# Install frontend dependencies if needed
if [ ! -d "frontend/node_modules" ]; then
    echo "📦 Installing frontend dependencies..."
    cd frontend && yarn install && cd ..
fi

# Install backend dependencies if needed
cd backend
if [ ! -f "go.sum" ]; then
    echo "📦 Installing backend dependencies..."
    go mod tidy
fi
cd ..

echo "🚀 Starting backend and frontend..."

# Get the script's directory for reliable path navigation
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" &> /dev/null && pwd)"

# Start backend in background
cd "$SCRIPT_DIR/backend" && go run . &
BACKEND_PID=$!

# Wait a bit for backend to start
sleep 2

# Start frontend
cd "$SCRIPT_DIR/frontend" && yarn dev &
FRONTEND_PID=$!

echo "✅ Application started!"
echo "   Frontend: http://localhost:3000"
echo "   Backend:  http://localhost:8080"
echo ""
echo "Press Ctrl+C to stop both services..."

# Function to cleanup processes
cleanup() {
    echo ""
    echo "🛑 Stopping services..."
    kill $BACKEND_PID 2>/dev/null
    kill $FRONTEND_PID 2>/dev/null
    wait $BACKEND_PID 2>/dev/null
    wait $FRONTEND_PID 2>/dev/null
    echo "✅ Services stopped."
    exit 0
}

# Trap Ctrl+C
trap cleanup SIGINT

# Wait for both processes
wait $BACKEND_PID
wait $FRONTEND_PID
