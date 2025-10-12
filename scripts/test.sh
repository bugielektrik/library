#!/bin/bash

set -e

echo "Running tests..."

# Run tests with coverage
go test -v -race -coverprofile=coverage.out -covermode=atomic ./...

# Display coverage summary
echo ""
echo "Coverage summary:"
go tool cover -func=coverage.out | tail -1

# Generate HTML coverage report (optional)
if [ "$1" == "--html" ]; then
    go tool cover -html=coverage.out -o coverage.html
    echo "HTML coverage report generated: coverage.html"
fi
