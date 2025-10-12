#!/bin/bash
# Development Environment Setup Script
# Sets up complete development environment from scratch

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}========================================"
echo " Library Service - Dev Setup"
echo "========================================${NC}"
echo ""

#######################################
# Prerequisites Check
#######################################

echo -e "${YELLOW}🔍 Checking prerequisites...${NC}"

# Check Go
if ! command -v go &> /dev/null; then
    echo -e "${RED}❌ Go is not installed${NC}"
    echo "   Please install Go 1.21 or later from https://go.dev/dl/"
    exit 1
fi
GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
echo -e "${GREEN}✓ Go $GO_VERSION${NC}"

# Check Docker
if ! command -v docker &> /dev/null; then
    echo -e "${RED}❌ Docker is not installed${NC}"
    echo "   Please install Docker from https://docs.docker.com/get-docker/"
    exit 1
fi
echo -e "${GREEN}✓ Docker$(docker --version | awk '{print " " $3}' | sed 's/,//')${NC}"

# Check Docker Compose
if ! command -v docker-compose &> /dev/null && ! docker compose version &> /dev/null; then
    echo -e "${RED}❌ Docker Compose is not installed${NC}"
    echo "   Please install Docker Compose"
    exit 1
fi
echo -e "${GREEN}✓ Docker Compose${NC}"

# Check Make
if ! command -v make &> /dev/null; then
    echo -e "${RED}❌ Make is not installed${NC}"
    echo "   Please install Make"
    exit 1
fi
echo -e "${GREEN}✓ Make${NC}"

# Check psql (optional but helpful)
if command -v psql &> /dev/null; then
    echo -e "${GREEN}✓ PostgreSQL client${NC}"
else
    echo -e "${YELLOW}⚠️  PostgreSQL client not found (optional)${NC}"
fi

echo ""

#######################################
# Install Dependencies
#######################################

echo -e "${YELLOW}📦 Installing Go dependencies...${NC}"
go mod download
echo -e "${GREEN}✓ Dependencies downloaded${NC}"
echo ""

echo -e "${YELLOW}📚 Vendoring dependencies...${NC}"
go mod vendor
echo -e "${GREEN}✓ Dependencies vendored${NC}"
echo ""

#######################################
# Install Development Tools
#######################################

echo -e "${YELLOW}🔧 Installing development tools...${NC}"

tools=(
    "github.com/golangci/golangci-lint/cmd/golangci-lint@latest|golangci-lint"
    "github.com/swaggo/swag/cmd/swag@latest|swag"
    "github.com/cosmtrek/air@latest|air"
)

for tool_info in "${tools[@]}"; do
    tool_path="${tool_info%|*}"
    tool_name="${tool_info#*|}"
    echo -e "  - Installing ${tool_name}..."
    go install "$tool_path" 2>&1 | grep -v "go: downloading" || true
done

echo -e "${GREEN}✓ Development tools installed${NC}"
echo ""

#######################################
# Setup Git Hooks
#######################################

echo -e "${YELLOW}🪝 Installing git hooks...${NC}"
if [ -f ".githooks/pre-commit" ]; then
    chmod +x .githooks/pre-commit
    git config core.hooksPath .githooks
    echo -e "${GREEN}✓ Git hooks installed${NC}"
else
    echo -e "${YELLOW}⚠️  No pre-commit hook found, skipping${NC}"
fi
echo ""

#######################################
# Setup Environment
#######################################

echo -e "${YELLOW}📝 Setting up environment...${NC}"

# Check for .env file
if [ ! -f .env ]; then
    if [ -f .env.example ]; then
        cp .env.example .env
        echo -e "${GREEN}✓ Created .env from .env.example${NC}"
        echo -e "${YELLOW}⚠️  Please review and update .env with your settings${NC}"
    else
        echo -e "${YELLOW}⚠️  No .env.example found, creating minimal .env${NC}"
        cat > .env << 'EOF'
# Database
POSTGRES_DSN=postgres://library:library123@localhost:5432/library?sslmode=disable

# Redis
REDIS_HOST=localhost:6379

