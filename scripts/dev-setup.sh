#!/bin/bash

# DixitMe Development Setup Script

set -e

echo "🎮 Setting up DixitMe Development Environment"

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "❌ Go is not installed. Please install Go 1.21+ first."
    exit 1
fi

# Check if Node.js is installed
if ! command -v node &> /dev/null; then
    echo "❌ Node.js is not installed. Please install Node.js 16+ first."
    exit 1
fi

# Check if PostgreSQL is running (optional check)
echo "📊 Checking PostgreSQL..."
if command -v psql &> /dev/null; then
    echo "✅ PostgreSQL is available"
else
    echo "⚠️  PostgreSQL not found. Make sure it's installed and running."
fi

# Check if Redis is running (optional check)
echo "🔄 Checking Redis..."
if command -v redis-cli &> /dev/null; then
    if redis-cli ping &> /dev/null; then
        echo "✅ Redis is running"
    else
        echo "⚠️  Redis is not responding. Make sure it's running."
    fi
else
    echo "⚠️  Redis not found. Make sure it's installed and running."
fi

# Install Go dependencies
echo "📦 Installing Go dependencies..."
go mod tidy

# Install Node.js dependencies
echo "📦 Installing Node.js dependencies..."
cd web
npm install
cd ..

# Generate placeholder cards if they don't exist
if [ ! -d "assets/cards" ] || [ -z "$(ls -A assets/cards)" ]; then
    echo "🃏 Generating placeholder cards..."
    go run scripts/generate-cards.go
else
    echo "✅ Card assets already exist"
fi

# Create .env file if it doesn't exist
if [ ! -f ".env" ]; then
    echo "⚙️  Creating .env file..."
    cp config.env.example .env
    echo "📝 Please edit .env with your database and Redis URLs"
else
    echo "✅ .env file already exists"
fi

echo ""
echo "🎉 Setup complete!"
echo ""
echo "To start development:"
echo "1. Make sure PostgreSQL and Redis are running"
echo "2. Update .env with your database URLs"
echo "3. Run 'go run cmd/server/main.go' to start the backend"
echo "4. In another terminal, run 'cd web && npm start' to start the frontend"
echo ""
echo "Backend will be available at: http://localhost:8080"
echo "Frontend will be available at: http://localhost:3000"
echo ""
echo "Happy coding! 🚀"
