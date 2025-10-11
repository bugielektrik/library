# Scripts

Development and deployment automation scripts.

## Available Scripts

### `dev-setup.sh`
Complete development environment setup from scratch.

**What it does:**
- Checks prerequisites (Go, Docker, Make)
- Installs development tools (golangci-lint, mockgen, swag)
- Starts Docker services (PostgreSQL, Redis)
- Runs database migrations
- Seeds test data
- Installs git hooks

**Usage:**
```bash
./scripts/dev-setup.sh
```

### `seed-data.sh`
Populates database with test data for development.

**Creates:**
- Test users (admin@library.com, user@library.com)
- Sample books with authors
- Test payments and subscriptions

**Usage:**
```bash
./scripts/seed-data.sh
```

### `build.sh`
Builds all binaries with version information.

**Outputs:**
- `bin/library-api` - API server
- `bin/library-worker` - Background worker
- `bin/library-migrate` - Migration tool

**Usage:**
```bash
./scripts/build.sh
# or with version
VERSION=1.0.0 ./scripts/build.sh
```

### `test.sh`
Runs tests with coverage reporting.

**Features:**
- Race detection enabled
- Coverage report generation
- Optional HTML output

**Usage:**
```bash
./scripts/test.sh
# With HTML coverage
./scripts/test.sh --html
```

## Script Standards

All scripts follow these conventions:

1. **Bash shebang**: `#!/bin/bash`
2. **Error handling**: `set -e` to exit on errors
3. **Colors for output**: RED, GREEN, YELLOW, BLUE
4. **Progress indicators**: ✓ for success, ✗ for failure
5. **Idempotent**: Safe to run multiple times

## Adding New Scripts

Template for new scripts:

```bash
#!/bin/bash
set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}=== Script Purpose ===${NC}"

# Check prerequisites
if ! command -v tool &> /dev/null; then
    echo -e "${RED}✗ Tool not found${NC}"
    exit 1
fi

# Main logic
echo -e "${YELLOW}→ Doing something...${NC}"
# commands here

echo -e "${GREEN}✓ Complete!${NC}"
```

## Integration with Make

Most scripts are wrapped by Makefile targets:

| Script | Make Target |
|--------|-------------|
| `dev-setup.sh` | `make dev-setup` |
| `seed-data.sh` | `make seed` |
| `build.sh` | `make build` |
| `test.sh` | `make test` |

Prefer using Make targets for consistency.

## Permissions

All scripts must be executable:
```bash
chmod +x scripts/*.sh
```

## Environment Variables

Scripts respect these environment variables:
- `POSTGRES_DSN` - Database connection string
- `REDIS_HOST` - Redis server address
- `JWT_SECRET` - JWT signing key
- `APP_MODE` - dev/prod mode

## CI/CD Integration

GitHub Actions uses these scripts:
- `.github/workflows/ci.yml` calls `test.sh`
- Docker builds use `build.sh`

Keep scripts CI-friendly (no interactive prompts).