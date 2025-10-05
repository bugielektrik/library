# First-Time Setup Guide

> **Step-by-step instructions to get the project running**

## Quick Setup (5 Minutes)

```bash
# 1. Clone and navigate
git clone <repository-url>
cd library

# 2. Initialize
make init && make up && make migrate-up

# 3. Set JWT secret
echo "JWT_SECRET=$(openssl rand -base64 32)" >> .env

# 4. Run
make run

# 5. Test
curl http://localhost:8080/health
```

## Detailed Setup

### Step 1: Verify Go Version

```bash
go version
# Should show: go version go1.25.0 or higher
```

**If wrong version:**
- **macOS:** `brew install go@1.25`
- **Linux:** Download from https://go.dev/dl/
- **Windows:** Download installer from https://go.dev/dl/

### Step 2: Install Dependencies

```bash
go mod download
go mod tidy

# If you get "module not found" errors:
rm -rf ~/go/pkg/mod && go mod download
```

### Step 3: Start Infrastructure

**Option A: Using Docker (Recommended)**
```bash
# Verify Docker is running
docker ps

# Start PostgreSQL and Redis
make up

# Check status
docker-compose -f deployments/docker/docker-compose.yml ps
```

**Option B: Local Installation**
```bash
# Install PostgreSQL
brew install postgresql@15  # macOS
sudo apt-get install postgresql-15  # Ubuntu

# Install Redis
brew install redis  # macOS
sudo apt-get install redis-server  # Ubuntu

# Start services
brew services start postgresql@15
brew services start redis
```

### Step 4: Configure Environment

```bash
# Copy example configuration
cp .env.example .env

# Required: Set JWT secret (use a strong random string)
echo "JWT_SECRET=$(openssl rand -base64 32)" >> .env

# Optional: Customize other settings
nano .env
```

**Critical Environment Variables:**
- `JWT_SECRET` - Required for authentication
- `POSTGRES_DSN` - Database connection string
- `REDIS_HOST` - Cache server (optional, uses in-memory if missing)

### Step 5: Run Database Migrations

```bash
make migrate-up

# Verify migrations
psql -h localhost -U library -d library -c "\dt"
```

### Step 6: Start the API Server

```bash
make run

# Or for development with hot reload:
make dev
```

### Step 7: Verify Installation

```bash
# Health check
curl http://localhost:8080/health

# Should return: {"status":"ok"}
```

## Troubleshooting

### "Connection Refused" Error

**Symptoms:** Can't connect to PostgreSQL or Redis

**Solutions:**
```bash
# Check if Docker containers are running
docker ps

# Restart containers
make down && make up

# Check container logs
make docker-logs

# Verify ports are not in use
lsof -ti:5432  # PostgreSQL
lsof -ti:6379  # Redis
```

### "Port Already Allocated" Error

**Symptoms:** Docker can't start because port is in use

**Solutions:**
```bash
# Kill process using the port
lsof -ti:8080 | xargs kill -9   # API server
lsof -ti:5432 | xargs kill -9   # PostgreSQL
lsof -ti:6379 | xargs kill -9   # Redis

# Or change ports in docker-compose.yml
nano deployments/docker/docker-compose.yml
```

### "Migration Failed" Error

**Symptoms:** Database migration errors

**Solutions:**
```bash
# Wait for PostgreSQL to fully start (5 seconds), then retry
sleep 5 && make migrate-up

# Check database connection
psql -h localhost -U library -d library

# Reset database (destructive!)
make migrate-down
make migrate-up

# Or completely reset
docker-compose down -v
make up
sleep 5
make migrate-up
```

### "Module Not Found" Error

**Symptoms:** Import errors or missing packages

**Solutions:**
```bash
# Clean module cache
go clean -modcache
go mod download
go mod tidy
go mod vendor

# Verify go.mod is correct
cat go.mod | grep "module library-service"
```

### "JWT Secret Not Set" Error

**Symptoms:** Authentication fails or panic on startup

