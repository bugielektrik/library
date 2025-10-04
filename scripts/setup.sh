#!/bin/bash
set -e

echo "========================================"
echo " Library Service - Development Setup"
echo "========================================"
echo ""

# Check Go installation
if ! command -v go &> /dev/null; then
    echo "❌ Go is not installed. Please install Go 1.21 or later"
    exit 1
fi

GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
echo "✓ Found Go $GO_VERSION"
echo ""

# Install dependencies
echo "📦 Installing dependencies..."
go mod download
echo "✓ Dependencies downloaded"
echo ""

# Install development tools
echo "🔧 Installing development tools..."
echo "  - air (hot reload)"
go install github.com/cosmtrek/air@latest

echo "  - swag (swagger docs)"
go install github.com/swaggo/swag/cmd/swag@latest

echo "  - golangci-lint (linter)"
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

echo "  - mockery (mock generation)"
go install github.com/vektra/mockery/v2@latest

echo "✓ Tools installed"
echo ""

# Vendor dependencies
echo "📚 Vendoring dependencies..."
go mod vendor
echo "✓ Vendor complete"
echo ""

# Generate mocks
echo "🎭 Generating mocks..."
mockery
echo "✓ Mocks generated"
echo ""

# Setup docker services
echo "🐳 Starting Docker services..."
if command -v docker-compose &> /dev/null; then
    docker-compose up -d postgres redis
    echo "✓ Docker services started"
    echo "  - PostgreSQL: localhost:5432"
    echo "  - Redis: localhost:6379"
else
    echo "⚠️  docker-compose not found, skipping Docker setup"
fi
echo ""

# Wait for store
if command -v docker-compose &> /dev/null; then
    echo "⏳ Waiting for database to be ready..."
    sleep 3
    echo "✓ Database should be ready"
    echo ""

    # Run migrations
    echo "🗄️  Running database migrations..."
    go run ./cmd/migrate/ -direction=up || echo "⚠️  Migrations may have failed"
    echo ""
fi

# Generate swagger docs
echo "📖 Generating Swagger documentation..."
swag init -g cmd/api/main.go -o docs
echo "✓ Swagger docs generated"
echo ""

# Create .env if it doesn't exist
if [ ! -f .env ]; then
    echo "📝 Creating .env file from .env.dist..."
    cp .env.dist .env
    echo "✓ .env file created"
    echo "⚠️  Please update .env with your configuration"
    echo ""
fi

echo "========================================"
echo " ✅ Setup Complete!"
echo "========================================"
echo ""
echo "Next steps:"
echo "  1. Update .env with your configuration"
echo "  2. Run: make dev      (with hot reload)"
echo "     or:  make run      (without hot reload)"
echo ""
echo "Available commands:"
echo "  make help            Show all available commands"
echo "  make test            Run tests"
echo "  make swagger         Regenerate API docs"
echo "  make lint            Run linters"
echo ""
