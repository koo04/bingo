#!/bin/bash

# Build script for the Bingo application with embedded frontend
# This script builds the frontend and embeds it in the backend

echo "Building Bingo application with embedded frontend..."

# Change to project root
cd "$(dirname "$0")"

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
go build -o ../server.exe .
if [ $? -ne 0 ]; then
    echo "Backend build failed!"
    exit 1
fi

cd ..

echo "Build completed successfully!"
echo "The server binary now contains the embedded frontend."
echo "You can run it with: ./server"