**Solutions:**
```bash
# Generate and set JWT secret
echo "JWT_SECRET=$(openssl rand -base64 32)" >> .env

# Or manually edit .env
nano .env
# Add: JWT_SECRET=your-very-long-random-secret-key-here
```

### "Tests Fail Randomly" Error

**Symptoms:** Tests pass sometimes, fail others

**Solutions:**
```bash
# Clear test cache
go clean -testcache

# Run with race detection
go test -race ./...

# Run verbose to see which test fails
go test -v ./...
```

### Platform-Specific Issues

**macOS:**
```bash
# Docker Desktop not starting
rm ~/Library/Containers/com.docker.docker/Data/vms/0/data/Docker.raw

# Permission denied on ports < 1024
# Use ports >= 1024 or run with sudo (not recommended)
```

**Linux:**
```bash
# Docker permission denied
sudo usermod -aG docker $USER
newgrp docker

# PostgreSQL authentication failed
sudo -u postgres psql -c "ALTER USER library PASSWORD 'library123';"
```

**Windows:**
```bash
# Path issues
# Use Git Bash or WSL2 for better compatibility

# Line ending issues
git config core.autocrlf false
```

## Development Tools Setup

### IDE Configuration

**VS Code:**
```bash
# Install Go extension
code --install-extension golang.go

# Settings (add to .vscode/settings.json)
{
  "go.useLanguageServer": true,
  "go.lintTool": "golangci-lint",
  "go.formatTool": "goimports",
  "go.testFlags": ["-v", "-race"]
}
```

**GoLand:**
- Go to Preferences → Go → Build Tags & Vendoring
- Enable "Use vendored dependencies"
- Set project GOROOT to Go 1.25+

### Install Development Tools

```bash
make install-tools

# Or manually:
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
go install github.com/golang/mock/mockgen@latest
go install github.com/swaggo/swag/cmd/swag@latest
```

### Database GUI Tools

**Recommended:**
- **TablePlus** (macOS/Windows) - https://tableplus.com
- **DBeaver** (Cross-platform) - https://dbeaver.io
- **pgAdmin** (PostgreSQL-specific) - https://www.pgadmin.org

**Connection Details:**
- Host: localhost
- Port: 5432
- Database: library
- Username: library
- Password: library123

### Redis GUI Tools

**Recommended:**
- **RedisInsight** - https://redis.com/redis-enterprise/redis-insight/
- **Medis** (macOS) - https://getmedis.com

**Connection Details:**
- Host: localhost
- Port: 6379
- Password: (none by default)

### API Testing Tools

**Recommended:**
- **cURL** (command line)
- **HTTPie** (command line) - https://httpie.io
- **Postman** - https://www.postman.com
- **Insomnia** - https://insomnia.rest

**Swagger UI:**
```bash
# Start server and visit
open http://localhost:8080/swagger/index.html
```

## Verification Checklist

After setup, verify everything works:

- [ ] `go version` shows 1.25.0+
- [ ] `docker ps` shows running containers
- [ ] `make test` passes all tests
- [ ] `make lint` shows no errors
- [ ] `curl http://localhost:8080/health` returns `{"status":"ok"}`
- [ ] Can register user via API
- [ ] Can login and get JWT token
- [ ] Can create a book with auth token

## Next Steps

1. Read [Architecture Guide](./architecture.md) to understand the codebase structure
2. Review [Development Workflow](./development.md) for daily development tasks
3. Check [Testing Guide](./testing.md) for testing patterns
4. Explore [API Documentation](./api.md) for endpoint details

## Getting Help

**Common Resources:**
- `make help` - Show all available commands
- `/docs` folder - Architecture and design docs
- GitHub Issues - Report problems or ask questions

**Still stuck?**
- Check Docker logs: `make docker-logs`
- Verify environment: `cat .env`
- Test database: `psql -h localhost -U library -d library`
- Check API logs: `tail -f service.log`