# JWT
JWT_SECRET=your-secret-key-change-in-production
JWT_EXPIRY=24h

# Server
APP_MODE=dev
SERVER_PORT=8080

# Logging
LOG_LEVEL=debug
EOF
        echo -e "${GREEN}✓ Created minimal .env file${NC}"
    fi
else
    echo -e "${GREEN}✓ .env file already exists${NC}"
fi
echo ""

#######################################
# Start Docker Services
#######################################

echo -e "${YELLOW}🐳 Starting Docker services...${NC}"
cd deployments/docker
docker-compose up -d
cd ../..
echo -e "${GREEN}✓ Docker services started${NC}"
echo "  - PostgreSQL: localhost:5432"
echo "  - Redis: localhost:6379"
echo ""

#######################################
# Wait for Database
#######################################

echo -e "${YELLOW}⏳ Waiting for database to be ready...${NC}"
max_attempts=30
attempt=0

while [ $attempt -lt $max_attempts ]; do
    if docker-compose -f deployments/docker/docker-compose.yml exec -T postgres pg_isready -U library &> /dev/null; then
        echo -e "${GREEN}✓ Database is ready${NC}"
        break
    fi
    attempt=$((attempt + 1))
    echo -n "."
    sleep 1
done

if [ $attempt -eq $max_attempts ]; then
    echo -e "${RED}❌ Database failed to start within 30 seconds${NC}"
    exit 1
fi
echo ""

#######################################
# Run Migrations
#######################################

echo -e "${YELLOW}🗄️  Running database migrations...${NC}"
export POSTGRES_DSN="postgres://library:library123@localhost:5432/library?sslmode=disable"
go run cmd/migrate/main.go up
echo -e "${GREEN}✓ Migrations complete${NC}"
echo ""

#######################################
# Seed Development Data
#######################################

echo -e "${YELLOW}🌱 Seeding development data...${NC}"
if [ -f "scripts/seed-data.sh" ]; then
    bash scripts/seed-data.sh
    echo -e "${GREEN}✓ Development data seeded${NC}"
else
    echo -e "${YELLOW}⚠️  No seed script found (scripts/seed-data.sh), skipping${NC}"
fi
echo ""

#######################################
# Generate API Documentation
#######################################

echo -e "${YELLOW}📖 Generating API documentation...${NC}"
if command -v swag &> /dev/null; then
    swag init -g cmd/api/main.go -o api/openapi --parseDependency --parseInternal 2>&1 | grep -v "ParseComment" || true
    echo -e "${GREEN}✓ API documentation generated${NC}"
else
    echo -e "${YELLOW}⚠️  swag not found, skipping documentation generation${NC}"
fi
echo ""

#######################################
# Build Project
#######################################

echo -e "${YELLOW}🔨 Building project...${NC}"
make build-api
echo -e "${GREEN}✓ Build successful${NC}"
echo ""

#######################################
# Summary
#######################################

echo -e "${GREEN}========================================"
echo " ✅ Development Setup Complete!"
echo "========================================${NC}"
echo ""
echo "Your development environment is ready!"
echo ""
echo -e "${BLUE}Quick Start:${NC}"
echo "  make run              Run API server"
echo "  make test             Run all tests"
echo "  make test-unit        Run unit tests only"
echo ""
echo -e "${BLUE}Useful Commands:${NC}"
echo "  make help             Show all available commands"
echo "  make lint             Run code linters"
echo "  make ci               Run full CI pipeline locally"
echo "  make gen-docs         Regenerate API documentation"
echo ""
echo -e "${BLUE}Services:${NC}"
echo "  API Server:     http://localhost:8080"
echo "  Swagger UI:     http://localhost:8080/swagger/index.html"
echo "  PostgreSQL:     localhost:5432"
echo "  Redis:          localhost:6379"
echo ""
echo -e "${BLUE}Test Accounts:${NC}"
echo "  (Run seed script if available)"
echo ""
echo -e "${YELLOW}⚠️  Important:${NC}"
echo "  - Review .env and update JWT_SECRET before production use"
echo "  - Pre-commit hooks are now active"
echo ""
