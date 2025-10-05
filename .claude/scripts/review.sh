#!/bin/bash
set -e

echo "🔍 Running self-review checks..."

echo "✓ Formatting code..."
make fmt

echo "✓ Running vet..."
make vet

echo "✓ Running linter..."
make lint

echo "✓ Running tests..."
make test

echo "✓ Checking race conditions..."
go test -race ./internal/...

echo "✓ Checking for TODO without owner..."
if grep -r "TODO:" --include="*.go" ./internal | grep -v "TODO([a-z]*):"; then
    echo "❌ Found TODOs without owner. Use: // TODO(username): description"
    exit 1
fi

echo "✓ Checking for debug statements..."
if grep -r "fmt.Println\|log.Println" --include="*.go" ./internal 2>/dev/null; then
    echo "⚠️  Warning: Found debug print statements"
fi

echo "✓ Checking for hardcoded secrets..."
if grep -ri "password.*=.*\"" --include="*.go" ./internal | grep -v "_test.go" 2>/dev/null; then
    echo "❌ Possible hardcoded password found!"
    exit 1
fi

echo "✓ Checking migrations..."
if ls migrations/postgres/*.up.sql 1> /dev/null 2>&1; then
    for up in migrations/postgres/*.up.sql; do
        down="${up%.up.sql}.down.sql"
        if [ ! -f "$down" ]; then
            echo "❌ Missing down migration for $up"
            exit 1
        fi
    done
fi

echo "✓ Building..."
make build

echo "✅ All checks passed! Ready to commit."
