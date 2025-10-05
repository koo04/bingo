#!/bin/bash

# Build script for the Bingo application with embedded frontend
# This script builds the frontend and embeds it in the backend

# Default values
ARCH="amd64"  # or arm64, etc.
OS="linux"    # or windows, darwin, etc.
PRODUCTION=false  # whether to use production environment

# Show usage information
show_usage() {
    echo "Usage: $0 [OPTIONS]"
    echo "Build the Bingo application with embedded frontend"
    echo ""
    echo "Options:"
    echo "  -a, --arch ARCH     Target architecture (default: amd64)"
    echo "  -o, --os OS         Target operating system (default: linux)"
    echo "  -p, --production    Use production environment (.env.production)"
    echo "  -h, --help          Show this help message"
    echo ""
    echo "Examples:"
    echo "  $0 --arch arm64 --os darwin"
    echo "  $0 -a amd64 -o windows --production"
    echo "  $0 -p"
    echo "  $0"
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -a|--arch)
            ARCH="$2"
            shift 2
            ;;
        -o|--os)
            OS="$2"
            shift 2
            ;;
        -p|--production)
            PRODUCTION=true
            shift
            ;;
        -h|--help)
            show_usage
            exit 0
            ;;
        *)
            echo "Unknown option: $1"
            show_usage
            exit 1
            ;;
    esac
done

echo "Building Bingo application with embedded frontend for $OS/$ARCH..."

# Change to project root
cd "$(dirname "$0")"

# load the appropriate .env file if it exists
if [ "$PRODUCTION" = true ]; then
    if [ -f .env.production ]; then
        echo "Loading environment variables from .env.production file..."
        export $(grep -v '^#' .env.production | grep -v '^$' | xargs)
    else
        echo "Warning: .env.production file not found, continuing without environment variables..."
    fi
else
    if [ -f .env ]; then
        echo "Loading environment variables from .env file..."
        export $(grep -v '^#' .env | grep -v '^$' | xargs)
    fi
fi

# Build the frontend
echo "Building frontend..."
cd frontend
npm run build
if [ $? -ne 0 ]; then
    echo "Frontend build failed!"
    exit 1
fi

# Frontend build automatically outputs to backend/dist via Vite config
cd ..

# Build the backend with embedded files
echo "Building backend with embedded frontend..."
cd backend
env CGO_ENABLED=0 GOARCH=$ARCH GOOS=$OS go build -o ../server .
if [ $? -ne 0 ]; then
    echo "Backend build failed!"
    exit 1
fi

cd ..

echo "Build completed successfully!"
echo "The server binary now contains the embedded frontend."
echo "You can run it with: ./server"
