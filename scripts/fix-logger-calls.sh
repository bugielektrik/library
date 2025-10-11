#!/bin/bash

# Fix UseCaseLogger calls that were broken by the previous script

echo "🔄 Fixing broken UseCaseLogger calls..."

# Find all files with broken logger syntax
FILES=$(find internal/usecase -name "*.go" -type f ! -name "*_test.go")

for file in $FILES; do
    # Fix broken UseCaseLogger calls
    # Pattern: logger := logutil.UseCaseLogger(ctx, "name", "operation")\n\t\tzap.Field...\n\t)
    # Should be: logger := logutil.UseCaseLogger(ctx, "name", "operation")

    # Remove orphaned zap.String lines after UseCaseLogger
    perl -i -0pe 's/logger := logutil\.UseCaseLogger\(ctx, "([^"]+)", "operation"\)\n\t\tzap\.[^\n]+,\n\t\)/logger := logutil.UseCaseLogger(ctx, "$1", "operation")/g' "$file"

    # Fix any remaining broken patterns
    perl -i -0pe 's/UseCaseLogger\(ctx, "([^"]+)", "operation"\)\n\t+zap\.[^\n]+,\n\t+\)/UseCaseLogger(ctx, "$1", "operation")/g' "$file"
done

echo "✅ Logger calls fixed!"